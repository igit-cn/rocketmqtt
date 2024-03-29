package main

import (
	"context"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"rocketmqtt/broker"
	"rocketmqtt/conf"
	"rocketmqtt/logger"
	"rocketmqtt/metric"
	"rocketmqtt/plugins/bridge"
	"runtime"

	"fmt"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/eclipse/paho.mqtt.golang/packets"
	"go.uber.org/zap"
)

var (
	log = logger.Instance.Named("main")
)

func main() {
	runtime.GOMAXPROCS(runtime.NumCPU())

	config := conf.RunConfig

	// 是否打开ppfrof性能分析
	if config.Broker.LogLevel == "debug" {
		go func() {
			log.Fatal("pprof", zap.Error(http.ListenAndServe(fmt.Sprintf("0.0.0.0:%s", config.Listen.PprofPort), nil)))
		}()
	}
	// prometheus 数据采集接口
	go metric.InitCollector()
	// 初始化broker
	b, err := broker.NewBroker(config)
	if err != nil {
		log.Fatal("New Broker error: ", zap.Error(err))
	}
	b.Start()
	broker.RunBroker = b
	for _, rmq := range bridge.Delivers.GetrocketMQClients() {
		c := rmq.GetProducer()
		if c != nil {
			subscribeRmq(c, b, rmq.GetConfig().SubscribeTopic, rmq.GetConfig().SubscribeTag)
		}
	}
	//go sendTest(b)
	s := waitForSignal()
	log.Info("signal received, broker closed.", zap.Any("signal", s))
}

func waitForSignal() os.Signal {
	signalChan := make(chan os.Signal, 1)
	defer close(signalChan)
	signal.Notify(signalChan, os.Kill, os.Interrupt)
	s := <-signalChan
	signal.Stop(signalChan)
	return s
}

// subscribe rocketmq, then send to mqtt
func subscribeRmq(rc rocketmq.PushConsumer, b *broker.Broker, topic string, tag string) {
	selector := consumer.MessageSelector{}
	if tag != "" {
		selector = consumer.MessageSelector{
			Type:       consumer.TAG,
			Expression: tag,
		}
	}
	err := rc.Subscribe(topic, selector, func(ctx context.Context,
		msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for i := range msgs {
			//tr := time.Now().UnixNano()
			msgTopic := msgs[i].GetProperty("topic")
			msgClientId := msgs[i].GetProperty("clientId")

			packet := packets.NewControlPacket(packets.Publish).(*packets.PublishPacket)
			packet.TopicName = msgs[i].GetProperty("topic")
			packet.Qos = 0
			packet.Payload = msgs[i].Body

			if msgClientId == "" || msgClientId == "-" {
				b.PublishMessage(packet)
			} else if msgTopic != "" {
				b.PublishMessageByCid(msgClientId, packet)
			} else {
				log.Warn("can't send message", zap.String("topic", msgTopic),
					zap.String("clientId", msgClientId), zap.Any("payload", packet.Payload))
			}
			// count downstream
			broker.CountIncrease(&broker.MessageDownCount)
		}
		return consumer.ConsumeSuccess, nil
	})
	if err != nil {
		log.Fatal("start producer error: %s", zap.Error(err))
	}
	err = rc.Start()
	if err != nil {
		log.Fatal("start producer error: %s", zap.Error(err))
		//os.Exit(-1)
	}
}
