package parts

import (
	"github.com/yamakiller/magicBehavior/core"
	"github.com/yamakiller/magicNet/engine/actor"
	"github.com/yamakiller/magicNet/service"
	"github.com/yamakiller/magicParts/robot/events"
)

//IRobotPlay desc
//@interface IRobotPlay robot player instance
type IRobotPlay interface {
	Tick(delta int64)
}

//RobotPlayer desc
//@struct RobotPlayer desc: Single robot actor
//@inherit (service.Service)
//@inherit (implement.NetConnectService)
//@member (*core.Ticker) robot behavior tick
type RobotPlayer struct {
	service.Service
	//implement.NetConnectService
	behaviorTick *core.Ticker
}

//Init desc
//@method Init desc:  Initialize the robot
func (slf *RobotPlayer) Init() {
	slf.Service.Init()
	slf.RegisterMethod(&events.TickEvent{}, slf.onTick)
}

//Tick desc
//@method Tick desc: Timed interval execution function
//@param (int) time interval millsec
func (slf *RobotPlayer) Tick(delta int64) {
	actor.DefaultSchedulerContext.Send(
		slf.GetPID(),
		&events.TickEvent{Delta: delta})
}

func (slf *RobotPlayer) onTick(context actor.Context,
	sender *actor.PID,
	message interface{}) {
	if slf.behaviorTick == nil ||
		slf.behaviorTick.GetTree() == nil {
		return
	}

	slf.behaviorTick.GetTree().Tick(0)
}
