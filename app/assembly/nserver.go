package assembly

import (
	"github.com/yamakiller/magicGateway/app/global"
	"github.com/yamakiller/magicNet/engine/actor"
	"github.com/yamakiller/magicNet/handler/implement"
)

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
	global.Instance().RegisterHandle(global.ConstNetServ, context.Self())
	slf.NetListenService.Started(context, sender, message)
}

//Stopping doc
//@Method Stopping @Summary Stopping Event Call Function
//@Param (actor.Context) source actor context
//@Param (*actor.PID) sender actor ID
//@Param (interface{}) a message
func (slf *NServer) Stopping(context actor.Context, sender *actor.PID, message interface{}) {
	global.Instance().UnRegisterHandle(global.ConstNetServ)
	slf.NetListenService.Stopping(context, sender, message)
}

//Shutdown doc
//@Summary Shutdown network service
//@Method Shutdown
func (slf *NServer) Shutdown() {
	slf.NetListenService.Shutdown()
}
