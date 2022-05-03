简体中文 | [English](./README.md)

# 缓存应用
此项目可以结合dtm文档中的 [缓存应用](https://dtm.pub/app/cache.html)阅读

## 概述
本项目主要演示了[dtm-labs/rockscache]和[dtm-labs/dtm]配合如何用于维护缓存一致性，包括以下几个方面
- 保证最终一致性：演示了rockscache如何彻底解决缓存与DB版本一致的问题
- 保证原子操作：如果更新完DB后，发生进程crash，也能够保证缓存能够更新
- 缓存管理的其他功能：演示了延迟删除的缓存管理方法，能够防击穿，防穿透，防雪崩（未演示）
- 强一致用法：演示了如何保证强一致
- 升降级中的强一致：如果在缓存的降级和升级中，保证应用访问数据是强一致的

## 启动dtm
[快速启动dtm](https://dtm.pub/guide/install.html)

## 运行本例子
`go run main.go`

### 保证最终一致性
代码主要在demo/api-version，例子主要演示了rockscache与dtm结合，解决了删除缓存未能解决的版本不一致问题
- 发起一个请求，演示了普通删除缓存方案下，版本不一致的问题 `curl http://localhost:8081/api/busi/version?mode=delete`
- 发起一个请求，演示了rockscache+dtm配合解决版本不一致的问题 `curl http://localhost:8081/api/busi/version?mode=rockscache`

在这个demo里面，主要有一下几点
1. 会将DB的数据初始化为v1，然后查询缓存，无数据后查询DB获得v1，睡眠几秒，模拟网络延迟，然后更新缓存。
2. 接着将数据修改为v2，然后查询缓存，无数据后查询DB获得v2，睡眠几毫秒，模拟无网络延迟，然后更新缓存。

- 对于mode=delete，由于网络延迟，导致v1写入缓存的时间在v2之后，导致缓存中的最终版本为v1，与数据库不一致
- 对于mode=rockscache，虽然遇见了网络延时，但是最终缓存中的版本是v2，与数据库中的一致
### 保证原子性
- 发起一个普通更新数据请求，模拟crash，导致DB与缓存不一致 `curl http://localhost:8081/api/busi/atomic?mode=none`
- 发起一个通过dtm更新数据请求，模拟crash，导致DB与缓存不一致，但是5s后，DB与缓存恢复一致 `curl http://localhost:8081/api/busi/atomic?mode=dtm`crash=1`

### 强一致访问

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
