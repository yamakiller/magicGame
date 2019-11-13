package parts

import (
	"bytes"
	"math"
	"sync"
	"time"

	"github.com/yamakiller/magicNet/engine/actor"
	"github.com/yamakiller/magicNet/network"
	"github.com/yamakiller/magicNet/service"
	"github.com/yamakiller/magicNet/service/implement"
	"github.com/yamakiller/magicNet/st/table"
	"github.com/yamakiller/magicParts/gateway/events"
	"github.com/yamakiller/magicParts/gateway/global"
	"github.com/yamakiller/magicParts/library"
)

//GWClient gateway network client
//@struct GWClient desc: link gateway server client servcie/actor
//@inherit (implement.NetClientService)
//@member  (implement.NetHandle) gateway handle/id[gateway id/handle/socket]
//@member  (uint64) authority time
//@member  (uint32) last wait response message serial
type GWClient struct {
	implement.NetClientService
	handle implement.NetHandle
	auth   uint64
	serial uint32
}

//Init desc
//@org    implement.NetClientService
//@method Init desc: initialize gateway client
func (slf *GWClient) Init() {
	slf.NetClientService.Init()
	slf.RegisterMethod(&events.RouteServerEvent{}, slf.onRouteServer)
}

//SetID desc
//@org    implement.NetClientService
//@method  SetID desc: Setting the client ID
//@param   (uint64) a handle/id
func (slf *GWClient) SetID(h uint64) {
	slf.handle.SetValue(h)
}

//GetID desc
//@org    implement.NetClientService
//@method GetID desc: Returns the client ID
//@return (uint64) return current  gateway client handle/id
func (slf *GWClient) GetID() uint64 {
	return slf.handle.GetValue()
}

//GetSocket desc
//@org    implement.NetClientService
//@method GetSocket desc: Returns the gateway client socket
//@return Returns the gateway client socket
func (slf *GWClient) GetSocket() int32 {
	return slf.handle.GetSocket()
}

//SetSocket desc
//@org     implement.NetClientService
//@method  SetSocket desc: Setting the gateway client socket
//@param   (int32) a socket id
func (slf *GWClient) SetSocket(sock int32) {
	slf.handle.Generate(slf.handle.GetServiceID(), slf.handle.GetHandle(), sock)
}

//Write desc
//@org  implement.NetClientService
//@method  Write desc: Send data to the client
//@param  ([]byte) a need send data
//@param  (int) need send data length
func (slf *GWClient) Write(d []byte, len int) {
	sock := slf.GetSocket()
	if sock <= 0 {
		slf.LogError("Write Data error:socket[0]")
		return
	}

	if err := network.OperWrite(sock, d, len); err != nil {
		slf.LogError("Write Data error:%+v", err)
	}
}

//SetAuth Setting the time for authentication
//@org  implement.NetClientService
//@method SetAuth desc: Setting author time
//@param (uint64) a author time
func (slf *GWClient) SetAuth(auth uint64) {
	slf.auth = auth
}

//GetAuth desc
//@org implement.NetClientService
//@method GetAuth desc: Returns the client author time
//@return (uint64) the client author time
func (slf *GWClient) GetAuth() uint64 {
	return slf.auth
}

//GetKeyPair desc
//@org implement.NetClientService
//@method GetKeyPair desc: Returns the client key pairs
//@return (interface{}) the client key pairs
func (slf *GWClient) GetKeyPair() interface{} {
	return nil
}

//BuildKeyPair desc
//@org implement.NetClientService
//@method BuildKeyPair desc: Building the client key pairs
func (slf *GWClient) BuildKeyPair() {

}

//GetKeyPublic desc
//@org implement.NetClientService
//@method GetKeyPublic desc: Returns the client public key
//@return (string) a public key
func (slf *GWClient) GetKeyPublic() string {
	return ""
}

//Shutdown desc
//@org implement.NetClientService
//@method Shutdown desc: Terminate this client
func (slf *GWClient) Shutdown() {
	slf.NetClientService.Shutdown()
}

//ptno
//@method ptno desc: The serial number to generate
//@return (uint32) a new serial number
func (slf *GWClient) ptno() uint32 {
	slf.serial++
	if slf.serial > math.MaxUint32 {
		slf.serial &= math.MaxUint32
		if slf.serial == 0 {
			slf.serial = 1
		}
	}

	return slf.serial
}

//onRouteServer desc
//@method onRouteServer desc: network message route event proccess
//@param (actor.Context) current service/actor context
//@param (*actor.PID) event source service/actor ID code
//@param (interface{}) event message
//@return (void)
func (slf *GWClient) onRouteServer(context actor.Context,
	sender *actor.PID,
	message interface{}) {
	sn := slf.ptno()
	evt := message.(*events.RouteServerEvent)
	evt.Serial = sn

	fpid := global.SSets.Get(global.RSName)
	if fpid == nil {
		slf.LogError("RPC Forawd Fail: error %s protocol find %s service",
			evt.ProtoName,
			global.RSName)
		return
	}

	future := actor.DefaultSchedulerContext.RequestFuture(fpid,
		evt,
		time.Millisecond*time.Duration(global.EnvInstance().ResponseTimeout))

	_, err := future.Result()
	if err != nil {
		//CancelFuture
		actor.DefaultSchedulerContext.Send(fpid, &events.CancelFutureEvent{Handle: evt.Handle, Serial: evt.Serial})
		slf.LogDebug("Gateway Client RPC Error:[%s,%s],%+v",
			evt.ProtoName,
			global.RSName,
			err)
		return
	}

	slf.LogDebug("Gateway Clinet RPC Complate:[%s,%s]",
		evt.ProtoName,
		global.RSName)
}

//clientComparator desc
//@method clientComparator desc: client object comparison function
//@param (interface{}) A Client object
//@param (interface{}) A Client object
//@return (int) 1 not equal, 0 equal
func clientComparator(a, b interface{}) int {
	c := a.(*GWClient)
	if c.handle.GetHandle() == int32(b.(uint32)) {
		return 0
	}

	return 1
}

//NewGWClientGroup desc
//@method NewGWClientGroup
//@param  (uint32) client max of number
//@return (*GWClientMgr) A Client Manager Object
func NewGWClientGroup(max uint32) *GWClientGroup {
	return &GWClientGroup{NetClientManager: implement.NetClientManager{Malloc: &GWClientAllocer{}},
		HashTable: table.HashTable{Mask: 0xFFFFFF, Max: max, Comp: clientComparator}}
}

//GWClientGroup desc
//@struct GWClientGroup desc: gateway client group
//@member (int32) current server id
//@inherit (table.HashTable)
//@member  (map[int32]implement.INetClient) socket id map client
//@inherit (sync.Mutex) synchronization mutex
type GWClientGroup struct {
	serverid int32
	table.HashTable
	implement.NetClientManager
	smp map[int32]implement.INetClient
	sync.Mutex
}

//Init desc
//@method Init desc: Initialize the gateway client manager
func (slf *GWClientGroup) Init() {
	slf.smp = make(map[int32]implement.INetClient)
	slf.HashTable.Init()
}

//Association desc
//@method Association desc: Associate current server ID\
//@param (int32) a server id
//@return (void)
func (slf *GWClientGroup) Association(id int32) {
	slf.serverid = id
}

//Occupy desc
//@method Occupy desc: occupy a client resouse
//@param (implement.INetClient) a client object
//@return (uint64) a resouse id
//@return (error) error informat
func (slf *GWClientGroup) Occupy(c implement.INetClient) (uint64, error) {
	slf.Lock()
	defer slf.Unlock()
	key, err := slf.Push(c)
	if err != nil {
		return 0, err
	}

	c.SetRef(2)
	h := c.GetID()
	nh := implement.NetHandle{}
	nh.SetValue(h)
	nh.Generate(slf.serverid, int32(key), nh.GetSocket())
	c.SetID(nh.GetValue())
	slf.smp[nh.GetSocket()] = c

	return nh.GetValue(), nil
}

//Grap desc
//@method Grap desc: return client and inc add 1
//@param (uint64) a client (Handle/ID)
//@return (implement.INetClient) a client
func (slf *GWClientGroup) Grap(h uint64) implement.INetClient {
	ea := implement.NetHandle{}
	ea.SetValue(h)

	slf.Lock()
	defer slf.Unlock()

	return slf.getClient(ea.GetHandle())
}

//GrapSocket desc
//@method GrapSocket desc: return client and ref add 1
//@param (int32) a socket id
//@return (implement.INetClient) a client
func (slf *GWClientGroup) GrapSocket(sock int32) implement.INetClient {
	slf.Lock()
	defer slf.Unlock()

	k, ok := slf.smp[sock]
	if !ok {
		return nil
	}

	k.IncRef()

	return k
}

//Erase desc
//@method Erase desc: remove client
//@param (uint64) a client is (Handle/ID)
//@return (void)
func (slf *GWClientGroup) Erase(h uint64) {
	slf.Lock()
	defer slf.Unlock()

	ea := implement.NetHandle{}
	ea.SetValue(h)
	if ea.GetSocket() > 0 {
		if _, ok := slf.smp[ea.GetSocket()]; ok {
			delete(slf.smp, ea.GetSocket())
		}
	}

	c := slf.Get(uint32(ea.GetHandle()))
	if c == nil {
		return
	}

	slf.Remove(uint32(ea.GetHandle()))

	if c.(implement.INetClient).DecRef() <= 0 {
		slf.Allocer().Delete(c.(implement.INetClient))
	}
}

//Release desc
//@method Release: release client grap
//@param (implement.INetClient): a client
func (slf *GWClientGroup) Release(net implement.INetClient) {
	slf.Lock()
	defer slf.Unlock()

	if net.DecRef() <= 0 {
		slf.Allocer().Delete(net)
	}
}

//GetHandles desc
//@method GetHandles desc: return client handles
//@param (void)
//@return ([]uint64) all client of (Handle/ID)
func (slf *GWClientGroup) GetHandles() []uint64 {
	slf.Lock()
	defer slf.Unlock()

	cs := slf.GetValues()
	if cs == nil {
		return nil
	}

	i := 0
	result := make([]uint64, len(cs))
	for _, v := range cs {
		result[i] = v.(implement.INetClient).GetID()
		i++
	}

	return result
}

func (slf *GWClientGroup) getClient(key int32) implement.INetClient {
	c := slf.Get(uint32(key))
	if c == nil {
		return nil
	}

	c.(implement.INetClient).IncRef()
	return c.(implement.INetClient)
}

//GWClientAllocer desc
//@struct GWClientAllocer desc:  Gateway Client memory allocator
//@inherit implement.IAllocer instance
type GWClientAllocer struct {
}

//New desc
//@method New desc: resource allocation
//@return (implement.INetClient) a client
func (slf *GWClientAllocer) New() implement.INetClient {
	return gwClientPool.Get().(implement.INetClient)
}

//Delete desc
//@method Delete desc: Release resources
//@param (implement.INetClient) need delete client
func (slf *GWClientAllocer) Delete(p implement.INetClient) {
	p.Shutdown()
	gwClientPool.Put(p)
}

//@gwClientName Client object name generator
var gwClientName = library.NameFactory{}

//@gwClientPool Client object pool
var gwClientPool = sync.Pool{
	New: func() interface{} {
		r := newGWClient()
		return r
	},
}

//newGWClient
//@method newGWClient desc: gateway client create
//@return (service.IService) a gateway client service/actor
func newGWClient() service.IService {
	r := service.Make(gwClientName.Spawn("Gateway/Client/"),
		func() service.IService {
			gwc := new(GWClient)
			if gwc.GetRecvBuffer() == nil {
				gwc.SetRecvBuffer(bytes.NewBuffer([]byte{}))
				gwc.GetRecvBuffer().Grow(global.EnvInstance().ClientBufferLimit)
			} else {
				gwc.GetRecvBuffer().Reset()
			}
			gwc.SetAuth(0)
			gwc.SetRef(0)

			gwc.Init()

			return gwc
		}).(*GWClient)
	return r
}
