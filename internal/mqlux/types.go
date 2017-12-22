package mqlux

import "time"

type Message struct {
	Time    time.Time
	Topic   string
	Payload []byte
}

type Record struct {
	Measurement string
	Tags        map[string]string
	Value       interface{}
}
