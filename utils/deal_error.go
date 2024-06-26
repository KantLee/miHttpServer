package utils

import (
	"miHttpServer/config"
	"miHttpServer/models"
)

// 处理请求错误
func DealRequestError(msg string, err error) models.ResponseData {
	var response models.ResponseData
	response.Code = config.Configs.Code.RequestError
	response.Msg = msg
	response.Data = err.Error()
	return response
}

// 处理服务器内部处理错误
func DealServerError(msg string, err error) models.ResponseData {
	var response models.ResponseData
	response.Code = config.Configs.Code.ServerError
	response.Msg = msg
	response.Data = err.Error()
	return response
}
