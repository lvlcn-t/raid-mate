package database

import (
	"database/sql"
	"errors"
	"fmt"

	_ "github.com/lib/pq"
)

// Config represents the database configuration.
type Config struct {
	// Host is the database host.
	Host string `yaml:"host" mapstructure:"host"`
	// Port is the database port.
	Port uint16 `yaml:"port" mapstructure:"port"`
	// Name is the name of the database.
	Name string `yaml:"name" mapstructure:"name"`
	// User is the user to connect to the database.
	User string `yaml:"user" mapstructure:"user"`
	// Password is the password of the user.
	Password string `yaml:"password" mapstructure:"password"`
}

func (c *Config) String() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Password, c.Name)
}

func (c *Config) Validate() error {
	var err error
	if c.Host == "" {
		err = errors.New("database.host is required")
	}
	if c.Port == 0 {
		err = errors.Join(err, errors.New("database.port is required"))
	}
	if c.Name == "" {
		err = errors.Join(err, errors.New("database.name is required"))
	}
	if c.User == "" {
		err = errors.Join(err, errors.New("database.user is required"))
	}
	if c.Password == "" {
		err = errors.Join(err, errors.New("database.password is required"))
	}
	return err
}

// New creates a new database connection.
func New(cfg *Config) (*sql.DB, error) {
	return sql.Open("postgres", cfg.String())
}
