# 订单应用

此项目可以结合dtm文档中的 [订单应用](https://dtm.pub/app/order.html)阅读

## 概述
本项目主要演示了dtm如何应用于非单体的订单系统，保证订单中的多个步骤，能够最终“原子”执行，保证最终一致性。

快速运行项目：

`go run main.go`

发起一个订单

`curl http://localhost:8081/api/busi/fireRequest`

本项目有以下内容
- main.go: 主程序文件
- service目录：相关各个服务文件
- conf目录：相关配置
- common目录：多个服务共享的代码
