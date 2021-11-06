package bridge

import (
	"rocketmqtt/logger"
)

const (
	//Connect mqtt connect
	Connect = "connect"
	//Publish mqtt publish
	Publish = "publish"
	//Subscribe mqtt sub
	Subscribe = "subscribe"
	//Unsubscribe mqtt sub
	Unsubscribe = "unsubscribe"
	//Disconnect mqtt disconenct
	Disconnect = "disconnect"
)

var (
	log = logger.Instance.Named("bridge")
)

//Elements kafka publish elements
type Elements struct {
	ClientID  string `json:"clientid"`
	Username  string `json:"username"`
	Topic     string `json:"topic"`
	Payload   []byte `json:"payload"`
	Timestamp int64  `json:"ts"`
	Size      int32  `json:"size"`
	Action    string `json:"action"`
}

type BridgeMQ interface {
	Publish(e *Elements) error
	//Publish(action string, mqttTopic string, payload []byte, clientId string) error
}

func InitBridgeMQ() BridgeMQ {
	targets.topicBitMap = make(map[string]int64, 1000)
	targets.clientsMap = make(map[string]*topicMatched, 1000)
	Delivers.kafkaClients = InitKafka()
	Delivers.rocketMQClients = InitRocketMQPushConsumer()
	return &Delivers
}
