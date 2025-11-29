package utils

import (
	"encoding/hex"
	"io"
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

// Hash file in sha256
func HashFile(file io.Reader) (string, error) {
	hash := sha256.New()
	_, err := io.Copy(hash, file)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
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
