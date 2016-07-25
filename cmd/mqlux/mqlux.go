package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/comail/colog"
	"github.com/ktt-ol/mqlux"
)

func main() {
	colog.Register()
	colog.ParseFields(true)
	colog.SetMinLevel(colog.LDebug)
	// mqtt.DEBUG = log.New(os.Stdout, "[mqtt] ", log.LstdFlags)

	configFile := flag.String("config", "mqlux.tml", "configuration")
	flag.Parse()

	config := mqlux.Config{}
	_, err := toml.DecodeFile(*configFile, &config)
	if err != nil {
		log.Fatal(err)
	}

	wait := 250 * time.Millisecond
	maxWait := 60 * time.Second
	for {
		// 2015-10-13: auto reconnect from mqtt is broken (keep alive keeps err-ing)
		// reconnect here
		if err := do(config); err != nil {
			log.Println("error: ", err)
			time.Sleep(wait)
			if wait < maxWait {
				wait *= 2
			}
		} else {
			wait = 250 * time.Millisecond
		}
	}
}

func do(config mqlux.Config) error {

	c, err := mqlux.NewMQTTClient(config)
	if err != nil {
		return err
	}

	if config.InfluxDB.URL != "" {
		fmt.Println(config.InfluxDB)
		db, err := mqlux.NewInfluxDBClient(config)
		if err != nil {
			return err
		}
		if err := c.Subscribe(config.Messages.Devices.Topic,
			mqlux.NetDeviceHandler(config, db.WriteDevices)); err != nil {
			return err
		}

		if err := c.Subscribe(config.Messages.SpaceStatus.Topic,
			mqlux.SpaceStatusHandler(config, db.WriteStatus)); err != nil {
			return err
		}
		for _, sensor := range config.Messages.Sensors {
			if err := c.Subscribe(sensor.Topic,
				mqlux.SensorHandler(config, sensor, db.WriteSensor)); err != nil {
				return err
			}
		}
	}

	if config.MQTT.CSVLog != "" {
		logger, err := mqlux.NewMQTTLogger(config.MQTT.CSVLog)
		if err != nil {
			return err
		}
		if err := c.Subscribe("/#", logger.Log); err != nil {
			return err
		}
	}

	c.WaitDisconnect()
	return nil
}
