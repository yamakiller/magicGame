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

type RouteOption struct {
	Server        string `xml:"server" yaml:"server" json:"server"`
	ServerAddr    string `xml:"server-address" yaml:"server address" json:"server-address"`
	Buffer        int    `xml:"buffer" yaml:"buffer" json:"buffer"`
	OutCChanSize  int    `xml:"occs" yaml:"occs" json:"occs"`
	SocketTimeout int    `xml:"socket-timeout" yaml:"socket timeout" json:"socket-timeout"`
	TimeOut       int    `xml:"timeout" yaml:"timeout" json:"timeout"`
	Idle          int    `xml:"idle" yaml:"idle" json:"idle"`
	Active        int    `xml:"active" yaml:"active" json:"active"`
	IdleTimeout   int    `xml:"idle-timeoust" yaml:"idle timeout" json:"idle-timeout"`
}

func newCtrl(opts *RouteOption) (*RouteCtrl, error) {
	rcl := &RouteCtrl{}
	p, err := rpcc.New(
		rpcc.SetName(opts.Server),
		rpcc.SetAddr(opts.ServerAddr),
		rpcc.SetBufferLimit(opts.Buffer),
		rpcc.SetOutChanSize(opts.OutCChanSize),
		rpcc.SetIdle(opts.Idle),
		rpcc.SetActive(opts.Active),
		rpcc.SetSocketTimeout(int64(opts.SocketTimeout)),
		rpcc.SetTimeout(int64(opts.TimeOut)),
		rpcc.SetIdleTimeout(int64(opts.IdleTimeout)),
	)

	if err != nil {
		return nil, err
	}

	rcl._pool = p

	return rcl, nil
}

type RouteCtrl struct {
	_pool *rpcc.RPCClientPool
	_ref  int32
}

func (slf *RouteCtrl) GetName() string {
	return slf._pool.GetName()
}

func (slf *RouteCtrl) IncRef() {
	atomic.AddInt32(&slf._ref, 1)
}

func (slf *RouteCtrl) DecRef() int {
	return int(atomic.AddInt32(&slf._ref, -1))
}

func (slf *RouteCtrl) Call(method string, param interface{}, ret interface{}) error {
	return slf._pool.Call(method, param, ret)
}

func (slf *RouteCtrl) Shutdown() {
	slf._pool.Shutdown()
}

func NewRouteSet(reps int) *RouteSet {
	return &RouteSet{_r: router.New(ctrlDelete, reps)}
}

type RouteSet struct {
	_r *router.RouteGroup
}

func (slf *RouteSet) Register(addr, srvAddr string, c *RouteCtrl) {
	slf._r.Register(addr, srvAddr, c)
}

func (slf *RouteSet) UnRegister(addr, key string) {
	slf._r.UnRegister(addr, key)
}

func (slf *RouteSet) Call(addr, method string, param, ret proto.Message) error {
	return slf._r.Call(addr, method, param, ret)
}

func (slf *RouteSet) Shutdown() {
	if slf._r != nil {
		slf._r.Shutdown()
		slf._r = nil
	}
}
