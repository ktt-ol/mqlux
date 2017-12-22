package config

type Config struct {
	MQTT     MQTT
	InfluxDB InfluxDB
	Messages struct {
		Devices     Devices
		SpaceStatus SpaceStatus
		Sensors     []Sensor `toml:"sensor"`
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

type SpaceStatus struct {
	Topic        string
	SpaceOpen    string
	SpaceClosing string
}

type Devices struct {
	Topic   string
	Unknown string
	Devices string
	People  string
}

type Sensor struct {
	Topic       string
	Measurement string
	Tags        map[string]string
}
