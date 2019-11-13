package parts

import "sync"

//RobotWorld desc
//@struct RobotWorld desc: robot world/manage all robot
type RobotWorld struct {
	worlds map[string]IRobotPlay
	syn    sync.Mutex
}

//Enter desc
//@method Enter desc: enter world
//@param (string) robot name
//@param (IRobotPlay) robot
func (slf *RobotWorld) Enter(name string, play IRobotPlay) {
	slf.syn.Lock()
	defer slf.syn.Unlock()
	slf.worlds[name] = play
}

//Leave desc
//@method Leave desc: Leave world
//@param (string) robot name
func (slf *RobotWorld) Leave(name string) {
	slf.syn.Lock()
	defer slf.syn.Unlock()

	if _, ok := slf.worlds[name]; ok {
		delete(slf.worlds, name)
	}
}

//get desc
//@method get desc: Return world IRobotPlay
//@param (string) Robot play name
//@return (IRobotPlay)
func (slf *RobotWorld) get(name string) IRobotPlay {
	slf.syn.Lock()
	defer slf.syn.Unlock()
	if v, ok := slf.worlds[name]; ok {
		return v
	}

	return nil
}

//getKeys desc
//@method getKeys desc: Return world Robot play all name
//@return ([]string)
func (slf *RobotWorld) getKeys() []string {
	slf.syn.Lock()
	defer slf.syn.Unlock()
	i := 0
	r := make([]string, len(slf.worlds))
	for k := range slf.worlds {
		r[i] = k
		i++
	}

	return r
}

//Tick desc
//@method Tick desc: tick world all robot
func (slf *RobotWorld) Tick(delta int64) {
	keys := slf.getKeys()
	for _, k := range keys {
		play := slf.get(k)
		play.Tick(delta)
	}
}
