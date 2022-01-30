package service

import (
	"github.com/dtm-labs/dtm-cases/order/common"
	"github.com/dtm-labs/dtm-cases/order/conf"
	"github.com/dtm-labs/dtmcli"
	"github.com/gin-gonic/gin"
	"github.com/lithammer/shortuuid/v3"
)

func AddAPIRoute(app *gin.Engine) {
	app.POST("/api/busi/submitOrder", common.WrapHandler(func(c *gin.Context) interface{} {
		req := common.MustGetReq(c)
		gid := "gid-" + req.OrderID
		saga := dtmcli.NewSaga(conf.DtmServer, gid).
			Add(conf.BusiUrl+"/orderCreate", conf.BusiUrl+"/orderCreateRevert", &req).
			Add(conf.BusiUrl+"/stockDeduct", conf.BusiUrl+"/stockDeductRevert", &req).
			Add(conf.BusiUrl+"/couponUse", conf.BusiUrl+"couponUseRevert", &req).
			Add(conf.BusiUrl+"/payCreate", conf.BusiUrl+"/payCreateRevert", &req)
		return saga.Submit()
	}))
	app.Any("/api/busi/fireRequest", common.WrapHandler(func(c *gin.Context) interface{} {
		req := common.Req{
			UserID:       1,
			OrderID:      shortuuid.New(),
			ProductID:    1,
			ProductCount: 1,
			CouponID:     0,
			Amount:       100,
		}
		resty := dtmcli.GetRestyClient()
		resp, err := resty.R().SetBody(&req).Post(conf.BusiUrl + "/submitOrder")
		if err != nil {
			return err
		}
		if resp.IsError() {
			return resp.Error()
		}
		return &req
	}))
	app.Any("/api/busi/fireRollbackRequest", common.WrapHandler(func(c *gin.Context) interface{} {
		req := common.Req{
			UserID:       1,
			OrderID:      shortuuid.New(),
			ProductID:    1,
			ProductCount: 1000,
			CouponID:     0,
			Amount:       100,
		}
		resty := dtmcli.GetRestyClient()
		resp, err := resty.R().SetBody(&req).Post(conf.BusiUrl + "/submitOrder")
		if err != nil {
			return err
		}
		if resp.IsError() {
			return resp.Error()
		}
		return &req
	}))
}
