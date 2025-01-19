package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

// Config represents the database configuration.
type Config struct {
	// Host is the database host.
	Host string `yaml:"host" mapstructure:"host" validate:"required"`
	// Port is the database port.
	Port uint16 `yaml:"port" mapstructure:"port" validate:"required"`
	// Name is the name of the database.
	Name string `yaml:"name" mapstructure:"name" validate:"required"`
	// User is the user to connect to the database.
	User string `yaml:"user" mapstructure:"user" validate:"required"`
	// Password is the password of the user.
	Password string `yaml:"password" mapstructure:"password" validate:"required"`
}

func (c *Config) String() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		c.Host, c.Port, c.User, c.Password, c.Name)
}

// New creates a new database connection.
func New(cfg *Config) (*sql.DB, error) {
	return sql.Open("postgres", cfg.String())
}
