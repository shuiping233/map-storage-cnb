package storage

import (
	"context"
	"io"
	"map-storage-cnb/src/model"
)

type Interface interface {
	Init() error

	Close() error

	// 写：从 reader 读直到 EOF 并计算哈希，返回最终 Meta
	Save(ctx context.Context, metaData model.MapMetaData, reader io.Reader) (*model.MapMetaData, error)

	// 读：把文件内容 copy 到 writer
	Get(ctx context.Context, hash string, writer io.Writer) (*model.MapMetaData, error)

	GetMeta(ctx context.Context, hash string) (*model.MapMetaData, error)

	GetHistory(ctx context.Context, hash string, limit int) ([]model.MapMetaData, error)

	// 是否存在
	Exists(ctx context.Context, hash string) (bool, error)

	// 精确文件名查询
	SearchExact(ctx context.Context, name string, limit int) ([]model.MapMetaData, error)

	// 模糊查询（LIKE %name%）
	Search(ctx context.Context, name string, limit int) ([]model.MapMetaData, error)

	// 列举
	List(ctx context.Context, page int, desc bool, orderField string, limit int) ([]model.MapMetaData, error)

	// 删除
	Delete(ctx context.Context, hash string) error
}
