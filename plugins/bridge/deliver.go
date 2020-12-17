package bridge

import (
	"errors"
	"rocketmqtt/conf"
	"strings"
)

const (
	Kafka = iota
	Rocketmq
)

var Delivers deliver

type PublishMessage interface {
	publish(topics map[string]bool, key string, msg *Elements) error
}

type deliver struct {
	rocketMQClients map[string]*rocketMQ
	kafkaClients    map[string]*kafka
}

var targets targetMemPool

func (d *deliver) ExistTargets() bool {
	if d.kafkaClients != nil || d.rocketMQClients != nil {
		return true
	}
	return false
}

func (d *deliver) GetrocketMQClients() map[string]*rocketMQ {
	return d.rocketMQClients
}

func (d *deliver) Publish(e *Elements) error {

	key := e.ClientID

	var bitMark int64

	switch e.Action {
	case Connect:
		//log.Debug("Connect", zap.String(e.ClientID, e.Action))
		//if config.ConnectTopic != "" {
		//	topics[config.ConnectTopic] = true
		//}
	case Publish:
		// foreach regexp map config
		if v, ok := targets.getMatch(e.Topic); ok {
			bitMark = v
		} else {
			bitMark = 0
			for _, target := range conf.RunConfig.DeliverMap {
				match := matchTopicSplit(target.NameSplit, e.Topic)
				if match {
					bitMark = bitMark | 1
				}
				bitMark = bitMark << 1
			}
			targets.storeClientTopicMatch(key, e.Topic, bitMark)
		}
	case Subscribe:
		//log.Debug("Connect", zap.String(e.ClientID, e.Action))
		//if config.SubscribeTopic != "" {
		//	topics[config.SubscribeTopic] = true
		//}
	case Unsubscribe:
		//log.Debug("Connect", zap.String(e.ClientID, e.Action))
		//if config.UnsubscribeTopic != "" {
		//	topics[config.UnsubscribeTopic] = true
		//}
	case Disconnect:
		//log.Debug("Connect", zap.String(e.ClientID, e.Action))
		//if config.DisconnectTopic != "" {
		//	topics[config.DisconnectTopic] = true
		//}
		targets.deleteClient(key)
	default:
		return errors.New("error action: " + e.Topic)
	}
	var err error
	for i := len(conf.RunConfig.DeliverMap); i >= 0; i-- {
		bit := bitMark & 1
		if bit == 1 {
			dm := conf.RunConfig.DeliverMap[i]
			switch dm.Plugin {
			case "kafka":
				err = Delivers.kafkaClients[dm.Target].publish(dm.Topic, key, e)
			case "rocketmq":
				err = Delivers.rocketMQClients[dm.Target].publish(dm.Topic, key, e, dm.Tag)
			}
		}
		bitMark = bitMark >> 1
	}
	return err
}

func match(subTopic []string, topic []string) bool {
	if len(subTopic) == 0 {
		if len(topic) == 0 {
			return true
		}
		return false
	}

	if len(topic) == 0 {
		if subTopic[0] == "#" {
			return true
		}
		return false
	}

	if subTopic[0] == "#" {
		return true
	}

	if (subTopic[0] == "+") || (subTopic[0] == topic[0]) {
		return match(subTopic[1:], topic[1:])
	}
	return false
}

func matchTopic(subTopic string, topic string) bool {
	return match(strings.Split(subTopic, "/"), strings.Split(topic, "/"))
}

func matchTopicSplit(subTopic *[]string, topic string) bool {
	return match(*subTopic, strings.Split(topic, "/"))
}
