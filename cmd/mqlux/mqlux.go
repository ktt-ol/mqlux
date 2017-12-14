package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/comail/colog"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/ktt-ol/mqlux"
	"github.com/ktt-ol/mqlux/topic"
)

func main() {
	colog.Register()
	colog.ParseFields(true)
	colog.SetMinLevel(colog.LInfo)

	configFile := flag.String("config", "mqlux.tml", "configuration")
	csvFile := flag.String("messages-csv", "", "read messages from CSV file; disables InfluxDB output")
	debug := flag.Bool("debug", false, "print debug messages")
	flag.Parse()

	if *debug {
		colog.SetMinLevel(colog.LDebug)
		// mqtt debug is very verbose
		// mqtt.DEBUG = log.New(os.Stdout, "[mqtt] ", log.LstdFlags)
	}

	config := mqlux.Config{}
	_, err := toml.DecodeFile(*configFile, &config)
	if err != nil {
		log.Fatal(err)
	}

	var writer mqlux.Writer
	if config.InfluxDB.URL != "" && *csvFile == "" {
		db, err := mqlux.NewInfluxDBClient(config)
		if err != nil {
			log.Fatal(err)
		}
		writer = db.Write
	} else {
		writer = func(recs []mqlux.Record) error {
			var buf bytes.Buffer
			for _, rec := range recs {
				buf.Reset()
				buf.WriteString("measurement ")
				buf.WriteString(rec.Measurement)
				fmt.Fprintf(&buf, " -> %v ", rec.Value)
				for k, v := range rec.Tags {
					buf.WriteString(k)
					buf.WriteString("='")
					buf.WriteString(v)
					buf.WriteString("' ")
				}
				log.Println(buf.String())
			}
			return nil
		}
	}

	handlers := []mqlux.Handler{}

	if config.MQTT.CSVLog != "" && *csvFile == "" {
		var out io.Writer
		if config.MQTT.CSVLog == "-" {
			out = os.Stdout
		} else {
			f, err := os.OpenFile(config.MQTT.CSVLog, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()
			out = f
		}
		logger, err := mqlux.NewMQTTLogger(out)
		if err != nil {
			log.Fatal(err)
		}
		defer logger.Stop()
		handlers = append(handlers, logger)
	}

	if config.MQTT.KeepAlive != "" {
		keepAlive, err := time.ParseDuration(config.MQTT.KeepAlive)
		if err != nil {
			log.Fatal("invalid keepalive duration", err)
		}
		watchdog := mqlux.NewWatchdogHandler(keepAlive)
		defer watchdog.Stop()
		handlers = append(handlers, watchdog)
	}

	if config.Messages.Devices.Topic != "" {
		handler, err := topic.New(
			config.Messages.Devices.Topic,
			"", nil, //set by NetDeviceParser
			mqlux.NetDeviceParser(config.Messages.Devices),
			writer,
		)
		if err != nil {
			log.Fatal(err)
		}
		handlers = append(handlers, handler)
	}

	if config.Messages.SpaceStatus.Topic != "" {
		handler, err := topic.New(
			config.Messages.SpaceStatus.Topic,
			"", nil, //set by SpaceStatusParser
			mqlux.SpaceStatusParser(config.Messages.SpaceStatus),
			writer,
		)
		if err != nil {
			log.Fatal(err)
		}
		handlers = append(handlers, handler)
	}

	for _, sensor := range config.Messages.Sensors {
		handler, err := topic.New(
			sensor.Topic,
			sensor.Measurement,
			sensor.Tags,
			mqlux.FloatParser,
			writer,
		)
		if err != nil {
			log.Fatal(err)
		}
		handlers = append(handlers, handler)
	}

	if *csvFile != "" {
		fwd := mqlux.Prepare(handlers)
		err = mqlux.MessagesFromCSV(*csvFile, fwd)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	log.Printf("debug: connecting to subscribe for %d handlers", len(handlers))
	_, err = mqlux.NewMQTTClient(config, func(c mqtt.Client) {
		log.Print("debug: on connect")
		mqlux.Subscribe(c, handlers)
	})
	if err != nil {
		log.Fatal(err)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	s := <-sigs
	log.Print("debug: exiting: ", s)
}
