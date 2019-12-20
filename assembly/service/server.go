package service

import (
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

	srv._rpcServer = rpcSrv

	return srv, nil
}

//Server service
type Server struct {
	_rpcServer *rpcsrv.RPCServer
	_ss
}

func (slf *Server) asyncAccept(clientHandle uint64) {

}

func (slf *Server) asyncClosed(clientHandle uint64) {

}
