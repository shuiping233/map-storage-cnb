package service

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	"map-storage-cnb/src/model"
	"map-storage-cnb/src/storage"

	"map-storage-cnb/src/utils"
)

const (
	MaxFormMem = 32 << 20 // 32 MB 内存表单阈值
)

type UploadAPI struct {
	Storage storage.Interface
}

func (u *UploadAPI) MapUploadApi(ctx *gin.Context) {
	// 1. 基础参数

	var request model.UploadFileRequest
	err := ctx.ShouldBind(&request)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, model.Fail("invalid params : "+err.Error()))
		return
	}

	hash := request.Sha256
	if hash != "" {
		exist, _ := u.Storage.Exists(ctx, hash)
		if exist {
			ctx.JSON(http.StatusConflict, model.Fail(request.Filename+" already uploaded"))
			return
		}

	}
	fileSize := request.File.Size
	filename := request.Filename
	if filename == "" {
		filename = request.File.Filename
	}
	file, err := request.File.Open()
	if err != nil {
		ctx.JSON(http.StatusBadRequest, model.Fail("open file failed : "+err.Error()))
		return
	}
	defer file.Close()

	fileData, err := io.ReadAll(file)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, model.Fail("read file failed : "+err.Error()))
		return
	}
	hash = utils.HashFile(fileData)

	meta, _ := u.Storage.GetMeta(ctx, hash)
	if meta != nil {
		ctx.JSON(http.StatusConflict, model.FailWithData(request.Filename+" already uploaded", meta))
		return
	}

	mapMetaData := model.NewMetaData(hash, filename)
	mapMetaData.Size = uint64(fileSize)

	_, err = u.Storage.Save(ctx, mapMetaData, fileData)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, model.Fail(err.Error()))
		return
	}

	ctx.JSON(http.StatusOK, model.OK(&model.UploadFileResponse{
		Sha256: hash,
		Size:   uint64(fileSize),
	}))
}
