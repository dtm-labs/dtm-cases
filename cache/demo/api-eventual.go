package demo

import (
	"database/sql"
	"time"

	"github.com/dtm-labs/dtm-cases/utils"
	"github.com/dtm-labs/dtmcli"
	"github.com/dtm-labs/dtmcli/logger"
	"github.com/gin-gonic/gin"
	"github.com/lithammer/shortuuid"
)

func eventualUpdateValue(value string) {
	msg := dtmcli.NewMsg(DtmServer, shortuuid.New()).
		Add(BusiUrl+"/delayDeleteKey", &Req{Key: rdbKey})
	msg.WaitResult = true // wait for result. when submit returned without error, cache has been deleted
	err := msg.DoAndSubmitDB(BusiUrl+"/queryPrepared", db, func(tx *sql.Tx) error {
		return updateInTx(tx, value)
	})
	logger.FatalIfError(err)
}

func eventualSleepGetDB() (string, error) {
	v, err := getDB()
	logger.Infof("get Data sleeping 1s")
	time.Sleep(1 * time.Second)
	return v, err
}

func eventualObtain() (result string, used int) {
	begin := time.Now()
	v, err := dc.Obtain(rdbKey, 86400, eventualSleepGetDB)
	logger.FatalIfError(err)
	return v, int(time.Since(begin).Seconds())
}

func addDelayDelete(app *gin.Engine) {
	app.GET(BusiAPI+"/eventualNormal", utils.WrapHandler(func(c *gin.Context) interface{} {
		updateDB("value1")
		v, err := dc.Obtain(rdbKey, 86400, getDB)
		logger.FatalIfError(err)
		return v
	}))
	app.GET(BusiAPI+"/eventualCases", utils.WrapHandler(func(c *gin.Context) interface{} {
		updateDB("value1")
		expected := "value1"
		// case-empty: no data exists
		_, err := rdb.Del(rdb.Context(), rdbKey).Result()
		logger.FatalIfError(err)
		go func() {
			v, _ := eventualObtain()
			ensure(v == expected, "case-empty: v == expected, v=%s, expected=%s", v, expected)
		}()

		// case-emptyWait: no data exists, but wait for data
		time.Sleep(200 * time.Millisecond)
		v, _ := eventualObtain()
		ensure(v == expected, "case-emptyWait: v == expected, v=%s, expected=%s", expected, v)

		// case-exists: data exists
		v, _ = eventualObtain()
		ensure(v == expected, "case-exists: v == expected, v=%s, expected=%s", v, expected)

		// case-delayDeleteQuery1: data exists, but delay deleted so return old value and get new data async
		eventualUpdateValue("value2")

		v, used := eventualObtain()
		ensure(v == expected, "case-delayDeleteQuery1: v == expected, v=%s, expected=%s", v, expected)
		ensure(used == 0, "case-delayDeleteQuery1: used == 0, used=%d", used)

		// case-delayDeleteQuery2: data exists, but delay deleted and locked return old value
		v, used = eventualObtain()
		ensure(v == expected, "case-delayDeleteQuery2: v == expected, v=%s, expected=%s", v, expected)
		ensure(used == 0, "case-delayDeleteQuery2: used == 0, used=%d", used)

		// case-delayDeleteQuery3: data already replaced by new data
		time.Sleep(1200 * time.Millisecond)
		expected = "value2"
		v, used = eventualObtain()
		ensure(v == expected, "case-delayDeleteQuery3: v == expected, v=%s, expected=%s", v, expected)
		ensure(used == 0, "case-delayDeleteQuery3: used == 0, used=%d", used)
		return "finished"
	}))

}
