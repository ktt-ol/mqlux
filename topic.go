package mqlux

import (
	"errors"
	"regexp"
	"strings"
)

type RegexpTopic struct {
	SubscribeTopic string
	Topic          string
	re             *regexp.Regexp
}

var NoRegexpTopic = errors.New("topic contains no regexp")

func NewRegexpTopic(topic string) (*RegexpTopic, error) {
	rt := RegexpTopic{Topic: topic}

	st, ok := nonRegexpTopic(topic)
	if !ok {
		return nil, NoRegexpTopic
	}

	rt.SubscribeTopic = st
	if !strings.HasSuffix(topic, "$") {
		topic += "$"
	}
	re, err := regexp.Compile(topic)
	if err != nil {
		return nil, err
	}
	rt.re = re
	return &rt, nil
}

func (rt *RegexpTopic) Match(topic string) map[string]string {
	var tags map[string]string
	tagNames := rt.re.SubexpNames()
	sub := rt.re.FindStringSubmatch(topic)
	for i := 1; i < len(sub); i++ {
		if sub[i] != "" {
			if tags == nil {
				tags = make(map[string]string)
			}
			tags[tagNames[i]] = sub[i]
		}
	}
	return tags
}

func nonRegexpTopic(topic string) (string, bool) {
	nonRegexp := regexp.MustCompile("^[a-zA-Z0-9-_]*$")
	var result string
	for _, part := range strings.Split(topic, "/") {
		if nonRegexp.MatchString(part) {
			result += part + "/"
		} else {
			return result + "#", true
		}
	}
	return topic, false
}
