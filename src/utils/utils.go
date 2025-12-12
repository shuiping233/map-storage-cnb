package utils

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
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
func TimeNowDay() string {
	return time.Now().Format("2006/01/02")
}

// "2006-01-02 15:04:05"
func TimeNowFull() string {
	return time.Now().Format("2006-01-02 15:04:05")
}

// "2006-01-02T15:04:05Z07:00"
func ISO8601LocalNow() string {
	return time.Now().Format("2006-01-02T15:04:05Z07:00")
}

// "2006-01-02T15:04:05Z07:00"
func ISO8601UTCNow() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05Z07:00")
}

// Get file size
func GetFileSize(path string) int64 {
	fi, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return fi.Size()
}

func DirExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return info.IsDir()
}

func FileExists(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		return false
	}
	return !info.IsDir()
}

// AddSuffixIfMissing 给文件名插入一段后缀。
// 如果已经有预期的后缀,则直接不处理
// 否则直接拼后缀
func AddSuffixIfMissing(name, suffix string) string {
	if !strings.HasPrefix(suffix, ".") {
		suffix = "." + suffix
	}
	ext := filepath.Ext(name)
	if ext == "" {
		return name + suffix
	}
	if ext == suffix {
		return name
	}
	return name + suffix

}
