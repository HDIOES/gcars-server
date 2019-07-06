package util

//Configuration struct holds values of settings
type Configuration struct {
	Port               int    `json:"port"`
	DatabaseURL        string `json:"database_url"`
	MaxOpenConnections int    `json:"maxOpenConnections"`
	MaxIdleConnections int    `json:"maxIdleConnections"`
	ConnectionTimeout  int    `json:"connectionTimeout"`
}
