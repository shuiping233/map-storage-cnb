package router

import (
	"github.com/gin-gonic/gin"

	"map-storage-cnb/src/service"
	"map-storage-cnb/src/storage"
)

func RegisterAll(engine *gin.Engine) {
	storage := storage.NewLocalStorage()
	storage.Init()
	uploadAPI := &service.UploadAPI{
		Storage: storage,
	}

	v1 := engine.Group("/api/v1")
	v1.POST("/upload", uploadAPI.MapUploadApi)

	engine.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
}
