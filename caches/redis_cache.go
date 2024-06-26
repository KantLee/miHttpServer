package caches

import (
	"miHttpServer/database"
	"miHttpServer/models"
)

// 更新reids缓存
func UpdateRedisCache(item_id int64, item models.Item) error {
	// 查找redis缓存中是否有相同数据
	itemCache := models.ItemCache{}
	ok, err := database.QueryItemCache(item_id, &itemCache)
	if err != nil {
		return err
	}
	if ok {
		// 如果缓存中有相同数据，更新缓存
		itemCache.Name = item.Name
		itemCache.Price = item.Price
		err = database.AddItemCache(item.ItemID, itemCache)
		if err != nil {
			return err
		}
	}
	return nil
}

// 查询redis缓存
func QueryRedisCache(item_id int64) (error, map[string]interface{}) {
	itemCache := models.ItemCache{}
	ok, err := database.QueryItemCache(item_id, &itemCache)
	if err != nil {
		return err, nil
	}
	if ok {
		storeInfo := make(map[string]interface{})
		storeInfo["store_info"] = map[string]interface{}{
			"item_id": itemCache.ItemID,
			"name":    itemCache.Name,
			"price":   itemCache.Price,
		}
		return nil, storeInfo
	}
	return nil, nil
}

// 新增redis缓存
func AddRedisCache(item_id int64, item models.Item) error {
	var itemCache models.ItemCache
	// 将数据存入Redis缓存
	itemCache.ItemID = item.ItemID
	itemCache.Name = item.Name
	itemCache.Price = item.Price
	err := database.AddItemCache(item.ItemID, itemCache)
	return err
}

// 删除redis缓存
func DeleteItemCache(item_id int64) error {
	itemCache := models.ItemCache{}
	ok, err := database.QueryItemCache(item_id, &itemCache)
	if err != nil {
		return err
	}
	if ok {
		// 如果Redis缓存中有相同数据，删除缓存
		err = database.DeleteItemCache(item_id)
		if err != nil {
			return err
		}
	}
	return nil
}
