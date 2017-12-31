package topic

import (
	"reflect"
	"testing"
)

func TestMatch(t *testing.T) {
	for _, test := range []struct {
		TopicTemplate string
		Topic         string
		Tags          map[string]string
	}{
		{TopicTemplate: `/net/wlan/ap-1`},
		{TopicTemplate: `/net/wlan/ap-(?P<ap>\d+)`, Topic: "/net/wlan/foo"},
		{TopicTemplate: `/net/wlan/ap-(?P<ap>\d+)`, Topic: "/net/wlan/ap-"},
		{TopicTemplate: `/net/wlan/ap-(?P<ap>\d+)`, Topic: "/net/wlan/ap-a"},
		{TopicTemplate: `/net/wlan/ap-(?P<ap>\d+)`, Topic: "/net/wlan/ap-1a"},
		{
			TopicTemplate: `/net/wlan/ap-(?P<ap>\d+)`,
			Topic:         "/net/wlan/ap-1",
			Tags:          map[string]string{"ap": "1"},
		},
		{
			TopicTemplate: `/net/wlan/ap-(?P<ap>\d+)/radio-(?P<radio>\d)/tx`,
			Topic:         "/net/wlan/ap-1/radio-2/tx",
			Tags:          map[string]string{"ap": "1", "radio": "2"},
		},
	} {
		rt, err := New(test.TopicTemplate, "", nil, nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		tags := rt.Tags(test.Topic)
		if !reflect.DeepEqual(tags, test.Tags) {
			t.Errorf("%s did not match %s %s != %s", test.Topic, test.TopicTemplate, tags, test.Tags)
		}
	}
}

func TestNonRegexpTopic(t *testing.T) {
	for _, test := range []struct {
		Topic    string
		Want     string
		IsRegexp bool
	}{
		{Topic: "/", Want: "/", IsRegexp: false},
		{Topic: "/foo", Want: "/foo", IsRegexp: false},
		{Topic: "/foo/", Want: "/foo/", IsRegexp: false},
		{Topic: "/foo/ba[rz]", Want: "/foo/#", IsRegexp: true},
		{Topic: `/net/wlan/ap-(?P<ap>\d+)/tx`, Want: "/net/wlan/#", IsRegexp: true},
	} {
		actual, isRegexp := nonRegexpTopic(test.Topic)
		if test.IsRegexp != isRegexp || actual != test.Want {
			t.Errorf("topic %s: %s(%v) != %s(%v)", test.Topic, test.Want, test.IsRegexp, actual, isRegexp)
		}
	}
}
