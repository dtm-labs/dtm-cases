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

var vdc = delay.NewVersionClient(rdb, 10)

func verUpdateValue(version int, nvalue string) {
	msg := dtmcli.NewMsg(DtmServer, shortuuid.New()).
		Add(BusiUrl+"/delayDelete", &delay.Req{Key: k})
	msg.WaitResult = true // wait for result. when submit returned without error, cache will be deleted
	err := msg.DoAndSubmitDB(BusiUrl+"/delayQueryPrepared", db, func(tx *sql.Tx) error {
		_, err := db.Exec("insert into cache1.ver(id, value, version) values(?, ?, ?) on duplicate key update value=values(value), version=values(version)", 1, nvalue, version)
		logger.FatalIfError(err)
		return nil
	})
	logger.FatalIfError(err)
}

func verGetData1() (int, string, error) {
	var v string
	var ver int
	err := db.QueryRow("select value, version from cache1.ver where id = ?", 1).Scan(&v, &ver)
	logger.FatalIfError(err)
	logger.Infof("get Data sleeping 1s")
	time.Sleep(1 * time.Second)
	return ver, v, nil
}

func verGetData2() (int, string, error) {
	ver, r, err := verGetData1()
	logger.Infof("get Data sleeping 3s")
	time.Sleep(3 * time.Second)
	return ver, r, err
}

func verObtain() (result string, used int) {
	begin := time.Now()
	v, err := vdc.Obtain(k, 86400, 3, verGetData1)
	logger.FatalIfError(err)
	return v, int(time.Since(begin).Seconds())
}

func addVersionDelayDelete(app *gin.Engine) {
	app.GET(BusiAPI+"/delayDeleteBug", utils.WrapHandler(func(c *gin.Context) interface{} {
		updateValue("value1")
		expected := "value1"
		// case-empty: no data exists
		_, err := rdb.Del(rdb.Context(), k).Result()
		logger.FatalIfError(err)
		v, _ := obtain()
		logger.FatalfIf(v != expected, "case-empty: expect %s, but got %s", expected, v)

		updateValue("value2")
		v, err = dc.Obtain(k, 86400, 4, getData2)
		logger.FatalIfError(err)
		logger.FatalfIf(v != expected, "case-delayDeleteVersionBugFirstObtain: expect %s, but got %s", expected, v)
		time.Sleep(200 * time.Millisecond) // wait for getData2 to finish intv update
		updateValue("value3")
		logger.FatalIfError(err)
		v, _ = obtain()
		logger.FatalfIf(v != expected, "case-delayDeleteVersionBugSecondObtain: expect %s, but got %s", expected, v)
		time.Sleep(4 * time.Second)
		expected = "value2"
		v, _ = obtain()
		logger.FatalfIf(v != expected, "case-delayDeleteVersionBugSecondObtain: expect %s, but got %s", expected, v)
		msg := fmt.Sprintf("finally, value is: %s, but value in cache is: %s, they are not matched", "value3", v)
		logger.Infof(msg)
		return msg
	}))

	app.GET(BusiAPI+"/delayDeleteVersion", utils.WrapHandler(func(c *gin.Context) interface{} {
		verUpdateValue(1, "value1")
		expected := "value1"

		_, err := rdb.Del(rdb.Context(), k).Result()
		logger.FatalIfError(err)
		v, _ := verObtain()
		logger.FatalfIf(v != expected, "case-empty: expect %s, but got %s", expected, v)

		verUpdateValue(2, "value2")
		v, err = vdc.Obtain(k, 86400, 4, verGetData2)
		logger.FatalIfError(err)
		logger.FatalfIf(v != expected, "case-delayDeleteVersionBugFirstObtain: expect %s, but got %s", expected, v)
		time.Sleep(200 * time.Millisecond) // wait for getData2 to finish intv update
		verUpdateValue(3, "value3")
		logger.FatalIfError(err)
		v, _ = verObtain()
		logger.FatalfIf(v != expected, "case-delayDeleteVersionBugSecondObtain: expect %s, but got %s", expected, v)
		time.Sleep(4 * time.Second)
		expected = "value3"
		v, _ = verObtain()
		logger.FatalfIf(v != expected, "case-delayDeleteVersionBugSecondObtain: expect %s, but got %s", expected, v)
		msg := fmt.Sprintf("finally, value is: %s, and value in cache is: %s, they are matched", expected, v)
		logger.Infof(msg)
		return msg
	}))
}
