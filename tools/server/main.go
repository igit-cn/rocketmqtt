package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/apache/rocketmq-client-go/v2"
	"github.com/apache/rocketmq-client-go/v2/consumer"
	"github.com/apache/rocketmq-client-go/v2/primitive"
	"github.com/apache/rocketmq-client-go/v2/producer"
	"go.uber.org/zap"
)

func main() {
	namesrv := flag.String("ns", "10.2.55.20:9876", "rocketmq nameserver :10.2.55.20:9876;10.2.55.21:9876")
	subTopic := flag.String("st", "mqtt2rmq", "subscribe rocketmq topic name")
	pubTopic := flag.String("pt", "rmq2mqtt", "publish rocketmq topic name")
	clientNum := flag.Int("c", 1000, "client topic num, test-mqtt:xxxx")
	//thread := flag.Int("tr", 10, "publish thread")
	interval := flag.Int("i", 1000, "publish interval (ms),every thread")
	flag.Parse()
	ns, err := primitive.NewNamesrvAddr(*namesrv)
	if err != nil {
		log.Fatal("name server error: ", zap.Error(err))
	}
	c1, err := rocketmq.NewPushConsumer(
		consumer.WithGroupName("test_server"),
		consumer.WithNameServer(ns),
	)
	if err != nil {
		log.Fatalf("new push consumer error: %s", err.Error())
	}
	p, _ := rocketmq.NewProducer(
		producer.WithNameServer(ns),
		producer.WithRetry(2),
	)
	err = p.Start()
	if err != nil {
		log.Fatalf("start producer error: %s", err.Error())
	}
	subscribeRmq(c1, *subTopic, []string{})
	publishRmq(0, int(*clientNum), *pubTopic, p, int(*interval))
}

func subscribeRmq(rc rocketmq.PushConsumer, topic string, tags []string) {
	selector := consumer.MessageSelector{}
	if len(tags) > 0 {
		e := ""
		for t := range tags {
			if t == 0 {
				e = tags[t]
			} else {
				e = e + " || " + tags[t]
			}
		}
		selector = consumer.MessageSelector{
			Type:       consumer.TAG,
			Expression: e,
		}
	}
	err := rc.Subscribe(topic, selector, func(ctx context.Context,
		msgs ...*primitive.MessageExt) (consumer.ConsumeResult, error) {
		for i := range msgs {
			m := make(map[string]interface{})
			json.Unmarshal(msgs[i].Body, &m)
			if m["ts"] == nil {
				log.Printf("%#v", m)
				continue
			}
			sts := m["ts"].(float64)
			rts := time.Now().UnixNano()
			log.Printf("[recived rmq] delay: %d ms, topic: %s, body: %v \n", (rts-int64(sts))/1000000, msgs[i].Topic, string(msgs[i].Body))
		}
		return consumer.ConsumeSuccess, nil
	})
	if err != nil {
		log.Fatalf("start producer error: %s", err.Error())
	}
	err = rc.Start()
	if err != nil {
		log.Fatalf("start producer error: %s", err.Error())
		//os.Exit(-1)
	}
}

// publish rmq loop
func publishRmq(mqtt_topic_start int, clientNum int, topic string, rp rocketmq.Producer, interval int) {
	for {
		i := 0
		for {
			i++
			payload := fmt.Sprintf(`{"taskId":%d, "msgId":%d, "ts": %d,"did":"865532048803320","rslt":1,"sts":4,"cmd":101,"ext":0,"ts":1570615080000,"tsOut":1570615080000,"seq":241821,"sig":"40BD59039AC50DF75751820CA8D90C4D"}`, mqtt_topic_start, i, time.Now().UnixNano())
			rmsg := primitive.NewMessage(topic,
				[]byte(payload))
			rmsg.WithProperty("topic", fmt.Sprintf("test-mqtt-down-%d", mqtt_topic_start+i))
			rmsg.WithProperty("clientId", fmt.Sprintf("cmd-%d", mqtt_topic_start+i))
			rmsg.WithTag("hmq-10000-0")

			// sync
			res, err := rp.SendSync(context.Background(), rmsg)
			if err != nil {
				log.Printf("[sent rmq] error: %s\n", err.Error())
			} else {
				log.Printf("[sent rmq]: Topic: %s  msgId: %s payload:%s \n", topic, res.MsgID, payload)
			}
			time.Sleep(time.Duration(interval) * time.Millisecond)
			if i >= clientNum {
				break
			}
		}
	}
}
