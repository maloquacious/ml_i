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
	symtab := newSymbolTable()
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
	for _, node := range nodes {
		// provide a default word for the instruction
		word := vm.Word{Op: node.Op} // default word to the current opcode

		// emit the word
		switch node.Op {

		// this section implements instructions that look like "OP"
		case op.ALIGN:
			// ALIGN emits no code
		case op.BMOVE, op.BSTK, op.CFSTK, op.CSS, op.FMOVE, op.FSTK:
			machine.Core[machine.PC], machine.PC = word, machine.PC+1
		case op.GOBRPC: // GOBRPC should not be available to callers
			return nil, fmt.Errorf("%d: %d: %s: internal error", node.Line, node.Col, node.Op)
		case op.PRGEN:
			machine.Core[machine.PC], machine.PC = vm.Word{Op: op.HALT}, machine.PC+1
		case op.UNKNOWN:
			return nil, fmt.Errorf("%d: %d: %s: internal error", node.Line, node.Col, node.Op)

		// this section implements instructions that look like "OP (CONSTANT_VAR|NUMBER)"
		case op.ANDL, op.CCN, op.LCN, op.NCH:
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

		// this section implements instructions that look like "OP (CONSTANT_VAR|NUMBER|N-OF)"
		case op.AAL, op.CAL, op.CON, op.LAL, op.LAM, op.LCM, op.MULTL, op.SAL, op.SBL:
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

		// this section implements instructions that look like "OP LABEL"
		case op.DCL:
			if minArgs := 1; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			switch label := node.Parameters[0]; label.Kind {
			case ast.Variable:
				if _, ok := symtab.Lookup(label.Text); !ok {
					if ok := symtab.InsertAddress(label.Text, machine.PC); !ok {
						return nil, fmt.Errorf("%d: %s: internal error: %s redefined", node.Line, node.Op, label.Kind)
					}
				} else {
					symtab.UpdateAddress(label.Text, machine.PC)
				}
				word.Text = label.Text
			default:
				return nil, fmt.Errorf("%d: %s: %s not allowed", node.Line, node.Op, label.Kind)
			}
			machine.Core[machine.PC], machine.PC = word, machine.PC+1
		case op.MDLABEL:
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
			// MDLABEL emits no code

		// this section implements instructions that look like "OP LABEL FLAG(PARNM|X) NUMBER"
		case op.SUBR:
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
				// add subroutine name for debugging
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

		// this section implements instructions that look like "OP LABEL FLAG(NUMBER|X)"
		case op.GOSUB:
			if minArgs := 2; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			switch v := node.Parameters[0]; v.Kind {
			case ast.Variable:
				symtab.UpdateAddress(v.Text, machine.PC)
			default:
				return nil, fmt.Errorf("%d: %s: %s: not allowed", node.Line, node.Op, v.Kind)
			}
			switch flag := node.Parameters[1]; flag.Kind {
			case ast.Number:
				// no action needed
			case ast.Variable:
				switch flag.Text {
				case "X": // calling MD-LOGIC!
					switch node.Parameters[0].Text {
					case "MDERCH", "MDQUIT":
						// no action needed for known MD functions
					default:
						return nil, fmt.Errorf("%d: GOSUB: undefined MD %q\n", node.Line, node.Parameters[0].Text)
					}
				default:
					return nil, fmt.Errorf("%d: %s: flag: want X: got %q (%q)", node.Line, node.Op, flag.Text)
				}
			default:
				return nil, fmt.Errorf("%d: %s: flag: want X or NUMBER: got %q", node.Line, node.Op, flag.Kind)
			}

		// this section implements instructions that look like "OP LABEL VARIABLE"
		case op.EQU:
			if minArgs := 2; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			var name string
			switch label := node.Parameters[0]; label.Kind {
			case ast.Variable:
				if _, ok := symtab.Lookup(label.Text); ok {
					return nil, fmt.Errorf("%d: %s: internal error: %q redefined", node.Line, node.Op, label.Text)
				}
				name = label.Text
			default:
				return nil, fmt.Errorf("%d: %s: %s not allowed", node.Line, node.Op, label.Kind)
			}
			switch v := node.Parameters[1]; v.Kind {
			case ast.Variable:
				symtab.InsertAlias(name, v.Text)
			default:
				return nil, fmt.Errorf("%d: %s: %s not allowed", node.Line, node.Op, v.Kind)
			}
			// EQU emits no code

		// this section implements instructions that look like "OP NUMBER LABEL"
		case op.EXIT:
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

		// this section implements instructions that look like "OP QUOTED_TEXT"
		case op.CCL:
			if minArgs := 1; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			switch text := node.Parameters[0]; text.Kind {
			case ast.QuotedText:
				if len(text.Text) != 1 {
					return nil, fmt.Errorf("%d: %s: want single character: got %q", node.Line, node.Op, text.Text)
				}
				word.Value = int(text.Text[0])
			default:
				return nil, fmt.Errorf("%d: %s: %s: not allowed", node.Line, node.Op, text.Kind)
			}
			machine.Core[machine.PC], machine.PC = word, machine.PC+1
		case op.MESS:
			if minArgs := 1; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			switch text := node.Parameters[0]; text.Kind {
			case ast.QuotedText:
				word.Text = text.Text
			default:
				return nil, fmt.Errorf("%d: %s: %s: not allowed", node.Line, node.Op, text.Kind)
			}
			machine.Core[machine.PC], machine.PC = word, machine.PC+1
		case op.NB: // ignore comments
			if minArgs := 1; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			switch text := node.Parameters[0]; text.Kind {
			case ast.QuotedText:
				// comments are ignored
			default:
				return nil, fmt.Errorf("%d: %s: %s: not allowed", node.Line, node.Op, text.Kind)
			}
			// NB emits no code
		case op.PRGST:
			if minArgs := 1; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			switch text := node.Parameters[0]; text.Kind {
			case ast.QuotedText:
				machine.Name, machine.Start = text.Text, machine.PC
			default:
				return nil, fmt.Errorf("%d: %s: %s: not allowed", node.Line, node.Op, text.Kind)
			}
			// PRGST emits no code
		case op.STR:
			if minArgs := 1; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			switch text := node.Parameters[0]; text.Kind {
			case ast.QuotedText:
				for _, ch := range text.Text {
					word.Value = int(ch)
					machine.Core[machine.PC], machine.PC = word, machine.PC+1
				}
			default:
				return nil, fmt.Errorf("%d: %s: %s: not allowed", node.Line, node.Op, text.Kind)
			}
			// STR emits code in the text loop above

		// this section implements instructions that look like "OP VARIABLE"
		case op.AAV, op.ABV, op.ANDV, op.CCI, op.CLEAR, op.GOADD, op.LBV, op.SAV, op.SBV, op.UNSTK:
			// implementation note: GOADD works by setting BranchPC. If GOxxx has the T flag and its PC matches BranchPC, then the branch is taken.
			if minArgs := 1; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			switch v := node.Parameters[0]; v.Kind {
			case ast.Variable:
				symtab.AddReference(v.Text, machine.PC)
			default:
				return nil, fmt.Errorf("%d: %s: %s: not allowed", node.Line, node.Op, v.Kind)
			}
			machine.Core[machine.PC], machine.PC = word, machine.PC+1

		// this section implements instructions that look like "OP VARIABLE FLAG(A|X)"
		case op.CAI, op.CAV:
			if minArgs := 2; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			switch v := node.Parameters[0]; v.Kind {
			case ast.Variable:
				symtab.AddReference(v.Text, machine.PC)
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
					return nil, fmt.Errorf("%d: %s: flag want A|X: got %q", node.Line, node.Op, flag)
				}
			default:
				return nil, fmt.Errorf("%d: %s: %s not allowed", node.Line, node.Op, flag.Kind)
			}
			machine.Core[machine.PC], machine.PC = word, machine.PC+1

		// this section implements instructions that look like "OP VARIABLE FLAG(C|D)"
		case op.LAA:
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
				case "C": // load A with the address of the table label
					// no special action needed
				case "D": // load A with the address of variable V
					// no special action needed
				default:
					return nil, fmt.Errorf("%d: %s: flag want C|D: got %q", node.Line, node.Op, flag.Text)
				}
			default:
				return nil, fmt.Errorf("%d: %s: %s not allowed", node.Line, node.Op, flag.Kind)
			}
			machine.Core[machine.PC], machine.PC = word, machine.PC+1

		// this section implements instructions that look like "OP VARIABLE FLAG(P|X)"
		case op.STI, op.STV:
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

		// this section implements instructions that look like "OP VARIABLE FLAG(R|X)"
		case op.LAI, op.LAV, op.LCI:
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
					return nil, fmt.Errorf("%d: %s: flag want R|X: got %q", node.Line, node.Op, flag.Text)
				}
			default:
				return nil, fmt.Errorf("%d: %s: %s not allowed", node.Line, node.Op, flag.Kind)
			}
			machine.Core[machine.PC], machine.PC = word, machine.PC+1

		// this section implements instructions that look like "OP VARIABLE NUMBER"
		case op.IDENT:
			if minArgs := 2; len(node.Parameters) < minArgs {
				return nil, fmt.Errorf("%d: %s: want %d args: got %d", node.Line, node.Op, minArgs, len(node.Parameters))
			}
			switch v := node.Parameters[0]; v.Kind {
			case ast.Variable:
				// nothing special
			default:
				return nil, fmt.Errorf("%d: %s: want variable: got %s", node.Line, node.Op, v.Kind)
			}
			switch constant := node.Parameters[1]; constant.Kind {
			case ast.Number:
				symtab.InsertConstant(node.Parameters[0].Text, constant.Number)
			default:
				return nil, fmt.Errorf("%d: %s: want constant: got %s", node.Line, node.Op, constant.Kind)
			}

		// this section implements instructions that look like "OP VARIABLE (NUMBER | CONSTANT_VAR | N-OF)"
		case op.BUMP:
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
				return nil, fmt.Errorf("%d: %s: %s: not allowed", node.Line, node.Op, nOF.Kind)
			}
			machine.Core[machine.PC], machine.PC = word, machine.PC+1

		// this section implements op codes that require a label spec
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
		// fmt.Printf("asm: set vm begin   %-12s %6d\n", "", sym.address)
		machine.Core[0] = vm.Word{Op: op.GO, Value: sym.address}
	}

	// detect and report undefined symbols
	undefinedSymbols := 0
	for _, sym := range symtab.symbols {
		if sym.defined {
			continue
		}
		fmt.Printf("asm: error: undefined symbol %q %q\n", sym.name, sym.kind)
		undefinedSymbols++
	}
	if undefinedSymbols != 0 {
		return nil, fmt.Errorf("found %d undefined symbols", undefinedSymbols)
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

	return machine, nil
}
