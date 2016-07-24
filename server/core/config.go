package core

import (
	"fmt"
	"os"

	"github.com/blendlabs/go-util"
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

func (c *config) AuthKey() string {
	if len(c.authKey) == 0 {
		c.authKey = os.Getenv("AUTH_KEY")
	}
	return c.authKey
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

// DBConfig is the basic config object for db connections.
type DBConfig struct {
	Server   string
	DBName   string
	Schema   string
	User     string
	Password string

	DSN string
}

// InitFromEnvironment initializes the db config from environment variables.
func (db *DBConfig) InitFromEnvironment() {
	dsn := os.Getenv("DATABASE_URL")
	if len(dsn) != 0 {
		db.InitFromDSN(dsn)
	} else {
		if len(os.Getenv("DB_HOST")) > 0 {
			db.Server = os.Getenv("DB_HOST")
		} else {
			db.Server = "localhost"
		}
		db.DBName = os.Getenv("DB_NAME")
		db.Schema = os.Getenv("DB_SCHEMA")
		db.User = os.Getenv("DB_USER")
		db.Password = os.Getenv("DB_PASSWORD")
	}
}

// InitFromDSN initializes the db config from a dsn.
func (db *DBConfig) InitFromDSN(dsn string) {
	db.DSN = dsn
}

// GetDSN returns the config as a postgres dsn.
func (db DBConfig) GetDSN() string {
	if len(db.DSN) > 0 {
		return db.DSN
	}
	if len(db.Password) > 0 {
		return fmt.Sprintf("postgres://%s:%s@%s/%s?sslmode=disable", db.User, db.Password, db.Server, db.DBName)
	}
	if len(db.User) > 0 {
		return fmt.Sprintf("postgres://%s@%s/%s?sslmode=disable", db.User, db.Server, db.DBName)
	}
	return fmt.Sprintf("postgres://%s/%s?sslmode=disable", db.Server, db.DBName)
}

// SetupDatabaseContext writes the config to spiffy.
func SetupDatabaseContext(config *DBConfig) error {
	spiffy.CreateDbAlias("main", spiffy.NewDbConnectionFromDSN(config.GetDSN()))
	spiffy.SetDefaultAlias("main")

	_, dbError := spiffy.DefaultDb().Open()
	if dbError != nil {
		return dbError
	}

	spiffy.DefaultDb().Connection.SetMaxIdleConns(50)
	return nil
}

// DBInit reads the config from the environment and sets up spiffy.
func DBInit() error {
	dbConfig := &DBConfig{}
	dbConfig.InitFromEnvironment()
	return SetupDatabaseContext(dbConfig)
}
