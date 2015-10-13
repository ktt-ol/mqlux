package mqlux

import (
	"net/url"
	"time"

	"github.com/influxdb/influxdb/client"
)

type InfluxDBClient struct {
	client   *client.Client
	database string
}

func NewInfluxDBClient(conf Config) (*InfluxDBClient, error) {

	// u, err := url.Parse("http://localhost:8086/")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	var err error
	clientCfg := client.NewConfig()
	u, err := url.Parse(conf.InfluxDB.URL)
	if err != nil {
		return nil, err
	}
	clientCfg.URL = *u
	clientCfg.Username = conf.InfluxDB.Username
	clientCfg.Password = conf.InfluxDB.Password

	c, err := client.NewClient(clientCfg)
	if err != nil {
		return nil, err
	}

	return &InfluxDBClient{
		client:   c,
		database: conf.InfluxDB.Database,
	}, nil
}

func (i *InfluxDBClient) WriteStatus(s SpaceStatus) error {
	return i.writePoints(statusPoints(s))
}

func (i *InfluxDBClient) WriteDevices(d Devices) error {
	return i.writePoints(devicesPoints(d))
}

func (i *InfluxDBClient) writePoints(pts []client.Point) error {
	bps := client.BatchPoints{
		Points:          pts,
		Database:        i.database,
		RetentionPolicy: "default",
	}

	_, err := i.client.Write(bps)
	if err != nil {
		return err
	}
	return nil
}

func statusPoints(s SpaceStatus) []client.Point {
	now := time.Now()
	open := 0
	if s.Open {
		open = 1
	}
	closing := 0
	if s.Closing {
		closing = 1
	}
	return []client.Point{
		{
			Measurement: "space_open",
			Fields: map[string]interface{}{
				"value": open,
			},
			Time:      now,
			Precision: "s",
		},
		{
			Measurement: "space_closing",
			Fields: map[string]interface{}{
				"value": closing,
			},
			Time:      now,
			Precision: "s",
		},
	}
}

func devicesPoints(d Devices) []client.Point {
	now := time.Now()
	return []client.Point{
		{
			Measurement: "devices_total",
			Fields: map[string]interface{}{
				"value": d.Devices,
			},
			Time:      now,
			Precision: "s",
		},
		{
			Measurement: "devices_unknown",
			Fields: map[string]interface{}{
				"value": d.Unknown,
			},
			Time:      now,
			Precision: "s",
		},
		{Measurement: "people",
			Fields: map[string]interface{}{
				"value": d.People,
			},
			Time:      now,
			Precision: "s",
		},
	}
}
