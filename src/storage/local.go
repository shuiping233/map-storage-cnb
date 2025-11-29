package storage

import (
	"context"
	"io"
	"log"
	"map-storage-cnb/src/model"
	"map-storage-cnb/src/utils"
	"os"
	"path/filepath"
)

const (
	DbPath = "./tmp/test.db"
	tmpDir = "./tmp"
)

type LocalStorage struct {
	DB *StorageDB
}

func joinTmpPath(name string) string {
	return filepath.Join(tmpDir, name)
}

func NewLocalStorage() *LocalStorage {
	return &LocalStorage{}
}

func (s *LocalStorage) Init() error {
	utils.InitDefaultDir()
	db, err := DBInit(DbPath)
	if err != nil {
		log.Fatal(err)
	}
	s.DB = db
	return nil
}
func (s *LocalStorage) Close() error {
	s.DB.Close()
	return nil
}

func (s *LocalStorage) Save(ctx context.Context, metaData model.MapMetaData, reader io.Reader) (*model.MapMetaData, error) {
	hash, err := utils.HashFile(reader)
	if err != nil {
		return nil, err
	}
	path, err := os.OpenFile(joinTmpPath(hash), os.O_CREATE, 0655)
	if err != nil {
		return nil, err
	}
	defer path.Close()
	_, err = io.Copy(path, reader)
	if err != nil {
		return nil, err
	}
	err = s.DB.Add(ctx, metaData)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

func (s *LocalStorage) Get(ctx context.Context, hash string, writer io.Writer) (*model.MapMetaData, error) {
	metaData, err := s.DB.Get(ctx, hash)
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

func (s *LocalStorage) GetMeta(ctx context.Context, hash string) (*model.MapMetaData, error) {
	metaData, err := s.DB.Get(ctx, hash)
	if err != nil {
		return nil, err
	}
	return metaData, nil
}

func (s *LocalStorage) GetHistory(ctx context.Context, hash string, limit int) ([]model.MapMetaData, error) {
	var result []model.MapMetaData
	for {
		metaData, err := s.DB.Get(ctx, hash)
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

func (s *LocalStorage) Exists(ctx context.Context, hash string) (bool, error) {
	_, err := s.DB.Get(ctx, hash)
	if err != nil {
		return false, err
	}
	return true, nil
}

// 精确查 name 地图名称, 返回元数据列表
func (s *LocalStorage) SearchExact(ctx context.Context, name string, limit int) ([]model.MapMetaData, error) {
	if limit <= 0 {
		limit = 10
	}
	result, err := s.DB.SearchExact(ctx, name, limit)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *LocalStorage) Search(ctx context.Context, name string, limit int) ([]model.MapMetaData, error) {
	if limit <= 0 {
		limit = 10
	}
	result, err := s.DB.Search(ctx, name, limit)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *LocalStorage) List(ctx context.Context, page int, desc bool, orderField string, limit int) ([]model.MapMetaData, error) {
	result, err := s.DB.List(ctx, page, desc, orderField, limit)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (s *LocalStorage) Delete(ctx context.Context, hash string) error {
	err := os.Remove(joinTmpPath(hash))
	if err != nil {
		return err
	}
	err = s.DB.Delete(ctx, hash)
	if err != nil {
		return err
	}
	return nil
}
