package topic

import (
	"log"
	"regexp"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/ktt-ol/mqlux"
)

type Topic struct {
	subscribeTopic string
	re             *regexp.Regexp
	measurement    string
	tags           map[string]string
	parser         mqlux.Parser
	writer         mqlux.Writer
}

func New(topic, measurement string, tags map[string]string, parser mqlux.Parser, writer mqlux.Writer) (*Topic, error) {
	t := Topic{
		measurement: measurement,
		tags:        tags,
		parser:      parser,
		writer:      writer,
	}

	st, ok := nonRegexpTopic(topic)
	if !ok {
		t.subscribeTopic = topic
		return &t, nil
	}

	t.subscribeTopic = st
	if !strings.HasSuffix(topic, "$") {
		topic += "$"
	}
	re, err := regexp.Compile(topic)
	if err != nil {
		return nil, err
	}
	t.re = re
	return &t, nil
}

func (t *Topic) Topic() string {
	return t.subscribeTopic
}

func (t *Topic) Match(topic string) bool {
	if t.re == nil {
		return topic == t.subscribeTopic
	}
	return t.re.MatchString(topic)
}

func (t *Topic) Receive(client mqtt.Client, msg mqtt.Message) {
	tags := t.Tags(msg.Topic())
	records, err := t.parser(msg.Topic(), msg.Payload(), t.measurement, tags)
	if err != nil {
		// TODO logger
		log.Println("error: parsing ", err)
		return
	}

	if records != nil {
		err := t.writer(records)
		if err != nil {
			// TODO logger
			log.Println("error: writing records", err)
		}
	}
}

func (t *Topic) Tags(topic string) map[string]string {
	if t.re == nil {
		return t.tags
	}
	tagNames := t.re.SubexpNames()
	tags := make(map[string]string, len(tagNames)+len(t.tags))
	for k, v := range t.tags {
		tags[k] = v
	}
	sub := t.re.FindStringSubmatch(topic)
	for i := 1; i < len(sub); i++ {
		if sub[i] != "" {
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
