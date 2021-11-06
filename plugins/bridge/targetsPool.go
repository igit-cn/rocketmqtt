package bridge

import (
	"sync"
)

type topicMatched struct {
	//ts     time.Time
	topics []string
}

type targetMemPool struct {
	clientsMap  map[string]*topicMatched
	topicBitMap map[string]int64
	sync.RWMutex
}

func (t *targetMemPool) deleteTopicMatch(clientId string) {
	t.Lock()
	tm, ok := t.clientsMap[clientId]
	if ok {
		for _, ts := range tm.topics {
			delete(t.topicBitMap, ts)
		}
	}
	t.Unlock()
}
func (t *targetMemPool) storeClientTopicMatch(clientId string, topic string, bitmap int64) {
	t.Lock()
	tm, ok := t.clientsMap[clientId]
	if ok {
		if _, ok := t.topicBitMap[topic]; ok {
			newTopics := append(tm.topics[1:9], topic)
			t.clientsMap[clientId] = &topicMatched{
				//ts:     time.Now(),
				topics: newTopics,
			}
		}
	} else {
		t.clientsMap[clientId] = &topicMatched{
			//ts:     time.Now(),
			topics: []string{topic},
		}
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
