package server

import (
	"fmt"
	"go-cdn/config"
	"go-cdn/database"
	"go-cdn/tracing"
	"go-cdn/utils"
	"io"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type GinState struct {
	Config      *config.Config
	RedisClient *database.RedisClient
	PgClient    *database.PostgresClient
	Sugar       *zap.SugaredLogger
}

func requestMetadataMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		err_ch := make(chan error, 1)
		c.Set("err_ch", err_ch)

		wg := sync.WaitGroup{}
		c.Set("wg", &wg)

		req_id, _ := uuid.NewRandom()
		c.Set("request.id", req_id.String())

		_, span := tracing.Tracer.Start(c.Request.Context(), "requestMetadataMiddleware",
			trace.WithAttributes(attribute.String("request.id", req_id.String())),
		)

		span.End() // defer would make the middleware span terminate at the end of the request
		c.Next()
	}
}
func errorPropagatorMiddleware() gin.HandlerFunc {
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

func SpawnGin(state *GinState) error {
	r := gin.Default()

	r.Use(otelgin.Middleware("gin-server"))
	r.Use(requestMetadataMiddleware())
	r.Use(errorPropagatorMiddleware())

	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	r.GET("/content/:hash", getFileHandler(state))

	if state.Config.HTTPServer.AllowInsertion {
		r.POST("/content/", postFileHandler(state))
	}

	if state.Config.HTTPServer.AllowDeletion {
		r.DELETE("/content/:hash", deleteFileHandler(state))
	}

	err := r.Run(fmt.Sprintf("0.0.0.0:%d", state.Config.HTTPServer.DeliveryPort))
	return err
}

// Returns the file, trying first from Redis and then from Postgres
func getFileHandler(state *GinState) gin.HandlerFunc {
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
		if state.Config.Redis.RedisEnable {
			bytes, err := state.RedisClient.GetFromCache(c.Request.Context(), hash)

			// Cache miss, the request is still good
			if err != nil {
				state.Sugar.Errorf("error while retrieving from Redis: %s", err)
				err_ch <- err // Only with a buffered ch
			}
			if bytes != nil {
				c.Data(http.StatusOK, "image", bytes)
				return
			}
		}

		// Get from Postgres
		stored_file, err := state.PgClient.GetFile(c.Request.Context(), hash)
		if err != nil {
			state.Sugar.Errorf("error while retrieving from Postgres: %s", err)
			c.String(http.StatusBadRequest, "")
			return
		}

		// Asynchronously add to Redis cache
		if state.Config.Redis.RedisEnable {
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := state.RedisClient.AddToCache(c.Request.Context(), hash, stored_file.Content)
				err_ch <- err
			}()
		}

		_, internal_span := tracing.Tracer.Start(c.Request.Context(), "sendData")
		c.Data(http.StatusOK, "image", stored_file.Content)
		internal_span.End()
	}
}

// POST with file, filename
func postFileHandler(state *GinState) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, span := tracing.Tracer.Start(c.Request.Context(), "postFileHandler")
		defer span.End()

		hash := utils.RandStringBytes(6)
		_, err := state.PgClient.GetFile(c.Request.Context(), hash)
		if err != nil {
			state.Sugar.Errorf("hash already set")
			c.String(http.StatusForbidden, "Invalid HashName")
		} else {
			filename := c.PostForm("filename")
			file, err := c.FormFile("file")
			if err != nil {
				state.Sugar.Errorf("got an error while uploading: %s", err)
				c.String(http.StatusBadRequest, "")
				return
			}

			stream, err := file.Open()
			if err != nil {
				state.Sugar.Errorf("got an error while uploading: %s", err)
				c.String(http.StatusBadRequest, "")
				return
			}
			defer stream.Close()

			bytes, err := io.ReadAll(stream)
			if err != nil {
				state.Sugar.Errorf("got an error while uploading: %s", err)
				c.String(http.StatusBadRequest, "")
				return
			}

			state.Sugar.Infow("adding an image",
				"filename", filename,
				"bytes", string(bytes)[:10], "err", err)

			err = state.PgClient.AddFile(c.Request.Context(), hash, filename, bytes)
			if err != nil {
				state.Sugar.Errorf("got an error adding a file: %s", err)
				c.String(http.StatusBadRequest, "")
				return
			}

			c.JSON(http.StatusOK, gin.H{
				"hash": hash,
			})
		}
	}
}

// Doesn't return an HTTP error by design
func deleteFileHandler(state *GinState) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, span := tracing.Tracer.Start(c.Request.Context(), "deleteFileHandler")
		defer span.End()

		hash := c.Param("hash")
		state.Sugar.Infof("removing %s image", hash)
		if state.Config.Redis.RedisEnable {
			// TODO make async
			_, err := state.RedisClient.RemoveFromCache(c.Request.Context(), hash)
			if err != nil {
				state.Sugar.Errorf("error while removing from Redis: %s", err)
			}
		}

		err := state.PgClient.RemoveFile(c.Request.Context(), hash)
		if err != nil {
			state.Sugar.Errorf("error while removing from Postgres: %s", err)
		}

		c.String(http.StatusOK, "OK")
	}
}
