package transform

import (
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/ktt-ol/mqlux"
	"github.com/robertkrimen/otto"
)

func TestValueToRecords(t *testing.T) {
	vm := otto.New()
	for _, test := range []struct {
		JS    string
		Want  []mqlux.Record
		Error string
	}{
		{
			JS:   `42`,
			Want: []mqlux.Record{{Value: float64(42)}},
		},
		{
			JS:   `null`,
			Want: []mqlux.Record{{Value: nil}},
		},
		{
			JS:   `true`,
			Want: []mqlux.Record{{Value: true}},
		},
		{
			JS:   `false`,
			Want: []mqlux.Record{{Value: false}},
		},
		{
			JS:   `[{value: "Hello"}, {value: 23.5}]`,
			Want: []mqlux.Record{{Value: "Hello"}, {Value: float64(23.5)}},
		},
		{
			JS:   `_ = {measurement: "temperature", value: 99.0}`,
			Want: []mqlux.Record{{Measurement: "temperature", Value: 99.0}},
		},
		{
			JS: `_ = {measurement: "temperature", value: 99.0, tags: {"room":"kitchen", "height": 1.5}}`,
			Want: []mqlux.Record{
				{Measurement: "temperature", Value: 99.0,
					Tags: map[string]string{"room": "kitchen", "height": "1.5"}}},
		},
		{
			JS:   `[]`,
			Want: nil,
		},
		{
			JS:    `[[]]`,
			Error: "no value",
		},
	} {
		t.Run(fmt.Sprintf("parsing `%s`", test.JS), func(t *testing.T) {
			v, err := vm.Run(test.JS)
			if err != nil {
				t.Fatal("parsing JS:", err)
			}
			recs, err := valueToRecords(v)
			if err != nil {
				if test.Error != "" {
					if strings.Contains(err.Error(), test.Error) {
						return
					}
					t.Fatalf("unexpected error `%#v` (does not contain `%s`)", err, test.Error)
				}
				t.Fatal(err)
				return
			}
			if !reflect.DeepEqual(recs, test.Want) {
				t.Errorf("records from `%s` is not %v but %v", test.JS, test.Want, recs)
			}
		})
	}
}

func TestTransform(t *testing.T) {
	for _, test := range []struct {
		Name        string
		JS          string
		Msg         []mqlux.Message
		Measurement string
		Tags        map[string]string
		Want        []mqlux.Record
		Error       string
	}{
		{
			Name:        "return static value",
			JS:          `function parse(topic, payload) { return 42; }`,
			Measurement: "temperature",
			Msg:         []mqlux.Message{{}},
			Tags:        map[string]string{"room": "kitchen"},
			Want: []mqlux.Record{
				{Measurement: "temperature", Value: 42.0, Tags: map[string]string{"room": "kitchen"}},
			},
		},
		{
			Name:        "record object with tags",
			JS:          `function parse(topic, payload) { return {"tags": {"room": "kitchen", "sensor": "dht22"}, "value": parseFloat(payload)}; }`,
			Msg:         []mqlux.Message{{Topic: "/sensors/kitchen/temp", Payload: []byte("42.5")}},
			Measurement: "temperature",
			Tags:        map[string]string{"script": "test"},
			Want: []mqlux.Record{
				{
					Measurement: "temperature",
					Value:       42.5,
					Tags:        map[string]string{"script": "test", "sensor": "dht22", "room": "kitchen"},
				},
			},
		},
		{
			Name: "persist data in js vm",
			JS:   `var counter = 0; function parse(topic, payload) { counter += parseFloat(payload); return counter };`,
			Msg: []mqlux.Message{
				{Topic: "/sensors/counter/incr", Payload: []byte("2")},
				{Topic: "/sensors/counter/incr", Payload: []byte("5")},
			},
			Measurement: "counter",
			Tags:        map[string]string{"script": "test"},
			Want: []mqlux.Record{
				{
					Measurement: "counter",
					Value:       2.0,
					Tags:        map[string]string{"script": "test"},
				},
				{
					Measurement: "counter",
					Value:       7.0,
					Tags:        map[string]string{"script": "test"},
				},
			},
		},
	} {
		t.Run(test.Name, func(t *testing.T) {
			transf, err := New(test.JS)
			if err != nil {
				t.Fatal("parsing JS:", err)
			}
			t.Log(test.Msg)
			var recs []mqlux.Record
			for _, msg := range test.Msg {
				rec, err := transf.Parse(msg, test.Measurement, test.Tags)
				if err != nil {
					if test.Error != "" {
						if strings.Contains(err.Error(), test.Error) {
							return
						}
						t.Fatalf("unexpected error `%#v` (does not contain `%s`)", err, test.Error)
					}
					t.Fatal(err)
					return
				}
				recs = append(recs, rec...)
			}
			if !reflect.DeepEqual(recs, test.Want) {
				t.Errorf("records from `%s` is not %v but %v", test.JS, test.Want, recs)
			}
		})
	}
}
