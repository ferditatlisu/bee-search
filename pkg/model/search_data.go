package model

import (
	"os"
	"strconv"
	"time"
)

type SearchData struct {
	PodName     string `json:"podName"`
	Key         string `json:"key"`
	Topic       string `json:"topic"`
	Value       string `json:"value"`
	MetadataKey string `json:"metadataKey"`
	StartDate   int64  `json:"startDate"`
	EndDate     int64  `json:"endDate"`
	ValueType   uint8  `json:"valueType"`
}

func NewSearchData() *SearchData {
	startDate := getStartDate()
	endDate := getEndDate()
	valueType := getValueType()

	return &SearchData{
		PodName:     os.Getenv("POD_NAME"),
		Key:         os.Getenv("KEY"),
		Topic:       os.Getenv("TOPIC"),
		Value:       os.Getenv("VALUE"),
		MetadataKey: os.Getenv("METADATA_KEY"),
		StartDate:   startDate,
		EndDate:     endDate,
		ValueType:   valueType,
	}
}

func getStartDate() int64 {
	value := time.Now().Add(time.Hour * -720).UnixMilli()
	sd := os.Getenv("START_DATE")
	if sd != "" {
		value, _ = strconv.ParseInt(sd, 10, 64)
	}

	return value
}

func getEndDate() int64 {
	value := time.Now().UnixMilli()
	sd := os.Getenv("END_DATE")
	if sd != "" {
		value, _ = strconv.ParseInt(sd, 10, 64)
	}

	return value
}

func getValueType() uint8 {
	value := int64(1)
	vt := os.Getenv("VALUE_TYPE")
	if vt != "" {
		value, _ = strconv.ParseInt(vt, 10, 64)
	}

	return uint8(value)
}
