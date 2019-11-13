package events

type RouteServerEvent struct {
	Handle    uint64
	ProtoName interface{}
	ServoName string
	Serial    uint32
	Data      []byte
}

type RouteClientEvent struct {
	Handle    uint64
	ProtoName interface{}
	ServoName string
	Serial    uint32
	Data      []byte
}
