package library

import (
	"sync"

	"github.com/yamakiller/magicDB"
)

var (
	onceRedis  sync.Once
	redisArray *redisDBArray
)

//redisDBArray
//@struct redisDBArray desc:  Redis pool object
//@member (map[int]*library.RedisDB) redis connection arrays
type redisDBArray struct {
	dbs map[int]*magicDB.RedisDB
}

func (slf *redisDBArray) get(db int) *magicDB.RedisDB {
	r, success := slf.dbs[db]
	if !success {
		return nil
	}
	return r
}

func (slf *redisDBArray) close() {
	for _, v := range slf.dbs {
		v.Close()
	}
	slf.dbs = make(map[int]*magicDB.RedisDB)
}

//redisInstance desc
//@method redistInstance desc: Redis connection pool interface
//@return (*redisDBArray)
func redisInstance() *redisDBArray {
	onceRedis.Do(func() {
		redisArray = &redisDBArray{dbs: make(map[int]*magicDB.RedisDB)}
	})

	return redisArray
}

//DoRedisDeployArray desc
//@method DoRedisDeployArray desc: deploy redis db array
//@param (*RedisDeployArray) (deploy array) informat
//@return (error) if fail return error inforamt else return nil
func DoRedisDeployArray(ds *magicDB.RedisDeployArray) error {
	for _, v := range ds.Deploys {
		if err := DoRedisDeploy(&v); err != nil {
			return err
		}
	}
	return nil
}

//DoRedisDeploy desc
//@method DoRedisDeploy desc: deplay redis db
//@param (*RedisDeploy) a deploy informat
//@return (error) if fail return error inforamt else return nil
func DoRedisDeploy(deploy *magicDB.RedisDeploy) error {
	tmpdb := &magicDB.RedisDB{}
	if err := magicDB.DoRedisDeploy(tmpdb, deploy); err != nil {
		return err
	}
	redisInstance().dbs[deploy.DB] = tmpdb
	return nil
}

//RedisDo desc
//@method RedisDo desc: Execute the Redis command
//@param (string) command name
//@param (...interface{}) command params
//@return (interface{}) return execute result
func RedisDo(db int, commandName string, args ...interface{}) (interface{}, error) {
	return redisInstance().get(db).Do(commandName, args...)
}

//RedisClose desc
//@method RedisClose desc: Close the entire redis
func RedisClose() {
	redisInstance().close()
}
