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
	r := Topic{
		measurement: measurement,
		tags:        tags,
		parser:      parser,
		writer:      writer,
	}

	st, ok := nonRegexpTopic(topic)
	if !ok {
		r.subscribeTopic = topic
		return &r, nil
	}

	r.subscribeTopic = st
	if !strings.HasSuffix(topic, "$") {
		topic += "$"
	}
	re, err := regexp.Compile(topic)
	if err != nil {
		return nil, err
	}
	r.re = re
	return &r, nil
}

func (r *Topic) Topic() string {
	return r.subscribeTopic
}

func (r *Topic) Match(topic string) bool {
	if r.re == nil {
		return topic == r.subscribeTopic
	}
	return r.re.MatchString(topic)
}

func (r *Topic) Receive(client mqtt.Client, msg mqtt.Message) {
	tags := r.Tags(msg.Topic())
	records, err := r.parser(msg.Topic(), msg.Payload(), r.measurement, tags)
	if err != nil {
		// TODO logger
		log.Println("error: parsing ", err)
		return
	}

	if records != nil {
		err := r.writer(records)
		if err != nil {
			// TODO logger
			log.Println("error: writing records", err)
		}
	}
}

func (r *Topic) Tags(topic string) map[string]string {
	if r.re == nil {
		return r.tags
		return nil
	}
	tagNames := r.re.SubexpNames()
	tags := make(map[string]string, len(tagNames)+len(r.tags))
	for k, v := range r.tags {
		tags[k] = v
	}
	sub := r.re.FindStringSubmatch(topic)
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
