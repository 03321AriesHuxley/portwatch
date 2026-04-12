package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/user/portwatch/internal/config"
)

const defaultConfigPath = "/etc/portwatch/config.yaml"

func main() {
	cfgPath := flag.String("config", defaultConfigPath, "path to portwatch YAML config file")
	validateOnly := flag.Bool("validate", false, "validate config and exit")
	flag.Parse()

	cfg, err := loadConfig(*cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "portwatch: %v\n", err)
		os.Exit(1)
	}

	if *validateOnly {
		fmt.Println("portwatch: config is valid")
		os.Exit(0)
	}

	log.Printf("portwatch starting (interval=%s, hook=%q)",
		cfg.Interval, cfg.AlertHook.URL)

	// TODO: wire scanner + alert loop
	select {}
}

// loadConfig reads and validates the configuration, falling back to defaults
// when the config file does not exist.
func loadConfig(path string) (config.Config, error) {
	cfg, err := config.Load(path)
	if err != nil {
		if os.IsNotExist(err) {
			log.Printf("portwatch: config file %q not found, using defaults", path)
			cfg = config.Defaults()
		} else {
			return config.Config{}, fmt.Errorf("loading config: %w", err)
		}
	}

	if err := config.Validate(cfg); err != nil {
		return config.Config{}, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}
