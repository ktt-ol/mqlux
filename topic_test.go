package mqlux

import (
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func TestMatch(t *testing.T) {
	for _, test := range []struct {
		TopicTemplate string
		Topic         string
		Tags          map[string]string
		NoRegexp      bool
	}{
		{TopicTemplate: `/net/wlan/ap-1`, NoRegexp: true},
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
		rt, err := NewRegexpTopic(test.TopicTemplate)
		if err == NoRegexpTopic && test.NoRegexp {
			continue
		}
		if err != nil {
			t.Fatal(err)
		}
		tags := rt.Match(test.Topic)
		if !reflect.DeepEqual(tags, test.Tags) {
			t.Errorf("%s did not match %s %s != %s", test.Topic, test.TopicTemplate, tags, test.Tags)
		}
	}
}

func Match(topicTemplate, topic string) (map[string]string, error) {
	if !strings.HasSuffix(topicTemplate, "$") {
		topicTemplate += "$"
	}
	re, err := regexp.Compile(topicTemplate)
	if err != nil {
		return nil, err
	}
	var tags map[string]string
	tagNames := re.SubexpNames()
	sub := re.FindStringSubmatch(topic)
	for i := 1; i < len(sub); i++ {
		if sub[i] != "" {
			if tags == nil {
				tags = make(map[string]string)
			}
			tags[tagNames[i]] = sub[i]
		}
	}
	return tags, nil
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
			t.Errorf("topic %s: %s(%s) != %s(%s)", test.Topic, test.Want, test.IsRegexp, actual, isRegexp)
		}
	}
}
