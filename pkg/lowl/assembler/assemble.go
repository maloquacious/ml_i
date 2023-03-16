// ml_i - an ML/I macro processor ported to Go
// Copyright (c) 2023 Michael D Henderson.
// All rights reserved.

// Package assembler assembles the instructions and returns a VM that can run them.
package assembler

import (
	"fmt"
	"github.com/maloquacious/ml_i/pkg/lowl/ast"
	"github.com/maloquacious/ml_i/pkg/lowl/op"
	"github.com/maloquacious/ml_i/pkg/lowl/vm"
)

func Assemble(nodes ast.Nodes) (*vm.VM, error) {
	// create symbol table and initialize it with required constants
	symtab := newSymbolTable(nodes)
	symtab.InsertConstant("DSTPT", 0)     // destination field pointer (stack moves)
	symtab.InsertConstant("LCH", 1)       // LCH is the length (in words) of a character
	symtab.InsertConstant("LNM", 1)       // LMN is the length (in words) of a number
	symtab.InsertConstant("LICH", 1)      // LICH is the inverse of LCH
	symtab.InsertConstant("NLREP", '\n')  // new-line
	symtab.InsertConstant("QUTREP", '"')  // quote mark
	symtab.InsertConstant("SPREP", ' ')   // space
	symtab.InsertConstant("SRCPT", 0)     // source field pointer (stack moves)
	symtab.InsertConstant("TABREP", '\t') // tab

	machine := &vm.VM{}
	// first instruction should be a halt. when we start running the machine,
	// the PC will be set to the first instruction in the program.
	machine.Core[machine.PC], machine.PC = vm.Word{Op: op.HALT}, machine.PC+1

	// assemble all the instructions
	for _, node := range nodes {
		// provide a default word for the instruction
		word := vm.Word{Op: node.Op} // default word to the current opcode

		// emit the word
		switch node.Op {
		// this section implements op codes that have no arguments

		// this section implements op codes that have 1 argument that must be a constant
		case op.ANDL: // bit-wise AND of register A with constant or literal value
			if minArgs := 1; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			// operand must be a constant
			switch constant := node.Parameters[0]; constant.Kind {
			case ast.Number:
				word.Value = constant.Number
			default:
				return nil, fmt.Errorf("%d: %s: %s not allowed", node.Line, node.Op, constant.Kind)
			}
			machine.Core[machine.PC], machine.PC = word, machine.PC+1
		case op.CCN: // compare register C with named character
			if minArgs := 1; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			// operand must be a constant
			switch constant := node.Parameters[0]; constant.Kind {
			case ast.Number:
				word.Value = constant.Number
			case ast.Variable:
				sym, ok := symtab.Lookup(constant.Text)
				if !ok {
					return nil, fmt.Errorf("%d: %s: %s %q: forward declared not implemented", node.Line, node.Op, constant.Kind, constant.Text)
				}
				switch sym.kind {
				case "constant":
					word.Value = sym.constant
				default:
					return nil, fmt.Errorf("%d: %s: %s: must be constant", node.Line, node.Op, constant.Text)
				}
			default:
				return nil, fmt.Errorf("%d: %s: %s not allowed", node.Line, node.Op, constant.Kind)
			}
			machine.Core[machine.PC], machine.PC = word, machine.PC+1
		case op.LCN: // load named character into register C
			if minArgs := 1; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			// operand must be a constant
			switch constant := node.Parameters[0]; constant.Kind {
			case ast.Number:
				word.Value = constant.Number
			case ast.Variable:
				sym, ok := symtab.Lookup(constant.Text)
				if !ok {
					return nil, fmt.Errorf("%d: %s: %s %q: forward declared not implemented", node.Line, node.Op, constant.Kind, constant.Text)
				}
				switch sym.kind {
				case "constant":
					word.Value = sym.constant
				default:
					return nil, fmt.Errorf("%d: %s: %s: must be constant", node.Line, node.Op, constant.Text)
				}
			default:
				return nil, fmt.Errorf("%d: %s: %s not allowed", node.Line, node.Op, constant.Kind)
			}
			machine.Core[machine.PC], machine.PC = word, machine.PC+1
		case op.NCH: // allocate memory and initialize it with a named character
			word.Op = op.HALT // non-executable instruction!
			if minArgs := 1; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			// operand must be a constant
			switch constant := node.Parameters[0]; constant.Kind {
			case ast.Number:
				word.Value = constant.Number
			case ast.Variable:
				sym, ok := symtab.Lookup(constant.Text)
				if !ok {
					return nil, fmt.Errorf("%d: %s: %s %q: forward declared not implemented", node.Line, node.Op, constant.Kind, constant.Text)
				}
				switch sym.kind {
				case "constant":
					word.Value = sym.constant
				default:
					return nil, fmt.Errorf("%d: %s: %s: must be constant", node.Line, node.Op, constant.Text)
				}
			default:
				return nil, fmt.Errorf("%d: %s: %s not allowed", node.Line, node.Op, constant.Kind)
			}
			machine.Core[machine.PC], machine.PC = word, machine.PC+1

		// this section implements op codes that have one argument which must be n-of
		case op.AAL: // add n-of value to register A
			if minArgs := 1; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			nOF := node.Parameters[0]
			switch nOF.Kind {
			case ast.Macro:
				value, err := evalMacro(nOF.Text, node.Parameters[1:], symtab.GetEnv())
				if err != nil {
					return nil, fmt.Errorf("%d: %s: %s %s: %w", node.Line, node.Op, nOF.Kind, nOF.Text, err)
				}
				word.Value = value
			case ast.Number:
				word.Value = nOF.Number
			default:
				return nil, fmt.Errorf("%d: %s: %s not allowed", node.Line, node.Op, nOF.Kind)
			}
			machine.Core[machine.PC], machine.PC = word, machine.PC+1
		case op.CAL: // compare register A with n-of
			if minArgs := 1; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			nOF := node.Parameters[0]
			switch nOF.Kind {
			case ast.Macro:
				value, err := evalMacro(nOF.Text, node.Parameters[1:], symtab.GetEnv())
				if err != nil {
					return nil, fmt.Errorf("%d: %s: %s %s: %w", node.Line, node.Op, nOF.Kind, nOF.Text, err)
				}
				word.Value = value
			case ast.Number:
				word.Value = nOF.Number
			case ast.Variable:
				// variable must be a constant
				if sym, ok := symtab.Lookup(nOF.Text); !ok {
					return nil, fmt.Errorf("%d: %s: %s: forward declaration not allowed", node.Line, node.Op, nOF.Text)
				} else if sym.kind != "constant" {
					return nil, fmt.Errorf("%d: %s: %s: must be constant", node.Line, node.Op, nOF.Text)
				} else {
					word.Value = sym.constant
				}
			default:
				return nil, fmt.Errorf("%d: %s: %s not allowed", node.Line, node.Op, nOF.Kind)
			}
			machine.Core[machine.PC], machine.PC = word, machine.PC+1
		case op.LAL, op.SAL:
			// LAL load register A with n-of
			// SAL subtract n-of from register A
			if minArgs := 1; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			nOF := node.Parameters[0]
			switch nOF.Kind {
			case ast.Macro:
				value, err := evalMacro(nOF.Text, node.Parameters[1:], symtab.GetEnv())
				if err != nil {
					return nil, fmt.Errorf("%d: %s: %s %s: %w", node.Line, node.Op, nOF.Kind, nOF.Text, err)
				}
				word.Value = value
			case ast.Number:
				word.Value = nOF.Number
			default:
				return nil, fmt.Errorf("%d: %s: %s: not allowed", node.Line, node.Op, nOF.Kind)
			}
			machine.Core[machine.PC], machine.PC = word, machine.PC+1

		// all other op codes
		case op.AAV: // add variable V to register A
			if minArgs := 1; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			v := node.Parameters[0]
			switch v.Kind {
			case ast.Variable:
				symtab.UpdateAddress(v.Text, machine.PC)
			default:
				return nil, fmt.Errorf("%d: %s: %s: not implemented", node.Line, node.Op, v.Kind)
			}
			machine.Core[machine.PC], machine.PC = word, machine.PC+1
		case op.ANDV: // bit-wise AND of register A with variable V
			if minArgs := 1; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			v := node.Parameters[0]
			switch v.Kind {
			case ast.Variable:
				symtab.UpdateAddress(v.Text, machine.PC)
			default:
				return nil, fmt.Errorf("%d: %s: %s: not implemented", node.Line, node.Op, v.Kind)
			}
			machine.Core[machine.PC], machine.PC = word, machine.PC+1
		case op.BMOVE: // backwards block move
			machine.Core[machine.PC], machine.PC = word, machine.PC+1
		case op.BSTK: // push register A onto stack BS
			machine.Core[machine.PC], machine.PC = word, machine.PC+1
		case op.BUMP: // add a literal or expression to a variable V
			if minArgs := 2; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			v, nOF := node.Parameters[0], node.Parameters[1]
			switch v.Kind {
			case ast.Variable:
				symtab.UpdateAddress(v.Text, machine.PC)
			default:
				return nil, fmt.Errorf("%d: %s: %s: not implemented", node.Line, node.Op, nOF.Kind)
			}
			switch nOF.Kind {
			case ast.Number:
				word.Value = nOF.Number
			default:
				return nil, fmt.Errorf("%d: %s: %s: not implemented", node.Line, node.Op, nOF.Kind)
			}
			machine.Core[machine.PC], machine.PC = word, machine.PC+1
		case op.CAV: // compare register A with variable V
			if minArgs := 1; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			v, axFlag := node.Parameters[0], node.Parameters[1].Text
			switch axFlag {
			case "A": // compare unsigned addresses
				// no special action needed
			case "X": // compare signed numbers
				// no special action needed
			default:
				return nil, fmt.Errorf("%d: %s: axFlag want A|X: got %q", node.Line, node.Op, axFlag)
			}
			switch v.Kind {
			case ast.Variable:
				symtab.UpdateAddress(v.Text, machine.PC)
			default:
				return nil, fmt.Errorf("%d: %s: %s: not implemented", node.Line, node.Op, v.Kind)
			}
			machine.Core[machine.PC], machine.PC = word, machine.PC+1
		case op.CLEAR: // set variable V to zero
			if minArgs := 1; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			v := node.Parameters[0]
			switch v.Kind {
			case ast.Variable:
				symtab.UpdateAddress(v.Text, machine.PC)
			default:
				return nil, fmt.Errorf("%d: %s: %s: not implemented", node.Line, node.Op, v.Kind)
			}
			machine.Core[machine.PC], machine.PC = word, machine.PC+1
		case op.CON: // allocate memory and initialize it
			word.Op = op.HALT // non-executable instruction!
			if minArgs := 1; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			// operand will be either a constant or the value of an expression
			operand := node.Parameters[0]
			switch operand.Kind {
			case ast.Number:
				word.Value = operand.Number
			default:
				return nil, fmt.Errorf("%d: %s: %s not implemented", node.Line, node.Op, operand.Kind)
			}
			machine.Core[machine.PC], machine.PC = word, machine.PC+1
		case op.CSS: // pop the return stack
			machine.Core[machine.PC], machine.PC = word, machine.PC+1
		case op.CFSTK: // push register C onto stack FS
			machine.Core[machine.PC], machine.PC = word, machine.PC+1
		case op.DCL: // allocate memory, create a variable, link them
			if minArgs := 1; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			name, addr := node.Parameters[0].Text, machine.PC
			if _, ok := symtab.Lookup(name); !ok {
				symtab.InsertAddress(name, addr)
			}
			symtab.UpdateAddress(name, machine.PC)
			machine.Core[machine.PC], machine.PC = vm.Word{}, machine.PC+1
		case op.EQU:
			if minArgs := 2; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			// create an alias to another variable
			name, alias := node.Parameters[0].Text, node.Parameters[1].Text
			symtab.InsertAlias(name, alias)
		case op.FMOVE: // forwards block move
			machine.Core[machine.PC], machine.PC = word, machine.PC+1
		case op.FSTK: // push register A onto stack FS
			machine.Core[machine.PC], machine.PC = word, machine.PC+1
		case op.GO, op.GOEQ, op.GOGE, op.GOLE, op.GOLT, op.GONE, op.GOGR: // GOxxx label,distance,(E|X),(C|T|X)
			if minArgs := 4; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			label, _, exFlag, ctxFlag := node.Parameters[0].Text, node.Parameters[1].Number, node.Parameters[2].Text, node.Parameters[3].Text
			symtab.AddReference(label, machine.PC)
			switch exFlag {
			case "E": // branch out of subroutine
				// no special action needed
			case "X": // normal branch
				// no special action needed
			default:
				return nil, fmt.Errorf("%d: %s: exFlag want E|X: got %q", node.Line, node.Op, exFlag)
			}
			switch ctxFlag {
			case "C": // exit following gosub
				// no special action needed
			case "T": // GOADD branch
				if node.Op != op.GO {
					return nil, fmt.Errorf("%d: %s: T: not implemented", node.Line, node.Op)
				}
				word.Op = op.GOBRPC
			case "X": // nothing special
				// no special action needed
			default:
				return nil, fmt.Errorf("%d: %s: ctxFlag want C|T|X: got %q", node.Line, node.Op, ctxFlag)
			}
			machine.Core[machine.PC], machine.PC = word, machine.PC+1
		case op.GOADD: // multi-way branch
			// works by setting BranchPC. If GOxxx has the T flag and it's PC matches BranchPC, then the branch is taken.
			if minArgs := 1; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			v := node.Parameters[0]
			switch v.Kind {
			case ast.Variable:
				symtab.UpdateAddress(v.Text, machine.PC)
			default:
				return nil, fmt.Errorf("%d: %s: %s: not implemented", node.Line, node.Op, v.Kind)
			}
			machine.Core[machine.PC], machine.PC = word, machine.PC+1
		case op.GOBRPC:
			panic("GOBRPC is not available to clients")
		case op.GOSUB: // branch to a subroutine
			if minArgs := 2; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			// operands are name of subroutine and distance to jump.
			// we always ignore the distance.
			v, _ := node.Parameters[0], node.Parameters[1].Number
			switch v.Kind {
			case ast.Variable:
				symtab.UpdateAddress(v.Text, machine.PC)
			default:
				return nil, fmt.Errorf("%d: %s: %s: not implemented", node.Line, node.Op, v.Kind)
			}
		case op.IDENT: // create a variable, initialize it
			if minArgs := 2; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			v, constant := node.Parameters[0], node.Parameters[1]
			switch v.Kind {
			case ast.Variable:
				// nothing special
			default:
				return nil, fmt.Errorf("%d: %s: want variable: got %s", node.Line, node.Op, v.Kind)
			}
			switch constant.Kind {
			case ast.Number:
				// nothing special
			default:
				return nil, fmt.Errorf("%d: %s: want constant: got %s", node.Line, node.Op, constant.Kind)
			}
			symtab.InsertConstant(v.Text, constant.Number)
		case op.LAA: // load register A with address of variable V
			if minArgs := 2; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			v, cdFlag := node.Parameters[0], node.Parameters[1]
			switch v.Kind {
			case ast.Variable:
				symtab.AddReference(v.Text, machine.PC)
			default:
				return nil, fmt.Errorf("%d: %s: %s not implemented", node.Line, node.Op, v.Kind)
			}
			switch cdFlag.Kind {
			case ast.Variable:
				switch cdFlag.Text {
				case "C": // load A with the address of the table label
					// no special action needed
				case "D": // load A with the address of variable V
					// no special action needed
				default:
					return nil, fmt.Errorf("%d: %s: cdFlag want C|D: got %q", node.Line, node.Op, cdFlag.Text)
				}
			default:
				return nil, fmt.Errorf("%d: %s: %s not implemented", node.Line, node.Op, cdFlag.Kind)
			}
			machine.Core[machine.PC], machine.PC = word, machine.PC+1
		case op.LAI, op.LCI: // indirect load of register A|C
			if minArgs := 2; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			v, rxFlag := node.Parameters[0], node.Parameters[1]
			switch v.Kind {
			case ast.Variable:
				symtab.AddReference(v.Text, machine.PC)
			default:
				return nil, fmt.Errorf("%d: %s: %s not implemented", node.Line, node.Op, v.Kind)
			}
			switch rxFlag.Kind {
			case ast.Variable:
				switch rxFlag.Text {
				case "R": // load may be redundant
					// no special action needed
				case "X": // load is not redundant
					// no special action needed
				default:
					return nil, fmt.Errorf("%d: %s: rxFlag want R|X: got %q", node.Line, node.Op, rxFlag.Text)
				}
			default:
				return nil, fmt.Errorf("%d: %s: %s not implemented", node.Line, node.Op, rxFlag.Kind)
			}
			machine.Core[machine.PC], machine.PC = word, machine.PC+1
		case op.LAM: // load register A from register B modified by n-of
			if minArgs := 1; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			nOF := node.Parameters[0]
			switch nOF.Kind {
			case ast.Macro:
				value, err := evalMacro(nOF.Text, node.Parameters[1:], symtab.GetEnv())
				if err != nil {
					return nil, fmt.Errorf("%d: %s: %s %s: %w", node.Line, node.Op, nOF.Kind, nOF.Text, err)
				}
				word.Value = value
			case ast.Number:
				word.Value = nOF.Number
			default:
				return nil, fmt.Errorf("%d: %s: %s: not implemented", node.Line, node.Op, nOF.Kind)
			}
			machine.Core[machine.PC], machine.PC = word, machine.PC+1
		case op.LAV, op.LBV: // load register A|B from variable V
			if minArgs := 1; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			v := node.Parameters[0]
			switch v.Kind {
			case ast.Variable:
				symtab.AddReference(v.Text, machine.PC)
			default:
				return nil, fmt.Errorf("%d: %s: %s not implemented", node.Line, node.Op, v.Kind)
			}
			switch node.Op {
			case op.LAV:
				rxFlag := node.Parameters[1]
				switch rxFlag.Kind {
				case ast.Variable:
					switch rxFlag.Text {
					case "R": // load may be redundant
						// no special action needed
					case "X": // normal load
						// no special action needed
					default:
						return nil, fmt.Errorf("%d: %s: exFlag want E|X: got %q", node.Line, node.Op, rxFlag)
					}
				default:
					return nil, fmt.Errorf("%d: %s: %s not implemented", node.Line, node.Op, rxFlag.Kind)
				}
			default:
				// no special action needed
			}
			machine.Core[machine.PC], machine.PC = word, machine.PC+1
		case op.MDLABEL: // create a label and link it to the current PC
			name, addr := node.Parameters[0].Text, machine.PC
			if _, ok := symtab.Lookup(name); !ok {
				symtab.InsertAddress(name, addr)
			}
			symtab.UpdateAddress(name, addr)
		case op.MESS: // write message to output stream
			word.Text = node.Parameters[0].Text
			machine.Core[machine.PC], machine.PC = word, machine.PC+1
		case op.NB: // ignore comments
			// no special action needed
		case op.PRGST: // name the machine
			name := node.Parameters[0].Text
			machine.Name = name
		case op.SAV: // subtract variable V from register A
			if minArgs := 1; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			v := node.Parameters[0]
			switch v.Kind {
			case ast.Variable:
				symtab.AddReference(v.Text, machine.PC)
			default:
				return nil, fmt.Errorf("%d: %s: %s not implemented", node.Line, node.Op, v.Kind)
			}
			machine.Core[machine.PC], machine.PC = word, machine.PC+1
		case op.STR: // allocate memory and initialize it with a string
			if minArgs := 1; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			for _, ch := range node.Parameters[0].Text {
				word.Value = int(ch)
				machine.Core[machine.PC], machine.PC = word, machine.PC+1
			}
		case op.STV: // store register A in variable V
			if minArgs := 2; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			v, pxFlag := node.Parameters[0], node.Parameters[1].Text
			switch pxFlag {
			case "P": // must preserve register A
				// no special action needed
			case "X": // okay to clobber register A
				// no special action needed
			default:
				return nil, fmt.Errorf("%d: %s: pxFlag want P|X: got %q", node.Line, node.Op, pxFlag)
			}
			switch v.Kind {
			case ast.Variable:
				symtab.AddReference(v.Text, machine.PC)
			default:
				return nil, fmt.Errorf("%d: %s: %s not implemented", node.Line, node.Op, v.Kind)
			}
			machine.Core[machine.PC], machine.PC = word, machine.PC+1
		case op.SUBR:
			//if minArgs := 4; len(node.Parameters) < minArgs {
			//	return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			//}
			//name, pxFlag := node.Parameters[0], node.Parameters[1].Text
			//if pxFlag == "X" { // subroutine has no parameters
			//	// no special action needed
			//} else { // subroutine has a single parameter
			//	parnm := node.Parameters[1].Text
			//}
			//var numberOfExits int
			//if len(node.Parameters) > 2 {
			//	noe := node.Parameters[2]
			//	switch noe.Kind {
			//	case ast.Number:
			//		numberOfExits = noe.Number
			//	default:
			//		return nil, fmt.Errorf("%d: %s: %s: not implemented", node.Line, node.Op, noe.Kind)
			//	}
			//} else {
			//	// give us a sane count, okay
			//	numberOfExits = 1
			//}
			panic("!!sub-routine!!")
		case op.UNSTK: // pop stack BS and store in variable V
			if minArgs := 1; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			v := node.Parameters[0]
			switch v.Kind {
			case ast.Variable:
				symtab.AddReference(v.Text, machine.PC)
			default:
				return nil, fmt.Errorf("%d: %s: %s: not implemented", node.Line, node.Op, v.Kind)
			}
			machine.Core[machine.PC], machine.PC = word, machine.PC+1

		default:
			return nil, fmt.Errorf("%d: %s: not implemented", node.Line, node.Op)
		}
	}

	// when we start running the machine, the PC should be set to the first
	// instruction in the program. if there is no BEGIN label, the PC will
	// point to a HALT instruction.
	if sym, ok := symtab.Lookup("BEGIN"); !ok {
		fmt.Printf("asm: warning: BEGIN not set\n")
	} else {
		if sym.kind != "address" {
			panic("BEGIN must be a label")
		}
		fmt.Printf("asm: set vm begin   %-12s %6d\n", "", sym.address)
		machine.Core[0] = vm.Word{Op: op.GO, Value: sym.address}
	}

	// allocate storage for all symbols at the end of the program
	for _, node := range nodes {
		if node.Op == op.DCL { // not an allocation
			continue
		}
		name, address := node.Parameters[0].Text, node.Parameters[1].Number
		if address != -1 { // already allocated
			continue
		}
		fmt.Printf("asm: alloc   var    %-12s %6d\n", name, machine.PC)
		symtab.UpdateAddress(node.Parameters[0].Text, machine.PC)
		machine.PC++
	}
	fmt.Printf("asm: %8d words created\n", machine.PC)
	for _, sym := range symtab.symbols {
		fmt.Printf("var %-12s address %8d\n", sym.name, sym.address)
	}

	machine.Core[0] = vm.Word{Op: op.GO, Value: machine.PC}

	panic("ast.Assemble is not implemented!")
}
