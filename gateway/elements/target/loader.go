package target

import (
	"github.com/yamakiller/magicNet/engine/actor"
	"github.com/yamakiller/magicNet/st/hash"
)

//NewLoadSet Assign a load set
func NewLoadSet() *TLoadSet {
	return &TLoadSet{set: make(map[string]*TLoader)}
}

//NewLoader Assign an equalizer
func NewLoader(replicas int) *TLoader {
	return &TLoader{Map: *hash.New(replicas)}
}

//TargeObject target
type TObject struct {
	ID     uint32
	Target *actor.PID
}

//TargetLoadSet Target load set
type TLoadSet struct {
	set map[string]*TLoader
}

//Add Add a cluster type
func (tlset *TLoadSet) Add(name string, tl *TLoader) {
	tlset.set[name] = tl
}

//Get Return to a cluster
func (tlset *TLoadSet) Get(name string) *TLoader {
	if v, ok := tlset.set[name]; ok {
		return v
	}
	return nil
}

//TargetLoader Provide load balancing management for servers
type TLoader struct {
	hash.Map
}

//AddTarget Join a target service
func (t *TLoader) AddTarget(key string, v *TObject) {
	t.Lock()
	defer t.Unlock()
	t.UnAdd(key, v)
}

//RemoveTarget Delete a target service
func (t *TLoader) RemoveTarget(key string) {
	t.Lock()
	defer t.Unlock()
	t.UnRemove(key)
}

//GetTarget Return Return a service target
func (t *TLoader) GetTarget(key string) *TObject {
	t.RLock()
	defer t.RUnlock()
	r, err := t.UnGet(key)
	if err != nil {
		return nil
	}
	return r.(*TObject)
}
