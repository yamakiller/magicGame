package library

import (
	"sync"

	"github.com/yamakiller/magicDB"
)

var (
	onceDB sync.Once
	sql    *MySQLOper
)

func sqlInstance() *MySQLOper {
	onceDB.Do(func() {
		sql = &MySQLOper{}
	})

	return sql
}

//DoMySQLDeploy desc
//@method DoMySQLDeploy desc: deploy mysql db
//@param (*MySQLDeployArray) mysql config informat
//@return (error) register mysql success/fail
func DoMySQLDeploy(sds *magicDB.MySQLDeploy) error {
	return magicDB.DoMySQLDeploy(&sqlInstance().conn, sds)
}

//MySQLOper desc
//@struct MySQLOper desc: mysql opertion
//@member (library.MySQLDB) a mysql connection pools
type MySQLOper struct {
	conn magicDB.MySQLDB
}

//MySQLClose desc
//@method MySQLClose desc: close mysql
func MySQLClose() {
	sqlInstance().conn.Close()
}

//MySQLQuery desc
//@method MySQLQuery desc: Query mysql database
//@param  (string) sql
//@param  (...interface{}) sql param
//@return (*magicDB.MySQLReader) query result
//@return (error) query error informat
func MySQLQuery(strSQL string, args ...interface{}) (*magicDB.MySQLReader, error) {
	return sqlInstance().conn.Query(strSQL, args...)
}

//MySQLInsert desc
//@method MySQLInsert desc: Mysql database insert
//@param (string) sql
//@param (...interface{}) sql param
//@return (int64) insert number
//@return (error) insert error informat
func MySQLInsert(strSQL string, args ...interface{}) (int64, error) {
	return sqlInstance().conn.Insert(strSQL, args...)
}

//MySQLUpdate desc
//@method DBUpdate desc: Mysql database update
//@param (string) sql
//@param (...interface{}) sql param
//@return (int64) update number
//@return (error) update error informat
func MySQLUpdate(strSQL string, args ...interface{}) (int64, error) {
	return sqlInstance().conn.Update(strSQL, args...)
}
