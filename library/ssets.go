package library

import (
	"sync"

	"github.com/yamakiller/magicNet/engine/actor"
)

//NewSSets desc
//@method NewSSets
func NewSSets() *SSets {
	return &SSets{s: make(map[string]actor.PID)}
}

//SSets desc
//@struct SSets desc: service pid manages
type SSets struct {
	s map[string]actor.PID
	sync.Mutex
}

//Get desc
//@method Get desc: return a *acotr pid
//@param (string) serivce name
//@return (*actor.PID) a acotr of pid
func (slf *SSets) Get(name string) *actor.PID {
	slf.Lock()
	defer slf.Unlock()

	if d, ok := slf.s[name]; ok {
		return &d
	}

	return nil
}

//Push desc
//@method Push desc: a service to manager
//@param (string) 		service name
//@param (*actor.PID) 	service pid
func (slf *SSets) Push(name string, pid *actor.PID) {
	slf.Lock()
	defer slf.Unlock()
	slf.s[name] = *pid
}

//Erase desc
//@method Erase desc: delete a service
//@param (string) service name
func (slf *SSets) Erase(name string) {
	slf.Lock()
	defer slf.Unlock()
	if _, ok := slf.s[name]; ok {
		delete(slf.s, name)
		return
	}
}
