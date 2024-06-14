package config

import (
	"fmt"
	"os"

	"github.com/llm-operator/common/pkg/db"
	"gopkg.in/yaml.v3"
)

// AuthConfig is the authentication configuration.
type AuthConfig struct {
	Enable                 bool   `yaml:"enable"`
	RBACInternalServerAddr string `yaml:"rbacInternalServerAddr"`
}

// Validate validates the configuration.
func (c *AuthConfig) Validate() error {
	if !c.Enable {
		return nil
	}
	if c.RBACInternalServerAddr == "" {
		return fmt.Errorf("rbacInternalServerAddr must be set")
	}
	return nil
}

// Config is the configuration.
type Config struct {
	GRPCPort int `yaml:"grpcPort"`
	HTTPPort int `yaml:"httpPort"`

	FileManagerServerAddr string    `yaml:"fileManagerServerAddr"`
	VectorDatabase        db.Config `yaml:"vectorDatabase"`
	Database              db.Config `yaml:"database"`

	AuthConfig AuthConfig `yaml:"auth"`
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.GRPCPort <= 0 {
		return fmt.Errorf("grpcPort must be greater than 0")
	}
	if c.HTTPPort <= 0 {
		return fmt.Errorf("httpPort must be greater than 0")
	}
	if c.FileManagerServerAddr == "" {
		return fmt.Errorf("file manager address must be set")
	}
	if err := c.VectorDatabase.Validate(); err != nil {
		return fmt.Errorf("vector database: %s", err)
	}
	if err := c.Database.Validate(); err != nil {
		return fmt.Errorf("database: %s", err)
	}
	if err := c.AuthConfig.Validate(); err != nil {
		return err
	}
	return nil
}

// Parse parses the configuration file at the given path, returning a new
// Config struct.
func Parse(path string) (Config, error) {
	var config Config

	b, err := os.ReadFile(path)
	if err != nil {
		return config, fmt.Errorf("config: read: %s", err)
	}

	if err = yaml.Unmarshal(b, &config); err != nil {
		return config, fmt.Errorf("config: unmarshal: %s", err)
	}
	return config, nil
}
