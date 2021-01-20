package bridge

import (
	"errors"
	"rocketmqtt/conf"
	"time"

	"github.com/Shopify/sarama"
	"go.uber.org/zap"
	"sync"
)

// var KafkaClients map[string]*kafka

type kafka struct {
	kafakConfig conf.Kafka
	kafkaClient sarama.AsyncProducer
	timerPool   sync.Pool
	Headers     []sarama.RecordHeader
}

func InitKafka() map[string]*kafka {
	var kafkas = make(map[string]*kafka)
	for _, r := range conf.RunConfig.Plugin.Kafka {
		if !r.Enable {
			continue
		}
		c := &kafka{kafakConfig: r}
		kafkas[r.Name] = c
		c.connect()
		if conf.RunConfig.WithBrokerId != "" {
			c.Headers = append(c.Headers, sarama.RecordHeader{
				Key:   []byte("bid"),
				Value: []byte(conf.RunConfig.WithBrokerId),
			})
		}
	}
	return kafkas
}

//connect
func (k *kafka) connect() {
	conf := sarama.NewConfig()
	conf.Version = sarama.V2_2_0_0
	kafkaClient, err := sarama.NewAsyncProducer(k.kafakConfig.Addr, conf)
	if err != nil {
		log.Fatal("create kafka async producer failed: ", zap.Error(err))
	}

	go func() {
		for err := range kafkaClient.Errors() {
			log.Error("send msg to kafka failed: ", zap.Error(err))
		}
	}()

	k.timerPool = sync.Pool{
		New: func() interface{} {
			return time.NewTimer(time.Duration(5) * time.Second)
		},
	}
	k.kafkaClient = kafkaClient
}

func (k *kafka) publish(topic string, key string, msg *Elements) error {
	//for _, topic := range topics {
	t := k.timerPool.Get().(*time.Timer)

	if !t.Stop() {
		<-t.C
	}
	t.Reset(time.Duration(5) * time.Second)
	//Headers := append(k.Headers, sarama.RecordHeader{
	//	Key: []byte("topic"),
	//	Value: []byte(msg.Topic),
	//})
	//Headers = append(Headers, sarama.RecordHeader{
	//	Key: []byte("clientId"),
	//	Value: []byte(msg.ClientID),
	//})
	select {
	case k.kafkaClient.Input() <- &sarama.ProducerMessage{
		Headers: k.Headers,
		Topic:   topic,
		Key:     sarama.StringEncoder(key),
		Value:   sarama.ByteEncoder(msg.Payload),
	}:
		k.timerPool.Put(t)
		//continue
	case <-t.C:
		k.timerPool.Put(t)
		return errors.New("write kafka timeout")
	}
	//}
	return nil
}
