package global

import "github.com/yamakiller/magicNet/engine/actor"

var (
	//NetworkServiceHandle network listen serivce pid
	NetworkServiceHandle *actor.PID
	//NetworkRouterServiceHandle network router service pid
	NetworkRouterServiceHandle *actor.PID
)
