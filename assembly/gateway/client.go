package gateway

import (
	srvc "github.com/yamakiller/magicNet/handler/implement/client"
)

type client struct {
	srvc.NetSSrvCleint
	_handle uint64
	_auth   int64
}

//Initial doc
//@Summary Initial gateway server accesser
//@Method Initial
func (slf *client) Initial() {
	slf._auth = 0
	slf.NetSSrvCleint.Initial()
}

//SetID doc
//@Summary Setting handle/id
//@Method SetID
//@Param uint64  handle/id
func (slf *client) SetID(id uint64) {
	slf._handle = id
}

//GetID doc
//@Summary Returns handle/id
//@Method GetID
//@Return uint64
func (slf *client) GetID() uint64 {
	return slf._handle
}
