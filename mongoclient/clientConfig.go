package mongoclient

import "fmt"

// ClientConfig config for Client
type ClientConfig struct {
	Database string
	Host     string
	Port     string
}

// NewClientConfig creates a new ClientConfig with default values for host and port
func NewClientConfig() ClientConfig {
	return ClientConfig{
		Host: "127.0.0.1",
		Port: "27017",
	}
}

func (c ClientConfig) url() string {
	return fmt.Sprintf("mongodb://%s:%s", c.Host, c.Port)
}
