package demo

import (
	"database/sql"
	"time"

	"github.com/dtm-labs/dtm-cases/utils"
	"github.com/dtm-labs/dtmcli"
	"github.com/dtm-labs/dtmcli/logger"
	"github.com/dtm-labs/rockscache"
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
		Add(BusiUrl+"/delayDeleteKey", &Req{Key: rdbKey})
	msg.TimeoutToFail = 3

	err := msg.DoAndSubmit(BusiUrl+"/queryPrepared", func(bb *dtmcli.BranchBarrier) error {
		return bb.CallWithDB(db, func(tx *sql.Tx) error {
			return updateInTx(tx, value)
		})
	})
	logger.FatalIfError(err)
}

func strongRead(confReadCache string, readCache bool) string {
	checkStatusCompatible(confReadCache, readCache)
	if !readCache {
		v, err := getDB()
		logger.FatalIfError(err)
		return v
	}
	options := rockscache.NewDefaultOptions()
	options.StrongConsistency = true
	sc := rockscache.NewClient(rdb, options)
	r, err := sc.Fetch(rdbKey, 600, func() (string, error) {
		return getDB()
	})
	logger.FatalIfError(err)
	return r
}

func addStrongConsistency(app *gin.Engine) {
	app.GET(BusiAPI+"/strongDemo", utils.WrapHandler(func(c *gin.Context) interface{} {
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

		// begin to upgrade to use cache
		confWriteCache = "partial" // begin to switch to write to cache. In a distributed app, this switch will take effect gradually
		strongWrite(expected, confWriteCache, true)
		clearCache()
		eventualObtain() // simulate a read. it will populate cache.

		expected = "value2"
		strongWrite(expected, confWriteCache, false)

		v := strongRead("parital", true) // error! if you errorly switch to read from cache, some request will read error data
		ensure(v != expected, "upgrading bug occur partial-write-partial-read: expecting v != expected, v=%s, expected=%s", v, expected)

		time.Sleep(2 * time.Second)
		confWriteCache = "full" // finish to switch to write to cache. all writes will be written to cache now.
		strongWrite(expected, confWriteCache, true)

		confReadCache = "patial"            // begin to switch to read from cache. In a distributed app, this switch will take effect gradually
		v = strongRead(confReadCache, true) // now read from cache, all reads are ok
		ensure(v == expected, "full-write-partial-read: expecting v == expected, v=%s, expected=%s", v, expected)
		time.Sleep(2 * time.Second)
		confReadCache = "full" // finish to switch to read from cache
		// upgrade to use cache ok

		// if redis got some problem, now we need to downgrade to not use cache
		confReadCache = "patial" // begin to switch to read not from cache. In a distributed app, this switch will take effect gradually
		expected = "value3"

		strongWrite(expected, "partial", false) // error! if you errorly switch not to write cache, some request will read error data
		v = strongRead(confReadCache, true)
		ensure(v != expected, "downgrading bug occur partial-read-patial-write: expecting v != expected, v=%s, expected=%s", v, expected)

		time.Sleep(2 * time.Second)
		confReadCache = "none" // finish to switch to read only from db

		v = strongRead(confReadCache, false) // all reads are from db, ok
		ensure(v == expected, "none-read-partial-write: expecting v == expected, v=%s, expected=%s", v, expected)

		confWriteCache = "partial" // begin to switch to write only to db, In a distributed app, this switch will take effect gradually
		time.Sleep(2 * time.Second)
		confWriteCache = "none" // finish to switch to write only to db
		return "finished"
	}))
}
