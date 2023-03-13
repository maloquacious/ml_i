package main

import (
	"flag"
	"fmt"
	"github.com/peterbourgon/ff/v3"
	"os"
)

type config struct {
	version    string
	debug      bool
	sourcefile string
}

func getConfig() (*config, error) {
	// create the config structure with default values
	cfg := &config{
		version: "L4A",
	}

	// create a flag set and then parse the command line (and optional configuration file)
	fs := flag.NewFlagSet("my-program", flag.ContinueOnError)
	var (
		_ = fs.String("config", "", "config file (optional, json)")
	)
	fs.BoolVar(&cfg.debug, "debug", cfg.debug, "log debug information (optional)")
	fs.StringVar(&cfg.sourcefile, "source", cfg.sourcefile, "assembly source file (required)")
	if err := ff.Parse(fs, os.Args[1:], ff.WithEnvVarPrefix("LASM"), ff.WithConfigFileFlag("config"), ff.WithConfigFileParser(ff.JSONParser), ff.WithIgnoreUndefined(false)); err != nil {
		return nil, err
	} else if cfg.sourcefile == "" {
		return nil, fmt.Errorf("--source is required")
	}

	return cfg, nil
}
