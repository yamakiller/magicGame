package events

//TickEvent desc
//@method TickEvent desc: Robot Tick event
//@member (int64) Tick ​​interval Unit millisecond
type TickEvent struct {
	Delta int64
}
