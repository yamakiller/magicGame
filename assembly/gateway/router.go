package gateway

import (
	"sync/atomic"

	"github.com/gogo/protobuf/proto"
	"github.com/yamakiller/magicLibs/router"
	rpcc "github.com/yamakiller/magicRpc/assembly/client"
)

func ctrlDelete(p router.IRouteCtrl) {
	p.(*RouteCtrl).Shutdown()
}

//RouteOption doc
//@Summary route option informat
type RouteOption struct {
	Server        string `xml:"server" yaml:"server" json:"server"`
	ServerAddr    string `xml:"server-address" yaml:"server address" json:"server-address"`
	Buffer        int    `xml:"buffer" yaml:"buffer" json:"buffer"`
	OutCChanSize  int    `xml:"occs" yaml:"occs" json:"occs"`
	SocketTimeout int    `xml:"socket-timeout" yaml:"socket timeout" json:"socket-timeout"`
	TimeOut       int    `xml:"timeout" yaml:"timeout" json:"timeout"`
	Idle          int    `xml:"idle" yaml:"idle" json:"idle"`
	Active        int    `xml:"active" yaml:"active" json:"active"`
	IdleTimeout   int    `xml:"idle-timeout" yaml:"idle timeout" json:"idle-timeout"`
}

func newCtrl(opts *RouteOption, asyncConnected func(*rpcc.RPCClient)) (*RouteCtrl, error) {
	rcl := &RouteCtrl{}
	p, err := rpcc.New(
		rpcc.WithName(opts.Server),
		rpcc.WithAddr(opts.ServerAddr),
		rpcc.WithBufferCap(opts.Buffer),
		rpcc.WithOutChanSize(opts.OutCChanSize),
		rpcc.WithIdle(opts.Idle),
		rpcc.WithActive(opts.Active),
		rpcc.WithSocketTimeout(int64(opts.SocketTimeout)),
		rpcc.WithTimeout(int64(opts.TimeOut)),
		rpcc.WithIdleTimeout(int64(opts.IdleTimeout)),
		rpcc.WithAsyncConnected(asyncConnected),
	)

	if err != nil {
		return nil, err
	}

	rcl._pool = p

	return rcl, nil
}

//RouteCtrl doc
//@Summary route control
type RouteCtrl struct {
	_pool *rpcc.RPCClientPool
	_ref  int32
}

//GetName Return Control name
func (slf *RouteCtrl) GetName() string {
	return slf._pool.GetName()
}

//IncRef Increment the reference counter
func (slf *RouteCtrl) IncRef() {
	atomic.AddInt32(&slf._ref, 1)
}

//DecRef Decrement the reference counter
func (slf *RouteCtrl) DecRef() int {
	return int(atomic.AddInt32(&slf._ref, -1))
}

//Call remote method
func (slf *RouteCtrl) Call(remoteMethod string, param interface{}, ret interface{}) error {
	return slf._pool.Call(remoteMethod, param, ret)
}

//Shutdown shutdown route control
func (slf *RouteCtrl) Shutdown() {
	slf._pool.Shutdown()
}

//NewRouteSet Create a route set
func NewRouteSet(reps int) *RouteSet {
	return &RouteSet{_r: router.New(ctrlDelete, reps)}
}

//RouteSet route sets
type RouteSet struct {
	_r *router.RouteGroup
}

//IsExist Whether the destination route exists
func (slf *RouteSet) IsExist(addr, srvAddr string) bool {
	return slf._r.IsExist(addr, srvAddr)
}

//Register Registered a route
func (slf *RouteSet) Register(addr, srvAddr string, c *RouteCtrl) {
	slf._r.Register(addr, srvAddr, c)
}

//UnRegister Unregister a route
func (slf *RouteSet) UnRegister(addr, srvAddr string) {
	slf._r.UnRegister(addr, srvAddr)
}

//Call Call a specified remote method
func (slf *RouteSet) Call(addr, method string, param, ret proto.Message) error {
	return slf._r.Call(addr, method, param, ret)
}

//Shutdown shutdown the route set
func (slf *RouteSet) Shutdown() {
	if slf._r != nil {
		slf._r.Shutdown()
		slf._r = nil
	}
}
