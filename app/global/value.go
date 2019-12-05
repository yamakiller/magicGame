package global

import (
	"sync"

	"github.com/yamakiller/magicNet/engine/actor"
)

var (
	defaultTable *Table
	globalOne    sync.Once
)

//Instance doc
func Instance() *Table {
	globalOne.Do(func() {
		defaultTable = new(Table)
		defaultTable._handles = make(map[int]*actor.PID)
	})
	return defaultTable
}

//Table doc
//@Summary global var table
//@Struct Table
//@Member NetworkServiceHandle network listen serivce pid
//@Member NetworkRouterServiceHandle network router service pid
type Table struct {
	_handles map[int]*actor.PID
	_sync    sync.RWMutex
}

//RegisterHandle doc
//@Summary register service pid
//@Member RegisterHandle
//@Param int key doc to global tag.go
//@Param *actor.PID
func (slf *Table) RegisterHandle(key int, pid *actor.PID) {
	slf._sync.Lock()
	defer slf._sync.Unlock()
	slf._handles[key] = pid
}

//UnRegisterHandle doc
//@Summary unregister service pid
//@Member UnRegisterHandle
//@Param int key doc to global tag.go
func (slf *Table) UnRegisterHandle(key int) {
	slf._sync.Lock()
	defer slf._sync.Unlock()
	if _, ok := slf._handles[key]; ok {
		delete(slf._handles, key)
	}
}

//GetHandle doc
//@Summary Return service handle/pid
//@Member GetHandle
//@Param int key doc to global tag.go
//@Return *actor.PID
func (slf *Table) GetHandle(key int) *actor.PID {
	slf._sync.RLock()
	defer slf._sync.RUnlock()
	if v, ok := slf._handles[key]; ok {
		return v
	}
	return nil
}
