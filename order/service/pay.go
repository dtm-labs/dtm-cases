package service

import (
	"database/sql"

	"github.com/dtm-labs/dtm-cases/order/common"
	"github.com/dtm-labs/dtmcli/dtmimp"
	"github.com/gin-gonic/gin"
)

func AddPayRoute(app *gin.Engine) {
	app.POST("/api/busi/payCreate", common.WrapHandler(func(c *gin.Context) interface{} {
		req := common.MustGetReq(c)
		bb := common.MustBarrierFrom(c)
		return bb.CallWithDB(common.DBGet(), func(tx *sql.Tx) error {
			_, err := dtmimp.DBExec(tx,
				"insert into busi.pay(user_id, order_id, amount, status) values(?,?,?,'CREATED')",
				req.UserID, req.OrderID, req.Amount)
			return err
		})
	}))
	app.POST("/api/busi/payCreateRevert", common.WrapHandler(func(c *gin.Context) interface{} {
		req := common.MustGetReq(c)
		bb := common.MustBarrierFrom(c)
		return bb.CallWithDB(common.DBGet(), func(tx *sql.Tx) error {
			_, err := dtmimp.DBExec(tx,
				"update busi.pay set status='CANCELED', update_time=now() where order_id=?",
				req.OrderID)
			return err
		})
	}))
}
