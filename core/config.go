package core

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

// Config stores application configuration
type Config struct {
	// a prefix used for all app specific environment variables
	envVarPrefix string
	// the name of the host the application is running on
	hostname string
	// address to listen on
	listenAddr string
	// port to listen on
	port int
}

// NewConfig creates and returns a new Config.
func NewConfig(envVarPrefix string) *Config {
	c := Config{
		envVarPrefix: envVarPrefix,
		listenAddr:   "",
		port:         8900,
	}
	flag.StringVar(&c.listenAddr, "addr", c.listenAddr, "address to listen on")
	flag.IntVar(&c.port, "port", c.port, "port to listen on")
	flag.Usage = c.Usage
	return &c
}

// Parse parses the command-line flags and checks for environment variable overrides.
func (c *Config) Parse() error {
	flag.Parse()

	key := c.envVarPrefix + "ADDR"
	val := os.Getenv(key)
	if val != "" {
		c.listenAddr = os.Getenv(key)
	}

	key = c.envVarPrefix + "PORT"
	val = os.Getenv(key)
	if val != "" {
		num, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("port=%v is not a valid number", val)
		}
		c.port = num
	}
	return nil
}

// Usage outputs how to use the application.
func (c *Config) Usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Note: environment variables can be defined to override any command-line flag.")
	fmt.Fprintf(os.Stderr, "The variables are equivalent to the command-line flag names, except that they should be upper-case and prefixed with \"%s\" (excluding double quotes)\n", c.envVarPrefix)
}
