package parts

import (
	"bytes"
	"math/rand"
	"strconv"
	"sync"
	"time"

	"github.com/yamakiller/magicNet/engine/actor"
	"github.com/yamakiller/magicNet/service"
	"github.com/yamakiller/magicNet/service/implement"
	"github.com/yamakiller/magicParts/gateway/elements/target"
	"github.com/yamakiller/magicParts/gateway/events"
	"github.com/yamakiller/magicParts/gateway/global"
)

//GWRouter desc
//@struct GWRouter desc: Gateway client route=>server/gateway remote client route=>client
//@inherit (service.Service)
//@member  ([]*GWRemoteClient) gateway connection logic server,client group
//@member  (map[string]*actor.PID) future client service pids
//@member  (*time.Ticker) remote client connect state checking/tick
//@member  (chan bool) tick close sign
//@member  (sync.WaitGroup)
//@member  (bool) is checking
//@member  (bool) is shutdown
//@member  (IGWRemoteCCodecMethod) gateway remote client network encode/Decode function
//@member  (IGWRemoteCSreachProto) gateway remote client protocol name map reflect.Type
//@member  (GWRCFDecodeRegisterRepMethod) gateway remote client network(register response decode function)
//@member  (IGWRemoteCRegisterProto) gateway remote client register protocol encode/Decode function
//@member  (RegisterRCMethodCall) /event method registration callback function at initialization time remote client
type GWRouter struct {
	service.Service
	conns      []*GWRemoteClient
	futures    map[string]*actor.PID
	authTicker *time.Ticker
	authSign   chan bool
	authWait   sync.WaitGroup
	isChecking bool
	isShutdown bool

	CCodec               IGWRemoteCCodecMethod
	Protocol             IGWRemoteCSreachProto
	RegisterProto        IGWRemoteCRegisterProto
	RegisterRCMethodCall func(client *GWRemoteClient)
}

type autoCheckEvent struct {
}

//Init desc
//@method Init desc: Initialize the repeater
func (slf *GWRouter) Init() {
	slf.Service.Init()
	slf.futures = make(map[string]*actor.PID)
	slf.RegisterMethod(&actor.Started{}, slf.Started)
	slf.RegisterMethod(&actor.Stopping{}, slf.Stopping)
	slf.RegisterMethod(&autoCheckEvent{}, slf.onCheckClient)
	slf.RegisterMethod(&events.RouteClientEvent{}, slf.onRouteClient)
	slf.RegisterMethod(&events.RouteServerEvent{}, slf.onRouteServer)
	slf.RegisterMethod(&events.CancelFutureEvent{}, slf.onCancelFuture)
}

//Started desc
//@method Started desc: Start router service
//@param (actor.Context) the service/actor context
//@param (*actor.PID) the event source service/actor id
//@param (interface{}) the event data
func (slf *GWRouter) Started(context actor.Context,
	sender *actor.PID,
	message interface{}) {

	slf.Service.Assignment(context)
	slf.LogInfo("Service Startup %s", slf.Name())
	tset := global.TSets.GetValues()
	if tset == nil {
		goto end
	}

	slf.LogInfo("%d connection target numbers", len(tset))
	slf.LogInfo("Startup trying to create a machine")
	for _, v := range tset {
		t := v.(*target.TConn)
		name := t.Name + "#" + strconv.Itoa(int(t.ID))
		slf.LogInfo("Start generating %s connectors address:%s", name, t.Addr)
		con := service.Make(name, func() service.IService {
			c := &GWRemoteClient{ServerID: global.EnvInstance().ServerID,
				AutoErrRetry:     global.EnvInstance().LinkerForwardErr,
				AutoErrRetryTime: global.EnvInstance().LinkerForwardErrInterval,
				Codec:            slf.CCodec,
				Protocol:         slf.RegisterProto,
				NetConnectService: implement.NetConnectService{
					Handle: &GWHRemoteClient{ReceiveBufferLimit: global.EnvInstance().ResponseClientBufferLimit,
						receiveBuffer: bytes.NewBuffer([]byte{})},
					Deleate: &GWRemoteClientDeleate{CCodec: slf.CCodec,
						Protocol: slf.Protocol},
					Target: t}}

			c.Handle.GetReceiveBuffer().Grow(c.Handle.GetReceiveBufferLimit())
			c.Init()
			if slf.RegisterRCMethodCall != nil {
				slf.RegisterRCMethodCall(c)
			}
			return c
		})

		slf.conns = append(slf.conns, con.(*GWRemoteClient))
		slf.LogInfo("Start generating %s connectors address:%s complete", name, t.Addr)
	}

	//auto connect==========================================
	slf.isChecking = true
	actor.DefaultSchedulerContext.Send(slf.GetPID(),
		&autoCheckEvent{})
	//======================================================

	slf.authWait.Add(1)
	slf.authTicker = time.NewTicker(time.Duration(global.EnvInstance().ResponseClientCheckInterval) * time.Millisecond)
	slf.authSign = make(chan bool, 1)
	go func(t *time.Ticker) {
		defer slf.authWait.Done()
		for {
			select {
			case <-t.C:
				if !slf.isChecking {
					slf.isChecking = true
					actor.DefaultSchedulerContext.Send(slf.GetPID(),
						&autoCheckEvent{})
				}
			case stop := <-slf.authSign:
				if stop {
					return
				}
			}
		}
	}(slf.authTicker)

	slf.LogInfo("%s Service Startup completed", slf.Name())
end:
	slf.Service.Started(context, sender, message)
}

//Stopping desc
//@method Stopping desc: Stop router service
//@param (actor.Context) the service/actor context
//@param (*actor.PID) the event source service/actor id
//@param (interface{}) the event data
func (slf *GWRouter) Stopping(context actor.Context,
	sender *actor.PID,
	message interface{}) {
	n := len(slf.conns)
	slf.LogInfo("Service Stopping [connecting:%d]", n)
	for _, v := range slf.conns {
		slf.LogInfo("Connection Stopping %d name:%s address:%s", n, v.Target.GetName(), v.Target.GetAddr())
		v.Shutdown()
		slf.LogInfo("Connection Stoped name:%s address:%s", v.Target.GetName(), v.Target.GetAddr())
		n--
	}
	slf.conns = slf.conns[:0]
	slf.futures = make(map[string]*actor.PID)
	slf.LogInfo("Service Stoped")
}

//Shutdown desc
//@method Shutdown desc: Termination of service
func (slf *GWRouter) Shutdown() {
	slf.isShutdown = false
	slf.authSign <- true
	slf.authWait.Wait()
	close(slf.authSign)
	slf.authTicker.Stop()
	slf.Service.Shutdown()
}

//onCheckConnect desc
//@method onCheckConnect desc: check remote client connection state
//@param (actor.Context) the service/actor context
//@param (*actor.PID) the event source service/actor id
//@param (interface{}) the event data
func (slf *GWRouter) onCheckClient(context actor.Context,
	sender *actor.PID,
	message interface{}) {
	defer func() {
		slf.isChecking = false
	}()

	for _, v := range slf.conns {
		//End and exit
		if slf.isShutdown {
			return
		}

		switch v.Target.GetEtat() {
		case implement.Connected:
		case implement.Connecting:
			fallthrough
		case implement.Verify:
			if outTm := v.Target.IsTimeout(); outTm > 0 {
				v.Handle.Close()
			}
		case implement.UnConnected:
			if v.GetPID() != nil {
				v.Target.SetEtat(implement.Connecting)
				actor.DefaultSchedulerContext.Send(v.GetPID(),
					&implement.NetConnectEvent{})
			}
		default:
			slf.LogDebug("Exception non-existent logic")
		}
	}

}

//onRouteClient desc
//@method onRouteClient desc:
//@param (actor.Context) the service/actor context
//@param (*actor.PID) the event source service/actor id
//@param (interface{}) the event data
func (slf *GWRouter) onRouteClient(context actor.Context,
	sender *actor.PID,
	message interface{}) {

	msg := message.(*events.RouteClientEvent)
	msgType := slf.Protocol.Sreach(msg.ProtoName)
	if msgType == nil {
		slf.LogError("The %s protocol is not defined, and the data is discarded depending on the abnormal operation.",
			msg.ProtoName)
		return
	}

	adr := global.FAddr.Sreach(msgType)
	if adr == nil || !(adr.ServoName == "client") {
		slf.LogError("Protocol not registered route or routing address error: protocol name:%s",
			msg.ProtoName)
		return
	}

	//wakeup---
	futureKey := strconv.FormatUint(msg.Handle, 10) + "#" + strconv.FormatUint(uint64(msg.Serial), 10)
	if pid, ok := slf.futures[futureKey]; ok {
		actor.DefaultSchedulerContext.Send(pid, struct{}{})
		delete(slf.futures, futureKey)
	}

	pid := global.SSets.Get(global.NSName)
	if pid == nil {
		slf.LogError("Network Service Department exists")
		return
	}

	actor.DefaultSchedulerContext.Send(pid, msg)
	slf.LogDebug("The send request has been handed over to the network service module: protocol name:%s",
		msg.ProtoName)
}

//onRouteServer desc
//@method onRouteServer desc: user client data send to Logic server
//@param (actor.Context) the service/actor context
//@param (*actor.PID) the event source service/actor id
//@param (interface{}) the event data
func (slf *GWRouter) onRouteServer(context actor.Context,
	sender *actor.PID,
	message interface{}) {
	msg := message.(*events.RouteServerEvent)
	loader := global.TLSets.Get(msg.ServoName)
	if loader == nil {
		slf.LogError("The %s target server was not found"+
			" and the data was discarded.", msg.ServoName)
		return
	}

	ick := 0
	var to *target.TObject
	for {
		to = loader.GetTarget(strconv.Itoa(rand.Intn(10000)))
		if to != nil {
			break
		}

		ick++
		if ick >= 6 {
			slf.LogError("[%s]No attempts were made to find"+
				" available nodes and data was dropped", msg.ServoName)
			goto error_resume
		}
	}

	if sender != nil {
		futureKey := strconv.FormatUint(msg.Handle, 10) + "#" + strconv.FormatUint(uint64(msg.Serial), 10)
		slf.futures[futureKey] = sender
	}

	actor.DefaultSchedulerContext.Send(to.Target, msg)
	slf.LogDebug("Data send request has been pushed successfully")
	return
error_resume:
	if msg.Serial > 0 && sender != nil {
		actor.DefaultSchedulerContext.Send(sender, struct{}{})
	}
}

//onCancelFuture desc
//@method onCancelFuture desc: cancel futuring user client service/actor
//@param (actor.Context) the service/actor context
//@param (*actor.PID) the event source service/actor id
//@param (interface{}) the event data
func (slf *GWRouter) onCancelFuture(context actor.Context,
	sender *actor.PID,
	message interface{}) {
	msg := message.(*events.CancelFutureEvent)
	futureKey := strconv.FormatUint(msg.Handle, 10) + "#" + strconv.FormatUint(uint64(msg.Serial), 10)
	if _, ok := slf.futures[futureKey]; ok {
		delete(slf.futures, futureKey)
	}
}
