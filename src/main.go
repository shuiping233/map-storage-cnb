package main

import (
	"github.com/gin-gonic/gin"

	"map-storage-cnb/src/router"
	"map-storage-cnb/src/service"
)

const (
	// 默认端口
	defaultPort = "8080"
)

func main() {
	engine := gin.Default()
	router.RegisterAll(engine)
	engine.MaxMultipartMemory = service.MaxFormMem
	engine.Run(":" + defaultPort) // 默认监听 0.0.0.0:8080
}
