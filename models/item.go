package models

import "time"

// 和MySQL表同步的结构体
type Item struct {
	ItemID    int64     `xorm:"'item_id' pk autoincr" json:"item_id"`
	Name      string    `xorm:"varchar(255)" json:"name"`
	Price     float64   `xorm:"decimal(10,2)" json:"price"`
	CreatedAt time.Time `xorm:"created" json:"created_at"`
	UpdatedAt time.Time `xorm:"updated" json:"updated_at"`
}

// Redis缓存的结构体
type ItemCache struct {
	ItemID int64   `json:"item_id"`
	Name   string  `json:"name"`
	Price  float64 `json:"price"`
}

// 响应数据结构体
// 其实这个响应结构体不是必须的
// 因为map的迭代顺序是不确定的，可能会出现msg在data下面的情况
// 因此为了统一格式，定义一个响应结构体
type ResponseData struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// 增加和更新商品信息时请求的结构体
type RequestData struct {
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

// 删除商品时保存item_id和删除时间
var ItemDeleteTime = make(map[int64]string, 20)
