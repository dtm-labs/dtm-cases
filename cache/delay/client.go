package delay

import (
	"time"

	"github.com/dtm-labs/dtmcli/logger"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/lithammer/shortuuid"
)

type Req struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Expire int64  `json:"expire"`
}

type Client struct {
	rdb   *redis.Client
	delay int64
}

func MustReqFrom(c *gin.Context) *Req {
	var req Req
	err := c.BindJSON(&req)
	logger.FatalIfError(err)
	return &req
}

func NewClient(rdb *redis.Client, delay int64) *Client {
	return &Client{rdb, delay}
}

func now() int64 {
	return time.Now().Unix()
}

func callLua(rdb *redis.Client, script string, keys []string, args []interface{}) (interface{}, error) {
	logger.Debugf("DelayCallLua: script=%s, keys=%v, args=%v", script, keys, args)
	v, err := rdb.Eval(rdb.Context(), script, keys, args).Result()
	if err == redis.Nil {
		err = nil
	}
	return v, err
}

func (c *Client) Delete(key string) error {
	logger.Debugf("delay.Delete: key=%s", key)
	_, err := callLua(c.rdb, ` --  delay.Delete
local v = redis.call('HGET', KEYS[1], 'value')
if v == false then
	return
end
redis.call('HSET', KEYS[1], 'lockUtil', ARGV[1])
redis.call('HDEL', KEYS[1], 'lockOwner')
redis.call('EXPIRE', KEYS[1], ARGV[2])
	`, []string{key}, []interface{}{time.Now().Add(-1 * time.Second).Unix(), c.delay})
	return err
}

func (c *Client) Obtain(key string, expire int, maxCalTime int64, fn func() (string, error)) (string, error) {
	logger.Debugf("delay.Obtain: key=%s", key)
	owner := shortuuid.New()
	redisGet := func() ([]interface{}, error) {
		res, err := callLua(c.rdb, ` -- delay.Obtain
		local v = redis.call('HGET', KEYS[1], 'value')
		local lu = redis.call('HGET', KEYS[1], 'lockUtil')
		if lu ~= false and tonumber(lu) < tonumber(ARGV[1]) or lu == false and v == false then
			redis.call('HSET', KEYS[1], 'lockUtil', ARGV[1])
			redis.call('HSET', KEYS[1], 'lockOwner', ARGV[2])
			return { v, 'LOCKED' }
		end
		return {v, lu}
		`, []string{key}, []interface{}{now() + maxCalTime, owner})
		if err != nil {
			return nil, err
		}
		return res.([]interface{}), nil
	}
	r, err := redisGet()
	logger.Debugf("r is: %v", r)

	for err == nil && r[0] == nil && r[1].(string) != "LOCKED" {
		logger.Debugf("lock by other, so sleep 1s")
		time.Sleep(1 * time.Second)
		r, err = redisGet()
	}
	if err != nil {
		return "", err
	}
	if r[1] != "LOCKED" {
		return r[0].(string), nil
	}
	getNew := func() (string, error) {
		result, err := fn()
		if err != nil {
			return "", err
		}
		_, err = callLua(c.rdb, `-- delay.Set
	local o = redis.call('HGET', KEYS[1], 'lockOwner')
	if o ~= false and o ~= ARGV[2] then
			return
	end
	redis.call('HSET', KEYS[1], 'value', ARGV[1])
	redis.call('HDEL', KEYS[1], 'lockUtil')
	redis.call('HDEL', KEYS[1], 'lockOwner')
	redis.call('EXPIRE', KEYS[1], ARGV[3])
	`, []string{key}, []interface{}{result, owner, expire})
		return result, err
	}
	if r[0] == nil {
		return getNew()
	}
	go getNew()
	return r[0].(string), nil
}
