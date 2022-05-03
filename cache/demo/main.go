package demo

import (
	"fmt"
	"log"
	"time"

	"github.com/dtm-labs/dtmcli/logger"
	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

var DtmServer = "http://localhost:36789/api/dtmsvr"

const BusiAPI = "/api/busi"
const BusiPort = 8081

var BusiUrl = fmt.Sprintf("http://localhost:%d%s", BusiPort, BusiAPI)

var BusiApp = gin.Default()

func init() {
	gin.SetMode(gin.ReleaseMode)
}

func Main() {
	logger.InitLog("debug")
	startSvr()
	time.Sleep(200 * time.Millisecond)
	select {}
}

func startSvr() {
	log.Printf("cache examples listening at %d", BusiPort)
	go BusiApp.Run(fmt.Sprintf(":%d", BusiPort))
	time.Sleep(100 * time.Millisecond)
}
