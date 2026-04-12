package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config holds all portwatch daemon configuration.
type Config struct {
	Interval  time.Duration `yaml:"interval"`
	LogFile   string        `yaml:"log_file"`
	AlertHook AlertHook     `yaml:"alert_hook"`
	ProcNet   ProcNet       `yaml:"proc_net"`
}

// AlertHook configures the outbound webhook for change notifications.
type AlertHook struct {
	URL     string            `yaml:"url"`
	Headers map[string]string `yaml:"headers"`
	Timeout time.Duration     `yaml:"timeout"`
}

// ProcNet configures which /proc/net files to scan.
type ProcNet struct {
	TCP  string `yaml:"tcp"`
	TCP6 string `yaml:"tcp6"`
	UDP  string `yaml:"udp"`
}

// Defaults returns a Config pre-populated with sensible defaults.
func Defaults() Config {
	return Config{
		Interval: 5 * time.Second,
		LogFile:  "",
		AlertHook: AlertHook{
			Timeout: 5 * time.Second,
		},
		ProcNet: ProcNet{
			TCP:  "/proc/net/tcp",
			TCP6: "/proc/net/tcp6",
			UDP:  "/proc/net/udp",
		},
	}
}

// Load reads a YAML config file from path and merges it over the defaults.
func Load(path string) (Config, error) {
	cfg := Defaults()

	f, err := os.Open(path)
	if err != nil {
		return cfg, err
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	decoder.KnownFields(true)
	if err := decoder.Decode(&cfg); err != nil {
		return cfg, err
	}

	return cfg, nil
}
