package forward

import "errors"

var (
	//ErrForwardAgreeUnDefined Error informat Agreement not defined
	ErrForwardAgreeUnDefined = errors.New("Agreement not defined")
	//ErrForwardAgreeUnRegister Error informat Agreement not register
	ErrForwardAgreeUnRegister = errors.New("Agreement not register")
	//ErrForwardClientUnverified Error informat Client Unverified
	ErrForwardClientUnverified = errors.New("Client Unverified")
	//ErrForwardServiceNotStarted Error informat Service has not started yet
	ErrForwardServiceNotStarted = errors.New("Service has not started yet")
	//ErrForwardRateAbnormal Error Sending frequency is too large to block
	ErrForwardRateAbnormal = errors.New("Sending frequency is too large to block")
)

// NewTable Generate a routing table
func NewTable() *Table {
	return &Table{tb: make(map[interface{}]Unit, 128)}
}

//Table Routing address table
type Table struct {
	tb map[interface{}]Unit
}

//Register Register routing data
func (t *Table) Register(proto interface{}, protoName, serverName string, auth bool) {
	t.tb[proto] = Unit{Proto: proto, ProtoName: protoName, ServoName: serverName, Auth: auth}
}

// Sreach Query a routing information
func (t *Table) Sreach(proto interface{}) *Unit {
	if p, ok := t.tb[proto]; ok {
		return &p
	}
	return nil
}
