package main

import (
	"errors"
	"fmt"
	"log"
	"miHttpServer/config"
	"miHttpServer/database"
	"miHttpServer/handlers"
	"miHttpServer/logger"
	"miHttpServer/middlewares"
	"miHttpServer/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

func main() {
	// 创建logs目录，保存日志文件
	logger.MakeLogsFolder()

	// 设置全局的日志文件
	globalLogFile := logger.SetGlobalLogger()
	defer globalLogFile.Close()

	// 设置gin的运行模式，ReleaseMode表示生产模式，不显示日志的调试信息
	gin.SetMode(gin.ReleaseMode)

	// 设置Gin日志文件
	ginLogFile := logger.SetGinLogger()
	defer ginLogFile.Close()

	// 解析配置文件
	utils.ParseYaml()

	// 创建一个服务（不使用默认的中间件）
	ginServer := gin.New()
	// 使用自定义的控制台日志格式
	ginServer.Use(gin.LoggerWithFormatter(middlewares.CustomConsoleLogger))
	// 使用自定义的文件日志格式
	ginServer.Use(middlewares.CustomFileLogger(ginLogFile))
	// 防止服务器产生panic而崩溃，同时返回一个500的HTTP状态码
	ginServer.Use(gin.Recovery())
	// 请求头根据url添加app_local参数
	ginServer.Use(middlewares.SetAppLocal())
	// 连接MySQL
	err := database.InitMySQL()
	if err != nil {
		log.Fatalln("连接数据库失败:", err)
	}
	// 关闭MySQL连接
	defer database.CloseMySQL()

	// 连接redis
	database.InitRedis()
	// 关闭redis连接
	defer database.CloseRedis()

	// 增加商品信息（从JSON获取）
	ginServer.PUT("/:app_local/item", handlers.AddItem)
	ginServer.POST("/:app_local/item", handlers.AddItem)

	// 修改商品信息
	ginServer.POST("/:app_local/item/:item_id", handlers.UpdateItem)

	// 查询商品信息
	ginServer.GET("/:app_local/item/:item_id", handlers.QueryItem)

	// 删除商品信息
	ginServer.DELETE("/:app_local/item/:item_id", handlers.DeleteItem)

	// 未匹配到任何路由的请求
	ginServer.NoRoute(func(ctx *gin.Context) {
		response := utils.DealRequestError(
			"未找到相关路由",
			errors.New("请检查请求的URL是否正确"),
		)
		ctx.JSON(http.StatusNotFound, response)
	})

	port := fmt.Sprintf(":%d", config.Configs.Server.Port)
	err = ginServer.Run(port)
	if err != nil {
		log.Fatal("项目启动失败:", err)
	}
}
