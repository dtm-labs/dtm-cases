package demo

import (
	"database/sql"
	"fmt"

	"github.com/dtm-labs/dtm-cases/cache/delay"
	"github.com/dtm-labs/dtmcli/logger"
	"github.com/go-redis/redis/v8"
)

const dbKey = 1

var rdbKey = "t1-id-1"

var redisOption = redis.Options{
	Addr:     "dtm.pub:6379",
	Username: "root",
	Password: "",
}

var rdb = redis.NewClient(&redisOption)

var dc = delay.NewClient(rdb, 3, 30)

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("mysql", "dtm:passwd123dtm@tcp(dtm.pub:3306)/cache1?charset=utf8")
	logger.FatalIfError(err)
}

func updateDB(value string) {
	_, err := db.Exec("insert into cache1.t1(id, value) values(?, ?) on duplicate key update value=values(value)", dbKey, value)
	logger.FatalIfError(err)
	logger.Infof("update db: %s", value)
}

func updateInTx(tx *sql.Tx, value string) error {
	_, err := tx.Exec("insert into cache1.t1(id, value) values(?, ?) on duplicate key update value=values(value)", dbKey, value)
	logger.Infof("updateInTx: %s", value)
	return err
}

func getDB() string {
	var value string
	err := db.QueryRow("select value from cache1.t1 where id=?", dbKey).Scan(&value)
	logger.FatalIfError(err)
	logger.Infof("get db: %s", value)
	return value
}

func obtainValue() string {
	v, err := rdb.Get(rdb.Context(), rdbKey).Result()
	if err == redis.Nil {
		value := getDB()
		_, err := rdb.Set(rdb.Context(), rdbKey, value, 0).Result()
		logger.FatalIfError(err)
		logger.Infof("obtainValue: %s", value)
		return value
	}
	logger.FatalIfError(err)
	logger.Infof("obtainValue: %s", v)
	return v
}

func ensure(condition bool, format string, v ...interface{}) {
	hint := "ok"
	if !condition {
		hint = "failed"
	}
	logger.Infof("ensure: %s for %s", hint, fmt.Sprintf(format, v...))
	if !condition {
		panic("ensure failed")
	}
}
