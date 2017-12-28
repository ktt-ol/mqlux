package config

type Config struct {
	MQTT     MQTT
	InfluxDB InfluxDB
	Messages struct {
		Sensors []Sensor `toml:"sensor"`
	}
	CACertFiles []string
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

type Sensor struct {
	Topic       string
	Measurement string
	Tags        map[string]string
	Script      string
}
