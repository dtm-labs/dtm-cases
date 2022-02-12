package demo

import (
	"fmt"
	"time"

	"github.com/dtm-labs/dtm-cases/cache/delay"
	"github.com/dtm-labs/dtm-cases/utils"
	"github.com/dtm-labs/dtmcli/logger"
	"github.com/gin-gonic/gin"
)

func addDelayDelete(app *gin.Engine) {
	app.POST(BusiAPI+"/delayDelete", utils.WrapHandler(func(c *gin.Context) interface{} {
		req := delay.MustReqFrom(c)
		err := dc.Delete(req.Key)
		logger.FatalIfError(err)
		return nil
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
		k := "key1"
		intv := 1
		value := "value1"
		getData1 := func() (string, error) {
			logger.Infof("get Data used 1s")
			value = fmt.Sprintf("value%d", intv)
			intv++
			v := value
			time.Sleep(1 * time.Second)
			return v, nil
		}
		getData2 := func() (string, error) {
			r, err := getData1()
			time.Sleep(3 * time.Second)
			return r, err
		}
		obtain := func() (result string, used int) {
			begin := time.Now()
			v, err := dc.Obtain(k, 86400, 3, getData1)
			logger.FatalIfError(err)
			return v, int(time.Since(begin).Seconds())
		}
		// case-empty: no data exists
		_, err := rdb.Del(rdb.Context(), k).Result()
		logger.FatalIfError(err)
		go func() {
			v, _ := obtain()
			logger.FatalfIf(v != value, "case-empty: expect %s, but got %s", value, v)
		}()

		// case-emptyWait: no data exists, but wait for data
		time.Sleep(200 * time.Millisecond)
		v, _ := obtain()
		logger.FatalfIf(v != value, "case-exists: expect %s, but got %s", value, v)

		// case-exists: data exists
		v, _ = obtain()
		logger.FatalfIf(v != value, "case-exists: expect %s, but got %s", value, v)

		// case-delayDeleteQuery1: data exists, but delay deleted so return old value and get new data async
		err = dc.Delete(k)
		logger.FatalIfError(err)
		v, used := obtain()
		logger.FatalfIf(v != "value1", "case-delayDeleteQuery1: expect %s, but got %s", "value1", v)
		logger.FatalfIf(used > 0, "case-delayDelete: expect 0, but got %d", used)

		// case-delayDeleteQuery2: data exists, but delay deleted and locked return old value
		v, used = obtain()
		logger.FatalfIf(v != "value1", "case-delayDeleteQuery2: expect %s, but got %s", "value1", v)
		logger.FatalfIf(used > 0, "case-delayDelete: expect 0, but got %d", used)

		// case-delayDeleteQuery3: data already replaced by new data
		time.Sleep(1 * time.Second)
		v, used = obtain()
		logger.FatalfIf(v != value, "case-delayDeleteQuery3: expect %s, but got %s", value, v)
		logger.FatalfIf(used > 0, "case-delayDeleteQuery3: expect 0, but got %d", used)

		// case-delayDeleteVersionBug
		dc = delay.NewClient(rdb, 3) // make value key expire in 3s. which is less than getData2.
		err = dc.Delete(k)
		logger.FatalIfError(err)
		go func() {
			v, err := dc.Obtain(k, 86400, 4, getData2)
			logger.FatalIfError(err)
			logger.FatalfIf(v != "value2", "case-delayDeleteVersionBugFirstObtain: expect %s, but got %s", "value2", v)
		}()
		time.Sleep(200 * time.Millisecond)
		err = dc.Delete(k)
		logger.FatalIfError(err)
		v, used = obtain()
		logger.FatalfIf(v != "value2", "case-delayDeleteVersionBugSecondObtain: expect %s, but got %s", "value2", v)
		time.Sleep(4 * time.Second)
		v, used = obtain()
		logger.FatalfIf(v != "value3", "case-delayDeleteVersionBugSecondObtain: expect %s, but got %s", "value3", v)
		logger.Infof("finally, intv is: %d, but value in cache is: %s, they are not matched", intv, v)
		return v
	}))

}
