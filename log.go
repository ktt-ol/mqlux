package mqlux

import (
	"encoding/csv"
	"io"
	"log"
	"time"
)

func NewMQTTLogger(out io.Writer) (*MQTTLogger, error) {
	logger := &MQTTLogger{
		csvWriter: csv.NewWriter(out),
		records:   make(chan Message, 64),
	}
	go logger.run()
	return logger, nil
}

type MQTTLogger struct {
	csvWriter *csv.Writer
	records   chan Message
}

func (w *MQTTLogger) Receive(msg Message) {
	w.records <- msg
}

func (w *MQTTLogger) Topic() string {
	return "/#"
}

func (w *MQTTLogger) Match(topic string) bool {
	return true
}

func (w *MQTTLogger) Stop() {
	close(w.records)
}

func (w *MQTTLogger) run() {
	for r := range w.records {
		err := w.csvWriter.Write([]string{
			r.Time.Format(time.RFC3339),
			r.Topic,
			string(r.Payload),
		})

		if err != nil {
			log.Println("error: unable to write CSV", err)
		}
		w.csvWriter.Flush()
	}
}
