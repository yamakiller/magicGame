package test

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/yamakiller/magicGame/assembly/gateway"
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
	d := &df{_data: bytes.NewBuffer([]byte{})}
	d._data.Grow(4096)

	ed := gateway.TestEncoder("ddddtest", []byte("css001-gb-01k2"))
	d.WriteBuffer(ed)
	fmt.Printf("data len:%d\n", d.GetBufferLen())

	name, data, err := gateway.TestDecoder(d)
	fmt.Printf("decode:%s-%s-%+v\n", name, string(data), err)
}
