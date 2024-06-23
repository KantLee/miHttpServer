package middlewares

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// 自定义控制台日志输出格式
func CustomConsoleLogger(params gin.LogFormatterParams) string {
	return fmt.Sprintf(
		"[miHttpServer] %s |%s %d %s|  %s %s %s	 %s\n",
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
