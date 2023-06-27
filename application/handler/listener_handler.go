package handler

import (
	"bee-search/pkg/client"
	"bee-search/pkg/logger"
	"bee-search/pkg/model"
	"bee-search/pkg/model/valuetype"
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
	ctx := context.Background()
	kc := l.c.CreateConsumer(l.searchData.Topic, l.partitionId)
	defer kc.Close()
	_ = kc.SetOffsetAt(ctx, time.UnixMilli(l.searchData.StartDate))
	lag, _ := kc.ReadLag(ctx)
	endDate := l.searchData.EndDate
	checkKey := l.searchData.ValueType>>valuetype.KEY&1 == 1
	checkValue := l.searchData.ValueType>>valuetype.VALUE&1 == 1
	checkHeader := l.searchData.ValueType>>valuetype.HEADER&1 == 1

	for {
		if lag <= 0 {
			break
		}
		m, err := kc.ReadMessage(ctx)
		if err != nil || m.Time.UnixMilli() > endDate {
			logger.Logger().Info(err.Error())
		}
		lag -= 1

		if checkKey && m.Key != nil && len(m.Key) > 0 {
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

		if checkValue && m.Value != nil && len(m.Value) > 0 {
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

		if checkHeader && m.Headers != nil && len(m.Headers) > 0 {
			headerFounded := false
			for i := range m.Headers {
				h := m.Headers[i]
				hasKey := strings.Contains(h.Key, l.searchData.Value)
				hasValue := strings.Contains(string(h.Value), l.searchData.Value)
				if hasKey || hasValue {
					l.founded(&m)
					headerFounded = true
					break
				}
			}

			if headerFounded && l.isReachLimit() {
				break
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
	value := map[string]interface{}{}
	if err := json.Unmarshal(m.Value, &value); err != nil {
		l.r.SearchException(l.searchData.Key, err)
		return
	}

	data := map[string]interface{}{}
	data["value"] = value

	data["offset"] = m.Offset
	data["partition"] = m.Partition
	data["publish_date_utc"] = m.Time.UnixMilli()
	if m.Key != nil && len(m.Key) > 0 {
		data["key"] = string(m.Key)
	}

	if m.Headers != nil && len(m.Headers) > 0 {
		var headers []map[string]string
		for i := range m.Headers {
			header := m.Headers[i]
			hm := map[string]string{}
			hm[header.Key] = string(header.Value)
			headers = append(headers, hm)
		}

		data["headers"] = headers
	}

	d, err := json.Marshal(data)
	if err != nil {
		l.r.SearchException(l.searchData.Key, err)
		return
	}
	l.r.Save(l.searchData.Key, string(d))
}
