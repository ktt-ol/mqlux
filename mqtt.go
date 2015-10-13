package mqlux

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"log"
	"regexp"
	"time"

	mqtt "git.eclipse.org/gitroot/paho/org.eclipse.paho.mqtt.golang.git"
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

type MQTTClient struct {
	client *mqtt.Client
	openc  chan struct{}
}

func NewMQTTClient(conf Config) (*MQTTClient, error) {
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
		RootCAs: certs,
	}

	opts.SetTLSConfig(tlsConf)

	opts.SetAutoReconnect(false)
	opts.SetKeepAlive(5 * time.Second)
	// opts.SetMaxReconnectInterval(10 * time.Second)

	// close channel to notify disconnect
	openc := make(chan struct{})
	opts.SetConnectionLostHandler(func(c *mqtt.Client, err error) {
		close(openc)
	})

	mc := mqtt.NewClient(opts)
	if tok := mc.Connect(); tok.WaitTimeout(5*time.Second) && tok.Error() != nil {
		return nil, tok.Error()
	}

	return &MQTTClient{
		client: mc,
		openc:  openc,
	}, nil
}

func (m *MQTTClient) WaitDisconnect() {
	<-m.openc // block till channel gets closed
}

func (m *MQTTClient) Subscribe(topic string, cb mqtt.MessageHandler) error {
	qos := 0
	tok := m.client.Subscribe(topic, byte(qos), cb)
	tok.WaitTimeout(2 * time.Second)
	return tok.Error()
}

type Devices struct {
	People  int
	Unknown int
	Devices int
}

func NetDeviceHandler(conf Config, f func(Devices) error) mqtt.MessageHandler {
	callback := func(client *mqtt.Client, message mqtt.Message) {
		log.Printf("debug: got net message for %s: %s", message.Topic(), message.Payload())
		var msg map[string]interface{}
		if err := json.Unmarshal(message.Payload(), &msg); err != nil {
			log.Printf("error: unable to unmarshal json: %s `%s`", err, message.Payload())
			return
		}
		devices := Devices{}
		if v, ok := msg[conf.Messages.Devices.People].(float64); ok {
			devices.People = int(v)
		}
		if v, ok := msg[conf.Messages.Devices.Unknown].(float64); ok {
			devices.Unknown = int(v)
		}
		if v, ok := msg[conf.Messages.Devices.Devices].(float64); ok {
			devices.Devices = int(v)
		}

		log.Printf("info: net-devices devices=%d unknown=%d people=%d",
			devices.Devices, devices.Unknown, devices.People,
		)

		if err := f(devices); err != nil {
			log.Printf("error: unable to process devices message: %s", err)
		}
	}
	return callback
}

type SpaceStatus struct {
	Open    bool
	Closing bool
}

func SpaceStatusHandler(conf Config, f func(SpaceStatus) error) mqtt.MessageHandler {
	closing := regexp.MustCompile(conf.Messages.SpaceStatus.SpaceClosing)
	open := regexp.MustCompile(conf.Messages.SpaceStatus.SpaceOpen)
	callback := func(client *mqtt.Client, message mqtt.Message) {
		log.Printf("debug: got status message for %s: %s", message.Topic(), message.Payload())

		status := SpaceStatus{}
		if open.Match(message.Payload()) {
			status.Open = true
		}
		if closing.Match(message.Payload()) {
			status.Closing = true
		}

		log.Printf("info: net-devices open=%v closing=%v",
			status.Open, status.Closing,
		)

		if err := f(status); err != nil {
			log.Printf("error: unable to process spacestatus message: %s", err)
		}
	}
	return callback
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