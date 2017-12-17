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
