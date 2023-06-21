package handler

import (
	"bee-search/pkg/client"
	"bee-search/pkg/model"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

type ListenerHandler struct {
	searchData  *model.SearchData
	c           client.KafkaConsumer
	partitionId int
	r           *client.RedisService
}

func NewListenerHandler(sd *model.SearchData, consumer client.KafkaConsumer, partitionId int, r *client.RedisService) *ListenerHandler {
	return &ListenerHandler{searchData: sd, c: consumer, partitionId: partitionId, r: r}
}

func (l *ListenerHandler) Handle() {
	sd := time.Now().UTC().UnixMilli()
	kc := l.c.CreateConsumer(l.searchData.Topic, l.partitionId)
	defer kc.Close()
	ctx, cancel := context.WithCancel(context.Background())
	t := time.AfterFunc(time.Second*10, func() {
		cancel()
	})

	for {
		m, err := kc.ReadMessage(ctx)
		if err != nil || m.Time.UnixMilli() > sd {
			break
		}

		t.Reset(time.Second * 5)
		if m.Key != nil && len(m.Key) > 0 {
			kStr := string(m.Key)
			has := strings.Contains(kStr, l.searchData.Value)
			if has {
				l.founded(&m)
				if l.isReachLimit() {
					break
				}
				continue
			}
		}

		if m.Value != nil && len(m.Value) > 0 {
			vStr := string(m.Value)
			has := strings.Contains(vStr, l.searchData.Value)
			if has {
				l.founded(&m)
				if l.isReachLimit() {
					break
				}
				continue
			}
		}
	}
}

func (l *ListenerHandler) isReachLimit() bool {
	length, err := l.r.Length(l.searchData.Key)
	if err == nil && length >= 500 {
		err = errors.New("searching stopped. you hit more than 500 messages. please use a specific value")
		l.r.SearchException(l.searchData.Key, err)
		return true
	}

	return false
}

func (l *ListenerHandler) founded(m *kafka.Message) {
	sec := map[string]interface{}{}
	if err := json.Unmarshal(m.Value, &sec); err != nil {
		l.r.SearchException(l.searchData.Key, err)
		return
	}
	sec["__kafka_offset"] = m.Offset
	sec["__kafka_partition"] = m.Partition
	sec["__kafka_publish_date_utc"] = m.Time.UnixMilli()
	if m.Key != nil && len(m.Key) > 0 {
		sec["__kafka_key"] = string(m.Key)
	}
	value, err := json.Marshal(sec)
	if err != nil {
		l.r.SearchException(l.searchData.Key, err)
		return
	}
	l.r.Save(l.searchData.Key, string(value))
}
