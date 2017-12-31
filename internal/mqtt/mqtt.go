package mqtt

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/ktt-ol/mqlux/internal/config"
	"github.com/ktt-ol/mqlux/internal/mqlux"
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

func NewMQTTClient(conf config.Config, onConnect mqtt.OnConnectHandler) (mqtt.Client, error) {
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

// Subscribe subscribes the handler function to /#.
// Should only be called once for each client.
func Subscribe(client mqtt.Client, fwd func(mqlux.Message)) error {
	// mqtt.Client only supports one callback for each topic and
	// subscribing to a wildcard topic can overwrite callbacks
	// to specific topics. We use our own router to work around
	// this limitation.
	tok := client.Subscribe("/#", 0, func(client mqtt.Client, message mqtt.Message) {
		msg := mqlux.Message{
			Time:     time.Now(),
			Payload:  message.Payload(),
			Topic:    message.Topic(),
			Retained: message.Retained(),
		}
		fwd(msg)
	})
	tok.WaitTimeout(30 * time.Second)
	return tok.Error()
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
