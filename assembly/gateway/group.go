package gateway

import (
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/yamakiller/magicLibs/util"
	"github.com/yamakiller/magicNet/handler"
	"github.com/yamakiller/magicNet/handler/implement/buffer"
	"github.com/yamakiller/magicNet/handler/net"
)

const (
	constIDMask  = 0x1F
	constIDShift = 5
)

type clientAllocer struct {
	_parent *clientGroup
	_pool   *sync.Pool
	_sn     uint32
}

//Initial doc
//@Summary 初始化分配器
//@Method Initial
func (slf *clientAllocer) Initial() {
	slf._pool = &sync.Pool{
		New: func() interface{} {
			c := new(client)
			c.ReceiveBuffer = buffer.NewBuffer(slf._parent._bfSize)
			return c
		},
	}
}

//New doc
//@Summary Allocate a new gateway visitor object
//@Method New
//@Return net.INetClient 新的一个gateway访问者对象
func (slf *clientAllocer) New() net.INetClient {
	c := handler.Spawn(fmt.Sprintf("gateway/client/%d/%d",
		slf._parent._id,
		atomic.AddUint32(&slf._sn, 1)), func() handler.IService {

		h := slf._pool.Get().(*client)
		h.ClearBuffer()
		h.Initial()
		return h
	})

	return c.(net.INetClient)
}

//Delete doc
//@Summary 释放一个gateway访问者对象
//@Method Delete
//@Param  p net.INetClient 需要释放的对象
func (slf *clientAllocer) Delete(p net.INetClient) {
	p.Shutdown()
	slf._pool.Put(p)
}

type clientGroup struct {
	_id        int
	_handles   map[uint64]net.INetClient
	_sockets   map[int32]net.INetClient
	_snowflake *util.SnowFlake
	_allocer   *clientAllocer
	_bfSize    int
	_cap       int
	_sz        int
	_sync      sync.Mutex
}

//Initial doc
//@Summary initialization gateway server client manage
//@Method Initial
func (slf *clientGroup) Initial() {
	slf._handles = make(map[uint64]net.INetClient)
	slf._sockets = make(map[int32]net.INetClient)
	slf._allocer = &clientAllocer{_parent: slf}
	workerID := int((slf._id >> constIDShift) & constIDMask)
	subWorkerID := (slf._id & constIDMask)
	slf._snowflake = util.NewSnowFlake(int64(workerID), int64(subWorkerID))
	slf._allocer.Initial()
}

//Allocer doc
//@Summary Retruns 一个分配器
//@Method Allocer
//@Return net.IAllocer
func (slf *clientGroup) Allocer() net.IAllocer {
	return slf._allocer
}

//Occupy doc
//@Summary  occupy a client resouse
//@Method Occupy
//@Param (implement.INetClient) a client object
//@Return (uint64) a resouse id
//@Return (error) error informat
func (slf *clientGroup) Occupy(c net.INetClient) (uint64, error) {
	slf._sync.Lock()
	defer slf._sync.Unlock()

	handle, err := slf._snowflake.NextID()
	if err != nil {
		return 0, err
	}

	handleKey := uint64(handle)

	slf._handles[handleKey] = c
	slf._sockets[c.GetSocket()] = c

	c.SetRef(2)
	c.WithID(handleKey)
	slf._sz++

	return handleKey, nil
}

//Grap doc
//@Summary return client and inc add 1
//@Method Grap
//@Param (uint64) a client (Handle/ID)
//@Return (implement.INetClient) a client
func (slf *clientGroup) Grap(h uint64) net.INetClient {
	slf._sync.Lock()
	defer slf._sync.Unlock()

	if v, ok := slf._handles[h]; ok {
		v.IncRef()
		return v
	}

	return nil
}

//GrapSocket doc
//@Summary return client and ref add 1
//@Method GrapSocket desc
//@Param (int32) a socket id
//@Return (implement.INetClient) a client
func (slf *clientGroup) GrapSocket(sock int32) net.INetClient {
	slf._sync.Lock()
	defer slf._sync.Unlock()

	if v, ok := slf._sockets[sock]; ok {
		v.IncRef()
		return v
	}

	return nil
}

//Erase doc
//@Summary remove client
//@Method Erase
//@Param (uint64) a client is (Handle/ID)
func (slf *clientGroup) Erase(h uint64) {
	slf._sync.Lock()
	defer slf._sync.Unlock()

	c, ok := slf._handles[h]
	if !ok {
		return
	}

	s := c.GetSocket()
	if s > 0 {
		if _, ok = slf._sockets[s]; ok {
			delete(slf._sockets, s)
		}
	}

	delete(slf._handles, h)

	if c.DecRef() <= 0 {
		slf.Allocer().Delete(c)
	}
	slf._sz--
}

//Release doc
//@Summary release client grap
//@Method Release
//@Param implement.INetClient a client
func (slf *clientGroup) Release(c net.INetClient) {
	slf._sync.Lock()
	defer slf._sync.Unlock()

	if c.DecRef() <= 0 {
		slf.Allocer().Delete(c)
	}
}

//Size doc
//@Summary Return 连接客户端数
//@Method Size
//@Return int
func (slf *clientGroup) Size() int {
	slf._sync.Lock()
	defer slf._sync.Unlock()
	return slf._sz
}

//Cap doc
//@Summary Returns 最大客户端连接数
func (slf *clientGroup) Cap() int {
	return slf._cap
}

//GetHandles doc
//@Summary Returns 所有连接中的客户端Handle/ID
//@Method GetHandles
//@Return ([]uint64) all client of (Handle/ID)
func (slf *clientGroup) GetHandles() []uint64 {
	slf._sync.Lock()
	defer slf._sync.Unlock()

	i := 0
	result := make([]uint64, slf._sz)
	for _, v := range slf._handles {
		result[i] = v.GetID()
		i++
	}

	return result
}
