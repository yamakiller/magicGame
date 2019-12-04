package app

import (
	"github.com/yamakiller/magicLibs/envs"
	"github.com/yamakiller/magicLibs/logger"
	"github.com/yamakiller/magicNet/core"
)

//GWDeploy doc
//@Summary gateway config data
//@Struct GWDeploy
//@Member int client max connection of number
//@Member int client receive buffer byte size
//@Member int client receive event chan size
type GWDeploy struct {
	Clients                int `yaml:"clients"`
	ClientReceive          int `yaml:"client_receive_buffer"`
	ClinetReceiveEventChan int `yaml:"client_receive_event_chan"`
}

//Gateway doc
//@Summary gateway server
//@Struct Gateway
//@Member (logger.Logger) log object
type Gateway struct {
	core.DefaultBoot
	core.DefaultService
	core.DefaultWait
}

//Initial doc
//@Summary Initialize gateway
//@Method Initial
//@Return error: Initialize fail return error
func (slf *Gateway) Initial() error {
	//Build global environment variable Manager
	//------------
	env := &envs.YAMLEnv{}
	env.Initial()
	envs.With(env)
	//------------------------------------------
	return slf.DefaultBoot.Initial()
}

//InitService doc
//@Summary init gateway system
//@Method InitService
//@Return (error) a error informat
func (slf *Gateway) InitService() error {
	return nil
}

//CloseService doc
//@Summary Close service
//@Method CloseService desc
func (slf *Gateway) CloseService() {
}

//Destory doc
//@Summary destory system reouse
//@Method Destory
func (slf *Gateway) Destory() {
	logger.Info(0, "Gateway Destory")
	slf.DefaultBoot.Destory()
}
