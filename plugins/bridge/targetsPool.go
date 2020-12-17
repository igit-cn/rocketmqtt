package bridge

import (
	"sync"
)

type targetMemPool struct {
	clientsMap  map[string][]string
	topicBitMap map[string]int64
	sync.RWMutex
}

func (t *targetMemPool) deleteClient(clientId string) {
	t.Lock()
	topics, ok := t.clientsMap[clientId]
	if ok {
		for _, ts := range topics {
			delete(t.topicBitMap, ts)
		}
	}
	t.Unlock()
}
func (t *targetMemPool) storeClientTopicMatch(clientId string, topic string, bitmap int64) {
	t.Lock()
	topics, ok := t.clientsMap[clientId]
	if ok {
		if _, ok := t.topicBitMap[topic]; ok {
			// if more than 10 topics,delete first topic
			if len(topics) >= 10 {
				t.clientsMap[clientId] = append(topics[1:9], topic)
			} else {
				t.clientsMap[clientId] = append(topics, topic)
			}
		}
	} else {
		t.clientsMap[clientId] = []string{topic}
	}
	t.topicBitMap[topic] = bitmap
	t.Unlock()
}
func (t *targetMemPool) getMatch(topic string) (bitmap int64, ok bool) {
	t.RLock()
	bitmap, ok = t.topicBitMap[topic]
	t.RUnlock()
	return bitmap, ok
}
