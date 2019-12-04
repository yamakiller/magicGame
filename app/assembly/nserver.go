package assembly

import (
	"github.com/yamakiller/magicNet/engine/actor"
	"github.com/yamakiller/magicNet/handler/implement"
)

//NServerDeleate doc
//@Summary listen service deleate
//@Struct NServerDeleate
type NServerDeleate struct {
}

//Handshake doc
//@Summary accept client handshake
//@Method Handshake
//@Param implement.INetClient client interface
//@Return error
func (slf *NServerDeleate) Handshake(c implement.INetClient) error {
	return nil
}

//Decode doc
//@Summary client receive decode
//@Method Decode
//@Param actor.Context listen service context
//@Param *implement.NetListenService listen service
//@Param implement.INetClient client
//@Return error
func (slf *NServerDeleate) Decode(context actor.Context,
	nets *implement.NetListenService,
	c implement.INetClient) error {
	return nil
}

//UnOnlineNotification doc
//@Summary  Offline notification
//@Method UnOnlineNotification
//@Param  (uint64) client handle
//@Return error
func (slf *NServerDeleate) UnOnlineNotification(h uint64) error {
	return nil
}

//NServer doc
//@Summary network service
//@Struct NServer
type NServer struct {
	implement.NetListenService
}

//Initial doc
//@Summary Initialize network service
//@Method Initial
func (slf *NServer) Initial() {
	slf.NetClients.Initial()
	slf.NetListenService.Initial()
	slf.RegisterMethod(&actor.Started{}, slf.Started)
	slf.RegisterMethod(&actor.Stopping{}, slf.Stopping)
}

//Started doc
//@Summary Started Event Call Function
//@Method Started
//@Param (actor.Context) source actor context
//@Param (*actor.PID) sender actor ID
//@Param (interface{}) a message
func (slf *NServer) Started(context actor.Context,
	sender *actor.PID,
	message interface{}) {
	//Register Service name
	slf.NetListenService.Started(context, sender, message)
}

//Stopping doc
//@Method Stopping @Summary Stopping Event Call Function
//@Param (actor.Context) source actor context
//@Param (*actor.PID) sender actor ID
//@Param (interface{}) a message
func (slf *NServer) Stopping(context actor.Context, sender *actor.PID, message interface{}) {
	slf.NetListenService.Stopping(context, sender, message)
}

//Shutdown doc
//@Summary Shutdown network service
//@Method Shutdown
func (slf *NServer) Shutdown() {
	slf.NetListenService.Shutdown()
}
