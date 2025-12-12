package config

import (
	"log"
	"map-storage-cnb/src/model"
	"os"

	"github.com/gookit/config/v2"
)

const (
	// 默认端口
	DefaultPort     = "8080"
	LocalStorageDir = "./uploads"
	DbName          = "FileMeta.db"
)

// 默认写默认json配置到目标路径
func WriteDefaultConfig(path string) error {
	log.Printf("Writing default config to %q", path)
	loader := config.New("config").WithOptions(config.ParseDefault)

	var cfg model.Config
	if err := loader.Decode(&cfg); err != nil {
		return err
	}

	jsonBytes, err := cfg.JsonDumpBytes("")
	if err != nil {
		return err
	}
	err = os.WriteFile(path, jsonBytes, 0644)
	if err != nil {
		return err
	}
	return nil
}

func Load(path string) (*model.Config, error) {
	loader := config.New("config").WithOptions(config.ParseDefault)
	loader.AddDriver(config.JSONDriver)

	err := loader.LoadFiles(path)
	if err != nil {
		log.Printf("load config error: %v", err)
	}

	var cfg model.Config
	if err := loader.Decode(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
