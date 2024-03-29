package bridge

import (
	"context"
	"fmt"
	"os"
	"rocketmqtt/conf"
	"rocketmqtt/logger"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"github.com/apache/rocketmq-client-go/v2/rlog"
	"go.uber.org/zap"
)

type rocketMQ struct {
	rocketMQConfig       conf.Rocketmq
	rocketMQPushConsumer rocketmq.PushConsumer
	rocketMQProducer     rocketmq.Producer
}

//Init RocketMQ Push client
func InitRocketMQPushConsumer() map[string]*rocketMQ {
	rmqLogger := logger.RmqLogger{}
	rmqLogger.Init(logger.Instance.Sugar().Named("rocketmq"))
	rlog.SetLogger(&rmqLogger)
	var rmqs = make(map[string]*rocketMQ)
	for _, r := range conf.RunConfig.Plugins.Rocketmq {
		if !r.Enable {
			continue
		}
		c := &rocketMQ{rocketMQConfig: r}
		rmqs[r.Name] = c
		c.connect()
	}
	return rmqs
}

func (r *rocketMQ) GetProducer() rocketmq.PushConsumer {
	return r.rocketMQPushConsumer
}
func (r *rocketMQ) GetConfig() conf.Rocketmq {
	return r.rocketMQConfig
}

func (r *rocketMQ) connect() {
	ns, err := primitive.NewNamesrvAddr(r.rocketMQConfig.NameSrv)
	if err != nil {
		log.Fatal("name server error: ", zap.Error(err))
	}
	var c rocketmq.PushConsumer
	if r.rocketMQConfig.EnableSubscribe {
		msgModel := consumer.Clustering
		if r.rocketMQConfig.SubscribeModel == "BroadCasting" {
			msgModel = consumer.BroadCasting
		}
		c, err = rocketmq.NewPushConsumer(
			consumer.WithGroupName(r.rocketMQConfig.GroupName),
			consumer.WithNameServer(ns),
			consumer.WithConsumerModel(msgModel),
			consumer.WithInstance(fmt.Sprintf("%s-%s", conf.RunConfig.Broker.ID, r.rocketMQConfig.GroupName)),
		)
		if err != nil {
			log.Fatal("new push consumer error: ", zap.Error(err))
		}
	}
	p, _ := rocketmq.NewProducer(
		producer.WithNameServer(ns),
		producer.WithRetry(2),
		producer.WithGroupName(r.rocketMQConfig.GroupName),
		//producer.WithInstanceName(r.rocketMQConfig.GroupName),
	)
	err = p.Start()
	if err != nil {
		log.Fatal("start producer error: %s", zap.Error(err))
		os.Exit(1)
	}

	r.rocketMQPushConsumer = c
	r.rocketMQProducer = p
}

func (r *rocketMQ) publish(topic string, key string, msg *Elements, tag string) error {

	//log.Debug("send rmq",zap.Any("msg", payload))

	// payload, err := json.Marshal(msg)
	// if err != nil {
	// 	return err
	// }

	//for _, topic := range topics {
	rmsg := primitive.NewMessage(topic,
		msg.Payload)
	rmsg.WithProperty("clientId", msg.ClientID)
	rmsg.WithProperty("topic", msg.Topic)
	if conf.RunConfig.Broker.ID != "" {
		rmsg.WithProperty("bid", conf.RunConfig.Broker.ID)
	}
	if tag != "" {
		rmsg.WithTag(tag)
	}

	// sync
	res, err := r.rocketMQProducer.SendSync(context.Background(), rmsg)
	if err != nil {
		log.Warn("send message error: %s\n", zap.Error(err))
		return err
	} else {
		log.Info("send message success: result=%s\n", zap.ByteString(res.MsgID, []byte(res.String())))
	}
	//}

	return nil
}
