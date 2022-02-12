package demo

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/dtm-labs/dtm-cases/cache/delay"
	"github.com/dtm-labs/dtm-cases/utils"
	"github.com/dtm-labs/dtmcli"
	"github.com/dtm-labs/dtmcli/logger"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/lithammer/shortuuid"
)

func updateDB(value string) {
	_, err := db.Exec("insert into cache1.t1(id, value) values(?, ?) on duplicate key update value=values(value)", 1, value)
	logger.FatalIfError(err)
}

func queryValue() string {
	v, err := rdb.Get(rdb.Context(), key).Result()
	if err == redis.Nil {
		logger.Debugf("querying db")
		var value string
		err = db.QueryRow("select value from cache1.t1 where id = ?", 1).Scan(&value)
		_, err = rdb.Set(rdb.Context(), key, value, 0).Result()
		logger.FatalIfError(err)
		return value
	}
	logger.FatalIfError(err)
	return v
}

func addConsistencyRoute(app *gin.Engine) {
	app.GET(BusiAPI+"/normalUpdate", utils.WrapHandler(func(c *gin.Context) interface{} {
		// set up
		updateDB("none")
		_, err := rdb.Set(rdb.Context(), key, "none", 0).Result()
		logger.FatalIfError(err)

		// update DB and Del and Query
		value := "normalUpdate-" + shortuuid.New()
		go func() {
			crash := c.Query("crash")
			updateDB(value)
			if crash != "" {
				select {}
			}
			_, err = rdb.Del(rdb.Context(), key).Result()
			logger.FatalIfError(err)
		}()
		v := queryValue()
		if v == value {
			return fmt.Sprintf("cache and db match. both are: %s", value)
		}
		return fmt.Sprintf("cache and db mismatch. cache is: %s, db is: %s", v, value)
	}))

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
	app.GET(BusiAPI+"/dtmUpdate", utils.WrapHandler(func(c *gin.Context) interface{} {
		// set up
		updateDB("none")
		_, err := rdb.Set(rdb.Context(), key, "none", 0).Result()
		logger.FatalIfError(err)

		// update DB and Del and Query
		value := "normalUpdate-" + shortuuid.New()
		crash := c.Query("crash")
		go func() {
			msg := dtmcli.NewMsg(DtmServer, shortuuid.New()).
				Add(BusiUrl+"/dtmDelKey", &delay.Req{Key: key})
			msg.TimeoutToFail = 3
			err = msg.DoAndSubmit(BusiUrl+"/dtmQueryPrepared", func(bb *dtmcli.BranchBarrier) error {
				err := bb.CallWithDB(db, func(tx *sql.Tx) error {
					_, err := tx.Exec("insert into cache1.t1(id, value) values(?, ?) on duplicate key update value=values(value)", 1, value)
					return err
				})
				if crash != "" {
					select {}
				}
				return err
			})
			logger.FatalIfError(err)
		}()

		logger.Infof("sleeping for 6 seconds to wait for the update")
		if crash != "" {
			time.Sleep(6 * time.Second)
		} else {
			time.Sleep(1 * time.Second)
		}
		v := queryValue()
		logger.FatalfIf(v != value, "expect %s, but got %s", value, v)
		return fmt.Sprintf("cache and db match. both are: %s", value)
	}))
}
