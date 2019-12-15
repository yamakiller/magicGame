package gateway

import (
	"sync"
	"time"

	"github.com/yamakiller/magicLibs/coroutine"

	"github.com/yamakiller/magicNet/network"

	"github.com/gogo/protobuf/proto"
	"github.com/yamakiller/magicNet/engine/actor"
	"github.com/yamakiller/magicNet/handler"
	"github.com/yamakiller/magicNet/handler/implement/listener"
	"github.com/yamakiller/magicNet/handler/net"
)

const (
	//TCPNet tcp mode
	TCPNet = 0
	//WWSNet websocket mode
	WWSNet = 1
)

type Options struct {
	Name          string
	ServerID      int
	SocketMode    int
	BufferLimit   int
	KeepTime      int
	OutCChanSize  int
	Cap           int
	Replicas      int
	AuthTimeout   int64
	GuardInterval int64
	Delegate      IServerDelegate
}

type Option func(*Options) error

func SetName(name string) Option {
	return func(o *Options) error {
		o.Name = name
		return nil
	}
}

// SetID setting id option
func SetID(id int) Option {
	return func(o *Options) error {
		o.ServerID = id
		return nil
	}
}

//SetClientCap Set accesser cap option
func SetClientCap(cap int) Option {
	return func(o *Options) error {
		o.Cap = cap
		return nil
	}
}

//SetClientBufferLimit Set client buffer limit option
func SetClientBufferLimit(limit int) Option {
	return func(o *Options) error {
		o.BufferLimit = limit
		return nil
	}
}

func SetRouteReplicas(reps int) Option {
	return func(o *Options) error {
		o.Replicas = reps
		return nil
	}
}

//SetClientOutChanSize doc
//@Summary Set the connection client transaction pipeline buffer size
//@Method SetClientOutChanSize
//@Param  int Pipe buffer size
func SetClientOutChanSize(ch int) Option {
	return func(o *Options) error {
		o.OutCChanSize = ch
		return nil
	}
}

//SetClientKeepTime doc
//@Summary Set the heartbeat interval of the connected client in milliseconds
//@Param   int Interval time in milliseconds
//@Return  Option
func SetClientKeepTime(tm int) Option {
	return func(o *Options) error {
		o.KeepTime = tm
		return nil
	}
}

func SetAuthTimeout(tm int64) Option {
	return func(o *Options) error {
		o.AuthTimeout = tm
		return nil
	}
}

func SetGuardInterval(tm int64) Option {
	return func(o *Options) error {
		o.GuardInterval = tm
		return nil
	}
}

func SetDelegate(delegate IServerDelegate) Option {
	return func(o *Options) error {
		o.Delegate = delegate
		return nil
	}
}

var (
	defaultOption = Options{Name: "Gateway",
		ServerID:      1,
		SocketMode:    TCPNet,
		BufferLimit:   8196,
		KeepTime:      5 * 1000,
		OutCChanSize:  32,
		Cap:           4096,
		Replicas:      32,
		AuthTimeout:   2 * 1000,
		GuardInterval: 5 * 1000,
	}
)

//New Create a gateway service and set related parameters
func New(options ...Option) (*Server, error) {
	opts := Options{}
	for _, opt := range options {
		if err := opt(&opts); err != nil {
			return nil, err
		}
	}

	srv := &Server{}
	handler.Spawn(opts.Name, func() handler.IService {
		cGroup := &clientGroup{_id: opts.ServerID, _bfSize: opts.BufferLimit, _cap: opts.Cap}
		var s net.INetListener
		if opts.SocketMode == TCPNet {
			s = &net.TCPListen{}
		} else {
			s = &net.WSSListen{}
		}

		h, err := listener.Spawn(
			listener.SetListener(s),
			listener.SetAsyncError(srv.asyncError),
			listener.SetClientKeepTime(opts.KeepTime),
			listener.SetClientOutChanSize(opts.OutCChanSize),
			listener.SetAsyncComplete(srv.asyncComplate),
			listener.SetAsyncAccept(srv.asyncAccept),
			listener.SetAsyncClosed(srv.asyncClosed),
			listener.SetClientGroups(cGroup),
			listener.SetClientDecoder(srv.defaultDecode),
		)

		if err != nil {
			return nil
		}

		srv._name = opts.Name
		srv._listenHandle = h
		srv._delegate = opts.Delegate
		srv._authTimeout = opts.AuthTimeout
		srv._guardInterval = opts.GuardInterval
		srv._rss = NewRouteSet(opts.Replicas)
		srv._listenHandle.Initial()
		return srv._listenHandle
	})

	return srv, nil
}

type IServerDelegate interface {
	AsyncDecode(net.INetClient) (*AgreMessage, error)
	AsyncEncode(response interface{}) ([]byte, error)
	AsyncAccept(net.INetClient) error
	AsyncClosed(uint64) error

	QueryAsyncMethod(interface{}) (string, interface{}, bool, error)
}

//Server doc: Gateway Server
type Server struct {
	_delegate      IServerDelegate
	_name          string
	_listenHandle  *listener.NetListener
	_listenWait    sync.WaitGroup
	_rss           *RouteSet
	_authTimeout   int64
	_guardInterval int64
	_err           error
	_ishutdown     bool
}

//Control Create a Control
//@Param  *RouteOption control option
//@Return *RouteCtrl
//@Return  error
func (slf *Server) Control(opts *RouteOption) (*RouteCtrl, error) {
	return newCtrl(opts)
}

//Router Add a route
//@Param route address
//@Param route server name
//@Param network agreement name/object
//@Param network agreement => method
//@Param route control option
func (slf *Server) Router(addr string, server string, ctrl *RouteCtrl) {
	slf._rss.Register(addr, server, ctrl)
}

//Router Dynamically calling the Retmote method via a route
func (slf *Server) RouteCall(addr, method string, param, ret proto.Message) error {
	return slf._rss.Call(addr, method, param, ret)
}

//Listen Start listen
func (slf *Server) Listen(addr string) error {
	slf._listenWait.Add(1)
	if slf._err = slf._listenHandle.Listen(addr); slf._err != nil {
		slf._listenWait.Done()
		return slf._err
	}

	slf._listenWait.Wait()
	return slf._err
}

func (slf *Server) defaultDecode(context actor.Context, params ...interface{}) error {
	c := params[1].(client)
	argee, err := slf._delegate.AsyncDecode(params[1].(net.INetClient))
	if err != nil {
		return err
	}

	actor.DefaultSchedulerContext.Send(c.GetPID(), argee)

	return nil
}

func (slf *Server) asyncGuard([]interface{}) {
	startTime := (time.Now().UnixNano() / int64(time.Millisecond))
	curreTime := startTime
	interval := curreTime
	defer slf._listenWait.Done()
	for {
		if slf._ishutdown {
			break
		}

		curreTime = (time.Now().UnixNano() / int64(time.Millisecond))
		interval = curreTime - startTime
		var c net.INetClient
		var cs *client
		clients := slf._listenHandle.GetClients()
		for _, v := range clients {
			c = slf._listenHandle.Grap(v)
			if c == nil {
				continue
			}

			cs = c.(*client)
			if cs._auth > 1 {
				continue
			}

			if cs._auth == 0 {
				continue
			}

			cs._authLastTime -= interval
			if cs._authLastTime <= 0 {
				network.OperClose(cs.GetSocket())
			}
		}
		startTime = curreTime
		time.Sleep(time.Duration(slf._guardInterval) * time.Millisecond)

	}
}

func (slf *Server) asyncAccept(c net.INetClient) error {
	c.(*client)._parent = slf
	c.(*client)._auth = 1
	c.(*client)._authLastTime = slf._authTimeout
	network.OperOpen(c.GetSocket())
	if slf._delegate != nil {
		return slf._delegate.AsyncAccept(c)
	}
	return nil
}

func (slf *Server) asyncClosed(h uint64) error {
	if slf._delegate != nil {
		return slf._delegate.AsyncClosed(h)
	}
	return nil
}

func (slf *Server) asyncError(err error) {
	defer slf._listenWait.Done()
	slf._err = err
}

func (slf *Server) asyncComplate(sock int32) {
	defer slf._listenWait.Done()
	slf._err = nil
	slf._listenWait.Add(1)
	coroutine.Instance().Go(slf.asyncGuard)
}

//Shutdown shutdown server
func (slf *Server) Shutdown() {
	slf._ishutdown = true
	slf._listenWait.Wait()
	if slf._listenHandle != nil {
		slf._listenHandle.Shutdown()
		slf._listenHandle = nil
	}
}
