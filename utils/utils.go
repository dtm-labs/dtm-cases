package utils

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dtm-labs/dtmcli"
	"github.com/dtm-labs/dtmcli/logger"
	"github.com/gin-gonic/gin"
)

func WrapHandler(fn func(*gin.Context) interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		began := time.Now()
		ret := fn(c)
		status := dtmcli.Result2HttpCode(ret)

		b, _ := json.Marshal(ret)
		if status == http.StatusOK || status == http.StatusTooEarly {
			logger.Infof("%2dms %d %s %s %s", time.Since(began).Milliseconds(), status, c.Request.Method, c.Request.RequestURI, string(b))
		} else {
			logger.Errorf("%2dms %d %s %s %s", time.Since(began).Milliseconds(), status, c.Request.Method, c.Request.RequestURI, string(b))
		}
		c.JSON(status, ret)
	}
}

func MustBarrierFrom(c *gin.Context) *dtmcli.BranchBarrier {
	bb, err := dtmcli.BarrierFromQuery(c.Request.URL.Query())
	if err != nil {
		panic(err)
	}
	return bb
}
