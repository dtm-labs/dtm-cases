package common

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/dtm-labs/dtm-cases/order/conf"
	"github.com/dtm-labs/dtmcli"
	"github.com/dtm-labs/dtmcli/dtmimp"
	"github.com/dtm-labs/dtmcli/logger"
	"github.com/gin-gonic/gin"
)

func getStatusFromResult(result interface{}) int {
	err, _ := result.(error)
	if errors.Is(err, dtmcli.ErrFailure) {
		return http.StatusConflict
	} else if errors.Is(err, dtmcli.ErrOngoing) {
		return http.StatusTooEarly
	} else if err != nil {
		return http.StatusInternalServerError
	}
	return http.StatusOK
}

func WrapHandler(fn func(*gin.Context) interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		began := time.Now()
		ret := fn(c)
		status := getStatusFromResult(ret)

		b, _ := json.Marshal(ret)
		if status == http.StatusOK || status == http.StatusTooEarly {
			logger.Infof("%2dms %d %s %s %s", time.Since(began).Milliseconds(), status, c.Request.Method, c.Request.RequestURI, string(b))
		} else {
			logger.Errorf("%2dms %d %s %s %s", time.Since(began).Milliseconds(), status, c.Request.Method, c.Request.RequestURI, string(b))
		}
		c.JSON(status, ret)
	}
}

func DBGet() *sql.DB {
	db, err := dtmimp.PooledDB(conf.DBConf)
	if err != nil {
		panic(err)
	}
	return db
}

type Req struct {
	UserID       int    `json:"user_id"`
	OrderID      string `json:"order_id"`
	Amount       int    `json:"amount"` // money amount
	ProductID    int    `json:"product_id"`
	ProductCount int    `json:"product_count"` // how many product to order
	CouponID     int    `json:"coupon_id"`     // optional
}

func MustGetReq(c *gin.Context) *Req {
	var req Req
	err := c.BindJSON(&req)
	if err != nil {
		panic(err)
	}
	if req.UserID == 0 {
		panic("user_id not specified")
	}
	if req.OrderID == "" {
		panic("order_Id not specified")
	}
	if req.Amount == 0 {
		panic("amount not specified")
	}
	if req.ProductID == 0 {
		panic("product_id not specified")
	}
	return &req
}

func MustBarrierFrom(c *gin.Context) *dtmcli.BranchBarrier {
	bb, err := dtmcli.BarrierFromQuery(c.Request.URL.Query())
	if err != nil {
		panic(err)
	}
	return bb
}
