package parser

import (
	"strconv"
	"strings"

	"github.com/ktt-ol/mqlux/internal/mqlux"
	"github.com/pkg/errors"
)

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
