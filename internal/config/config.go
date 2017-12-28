package config

type Config struct {
	MQTT          MQTT
	InfluxDB      InfluxDB
	Subscriptions []Subscription `toml:"subscription"`
	CACertFiles   []string
}

type MQTT struct {
	URL       string
	Username  string
	Password  string
	ClientID  string
	CSVLog    string
	KeepAlive string
}

type InfluxDB struct {
	URL      string
	Username string
	Password string
	Database string
}

type Subscription struct {
	Topic       string
	Measurement string
	Tags        map[string]string
	Script      string
}
