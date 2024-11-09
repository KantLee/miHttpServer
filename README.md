# miHttpServer

## 开发环境
1. go 1.22.4
2. mysql 8.4.0
3. redis 7.2.5
4. Gin后端框架 github.com/gin-gonic/gin
5. 操作MySQL的库 xorm.io/xorm
6. 操作Redis的库 github.com/gomodule/redigo/redis

## 开发目标

- [x] 增加商品信息（利用 Redis 分布式锁防止并发问题）
- [x] 修改商品信息 （利用 Redis 分布式锁防止并发问题）
- [x] 查询商品信息（利用 Mysql持久化数据，redis和本地缓存实现双层缓存）
- [x] 删除商品信息（符合幂等性，响应结果给出删除时间）

## 开发进度

- [x] 需求分析
- [x] 设计项目架构
- [x] 实现基础日志功能
- [x] 增加商品信息（未使用Redis）
- [x] 修改商品信息（未使用Redis）
- [x] 查询商品信息（未使用Redis）
- [x] 删除商品信息（未考虑不同时区）
- [x] 删除商品信息（响应时间和时区关联）
- [x] 添加Redis缓存
- [x] 实现Redis分布式锁
- [x] 重构代码，进行分层
- [x] 删除商品时符合幂等性
- [x] 实现基于LRU策略的本地缓存
