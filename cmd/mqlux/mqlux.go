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
	paho "github.com/eclipse/paho.mqtt.golang"
	"github.com/ktt-ol/mqlux/internal/config"
	"github.com/ktt-ol/mqlux/internal/debug"
	"github.com/ktt-ol/mqlux/internal/handler/csv"
	"github.com/ktt-ol/mqlux/internal/handler/keepalive"
	"github.com/ktt-ol/mqlux/internal/handler/topic"
	"github.com/ktt-ol/mqlux/internal/influxdb"
	"github.com/ktt-ol/mqlux/internal/mqlux"
	"github.com/ktt-ol/mqlux/internal/mqtt"
	"github.com/ktt-ol/mqlux/internal/parser"
)

func main() {
	colog.Register()
	colog.ParseFields(true)
	colog.SetMinLevel(colog.LInfo)

	configFile := flag.String("config", "mqlux.tml", "configuration")
	csvFile := flag.String("messages-csv", "", "read messages from CSV file; disables InfluxDB output")
	isDebug := flag.Bool("debug", false, "print debug messages")
	flag.Parse()

	if *isDebug {
		colog.SetMinLevel(colog.LDebug)
		// mqtt debug is very verbose
		// mqtt.DEBUG = log.New(os.Stdout, "[mqtt] ", log.LstdFlags)
	}

	config := config.Config{}
	_, err := toml.DecodeFile(*configFile, &config)
	if err != nil {
		log.Fatal(err)
	}

	var writer influxdb.Writer
	if config.InfluxDB.URL != "" && *csvFile == "" {
		db, err := influxdb.NewInfluxDBClient(config)
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

	handlers := []mqtt.Handler{}

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
		logger, err := csv.NewMQTTLogger(out)
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
		watchdog := keepalive.NewWatchdogHandler(keepAlive)
		defer watchdog.Stop()
		handlers = append(handlers, watchdog)
	}

	if config.Messages.Devices.Topic != "" {
		handler, err := topic.New(
			config.Messages.Devices.Topic,
			"", nil, //set by NetDeviceParser
			parser.NetDeviceParser(config.Messages.Devices),
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
			parser.SpaceStatusParser(config.Messages.SpaceStatus),
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
			parser.FloatParser,
			writer,
		)
		if err != nil {
			log.Fatal(err)
		}
		handlers = append(handlers, handler)
	}

	if *csvFile != "" {
		fwd := mqtt.Prepare(handlers)
		err = debug.MessagesFromCSV(*csvFile, fwd)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	log.Printf("debug: connecting to subscribe for %d handlers", len(handlers))
	_, err = mqtt.NewMQTTClient(config, func(c paho.Client) {
		log.Print("debug: on connect")
		mqtt.Subscribe(c, handlers)
	})
	if err != nil {
		log.Fatal(err)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	s := <-sigs
	log.Print("debug: exiting: ", s)
}
