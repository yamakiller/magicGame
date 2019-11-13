package global

import (
	"github.com/yamakiller/magicParts/gateway/elements/forward"
	"github.com/yamakiller/magicParts/gateway/elements/target"
	"github.com/yamakiller/magicParts/library"
)

var (
	//FAddr
	FAddr *forward.Table
	//SSets Service set
	SSets *library.SSets
	//TSets Connection configuration status information of the target server
	TSets *target.TSet
	//
	TLSets *target.TLoadSet
	//
	RSName string
	//
	NSName string
)

func init() {
	FAddr = forward.NewTable()
	SSets = library.NewSSets()
	TSets = target.NewTargetSet()
	TLSets = target.NewLoadSet()
}
