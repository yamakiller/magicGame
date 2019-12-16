package gateway

import (
	"encoding/binary"
	"reflect"

	"github.com/yamakiller/magicLibs/encryption/dh64"

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
	_secert       []byte
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
func (slf *client) WithEncrypt(encrypt encryption.INetEncryption) error {
	slf._encrypt = encrypt
	return slf._encrypt.Cipher(slf._secert)
}

//GetID doc
//@Summary Returns handle/id
//@Return uint64
func (slf *client) GetID() uint64 {
	return slf._handle
}

//KeyPair doc
//@Summary Build Key Pair
//@Return publicKey
func (slf *client) KeyPair() uint64 {
	prvKey, publicKey := dh64.KeyPair()
	slf._prvKey = prvKey
	return publicKey
}

//BuildSecert doc
//@Summary Build Secert
func (slf *client) BuildSecert(publicKey uint64) {
	secert := dh64.Secret(slf._prvKey, publicKey)
	if slf._secert == nil {
		slf._secert = make([]byte, 8)
	}
	binary.BigEndian.PutUint64(slf._secert, secert)
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
