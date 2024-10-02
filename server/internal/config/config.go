package config

import (
	"fmt"
	"os"

	"github.com/llm-operator/inference-manager/pkg/llmkind"
	"github.com/llmariner/api-usage/pkg/sender"
	"github.com/llmariner/common/pkg/db"
	"gopkg.in/yaml.v3"
)

// AssumeRoleConfig is the assume role configuration.
type AssumeRoleConfig struct {
	RoleARN    string `yaml:"roleArn"`
	ExternalID string `yaml:"externalId"`
}

func (c *AssumeRoleConfig) validate() error {
	if c.RoleARN == "" {
		return fmt.Errorf("roleArn must be set")
	}
	return nil
}

// S3Config is the S3 configuration.
type S3Config struct {
	EndpointURL string `yaml:"endpointUrl"`
	Region      string `yaml:"region"`
	Bucket      string `yaml:"bucket"`

	AssumeRole *AssumeRoleConfig `yaml:"assumeRole"`
}

// ObjectStoreConfig is the object store configuration.
type ObjectStoreConfig struct {
	S3 S3Config `yaml:"s3"`
}

// Validate validates the object store configuration.
func (c *ObjectStoreConfig) Validate() error {
	if c.S3.Region == "" {
		return fmt.Errorf("s3 region must be set")
	}
	if c.S3.Bucket == "" {
		return fmt.Errorf("s3 bucket must be set")
	}
	if ar := c.S3.AssumeRole; ar != nil {
		if err := ar.validate(); err != nil {
			return fmt.Errorf("assumeRole: %s", err)
		}
	}
	return nil
}

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
	GRPCPort         int `yaml:"grpcPort"`
	HTTPPort         int `yaml:"httpPort"`
	InternalGRPCPort int `yaml:"internalGrpcPort"`

	LLMEngine                     string `yaml:"llmEngine"`
	LLMEngineAddr                 string `yaml:"llmEngineAddr"`
	FileManagerServerAddr         string `yaml:"fileManagerServerAddr"`
	FileManagerServerInternalAddr string `yaml:"fileManagerServerInternalAddr"`

	VectorDatabase db.Config         `yaml:"vectorDatabase"`
	Database       db.Config         `yaml:"database"`
	ObjectStore    ObjectStoreConfig `yaml:"objectStore"`

	// Model is the embedding model name.
	Model string `yaml:"model"`

	AuthConfig  AuthConfig    `yaml:"auth"`
	UsageSender sender.Config `yaml:"usageSender"`
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	if c.GRPCPort <= 0 {
		return fmt.Errorf("grpcPort must be greater than 0")
	}
	if c.HTTPPort <= 0 {
		return fmt.Errorf("httpPort must be greater than 0")
	}
	if c.InternalGRPCPort <= 0 {
		return fmt.Errorf("internalGrpcPort must be greater than 0")
	}
	if c.LLMEngineAddr == "" {
		return fmt.Errorf("LLM engine addr must be set")
	}
	if c.FileManagerServerAddr == "" {
		return fmt.Errorf("file manager address must be set")
	}
	if c.FileManagerServerInternalAddr == "" {
		return fmt.Errorf("file manager server internal address must be set")
	}
	if c.Model == "" {
		return fmt.Errorf("model must be set")
	}
	if err := c.VectorDatabase.Validate(); err != nil {
		return fmt.Errorf("vector database: %s", err)
	}
	if err := c.Database.Validate(); err != nil {
		return fmt.Errorf("database: %s", err)
	}
	if err := c.ObjectStore.Validate(); err != nil {
		return fmt.Errorf("object store: %s", err)
	}
	if err := c.AuthConfig.Validate(); err != nil {
		return err
	}
	if err := c.UsageSender.Validate(); err != nil {
		return err
	}
	switch c.LLMEngine {
	case llmkind.Ollama, llmkind.VLLM:
		break
	default:
		return fmt.Errorf("unsupported llm engine: %q", c.LLMEngine)
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
