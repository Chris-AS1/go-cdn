package server

import (
	"context"
	"fmt"
	"go-cdn/internal/config"
	"go-cdn/internal/database"
	"go-cdn/internal/tracing"
	"go-cdn/pkg/model"
	"go-cdn/pkg/utils"
	"io"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/ratelimit"
	"go.uber.org/zap"
)

type GinState struct {
	Config *config.Config
	Cache  *database.Controller
	DB     *database.Controller
	Sugar  *zap.SugaredLogger
	limit  ratelimit.Limiter
	rps    int
}

func New(cfg *config.Config, db *database.Controller, cache *database.Controller, sugar *zap.SugaredLogger) *GinState {
	g := &GinState{
		Config: cfg,
		Cache:  cache,
		DB:     db,
		Sugar:  sugar,
	}

	if g.Config.HTTPServer.RateLimitEnable {
		g.rps = g.Config.HTTPServer.RateLimit
		// The rate limiter gets applied on *concurrent* requests, to change the behavior use WithoutSlack
		g.limit = ratelimit.New(g.rps, ratelimit.WithSlack(10))
		g.Sugar.Infow("using leakyBucket", "rps", g.rps)
	}

	return g
}

func (g *GinState) requestMetadataMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Adds the error channel, wg, and id to the request's Context
		err_ch := make(chan error, 1)
		c.Set("err_ch", err_ch)

		wg := sync.WaitGroup{}
		c.Set("wg", &wg)

		req_id, _ := uuid.NewRandom()
		req_path := c.Request.URL.Path
		c.Set("request.id", req_id.String())

		// Attaches request.id to the root span
		span := trace.SpanFromContext(c.Request.Context())
		span.SetAttributes(attribute.String("request.id", req_id.String()))
		span.SetAttributes(attribute.String("request.path", req_path))
		span.SetAttributes(attribute.String("service.id", g.Config.Consul.ConsulServiceID))

		c.Next()
	}
}

func (g *GinState) errorPropagatorMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		span := trace.SpanFromContext(c.Request.Context()) // Here to attach the error to the root span
		c.Next()

		// _, span := tracing.Tracer.Start(c.Request.Context(), "errorPropagatorMiddleware") // Use this to attach to a dedicated span
		defer span.End()

		err_ch := c.MustGet("err_ch").(chan error)
		wg := c.MustGet("wg").(*sync.WaitGroup)

		go func() {
			wg.Wait()
			close(err_ch)
			// Can't put error propagation in here since it would end after the span_end
		}()

		for err := range err_ch {
			if err != nil {
				span.RecordError(err)
				span.SetStatus(codes.Error, err.Error())
			}
		}
	}
}

func (g *GinState) leakBucket() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		g.limit.Take()
	}
}

type OptFunc func()

func WithMode(mode string) OptFunc {
	return func() {
		gin.SetMode(mode)
	}
}
func (g *GinState) Spawn(opts ...OptFunc) {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Apply optional values
	for _, opt := range opts {
		opt()
	}

	r := gin.Default()

	if g.Config.HTTPServer.RateLimitEnable {
		r.Use(g.leakBucket())
	}

	r.Use(otelgin.Middleware("gin-server"))
	r.Use(g.requestMetadataMiddleware())
	r.Use(g.errorPropagatorMiddleware())

	r.GET("/health", func(c *gin.Context) {
		String(c, http.StatusOK, "OK")
	})

	r.GET("/content/:hash", g.getFileHandler())
	r.GET("/content/list", g.getFileListHandler())

	if g.Config.HTTPServer.AllowInsertion {
		r.POST("/content/", g.postFileHandler())
	}

	if g.Config.HTTPServer.AllowDeletion {
		r.DELETE("/content/:hash", g.deleteFileHandler())
	}

	srv := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", g.Config.HTTPServer.DeliveryPort),
		Handler: r,
	}

	// Start the server
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			g.Sugar.Panicw("listen", "err", err)
		}
	}()

	// Listen for the interrupt signal.
	<-ctx.Done()

	// Restore default behavior on the interrupt signal and notify user of shutdown.
	stop()
	g.Sugar.Info("shutting down gracefully, press Ctrl+C again to force")

	// The context is used to inform the server it has 5 seconds to finish the request it is currently handling
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		g.Sugar.Panicw("server forced to shutdown", "err", err)
	}
}

// GET handler to retrieve an image
func (g *GinState) getFileHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		rq_ctx, span := tracing.Tracer.Start(c.Request.Context(), "getFileHandler")
		c.Request = c.Request.WithContext(rq_ctx)
		defer span.End()

		// Setup error propagation
		err_ch := c.MustGet("err_ch").(chan error)
		wg := c.MustGet("wg").(*sync.WaitGroup)

		hash := c.Param("hash")

		if g.Config.Redis.RedisEnable {
			file, err := g.Cache.GetFile(c.Request.Context(), hash)
			bytes := file.Content

			// Cache miss, the request is still good
			if err != nil {
				g.Sugar.Infow("cache miss", "err", err)
				err_ch <- err // Only with a buffered ch
			}
			if bytes != nil {
				Data(c, http.StatusOK, "image", bytes)
				return
			}
		}

		stored_file, err := g.DB.GetFile(c.Request.Context(), hash)
		if err != nil {
			g.Sugar.Errorw("db file miss", "err", err)

			String(c, http.StatusBadRequest, "")
			return
		}

		// Asynchronously add to Redis cache
		if g.Config.Redis.RedisEnable {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := g.Cache.AddFile(c.Request.Context(), stored_file)
				err_ch <- err
			}()
		}

		Data(c, http.StatusOK, "image", stored_file.Content)
	}
}

// POST handler to add an image
func (g *GinState) postFileHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, span := tracing.Tracer.Start(c.Request.Context(), "postFileHandler")
		defer span.End()

		hash := utils.RandStringBytes(6)
		_, err := g.DB.GetFile(c.Request.Context(), hash)

		if err != nil {
			g.Sugar.Infow("hash already set", "err", err)
			String(c, http.StatusForbidden, "Invalid HashName")
		} else {
			filename := c.PostForm("filename")
			file, err := c.FormFile("file")
			if err != nil {
				g.Sugar.Errorw("FormFile", "err", err)
				String(c, http.StatusBadRequest, "")
				return
			}

			stream, err := file.Open()
			if err != nil {
				g.Sugar.Errorw("FileOpen", "err", err)
				String(c, http.StatusBadRequest, "")
				return
			}
			defer stream.Close()

			bytes, err := io.ReadAll(stream)
			if err != nil {
				g.Sugar.Errorw("ReadAll", "err", err)
				String(c, http.StatusBadRequest, "")
				return
			}

			g.Sugar.Infow("adding an image",
				"filename", filename,
				"bytes", string(bytes)[:6],
				"err", err)

			err = g.DB.AddFile(c.Request.Context(), &model.StoredFile{
				IDHash:   hash,
				Filename: filename,
				Content:  bytes,
			})

			if err != nil {
				g.Sugar.Errorw("db add file", "err", err)
				String(c, http.StatusBadRequest, "")
				return
			}

			JSON(c, http.StatusOK, gin.H{
				"hash": hash,
			})
		}
	}
}

// DELETE handler to remove an image. Doesn't return an HTTP error by design
func (g *GinState) deleteFileHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, span := tracing.Tracer.Start(c.Request.Context(), "deleteFileHandler")
		defer span.End()

		// Setup error propagation
		err_ch := c.MustGet("err_ch").(chan error)
		wg := c.MustGet("wg").(*sync.WaitGroup)
		hash := c.Param("hash")

		if g.Config.Redis.RedisEnable {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := g.Cache.RemoveFile(c.Request.Context(), hash)
				err_ch <- err
			}()
		}

		err := g.DB.RemoveFile(c.Request.Context(), hash)
		if err != nil {
			g.Sugar.Errorw("db remove file", "err", err)
			wg.Add(1)
			go func(err error) {
				defer wg.Done()
				err_ch <- err
			}(err)
		}

		String(c, http.StatusOK, "OK")
	}
}

// GET handler to retrieve a list of currently stored files
func (g *GinState) getFileListHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		rq_ctx, span := tracing.Tracer.Start(c.Request.Context(), "getFileListHandler")
		c.Request = c.Request.WithContext(rq_ctx)
		defer span.End()

		// Setup error propagation
		err_ch := c.MustGet("err_ch").(chan error)
		wg := c.MustGet("wg").(*sync.WaitGroup)

		file_list, err := g.DB.GetFileList(c.Request.Context())
		if err != nil {
			g.Sugar.Errorw("db get file list", "err", err)
			wg.Add(1)
			go func(err error) {
				defer wg.Done()
				err_ch <- err
			}(err)
		}

		JSON(c, http.StatusOK, gin.H{
			"list": file_list,
		})
	}
}
