package demo

import (
	"github.com/dtm-labs/dtm-cases/utils"
	"github.com/dtm-labs/dtmcli"
	"github.com/dtm-labs/dtmcli/logger"
	"github.com/gin-gonic/gin"
)

func init() {

	BusiApp.POST(BusiAPI+"/deleteKey", utils.WrapHandler(func(c *gin.Context) interface{} {
		req := MustReqFrom(c)
		logger.Infof("deleting key: %s", req.Key)
		_, err := rdb.Del(rdb.Context(), req.Key).Result()
		return err
	}))
	BusiApp.GET(BusiAPI+"/queryPrepared", utils.WrapHandler(func(c *gin.Context) interface{} {
		bb, err := dtmcli.BarrierFromQuery(c.Request.URL.Query())
		if err == nil {
			err = bb.QueryPrepared(db)
		}
		return err
	}))
	BusiApp.POST(BusiAPI+"/delayDeleteKey", utils.WrapHandler(func(c *gin.Context) interface{} {
		req := MustReqFrom(c)
		return dc.DelayDelete(req.Key)
	}))
	BusiApp.POST(BusiAPI+"/deleteCache", utils.WrapHandler(func(c *gin.Context) interface{} {
		body := MustMapBodyFrom(c)
		mode := body["mode"].(string)
		if mode == "delete" {
			return rdb.Del(rdb.Context(), body["key"].(string))
		} else {
			return dc.DelayDelete(body["key"].(string))
		}
	}))
}
