package handler

import (
	"bee-search/pkg/client"
	"bee-search/pkg/config"
	"bee-search/pkg/logger"
	"bee-search/pkg/model"
	"bee-search/pkg/model/status"
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"sync"
	"time"
)

type ConsumeHandler struct {
	c     client.KafkaConsumer
	redis *client.RedisService
	data  *model.SearchData
	conf  *config.ApplicationConfig
}

func NewConsumeHandler(consumer client.KafkaConsumer, redis *client.RedisService, appConf *config.ApplicationConfig, sd *model.SearchData) *ConsumeHandler {
	return &ConsumeHandler{c: consumer, redis: redis, conf: appConf, data: sd}
}

func (ch *ConsumeHandler) Handle() {
	s, err := ch.redis.GetStateOfSearch(ch.data.MetadataKey)
	if err != nil {
		logger.Logger().Error("error occurred when HMGET from Redis", zap.Error(err))
		ch.stopProcess(err)
		return
	}

	if s != status.INQUEUE {
		logger.Logger().Error("There is already process.")
		ch.stopProcess(nil)
		return
	}

	ch.redis.SearchStarted(ch.data.MetadataKey)
	executeTime := time.Now().UnixMilli()
	ch.Consume()
	executeTime = time.Now().UnixMilli() - executeTime
	ch.redis.SearchFinished(ch.data.MetadataKey, executeTime)
	ch.stopProcess(nil)
}

func (ch *ConsumeHandler) Consume() {
	pIds := ch.getPartitionIds(ch.data)
	if len(pIds) == 0 {
		return
	}

	var wg sync.WaitGroup
	wg.Add(len(pIds))

	for i := range pIds {
		pId := pIds[i]
		go func(sd *model.SearchData, p int, r *client.RedisService) {
			h := NewListenerHandler(sd, ch.c, p, r)
			h.Handle()
			wg.Done()
		}(ch.data, pId, ch.redis)
	}

	wg.Wait()
}

func (ch *ConsumeHandler) getPartitionIds(searchData *model.SearchData) []int {
	partitions, err := ch.c.GetPartitions(searchData.Topic)
	if err != nil {
		return nil
	}

	pIds := make([]int, len(partitions))
	for i := range partitions {
		pIds[i] = partitions[i].ID
	}

	return pIds
}

func (ch *ConsumeHandler) stopProcess(err error) {
	if err != nil {
		ch.redis.SearchException(ch.data.Key, err)
		logger.Logger().Error("process can not be complete", zap.Error(err))
	}

	ch.deleteRequest(ch.data.PodName)
}

func (ch *ConsumeHandler) deleteRequest(podName string) {
	fmt.Println("waiting before kill ...")
	cli := &http.Client{}
	url := ch.conf.Master.Url + "/master?key=" + podName
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		logger.Logger().Error("Request couldn't created!", zap.Error(err))
		return
	}

	_, err = cli.Do(req)
	if err != nil {
		logger.Logger().Error("Request couldn't send to pod-master!", zap.Error(err))
		return
	}
}
