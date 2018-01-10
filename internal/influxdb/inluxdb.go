package influxdb

import (
	"net/url"
	"time"

	"github.com/influxdb/influxdb/client"
	"github.com/ktt-ol/mqlux/internal/config"
	"github.com/ktt-ol/mqlux/internal/mqlux"
)

type InfluxDBClient struct {
	client          *client.Client
	database        string
	retentionPolicy string
}

func NewInfluxDBClient(conf config.Config) (*InfluxDBClient, error) {
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
		client:          c,
		database:        conf.InfluxDB.Database,
		retentionPolicy: conf.InfluxDB.RetentionPolicy,
	}, nil
}

func (i *InfluxDBClient) Write(recs []mqlux.Record) error {
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
		RetentionPolicy: i.retentionPolicy,
	}

	_, err := i.client.Write(bps)
	if err != nil {
		return err
	}
	return nil
}
