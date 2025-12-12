package storage

import (
	"context"
	"log"

	"map-storage-cnb/src/model"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type StorageDB struct {
	cfg model.StorageDBConfig
	DB  *gorm.DB
}

func DBInit(cfg model.StorageConfig) (*StorageDB, error) {
	// TODO  mysql之类的适配器还没做

	db, err := gorm.Open(sqlite.Open(cfg.DB.URL), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	err = db.AutoMigrate(&model.MapMetaData{})
	if err != nil {
		log.Fatal(err)
	}
	return &StorageDB{DB: db, cfg: cfg.DB}, nil

}

func (s *StorageDB) Add(ctx context.Context, metaData model.MapMetaData) error {
	return s.DB.WithContext(ctx).Create(&metaData).Error
}

func (s *StorageDB) Delete(ctx context.Context, hash string) error {
	return s.DB.WithContext(ctx).Delete(&model.MapMetaData{}, "hash = ?", hash).Error
}

// 只更新非零字段
func (s *StorageDB) Update(ctx context.Context, metaData model.MapMetaData) error {
	return s.DB.WithContext(ctx).
		Where("hash = ?", metaData.Hash).
		Updates(&metaData).Error
}

func (s *StorageDB) Get(ctx context.Context, hash string) (*model.MapMetaData, error) {
	var result model.MapMetaData
	err := s.DB.WithContext(ctx).
		Where("hash = ?", hash).
		First(&result).Error // 找到第一条记录
	if err != nil {
		return nil, err
	}
	return &result, nil
}

// 模糊查询 name , 也就是地图名称 , 默认查10个
func (s *StorageDB) Search(ctx context.Context, keyword string, limit int) ([]model.MapMetaData, error) {
	if limit <= 0 {
		limit = 10
	}
	var result []model.MapMetaData
	err := s.DB.WithContext(ctx).
		Where("name LIKE ?", "%"+keyword+"%").
		Limit(limit).
		Find(&result).Error
	return result, err
}

// 精确查询 name , 也就是地图名称 , 默认查10个
func (s *StorageDB) SearchExact(ctx context.Context, keyword string, limit int) ([]model.MapMetaData, error) {
	if limit <= 0 {
		limit = 10
	}
	var result []model.MapMetaData
	err := s.DB.WithContext(ctx).
		Where("name = ?", keyword).
		Limit(limit).
		Find(&result).Error
	return result, err
}

// 列出元数据
func (s *StorageDB) List(ctx context.Context, page int, desc bool, orderField string, limit int) ([]model.MapMetaData, error) {
	if limit <= 0 {
		limit = 10
	}
	if page <= 0 {
		page = 1
	}
	order := orderField + "ASC"
	if desc {
		order = orderField + "DESC"
	}

	var result []model.MapMetaData
	err := s.DB.WithContext(ctx).
		Order(order).
		Limit(limit).
		Offset((page - 1) * limit).
		Find(&result).Error
	return result, err
}
func (s *StorageDB) Close() error {
	db, err := s.DB.DB()
	if err != nil {
		return err
	}
	return db.Close()
}
