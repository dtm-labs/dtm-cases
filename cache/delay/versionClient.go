package delay

import (
	"time"

	"github.com/dtm-labs/dtmcli/logger"
	"github.com/go-redis/redis/v8"
)

type VersionClient struct {
	c *Client
}

func NewVersionClient(rdb *redis.Client, delay int64) *VersionClient {
	return &VersionClient{c: NewClient(rdb, delay)}
}

func (s *VersionClient) Delete(key string) error {
	return s.c.Delete(key)
}

func (s *VersionClient) Obtain(key string, expire int, maxCalTime int64, fn func() (int, string, error)) (string, error) {
	c := s.c
	redisGet := func() ([]interface{}, error) {
		res, err := callLua(c.rdb, ` -- delay.Obtain
		local v = redis.call('HGET', KEYS[1], 'value')
		local lu = redis.call('HGET', KEYS[1], 'lockUtil')
		if lu ~= false and tonumber(lu) < tonumber(ARGV[1]) or lu == false and v == false then
			redis.call('HSET', KEYS[1], 'lockUtil', ARGV[1])
			return { v, 'LOCKED' }
		end
		return {v, lu}
		`, []string{key}, []interface{}{now() + maxCalTime})
		if err != nil {
			return nil, err
		}
		return res.([]interface{}), nil
	}
	r, err := redisGet()
	logger.Debugf("r is: %v", r)

	for err == nil && r[0] == nil && r[1].(string) != "LOCKED" {
		logger.Debugf("locked by other, so sleep 1s")
		time.Sleep(1 * time.Second)
		r, err = redisGet()
	}
	if err != nil {
		return "", err
	}
	if r[1] != "LOCKED" {
		return r[0].(string), nil
	}
	getNew := func() (int, string, error) {
		ver, result, err := fn()
		if err != nil {
			return 0, "", err
		}
		_, err = callLua(c.rdb, `-- delay.Set
	local o = redis.call('HGET', KEYS[1], 'version')
	if o ~= false and tonumber(o) >= tonumber(ARGV[2]) then
			return
	end
	redis.call('HSET', KEYS[1], 'value', ARGV[1])
	redis.call('HSET', KEYS[1], 'version', ARGV[2])
	redis.call('HDEL', KEYS[1], 'lockUtil')
	redis.call('EXPIRE', KEYS[1], ARGV[3])
	`, []string{key}, []interface{}{result, ver, expire})
		return ver, result, err
	}
	if r[0] == nil {
		_, v, err := getNew()
		return v, err
	}
	go getNew()
	return r[0].(string), nil
}
