package model

type SearchData struct {
	PodName     string `json:"podName"`
	Key         string `json:"key"`
	Topic       string `json:"topic"`
	Value       string `json:"value"`
	MetadataKey string `json:"metadataKey"`
}
