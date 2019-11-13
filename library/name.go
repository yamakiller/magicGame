package library

import (
	"strconv"
	"sync/atomic"
)

//NameFactory desc
//@struct NameFactory desc:
//@member (uint32) sequence number
type NameFactory struct {
	sid uint32
}

//Spawn desc
//@method Spawn desc: make a name
//@param  (string) base name
//@return (string) base name + sequence number
func (slf *NameFactory) Spawn(name string) string {
	newid := atomic.AddUint32(&slf.sid, 1)
	return name + strconv.FormatUint(uint64(newid), 10)
}
