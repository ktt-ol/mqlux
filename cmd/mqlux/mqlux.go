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

	"github.com/ktt-ol/mqlux/internal/router"

	"github.com/BurntSushi/toml"
	"github.com/comail/colog"
	"github.com/ktt-ol/mqlux/internal/config"
	"github.com/ktt-ol/mqlux/internal/debug"
	"github.com/ktt-ol/mqlux/internal/handler/csv"
	"github.com/ktt-ol/mqlux/internal/handler/keepalive"
	"github.com/ktt-ol/mqlux/internal/handler/topic"
	"github.com/ktt-ol/mqlux/internal/influxdb"
	"github.com/ktt-ol/mqlux/internal/mqlux"
	"github.com/ktt-ol/mqlux/internal/mqtt"
	"github.com/ktt-ol/mqlux/internal/parser"
	"github.com/ktt-ol/mqlux/internal/parser/script"
)

var version = "master"

func main() {
	colog.Register()
	colog.ParseFields(true)
	colog.SetMinLevel(colog.LInfo)

	configFile := flag.String("config", "mqlux.tml", "configuration")
	csvFile := flag.String("messages-csv", "", "read messages from CSV file; disables InfluxDB output")
	isDebug := flag.Bool("debug", false, "print debug messages")
	printVersion := flag.Bool("version", false, "print version and exit")

	flag.Parse()

	if *printVersion {
		fmt.Printf("mqlux %s\n", version)
		os.Exit(0)
	}

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

	var writer mqlux.Writer
	if config.InfluxDB.URL != "" && *csvFile == "" {
		db, err := influxdb.NewInfluxDBClient(config)
		if err != nil {
			log.Fatal(err)
		}
		writer = db.Write
	} else {
		writer = func(recs []mqlux.Record) error { return nil }
	}

	if *isDebug {
		// wrap original writer with debug logger
		origWriter := writer
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
			return origWriter(recs)
		}
	}

	r := router.New()

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
		r.Add("/#", logger)
	}

	if config.MQTT.KeepAlive != "" && *csvFile == "" {
		keepAlive, err := time.ParseDuration(config.MQTT.KeepAlive)
		if err != nil {
			log.Fatal("invalid keepalive duration", err)
		}
		watchdog := keepalive.NewWatchdogHandler(keepAlive)
		defer watchdog.Stop()
		r.Add("/#", watchdog)
	}

	for _, sub := range config.Subscriptions {
		var p mqlux.Parser
		if sub.Script != "" {
			vm, err := script.New(sub.Script)
			if err != nil {
				log.Fatal(err)
			}
			p = vm.Parse
		} else {
			p = parser.FloatParser
		}

		handler, err := topic.New(
			sub.Topic,
			sub.Measurement,
			sub.Tags,
			p,
			writer,
		)
		handler.IncludeRetained(sub.IncludeRetained)

		if err != nil {
			log.Fatal(err)
		}
		r.Add(handler.Topic(), handler)
	}

	if *csvFile != "" {
		err = debug.MessagesFromCSV(*csvFile, r.Receive)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	log.Printf("debug: connecting to subscribe")
	_, err = mqtt.Subscribe(config.MQTT, r.Receive)
	if err != nil {
		log.Fatal(err)
	}

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	s := <-sigs
	log.Print("debug: exiting: ", s)
}
