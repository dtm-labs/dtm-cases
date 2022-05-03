package demo

import (
	"database/sql"
	"errors"
	"time"

	"github.com/dtm-labs/dtm-cases/utils"
	"github.com/dtm-labs/dtmcli"
	"github.com/dtm-labs/dtmcli/logger"
	"github.com/gin-gonic/gin"
	"github.com/lithammer/shortuuid"
)

func init() {
	BusiApp.GET(BusiAPI+"/version", utils.WrapHandler(func(c *gin.Context) interface{} {
		clearData()
		mode := c.Query("mode")
		if mode != "delete" && mode != "rockscache" {
			return errors.New("mode should be delete or rockscache")
		}
		data := "v1"
		_ = Post(BusiUrl+"/versionUpdateData", map[string]interface{}{
			"key":       DataKey,
			"value":     data,
			"time_cost": "2s",
			"mode":      mode,
		})
		_ = Post(BusiUrl+"/versionQueryData", map[string]interface{}{
			"key":  DataKey,
			"mode": mode,
		})
		data = "v2"
		time.Sleep(200 * time.Millisecond)
		_ = Post(BusiUrl+"/versionUpdateData", map[string]interface{}{
			"key":       DataKey,
			"value":     data,
			"time_cost": "5ms",
			"mode":      mode,
		})
		_ = Post(BusiUrl+"/versionQueryData", map[string]interface{}{
			"key":  DataKey,
			"mode": mode,
		})
		time.Sleep(2500 * time.Millisecond)
		dbv := GetDBValue(DataKey)
		cachev := GetCacheValue(DataKey, mode)
		ensure(dbv.V == "v2", "db value should be v2")
		if mode == "delete" {
			ensure(cachev == "v1", "cache value should be v1")
			logger.Infof("for mode delete, db value is: %s, cache value is: %s, not matched", dbv, cachev)
		} else {
			ensure(cachev == "v1", "cache value should be v1")
			logger.Infof("for mode rockscache, db value is: %s, cache value is: %s, matched", dbv, cachev)
		}
		return "ok"
	}))
	BusiApp.POST(BusiAPI+"/versionUpdateData", utils.WrapHandler(func(c *gin.Context) interface{} {
		body := MustMapBodyFrom(c)
		gid := shortuuid.New()
		msg := dtmcli.NewMsg(DtmServer, gid).
			Add(BusiUrl+"/deleteCache", body)
		msg.WaitResult = true // when return success, the global transaction has finished
		return msg.DoAndSubmit(BusiUrl+"/queryPrepared", func(bb *dtmcli.BranchBarrier) error {
			return bb.CallWithDB(db, func(tx *sql.Tx) error {
				return UpdateInTx(tx, &DBRow{
					K:        body["key"].(string),
					V:        body["value"].(string),
					TimeCost: body["time_cost"].(string),
				})
			})
		})
	}))
	BusiApp.POST(BusiAPI+"/versionQueryData", utils.WrapHandler(func(c *gin.Context) interface{} {
		body := MustMapBodyFrom(c)
		mode := body["mode"].(string)
		go func() {
			fetch := func() (string, error) {
				row := GetDBValue(body["key"].(string))
				if row.TimeCost != "" {
					duration, err := time.ParseDuration(row.TimeCost)
					logger.FatalIfError(err)
					logger.Debugf("before sleep %s, return %s", duration, row.V)
					time.Sleep(duration)
					logger.Debugf("after sleep %s, return %s", duration, row.V)
				}
				return row.V, nil
			}
			if mode == "rockscache" {
				_, _ = dc.Fetch(body["key"].(string), 300, fetch)
			} else {
				_, _ = NormalFetch(body["key"].(string), 300, fetch)
			}

		}()
		return "query data started asynchronously"
	}))
}
