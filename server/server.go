package server

import (
	"fmt"
	"go-cdn/config"
	"go-cdn/database"
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

func SpawnGin(state *GinState, available_files *map[string]int) error {
	r := gin.Default()
	r.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	// Debug endpoint to quickly check if a file is available
	r.GET("/debug/:hash", getDebugFileHandler(available_files))

	r.GET("/content/:hash", getFileHandler(state, available_files))
	if state.Config.HTTPServer.AllowInsertion {
		r.POST("/content/", postFileHandler(state, available_files))
	}

	err := r.Run(fmt.Sprintf("0.0.0.0:%d", state.Config.HTTPServer.DeliveryPort))
	return err
}

func getFileHandler(state *GinState, available_files *map[string]int) gin.HandlerFunc {
	return func(c *gin.Context) {
		hash := c.Param("hash")
		_, ok := (*available_files)[hash]
		if ok {
			// 1. Check if in Redis
			if state.RedisClient != nil {
				bytes, err := state.RedisClient.GetFromCache(hash)
				if err != nil {
					state.Sugar.Errorf("error while retrieving from Redis: %s", err)
					c.String(http.StatusBadRequest, "")
					return
				}
				if bytes != nil {
					c.Data(http.StatusOK, "image", bytes)
					return
				}
			}
			// 2. Check if in Postgres
		} else {
            state.PgClient.GetFile(hash)
			c.String(http.StatusOK, "BOOO")
		}
	}
}

func getDebugFileHandler(available_files *map[string]int) gin.HandlerFunc {
	return func(c *gin.Context) {
		_, ok := (*available_files)[c.Param("hash")]
		if ok {
			c.String(http.StatusOK, "OK")
		} else {
			c.String(http.StatusOK, "BOOO")
		}
	}
}

func postFileHandler(state *GinState, available_files *map[string]int) gin.HandlerFunc {
	return func(c *gin.Context) {
		hash := c.PostForm("hash")
		_, ok := (*available_files)[hash]
		if ok {
			state.Sugar.Errorf("hash already set")
			c.String(http.StatusForbidden, "Invalid Parameters")
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
				"bytes", string(bytes), "err", err)

			if err != nil {
				state.Sugar.Errorf("got an error while uploading: %s", err)
				c.String(http.StatusBadRequest, "")
				return
			}

			err = state.PgClient.AddFile(hash, filename, bytes)
			if err != nil {
				state.Sugar.Errorf("got an error adding a file: %s", err)
				c.String(http.StatusBadRequest, "")
				return
			}

			available_files, err = state.PgClient.GetFileList()
			if err != nil {
				state.Sugar.Errorf("got an error refreshing current files: %s", err)
				c.String(http.StatusBadRequest, "")
				return
			}

			c.String(http.StatusOK, "OK")
		}
	}
}
