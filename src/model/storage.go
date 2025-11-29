package model

import "time"

// 地图元数据,理论上都可以从地图本身计算和获取的到
type MapMetaData struct {
	Hash       string `gorm:"primaryKey"` // SHA-256 hex
	Name       string
	Size       uint64
	CreateTime int64  // UnixNano，方便列举排序
	PrevHash   string // 指向上一个版本，首版留空
	Message    string // 提交备注
	Authors    string
}

// TODO 查询用
type MapMetaDataSearch struct {
	Hash      *string // 指针，能区分“没传”和“传了空串”
	Name      *string
	MinSize   *uint64
	MaxSize   *uint64
	StartTime *int64 // 时间区间
	EndTime   *int64
	PrevHash  *string
	Message   *string
	Authors   *string // 切片，nil 表示不限
	OrderDesc bool
	Limit     uint
}

type Option func(MapMetaData)

func NewMetaData(hash string, name string, opts ...Option) MapMetaData {
	m := MapMetaData{
		Hash:       hash,
		Name:       name,
		Size:       0,
		CreateTime: time.Now().UnixNano(),
		Message:    "",
		Authors:    "",
		PrevHash:   "",
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}
