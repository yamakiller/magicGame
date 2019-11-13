package events

import (
	"github.com/yamakiller/magicNet/engine/actor"
	"github.com/yamakiller/magicNet/service/implement"
)

//ResponseNetMethod Intranet packet event
type ResponseNetMethodEvent struct {
	H uint64
	S uint32
	implement.NetMethodEvent
}

//ResponseNetMethod Intranet client packet event
type ResponseNetMethodClientEvent struct {
	Context actor.Context
	ResponseNetMethodEvent
}
