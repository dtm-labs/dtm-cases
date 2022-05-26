English | [简体中文](./README-cn.md)

# Cache Applications
This project can be read in conjunction with the [Caching Applications](https://en.dtm.pub/app/cache.html) in the dtm documentation

## Overview
This project demonstrates how dtm can be used to maintain cache consistency, in the following ways
- Consistency assurance: if a process crash occurs, the cache can be updated to ensure that the DB is consistent with the cache eventually
- Delayed delete: demonstrates a cache management approach to delayed delete that is hit-proof
- Strong Consistency Usage: demonstrates how to ensure strong consistency if the application accesses data in a cache downgrade or upgrade

## Start dtm
[Quick start dtm](https://en.dtm.pub/guide/install.html)

## Run this example
`go run main.go`

## Ensure consistency
The code is mainly in demo/api-crash, the example mainly demonstrates writing DBs, the normal write method for deleting cache and the dtm suggested write method
- Initiate a normal write request, no crash, consistent data `curl http://localhost:8081/api/busi/normalUpdate?crash=`
- Initiate a normal write request that simulates a crash and causes the DB to be inconsistent with the cache `curl http://localhost:8081/api/busi/normalUpdate?crash=1`
- Initiate a dtm write request, no crash, consistent data `curl http://localhost:8081/api/busi/dtmUpdate?crash=`
- Initiate a dtm write request, simulating a crash, resulting in the DB inconsistent with the cache, but after several seconds the DB is consistent with the cache `curl http://localhost:8081/api/busi/dtmUpdate?crash=1`

### Delayed deletion

#### Delayed deletion of library
The code is in delay/client.go and the main functions are described as follows.
- `func NewClient(rdb *redis.Client) *Client` creates a delayed delete client
- `func (c *Client) Delete(key string) error` Delay deleting a piece of data
- ` func (c *Client) Obtain(key string, expire int, fn func() (string, error)) (string, error) ` Get data

The Obtain function is used to fetch data and the parameters are described as follows.
- key the key of the cache
- expire the expiry time of the cache
- fn Callback function to fetch data if the cache does not hit

#### Delayed delete use case
The code is mainly in demo/api-eventual and the examples demonstrate the various features of the delayed delete method.

`curl http://localhost:8081/api/busi/eventualCases`

The code explains the following.
- case-empty: when the data is empty, getData1 will be called to get the data, taking 1s
- case-emptyWait: when the data is empty and the cache is locked, it will sleep and wait for the cached data
- case-exists: the data exists and is returned normally
- case-delayDeleteQuery1: data exists but has been delay-deleted, an asynchronous concurrent process is started to compute the new result and immediately return the old value from the cache
- case-delayDeleteQuery2: the data exists, but has been delay-deleted and is locked, the new result is not calculated and the old value is returned immediately from the cache
- case-delayDeleteQuery3: the data exists and the concurrent process in Query1 has calculated the new result and updated it to the cache, returning the new value in the cache

### Strongly consistent access

#### Strongly Consistent Cache Library
The code is in delay/client.go and the main functions are described as follows.
- `func (c *Client) StrongObtain(key string, expire int, fn func() (string, error)) (string, error) ` Get data

#### Strong Consistent Access use case
The code is mainly in demo/api-strong, and the examples demonstrate the various features of the strong consistent access, run the strong consistent access test case with the following code

`curl http://localhost:8081/api/busi/strongDemo`

The code explains the following.
- The initial state does not use cache, just DB
- The write cache is turned on, where some write requests are to DB only, and some write requests are to both cache and DB
- The write cache switch is active for all applications, and the read cache switch is turned on to ensure that the cache reads the correct results
- The read cache switch is valid for all applications
- If a cache failure occurs, downgrade
- Turn off the read cache switch so that some of the read requests read the cache and some read the DB
- The switch not to read cache is in effect for all applications, where all reads are DB only
- Turn off the write cache switch
- The switch not to write cache is in effect for all applications, when the cache is no longer being used and can be taken offline
