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
	symtab.InsertConstant("LCH", 1)       // LCH is the length (in words) of a character
	symtab.InsertConstant("LNM", 1)       // LMN is the length (in words) of a number
	symtab.InsertConstant("LICH", 1)      // LICH is the inverse of LCH
	symtab.InsertConstant("NLREP", '\n')  // new-line
	symtab.InsertConstant("QUTREP", '"')  // quote mark
	symtab.InsertConstant("SPREP", ' ')   // space
	symtab.InsertConstant("TABREP", '\t') // tab

	machine := &vm.VM{}
	// first instruction should be a halt. when we start running the machine,
	// the PC will be set to the first instruction in the program.
	machine.Core[machine.PC], machine.PC = vm.Word{Op: op.HALT}, machine.PC+1

	// insert variables required by the LOWL specifications
	for _, lvar := range []struct {
		name  string
		value int
	}{
		{"DSTPT", 0}, // destination field pointer (stack moves)
		{"PARNM", 0}, // subroutine named parameter
		{"SRCPT", 0}, // source field pointer (stack moves)
	} {
		if ok := symtab.InsertAddress(lvar.name, machine.PC); !ok {
			return nil, fmt.Errorf("%d: setup: internal error: %s redefined", 0, lvar.name)
		}
		machine.Core[machine.PC], machine.PC = vm.Word{Value: lvar.value}, machine.PC+1
	}

	// the current subroutine name is set whenever we get a SUBR instruction.
	// it is used as a sanity check in the EXIT calls
	var currSubroutine struct {
		name          string
		numberOfExits int
	}

	// assemble all the instructions
	foundPrgst, foundPrgen := false, false
	for _, node := range nodes {
		if foundPrgen {
			fmt.Printf("asm: %d: %d: %q after PRGEN\n", node.Line, node.Col, node.Op)
		}
		// provide a default word for the instruction
		word := vm.Word{Op: node.Op} // default word to the current opcode

		// emit the word
		switch node.Op {
		//
		//
		// this section implements op codes that have no arguments
		//
		case op.ALIGN:
			// ALIGN  no code emitted for this op code
		case op.BMOVE, op.BSTK, op.CFSTK, op.FMOVE, op.FSTK:
			// BMOVE backwards block move
			// BSTK  push register A onto stack BS
			// CSTK  push register C onto stack FS
			// FMOVE forwards block move
			// FSTK  ush register A onto stack FS
			machine.Core[machine.PC], machine.PC = word, machine.PC+1
		case op.GOBRPC:
			// GOBRPC is not available to callers
			return nil, fmt.Errorf("%d: %d: %s: internal error", node.Line, node.Col, node.Op)
		case op.PRGEN:
			foundPrgen = true
		case op.UNKNOWN:
			return nil, fmt.Errorf("%d: %d: %s: internal error", node.Line, node.Col, node.Op)

		//
		//
		// this section implements op codes that have 1 argument that must be a constant
		//
		//
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
					return nil, fmt.Errorf("%d: %s: %s %q: forward declaration not allowed here", node.Line, node.Op, constant.Kind, constant.Text)
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
					return nil, fmt.Errorf("%d: %s: %s %q: forward declaration not allowed here", node.Line, node.Op, constant.Kind, constant.Text)
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
					return nil, fmt.Errorf("%d: %s: %s %q: forward declaration not allowed here", node.Line, node.Op, constant.Kind, constant.Text)
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

		//
		//
		// this section implements op codes that have one argument which must be n-of
		//   AAL   add n-of value to register A
		//   CAL   compare register A with n-of
		//   LAL   load register A with n-of
		//   LAM   load register A from register B modified by n-of
		//   LCM   load register C from register B modified by n-of
		//   MULTL multiply register A by n-of
		//   SAL   subtract n-of from register A
		//   SBL   subtract n-of value from register B
		case op.AAL, op.CAL, op.LAL, op.LAM, op.LCM, op.MULTL, op.SAL, op.SBL:
			if minArgs := 1; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			switch nOF := node.Parameters[0]; nOF.Kind {
			case ast.Macro:
				if minArgs := 2; len(node.Parameters) < minArgs {
					return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
				}
				expr := node.Parameters[1]
				value, err := evalMacro(nOF.Text, expr, symtab.GetEnv())
				if err != nil {
					return nil, fmt.Errorf("%d: %s: %s %s: %w", node.Line, node.Op, nOF.Kind, nOF.Text, err)
				}
				word.Value = value
			case ast.Number:
				word.Value = nOF.Number
			case ast.Variable:
				// variable must be a constant
				sym, ok := symtab.Lookup(nOF.Text)
				if !ok {
					return nil, fmt.Errorf("%d: %s: %s %q: forward declaration not allowed here", node.Line, node.Op, nOF.Kind, nOF.Text)
				}
				switch sym.kind {
				case "constant":
					word.Value = sym.constant
				default:
					return nil, fmt.Errorf("%d: %s: %s: must be constant", node.Line, node.Op, nOF.Text)
				}
			default:
				return nil, fmt.Errorf("%d: %s: %s not allowed", node.Line, node.Op, nOF.Kind)
			}
			machine.Core[machine.PC], machine.PC = word, machine.PC+1

		//
		//
		// this section implements op codes that have one argument which must be a variable
		//   AAV   add variable V to register A
		//   ABV   add variable V to register B
		//   ANDV  bit-wise AND of register A with variable V
		//   CCI   compare register C with contents of address pointed to by V
		//   SAV   subtract variable V from register A
		//   SBV   subtract variable V from register B
		//   UNSTK pop stack BS and store in variable V
		case op.AAV, op.ABV, op.ANDV, op.CCI, op.SAV, op.SBV, op.UNSTK:
			if minArgs := 1; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			switch v := node.Parameters[0]; v.Kind {
			case ast.Variable:
				symtab.UpdateAddress(v.Text, machine.PC)
			default:
				return nil, fmt.Errorf("%d: %s: %s: not allowed", node.Line, node.Op, v.Kind)
			}
			machine.Core[machine.PC], machine.PC = word, machine.PC+1

		//
		//
		// this section implements op codes that require a variable and an A|X flag
		//   CAI indirect compare of register A with address pointed to by variable V
		//   CAV compare register A with variable V
		case op.CAI, op.CAV:
			if minArgs := 1; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			switch v := node.Parameters[0]; v.Kind {
			case ast.Variable:
				symtab.UpdateAddress(v.Text, machine.PC)
			default:
				return nil, fmt.Errorf("%d: %s: %s: not allowed", node.Line, node.Op, v.Kind)
			}
			switch flag := node.Parameters[1]; flag.Kind {
			case ast.Variable:
				switch flag.Text {
				case "A": // compare unsigned addresses
					// no special action needed
				case "X": // compare signed numbers
					// no special action needed
				default:
					return nil, fmt.Errorf("%d: %s: axFlag want A|X: got %q", node.Line, node.Op, flag)
				}
			default:
				return nil, fmt.Errorf("%d: %s: %s not allowed", node.Line, node.Op, flag.Kind)
			}
			machine.Core[machine.PC], machine.PC = word, machine.PC+1

		//
		//
		// this section implements op codes that require a label spec
		//   GOxxx label,distance,(E|X),(C|T|X)
		case op.GO, op.GOEQ, op.GOGE, op.GOLE, op.GOLT, op.GOND, op.GONE, op.GOGR, op.GOPC:
			if minArgs := 4; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			switch label := node.Parameters[0]; label.Kind {
			case ast.Variable:
				symtab.AddReference(label.Text, machine.PC)
			default:
				return nil, fmt.Errorf("%d: %s: %s not allowed", node.Line, node.Op, label.Kind)
			}
			switch distance := node.Parameters[1]; distance.Kind {
			case ast.Number:
				// no special action needed
			default:
				return nil, fmt.Errorf("%d: %s: %s not allowed", node.Line, node.Op, distance.Kind)
			}
			switch flag := node.Parameters[2]; flag.Kind {
			case ast.Variable:
				switch flag.Text {
				case "E": // branch out of subroutine
					// no special action needed
				case "X": // normal branch
					// no special action needed
				default:
					return nil, fmt.Errorf("%d: %s: flag wants E|X: got %q", node.Line, node.Op, flag.Text)
				}
			}
			switch flag := node.Parameters[3]; flag.Kind {
			case ast.Variable:
				switch flag.Text {
				case "C": // exit following gosub
					// no special action needed
				case "T": // GOADD branch
					if node.Op != op.GO {
						return nil, fmt.Errorf("%d: %s: T: not allowed", node.Line, node.Op)
					}
					word.Op = op.GOBRPC
				case "X": // nothing special
					// no special action needed
				default:
					return nil, fmt.Errorf("%d: %s: flag wants C|T|X: got %q", node.Line, node.Op, flag.Text)
				}
			}
			machine.Core[machine.PC], machine.PC = word, machine.PC+1

		//
		//
		// this section implements op codes that require a variable and an R|X flag
		//   LAI indirect load of register A
		//   LCI indirect load of register C
		case op.LAI, op.LCI:
			if minArgs := 2; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			switch v := node.Parameters[0]; v.Kind {
			case ast.Variable:
				symtab.AddReference(v.Text, machine.PC)
			default:
				return nil, fmt.Errorf("%d: %s: %s not allowed", node.Line, node.Op, v.Kind)
			}
			switch flag := node.Parameters[1]; flag.Kind {
			case ast.Variable:
				switch flag.Text {
				case "R": // load may be redundant
					// no special action needed
				case "X": // load is not redundant
					// no special action needed
				default:
					return nil, fmt.Errorf("%d: %s: rxFlag want R|X: got %q", node.Line, node.Op, flag.Text)
				}
			default:
				return nil, fmt.Errorf("%d: %s: %s not allowed", node.Line, node.Op, flag.Kind)
			}
			machine.Core[machine.PC], machine.PC = word, machine.PC+1

		//
		//
		// all other op codes
		//
		//
		case op.BUMP: // add a literal or expression to a variable V
			if minArgs := 2; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			switch v := node.Parameters[0]; v.Kind {
			case ast.Variable:
				symtab.UpdateAddress(v.Text, machine.PC)
			default:
				return nil, fmt.Errorf("%d: %s: %s: not allowed", node.Line, node.Op, v.Kind)
			}
			switch nOF := node.Parameters[1]; nOF.Kind {
			case ast.Macro:
				if minArgs := 3; len(node.Parameters) < minArgs {
					return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
				}
				expr := node.Parameters[2]
				value, err := evalMacro(nOF.Text, expr, symtab.GetEnv())
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
		case op.CCL: // compare register C with literal
			if minArgs := 1; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			switch ch := node.Parameters[0]; ch.Kind {
			case ast.QuotedText:
				if len(ch.Text) != 1 {
					return nil, fmt.Errorf("%d: %s: want single character: got %q", node.Line, node.Op, ch.Text)
				}
				word.Value = int(ch.Text[0])
			default:
				return nil, fmt.Errorf("%d: %s: %s: not allowed", node.Line, node.Op, ch.Kind)
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
				return nil, fmt.Errorf("%d: %s: %s: not allowed", node.Line, node.Op, v.Kind)
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
				return nil, fmt.Errorf("%d: %s: %s not allowed", node.Line, node.Op, operand.Kind)
			}
			machine.Core[machine.PC], machine.PC = word, machine.PC+1
		case op.CSS: // pop the return stack
			machine.Core[machine.PC], machine.PC = word, machine.PC+1
		case op.DCL: // allocate memory, create a variable, link them
			if minArgs := 1; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			switch name := node.Parameters[0]; name.Kind {
			case ast.Variable:
				if _, ok := symtab.Lookup(name.Text); !ok {
					if ok := symtab.InsertAddress(name.Text, machine.PC); !ok {
						return nil, fmt.Errorf("%d: %s: internal error: %s redefined", node.Line, node.Op, name.Kind)
					}
				} else {
					symtab.UpdateAddress(name.Text, machine.PC)
				}
			default:
				return nil, fmt.Errorf("%d: %s: %s not allowed", node.Line, node.Op, name.Kind)
			}
			machine.Core[machine.PC], machine.PC = vm.Word{Text: node.Parameters[0].Text}, machine.PC+1
		case op.EQU:
			if minArgs := 2; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			// create an alias to another variable
			name, alias := node.Parameters[0].Text, node.Parameters[1].Text
			symtab.InsertAlias(name, alias)
		case op.EXIT: // exit a subroutine
			if minArgs := 2; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			switch exits := node.Parameters[0]; exits.Kind {
			case ast.Number:
				if exits.Number < 1 {
					return nil, fmt.Errorf("%d: %s: exit-number %d: invalid", node.Line, node.Op, exits.Number)
				} else if exits.Number > currSubroutine.numberOfExits {
					return nil, fmt.Errorf("%d: %s: exit-number %d: exceeds %d", node.Line, node.Op, exits.Number, currSubroutine.numberOfExits)
				}
				// the machine will take the exit value and add it to the return stack index, so we must decrement it here
				word.Value = exits.Number - 1
			default:
				return nil, fmt.Errorf("%d: %s: %s not allowed", node.Line, node.Op, exits.Kind)
			}
			// machine expects that label will match the subroutine name that we're currently in
			switch label := node.Parameters[1]; label.Kind {
			case ast.Variable:
				if label.Text != currSubroutine.name {
					return nil, fmt.Errorf("%d: %s: exit wants %q: got %q", node.Line, node.Op, currSubroutine.name, label.Text)
				}
			default:
				return nil, fmt.Errorf("%d: %s: %s not allowed", node.Line, node.Op, label.Kind)
			}
			machine.Core[machine.PC], machine.PC = word, machine.PC+1
		case op.GOADD: // multi-way branch
			// works by setting BranchPC. If GOxxx has the T flag and its PC matches BranchPC, then the branch is taken.
			if minArgs := 1; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			v := node.Parameters[0]
			switch v.Kind {
			case ast.Variable:
				symtab.UpdateAddress(v.Text, machine.PC)
			default:
				return nil, fmt.Errorf("%d: %s: %s: not allowed", node.Line, node.Op, v.Kind)
			}
			machine.Core[machine.PC], machine.PC = word, machine.PC+1
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
				return nil, fmt.Errorf("%d: %s: %s: not allowed", node.Line, node.Op, v.Kind)
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
				return nil, fmt.Errorf("%d: %s: %s not allowed", node.Line, node.Op, v.Kind)
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
				return nil, fmt.Errorf("%d: %s: %s not allowed", node.Line, node.Op, cdFlag.Kind)
			}
			machine.Core[machine.PC], machine.PC = word, machine.PC+1
		case op.LAV, op.LBV:
			// LAV load register A from variable V
			// LBV load register B from variable V
			if minArgs := 1; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			v := node.Parameters[0]
			switch v.Kind {
			case ast.Variable:
				symtab.AddReference(v.Text, machine.PC)
			default:
				return nil, fmt.Errorf("%d: %s: %s not allowed", node.Line, node.Op, v.Kind)
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
					return nil, fmt.Errorf("%d: %s: %s not allowed", node.Line, node.Op, rxFlag.Kind)
				}
			default:
				// no special action needed
			}
			machine.Core[machine.PC], machine.PC = word, machine.PC+1
		case op.MDLABEL: // create a label and link it to the current PC
			if minArgs := 1; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			switch name := node.Parameters[0]; name.Kind {
			case ast.Label:
				if _, ok := symtab.Lookup(name.Text); !ok {
					if ok := symtab.InsertAddress(name.Text, machine.PC); !ok {
						return nil, fmt.Errorf("%d: %s: internal error: %s redefined", node.Line, node.Op, name.Kind)
					}
				} else {
					symtab.UpdateAddress(name.Text, machine.PC)
				}
			default:
				return nil, fmt.Errorf("%d: %s: %s not allowed", node.Line, node.Op, name.Kind)
			}
		case op.MESS: // write message to output stream
			word.Text = node.Parameters[0].Text
			machine.Core[machine.PC], machine.PC = word, machine.PC+1
		case op.NB: // ignore comments
			// no special action needed
		case op.PRGST: // name the machine
			if minArgs := 1; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			foundPrgst = true
			name := node.Parameters[0].Text
			machine.Name = name
		case op.STI, op.STV:
			// STI indirect store of register A into variable V
			// STV store register A in variable V
			if minArgs := 2; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			switch v := node.Parameters[0]; v.Kind {
			case ast.Variable:
				symtab.AddReference(v.Text, machine.PC)
			default:
				return nil, fmt.Errorf("%d: %s: %s not allowed", node.Line, node.Op, v.Kind)
			}
			switch pxFlag := node.Parameters[1]; pxFlag.Kind {
			case ast.Variable:
				switch pxFlag.Text {
				case "P": // must preserve register A
					// no special action needed
				case "X": // okay to clobber register A
					// no special action needed
				default:
					return nil, fmt.Errorf("%d: %s: rxFlag want R|X: got %q", node.Line, node.Op, pxFlag.Text)
				}
			default:
				return nil, fmt.Errorf("%d: %s: %s not allowed", node.Line, node.Op, pxFlag.Kind)
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
		case op.SUBR:
			// declare a subroutine
			if minArgs := 3; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			switch name := node.Parameters[0]; name.Kind {
			case ast.Variable:
				if _, ok := symtab.Lookup(name.Text); !ok {
					if ok := symtab.InsertAddress(name.Text, machine.PC); !ok {
						return nil, fmt.Errorf("%d: %s: internal error: %s redefined", node.Line, node.Op, name.Kind)
					}
				} else {
					symtab.UpdateAddress(name.Text, machine.PC)
				}
				word.Text = name.Text
				currSubroutine.name = name.Text
			default:
				return nil, fmt.Errorf("%d: %s: %s not allowed", node.Line, node.Op, name.Kind)
			}
			switch flag := node.Parameters[1]; flag.Kind {
			case ast.Variable:
				switch flag.Text {
				case "X": // no parameters
					// emit a noop
					word.Op = op.NOOP
				case "PARNM": // named parameter
					// emit code to store register A into the named parameter
					symtab.AddReference(flag.Text, machine.PC)
					word.Op = op.STV
				default:
					return nil, fmt.Errorf("%d: %s: invalid parameter %q", node.Line, node.Op, flag.Text)
				}
			default:
				return nil, fmt.Errorf("%d: %s: %s not allowed", node.Line, node.Op, flag.Kind)
			}
			switch exits := node.Parameters[2]; exits.Kind {
			case ast.Number:
				currSubroutine.numberOfExits = exits.Number
			default:
				return nil, fmt.Errorf("%d: %s: %s not allowed", node.Line, node.Op, exits.Kind)
			}
			machine.Core[machine.PC], machine.PC = word, machine.PC+1

		default:
			return nil, fmt.Errorf("%d: %s: not implemented", node.Line, node.Op)
		}
	}

	if !foundPrgst {
		return nil, fmt.Errorf("PRGST missing")
	} else if !foundPrgen {
		return nil, fmt.Errorf("PRGEN missing")
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
		// fmt.Printf("asm: set vm begin   %-12s %6d\n", "", sym.address)
		machine.Core[0] = vm.Word{Op: op.GO, Value: sym.address}
	}

	// back-fill as needed
	for _, sym := range symtab.symbols {
		switch sym.kind {
		case "address":
			for _, addr := range sym.backFill {
				machine.Core[addr].Value = sym.address
			}
		case "constant":
			for _, addr := range sym.backFill {
				machine.Core[addr].Value = sym.constant
			}
		case "alias":
			aliasOf, ok := symtab.Lookup(sym.name)
			if !ok {
				return nil, fmt.Errorf("alias %q never defined", sym.name)
			}
			fmt.Printf("asm: alias %q to %q %s\n", sym.name, aliasOf.name, aliasOf.kind)
			switch aliasOf.kind {
			case "address":
				for _, addr := range sym.backFill {
					machine.Core[addr].Value = aliasOf.address
				}
			case "constant":
				for _, addr := range sym.backFill {
					machine.Core[addr].Value = aliasOf.constant
				}
			default:
				panic(fmt.Sprintf("assert(aliasOf.kind != %q)", aliasOf.kind))
			}
		default:
			panic(fmt.Sprintf("assert(sym.kind != %q)", sym.kind))
		}
	}

	machine.Core[0] = vm.Word{Op: op.GO, Value: machine.PC}

	panic("ast.Assemble is not implemented!")
}
