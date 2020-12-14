package broker

import "sync/atomic"


var ClientCount uint64
var MessageDownCount uint64
var MessageUpCount uint64

func CountIncrease(c *uint64){
	atomic.AddUint64(c, 1)
}

func (b *Broker) SessionCount() int {
	return b.sessionMgr.Count()
}
