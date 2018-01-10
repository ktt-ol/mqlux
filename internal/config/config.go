package config

type Config struct {
	MQTT          MQTT
	InfluxDB      InfluxDB
	Subscriptions []Subscription `toml:"subscription"`
	CACertFiles   []string
}

type MQTT struct {
	URL               string
	Username          string
	Password          string
	ClientID          string
	CSVLog            string
	KeepAlive         string
	TLSServerCert     string `toml:"tls_server_cert"`
	TLSServerInsecure bool   `toml:"tls_server_insecure"`
}

type InfluxDB struct {
	URL             string
	Username        string
	Password        string
	Database        string
	RetentionPolicy string `toml:"retention_policy"`
}

type Subscription struct {
	Topic           string
	Measurement     string
	Tags            map[string]string
	Script          string
	IncludeRetained bool `toml:"include_retained"`
}
