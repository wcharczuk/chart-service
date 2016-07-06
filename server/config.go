package server

import (
	"os"

	"github.com/blendlabs/go-util"
)

const (
	// DefaultPort is the default port the app runs on.
	DefaultPort = "8080"

	// DefaultEnv is the default env.
	DefaultEnv = "dev"
)

var (
	// Config contains common configuration parameters.
	Config = &config{}
)

type config struct {
	key  []byte
	port string
	env  string
}

// Port is the port the server should listen on.
func (c *config) Port() string {
	if len(c.port) == 0 {
		envPort := os.Getenv("PORT")
		if !util.IsEmpty(envPort) {
			c.port = envPort
		} else {
			c.port = DefaultPort
		}
	}
	return c.port
}

// ConfigKey is the app secret we use to encrypt things.
func (c *config) Key() []byte {
	if c.key == nil {
		keyBlob := os.Getenv("ENCRYPTION_KEY")
		if len(keyBlob) != 0 {
			key, keyErr := util.Base64Decode(keyBlob)
			if keyErr != nil {
				println(keyErr.Error())
				return key
			}
			c.key = key
		} else {
			c.key = []byte{}
		}
	}
	return c.key
}

// Environment returns the current environment.
func (c *config) Environment() string {
	if len(c.env) == 0 {
		env := os.Getenv("ENV")
		if len(env) != 0 {
			c.env = env
		} else {
			c.env = DefaultEnv
		}
	}
	return c.env
}

// IsProduction returns if the app is running in production mode.
func (c *config) IsProduction() bool {
	return util.CaseInsensitiveEquals(c.Environment(), "prod")
}
