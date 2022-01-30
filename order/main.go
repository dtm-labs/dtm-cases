package main

import (
	"fmt"
	"log"
	"time"

	"github.com/dtm-labs/dtm-cases/order/conf"
	"github.com/dtm-labs/dtm-cases/order/service"
	"github.com/dtm-labs/dtmcli"
	"github.com/dtm-labs/dtmcli/logger"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql" // register mysql driver
)

// 事务参与者的服务地址

func startSvr() {
	gin.SetMode(gin.ReleaseMode)
	app := gin.Default()
	addRoutes(app)
	log.Printf("order examples listening at %d", conf.BusiPort)
	go app.Run(fmt.Sprintf(":%d", conf.BusiPort))
	time.Sleep(100 * time.Millisecond)
}

func addRoutes(app *gin.Engine) {
	service.AddAPIRoute(app)
	service.AddCouponRoute(app)
	service.AddOrderRoute(app)
	service.AddPayRoute(app)
	service.AddStockRoute(app)
}

func main() {
	logger.InitLog("debug")
	startSvr()
	fireRequest()
	time.Sleep(1000 * time.Second)
}

func fireRequest() {
	resp, err := dtmcli.GetRestyClient().R().Post(conf.BusiUrl + "/fireRequest")
	if err != nil {
		panic(err)
	}
	if resp.IsError() {
		panic(resp.Error())
	}
}
