package framework

import (
	"github.com/yamakiller/magicNet/core"
	"github.com/yamakiller/magicParts/gateway/parts"
)

//GWFrame desc
//@struct GWFrame desc: Gateway infrastructure module
//@inherit (core.DefaultStart)
//@inherit (core.DefaultEnv)
//@inherit (core.DefaultLoop)
//@inherit (core.DefaultService)
//@inherit (core.DefaultCMDLineOption)
//@member  (*parts.GWLua) Script service module for executing the startup initialization script
//@member  (*parts.GWRouter) For routing services
//@member  (*parts.GWListener) Network monitoring service
type GWFrame struct {
	core.DefaultStart
	core.DefaultEnv
	core.DefaultLoop
	core.DefaultService
	core.DefaultCMDLineOption

	luaService *parts.GWLua
	rouService *parts.GWRouter
	nstService *parts.GWListener
}
