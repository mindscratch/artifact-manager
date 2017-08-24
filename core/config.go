package core

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config stores application configuration
type Config struct {
	// address to listen on
	Addr string
	// enable debug logging
	Debug bool
	// the directory used for managing files
	Dir string
	// a prefix used for all app specific environment variables
	EnvVarPrefix string
        // if the application is run inside a container, the external directory would be the location on
        // the host.
        ExternalDir string
	// the name of the host the application is running on
	Hostname string
	// enable debugging by the go-marathon library
	MarathonDebug bool
	// Marathon hosts to interact with, can be one or more "host:port" separated by commas
	MarathonHosts string
	// the period in-between querying marathon
	MarathonQueryInterval time.Duration
	// port to listen on
	Port int
}

// NewConfig creates and returns a new Config.
func NewConfig(envVarPrefix string) *Config {
	c := Config{
		Addr:                  "",
		Debug:                 false,
		Dir:                   "/tmp",
		EnvVarPrefix:          envVarPrefix,
                ExternalDir:           "/tmp",
		MarathonDebug:         false,
		MarathonHosts:         "localhost:8080",
		MarathonQueryInterval: 10 * time.Second,
		Port: 8900,
	}
	if flag.Lookup("addr") == nil {
		flag.StringVar(&c.Addr, "addr", c.Addr, "address to listen on")
	}
	if flag.Lookup("debug") == nil {
		flag.BoolVar(&c.Debug, "debug", c.Debug, "enable debug logging")
	}
	if flag.Lookup("dir") == nil {
		flag.StringVar(&c.Dir, "dir", c.Dir, "directory where files will be managed")
	}
	if flag.Lookup("external-dir") == nil {
		flag.StringVar(&c.ExternalDir, "external-dir", c.ExternalDir, "if running in a container, this is the directory on the host that maps to `dir` inside the container")
	}
	if flag.Lookup("marathon-debug") == nil {
		flag.BoolVar(&c.MarathonDebug, "marathon-debug", c.MarathonDebug, "enable go-marathon library debug logging")
	}
	if flag.Lookup("marathon-hosts") == nil {
		flag.StringVar(&c.MarathonHosts, "marathon-hosts", c.MarathonHosts, "comma-delimited list of marathon hosts, \"host:port\"")
	}
	if flag.Lookup("marathon-query-interval") == nil {
		flag.DurationVar(&c.MarathonQueryInterval, "marathon-query-interval", c.MarathonQueryInterval, "time to wait between queries to marathon")
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

	key = c.EnvVarPrefix + "DEBUG"
	val = os.Getenv(key)
	if strings.HasPrefix(strings.ToLower(val), "t") {
		c.Debug = true
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

	key = c.EnvVarPrefix + "EXTERNAL_DIR"
	val = os.Getenv(key)
	if val != "" {
		c.ExternalDir = os.Getenv(key)
	} else {
                // if no value is specified, default to being equal to `dir`
                c.ExternalDir = c.Dir
        }

	key = c.EnvVarPrefix + "MARATHON_DEBUG"
	val = os.Getenv(key)
	if strings.HasPrefix(strings.ToLower(val), "t") {
		c.MarathonDebug = true
	}

	key = c.EnvVarPrefix + "MARATHON_HOSTS"
	val = os.Getenv(key)
	if val != "" {
		c.MarathonHosts = os.Getenv(key)
	}

	key = c.EnvVarPrefix + "MARATHON_QUERY_INTERVAL"
	val = os.Getenv(key)
	if val != "" {
		d, err := time.ParseDuration(val)
		if err != nil {
			return fmt.Errorf("marathon-query-interval=%v is not a valid duration: %v", val, err)
		}
		c.MarathonQueryInterval = d
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
	fmt.Fprintf(os.Stderr, "The variables are equivalent to the command-line flag names, except that they should be upper-case, hypens replaced by underscores and prefixed with \"%s\" (excluding double quotes)\n", c.EnvVarPrefix)
}

// ServeAddr returns the address the server should listen on.
func (c *Config) ServeAddr() string {
	return fmt.Sprintf("%s:%d", c.Addr, c.Port)
}
