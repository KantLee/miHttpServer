package logger

import (
	"log"
	"os"
	"xorm.io/xorm"
	xormLog "xorm.io/xorm/log"
)

// 创建logs目录，保存日志文件
func MakeLogsFolder() {
	// 确保logs目录存在
	if err := os.MkdirAll("./logs", 0755); err != nil {
		log.Fatalf("创建logs目录失败: %v", err)
	}
}

// 设置全局的运行日志文件
func SetGlobalLogger() *os.File {
	logFile, err := os.OpenFile("./logs/miHttpServer.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatalf("无法创建全局日志文件: %v\n", err)
	}
	log.SetPrefix("[miHttpServer] ")
	log.SetOutput(logFile)
	return logFile
}

// 设置gin的日志文件
func SetGinLogger() *os.File {
	ginLogFile, err := os.Create("./logs/gin.log")
	if err != nil {
		log.Fatalf("无法创建Gin日志文件: %v\n", err)
	}
	return ginLogFile
}

// 设置mysql的日志文件
func SetMySQLLogger(engine *xorm.Engine) *os.File {
	file, err := os.Create("./logs/xorm.log")
	if err != nil {
		log.Fatalln("数据库日志文件创建失败:", err)
	} else {
		engine.SetLogger(xormLog.NewSimpleLogger(file))
	}
	// 日志相关设置
	engine.ShowSQL(true) // 开启SQL语句记录
	// 设置日志记录级别
	engine.Logger().SetLevel(xormLog.LOG_DEBUG)
	return file
}
