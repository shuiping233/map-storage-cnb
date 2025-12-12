package storage

import (
	"context"
	"errors"
	"io"
	"log"
	"map-storage-cnb/src/config"
	"map-storage-cnb/src/model"
	"map-storage-cnb/src/utils"
	"os"
	"path/filepath"

	"gorm.io/gorm"
)

const (
	StorageTypeLocalStorage model.StorageType = "LocalStorage"
)

type LocalStorage struct {
	cfg model.LocalStorageConfig
	DB  *StorageDB
}

func joinTmpPath(name string) string {
	return filepath.Join(config.LocalStorageDir, name)
}

func NewLocalStorage() *LocalStorage {
	return &LocalStorage{}
}

func (g *LocalStorage) Init(cfg model.StorageConfig) error {
	g.cfg = cfg.LocalStorage
	utils.InitDefaultDir()
	db, err := DBInit(cfg)
	if err != nil {
		log.Fatal(err)
	}
	g.DB = db
	return nil
}
func (s *LocalStorage) Close() error {
	s.DB.Close()
	return nil
}

func (g *LocalStorage) Save(ctx context.Context, metaData model.MapMetaData, data []byte) (*model.MapMetaData, error) {
	metaData.SetStorageType(StorageTypeLocalStorage)
	if err := os.WriteFile(joinTmpPath(metaData.Hash), data, 0655); err != nil {
		return nil, err
	}
	metaData.SetStorageStatus(model.MapUploadStatusSuccess, "")
	if err := g.DB.Add(ctx, metaData); err != nil {
		return nil, err
	}
	return nil, nil
}

func (g *LocalStorage) Get(ctx context.Context, hash string, writer io.Writer) (*model.MapMetaData, error) {
	metaData, err := g.DB.Get(ctx, hash)
	if err != nil {
		return nil, err
	}
	file, err := os.Open(joinTmpPath(hash))
	if err != nil {
		return nil, err
	}
	defer file.Close()

	_, err = io.Copy(writer, file)
	if err != nil {
		return nil, err
	}

	return metaData, nil
}

func (g *LocalStorage) GetMeta(ctx context.Context, hash string) (*model.MapMetaData, error) {
	metaData, err := g.DB.Get(ctx, hash)
	if err != nil {
		return nil, err
	}
	return metaData, nil
}

func (g *LocalStorage) GetHistory(ctx context.Context, hash string, limit int) ([]model.MapMetaData, error) {
	var result []model.MapMetaData
	for {
		metaData, err := g.DB.Get(ctx, hash)
		if err != nil {
			return nil, err
		}
		result = append(result, *metaData)
		hash = metaData.PrevHash
		if hash == "" {
			break
		}
	}
	return result, nil
}

func (g *LocalStorage) Exists(ctx context.Context, hash string) (bool, error) {
	_, err := g.DB.Get(ctx, hash)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// 精确查 name 地图名称, 返回元数据列表
func (g *LocalStorage) SearchExact(ctx context.Context, name string, limit int) ([]model.MapMetaData, error) {
	return g.DB.SearchExact(ctx, name, limit)
}

func (g *LocalStorage) Search(ctx context.Context, name string, limit int) ([]model.MapMetaData, error) {
	result, err := g.DB.Search(ctx, name, limit)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (g *LocalStorage) List(ctx context.Context, page int, desc bool, orderField string, limit int) ([]model.MapMetaData, error) {
	return g.DB.List(ctx, page, desc, orderField, limit)
}

func (g *LocalStorage) Delete(ctx context.Context, hash string) error {
	err := os.Remove(joinTmpPath(hash))
	if err != nil {
		return err
	}
	err = g.DB.Delete(ctx, hash)
	if err != nil {
		return err
	}
	return nil
}
