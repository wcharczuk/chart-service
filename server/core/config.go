package core

import (
	"os"

	"github.com/blendlabs/go-util"
	"github.com/blendlabs/go-util/env"
	"github.com/blendlabs/spiffy"
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
	key     []byte
	port    string
	env     string
	authKey string
}

// Port is the port the server should listen on.
func (c *config) Port() string {
	return env.Env().String("PORT", DefaultPort)
}

// ConfigKey is the app secret we use to encrypt things.
func (c *config) Key() []byte {
	if c.key == nil {
		keyBlob := os.Getenv("ENCRYPTION_KEY")
		if len(keyBlob) != 0 {
			key, keyErr := util.String.Base64Decode(keyBlob)
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

func (c *config) AuthKey() string {
	if len(c.authKey) == 0 {
		c.authKey = os.Getenv("AUTH_KEY")
	}
	return c.authKey
}

func (c *config) IsProduction() bool {
	return util.String.CaseInsensitiveEquals(env.Env().String("ENV", DefaultEnv), "prod")
}

// SetupDatabaseContext writes the config to spiffy.
func SetupDatabaseContext() error {
	err := spiffy.OpenDefault(spiffy.NewConnectionFromEnvironment())
	if err != nil {
		return err
	}

	spiffy.Default().Connection.SetMaxIdleConns(64)
	spiffy.Default().Connection.SetMaxOpenConns(64)
	return nil
}
