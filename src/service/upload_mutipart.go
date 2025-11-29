package service

import "sync"

var (
	// 合并锁，防止并发同时写同一目标
	mergeLock sync.Map
)
