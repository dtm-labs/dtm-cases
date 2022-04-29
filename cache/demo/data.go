package demo

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/dtm-labs/dtmcli/logger"
	"github.com/dtm-labs/rockscache"
	"github.com/gin-gonic/gin"
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

var dc = rockscache.NewClient(rdb, rockscache.NewDefaultOptions())

var db *sql.DB

func init() {
	var err error
	db, err = sql.Open("mysql", "dtm:passwd123dtm@tcp(dtm.pub:3306)/cache1?charset=utf8")
	logger.FatalIfError(err)
}

// Req is request
type Req struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Expire int64  `json:"expire"`
}

// MustReqFrom gin.Context to Req
func MustReqFrom(c *gin.Context) *Req {
	var req Req
	err := c.BindJSON(&req)
	logger.FatalIfError(err)
	return &req
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

func getDB() (string, error) {
	var value string
	err := db.QueryRow("select value from cache1.t1 where id=?", dbKey).Scan(&value)
	logger.Infof("get db: %s, %v", value, err)
	return value, err
}

func clearCache() {
	err := rdb.Del(rdb.Context(), rdbKey).Err()
	logger.FatalIfError(err)
}

func obtainValue() string {
	v, err := rdb.Get(rdb.Context(), rdbKey).Result()
	if err == redis.Nil {
		value, err := getDB()
		logger.FatalIfError(err)
		_, err = rdb.Set(rdb.Context(), rdbKey, value, 0).Result()
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
		os.Exit(1)
	}
}
