package gateway

import (
	"sync"

	"github.com/gogo/protobuf/proto"
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
	Name         string
	ServerID     int
	SocketMode   int
	BufferLimit  int
	KeepTime     int
	OutCChanSize int
	Cap          int
	Replicas     int
	Decoder      net.INetDecoder
	AsyncAccept  func(net.INetClient) error
	AsyncClosed  func(uint64) error
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

//SetClientDecoder doc
//@Summary Set the connection client data decoder
//@Method SetClientDecoder
//@Param  net.INetDecoder decoder
//@Return Option
func SetClientDecoder(d net.INetDecoder) Option {
	return func(o *Options) error {
		o.Decoder = d
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

//SetAsyncAccept doc
//@Summary  Set listen accept asynchronous callback function
//@Method   SetAsyncAccept
//@Param    func(net.INetClient) error  Callback
//@Return   Option
func SetAsyncAccept(f func(net.INetClient) error) Option {
	return func(o *Options) error {
		o.AsyncAccept = f
		return nil
	}
}

//SetAsyncClose doc
//@Summary Set the client to close the asynchronous callback function
//@Method Close
//@Param  func(uint64) error Callback
//@Return Option
func SetAsyncClosed(f func(uint64) error) Option {
	return func(o *Options) error {
		o.AsyncClosed = f
		return nil
	}
}

//New 创建一个网关服务,并设置相关参数
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
			listener.SetAsyncAccept(opts.AsyncAccept),
			listener.SetAsyncClosed(opts.AsyncClosed),
			listener.SetClientGroups(cGroup),
			listener.SetClientDecoder(opts.Decoder),
		)

		if err != nil {
			return nil
		}

		srv._name = opts.Name
		srv._listenHandle = h
		srv._rss = NewRouteSet(opts.Replicas)
		srv._listenHandle.Initial()
		return srv._listenHandle
	})

	return srv, nil
}

type Server struct {
	_name         string
	_listenHandle *listener.NetListener
	_listenWait   sync.WaitGroup
	_rss          *RouteSet
	_err          error
}

//Control 创建一个控制器
func (slf *Server) Control(opts *RouteOption) (*RouteCtrl, error) {
	return newCtrl(opts)
}

//Router 添加一个路由
func (slf *Server) Router(addr string, server string, ctrl *RouteCtrl) {
	slf._rss.Register(addr, server, ctrl)
}

//Router 通过路由动态调用Retmote方法
func (slf *Server) RouteCall(addr, method string, param, ret proto.Message) error {
	return slf._rss.Call(addr, method, param, ret)
}

//Listen 监听服务
func (slf *Server) Listen(addr string) error {
	slf._listenWait.Add(1)
	if err := slf._listenHandle.Listen(addr); err != nil {
		slf._listenWait.Done()
		return err
	}

	slf._listenWait.Wait()
	return slf._err
}

func (slf *Server) asyncError(err error) {
	defer slf._listenWait.Done()
	slf._err = err
}

func (slf *Server) asyncComplate(sock int32) {
	defer slf._listenWait.Done()
	slf._err = nil
}

//Shutdown 终止服务
func (slf *Server) Shutdown() {
	if slf._listenHandle != nil {
		slf._listenHandle.Shutdown()
		slf._listenHandle = nil
	}
}
