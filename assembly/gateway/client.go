package gateway

import (
	"reflect"

	"github.com/yamakiller/magicLibs/encryption/dh64"

	"github.com/yamakiller/magicNet/network"

	"github.com/yamakiller/magicNet/engine/actor"
	srvc "github.com/yamakiller/magicNet/handler/implement/client"
)

type client struct {
	srvc.NetSSrvCleint
	_parent       *Server
	_handle       uint64
	_authLastTime int64
	_auth         int64
	_prvKey       uint64
	_secert       uint64
}

//Initial doc
//@Summary Initial gateway server accesser
//@Method Initial
func (slf *client) Initial() {
	slf.NetSSrvCleint.Initial()
	slf.RegisterMethod(&AgreMessage{}, slf.onAgreement)
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

//KeyPair doc
//@Summary Build Key Pair
//@Method KeyPair
//@Return publicKey
func (slf *client) KeyPair() uint64 {
	prvKey, publicKey := dh64.KeyPair()
	slf._prvKey = prvKey
	return publicKey
}

//BuildSecert doc
//@Summary Build Secert
//@Method BuildSecert
func (slf *client) BuildSecert(publicKey uint64) {
	slf._secert = dh64.Secret(slf._prvKey, publicKey)
}

//Secert doc
//@Summary Return client secert
//@Method Secert
//@Return secert uint64
func (slf *client) Secert() uint64 {
	return slf._secert
}

func (slf *client) onAgreement(context actor.Context, sender *actor.PID, message interface{}) {
	req := message.(*AgreMessage)
	addr, m, isauth, err := slf._parent._delegate.QueryAsyncMethod(req.Agreement)
	if err != nil {
		slf.LogError("client %s => %d %+v", slf.GetAddr(), slf.GetSocket(), err)
		return
	}

	if isauth {
		if slf._auth <= 1 {
			slf.LogError("client %s => %d %+v agreement need Authorization", slf.GetAddr(), slf.GetSocket(), req.Agreement)
			network.OperClose(slf.GetSocket())
			return
		}
	}

	params := make([]reflect.Value, 2)
	params[0] = reflect.ValueOf(addr)
	params[1] = reflect.ValueOf(req.AgreementData)

	rs := reflect.ValueOf(m).Call(params)
	if len(rs) > 0 {
		numrs := len(rs)
		if err, ok := rs[numrs].Interface().(error); ok {
			if err != nil {
				slf.LogError("client %s => %d %+v", slf.GetAddr(), slf.GetSocket(), err)
				return
			}

			if len(rs) == 1 {
				return
			}
		}

		d, err := slf._parent._delegate.AsyncEncode(rs[0].Interface())
		if err != nil {
			slf.LogError("client %s => %d %+v", slf.GetAddr(), slf.GetSocket(), err)
			return
		}

		if err = slf.SendTo(d); err != nil {
			slf.LogError("client %s => %d %+v", slf.GetAddr(), slf.GetSocket(), err)
			return
		}
	}

}

func (slf *client) Shutdown() {
	slf.NetSSrvCleint.Shutdown()
	slf._auth = 0
	slf._handle = 0
	slf._parent = nil
	slf._authLastTime = 0
}
