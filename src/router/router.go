package router

import (
	"fmt"

	"github.com/gin-gonic/gin"

	"map-storage-cnb/src/model"
	"map-storage-cnb/src/service"
	"map-storage-cnb/src/storage"
)

func InitStorage(cfg model.StorageConfig) (*storage.Interface, error) {
	var storageService storage.Interface
	switch cfg.Type {
	case storage.StorageTypeLocalStorage:
		storageService = storage.NewLocalStorage()
	case storage.StorageTypeGitStorage:
		storageService = storage.NewGitStorage()
	default:
		return nil, fmt.Errorf("unknown storage type %q", cfg.Type)
	}

	storageService.Init(cfg)

	return &storageService, nil
}
func RegisterAll(engine *gin.Engine, cfg model.StorageConfig) error {

	storage, err := InitStorage(cfg)
	if err != nil {
		return err
	}

	uploadAPI := &service.UploadAPI{
		Storage: *storage,
	}

	v1 := engine.Group("/api/v1")
	v1.POST("/upload", uploadAPI.MapUploadApi)

	return nil
}
