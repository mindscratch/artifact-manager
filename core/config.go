package core

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

// Config stores application configuration
type Config struct {
	// the directory used for managing files
	Dir string
	// a prefix used for all app specific environment variables
	EnvVarPrefix string
	// the name of the host the application is running on
	Hostname string
	// address to listen on
	Addr string
	// port to listen on
	Port int
}

// NewConfig creates and returns a new Config.
func NewConfig(envVarPrefix string) *Config {
	c := Config{
		Dir:          "/tmp",
		EnvVarPrefix: envVarPrefix,
		Addr:         "",
		Port:         8900,
	}
	if flag.Lookup("addr") == nil {
		flag.StringVar(&c.Addr, "addr", c.Addr, "address to listen on")
	}
	if flag.Lookup("dir") == nil {
		flag.StringVar(&c.Dir, "dir", c.Dir, "directory where files will be managed")
	}
	if flag.Lookup("port") == nil {
		flag.IntVar(&c.Port, "port", c.Port, "port to listen on")
	}
	flag.Usage = c.Usage
	return &c
}

// Parse parses the command-line flags and checks for environment variable overrides.
func (c *Config) Parse() error {
	flag.Parse()

	key := c.EnvVarPrefix + "ADDR"
	val := os.Getenv(key)
	if val != "" {
		c.Addr = os.Getenv(key)
	}

	key = c.EnvVarPrefix + "DIR"
	val = os.Getenv(key)
	if val != "" {
		c.Dir = os.Getenv(key)
	}
	_, err := os.Stat(c.Dir)
	if os.IsNotExist(err) {
		return fmt.Errorf("directory=%s does not exist", c.Dir)
	}

	key = c.EnvVarPrefix + "PORT"
	val = os.Getenv(key)
	if val != "" {
		num, err := strconv.Atoi(val)
		if err != nil {
			return fmt.Errorf("port=%v is not a valid number", val)
		}
		c.Port = num
	}
	return nil
}

// Usage outputs how to use the application.
func (c *Config) Usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	flag.PrintDefaults()
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Note: environment variables can be defined to override any command-line flag.")
	fmt.Fprintf(os.Stderr, "The variables are equivalent to the command-line flag names, except that they should be upper-case and prefixed with \"%s\" (excluding double quotes)\n", c.EnvVarPrefix)
}

// ServeAddr returns the address the server should listen on.
func (c *Config) ServeAddr() string {
	return fmt.Sprintf("%s:%d", c.Addr, c.Port)
}
