package test

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"testing"

	"github.com/yamakiller/magicGame/assembly/gateway"
	"github.com/yamakiller/magicNet/handler/encryption"

	"github.com/yamakiller/magicLibs/encryption/dh64"
)

type df struct {
	_data *bytes.Buffer
}

func (slf *df) GetBufferCap() int {
	return slf._data.Cap()
}

func (slf *df) GetBufferLen() int {
	return slf._data.Len()
}

func (slf *df) GetBufferBytes() []byte {
	return slf._data.Bytes()
}
func (slf *df) ClearBuffer() {
	slf._data.Reset()
}
func (slf *df) TrunBuffer(n int) {
	slf._data.Next(n)
}
func (slf *df) WriteBuffer(b []byte) (int, error) {
	return slf._data.Write(b)
}

func (slf *df) ReadBuffer(n int) []byte {
	return slf._data.Next(n)
}

//TestGatewayDecode doc
func TestGatewayDecode(t *testing.T) {
	/*d := &df{_data: bytes.NewBuffer([]byte{})}
	d._data.Grow(4096)

	ed := gateway.TestEncoder("ddddtest", []byte("css001-gb-01k2"))
	d.WriteBuffer(ed)
	fmt.Printf("data len:%d\n", d.GetBufferLen())

	name, data, err := gateway.TestDecoder(d)
	fmt.Printf("decode:%s-%s-%+v\n", name, string(data), err)*/
	ServerKex := &dh64.KeyExchange{P: dh64.DefaultP, G: dh64.DefaultG}
	ClientKex := &dh64.KeyExchange{P: dh64.DefaultP, G: dh64.DefaultG}

	serverPrvKey, serverPubKey := ServerKex.KeyPair()
	clientPrvKey, clientPubKey := ClientKex.KeyPair()

	serverSecret := make([]byte, 8)
	clientSecret := make([]byte, 8)

	binary.BigEndian.PutUint64(serverSecret, ServerKex.Secret(serverPrvKey, clientPubKey))
	binary.BigEndian.PutUint64(clientSecret, ClientKex.Secret(clientPrvKey, serverPubKey))

	serverEncrypt := &encryption.NetRC4Encrypt{}
	clientEncrypt := &encryption.NetRC4Encrypt{}

	if err := serverEncrypt.Cipher(serverSecret); err != nil {
		fmt.Println(err)
		return
	}

	if err := clientEncrypt.Cipher(clientSecret); err != nil {
		fmt.Println(err)
		return
	}

	serverToClientData := &df{_data: bytes.NewBuffer([]byte{})}
	serverToClientData._data.Grow(4096)

	serverEData := gateway.TestEncoder(serverEncrypt, "ddddtest", []byte("css001-gb-01k2"))
	serverToClientData.WriteBuffer(serverEData)
	sToCName, sToCData, sToCErr := gateway.TestDecoder(clientEncrypt, serverToClientData)
	fmt.Printf("server To Cleint:%s-%s-%+v\n", sToCName, string(sToCData), sToCErr)

	clientToServerData := &df{_data: bytes.NewBuffer([]byte{})}
	clientToServerData._data.Grow(4096)

	clientEData := gateway.TestEncoder(clientEncrypt, "ddddtest", []byte("css001-gb-01k2"))
	clientToServerData.WriteBuffer(clientEData)
	cToSName, cToSData, cToSErr := gateway.TestDecoder(serverEncrypt, clientToServerData)
	fmt.Printf("client To Server:%s-%s-%+v\n", cToSName, string(cToSData), cToSErr)

}
