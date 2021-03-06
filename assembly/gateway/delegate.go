package gateway

import (
	"encoding/binary"
	"fmt"
	"reflect"

	"github.com/yamakiller/magicLibs/encryption/dh64"
	"github.com/yamakiller/magicNet/handler/encryption"

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

func decoder(encrypt encryption.INetEncryption, bf net.INetReceiveBuffer) (string, []byte, error) {

	/***************************************************************|
	|      24 Bit   |    8 Bit	  |     N Bit    |    （N） Bit      |
	|-------------  |-------------|--------------|------------------|
	|  Data Length  |  Data Name  |  Data Name   |      Data        |
	|			    |	Length    |				 |					|
	****************************************************************/

	if bf.GetBufferLen() < constHeadByte {
		return "", nil, net.ErrAnalysisProceed
	}

	tmpByte := make([]byte, 4)
	copy(tmpByte, bf.GetBufferBytes()[:constHeadByte])
	if encrypt != nil {
		encrypt.Decode(tmpByte, tmpByte)
	}
	header := binary.BigEndian.Uint32(tmpByte)
	tmpDataLength := getDataLength(header)
	tmpDataNameLength := getDataNameLength(header)

	if (tmpDataLength + tmpDataNameLength + constHeadByte) > bf.GetBufferLen() {
		return "", nil, net.ErrAnalysisProceed
	}

	if (constHeadByte + tmpDataLength + tmpDataNameLength) > (bf.GetBufferCap() << 1) {
		return "", nil, code.ErrDataOverflow
	}

	bf.TrunBuffer(constHeadByte)
	tmpByte = bf.ReadBuffer(tmpDataNameLength + tmpDataLength)
	if encrypt != nil {
		encrypt.Decode(tmpByte, tmpByte)
	}

	name := string(tmpByte[:tmpDataNameLength])
	data := tmpByte[tmpDataNameLength:]

	return name, data, nil
}

func encoder(encrypt encryption.INetEncryption, dataName string, data []byte) []byte {
	dataNameLength := len([]byte(dataName))
	dataLength := len(data)

	header := uint32(((dataLength & constDataLengthMask) << constDataLengthShift))
	header = (header | uint32(dataNameLength&constDataNameLengthMask))

	result := make([]byte, constHeadByte+dataNameLength+dataLength)
	binary.BigEndian.PutUint32(result, header)
	copy(result[constHeadByte:], []byte(dataName))
	copy(result[constHeadByte+dataNameLength:], data)

	if encrypt != nil {
		encrypt.Encrypt(result, result)
	}

	return result
}

//DefaultAgreement doc
//@Summary default agreement method
//@Member  route address
//@Member  local method
//@Member  remove method
//@Member  is auth
type DefaultAgreement struct {
	Addr        string
	LocalMethod interface{}
	Auth        bool
}

//DefaultDelegate doc
//@Summary default gateserver delegate instance
//@Member  map  method
type DefaultDelegate struct {
	KeyExc  *dh64.KeyExchange
	Encrypt bool
	Maps    map[interface{}]interface{}
}

//PutLocalCall doc
//@Summary register agreement
//@Param agreement
//@Param route address
//@Param local method
//@Param remote method
//@Param is need auth
func (slf *DefaultDelegate) PutLocalCall(param interface{}, localMethod interface{}) {

	slf.Maps[reflect.TypeOf(param)] = localMethod
}

func (slf *DefaultDelegate) getLocalCall(param interface{}) interface{} {
	if v, ok := slf.Maps[reflect.TypeOf(param)]; ok {
		return v
	}

	return nil
}

//AsyncAccept doc
//@Summary server accept event
//@Param  client
//@Return error
func (slf *DefaultDelegate) AsyncAccept(c net.INetClient) error {
	privateKey, publicKey := slf.KeyExc.KeyPair()
	c.(*client).WithPrvKey(privateKey)
	x := make([]byte, 8)
	binary.BigEndian.PutUint64(x, publicKey)
	if err := c.SendTo(x); err != nil {
		return err
	}
	return nil
}

//AsyncClosed doc
//@Summary server closed client event
//@Param client handle
//@Return error
func (slf *DefaultDelegate) AsyncClosed(handle uint64) error {
	return nil
}

//AsyncDecode doc
//@Summary network data decode method
//@Param   client
//@Return  AgreMessage
//@Return  error
func (slf *DefaultDelegate) AsyncDecode(c net.INetClient) (*AgreMsg, error) {
	gwClient := c.(*client)
	if gwClient.Encrypt() == nil {
		if c.GetBufferLen() < 8 {
			return nil, net.ErrAnalysisProceed
		}

		publicKey := binary.BigEndian.Uint64(c.ReadBuffer(8))
		secret := slf.KeyExc.Secret(gwClient.GetPrvKey(), publicKey)
		rc4 := &encryption.NetRC4Encrypt{}
		secretByte := make([]byte, 8)
		binary.BigEndian.PutUint64(secretByte, secret)
		if err := rc4.Cipher(secretByte); err != nil {
			return nil, err
		}
		gwClient.WithEncrypt(rc4)

	}

	name, data, err := decoder(slf.getEncrypt(gwClient), c)
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

	return &AgreMsg{name, &msg}, nil
}

//AsyncEncode doc
//@Summary network data encode method
//@Param   need encode data
//@Return  encode result
//@Return  error
func (slf *DefaultDelegate) AsyncEncode(c net.INetClient,
	response interface{}) ([]byte, error) {
	d, err := proto.Marshal(response.(proto.Message))
	if err != nil {
		return nil, err
	}

	msgName := proto.MessageName(response.(proto.Message))

	return encoder(slf.getEncrypt(c.(*client)), msgName, d), nil
}

func (slf *DefaultDelegate) getEncrypt(c *client) encryption.INetEncryption {
	if slf.Encrypt {
		return c.Encrypt()
	}
	return nil
}

/*func TestDecoder(encrypt encryption.INetEncryption, bf net.INetReceiveBuffer) (string, []byte, error) {
	return decoder(encrypt, bf)
}

func TestEncoder(encrypt encryption.INetEncryption, dataName string, data []byte) []byte {
	return encoder(encrypt, dataName, data)
}*/
