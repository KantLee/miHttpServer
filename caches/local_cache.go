package caches

import (
	"miHttpServer/config"
	"miHttpServer/models"
	"sync"
	"time"
)

type Node struct {
	key   int64
	value models.ItemCache
	pre   *Node
	next  *Node
	// 过期时间
	expireAt time.Time
}

type LRUCache struct {
	// 缓存的最大容量
	capacity int
	// 用于快速查找节点
	cache map[int64]*Node
	// 指向双向链表的头节点
	head *Node
	// 指向双向链表的尾节点
	end *Node
	// 互斥锁
	mutex sync.Mutex
}

// 本地缓存
var LocalCache *LRUCache

// 构造函数，初始化LRUCache
func NewLRUCache(capacity int) *LRUCache {
	return &LRUCache{
		capacity: config.Configs.LocalCache.Capacity,
		cache:    make(map[int64]*Node, capacity),
	}
}

// 获取数据，返回额外的布尔值表示是否找到
func (lruCache *LRUCache) Get(key int64) (models.ItemCache, bool) {
	lruCache.mutex.Lock()
	defer lruCache.mutex.Unlock()

	if node, ok := lruCache.cache[key]; ok {
		// 如果找到了节点，则判断是否过期
		if node.expireAt.Before(time.Now()) {
			// 如果过期了，则删除节点
			removeKey := lruCache.removeNode(node)
			delete(lruCache.cache, removeKey)
			return models.ItemCache{}, false
		}
		// 如果没有过期，则将节点移动到链表尾部
		lruCache.moveNodeToEnd(node)
		return node.value, true
	}
	return models.ItemCache{}, false
}

// 添加数据
func (lruCache *LRUCache) Put(key int64, value models.ItemCache) {
	lruCache.mutex.Lock()
	defer lruCache.mutex.Unlock()
	duration := time.Duration(config.Configs.LocalCache.ExpireSec) * time.Second
	expireAt := time.Now().Add(duration)
	if node, ok := lruCache.cache[key]; ok {
		// 如果key已经存在，则更新value和过期时间
		node.value = value
		node.expireAt = expireAt
		// 将节点移动到链表尾部
		lruCache.moveNodeToEnd(node)
	} else {
		// 如果key不存在，则判断容量是否已满，如果已满则删除头节点
		if len(lruCache.cache) >= lruCache.capacity {
			removeKey := lruCache.removeNode(lruCache.head)
			delete(lruCache.cache, removeKey)
		}
		// 添加新节点到链表尾部
		node := &Node{
			key:      key,
			value:    value,
			expireAt: expireAt,
		}
		lruCache.addNode(node)
		lruCache.cache[key] = node
	}
}

// 移动节点到双向链表尾部
func (lruCache *LRUCache) moveNodeToEnd(node *Node) {
	if node != lruCache.end {
		lruCache.removeNode(node)
		lruCache.addNode(node)
	}
}

// 移除节点
func (lruCache *LRUCache) removeNode(node *Node) int64 {
	if node == lruCache.end {
		// 如果移除的是尾节点，则更新end指针
		lruCache.end = node.pre
		if lruCache.end == nil {
			// 如果end为空，则表示链表为空
			lruCache.head = nil
		}
	} else if node == lruCache.head {
		// 如果移除的是头节点，则更新head指针
		lruCache.head = node.next
		if lruCache.head != nil {
			// 如果head不为空，说明链表中还有节点，则将head的pre指针置空
			lruCache.head.pre = nil
		}
	} else {
		// 如果移除的是中间节点，则更新前后节点的指针
		node.pre.next = node.next
		node.next.pre = node.pre
	}
	// 将节点的前后指针置空
	node.pre = nil
	node.next = nil
	return node.key
}

// 添加节点
func (lruCache *LRUCache) addNode(node *Node) {
	if node == nil {
		return
	}
	// 将节点添加到链表尾部
	if lruCache.end != nil {
		// 如果链表不为空，则更新end的next指针
		lruCache.end.next = node
		node.pre = lruCache.end
	}
	// 更新end指针
	node.next = nil
	lruCache.end = node
	if lruCache.head == nil {
		// 如果链表为空，则更新head指针
		lruCache.head = node
	}
}

// 查询本地缓存
func QueryLocalCache(key int64) (bool, map[string]interface{}) {
	if value, ok := LocalCache.Get(key); ok {
		storeInfo := make(map[string]interface{})
		storeInfo["store_info"] = map[string]interface{}{
			"item_id": value.ItemID,
			"name":    value.Name,
			"price":   value.Price,
		}
		return true, storeInfo
	}
	return false, nil
}

// 添加本地缓存
func AddLocalCache(key int64, value models.ItemCache) {
	LocalCache.Put(key, value)
}

// 更新本地缓存
func UpdateLocalCache(key int64, value models.ItemCache) {
	// 检查本地缓存是否存在该数据
	if _, ok := LocalCache.Get(key); ok {
		// 存在且未过期，则更新数据
		LocalCache.Put(key, value)
	}
}

// 删除本地缓存
func DeleteLocalCache(key int64) {
	LocalCache.mutex.Lock()
	defer LocalCache.mutex.Unlock()
	if node, ok := LocalCache.cache[key]; ok {
		LocalCache.removeNode(node)
		delete(LocalCache.cache, key)
	}
}
