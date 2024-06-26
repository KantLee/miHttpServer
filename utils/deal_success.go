package utils

import (
	"miHttpServer/config"
	"miHttpServer/models"
)

// 处理增加/查询/修改/删除请求成功
func DealSuccess(msg string, data map[string]interface{}) models.ResponseData {
	var response models.ResponseData
	response.Code = config.Configs.Code.Success
	response.Msg = msg
	response.Data = data
	return response
}
