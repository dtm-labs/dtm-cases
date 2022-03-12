package demo

import (
	"database/sql"
	"time"

	"github.com/dtm-labs/dtm-cases/cache/delay"
	"github.com/dtm-labs/dtm-cases/utils"
	"github.com/dtm-labs/dtmcli"
	"github.com/dtm-labs/dtmcli/logger"
	"github.com/gin-gonic/gin"
	"github.com/lithammer/shortuuid"
)

func addConsistencyRoute(app *gin.Engine) {
	app.GET(BusiAPI+"/normalUpdate", utils.WrapHandler(func(c *gin.Context) interface{} {
		// set up
		updateDB("none")
		_, err := rdb.Set(rdb.Context(), rdbKey, "none", 0).Result()
		logger.FatalIfError(err)

		// update DB and Del and Query
		value := "normalUpdate-" + shortuuid.New()
		crash := c.Query("crash")
		go func() {
			updateDB(value)
			if crash != "" { // simulate crash
				select {}
			}
			_, err = rdb.Del(rdb.Context(), rdbKey).Result()
			logger.FatalIfError(err)
		}()
		time.Sleep(time.Millisecond * 200)
		v := obtainValue()
		if crash == "" {
			ensure(v == value, "expecting v == value v=%s, value=%s", v, value)
		} else {
			ensure(v != value, "expecting v != value v=%s, value=%s", v, value)
		}
		return nil
	}))

	app.GET(BusiAPI+"/dtmUpdate", utils.WrapHandler(func(c *gin.Context) interface{} {
		// set up
		updateDB("none")
		_, err := rdb.Set(rdb.Context(), rdbKey, "none", 0).Result()
		logger.FatalIfError(err)

		// update DB and Del and Query
		value := "normalUpdate-" + shortuuid.New()
		crash := c.Query("crash")
		go func() {
			msg := dtmcli.NewMsg(DtmServer, shortuuid.New()).
				Add(BusiUrl+"/deleteKey", &delay.Req{Key: rdbKey})
			msg.TimeoutToFail = 3
			err = msg.DoAndSubmit(BusiUrl+"/queryPrepared", func(bb *dtmcli.BranchBarrier) error {
				err := bb.CallWithDB(db, func(tx *sql.Tx) error {
					return updateInTx(tx, value)
				})
				if crash != "" { // simulate crash
					select {}
				}
				return err
			})
			logger.FatalIfError(err)
		}()

		logger.Infof("sleeping for 7 seconds to wait for the update")
		if crash != "" {
			time.Sleep(7 * time.Second)
		} else {
			time.Sleep(200 * time.Millisecond)
		}
		v := obtainValue()
		ensure(v == value, "expecting v == value v=%s, value=%s", v, value)
		return nil
	}))
}
