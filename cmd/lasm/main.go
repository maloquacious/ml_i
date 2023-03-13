// Package main implements a LOWL assembler.
package main

import (
	"fmt"
	"github.com/maloquacious/ml_i/pkg/lowl"
	"log"
	"os"
)

func main() {
	cfg, err := getConfig()
	if err != nil {
		log.Fatal(err)
	}

	if err := run(cfg); err != nil {
		log.Fatal(err)
	}
}

func run(cfg *config) error {
	//commands, err := lowl.Parse(cfg.sourcefile)
	//if err != nil {
	//	return err
	//}

	vm, listing, err := lowl.Assemble(cfg.sourcefile)
	if err != nil {
		return fmt.Errorf("assemble: %w", err)
	} else if listing == nil {
		return fmt.Errorf("package is empty")
	} else if err = os.WriteFile("lowl_app.txt", listing.Bytes(), 0644); err != nil {
		return err
	}

	return vm.Run()
}
