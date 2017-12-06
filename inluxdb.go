package mqlux

import (
	"net/url"
	"time"

	"github.com/influxdb/influxdb/client"
)

type Record struct {
	Measurement string
	Tags        map[string]string
	Value       interface{}
}

type Writer func([]Record) error

type InfluxDBClient struct {
	client   *client.Client
	database string
}

func NewInfluxDBClient(conf Config) (*InfluxDBClient, error) {
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

func (i *InfluxDBClient) Write(recs []Record) error {
	pts := make([]client.Point, len(recs))
	for i, rec := range recs {
		pts[i] = client.Point{
			Measurement: rec.Measurement,
			Fields: map[string]interface{}{
				"value": rec.Value,
			},
			Tags: rec.Tags,
			Time: time.Now(),
		}
	}
	return i.writePoints(pts)
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
