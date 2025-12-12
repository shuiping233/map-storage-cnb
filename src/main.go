package main

import (
	"log"
	"map-storage-cnb/src/config"
	"map-storage-cnb/src/router"
	"map-storage-cnb/src/service"
	"map-storage-cnb/src/utils"
	"os"
	"path"

	"github.com/gin-gonic/gin"
)

const ConfigPath = "./config/config.json"

func main() {
	log.SetFlags(log.Lshortfile)

	err := os.MkdirAll(path.Dir(ConfigPath), 0755)
	if err != nil {
		log.Fatalln(err)
	}
	if !utils.FileExists(ConfigPath) {
		err := config.WriteDefaultConfig(ConfigPath)
		if err != nil {
			log.Fatalln(err)
		}
		os.Exit(1)
	}

	cfg, err := config.Load(ConfigPath)
	if err != nil {
		log.Fatalln(err)
	}
	cfg.Print("")

	engine := gin.Default()
	err = router.RegisterAll(engine, cfg.Storage)
	if err != nil {
		log.Fatalln(err)
	}
	engine.MaxMultipartMemory = service.MaxFormMem
	engine.Run(cfg.Service.Host + ":" + cfg.Service.Port)
}
