package database

import (
	"fmt"
	"log"
	"miHttpServer/config"
	"miHttpServer/logger"
	"miHttpServer/models"

	_ "github.com/go-sql-driver/mysql"
	"xorm.io/xorm"
)

var Engine *xorm.Engine

func InitMySQL() error {
	// 数据库连接基本信息
	var (
		userName  string = config.Configs.MySQL.Username
		password  string = config.Configs.MySQL.Password
		ipAddress string = config.Configs.MySQL.Host
		port      int    = config.Configs.MySQL.Port
		dbName    string = config.Configs.MySQL.Database
		charset   string = config.Configs.MySQL.Charset
	)
	// 构建数据库的连接信息
	dataSourceName := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s", userName, password, ipAddress, port, dbName, charset)
	var err error
	Engine, err = xorm.NewEngine("mysql", dataSourceName)
	if err != nil {
		return err
	}

	// 设置日志
	xormLogFile := logger.SetMySQLLogger(Engine)
	defer xormLogFile.Close()

	// 同步表结构
	err = Engine.Sync2(new(models.Item))
	if err != nil {
		return err
	}

	return nil
}

// 关闭MySQL连接
func CloseMySQL() {
	if Engine != nil {
		Engine.Close()
	}
}

// 插入数据
func InsertItem(item *models.Item) (int64, error) {
	// 单条插入数据，一般不需要使用事务
	n, err := Engine.Insert(item)
	if err != nil {
		log.Println("插入失败:", err)
	}
	return n, err
}

// 更新数据
func UpdateItem(item_id int64, item *models.Item) (int64, error) {
	n, err := Engine.Where("item_id = ?", item_id).Update(item)
	if err != nil {
		log.Println("更新数据失败:", err)
	}
	return n, err
}

// 根据item_id查询数据
func QueryItem(item_id int64, item *models.Item) (bool, error) {
	success, err := Engine.Where("item_id = ?", item_id).Get(item)
	return success, err
}

// 根据item_id删除数据
func DeleteItem(item_id int64) (int64, error) {
	session := Engine.NewSession()
	defer session.Close()
	session.Begin()
	defer func() {
		if err := recover(); err != nil {
			session.Rollback()
		} else {
			session.Commit()
		}
	}()
	n, err := session.ID(item_id).Delete(&models.Item{})
	if err != nil {
		panic(err)
	}
	return n, err
}
