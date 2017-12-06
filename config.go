package mqlux

type Config struct {
	MQTT     MQTTConfig
	InfluxDB InfluxDBConfig
	Messages struct {
		Devices     DevicesConfig
		SpaceStatus SpaceStatusConfig
		Sensors     []SensorConfig `toml:"sensor"`
	}
	CACertFiles []string
}

type MQTTConfig struct {
	URL       string
	Username  string
	Password  string
	ClientID  string
	CSVLog    string
	KeepAlive string
}

type InfluxDBConfig struct {
	URL      string
	Username string
	Password string
	Database string
}

type SpaceStatusConfig struct {
	Topic        string
	SpaceOpen    string
	SpaceClosing string
}

type DevicesConfig struct {
	Topic   string
	Unknown string
	Devices string
	People  string
}

type SensorConfig struct {
	Topic       string
	Measurement string
	Tags        map[string]string
}
