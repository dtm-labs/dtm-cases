package main

import (
	"fmt"
	"log"
	"time"

	"github.com/dtm-labs/dtmcli/logger"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql" // register mysql driver
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

func main() {
	logger.InitLog("debug")
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
}
