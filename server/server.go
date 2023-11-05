package server

import (
	"context"
	"fmt"
	"go-cdn/config"
	"go-cdn/database"
	"go-cdn/tracing"
	"go-cdn/utils"
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

type GinServer struct {
	Config      *config.Config
	RedisClient *database.RedisClient
	PgClient    *database.PostgresClient
	Sugar       *zap.SugaredLogger
	limit       ratelimit.Limiter
	rps         int
}

func (g *GinServer) requestMetadataMiddleware() gin.HandlerFunc {
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

func (g *GinServer) errorPropagatorMiddleware() gin.HandlerFunc {
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

func (g *GinServer) leakBucket() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		g.limit.Take()
		// now := g.limit.Take()
		// _ = now.Sub(prev)
		// g.Sugar.Infof("%v", now.Sub(prev))
	}
}

func (g *GinServer) Spawn() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	r := gin.Default()

	if g.Config.HTTPServer.RateLimitEnable {
		g.rps = g.Config.HTTPServer.RateLimit
		// The rate limiter gets applied on *concurrent* requests, to change the behavior use WithoutSlack
		g.limit = ratelimit.New(g.rps, ratelimit.WithSlack(10))
		g.Sugar.Infof("Using leakyBucket %d/rps", g.rps)
		r.Use(g.leakBucket())
	}

	r.Use(otelgin.Middleware("gin-server"))
	r.Use(g.requestMetadataMiddleware())
	r.Use(g.errorPropagatorMiddleware())

	r.GET("/health", func(c *gin.Context) {
		String(c, http.StatusOK, "OK")
	})

	r.GET("/content/:hash", g.getFileHandler())

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
			g.Sugar.Panicf("listen: %s", err)
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
		g.Sugar.Panicf("server forced to shutdown: %s", err)
	}
}

// GET handler to retrieve an image
func (g *GinServer) getFileHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		rq_ctx, span := tracing.Tracer.Start(c.Request.Context(), "getFileHandler")
		c.Request = c.Request.WithContext(rq_ctx)
		defer span.End()

		// Setup error propagation
		err_ch := c.MustGet("err_ch").(chan error)
		wg := c.MustGet("wg").(*sync.WaitGroup)

		// Param
		hash := c.Param("hash")

		// Check if in Redis
		if g.Config.Redis.RedisEnable {
			bytes, err := g.RedisClient.GetFromCache(c.Request.Context(), hash)

			// Cache miss, the request is still good
			if err != nil {
				g.Sugar.Errorf("error while retrieving from Redis: %s", err)
				err_ch <- err // Only with a buffered ch
			}
			if bytes != nil {
				Data(c, http.StatusOK, "image", bytes)
				return
			}
		}

		// Get from Postgres
		stored_file, err := g.PgClient.GetFile(c.Request.Context(), hash)
		if err != nil {
			g.Sugar.Errorf("error while retrieving from Postgres: %s", err)

			String(c, http.StatusBadRequest, "")
			return
		}

		// Asynchronously add to Redis cache
		if g.Config.Redis.RedisEnable {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := g.RedisClient.AddToCache(c.Request.Context(), hash, stored_file.Content)
				err_ch <- err
			}()
		}

		Data(c, http.StatusOK, "image", stored_file.Content)
	}
}

// POST handler to add an image
func (g *GinServer) postFileHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, span := tracing.Tracer.Start(c.Request.Context(), "postFileHandler")
		defer span.End()

		hash := utils.RandStringBytes(6)
		_, err := g.PgClient.GetFile(c.Request.Context(), hash)
		if err != nil {
			g.Sugar.Errorf("hash already set")
			String(c, http.StatusForbidden, "Invalid HashName")
		} else {
			filename := c.PostForm("filename")
			file, err := c.FormFile("file")
			if err != nil {
				g.Sugar.Errorf("got an error while uploading: %s", err)
				String(c, http.StatusBadRequest, "")
				return
			}

			stream, err := file.Open()
			if err != nil {
				g.Sugar.Errorf("got an error while uploading: %s", err)
				String(c, http.StatusBadRequest, "")
				return
			}
			defer stream.Close()

			bytes, err := io.ReadAll(stream)
			if err != nil {
				g.Sugar.Errorf("got an error while uploading: %s", err)
				String(c, http.StatusBadRequest, "")
				return
			}

			g.Sugar.Infow("adding an image",
				"filename", filename,
				"bytes", string(bytes)[:10], "err", err)

			err = g.PgClient.AddFile(c.Request.Context(), hash, filename, bytes)
			if err != nil {
				g.Sugar.Errorf("got an error adding a file: %s", err)
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
func (g *GinServer) deleteFileHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, span := tracing.Tracer.Start(c.Request.Context(), "deleteFileHandler")
		defer span.End()

		// Setup error propagation
		err_ch := c.MustGet("err_ch").(chan error)
		wg := c.MustGet("wg").(*sync.WaitGroup)

		hash := c.Param("hash")
		g.Sugar.Infof("removing %s image", hash)
		if g.Config.Redis.RedisEnable {
			wg.Add(1)
			go func() {
				defer wg.Done()
				_, err := g.RedisClient.RemoveFromCache(c.Request.Context(), hash)
				err_ch <- err
			}()
		}

		err := g.PgClient.RemoveFile(c.Request.Context(), hash)
		if err != nil {
			g.Sugar.Errorf("error while removing from Postgres: %s", err)
		}

		String(c, http.StatusOK, "OK")
	}
}
