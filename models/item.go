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

var engine *xorm.Engine
var session *xorm.Session

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
	// 构建数据库的连接信息
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s", userName, password, ipAddress, port, dbName, charset)
	var err error
	engine, err = xorm.NewEngine("mysql", dataSourceName)
	if err != nil {
		return err
	}

	// 设置日志记录的文件
	f, err := os.Create("./logs/xorm.log")
	if err != nil {
		fmt.Println("数据库日志文件创建失败:", err)
	} else {
		engine.SetLogger(log.NewSimpleLogger(f))
	}
	// 日志相关设置
	engine.ShowSQL(true) // 开启SQL语句记录
	// 设置日志记录级别
	engine.Logger().SetLevel(log.LOG_DEBUG)

	// 同步表结构
	err = engine.Sync2(new(Item))
	if err != nil {
		return err
	}

	session = engine.NewSession()

	return nil
}

// 插入数据
func InsertItem(item *Item) (int64, error) {
	session.Begin()
	defer session.Close()
	defer func() {
		if err := recover(); err != nil {
			session.Rollback()
		} else {
			session.Commit()
		}
	}()
	n, err := session.Insert(item)
	if err != nil {
		panic(err)
	}
	return n, err
}

// 更新数据
func UpdateItem(item_id int64, item *Item) (int64, error) {
	session.Begin()
	defer session.Close()
	defer func() {
		if err := recover(); err != nil {
			session.Rollback()
		} else {
			session.Commit()
		}
	}()
	n, err := session.Where("item_id = ?", item_id).Update(item)
	if err != nil {
		panic(err)
	}
	return n, err
}

// 根据item_id查询数据，查询就不使用事务了
func QueryItem(item_id int64, item *Item) (bool, error) {
	success, err := engine.Where("item_id = ?", item_id).Get(item)
	return success, err
}

// 根据item_id删除数据
func DeleteItem(item_id int64) (int64, error) {
	session.Begin()
	defer session.Close()
	defer func() {
		if err := recover(); err != nil {
			session.Rollback()
		} else {
			session.Commit()
		}
	}()
	n, err := session.ID(item_id).Delete(&Item{})
	if err != nil {
		panic(err)
	}
	return n, err
}
