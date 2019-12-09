package mongo

// DbConfig mongo config
type DbConfig struct {
	IP     string `json:"ip"`
	DbName string `json:"dbName"`
}

// ServerConfig server config
type ServerConfig struct {
	Port string `json:"port"`
}

// AuthConfig authentication config
type AuthConfig struct {
	Secret string `json:"secret"`
}

// Config config
type Config struct {
	Mongo  *DbConfig     `json:"mongo"`
	Server *ServerConfig `json:"server"`
	Auth   *AuthConfig   `json:"auth"`
}
