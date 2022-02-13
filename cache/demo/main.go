package demo

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/dtm-labs/dtm-cases/cache/delay"
	"github.com/dtm-labs/dtmcli/logger"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	_ "github.com/go-sql-driver/mysql"
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

var dc = delay.NewClient(rdb, 3)

var db *sql.DB

var key = "t1-id-1"

func init() {
	var err error
	db, err = sql.Open("mysql", "dtm:passwd123dtm@tcp(dtm.pub:3306)/cache1?charset=utf8")
	logger.FatalIfError(err)
}

var stockKey = "{a}--stock-1"
var orderCreated int64

func Main() {
	logger.InitLog("debug")
	startSvr()
	time.Sleep(200 * time.Millisecond)
	// dtmcli.GetRestyClient().R().Get(BusiUrl + "/delayDeleteCases")
	select {}
}

func startSvr() {
	gin.SetMode(gin.ReleaseMode)
	app := gin.Default()
	addRoutes(app)
	log.Printf("cache examples listening at %d", BusiPort)
	go app.Run(fmt.Sprintf(":%d", BusiPort))
	time.Sleep(100 * time.Millisecond)
}

func addRoutes(app *gin.Engine) {
	addConsistencyRoute(app)
	addDelayDelete(app)
	addStringConsistency(app)
}
