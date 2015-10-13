package mqlux

import (
	"encoding/csv"
	"log"
	"os"
	"time"

	"git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
)

type record struct {
	time    time.Time
	topic   string
	payload []byte
}

func NewMQTTLogger(file string) (*MQTTLogger, error) {
	f, err := os.OpenFile(file, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, err
	}
	logger := &MQTTLogger{
		file:      f,
		csvWriter: csv.NewWriter(f),
		records:   make(chan record, 64),
	}
	go logger.run()
	return logger, nil
}

type MQTTLogger struct {
	file      *os.File
	csvWriter *csv.Writer
	records   chan record
}

func (w *MQTTLogger) Log(client *mqtt.Client, message mqtt.Message) {
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
	defer w.file.Close()
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
