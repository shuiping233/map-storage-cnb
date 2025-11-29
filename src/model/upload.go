package model

import (
	"mime/multipart"
)

type UploadFileRequest struct {
	File       *multipart.FileHeader `form:"file" binding:"required"`
	Filename   string                `form:"filename"`
	Sha256     string                `form:"sha256"`
}

type UploadFileResponse struct {
	Size   uint64 `json:"size"`
	Sha256 string `json:"sha256"`
}
