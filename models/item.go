package models

import (
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"xorm.io/xorm"
	"xorm.io/xorm/log"
)

// 定义一个和表同步的结构体
type Item struct {
	ItemID    int64     `xorm:"'item_id' pk autoincr" json:"item_id"`
	Name      string    `xorm:"varchar(255)" json:"name"`
	Price     float64   `xorm:"decimal(10,2)" json:"price"`
	CreatedAt time.Time `xorm:"created" json:"created_at"`
	UpdatedAt time.Time `xorm:"updated" json:"updated_at"`
}

var Engine *xorm.Engine

func InitDB() error {
	// 数据库连接基本信息
	var (
		userName  string = "root"
		password  string = "admin"
		ipAddress string = "192.168.96.130"
		port      int    = 3306
		dbName    string = "miHttpServer"
		charset   string = "utf8mb4"
	)
	var err error
	// 构建数据库的连接信息
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s", userName, password, ipAddress, port, dbName, charset)
	Engine, err = xorm.NewEngine("mysql", dataSourceName)
	if err != nil {
		return err
	}
	err = Engine.Sync2(new(Item))
	if err != nil {
		return err
	}
	// 日志相关设置
	Engine.ShowSQL(true) // 开启SQL语句记录
	// 设置日志记录级别
	Engine.Logger().SetLevel(log.LOG_DEBUG)
	// 设置日志记录的文件
	f, err := os.Create("./logs/xorm.log")
	if err != nil {
		fmt.Println("数据库日志文件创建失败:", err)
	} else {
		Engine.SetLogger(log.NewSimpleLogger(f))
	}
	return nil
}

func InsertItem(item *Item) (int64, error) {
	session := Engine.NewSession()
	defer session.Close()
	defer func() {
		if err := recover(); err != nil {
			session.Rollback()
		} else {
			session.Commit()
		}
	}()
	return Engine.Insert(item)
}