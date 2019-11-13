package parts

import (
	"bytes"
	"fmt"
	"reflect"

	"github.com/yamakiller/magicNet/engine/actor"
	"github.com/yamakiller/magicNet/network"
	"github.com/yamakiller/magicNet/network/parser"
	"github.com/yamakiller/magicNet/service/implement"
	"github.com/yamakiller/magicNet/timer"
	"github.com/yamakiller/magicParts/gateway/elements/forward"
	"github.com/yamakiller/magicParts/gateway/events"
	"github.com/yamakiller/magicParts/gateway/global"
)

//IGWCCodecMethod desc
//@Instance IGWCCodecMethod desc: GWClient Codec instance
//@func Decode
//@func Encode
type IGWCCodecMethod interface {
	Decode(keyPair interface{}, b *bytes.Buffer) (parser.IResult, error)
	Encode(keyPair interface{}, p interface{}) []byte
}

//IGWCHandshakeMethod desc
//@Instance IGWCHandshakeMethod desc: GWClient Hanshake Protocol encode instance
//@func Encode
type IGWCHandshakeMethod interface {
	Encode(key interface{}) []byte
}

//IGWCUnOnlineMethod desc
//@Instance IGWCUnOnlineMethod desc: GWClient unonline Protocol encode instance
//@func Encode
type IGWCUnOnlineMethod interface {
	Encode(h uint64) *events.RouteServerEvent
}

//IGWCSreachProto desc
//@Instance IGWCSreachProto desc: GWClient Protocol sreacher
//@func Sreach
type IGWCSreachProto interface {
	Sreach(name interface{}) reflect.Type
}

//SpawnGWLDeleate desc
//@method SpawnGWLDeleate desc: make Gateway Listen Deleate
//@param (IGWCCodecMethod) a Gateway Client Codec Method instance
//@param (IGWCHandshakeMethod) a Gateway Client Handshake Encode Method instance
//@param (IGWCUnOnlineMethod) a Gateway Client UnOnline Event Encode Method instance
//@param (IGWCSreachProto) a Gateway Client Network Protocol Sreacher instance
//@return (GWListenDeleate) a Gateway Listen Deleate object
func SpawnGWLDeleate(codec IGWCCodecMethod,
	hanshake IGWCHandshakeMethod,
	unOnline IGWCUnOnlineMethod,
	protocol IGWCSreachProto) GWListenDeleate {
	return GWListenDeleate{codec, hanshake, unOnline, protocol}
}

//GWListenDeleate desc
//@struct GWListenDeleate desc: network listening service, delegate logic
type GWListenDeleate struct {
	codec    IGWCCodecMethod
	hanshake IGWCHandshakeMethod
	unOnline IGWCUnOnlineMethod
	protocol IGWCSreachProto
}

//Handshake desc
//@method Handshake desc: network connect handshake processing
//@param  (implement.INetClient) connect client
//@return error: return error, success: return nil
func (slf *GWListenDeleate) Handshake(c implement.INetClient) error {
	c.BuildKeyPair()
	publicKey := c.GetKeyPublic()
	shake := slf.hanshake.Encode(publicKey)
	if err := network.OperWrite(c.GetSocket(), shake, len(shake)); err != nil {
		return err
	}
	return nil
}

//Analysis desc
//@method Analysis desc: Analysis of network data
//@param (actor.Context) current service/actor context
//@param (*implement.NetListenService) current network listen service
//@param (implement.INetClient) current recvice data client
//@return fail:return error,success return nil
func (slf *GWListenDeleate) Analysis(context actor.Context,
	nets *implement.NetListenService,
	c implement.INetClient) error {
	wrap, err := slf.codec.Decode(c.GetKeyPair(), c.GetRecvBuffer())
	if err != nil {
		return err
	}

	if wrap == nil {
		return implement.ErrAnalysisProceed
	}

	var unit *forward.Unit
	msgType := slf.protocol.Sreach(wrap.GetCommand())
	if msgType != nil {
		if f := nets.NetMethod.GetType(msgType); f != nil {
			f(implement.NetMethodEvent{
				Name:   wrap.GetCommand(),
				Socket: c.GetSocket(),
				Wrap:   wrap.GetWrap().([]byte),
			})
			goto end
		}
	}

	unit = global.FAddr.Sreach(msgType)
	if unit == nil {
		return fmt.Errorf("Abnormal protocol, no protocol information defined")
	}

	if unit.Auth && c.GetAuth() == 0 {
		return forward.ErrForwardClientUnverified
	}

	actor.DefaultSchedulerContext.Send((c.(*GWClient)).GetPID(),
		&events.RouteServerEvent{Handle: c.GetID(),
			ProtoName: wrap.GetCommand(),
			ServoName: unit.ServoName,
			Data:      wrap.GetWrap().([]byte)})
end:
	return implement.ErrAnalysisSuccess
}

//UnOnlineNotification desc
//@method UnOnlineNotification desc:  Offline notification
//@param  (uint64) client handle
func (slf *GWListenDeleate) UnOnlineNotification(h uint64) error {
	evt := slf.unOnline.Encode(h)
	fpid := global.SSets.Get(global.NSName)
	if fpid == nil {
		return forward.ErrForwardServiceNotStarted
	}

	actor.DefaultSchedulerContext.Send(fpid, evt)

	return nil
}

//GWListener desc
//@struct GWListener desc: gateway internet monitoring service
//@inherit (implement.NetListenService)
//@member  (GWNFEncodeMethod) a internet message encode function
type GWListener struct {
	implement.NetListenService
	codec IGWCCodecMethod
}

//Init desc
//@method Init desc: initialize gateway listener
func (slf *GWListener) Init() {
	slf.NetClients.Init()
	slf.NetListenService.Init()
	slf.RegisterMethod(&events.RouteClientEvent{}, slf.onRouteClient)
}

//SetCodecer desc
//@method SetCodecer desc: Setting codec instance
//@param (IGWCCodecMethod) gateway client codec method instance
func (slf *GWListener) SetCodecer(codec IGWCCodecMethod) {
	slf.codec = codec
}

//onRouteClient desc
//@method onRouteClient desc: route directly to the client
//@param (actor.Context) current service/actor context
//@param (*actor.PID) event source service/actor ID code
//@param (interface{}) event message
//@return (void)
func (slf *GWListener) onRouteClient(context actor.Context,
	sender *actor.PID,
	message interface{}) {
	msg := message.(*events.RouteClientEvent)
	h := implement.NetHandle{}
	h.SetValue(msg.Handle)
	c := slf.NetClients.Grap(h.GetValue())
	if c == nil {
		slf.LogError("Failed to send data, %+v client does not exist", msg.Handle)
		return
	}

	defer slf.NetClients.Release(c)

	sd := slf.codec.Encode(c.GetKeyPair(), msg)
	if sd == nil {
		slf.LogError("Protocol package failed to be assembled: protocol name %s",
			msg.ProtoName)
		return
	}

	if err := network.OperWrite(h.GetSocket(), sd, len(sd)); err != nil {
		slf.LogError("Failed to send data to client socket: %+v protocol name %s",
			err,
			msg.ProtoName)
		return
	}

	c.GetStat().UpdateRead(timer.Now(), uint64(len(sd)))
	slf.LogDebug("Already sent data to the client: protocol name %s", msg.ProtoName)
}

//Shutdown desc
//@method Shutdown desc: Turn off the gateway network listening service
func (slf *GWListener) Shutdown() {
	name := slf.Key()
	slf.NetListenService.Shutdown()
	global.SSets.Erase(name)
}
