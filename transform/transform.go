package transform

import (
	"github.com/ktt-ol/mqlux"
	"github.com/pkg/errors"
	"github.com/robertkrimen/otto"
)

// func main() {

// 	vm := otto.New()
// 	script, err := vm.Compile("", `function foo(msg) { return JSON.parse(msg); }`)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	val, err := vm.Run(script)
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	for i := 0; i < 1; i++ {
// 		val, err = vm.Run(`foo("{\"measurement\": \"hello\", \"tags\": {\"foo\": \"bar\"}, \"value\": null}")`)
// 		if err != nil {
// 			log.Fatal(err)
// 		}
// 		ToRecord(val)
// 	}
// 	log.Println(ToRecord(val))
// }

func valueToRecords(v otto.Value) ([]mqlux.Record, error) {
	if v.IsObject() && v.Object().Class() == "Array" {
		return arrayToRecords(v)
	}
	if v.IsObject() && v.Object().Class() == "Object" {
		rec, err := objectToRecord(v)
		if err != nil {
			return nil, err
		}
		return []mqlux.Record{*rec}, nil
	}
	val, err := valueToValue(v)
	if err != nil {
		return nil, err
	}
	return []mqlux.Record{{Value: val}}, nil
}

func valueToStringMap(v otto.Value) (map[string]string, error) {
	tags := make(map[string]string)
	if !v.IsObject() {
		return nil, errors.New("not an object")
	}
	obj := v.Object()
	for _, k := range obj.Keys() {
		vv, err := obj.Get(k)
		if err != nil {
			return nil, err
		}
		tags[k], err = vv.ToString()
		if err != nil {
			return nil, err
		}
	}
	return tags, nil
}
func valueToValue(v otto.Value) (interface{}, error) {
	var result interface{}
	var err error
	if v.IsUndefined() {
		return nil, errors.New("no value")
	}
	if v.IsNumber() {
		result, err = v.ToFloat()
		if err != nil {
			return nil, err
		}
	} else if v.IsBoolean() {
		result, err = v.ToBoolean()
		if err != nil {
			return nil, err
		}
	} else if v.IsNull() {
	} else {
		result, err = v.ToString()
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

func objectToRecord(v otto.Value) (*mqlux.Record, error) {
	rec := mqlux.Record{}
	if !v.IsObject() {
		return nil, errors.Errorf("not an object but class: %s", v.Class())
	}
	o := v.Object()

	var err error
	// measurement
	v, err = o.Get("measurement")
	if err != nil {
		return nil, err
	}
	if v.IsString() {
		rec.Measurement, err = v.ToString()
		if err != nil {
			return nil, err
		}
	}

	// value
	v, err = o.Get("value")
	if err != nil {
		return nil, err
	}
	if rec.Value, err = valueToValue(v); err != nil {
		return nil, err
	}

	// tags
	v, err = o.Get("tags")
	if err != nil {
		return nil, err
	}
	if !v.IsUndefined() {
		if rec.Tags, err = valueToStringMap(v); err != nil {
			return nil, err
		}
	}

	return &rec, nil
}

func arrayToRecords(v otto.Value) ([]mqlux.Record, error) {
	if !v.IsObject() {
		return nil, errors.New("not an object, expected array object")
	}

	o := v.Object()
	if o.Class() != "Array" {
		return nil, errors.New("not an array")
	}
	var recs []mqlux.Record
	for _, k := range o.Keys() {
		v, err := o.Get(k)
		if err != nil {
			return nil, err
		}
		rec, err := objectToRecord(v)
		if err != nil {
			return nil, errors.Wrap(err, "extracting record object")
		}
		recs = append(recs, *rec)
	}

	return recs, nil
}
