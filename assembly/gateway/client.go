package gateway

import (
	"reflect"

	"github.com/yamakiller/magicNet/network"

	"github.com/yamakiller/magicNet/engine/actor"
	"github.com/yamakiller/magicNet/handler/encryption"
	srvc "github.com/yamakiller/magicNet/handler/implement/client"
)

type client struct {
	srvc.NetSSrvCleint
	_parent       *Server
	_handle       uint64
	_authLastTime int64
	_auth         int64
	_prvKey       uint64
	_encrypt      encryption.INetEncryption
}

//Initial doc
//@Summary Initial gateway server accesser
func (slf *client) Initial() {
	slf.NetSSrvCleint.Initial()
	slf.RegisterMethod(&AgreMsg{}, slf.onAgreement)
}

//WithID doc
//@Summary Set handle/id
//@Param uint64  handle/id
func (slf *client) WithID(id uint64) {
	slf._handle = id
}

//WithEncrypt doc
//@Summary Set Encryptor
func (slf *client) WithEncrypt(encrypt encryption.INetEncryption) {
	slf._encrypt = encrypt
}

//WithPrvKey doc
//@Summary Set private key
func (slf *client) WithPrvKey(prvKey uint64) {
	slf._prvKey = prvKey
}

//GetPrvKey doc
//@Summary Return private key
func (slf *client) GetPrvKey() uint64 {
	return slf._prvKey
}

//GetID doc
//@Summary Returns handle/id
//@Return uint64
func (slf *client) GetID() uint64 {
	return slf._handle
}

//Encrypt doc
//@Summary Returns a encryptor
func (slf *client) Encrypt() encryption.INetEncryption {
	return slf._encrypt
}

//Secert doc
//@Summary Return client secert
//@Return secert uint64
/*func (slf *client) Secert() uint64 {
	return slf._secert
}*/

func (slf *client) onAgreement(context actor.Context, sender *actor.PID, message interface{}) {
	req := message.(*AgreMsg)
	addr, m, isauth, err := slf._parent._delegate.QueryLocalAgreement(req.Agreement)
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

		d, err := slf._parent._delegate.AsyncEncode(slf, rs[0].Interface())
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
	if slf._encrypt != nil {
		slf._encrypt.Destory()
		slf._encrypt = nil
	}

	slf._auth = 0
	slf._handle = 0
	slf._parent = nil
	slf._authLastTime = 0
}
