package parts

import (
	"bytes"
	"reflect"
	"time"

	"github.com/yamakiller/magicNet/network/parser"
	"github.com/yamakiller/magicNet/timer"

	"github.com/yamakiller/magicNet/engine/actor"
	"github.com/yamakiller/magicNet/network"
	"github.com/yamakiller/magicNet/service/implement"
	"github.com/yamakiller/magicNet/service/net"
	"github.com/yamakiller/magicParts/gateway/elements/forward"
	"github.com/yamakiller/magicParts/gateway/elements/target"
	"github.com/yamakiller/magicParts/gateway/events"
	"github.com/yamakiller/magicParts/gateway/global"
)

//GWHRemoteClient desc
//@struct GWHRemoteClient desc: gateway connect logic server client
//@inherit net.TCPConnection
//@member  (bytes.Buffer) Receive buffer
//@member  (int) Receive buffer size limit
//@member  (uint64) Author time
type GWHRemoteClient struct {
	net.TCPConnection
	receiveBuffer      *bytes.Buffer
	ReceiveBufferLimit int
	auth               uint64
	stat               implement.NetStat
}

//GetReceiveBufferLimit desc
//@method GetReceiveBufferLimit desc: Return this connection receive buffer limit
//@return (int) receive buffer limit
func (slf *GWHRemoteClient) GetReceiveBufferLimit() int {
	return slf.ReceiveBufferLimit
}

//GetReceiveBuffer desc
//@method GetReceiveBuffer desc: Return receive buffer object
//@return (*bytes.Buffer) receive buffer object
func (slf *GWHRemoteClient) GetReceiveBuffer() *bytes.Buffer {
	return slf.receiveBuffer
}

//GetStat desc
//@method GetStat desc: Return data status information
//@return (net.INetConnectionDataStat) a data status information
func (slf *GWHRemoteClient) GetStat() net.INetConnectionDataStat {
	return &slf.stat
}

//GetAuth desc
//@method GetAuth desc: Return the authentication time
//@return (uint64) time
func (slf *GWHRemoteClient) GetAuth() uint64 {
	return slf.auth
}

//SetAuth desc
//@method SetAuth desc: Setting the authentication time
//@param (uint64) time
func (slf *GWHRemoteClient) SetAuth(auth uint64) {
	slf.auth = auth
}

//Close desc
//@method Close the client
func (slf *GWHRemoteClient) Close() {
	slf.TCPConnection.Close()
}

//type GWFDecodeMethod func(keyPair interface{}, b *bytes.Buffer) (parser.IResult, error)
//type GWFGetMessageType func(name interface{}) reflect.Type

//IGWRemoteCCodecMethod desc
//@Instance IGWRemoteCCodecMethod desc: remote client Codec instance
//@func Decode
//@func Encode
type IGWRemoteCCodecMethod interface {
	Decode(keyPair interface{}, b *bytes.Buffer) (parser.IResult, error)
	Encode(keyPair interface{}, p interface{}) []byte
}

//IGWRemoteCSreachProto desc
//@Instance IGWRemoteCSreachProto desc: remote client Protocol sreacher
//@func Sreach
type IGWRemoteCSreachProto interface {
	Sreach(name interface{}) reflect.Type
}

//IGWRemoteCRegisterProto desc
//@Instance IGWRemoteCRegisterProto desc: remote client register protocol instance
//@func EncodeRegisterRequest encode register request
//@func DecodeRegisterResponse decode register response
type IGWRemoteCRegisterProto interface {
	EncodeRegisterRequest() ([]byte, error)
	DecodeRegisterResponse([]byte) (interface{}, error)
}

//GWRemoteClientDeleate desc
//@struct GWRemoteClientDeleate desc: gateway remove client deleate
type GWRemoteClientDeleate struct {
	CCodec   IGWRemoteCCodecMethod
	Protocol IGWRemoteCSreachProto
}

//Connected desc
//@method Connected desc Connected proccess
//@param (actor.Context) the service/actor context
//@param (implement.NetConnectService) the RemoteClient instance
//@return (error) fail:return error, success:nil
func (slf *GWRemoteClientDeleate) Connected(context actor.Context,
	nets *implement.NetConnectService) error {
	return nil
}

//Analysis Packet decomposition
func (slf *GWRemoteClientDeleate) Analysis(context actor.Context,
	nets *implement.NetConnectService) error {

	wrap, err := slf.CCodec.Decode(nil, nets.Handle.GetReceiveBuffer())
	if err != nil {
		return err
	}

	if wrap == nil {
		return implement.ErrAnalysisProceed
	}

	var fpid *actor.PID
	msgType := slf.Protocol.Sreach(wrap.GetCommand())

	if msgType != nil {
		if f := nets.NetMethod.GetType(msgType); f != nil {
			f(&events.ResponseNetMethodEvent{
				H: wrap.GetExtern().(uint64),
				S: wrap.GetSerial(),
				NetMethodEvent: implement.NetMethodEvent{Name: wrap.GetCommand(),
					Socket: nets.Handle.Socket(),
					Wrap:   wrap.GetWrap().([]byte)},
			})
			goto end
		}
	}

	fpid = global.SSets.Get(global.RSName)
	if fpid == nil {
		return forward.ErrForwardServiceNotStarted
	}

	actor.DefaultSchedulerContext.Send(fpid,
		&events.RouteClientEvent{Handle: wrap.GetExtern().(uint64),
			Serial:    wrap.GetSerial(),
			ProtoName: wrap.GetCommand(),
			Data:      wrap.GetWrap().([]byte)})
end:
	//return name, data, err
	return implement.ErrAnalysisSuccess
}

//GWRCFEncodeMethod network encode message
//type GWRCFEncodeMethod func(keyPair interface{}, p interface{}) []byte

//GWRCFEncodeRegisterRetMethod network encode register request message
//type GWRCFEncodeRegisterRetMethod func() ([]byte, error)

//GWRCFDecodeRegisterRepMethod network decode register response message
//type GWRCFDecodeRegisterRepMethod func([]byte) (interface{}, error)

//GWRCFGetRepEventMethod retrun a event object use register
//type GWRCFGetEventMethod func() interface{}

//GWRemoteClient desc
//@method GWRemoteClient desc: Gateway connector
//@inherit (implement.NetConnectServicServere)
//@member  (library.INetDeleate) Network data processing agent
//@member (int32) the server id
//@member (int) Automatic connection failure retries
//@member (int) Automatic connection failure retry interval unit milliseconds
type GWRemoteClient struct {
	implement.NetConnectService
	Codec    IGWRemoteCCodecMethod
	Protocol IGWRemoteCRegisterProto

	ServerID         int32
	AutoErrRetry     int
	AutoErrRetryTime int
}

//Init desc
//@method Init desc: initialization gateway remote client
func (slf *GWRemoteClient) Init() {
	slf.NetConnectService.Init()
	slf.RegisterMethod(&network.NetClose{}, slf.OnClose)
	slf.RegisterMethod(&events.RouteServerEvent{}, slf.onRouteServer)
}

//OnClose desc
//@method OnClose desc: Handling connection close events
//@param (actor.Context) the service/actor context
//@param (*actor.PID) event source service/actor id
//@param (interface{}) message
func (slf *GWRemoteClient) OnClose(context actor.Context,
	sender *actor.PID,
	message interface{}) {

	group := global.TLSets.Get(slf.Target.GetName())
	if group != nil {
		//Delete equalizer
		group.RemoveTarget(slf.Name())
	} else {
		slf.LogError("%s class equalizer does not exist", slf.Target.GetName())
	}
	slf.NetConnectService.OnClose(context, sender, message)
}

//onRouteServer desc
//@method onRouteServer desc: send data to logic server event
//@param (actor.Context) the service/actor context
//@param (*actor.PID) event source service/actor id
//@param (interface{}) message
func (slf *GWRemoteClient) onRouteServer(context actor.Context,
	sender *actor.PID,
	message interface{}) {

	ick := 0
	var err error
	request := message.(*events.RouteServerEvent)
	for {
		if slf.IsShutdown() {
			slf.LogWarning("Service has been terminated, data is discarded:%s",
				request.ProtoName)
			return
		}

		if slf.Target.GetEtat() != implement.Connected {
			if slf.Target.GetEtat() == implement.UnConnected {
				err = slf.NetConnectService.AutoConnect(context)
				if err != nil {
					slf.LogError("onForwardServer: AutoConnect fail:%+v", err)

				}
			} else {
				if outTime := slf.Target.IsTimeout(); outTime > 0 {
					slf.LogWarning("onForwardServer: AutoConnect TimeOut:%d millisecond",
						outTime)
					slf.Target.RestTimeout()
					slf.Handle.Close()
				}
			}
		} else {
			break
		}

		ick++
		if ick > slf.AutoErrRetry {
			slf.LogError("OnForwardServer AutoConnect fail, "+
				"Data is discarded %+v %+v %s %s %d check num:%d",
				request.Handle,
				request.Serial,
				request.ProtoName,
				request.ServoName,
				len(request.Data),
				ick)
			return
		}
		time.Sleep(time.Duration(slf.AutoErrRetryTime) * time.Millisecond)
	}

	fdata := slf.Codec.Encode(nil, request)
	if err = slf.Handle.Write(fdata, len(fdata)); err != nil {
		slf.LogError("OnForwardServer Send fail, Data is discarded %+v %s %s %d",
			request.Handle,
			request.ProtoName,
			request.ServoName,
			len(request.Data))
		return
	}

	slf.LogDebug("OnForwardServer Send %s success", request.ProtoName)
}

//OnNetHandshake desc
//@method OnNetHandshake desc: logic response handshake event
//@param (implement.INetMethodEvent)  event
func (slf *GWRemoteClient) OnNetHandshake(event implement.INetMethodEvent) {
	//Internal communication does not consider encrypted communication

	if slf.Target.GetEtat() != implement.Connecting {
		slf.LogError("onNetHandshake: handshake fail: current status %+v,%+v",
			slf.Target.GetEtat(), implement.Connecting)
		return
	}

	request, err := slf.Protocol.EncodeRegisterRequest()
	if err != nil {
		slf.LogError("onNetHandshake: handshake fail:%+v", err)
		goto unend
	}

	err = slf.Handle.Write(request, len(request))
	if err != nil {
		slf.LogError("onNetHandshake: Register ID fail:%+v", err)
		goto unend
	}

	slf.Target.SetEtat(implement.Verify)
	return
unend:
	slf.Target.SetEtat(implement.UnConnected)
}

//OnNetRegisterResponse desc
//@method OnNetRegisterResponse desc: logic response handshake event
//@param (implement.INetMethodEvent)  event
func (slf *GWRemoteClient) OnNetRegisterResponse(evt implement.INetMethodEvent) {
	response := evt.(*events.ResponseNetMethodEvent)
	slf.LogDebug("onNetRegisterResponse: remote handle:%+v %s",
		response.H, response.Name)
	_, err := slf.Protocol.DecodeRegisterResponse(response.Wrap)
	now := timer.Now()
	if err != nil {
		slf.LogError("onNetRegisterResponse: unmarshal fail:%+v", err)
		slf.Handle.Close()
		goto unend
	}

	if slf.Target.GetEtat() != implement.Verify {
		slf.LogError("onNetRegisterResponse: register fail: current status %+v,%+v",
			slf.Target.GetEtat(), implement.Verify)
		return
	}

	slf.Handle.SetAuth(now)
	slf.Target.SetEtat(implement.Connected)

	if group := global.TLSets.Get(slf.Target.GetName()); group != nil {
		group.AddTarget(slf.Name(),
			&target.TObject{ID: slf.Target.(*target.TConn).GetVirtualID(),
				Target: slf.GetPID()})
	} else { //registration failed
		slf.Handle.Close()
		slf.LogError("Registration to the loader failed, %s such loader does not exist",
			slf.Target.GetName())
		goto unend
	}

	slf.LogInfo("onNetRegisterResponse: connected address:%s time:%+v success ",
		slf.Target.GetAddr(), now)
	return
unend:
	slf.Target.SetEtat(implement.UnConnected)
}
