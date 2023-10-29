package server

import (
	"fmt"
	"go-cdn/config"
	"go-cdn/database"
	"go-cdn/utils"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

type GinState struct {
	Config      *config.Config
	RedisClient *database.RedisClient
	PgClient    *database.PostgresClient
	Sugar       *zap.SugaredLogger
}

func SpawnGin(state *GinState) error {
	r := gin.Default()
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
		hash := c.Param("hash")
		// 1. Check if in Redis
		if state.Config.Redis.RedisEnable {
			bytes, err := state.RedisClient.GetFromCache(hash)
			if err != nil {
				// Cache miss, the request is still good
				state.Sugar.Errorf("error while retrieving from Redis: %s", err)
			}
			if bytes != nil {
				c.Data(http.StatusOK, "image", bytes)
				return
			}
		}

		// 2. Get from Postgres
		stored_file, err := state.PgClient.GetFile(hash)
		if err != nil {
			state.Sugar.Errorf("error while retrieving from Postgres: %s", err)
			c.String(http.StatusBadRequest, "")
			return
		}
		c.Data(http.StatusOK, "image", stored_file.Content)
	}
}

// POST with file, filename
func postFileHandler(state *GinState) gin.HandlerFunc {
	return func(c *gin.Context) {
		hash := utils.RandStringBytes(6)
		_, err := state.PgClient.GetFile(hash)
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

			err = state.PgClient.AddFile(hash, filename, bytes)
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
		hash := c.Param("hash")
		state.Sugar.Infof("removing %s image", hash)
		if state.Config.Redis.RedisEnable {
			_, err := state.RedisClient.RemoveFromCache(hash)
			if err != nil {
				state.Sugar.Errorf("error while removing from Redis: %s", err)
			}
		}

		err := state.PgClient.RemoveFile(hash)
		if err != nil {
			state.Sugar.Errorf("error while removing from Postgres: %s", err)
		}

		c.String(http.StatusOK, "OK")
	}
}
