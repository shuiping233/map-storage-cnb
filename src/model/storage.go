package model

import "time"

type MapStorageStatus uint
type StorageType string

const (
	MapUploadStatusSuccess MapStorageStatus = iota
	MapUploadStatusOnProgress
	MapUploadStatusFailed
)

const (
	MapUploadStatusMsgSuccess    = "success"
	MapUploadStatusMsgOnProgress = "on progress"
	MapUploadStatusMsgUnknown    = "unknown failed"
)

// 地图元数据,理论上都可以从地图本身计算和获取的到
type MapMetaData struct {
	Hash             string `gorm:"primaryKey"` // SHA-256 hex
	Name             string
	Size             uint64
	CreateTime       int64  // UnixNano，方便列举排序
	PrevHash         string // 指向上一个版本，首版留空
	Message          string // 提交备注
	Authors          string
	StorageType      StorageType
	StorageStatus    MapStorageStatus
	StorageStatusMsg string
}

type Option func(MapMetaData)

func NewMetaData(hash string, name string, opts ...Option) MapMetaData {
	m := MapMetaData{
		Hash:             hash,
		Name:             name,
		Size:             0,
		CreateTime:       time.Now().UnixNano(),
		Message:          "",
		Authors:          "",
		PrevHash:         "",
		StorageType:      "",
		StorageStatus:    MapUploadStatusFailed,
		StorageStatusMsg: MapUploadStatusMsgUnknown,
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func (m *MapMetaData) SetStorageType(storageType StorageType) {
	m.StorageType = storageType
}

// reason传空字符串则默认设置存储失败原因为"Unknown,失败状态才会使用传入的reason"
func (m *MapMetaData) SetStorageStatus(statusType MapStorageStatus, reason string) {
	if reason == "" {
		reason = MapUploadStatusMsgUnknown
	}
	switch statusType {
	case MapUploadStatusSuccess:
		m.StorageStatus = MapUploadStatusSuccess
		m.StorageStatusMsg = MapUploadStatusMsgSuccess

	case MapUploadStatusOnProgress:
		m.StorageStatus = MapUploadStatusOnProgress
		m.StorageStatusMsg = MapUploadStatusMsgOnProgress

	case MapUploadStatusFailed:
		m.StorageStatus = MapUploadStatusFailed
		m.StorageStatusMsg = reason

	}

}

// TODO 查询用
type MapMetaDataSearch struct {
	Hash      *string
	Name      *string
	MinSize   *uint64
	MaxSize   *uint64
	StartTime *int64 // 时间区间
	EndTime   *int64
	PrevHash  *string
	Message   *string
	Authors   *string // 英文半角逗号做分隔符,不用字符串列表是因为sqlite不支持
	OrderDesc bool
	Limit     uint
}
