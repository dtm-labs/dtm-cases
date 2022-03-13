简体中文 | [English](./README.md)

# 缓存应用
此项目可以结合dtm文档中的 [缓存应用](https://dtm.pub/app/cache.html)阅读

## 概述
本项目主要演示了dtm如何用于维护缓存一致性，包括以下几个方面
- 保证一致性：如果发生进程crash，也能够保证缓存能够更新，保证DB与缓存的最终一致
- 延迟删除：演示了延迟删除的缓存管理方法，能够防击穿
- 强一致用法：演示了如何保证强一致，如果在缓存的降级和升级中，保证应用访问数据是强一致的

## 启动dtm
[快速启动dtm](https://dtm.pub/guide/install.html)

## 运行本例子
`go run main.go`

### 保证一致性
代码主要在demo/api-crash，例子主要演示了写DB，删除缓存的普通写法和dtm建议写法
- 发起一个普通写法请求，没有发生crash，数据一致 `curl http://localhost:8081/api/busi/normalUpdate?crash=`
- 发起一个普通写法请求，模拟crash，导致DB与缓存不一致 `curl http://localhost:8081/api/busi/normalUpdate?crash=1`
- 发起一个dtm写法请求，没有发生crash，数据一致 `curl http://localhost:8081/api/busi/dtmUpdate?crash=`
- 发起一个dtm写法请求，模拟crash，导致DB与缓存不一致，但是5s后，DB与缓存恢复一致 `curl http://localhost:8081/api/busi/dtmUpdate?crash=1`

### 延迟删除

#### 延迟删除库
代码在delay/client.go，主要函数介绍如下：
- `func NewClient(rdb *redis.Client) *Client` 创建一个延迟删除的对象
- `func (c *Client) Delete(key string) error` 延迟删除一条数据
- `func (c *Client) Obtain(key string, expire int, fn func() (string, error)) (string, error) ` 获取数据

Obtain函数用于获取数据，参数介绍如下：
- key 缓存的key
- expire 缓存的过期时间
- fn 缓存未命中时，获取数据的回调函数

#### 延迟删除用例
代码主要在demo/api-eventual

`curl http://localhost:8081/api/busi/eventualCases`

代码中解释如下：
- case-empty: 数据为空时，会调用getData1获取数据，耗时1s
- case-emptyWait：数据为空，并且缓存被锁，此时会sleep等待缓存数据
- case-exists：数据存在，正常返回
- case-delayDeleteQuery1：数据存在，但是已经被延迟删除，此时会启动异步协程去计算新结果，并立即返回缓存中的旧值
- case-delayDeleteQuery2：数据存在，但是已经被延迟删除，并且被锁，此时不会计算新结果，只会立即返回缓存中的旧值
- case-delayDeleteQuery3：数据存在，Query1中的协程已计算出新结果，并更新到缓存，返回缓存中的新值

### 强一致访问

#### 强一致缓存库
代码在delay/client.go，主要函数介绍如下：
- `func (c *Client) StrongObtain(key string, expire int, fn func() (string, error)) (string, error) ` 获取数据

#### 强一致访问用例
代码主要在demo/api-strong，例子主要演示了强一致访问升降级的各种特性，通过下面代码运行强一致访问的测试用例

`curl http://localhost:8081/api/busi/strongDemo`

代码中解释如下：
- 初始状态没有使用cache，只是使用DB
- 打开写缓存，此时部分写请求只写DB，部分写请求同时写cache和DB
- 写缓存开关对所有应用生效，此时打开读缓存开关，就保证缓存读取的结果是正确的
- 读缓存开关对所有应用生效
- 如果遇见缓存故障，要进行降级
- 关闭读缓存开关，此时部分读请求读取缓存，部分读取DB
- 读缓存开关对所有应用生效，此时所有读取，都只读DB
- 关闭写缓存开关
- 写缓存开关对所有应用生效，此时缓存不再被使用，可以下线