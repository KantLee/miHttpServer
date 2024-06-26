package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"miHttpServer/caches"
	"miHttpServer/database"
	"miHttpServer/models"
	"miHttpServer/utils"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// 增加商品信息
func AddItem(ctx *gin.Context) {
	data, err := ctx.GetRawData()
	var response models.ResponseData
	if err != nil {
		response = utils.DealRequestError(
			"客户端传递的json非法",
			err,
		)
		ctx.JSON(http.StatusBadRequest, response)
		return
	}
	var requestStr models.RequestData
	err = json.Unmarshal(data, &requestStr)
	if err != nil {
		response = utils.DealRequestError(
			"客户端传递的json非法",
			err,
		)
		ctx.JSON(http.StatusBadRequest, response)
		return
	}
	item := models.Item{
		Name:  requestStr.Name,
		Price: requestStr.Price,
	}

	// 尝试获取分布式锁
	id := uuid.New().String()
	lockKey := fmt.Sprintf("item_lock_name_%s", requestStr.Name)
	ok, err := utils.GetLock(lockKey, id)
	if err != nil {
		response = utils.DealServerError("获取分布式锁错误", err)
		ctx.JSON(http.StatusInternalServerError, response)
		return
	}
	if !ok {
		response = utils.DealServerError("获取分布式锁超时", errors.New("请稍后再试"))
		ctx.JSON(http.StatusConflict, response)
		return
	}
	// 确保最后释放锁
	defer utils.ReleaseLock(lockKey, id)

	_, err = database.InsertItem(&item)
	if err != nil {
		response = utils.DealServerError("插入数据失败", err)
		ctx.JSON(http.StatusInternalServerError, response)
		return
	}

	itemInfo := make(map[string]interface{})
	itemInfo["item_info"] = map[string]interface{}{
		"item_id": item.ItemID,
		"name":    item.Name,
		"price":   item.Price,
	}
	response = utils.DealSuccess("成功", itemInfo)
	ctx.JSON(http.StatusOK, response)
	log.Printf("增加商品，item_id: %d，name：%s", item.ItemID, item.Name)
}

// 修改商品信息（如果缓存中也存在，需要同步更新）
func UpdateItem(ctx *gin.Context) {
	itemIDStr := ctx.Param("item_id")
	var response models.ResponseData
	item_id, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		response = utils.DealRequestError("item_id非法", err)
		ctx.JSON(http.StatusBadRequest, response)
		return
	}

	data, err := ctx.GetRawData()
	if err != nil {
		response = utils.DealRequestError(
			"客户端传递的json非法",
			err,
		)
		ctx.JSON(http.StatusBadRequest, response)
		return
	}

	var requestStr models.RequestData
	err = json.Unmarshal(data, &requestStr)
	if err != nil {
		response = utils.DealRequestError(
			"客户端传递的json非法",
			err,
		)
		ctx.JSON(http.StatusBadRequest, response)
		return
	}
	item := models.Item{
		ItemID: item_id,
		Name:   requestStr.Name,
		Price:  requestStr.Price,
	}

	// 尝试获取分布式锁
	id := uuid.New().String()
	lockKey := fmt.Sprintf("item_lock_name_%s", requestStr.Name)
	ok, err := utils.GetLock(lockKey, id)
	if err != nil {
		response := utils.DealServerError("获取分布式锁错误", err)
		ctx.JSON(http.StatusInternalServerError, response)
		return
	}
	if !ok {
		response := utils.DealServerError("获取分布式锁超时", errors.New("请稍后再试"))
		ctx.JSON(http.StatusConflict, response)
		return
	}
	// 确保最后释放锁
	defer utils.ReleaseLock(lockKey, id)

	n, err := database.UpdateItem(item_id, &item)
	if err != nil {
		response = utils.DealServerError("更新数据失败", err)
		ctx.JSON(http.StatusInternalServerError, response)
		return
	}
	if n == 0 {
		response = utils.DealServerError(
			"未找到相关记录",
			fmt.Errorf("item_id为%v的商品不存在", item_id),
		)
		ctx.JSON(http.StatusInternalServerError, response)
		return
	}
	storeInfo := make(map[string]interface{})
	storeInfo["store_info"] = map[string]interface{}{
		"item_id": item.ItemID,
		"name":    item.Name,
		"price":   item.Price,
	}

	response = utils.DealSuccess("成功", storeInfo)
	ctx.JSON(http.StatusOK, response)
	log.Printf("修改商品，item_id: %d，name：%s", item.ItemID, item.Name)

	// 查找redis缓存中是否有相同数据，有则更新redis缓存
	err = caches.UpdateRedisCache(item_id, item)
	if err != nil {
		log.Printf("查询商品%d的Redis缓存失败: %s", item_id, err.Error())
	}
}

// 查询商品信息（先查询缓存，未命中再查询MySQL）
func QueryItem(ctx *gin.Context) {
	itemIDStr := ctx.Param("item_id")
	var response models.ResponseData
	item_id, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		response = utils.DealRequestError("item_id非法", err)
		ctx.JSON(http.StatusBadRequest, response)
		return
	}

	// 从redis缓存中查询数据
	err, storeInfo := caches.QueryRedisCache(item_id)
	if err != nil {
		log.Printf("查询商品%d的Redis缓存失败: %s", item_id, err.Error())
	} else {
		if storeInfo != nil {
			// 说明缓存中有数据
			response = utils.DealSuccess("成功", storeInfo)
			ctx.JSON(http.StatusOK, response)
			return
		}
	}

	// 从MySQL查询数据
	item := models.Item{}
	success, err := database.QueryItem(item_id, &item)
	if err != nil {
		response = utils.DealServerError("查询数据失败", err)
		ctx.JSON(http.StatusInternalServerError, response)
		return
	}
	if !success {
		response = utils.DealServerError(
			"未找到相关记录",
			fmt.Errorf("item_id为%v的商品不存在", item_id),
		)
		ctx.JSON(http.StatusInternalServerError, response)
		return
	}

	storeInfo = make(map[string]interface{})
	storeInfo["store_info"] = map[string]interface{}{
		"item_id": item.ItemID,
		"name":    item.Name,
		"price":   item.Price,
	}
	response = utils.DealSuccess("成功", storeInfo)
	ctx.JSON(http.StatusOK, response)

	// 将数据存入Redis缓存
	err = caches.AddRedisCache(item_id, item)
	if err != nil {
		log.Printf("增加商品%d的Redis缓存失败: %s", item_id, err.Error())
	}
}

// 删除商品信息（如果缓存中也存在，需要同步删除）
func DeleteItem(ctx *gin.Context) {
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
	var response models.ResponseData
	item_id, err := strconv.ParseInt(itemIDStr, 10, 64)
	if err != nil {
		response = utils.DealRequestError("item_id非法", err)
		ctx.JSON(http.StatusBadRequest, response)
		return
	}

	n, err := database.DeleteItem(item_id)
	if err != nil {
		response = utils.DealServerError("删除数据失败", err)
		ctx.JSON(http.StatusInternalServerError, response)
		return
	}
	if n == 0 {
		response = utils.DealServerError(
			"未找到相关记录",
			fmt.Errorf("item_id为%v的商品不存在", item_id),
		)
		ctx.JSON(http.StatusInternalServerError, response)
		return
	}

	deleteTime := make(map[string]interface{}, 1)
	deleteTime["delete_time"] = formattedTime
	response = utils.DealSuccess("成功", deleteTime)
	ctx.JSON(http.StatusOK, response)
	log.Printf("%s站点删除商品，item_id: %d，当地时间：%s", localCountry, item_id, formattedTime)

	// 查找redis缓存中是否有相同数据，有则删除
	err = caches.DeleteItemCache(item_id)
	if err != nil {
		log.Printf("删除商品%d的Redis缓存失败: %s", item_id, err.Error())
	}
}
