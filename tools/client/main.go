package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"sync"
	"time"
	//导入mqtt包
	MQTT "github.com/eclipse/paho.mqtt.golang"
)

var fcallback MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
	rtc := time.Now().UnixNano()
	m := make(map[string]interface{})
	_ = json.Unmarshal(msg.Payload(), &m)
	stc := int64(m["ts"].(float64))

	log.Printf("[Recived mqtt] delay %d ms TOPIC: %s MSG: %s \n", (rtc-stc)/1000000, msg.Topic(), msg.Payload())
}
var fail_nums int = 0

func main() {
	//生成连接的客户端数
	c := flag.Uint64("n", 1000, "client nums,default 1000")
	s := flag.Uint64("s", 0, "client start id,default 0")
	snum := flag.Uint64("t", 1, "subscribe topic numbers")
	host := flag.String("h", "tcp://127.0.0.1:1883", "mqtt server :tcp://172.16.84.15:1883")
	interval := flag.Uint64("i", 1000, "create client interval, default 1000(ms)")
	onlySubscribe := flag.Bool("o", false, "only subscribe topic, release after keepalive time")
	keepalive := flag.Uint64("k", 60, "keepalive time,default 60(s)")
	flag.Parse()
	nums := int(*c)
	begin := int(*s)
	keep := int(*keepalive)
	subscribeNum := int(*snum)
	fmt.Printf("create mqtt %d clients with %s\n", nums, *host)

	wg := sync.WaitGroup{}
	for i := 0; i <= nums; i++ {
		wg.Add(1)
		time.Sleep(time.Duration(*interval) * time.Millisecond)
		go createTask(keep, *host, begin+i, &wg, *onlySubscribe, subscribeNum)
	}
	wg.Wait()
}

func createTask(keepalive int, server string, taskId int, wg *sync.WaitGroup, onlySubscribe bool, subscribeNum int) {
	defer wg.Done()
	if taskId%100 == 0 {
		fmt.Printf("%d", taskId)
	} else {
		fmt.Printf(".")
	}
	opts := MQTT.NewClientOptions().AddBroker(server) //.SetUsername("test").SetPassword("test")

	opts.SetClientID(fmt.Sprintf("tc-%d", taskId))
	opts.SetDefaultPublishHandler(fcallback)
	opts.SetConnectTimeout(time.Duration(60) * time.Second)

	//创建连接
	c := MQTT.NewClient(opts)

	if token := c.Connect(); token.WaitTimeout(time.Duration(60)*time.Second) && token.Wait() && token.Error() != nil {
		fail_nums++
		fmt.Printf("taskId:%d,fail_nums:%d,error:%s \n", taskId, fail_nums, token.Error())
	}

	topicDownstream := fmt.Sprintf("test-mqtt-down:%d", taskId)

	c.Subscribe(topicDownstream, 1, fcallback)
	for i := 1; i < subscribeNum; i++ {
		c.Subscribe(fmt.Sprintf("%s-%d", topicDownstream, i), 1, fcallback)
	}
	if onlySubscribe {
		time.Sleep(time.Duration(keepalive) * time.Second)
		log.Printf("Task %d over", taskId)
		//c.Unsubscribe(topicDownstream)
		c.Disconnect(5000)
		return
	}
	//每隔25秒向topic发送一条消息 模型 10000车每秒400条，单车25秒1条
	i := 0
	for {
		i++
		text := fmt.Sprintf(`{
        "taskId": %d,
        "msgId": %d,
		}`, taskId, i)
		topicUpstream := fmt.Sprintf("test-mqtt-up:%d", taskId)
		token := c.Publish(topicUpstream, 1, false, text)
		token.Wait()
		log.Printf("[sent mqtt] Topic: %s Msg: %s", topicUpstream, text)
		time.Sleep(time.Duration(25) * time.Second)
	}
}
