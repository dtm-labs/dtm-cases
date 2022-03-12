package demo

import (
	"github.com/dtm-labs/dtm-cases/cache/delay"
	"github.com/dtm-labs/dtm-cases/utils"
	"github.com/dtm-labs/dtmcli"
	"github.com/dtm-labs/dtmcli/logger"
	"github.com/gin-gonic/gin"
)

func addBaseRoute(app *gin.Engine) {

	app.POST(BusiAPI+"/dtmDelKey", utils.WrapHandler(func(c *gin.Context) interface{} {
		req := delay.MustReqFrom(c)
		logger.Infof("deleting key: %s", req.Key)
		_, err := rdb.Del(rdb.Context(), req.Key).Result()
		logger.FatalIfError(err)
		return nil
	}))
	app.GET(BusiAPI+"/dtmQueryPrepared", utils.WrapHandler(func(c *gin.Context) interface{} {
		bb, err := dtmcli.BarrierFromQuery(c.Request.URL.Query())
		logger.FatalIfError(err)
		return bb.QueryPrepared(db)
	}))
}
