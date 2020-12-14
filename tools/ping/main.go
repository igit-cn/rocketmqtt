package main

import (
"encoding/json"
"flag"
"fmt"
"log"
"os"
"strconv"
"sync"
"time"
//导入mqtt包
MQTT "github.com/eclipse/paho.mqtt.golang"
)

var pingMap sync.Map

var fcallback MQTT.MessageHandler = func(client MQTT.Client, msg MQTT.Message) {
	rtc := time.Now().UnixNano() / 1000000
	m := make(map[string]interface{})
	_ = json.Unmarshal(msg.Payload(), &m)
	id := m["pong"].(string)
	sentTs, ok := pingMap.LoadAndDelete(id)
	if !ok {
		log.Printf("id: %s not exist", id)
	}
	brokerStc := int64(m["ts"].(float64))
	clientStc := sentTs.(int64)
	delay := rtc - clientStc
	warn := ""
	switch true {
	case delay > 1000:
		warn = " <-----"
	case delay > 100:
		warn = " <----"
	case delay > 10:
		warn = " <---"
	case delay > 5:
		warn = " <--"
	case delay > 1:
		warn = " <-"
	default:
	}
	log.Printf("[<-] ts: %d Pong: %s delay: %d brokerts: %d  %s \n", rtc, id, delay, brokerStc, warn)
}

func main() {
	//生成连接的客户端数
	server := flag.String("h", "tcp://127.0.0.1:1883", "mqtt server")
	interval := flag.Uint64("i", 1000, "create client interval, default 1000(ms)")
	flag.Parse()
	opts := MQTT.NewClientOptions().AddBroker(*server) //.SetUsername("test").SetPassword("test")
	i := int(*interval)

	log.Printf("start ping %s with interval %d", *server, i)

	opts.SetClientID("pingClient")
	opts.SetDefaultPublishHandler(fcallback)
	opts.SetConnectTimeout(time.Duration(10) * time.Second)

	//创建连接
	c := MQTT.NewClient(opts)
	c.Subscribe("ping", 1, fcallback)

	if token := c.Connect(); token.WaitTimeout(time.Duration(10)*time.Second) && token.Wait() && token.Error() != nil {
		fmt.Printf("connet mqtt error")
		os.Exit(1)
	}
	x := 0
	for {
		s := strconv.Itoa(x)
		token := c.Publish("ping", 0, false, s)
		token.Wait()
		t := time.Now().UnixNano() / 1000000
		log.Printf("[->] ts: %d Ping: %s", t, s)
		pingMap.Store(s, t)
		time.Sleep(time.Duration(i) * time.Millisecond)
		x++
	}
}

