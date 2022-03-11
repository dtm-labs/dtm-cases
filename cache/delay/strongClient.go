package delay

import (
	"time"

	"github.com/dtm-labs/dtmcli/logger"
	"github.com/go-redis/redis/v8"
	"github.com/lithammer/shortuuid"
)

type StrongClient struct {
	c *Client
}

func NewStrongClient(rdb *redis.Client, delay int, emptyExpire int) *StrongClient {
	return &StrongClient{c: NewClient(rdb, delay, emptyExpire)}
}

func (s *StrongClient) Delete(key string) error {
	return s.c.Delete(key)
}

func (c *StrongClient) Obtain(key string, expire int, maxCalTime int, fn func() (string, error)) (string, error) {
	logger.Debugf("delay.Obtain: key=%s", key)
	owner := shortuuid.New()
	redisGet := func() ([]interface{}, error) {
		res, err := callLua(c.c.rdb, ` -- delay.Obtain
		local v = redis.call('HGET', KEYS[1], 'value')
		local lu = redis.call('HGET', KEYS[1], 'lockUtil')
		if lu ~= false and tonumber(lu) < tonumber(ARGV[1]) or lu == false and v == false then
			redis.call('HSET', KEYS[1], 'lockUtil', ARGV[1])
			redis.call('HSET', KEYS[1], 'lockOwner', ARGV[2])
			return { v, 'LOCKED' }
		end
		return {v, lu}
		`, []string{key}, []interface{}{now() + int64(maxCalTime), owner})
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
		if result == "" {
			expire = c.c.EmptyExpire
		}
		_, err = callLua(c.c.rdb, `-- delay.Set
	local o = redis.call('HGET', KEYS[1], 'lockOwner')
	if o ~= ARGV[2] then
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
