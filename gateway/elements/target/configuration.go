package target

import (
	"github.com/yamakiller/magicNet/st/table"
)

const (
	constTargetMax = 2048
)

//NewTargetSet : Connection group information
func NewTargetSet() *TSet {
	r := &TSet{HashTable: table.HashTable{Mask: 0xFFFFFFFF, Max: constTargetMax, Comp: targetConnComparator}}
	r.Init()
	return r
}

func targetConnComparator(a, b interface{}) int {
	c := a.(*TConn)
	if c.virtualID == b.(uint32) {
		return 0
	}
	return 1
}

//TSet connection set
type TSet struct {
	table.HashTable
}

//Push Increase connection target
func (tset *TSet) Push(t *TConn) error {
	key, err := tset.HashTable.Push(t)
	if err != nil {
		return err
	}

	t.virtualID = key
	return nil
}

//Get Returns Target Connection Object in the Set
func (tset *TSet) Get(virtaulID uint32) *TConn {
	v := tset.HashTable.Get(virtaulID)
	if v == nil {
		return nil
	}

	return v.(*TConn)
}
