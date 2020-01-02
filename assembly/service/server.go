package service

import (
	"errors"
	"sync"

	"github.com/yamakiller/magicNet/handler/net"
	rpcsrv "github.com/yamakiller/magicRpc/assembly/server"
)

//Options Service Server Options
type Options struct {
	Name         string
	ServerID     int
	Cap          int
	KeepTime     int
	BufferCap    int
	OutCChanSize int
	Compare      func(a uint64, b uint64) int
}

//Option is a function on the options for a service.
type Option func(*Options) error

//WithName setting name option
func WithName(name string) Option {
	return func(o *Options) error {
		o.Name = name
		return nil
	}
}

//WithID setting id option
func WithID(id int) Option {
	return func(o *Options) error {
		o.ServerID = id
		return nil
	}
}

//WithClientCap Set accesser cap option
func WithClientCap(cap int) Option {
	return func(o *Options) error {
		o.Cap = cap
		return nil
	}
}

//WithClientKeepTime Set client keep time millsecond option
func WithClientKeepTime(tm int) Option {
	return func(o *Options) error {
		o.KeepTime = tm
		return nil
	}
}

//WithClientBufferCap Set client buffer limit option
func WithClientBufferCap(cap int) Option {
	return func(o *Options) error {
		o.BufferCap = cap
		return nil
	}
}

//WithClientOutSize Set client recvice call chan size option
func WithClientOutSize(outSize int) Option {
	return func(o *Options) error {
		o.OutCChanSize = outSize
		return nil
	}
}

//WithCompare Set Client find Compare function
func WithCompare(f func(a uint64, b uint64) int) Option {
	return func(o *Options) error {
		o.Compare = f
		return nil
	}
}

//New Create service
func New(options ...Option) (*Server, error) {

	opts := Options{}
	for _, opt := range options {
		if err := opt(&opts); err != nil {
			return nil, err
		}
	}

	srv := &Server{}

	rpcSrv, err := rpcsrv.New(
		rpcsrv.WithName(opts.Name),
		rpcsrv.WithID(opts.ServerID),
		rpcsrv.WithClientKeepTime(opts.KeepTime),
		rpcsrv.WithClientCap(opts.Cap),
		rpcsrv.WithClientBufferCap(opts.BufferCap),
		rpcsrv.WithClientOutSize(opts.OutCChanSize),
		rpcsrv.WithAsyncAccept(srv.asyncAccept),
	)

	if err != nil {
		return nil, err
	}

	srv._ss = make(map[uint64]uint64)
	srv._compare = opts.Compare
	srv._rpcServer = rpcSrv
	srv._rpcServer.RegRPC(&regCtrl{srv})

	return srv, nil
}

//Server service
type Server struct {
	_rpcServer *rpcsrv.RPCServer
	_ss        map[uint64]uint64 //[clietn id]socket handle
	_compare   func(a uint64, b uint64) int
	_sync      sync.RWMutex
}

//Listen listening
func (slf *Server) Listen(addr string) error {
	return slf._rpcServer.Listen(addr)
}

//Shutdown close server
func (slf *Server) Shutdown() {
	if slf._rpcServer != nil {
		slf._rpcServer.Shutdown()
	}
}

//PutCtrl put control
func (slf *Server) PutCtrl(ctrl interface{}) error {
	return slf._rpcServer.RegRPC(ctrl)
}

//Call call object client function
func (slf *Server) Call(client uint64, method string, param interface{}) error {

	var handle uint64
	slf._sync.RLock()
	for clientHandle, clientSocketHandle := range slf._ss {
		if slf._compare(clientHandle, client) == 0 {
			handle = clientSocketHandle
			break
		}
	}
	slf._sync.Unlock()
	if handle == 0 {
		return errors.New("unknown client")
	}
	return slf._rpcServer.Call(handle, method, param)
}

func (slf *Server) asyncAccept(socketHandle uint64) {
}

func (slf *Server) asyncClosed(socketHandle uint64) {
	slf._sync.Lock()
	defer slf._sync.Unlock()

	for k, v := range slf._ss {
		if v == socketHandle {
			delete(slf._ss, k)
		}
	}
}

type regCtrl struct {
	_parent *Server
}

//Register client connection
func (slf *regCtrl) SignIn(c net.INetClient, request *SignInReq) *SignInRsp {
	handle := c.GetID()

	slf._parent._sync.Lock()
	defer slf._parent._sync.Unlock()

	if v, ok := slf._parent._ss[request.ClientHandle]; ok {
		if v != handle {
			slf._parent._rpcServer.CloseClient(v)
		}
	}

	slf._parent._ss[request.ClientHandle] = handle
	return &SignInRsp{Code: 0}
}
