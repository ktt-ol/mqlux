package parser

import (
	"encoding/json"
	"log"
	"regexp"
	"strconv"
	"strings"

	"github.com/ktt-ol/mqlux/internal/config"
	"github.com/ktt-ol/mqlux/internal/mqlux"
	"github.com/ktt-ol/mqlux/internal/mqtt"
	"github.com/pkg/errors"
)

func NetDeviceParser(conf config.Devices) mqtt.Parser {
	return func(msg mqlux.Message, measurement string, tags map[string]string) ([]mqlux.Record, error) {

		var val map[string]interface{}
		if err := json.Unmarshal(msg.Payload, &val); err != nil {
			return nil, errors.Wrapf(err, "unable to marshal json `%s` from %s for %s", msg.Payload, msg.Topic, measurement)
		}

		var people, unknown, total int

		if v, ok := val[conf.People].(float64); ok {
			people = int(v)
		}
		if v, ok := val[conf.Unknown].(float64); ok {
			unknown = int(v)
		}
		if v, ok := val[conf.Devices].(float64); ok {
			total = int(v)
		}
		recs := []mqlux.Record{
			mqlux.Record{
				Measurement: "devices_unknown",
				Value:       unknown,
			},
			mqlux.Record{
				Measurement: "devices_total",
				Value:       total,
			},
			mqlux.Record{
				Measurement: "people",
				Value:       people,
			},
		}

		log.Printf("debug: net-devices devices=%d unknown=%d people=%d",
			total, unknown, people,
		)

		return recs, nil
	}
}

func SpaceStatusParser(conf config.SpaceStatus) mqtt.Parser {
	closing := regexp.MustCompile(conf.SpaceClosing)
	open := regexp.MustCompile(conf.SpaceOpen)

	return func(msg mqlux.Message, measurement string, tags map[string]string) ([]mqlux.Record, error) {
		val := 0.0

		if open.Match(msg.Payload) {
			log.Printf("debug: net-devices open")
			val = 1.0
		} else if closing.Match(msg.Payload) {
			log.Printf("debug: net-devices closing")
			val = 0.5
		} else {
			log.Printf("debug: net-devices closed")
		}

		return []mqlux.Record{
			{
				Measurement: "space_open",
				Value:       val,
			},
		}, nil
	}
}

func FloatParser(msg mqlux.Message, measurement string, tags map[string]string) ([]mqlux.Record, error) {
	v, err := strconv.ParseFloat(strings.TrimSpace(string(msg.Payload)), 32)
	if err != nil {
		return nil, errors.Wrapf(err, "parsing float %s from %s for %s", msg.Payload, msg.Topic, measurement)
	}

	return []mqlux.Record{
		{
			Measurement: measurement,
			Tags:        tags,
			Value:       v,
		},
	}, nil
}
