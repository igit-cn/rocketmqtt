package broker

import (
	"rocketmqtt/logger"
	"rocketmqtt/plugins/bridge"
	"go.uber.org/zap"
)

var (
	log = logger.Instance
)

func (b *Broker) Publish(e *bridge.Elements) {
	if bridge.Delivers.ExistTargets() {
		err := b.bridgeMQ.Publish(e)
		if err != nil {
			log.Error("send message to mq error.", zap.Error(err))
		}
	}
}
