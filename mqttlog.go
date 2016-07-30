package mqlux

import (
	"encoding/csv"
	"io"
	"log"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
)

type record struct {
	time    time.Time
	topic   string
	payload []byte
}

func NewMQTTLogger(out io.Writer) (*MQTTLogger, error) {
	logger := &MQTTLogger{
		csvWriter: csv.NewWriter(out),
		records:   make(chan record, 64),
	}
	go logger.run()
	return logger, nil
}

type MQTTLogger struct {
	csvWriter *csv.Writer
	records   chan record
}

func (w *MQTTLogger) Log(client mqtt.Client, message mqtt.Message) {
	w.records <- record{
		time:    time.Now(),
		topic:   message.Topic(),
		payload: message.Payload(),
	}
}

func (w *MQTTLogger) Stop() {
	close(w.records)
}

func (w *MQTTLogger) run() {
	for r := range w.records {
		err := w.csvWriter.Write([]string{
			r.time.Format(time.RFC3339),
			r.topic,
			string(r.payload),
		})

		if err != nil {
			log.Println("error: unable to write CSV", err)
		}
		w.csvWriter.Flush()
	}
}
