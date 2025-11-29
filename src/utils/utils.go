package utils

import (
	"encoding/hex"
	"os"
	"time"

	"crypto/sha256"
)

func InitDefaultDir() {
	err := os.MkdirAll("./tmp", 0755)
	if err != nil {
		panic(err)
	}
	err = os.MkdirAll("./uploads", 0755)
	if err != nil {
		panic(err)
	}
}

// Hash bytes in sha256
func HashFile(data []byte) string {
	hash := sha256.Sum256(data) // 返回 [32]byte
	return hex.EncodeToString(hash[:])
}

// "2006/01/02"
func TimeNowDir() string {
	return time.Now().Format("2006/01/02")
}

// Get file size
func GetFileSize(path string) int64 {
	fi, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return fi.Size()
}
