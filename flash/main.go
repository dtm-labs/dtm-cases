package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/dtm-labs/dtm-cases/utils"
	"github.com/dtm-labs/dtmcli"
	"github.com/dtm-labs/dtmcli/logger"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/lithammer/shortuuid"
)

var redisOption = redis.Options{
	Addr:     "dtm.pub:6379",
	Username: "root",
	Password: "",
}

var DtmServer = "http://localhost:36789/api/dtmsvr"

const BusiAPI = "/api/busi"
const BusiPort = 8081

var BusiUrl = fmt.Sprintf("http://localhost:%d%s", BusiPort, BusiAPI)

var rdb = redis.NewClient(&redisOption)

var stockKey = "{a}--stock-1"

func main() {
	logger.InitLog("debug")
	_, err := rdb.Set(context.Background(), stockKey, "4", 86400*time.Second).Result()
	logger.FatalIfError(err)
	startSvr()
	select {}
}

func startSvr() {
	gin.SetMode(gin.ReleaseMode)
	app := gin.Default()
	addRoutes(app)
	log.Printf("flash sales examples listening at %d", BusiPort)
	go app.Run(fmt.Sprintf(":%d", BusiPort))
	time.Sleep(100 * time.Millisecond)
}

func addRoutes(app *gin.Engine) {
	app.GET(BusiAPI+"/redisQueryPrepared", utils.WrapHandler(func(c *gin.Context) interface{} {
		bb := utils.MustBarrierFrom(c)
		return bb.RedisQueryPrepared(rdb, 7*86400)
	}))
	app.GET(BusiAPI+"/createOrder", utils.WrapHandler(func(c *gin.Context) interface{} {
		logger.Infof("createOrder")
		return nil
	}))
	app.Any(BusiAPI+"/flashSales", utils.WrapHandler(func(c *gin.Context) interface{} {
		gid := "{a}-" + shortuuid.New() // gid should contain same {a} as stockKey, so that the data will be in same redis slot
		msg := dtmcli.NewMsg(DtmServer, gid).
			Add(BusiUrl+"/createOrder", nil)
		return msg.DoAndSubmit(BusiUrl+"/redisQueryPrepared", func(bb *dtmcli.BranchBarrier) error {
			return bb.RedisCheckAdjustAmount(rdb, stockKey, -1, 86400)
		})
	}))
}
