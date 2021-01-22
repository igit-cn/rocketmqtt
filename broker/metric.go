package broker

import "sync/atomic"

var ClientCount uint64
var MessageDownCount uint64
var MessageUpCount uint64

func CountIncrease(c *uint64) {
	atomic.AddUint64(c, 1)
}

func (b *Broker) SessionCount() int {
	return b.sessionMgr.Count()
}

func (b *Broker) ConnectionCount() int {
	c := 1
	b.clients.Range(func(key, value interface{}) bool {
		c += 1
		return true
	})
	return c
}
