package bridge

import (
	"errors"
	"rocketmqtt/conf"
	"strings"

	"go.uber.org/zap"
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
			for _, target := range conf.RunConfig.DeliversRules {
				match := matchTopicSplit(target.NameSplit, e.Topic)
				if match {
					bitMark = bitMark | 1
				}
				bitMark = bitMark << 1
			}
			targets.storeClientTopicMatch(e.ClientID, e.Topic, bitMark)
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
		targets.deleteTopicMatch(e.ClientID)
	case Disconnect:
		//log.Debug("Connect", zap.String(e.ClientID, e.Action))
		//if config.DisconnectTopic != "" {
		//	topics[config.DisconnectTopic] = true
		//}
		targets.deleteTopicMatch(e.ClientID)
	default:
		return errors.New("error action: " + e.Topic)
	}
	var err error
	if bitMark == 0 {
		log.Warn("No matched deliver rule", zap.String("ClientID", e.ClientID), zap.Any("element", e))
		return nil
	}
	for i := len(conf.RunConfig.DeliversRules); i >= 0; i-- {
		bit := bitMark & 1
		if bit == 1 {
			dm := conf.RunConfig.DeliversRules[i]
			switch dm.Plugin {
			case "kafka":
				err = Delivers.kafkaClients[dm.Target].publish(dm.Topic, e.ClientID, e)
			case "rocketmq":
				err = Delivers.rocketMQClients[dm.Target].publish(dm.Topic, e.ClientID, e, dm.Tag)
			default:
				log.Warn("plugin not defined", zap.String("plugin name", dm.Plugin))
			}
		}
		bitMark = bitMark >> 1
	}
	return err
}

func match(subTopic []string, topic []string) bool {
	if len(subTopic) == 0 {
		return len(topic) == 0
	}

	if len(topic) == 0 {
		return subTopic[0] == "#"
	}

	if subTopic[0] == "#" {
		return true
	}

	if (subTopic[0] == "+") || (subTopic[0] == topic[0]) {
		return match(subTopic[1:], topic[1:])
	}
	return false
}

// func matchTopic(subTopic string, topic string) bool {
// 	return match(strings.Split(subTopic, "/"), strings.Split(topic, "/"))
// }

func matchTopicSplit(subTopic *[]string, topic string) bool {
	return match(*subTopic, strings.Split(topic, "/"))
}
