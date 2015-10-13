package mqlux

type Config struct {
	MQTT     MQTTConfig
	InfluxDB InfluxDBConfig
	Messages struct {
		Devices     DevicesConfig
		SpaceStatus SpaceStatusConfig
	}
	CACertFiles []string
}

type MQTTConfig struct {
	URL      string
	Username string
	Password string
	ClientID string
	CSVLog   string
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
