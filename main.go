package main

import (
	"encoding/json"
	"fmt"
	"log"
	"miHttpServer/databases"
	"miHttpServer/middlewares"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

/*
其实这个响应结构体不是必须的
因为map的迭代顺序是不确定的，可能会出现msg在data下面的情况
因此为了统一格式，定义一个响应结构体
*/
type ResponseData struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// 增加商品信息
func AddData(ctx *gin.Context) {
	data, _ := ctx.GetRawData()
	var response ResponseData
	var jsonStr map[string]interface{}
	err := json.Unmarshal(data, &jsonStr)
	// 这个只能判断json的key类型是否正确，不能判断value的类型是否正确
	if err != nil {
		response.Code = 1
		response.Msg = "客户端传递的json非法"
		response.Data = err.Error()
		ctx.JSON(http.StatusBadRequest, response)
		return
	}
	// 尝试获取分布式锁
	requestID := uuid.New().String()
	lockKey := fmt.Sprintf("item_lock_name_%s", jsonStr["name"].(string))
	locked, err := databases.Lock(lockKey, requestID, 30, 5*time.Second)
	if err != nil || !locked {
		response.Code = 1
		response.Msg = "无法获取锁"
		response.Data = "请稍后再试"
		ctx.JSON(http.StatusConflict, response)
		return
	}
	defer databases.Unlock(lockKey) // 确保最后释放锁

	// 如果json的value类型不正确，会导致panic
	defer func() {
		if err := recover(); err != nil {
			response.Code = 1
			response.Msg = "客户端传递的json值无效"
			response.Data = err.(error).Error()
			ctx.JSON(http.StatusInternalServerError, response)
		}
	}()
	item := databases.Item{
		Name:  jsonStr["name"].(string),
		Price: jsonStr["price"].(float64),
	}
	_, err = databases.InsertItem(&item)
	if err != nil {
		response.Code = 1
		response.Msg = "插入数据失败"
		response.Data = err.Error()
		ctx.JSON(http.StatusInternalServerError, response)
	} else {
		itemInfo := make(map[string]interface{})
		itemInfo["item_info"] = map[string]interface{}{
			"item_id": item.ItemID,
			"name":    item.Name,
			"price":   item.Price,
		}
		response.Code = 0
		response.Msg = "成功"
		response.Data = itemInfo
		ctx.JSON(http.StatusOK, response)
		log.Printf("增加商品，item_id: %d，name：%s", item.ItemID, item.Name)
	}
}

// 修改商品信息（如果缓存中也存在，需要同步更新）
func UpdateData(ctx *gin.Context) {
	itemIDStr := ctx.Param("item_id")
	var response ResponseData
	item_id, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		response.Code = 1
		response.Msg = "链接中的item_id非法"
		response.Data = err.Error()
		ctx.JSON(http.StatusBadRequest, response)
		return
	}

	data, _ := ctx.GetRawData()
	var jsonStr map[string]interface{}
	err = json.Unmarshal(data, &jsonStr)
	if err != nil {
		response.Code = 1
		response.Msg = "客户端传递的json非法"
		response.Data = err.Error()
		ctx.JSON(http.StatusBadRequest, response)
		return
	}
	// 尝试获取分布式锁
	requestID := uuid.New().String()
	lockKey := fmt.Sprintf("item_lock_id_%s", itemIDStr)
	locked, err := databases.Lock(lockKey, requestID, 30, 5*time.Second)
	if err != nil || !locked {
		response.Code = 1
		response.Msg = "无法获取锁"
		response.Data = "请稍后再试"
		ctx.JSON(http.StatusConflict, response)
		return
	}
	defer databases.Unlock(lockKey) // 确保最后释放锁

	defer func() {
		if err := recover(); err != nil {
			response.Code = 1
			response.Msg = "客户端传递的json值无效"
			response.Data = err.(error).Error()
			ctx.JSON(http.StatusInternalServerError, response)
		}
	}()
	item := databases.Item{
		ItemID: item_id,
		Name:   jsonStr["name"].(string),
		Price:  jsonStr["price"].(float64),
	}
	n, err := databases.UpdateItem(item_id, &item)
	if err != nil {
		response.Code = 1
		response.Msg = "更新数据失败，item_id：" + itemIDStr
		response.Data = err.Error()
		ctx.JSON(http.StatusInternalServerError, response)
	} else if n == 0 {
		response.Code = 1
		response.Msg = "未找到相关记录"
		response.Data = fmt.Sprintf("item_id为%v的商品不存在", item_id)
		ctx.JSON(http.StatusInternalServerError, response)
	} else {
		storeInfo := make(map[string]interface{})
		storeInfo["store_info"] = map[string]interface{}{
			"item_id": item.ItemID,
			"name":    item.Name,
			"price":   item.Price,
		}
		response.Code = 0
		response.Msg = "成功"
		response.Data = storeInfo
		ctx.JSON(http.StatusOK, response)
		log.Printf("修改商品，item_id: %d，name：%s", item.ItemID, item.Name)

		// 查找redis缓存中是否有相同数据
		itemCache := databases.ItemCache{}
		ok, err := databases.QueryItemCache(item_id, &itemCache)
		if err != nil {
			log.Printf("查询商品%d的Redis缓存失败: %s", item_id, err.Error())
		} else {
			if ok {
				// 如果缓存中有相同数据，更新缓存
				itemCache.Name = item.Name
				itemCache.Price = item.Price
				err = databases.AddItemCache(item.ItemID, itemCache)
				if err != nil {
					log.Printf("更新商品%d的Redis缓存失败: %s", item_id, err.Error())
				}
			}
		}
	}
}

// 查询商品信息（先查询缓存，未命中再查询MySQL）
func QueryData(ctx *gin.Context) {
	itemIDStr := ctx.Param("item_id")
	var response ResponseData
	item_id, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		response.Code = 1
		response.Msg = "链接中的item_id非法"
		response.Data = err.Error()
		ctx.JSON(http.StatusBadRequest, response)
		return
	}

	// 从redis缓存中查询数据
	itemCache := databases.ItemCache{}
	ok, err := databases.QueryItemCache(item_id, &itemCache)
	if err != nil {
		log.Printf("查询商品%d的Redis缓存失败: %s", item_id, err.Error())
	} else {
		if ok {
			storeInfo := make(map[string]interface{})
			storeInfo["store_info"] = map[string]interface{}{
				"item_id": itemCache.ItemID,
				"name":    itemCache.Name,
				"price":   itemCache.Price,
			}
			response.Code = 0
			response.Msg = "成功"
			response.Data = storeInfo
			ctx.JSON(http.StatusOK, response)
			return
		}
	}

	// 从MySQL查询数据
	item := databases.Item{}
	success, err := databases.QueryItem(item_id, &item)
	if err != nil {
		response.Code = 1
		response.Msg = "查询数据失败"
		response.Data = err.Error()
		ctx.JSON(http.StatusInternalServerError, response)
	} else if !success {
		response.Code = 1
		response.Msg = "未找到相关记录"
		response.Data = fmt.Sprintf("item_id为%v的商品不存在", item_id)
		ctx.JSON(http.StatusInternalServerError, response)
	} else {
		storeInfo := make(map[string]interface{})
		storeInfo["store_info"] = map[string]interface{}{
			"item_id": item.ItemID,
			"name":    item.Name,
			"price":   item.Price,
		}
		response.Code = 0
		response.Msg = "成功"
		response.Data = storeInfo
		ctx.JSON(http.StatusOK, response)

		// 将数据存入Redis缓存
		itemCache.ItemID = item.ItemID
		itemCache.Name = item.Name
		itemCache.Price = item.Price
		err = databases.AddItemCache(item.ItemID, itemCache)
		if err != nil {
			log.Printf("增加商品%d的Redis缓存失败: %s", item_id, err.Error())
		}
	}
}

// 删除商品信息（如果缓存中也存在，需要同步删除）
func DeleteData(ctx *gin.Context) {
	// 获取当前的UTC时间
	utcNow := time.Now().UTC()
	var location *time.Location
	var err error
	appLocal := ctx.Param("app_local")
	var localCountry string
	switch appLocal {
	case "uk":
		location, _ = time.LoadLocation("Europe/London")
		localCountry = "英国"
	case "jp":
		location, _ = time.LoadLocation("Asia/Tokyo")
		localCountry = "日本"
	case "ru":
		location, _ = time.LoadLocation("Europe/Moscow")
		localCountry = "俄罗斯"
	}
	// 将 UTC 时间转换为当地时间
	localTime := utcNow.In(location)
	// 格式化时间
	formattedTime := localTime.Format("2006-01-02 15:04:05")

	itemIDStr := ctx.Param("item_id")
	var response ResponseData
	item_id, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		response.Code = 1
		response.Msg = "链接中的item_id非法"
		response.Data = err.Error()
		ctx.JSON(http.StatusBadRequest, response)
		return
	}

	n, err := databases.DeleteItem(item_id)
	if err != nil {
		response.Code = 1
		response.Msg = "查询数据失败"
		response.Data = err.Error()
		ctx.JSON(http.StatusInternalServerError, response)
	} else if n == 0 {
		response.Code = 1
		response.Msg = "未找到相关记录"
		response.Data = fmt.Sprintf("item_id为%v的商品不存在", item_id)
		ctx.JSON(http.StatusInternalServerError, response)
	} else {
		deleteTime := make(map[string]string, 1)
		deleteTime["delete_time"] = formattedTime
		response.Code = 0
		response.Msg = "成功"
		response.Data = deleteTime
		ctx.JSON(http.StatusOK, response)
		log.Printf("%s站点删除商品，item_id: %d，当地时间：%s", localCountry, item_id, formattedTime)

		// 查找redis缓存中是否有相同数据
		itemCache := databases.ItemCache{}
		ok, err := databases.QueryItemCache(item_id, &itemCache)
		if err != nil {
			log.Printf("查询商品%d的Redis缓存失败: %s", item_id, err.Error())
		} else {
			if ok {
				// 如果Redis缓存中有相同数据，删除缓存
				err = databases.DeleteItemCache(item_id)
				if err != nil {
					log.Printf("删除商品%d的Redis缓存失败: %s", item_id, err.Error())
				}
			}
		}
	}
}

func main() {
	// 确保logs目录存在
	if err := os.MkdirAll("./logs", 0755); err != nil {
		log.Fatalf("创建logs目录失败: %v", err)
	}
	// 设置gin的运行模式，ReleaseMode表示生产模式，不显示日志的调试信息
	gin.SetMode(gin.ReleaseMode)
	// 设置全局的运行日志文件
	logFile, err := os.OpenFile("./logs/miHttpServer.log", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()
	log.SetPrefix("[miHttpServer] ")
	log.SetOutput(logFile)

	// 设置Gin日志文件
	ginLogFile, err := os.Create("./logs/gin.log")
	if err != nil {
		log.Printf("无法创建Gin日志文件: %v\n", err)
	}
	defer ginLogFile.Close()

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
	err = databases.InitMySQL()
	if err != nil {
		log.Fatalln("连接数据库失败:", err)
	}
	// 关闭MySQL连接
	defer databases.CloseMySQL()

	// 连接redis
	databases.InitRedis()
	// 关闭redis连接
	defer databases.CloseRedis()

	// 增加商品信息（从JSON获取）
	ginServer.PUT("/:app_local/item", AddData)
	ginServer.POST("/:app_local/item", AddData)

	// 修改商品信息
	ginServer.POST("/:app_local/item/:item_id", UpdateData)

	// 查询商品信息
	ginServer.GET("/:app_local/item/:item_id", QueryData)

	// 删除商品信息
	ginServer.DELETE("/:app_local/item/:item_id", DeleteData)

	// 未匹配到任何路由的请求
	ginServer.NoRoute(func(ctx *gin.Context) {
		response := ResponseData{}
		response.Code = 1
		response.Msg = "未找到相关路由"
		response.Data = "请检查请求的URL是否正确"
		ctx.JSON(http.StatusNotFound, response)
	})

	err = ginServer.Run(":8080")
	if err != nil {
		log.Fatal("项目启动失败:", err)
	}
}
