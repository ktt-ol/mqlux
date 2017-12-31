package mqlux

import "time"

// A Message stores data from incomming messages (typical MQTT messages).
type Message struct {
	Time     time.Time
	Topic    string
	Payload  []byte
	Retained bool
}

// A Record stores data for outgoing records (typical InfluxDB records).
type Record struct {
	Measurement string
	Tags        map[string]string
	Value       interface{}
}

// Parser converts one Message into zero or more Records.
// A parser can set different Measurement and Tags for each Record (e.g. based
// on the message or the parser configuration).
type Parser func(msg Message, measurement string, tags map[string]string) ([]Record, error)

// Writer writes all records.
type Writer func([]Record) error
