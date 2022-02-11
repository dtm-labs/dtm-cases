package service

import (
	"database/sql"

	"github.com/dtm-labs/dtm-cases/order/common"
	"github.com/dtm-labs/dtmcli/dtmimp"
	"github.com/gin-gonic/gin"
)

func AddOrderRoute(app *gin.Engine) {
	app.POST("/api/busi/orderCreate", common.WrapHandler(func(c *gin.Context) interface{} {
		req := common.MustGetReq(c)
		bb := common.MustBarrierFrom(c)
		return bb.CallWithDB(common.DBGet(), func(tx *sql.Tx) error {
			_, err := dtmimp.DBExec(tx,
				"insert into ord.order1(user_id, order_id, product_id, amount, status) values(?,?,?,?,'PAYING')",
				req.UserID, req.OrderID, req.ProductID, req.Amount)
			return err
		})
	}))
	app.POST("/api/busi/orderCreateRevert", common.WrapHandler(func(c *gin.Context) interface{} {
		req := common.MustGetReq(c)
		bb := common.MustBarrierFrom(c)
		return bb.CallWithDB(common.DBGet(), func(tx *sql.Tx) error {
			_, err := dtmimp.DBExec(tx,
				"update ord.order1 set status='FAILED', update_time=now() where order_id=?",
				req.OrderID)
			return err
		})
	}))
}
