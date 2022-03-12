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

func checkStatusCompatible(opSwitch string, doCache bool) {
	if opSwitch == "none" {
		logger.FatalfIf(doCache, "opSwitch is none, doCache should be false")
	}
	if opSwitch == "full" {
		logger.FatalfIf(!doCache, "opSwitch is full, doCache should be true")
	}
}

func strongWrite(value string, confWriteCache string, writeCache bool) {
	checkStatusCompatible(confWriteCache, writeCache)
	if !writeCache {
		updateDB(value)
		return
	}
	msg := dtmcli.NewMsg(DtmServer, shortuuid.New()).
		Add(BusiUrl+"/dtmDelKey", &delay.Req{Key: key})
	msg.TimeoutToFail = 3

	err := msg.DoAndSubmit(BusiUrl+"/dtmQueryPrepared", func(bb *dtmcli.BranchBarrier) error {
		return bb.CallWithDB(db, func(tx *sql.Tx) error {
			_, err := tx.Exec("insert into cache1.t1(id, value) values(?, ?) on duplicate key update value=values(value)", 1, value)
			return err
		})
	})
	logger.FatalIfError(err)
}

func strongRead(confReadCache string, readCache bool) string {
	checkStatusCompatible(confReadCache, readCache)
	if !readCache {
		return queryValue()
	}
	sc := delay.NewClient(rdb, 10, 30)
	r, err := sc.StrongObtain(key, 600, 3, func() (string, error) {
		return queryValue(), nil
	})
	logger.FatalIfError(err)
	return r
}

func addStrongConsistency(app *gin.Engine) {
	app.GET(BusiAPI+"/strongConsistencyDemo", utils.WrapHandler(func(c *gin.Context) interface{} {
		// set up
		// none: all read from db
		// partial: some read from db, some read from cache.
		// full: all read from cache
		var confReadCache = "none"

		// none: all write only db
		// partial: some write only db, some write both db and cache
		// full: all write both db and cache
		var confWriteCache = "none"
		expected := "value1"

		// 准备升级
		confWriteCache = "partial" // 打开写缓存开关，在分布式应用中，配置会逐步在各个进程生效。
		strongWrite(expected, confWriteCache, true)

		expected = "value2"
		strongWrite(expected, confWriteCache, false)

		v := strongRead("parital", true) // 如果此时错误的打开了读缓存，那么部分请求会读取到缓存中的脏数据，导致 v != expected
		logger.FatalfIf(v == expected, "strongRead('parital', true) expected: %s, actual: %s", expected, v)

		time.Sleep(2 * time.Second)
		confWriteCache = "full" // 写缓存的升级已全部完成，所有的写都会写DB+缓存

		confReadCache = "patial"            // 打开读缓存开关
		v = strongRead(confReadCache, true) // 此时读取缓存，能够读取缓存中的正确数据
		logger.FatalfIf(v != expected, "strongRead('parital', true) expected: %s, actual: %s", expected, v)
		time.Sleep(2 * time.Second)
		confReadCache = "full" // 读缓存的升级完成
		// 升级完成

		// 运行一段时间后，Redis出现故障，现在需要降级
		confReadCache = "patial" // 关闭读缓存开关，在分布式应用中，配置会逐步在各个进程生效。
		expected = "value3"

		strongWrite(expected, "partial", false) // 如果此时错误的关闭了写缓存，那么部分请求会只写DB
		v = strongRead(confReadCache, true)     // 此时部分请求会读取到缓存中的脏数据，导致 v != expected
		logger.FatalfIf(v == expected, "strongRead('parital', true) expected: %s, actual: %s", expected, v)

		time.Sleep(2 * time.Second)
		confReadCache = "none" // 关闭读缓存开关，所有进程上都已关闭，所有读都会从DB中读取

		v = strongRead(confReadCache, false) // 此时所有的读都从DB中读取，不会读取到脏数据
		logger.FatalfIf(v != expected, "strongRead('none', false) expected: %s, actual: %s", expected, v)

		confWriteCache = "partial" // 关闭写缓存开关，在分布式应用中，配置会逐步在各个进程生效。
		time.Sleep(2 * time.Second)
		confWriteCache = "none" // 关闭写缓存开关，所有进程上都已关闭，所有写都会只写DB
		return "finished"
	}))
}
