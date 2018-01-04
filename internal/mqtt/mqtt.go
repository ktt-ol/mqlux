package mqtt

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/ktt-ol/mqlux/internal/config"
	"github.com/ktt-ol/mqlux/internal/mqlux"
)

func connect(conf config.MQTT, onConnect mqtt.OnConnectHandler) (mqtt.Client, error) {
	opts := mqtt.NewClientOptions()

	opts.AddBroker(conf.URL)

	if conf.Username != "" {
		opts.SetUsername(conf.Username)
	}
	if conf.Password != "" {
		opts.SetPassword(conf.Password)
	}

	if conf.ClientID == "" {
		conf.ClientID = fmt.Sprintf("mqlux-%06d\n", time.Now().Nanosecond()/1000)
	}

	opts.SetClientID(conf.ClientID)

	var certs *x509.CertPool
	if conf.TLSServerCert != "" {
		certs = x509.NewCertPool()
		if !certs.AppendCertsFromPEM([]byte(conf.TLSServerCert)) {
			return nil, errors.New("unable to add tls_server_cert to CertPool")
		}
	}
	tlsConf := &tls.Config{
		InsecureSkipVerify: conf.TLSServerInsecure,
		RootCAs:            certs,
	}

	opts.SetTLSConfig(tlsConf)

	opts.SetAutoReconnect(true)

	opts.SetKeepAlive(30 * time.Second)
	opts.SetMaxReconnectInterval(5 * time.Minute)

	opts.SetOnConnectHandler(onConnect)

	mc := mqtt.NewClient(opts)
	if tok := mc.Connect(); tok.WaitTimeout(10*time.Second) && tok.Error() != nil {
		return nil, tok.Error()
	}

	return mc, nil
}

type Disconnector interface {
	Disconnect(waitms uint)
}

// Subscribe connects to the MQTT server and subscribes the handler function to /#.
// Should only be called once for each client.
func Subscribe(config config.MQTT, fwd func(mqlux.Message)) (Disconnector, error) {
	c, err := connect(config, func(c mqtt.Client) {
		log.Print("debug: on connect")
		// mqtt.Client only supports one callback for each topic and
		// subscribing to a wildcard topic can overwrite callbacks
		// to specific topics. We use our own router to work around
		// this limitation.
		tok := c.Subscribe("/#", 0, func(client mqtt.Client, message mqtt.Message) {
			msg := mqlux.Message{
				Time:     time.Now(),
				Payload:  message.Payload(),
				Topic:    message.Topic(),
				Retained: message.Retained(),
			}
			fwd(msg)
		})
		tok.WaitTimeout(30 * time.Second)
		if err := tok.Error(); err != nil {
			log.Print("error: on connect: ", err)
		}
	})

	return c, err
}
