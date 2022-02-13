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
	"github.com/lithammer/shortuuid"
)

var k = "key1"

func updateValue(nvalue string) {
	msg := dtmcli.NewMsg(DtmServer, shortuuid.New()).
		Add(BusiUrl+"/delayDelete", &delay.Req{Key: k})
	msg.WaitResult = true // wait for result. when submit returned without error, cache will be deleted
	err := msg.DoAndSubmitDB(BusiUrl+"/delayQueryPrepared", db, func(tx *sql.Tx) error {
		_, err := db.Exec("insert into cache1.t1(id, value) values(?, ?) on duplicate key update value=values(value)", 1, nvalue)
		logger.FatalIfError(err)
		return nil
	})
	logger.FatalIfError(err)
}

func getData1() (string, error) {
	var v string
	err := db.QueryRow("select value from cache1.t1 where id = ?", 1).Scan(&v)
	logger.FatalIfError(err)
	logger.Infof("get Data sleeping 1s")
	time.Sleep(1 * time.Second)
	return v, nil
}

func getData2() (string, error) {
	r, err := getData1()
	logger.Infof("get Data sleeping 3s")
	time.Sleep(3 * time.Second)
	return r, err
}

func obtain() (result string, used int) {
	begin := time.Now()
	v, err := dc.Obtain(k, 86400, 3, getData1)
	logger.FatalIfError(err)
	return v, int(time.Since(begin).Seconds())
}

func addDelayDelete(app *gin.Engine) {
	app.POST(BusiAPI+"/delayDelete", utils.WrapHandler(func(c *gin.Context) interface{} {
		req := delay.MustReqFrom(c)
		err := dc.Delete(req.Key)
		logger.FatalIfError(err)
		return nil
	}))
	app.GET(BusiAPI+"/delayQueryPrepared", utils.WrapHandler(func(c *gin.Context) interface{} {
		bb, err := dtmcli.BarrierFromQuery(c.Request.URL.Query())
		logger.FatalIfError(err)
		return bb.QueryPrepared(db)
	}))

	app.GET(BusiAPI+"/delayDeleteDemo", utils.WrapHandler(func(c *gin.Context) interface{} {
		_, err := db.Exec("insert ignore into cache1.t1(id, value) values(?, ?)", 1, "value1")
		logger.FatalIfError(err)
		v, err := dc.Obtain("t1-id-1", 86400, 3, func() (string, error) {
			logger.Debugf("querying db")
			var value string
			err := db.QueryRow("select value from cache1.t1 where id = ?", 1).Scan(&value)
			return value, err
		})
		logger.FatalIfError(err)
		return v
	}))
	app.GET(BusiAPI+"/delayDeleteCases", utils.WrapHandler(func(c *gin.Context) interface{} {
		updateValue("value1")
		expected := "value1"
		// case-empty: no data exists
		_, err := rdb.Del(rdb.Context(), k).Result()
		logger.FatalIfError(err)
		go func() {
			v, _ := obtain()
			logger.FatalfIf(v != expected, "case-empty: expect %s, but got %s", expected, v)
		}()

		// case-emptyWait: no data exists, but wait for data
		time.Sleep(200 * time.Millisecond)
		v, _ := obtain()
		logger.FatalfIf(v != expected, "case-exists: expect %s, but got %s", expected, v)

		// case-exists: data exists
		v, _ = obtain()
		logger.FatalfIf(v != expected, "case-exists: expect %s, but got %s", expected, v)

		// case-delayDeleteQuery1: data exists, but delay deleted so return old value and get new data async
		updateValue("value2")
		logger.FatalIfError(err)
		v, used := obtain()
		logger.FatalfIf(v != expected, "case-delayDeleteQuery1: expect %s, but got %s", expected, v)
		logger.FatalfIf(used > 0, "case-delayDelete: expect 0, but got %d", used)

		// case-delayDeleteQuery2: data exists, but delay deleted and locked return old value
		v, used = obtain()
		logger.FatalfIf(v != expected, "case-delayDeleteQuery2: expect %s, but got %s", expected, v)
		logger.FatalfIf(used > 0, "case-delayDelete: expect 0, but got %d", used)

		// case-delayDeleteQuery3: data already replaced by new data
		time.Sleep(1200 * time.Millisecond)
		expected = "value2"
		v, used = obtain()
		logger.FatalfIf(v != expected, "case-delayDeleteQuery3: expect %s, but got %s", expected, v)
		logger.FatalfIf(used > 0, "case-delayDeleteQuery3: expect 0, but got %d", used)

		// case-delayDeleteVersionBug
		err = dc.Delete(k)
		logger.FatalIfError(err)
		updateValue("value3")
		v, err = dc.Obtain(k, 86400, 4, getData2)
		logger.FatalIfError(err)
		logger.FatalfIf(v != expected, "case-delayDeleteVersionBugFirstObtain: expect %s, but got %s", expected, v)
		time.Sleep(200 * time.Millisecond) // wait for getData2 to finish intv update
		updateValue("value4")
		logger.FatalIfError(err)
		v, used = obtain()
		logger.FatalfIf(v != expected, "case-delayDeleteVersionBugSecondObtain: expect %s, but got %s", expected, v)
		time.Sleep(4 * time.Second)
		expected = "value3"
		v, used = obtain()
		logger.FatalfIf(v != expected, "case-delayDeleteVersionBugSecondObtain: expect %s, but got %s", expected, v)
		msg := fmt.Sprintf("finally, value is: %s, but value in cache is: %s, they are not matched", "value4", v)
		logger.Infof(msg)
		return msg
	}))

}
