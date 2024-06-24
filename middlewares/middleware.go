package middlewares

import (
	"fmt"
	"os"
	"time"

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
			"[miHttpServer] %s |%d|  %s	 %s\n",
			params.TimeStamp.Format("2006-01-02 15:04:05"),
			params.StatusCode,
			params.Method,
			params.Path,
		))
	}
}
