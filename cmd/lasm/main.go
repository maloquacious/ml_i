// Package main implements a LOWL assembler.
package main

import (
	"fmt"
	"github.com/maloquacious/ml_i/pkg/lowl/assembler"
	"github.com/maloquacious/ml_i/pkg/lowl/ast"
	"github.com/maloquacious/ml_i/pkg/lowl/cst"
	"log"
	"os"
)

func main() {
	cfg, err := getConfig()
	if err != nil {
		log.Fatal(err)
	}

	if err := run(cfg); err != nil {
		fmt.Println("")
		fmt.Println("")
		log.Fatal(err)
	}
}

func run(cfg *config) error {
	parseTree, err := cst.Parse(cfg.sourcefile, false, cfg.test.scanner)
	if cfg.test.scanner || err != nil {
		return err
	}
	foundErrors := false
	for _, node := range parseTree {
		if node.Error != nil {
			fmt.Printf("%d:%d %+v\n", node.Line, node.Col, node.Error)
			foundErrors = true
		}
	}
	if foundErrors == true {
		return fmt.Errorf("found errors")
	}

	syntaxTree, err := ast.Parse(parseTree)
	if err != nil {
		return err
	} else if cfg.test.astParser {
		if err = os.WriteFile("ast.txt", syntaxTree.Listing(), 0644); err != nil {
			return err
		}
		return nil
	}
	vm, err := assembler.Assemble(syntaxTree)
	if err != nil {
		return err
	}
	fmt.Println(vm)

	//vm, listing, err := lowl.Assemble(cfg.sourcefile)
	//if err != nil {
	//	return fmt.Errorf("assemble: %w", err)
	//} else if listing == nil {
	//	return fmt.Errorf("package is empty")
	//} else if err = os.WriteFile("lowl_app.txt", listing.Bytes(), 0644); err != nil {
	//	return err
	//}

	return vm.Run()
}
