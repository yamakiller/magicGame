package gateway

import (
	"encoding/binary"
	"fmt"
	"reflect"

	"github.com/gogo/protobuf/proto"
	"github.com/yamakiller/magicGame/assembly/code"
	"github.com/yamakiller/magicNet/handler/net"
)

const (
	constHeadSize = 32
	//
	constHeadByte = 4
	//
	constDataLengthSize = 24
	//data length start pos bit
	constDataLengthStart = 0
	//data length mask
	constDataLengthMask = 0xFFFFFF
	//data length shift
	constDataLengthShift = 8
	//data name length size
	constDataNameLengthSize = 8
	//data name length start pos bit
	constDataNameLengthStart = constDataLengthStart + constDataLengthSize
	//data name length mask
	constDataNameLengthMask = 0xFF
	//data name length shift
	constDataNameLengthShift = 0
)

func getDataLength(d uint32) int {
	return int((d >> constDataLengthShift) & constDataLengthMask)
}

func getDataNameLength(d uint32) int {
	return int(d & constDataNameLengthMask)
}

func decoder(bf net.INetReceiveBuffer) (string, []byte, error) {
	//#Data Header#######################################
	/******************************
	|      24 Bit   |    8 Bit	  |
	|-------------  |-------------|
	|  Data Length  |  Data Name  |
	|			    |	Length    |
	*******************************
	//#Data Chunk########################################
	/*********************************|
	|   Data Name Length  |    （N）  |
	|         Bit         |    Bit   |
	|---------------------|----------|
	|	   Data Name      |   Data   |
	|					  |			 |
	*********************************/

	if bf.GetBufferLen() > constHeadByte {
		return "", nil, net.ErrAnalysisProceed
	}

	header := binary.BigEndian.Uint32(bf.GetBufferBytes()[:constHeadByte])
	tmpDataLength := getDataLength(header)
	tmpDataNameLength := getDataNameLength(header)

	if (tmpDataLength + tmpDataNameLength + constHeadByte) > bf.GetBufferLen() {
		return "", nil, net.ErrAnalysisProceed
	}

	if (constHeadByte + tmpDataLength + tmpDataNameLength) > (bf.GetBufferCap() << 1) {
		return "", nil, code.ErrDataOverflow
	}

	bf.TrunBuffer(constHeadByte)
	name := string(bf.ReadBuffer(tmpDataNameLength))
	data := bf.ReadBuffer(tmpDataLength)

	return name, data, nil
}

func encoder(dataName string, data []byte) []byte {
	dataNameLength := len([]byte(dataName))
	dataLength := len(data)

	var header uint32 = uint32(((dataLength & constDataLengthMask) << constDataLengthShift))
	header = (header | uint32(dataNameLength&constDataNameLengthMask))

	result := make([]byte, constHeadByte+dataNameLength+dataLength)
	binary.BigEndian.PutUint32(result, header)
	copy(result[constHeadByte:], []byte(dataName))
	copy(result[constHeadByte+dataNameLength:], data)

	return result
}

type DefaultMethod struct {
	Addr         string
	Method       interface{}
	RemoteMethod string
	Auth         bool
}

type DefaultDelegate struct {
	_ms map[interface{}]*DefaultMethod
}

func (slf *DefaultDelegate) RegisterAyncMethod(key interface{}, addr string, method interface{}, remoteMethod string, auth bool) {
	slf._ms[reflect.TypeOf(key)] = &DefaultMethod{addr, method, remoteMethod, auth}
}

func (slf *DefaultDelegate) AsyncAccept(c net.INetClient) error {
	publicKey := c.(*client).KeyPair()
	x := make([]byte, 8)
	binary.BigEndian.PutUint64(x, publicKey)
	if err := c.SendTo(x); err != nil {
		return err
	}
	return nil
}

func (slf *DefaultDelegate) AsyncClosed(uint64) error {
	return nil
}

func (slf *DefaultDelegate) QueryAsyncMethod(key interface{}) (string, interface{}, bool, error) {

	if v, ok := slf._ms[key]; ok {
		return v.Addr, v.Method, v.Auth, nil
	}

	return "", nil, false, nil
}

func (slf *DefaultDelegate) QueryAsyncRemoteMethod(key interface{}) (string, error) {
	if v, ok := slf._ms[key]; ok {
		return v.RemoteMethod, nil
	}

	return "", fmt.Errorf("%+v protocol remote method is not defined", key)
}

func (slf *DefaultDelegate) AsyncDecode(c net.INetClient) (*AgreMessage, error) {
	if c.(*client).Secert() == 0 {
		if c.GetBufferLen() < 8 {
			return nil, net.ErrAnalysisProceed
		}

		publicKey := binary.BigEndian.Uint64(c.ReadBuffer(8))
		c.(*client).BuildSecert(publicKey)
	}

	name, data, err := decoder(c)
	if err != nil {
		return nil, err
	}

	msgType := proto.MessageType(name)
	if msgType == nil {
		return nil, fmt.Errorf("%s protocol is undefined", name)
	}

	msg := reflect.Indirect(reflect.New(msgType.Elem())).Addr().Interface().(proto.Message)

	err = proto.Unmarshal(data, msg)
	if err != nil {
		return nil, err
	}

	return &AgreMessage{name, &msg}, nil
}

func (slf *DefaultDelegate) AsyncEncode(response interface{}) ([]byte, error) {

	d, err := proto.Marshal(response.(proto.Message))
	if err != nil {
		return nil, err
	}

	return encoder(proto.MessageName(response.(proto.Message)), d), nil
}
