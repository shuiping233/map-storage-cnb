package main

import "github.com/gin-gonic/gin"

const (
	// 默认端口
	defaultPort = "8080"
)

func main() {
	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	router.Run(":" + defaultPort) // 默认监听 0.0.0.0:8080
}
