// ml_i - an ML/I macro processor ported to Go
// Copyright (c) 2023 Michael D Henderson.
// All rights reserved.

// Package main implements a LOWL assembler.
package main

import (
	"bytes"
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
		fmt.Printf("\n\nerror:\n%v\n\n", err)
		log.Fatal("error")
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
	} else if err = syntaxTree.Listing("ast_listing.txt"); err != nil {
		return err
	}

	vm, err := assembler.Assemble(syntaxTree)
	if err != nil {
		return err
	}

	stdout, stdmsg := &bytes.Buffer{}, &bytes.Buffer{}
	err = vm.Run(stdout, stdmsg)
	_ = os.WriteFile("vm_stdout.txt", stdout.Bytes(), 0644)
	_ = os.WriteFile("vm_stdmsg.txt", stdmsg.Bytes(), 0644)

	return err
}
