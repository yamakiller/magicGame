package library

import (
	"github.com/yamakiller/magicNet/engine/actor"
	"github.com/yamakiller/magicNet/service/implement"
)

//ListenerDeleate desc
//@struct  ListenerDeleate desc:
type ListenerDeleate struct {
}

//Handshake desc
//@method Handshake desc
func (slf *ListenerDeleate) Handshake(c implement.INetClient) error {
	return nil
}

//Analysis desc
//@method Analysis desc: Analysis of network data
//@param (actor.Context) current service/actor context
//@param (*implement.NetListenService) current network listen service
//@param (implement.INetClient) current recvice data client
//@return fail:return error,success return nil
func (slf *ListenerDeleate) Analysis(context actor.Context,
	nets *implement.NetListenService,
	c implement.INetClient) error {
	return nil
}

//UnOnlineNotification desc
//@method UnOnlineNotification desc:  Offline notification
//@param  (uint64) client handle
func (slf *ListenerDeleate) UnOnlineNotification(h uint64) error {
	return nil
}

//Listener desc
//@struct Listener desc:
type Listener struct {
	implement.NetListenService
}

//Init desc
//@method Init desc: initialize gateway listener
func (slf *Listener) Init() {
	slf.NetClients.Init()
	slf.NetListenService.Init()
}
