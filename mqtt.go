package mqlux

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

func defaultCertPool() *x509.CertPool {
	certs := x509.NewCertPool()

	if !certs.AppendCertsFromPEM([]byte(mainframeCert)) {
		log.Fatal("unable to add pem to CertPool")
	}
	if !certs.AppendCertsFromPEM([]byte(spacegateCert)) {
		log.Fatal("unable to add pem to CertPool")
	}
	return certs
}

func NewMQTTClient(conf Config, onConnect mqtt.OnConnectHandler) (mqtt.Client, error) {
	opts := mqtt.NewClientOptions()

	opts.AddBroker(conf.MQTT.URL)

	if conf.MQTT.Username != "" {
		opts.SetUsername(conf.MQTT.Username)
	}
	if conf.MQTT.Password != "" {
		opts.SetPassword(conf.MQTT.Password)
	}

	opts.SetClientID(conf.MQTT.ClientID)

	certs := defaultCertPool()

	tlsConf := &tls.Config{
		InsecureSkipVerify: true,
		RootCAs:            certs,
	}

	opts.SetTLSConfig(tlsConf)

	opts.SetAutoReconnect(true)

	opts.SetKeepAlive(10 * time.Second)
	opts.SetMaxReconnectInterval(5 * time.Minute)

	opts.SetOnConnectHandler(onConnect)

	mc := mqtt.NewClient(opts)
	if tok := mc.Connect(); tok.WaitTimeout(5*time.Second) && tok.Error() != nil {
		return nil, tok.Error()
	}

	return mc, nil
}

// func Subscribe(client mqtt.Client, topic string, cb mqtt.MessageHandler) error {
// 	qos := 0
// 	tok := client.Subscribe(topic, byte(qos), cb)
// 	tok.WaitTimeout(5 * time.Second)
// 	return tok.Error()
// }

// func SensorHandler(conf Config, s SensorConfig, f func(SensorConfig, map[string]string, float64) error) mqtt.MessageHandler {
// 	callback := func(client mqtt.Client, message mqtt.Message) {
// 		log.Printf("debug: got status message for %s: %s (%v)", message.Topic(), message.Payload(), s)
// 		var tags map[string]string
// 		if s.RegexpTopic != nil {
// 			tags = s.RegexpTopic.Tags(message.Topic())
// 			if tags == nil {
// 				return
// 			}
// 		}
//
// 		v, err := strconv.ParseFloat(strings.TrimSpace(string(message.Payload())), 64)
// 		if err != nil {
// 			log.Printf("error: unable to parse float ('%s'): %s", message.Payload(), err)
// 			return
// 		}
//
// 		if tags != nil {
// 			// append static Tags to regexp tags
// 			for k, v := range s.Tags {
// 				tags[k] = v
// 			}
// 		} else {
// 			tags = s.Tags
// 		}
// 		log.Printf("debug: sensor %v %v v=%f",
// 			s.Measurement, tags, v,
// 		)
// 		if f != nil {
// 			if err := f(s, tags, v); err != nil {
// 				log.Printf("error: unable to process sensor message: %s", err)
// 			}
// 		}
// 	}
// }
//
type Handler interface {
	// Topic returns the topic this handler should be subscribed to.
	// The topic can contain wildcards. Match will check if
	// a final topic should actualy be handled.
	Topic() string
	// Match returns whether this handler handles a specific topic.
	Match(topic string) bool
	// Receive takes and processes an incoming mqtt.Message.
	Receive(client mqtt.Client, message mqtt.Message)
}

type Parser func(topic string, payload []byte, measurement string, tags map[string]string) ([]Record, error)

func Subscribe(client mqtt.Client, handler []Handler) error {
	// group handler by topic
	topicHandler := make(map[string][]Handler)
	for _, h := range handler {
		topic := h.Topic()
		topicHandler[topic] = append(topicHandler[topic], h)
	}

	// subscribe, use topicforwarder for duplicate subscriptions
	for t, h := range topicHandler {
		var tok mqtt.Token
		if len(h) == 1 {
			// TODO QOS
			tok = client.Subscribe(t, 0, h[0].Receive)
		} else {
			tok = client.Subscribe(t, 0, TopicForwarder(h).Receive)
		}
		tok.Wait()
		if err := tok.Error(); err != nil {
			return err
		}
	}
	return nil
}

type TopicForwarder []Handler

func (t TopicForwarder) Receive(client mqtt.Client, message mqtt.Message) {
	for _, h := range t {
		if h.Match(message.Topic()) {
			h.Receive(client, message)
			return
		}
	}
}

var mainframeCert = `
-----BEGIN CERTIFICATE-----
MIICqjCCAZICCQDDcRiB/QxDITANBgkqhkiG9w0BAQsFADAXMRUwEwYDVQQDDAxt
YWluZnJhbWUuaW8wHhcNMTUwMjEzMjMyMjQzWhcNMjAwMjEyMjMyMjQzWjAXMRUw
EwYDVQQDDAxtYWluZnJhbWUuaW8wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEK
AoIBAQDEhqyZE2r6QOBPbLSr4FtFWfmaSJggFdZV71+r8unGoAEa7TL1JgSZXBap
JA8eKzP1O81enQw4utwnE7bmlVlauiMfcgQ8vgPNkG/XjBTlGOGRaqlQy/7ULdgJ
rdYAujODJsCvFQQ9agocsWMtbH79kFPueSA7Y8oIElTpahp4Slc8VQeX9D90GY2p
rETIoUNTWT0k9wgNOsdLDdN7XKYKQH2dq7WQyRsnfWUDsf/eKn0rSc6SFgST7/71
5ek284/zzAxr4rOQcdBnUL+vKM6LPrLs3t34BaIXYht+ttxj6jfJ4DZS1suLSFZY
wX+zqKCjIVRLrWDNWDGLqJf4js7VAgMBAAEwDQYJKoZIhvcNAQELBQADggEBAJGx
gPYQRZD23ueJdzuT9xHaVptXpaW3LPHsvfme1uJz731Kl1NuIp5h6oYE5z4c1Gyj
0+177v2QVy4N2hkH/rTYETQ4wtg9Y5VvP0L5xcF88mH3zjgrZ1RYm3UM8d30mnNm
GwRDMitAHCim0EPFSXZ2X00v3dhX5+0jjyfRt3azRcINsKXuRbJ3tfECIEi4lv4i
dXKaevyeaCrZvVoP9LyPcbH4KO8ObVowLnG6c/eQB9QpirC5bt2UDJqWXJKW/yqp
vmvGbBGwTWhXpvdoWmBj5+qielEyBR4a6TxEr2R/YwEX624TmhlyZcnh3K3Lejdg
CqTUiKTlyh9bur7Jfn0=
-----END CERTIFICATE-----
`

var spacegateCert = `
-----BEGIN CERTIFICATE-----
MIICqjCCAZICCQDDcRiB/QxDITANBgkqhkiG9w0BAQsFADAXMRUwEwYDVQQDDAxt
YWluZnJhbWUuaW8wHhcNMTUwMjEzMjMyMjQzWhcNMjAwMjEyMjMyMjQzWjAXMRUw
EwYDVQQDDAxtYWluZnJhbWUuaW8wggEiMA0GCSqGSIb3DQEBAQUAA4IBDwAwggEK
AoIBAQDEhqyZE2r6QOBPbLSr4FtFWfmaSJggFdZV71+r8unGoAEa7TL1JgSZXBap
JA8eKzP1O81enQw4utwnE7bmlVlauiMfcgQ8vgPNkG/XjBTlGOGRaqlQy/7ULdgJ
rdYAujODJsCvFQQ9agocsWMtbH79kFPueSA7Y8oIElTpahp4Slc8VQeX9D90GY2p
rETIoUNTWT0k9wgNOsdLDdN7XKYKQH2dq7WQyRsnfWUDsf/eKn0rSc6SFgST7/71
5ek284/zzAxr4rOQcdBnUL+vKM6LPrLs3t34BaIXYht+ttxj6jfJ4DZS1suLSFZY
wX+zqKCjIVRLrWDNWDGLqJf4js7VAgMBAAEwDQYJKoZIhvcNAQELBQADggEBAJGx
gPYQRZD23ueJdzuT9xHaVptXpaW3LPHsvfme1uJz731Kl1NuIp5h6oYE5z4c1Gyj
0+177v2QVy4N2hkH/rTYETQ4wtg9Y5VvP0L5xcF88mH3zjgrZ1RYm3UM8d30mnNm
GwRDMitAHCim0EPFSXZ2X00v3dhX5+0jjyfRt3azRcINsKXuRbJ3tfECIEi4lv4i
dXKaevyeaCrZvVoP9LyPcbH4KO8ObVowLnG6c/eQB9QpirC5bt2UDJqWXJKW/yqp
vmvGbBGwTWhXpvdoWmBj5+qielEyBR4a6TxEr2R/YwEX624TmhlyZcnh3K3Lejdg
CqTUiKTlyh9bur7Jfn0=
-----END CERTIFICATE-----
`
