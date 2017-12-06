package main

import (
	"flag"
	"log"
	"os"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/comail/colog"
	"github.com/eclipse/paho.mqtt.golang"
	"github.com/ktt-ol/mqlux"
	"github.com/ktt-ol/mqlux/topic"
)

func main() {
	colog.Register()
	colog.ParseFields(true)
	colog.SetMinLevel(colog.LInfo)

	configFile := flag.String("config", "mqlux.tml", "configuration")
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

	var onConnectHandler []mqtt.OnConnectHandler

	var writer mqlux.Writer
	if config.InfluxDB.URL != "" {
		db, err := mqlux.NewInfluxDBClient(config)
		if err != nil {
			log.Fatal(err)
		}
		writer = db.Write
	} else {
		writer = func(recs []mqlux.Record) error {
			for _, rec := range recs {
				log.Println(rec)
			}
			return nil
		}
	}

	handlers := []mqlux.Handler{}

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

	onConnectHandler = append(onConnectHandler, func(c mqtt.Client) {
		log.Print("debug: on connect")

		mqlux.Subscribe(c, handlers)

		// for _, sensor := range config.Messages.Sensors {
		// 	rt, err := .NewRegexpTopic(sensor.Topic)
		// 	if err == nil {
		// 		sensor.RegexpTopic = rt
		// 		sensor.Topic = rt.SubscribeTopic
		// 	} else if err == mqlux.NoRegexpTopic {
		// 		// just use regular sensor.Topic
		// 	} else {
		// 		log.Fatal(err)
		// 	}
		// 	log.Print("debug: subscribing for sensor", sensor)
		// 	if err := mqlux.Subscribe(c, sensor.Topic,
		// 		mqlux.SensorHandler(config, sensor, db.WriteSensor)); err != nil {
		// 		log.Fatal(err)
		// 	} else {
		// 		log.Print("debug: subscribed to", sensor.Topic)
		// 	}
		// }
	})

	var keepAliveC chan struct{}
	var keepAlive time.Duration
	// var keepAliveHandler mqtt.MessageHandler
	// if config.MQTT.KeepAlive != "" {
	// 	var err error
	// 	keepAlive, err = time.ParseDuration(config.MQTT.KeepAlive)
	// 	if err != nil {
	// 		log.Fatal("invalid keepalive duration", err)
	// 	}
	// 	keepAliveC = make(chan struct{}, 1)
	// 	keepAliveHandler = func(client mqtt.Client, message mqtt.Message) {
	// 		keepAliveC <- struct{}{}
	// 	}
	// }

	// var logHandler mqtt.MessageHandler
	// if config.MQTT.CSVLog != "" {
	// 	var out io.Writer
	// 	if config.MQTT.CSVLog == "-" {
	// 		out = os.Stdout
	// 	} else {
	// 		f, err := os.OpenFile(config.MQTT.CSVLog, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0644)
	// 		if err != nil {
	// 			log.Fatal(err)
	// 		}
	// 		defer f.Close()
	// 		out = f
	// 	}
	// 	logger, err := mqlux.NewMQTTLogger(out)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	logHandler = logger.Log
	// }

	//
	// // mqtt.Subscribe only supports one callback for each topic
	// // create allHandler and dispatch to log and keepalive handlers
	// if logHandler != nil || keepAliveHandler != nil {
	// 	allHandler := func(client mqtt.Client, message mqtt.Message) {
	// 		if logHandler != nil {
	// 			logHandler(client, message)
	// 		}
	// 		if keepAliveHandler != nil {
	// 			keepAliveHandler(client, message)
	// 		}
	// 	}
	// 	onConnectHandler = append(onConnectHandler, func(c mqtt.Client) {
	// 		if err := mqlux.Subscribe(c, "/#", allHandler); err != nil {
	// 			log.Fatal(err)
	// 		}
	// 	})
	// }

	_, err = mqlux.NewMQTTClient(config, func(c mqtt.Client) {
		log.Println("debug: connect handler called")
		for i := range onConnectHandler {
			onConnectHandler[i](c)
		}
	})

	if err != nil {
		log.Fatal(err)
	}

	t := time.NewTicker(10 * time.Second)
	lastKeepAlive := time.Now()
	for {
		select {
		case <-t.C:
			if keepAlive != 0 && time.Since(lastKeepAlive) > keepAlive {
				os.Exit(42)
			}
		case <-keepAliveC:
			lastKeepAlive = time.Now()
		}
	}
}
