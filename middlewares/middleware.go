package middlewares

import (
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

type ResponseData struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// 自定义控制台日志输出格式
func CustomConsoleLogger(params gin.LogFormatterParams) string {
	return fmt.Sprintf(
		"[miHttpServerGin] %s |%s %d %s|  %s %s %s	 %s\n",
		params.TimeStamp.Format("2006-01-02 15:04:05"),
		params.StatusCodeColor(),
		params.StatusCode,
		params.ResetColor(),
		params.MethodColor(),
		params.Method,
		params.ResetColor(),
		params.Path,
	)
}

// 自定义文件日志输出格式
func CustomFileLogger(ginLogFile *os.File) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()
		params := gin.LogFormatterParams{
			TimeStamp:  time.Now(),
			StatusCode: ctx.Writer.Status(),
			Method:     ctx.Request.Method,
			Path:       ctx.Request.URL.Path,
		}
		// 记录日志到文件
		ginLogFile.WriteString(fmt.Sprintf(
			"[miHttpServerGin] %s |%d|  %s	 %s\n",
			params.TimeStamp.Format("2006-01-02 15:04:05"),
			params.StatusCode,
			params.Method,
			params.Path,
		))
	}
}

// 根据请求URL的app_local参数设置请求头
func SetAppLocal() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		response := ResponseData{}
		appLocal := ctx.Param("app_local")
		if appLocal == "" {
			// @warn: 错误处理组件
			response.Code = 1
			response.Msg = "请求参数app_local为空"
			response.Data = "缺少app_local参数，应为uk、jp和ru中的一个，例如http://localhost:8080/uk/item/"
			ctx.JSON(400, response)
			ctx.Abort()
			return
		} else if appLocal != "uk" && appLocal != "jp" && appLocal != "ru" {
			response.Code = 1
			response.Msg = "请求参数app_local非法"
			response.Data = "app_local参数应为uk、jp和ru中的一个，例如http://localhost:8080/uk/item/"
			ctx.JSON(400, response)
			ctx.Abort()
			return
		} else {
			ctx.Header("app_local", appLocal)
			ctx.Next()
		}
	}
}
