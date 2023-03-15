package lowl

import (
	"bytes"
	"fmt"
	"github.com/maloquacious/ml_i/pkg/lowl/op"
	"github.com/maloquacious/ml_i/pkg/postfix"
	"github.com/maloquacious/ml_i/pkg/tokens"
	"os"
	"sort"
	"strconv"
	"strings"
)

func Assemble(name string) (*VM, *bytes.Buffer, error) {
	listing := &bytes.Buffer{}
	vm := &VM{}
	vm.heaps.msg = make(map[int]string)
	// vm.heaps.vars = []WORD{0} // reserve the first cell, ok

	// the label and symbol tables are map of name to address in memory
	symtab := newSymbolTable()

	// set up variables for OF macro and other things
	if err := symtab.defConstant("LCH", 0, LCH); err != nil {
		return nil, listing, fmt.Errorf("def-constant %w", err)
	} else if err = symtab.defConstant("LNM", 0, LNM); err != nil {
		return nil, listing, fmt.Errorf("def-constant %w", err)
	} else if err = symtab.defConstant("LICH", 0, LICH); err != nil {
		return nil, listing, fmt.Errorf("def-constant %w", err)
	} else if err = symtab.defConstant("NLREP", 0, '\n'); err != nil {
		return nil, listing, fmt.Errorf("def-constant %w", err)
	} else if err = symtab.defConstant("QUTREP", 0, '"'); err != nil {
		return nil, listing, fmt.Errorf("def-constant %w", err)
	} else if err = symtab.defConstant("SPREP", 0, ' '); err != nil {
		return nil, listing, fmt.Errorf("def-constant %w", err)
	} else if err = symtab.defConstant("TABREP", 0, '\t'); err != nil {
		return nil, listing, fmt.Errorf("def-constant %w", err)
	}

	input, err := os.ReadFile(name)
	if err != nil {
		return nil, listing, fmt.Errorf("assemble: %w", err)
	}

	// todo: set up boot area of the VM

	line := 0
	for toks, rest := tokens.NextLine(input); len(rest) != 0; toks, rest = tokens.NextLine(rest) {
		line++

		if len(toks) == 1 && len(toks[0]) == 0 || toks[0][0] == '\n' {
			fprintf(listing, "\n")
			continue
		}

		// extract label, op-code, and parameters from the line
		var label, opc string
		var parameters []string
		for n, tok := range toks {
			if len(tok) == 0 || tok[0] == '\n' {
				// ignore
			} else if n == 0 && tok[0] == '[' {
				label = string(tok[1 : len(tok)-1])
			} else if opc == "" {
				opc = string(tok)
			} else {
				parameters = append(parameters, string(tok))
			}
		}
		parameters = pToCSL(parameters)
		for idx, parm := range parameters {
			env := symtab.getConstants()
			//// maybe a constant
			//if n, ok := env[parm]; ok {
			//	parameters[idx] = fmt.Sprintf("%d", n)
			//	continue
			//}
			if strings.HasPrefix(parm, "OF(") && strings.HasSuffix(parm, ")") {
				if n, err := ofMacro(ParseOF(parm[2:]), env); err != nil {
					return nil, listing, err
				} else {
					parameters[idx] = fmt.Sprintf("%d", n)
					continue
				}
			}
		}

		// add the line to the listing
		if listing != nil {
			pout, sep := "", ""
			for _, parm := range parameters {
				pout += sep + parm
				sep = " "
			}
			if label == "" {
				fprintf(listing, "%-10s  %-10s %-55s ;; %4d\n", label, opc, pout, line)
			} else {
				fprintf(listing, "%-10s  %-10s %-55s ;; %4d\n", "["+label+"]", opc, pout, line)
			}
		}

		// assemble it here

		// [label] creates a label for the current address
		if label != "" {
			if label == "BEGIN" {
				fmt.Printf("assm: %06d: found label %q\n", vm.pc, label)
			}
			if err := symtab.defLabel(label, line, vm.pc); err != nil {
				return nil, listing, fmt.Errorf("%d: label %q: %w", line, label, err)

			}
		}

		switch opc {
		case "AAL":
			word := Word{op: op.AAL}
			// AAL N-OF
			if len(parameters) < 1 {
				return nil, listing, fmt.Errorf("%d: %s: want 1 args: got %d", line, opc, len(parameters))
			}
			// set the operand to the literal value
			if n, err := strconv.Atoi(parameters[0]); err != nil {
				return nil, listing, fmt.Errorf("%d: %s %w", line, opc, err)
			} else {
				word.data = WORD(n)
			}
			vm.core[vm.pc], vm.pc = word, vm.pc+1
		case "AAV": // add a variable to A
			// AAV V
			word := Word{op: op.AAV}
			if len(parameters) < 1 {
				return nil, listing, fmt.Errorf("%d: %s: want 1 args: got %d", line, opc, len(parameters))
			}
			vName := parameters[0]
			if sym, ok := symtab.getUnaliasedSymbol(vName); !ok || sym.kind == SymIsBackfill {
				symtab.refVariable(sym.name, vm.pc) // operand to be back-filled
			} else if sym.kind == SymIsOnHeap {
				word.data = WORD(sym.address) // operand is the variable address
			} else {
				return nil, listing, fmt.Errorf("%d: %s %q: invalid", line, opc, vName)
			}
			vm.core[vm.pc], vm.pc = word, vm.pc+1
		case "ABV":
			word := Word{op: op.ABV}
			vm.core[vm.pc], vm.pc = word, vm.pc+1
		case "ALIGN":
			vm.core[vm.pc], vm.pc = Word{op: op.ALIGN}, vm.pc+1
		case "ANDL":
			vm.core[vm.pc], vm.pc = Word{op: op.ANDL}, vm.pc+1
		case "ANDV":
			vm.core[vm.pc], vm.pc = Word{op: op.ANDV}, vm.pc+1
		case "BMOVE":
			vm.core[vm.pc], vm.pc = Word{op: op.BMOVE}, vm.pc+1
		case "BSTK":
			vm.core[vm.pc], vm.pc = Word{op: op.BSTK}, vm.pc+1
		case "BUMP":
			vm.core[vm.pc], vm.pc = Word{op: op.BUMP}, vm.pc+1
		case "CAI":
			vm.core[vm.pc], vm.pc = Word{op: op.CAI}, vm.pc+1
		case "CAL":
			word := Word{op: op.CAL}
			// CAL N-OF
			if len(parameters) < 1 {
				return nil, listing, fmt.Errorf("%d: %s: want 1 args: got %d", line, opc, len(parameters))
			}
			// set the operand to the literal value
			if n, err := strconv.Atoi(parameters[0]); err != nil {
				return nil, listing, fmt.Errorf("%d: %s %w", line, opc, err)
			} else {
				word.data = WORD(n)
			}
			vm.core[vm.pc], vm.pc = word, vm.pc+1
		case "CAV":
			word := Word{op: op.CAV}
			// CAV * V,(A)
			// CAV * V,(X)
			if len(parameters) < 2 {
				return nil, listing, fmt.Errorf("%d: %s: want 2 args: got %d", line, opc, len(parameters))
			}
			vName, flag := parameters[0], parameters[1]
			switch flag {
			case "A": // compare to address
				symtab.refVariable(vName, vm.pc) // operand to be back-filled
				if sym, ok := symtab.getUnaliasedSymbol(vName); !ok || sym.kind == SymIsBackfill {
					symtab.refVariable(vName, vm.pc) // operand to be back-filled
				} else if sym.kind == SymIsOnHeap {
					word.data = WORD(sym.heapIndex) // operand is the variable address
				} else {
					return nil, listing, fmt.Errorf("%d: %s %q: invalid", line, opc, vName)
				}
			case "X": // compare to signed value
				if sym, ok := symtab.getUnaliasedSymbol(vName); !ok || sym.kind == SymIsBackfill {
					symtab.refVariable(vName, vm.pc) // operand to be back-filled
				} else if sym.kind == SymIsOnHeap {
					word.data = WORD(sym.heapIndex) // operand is the variable address
				} else {
					return nil, listing, fmt.Errorf("%d: %s %q: invalid", line, opc, vName)
				}
			default:
				return nil, listing, fmt.Errorf("%d: %s: unknown flag %q", line, opc, flag)
			}
			vm.core[vm.pc], vm.pc = word, vm.pc+1
		case "CCI":
			vm.core[vm.pc], vm.pc = Word{op: op.CCI}, vm.pc+1
		case "CCL":
			vm.core[vm.pc], vm.pc = Word{op: op.CCL}, vm.pc+1
		case "CCN":
			vm.core[vm.pc], vm.pc = Word{op: op.CCN}, vm.pc+1
		case "CFSTK":
			vm.core[vm.pc], vm.pc = Word{op: op.CFSTK}, vm.pc+1
		case "CLEAR":
			vm.core[vm.pc], vm.pc = Word{op: op.CLEAR}, vm.pc+1
		case "CON":
			word := Word{op: op.CON}
			// CON ( N-OF)
			// CON (-N-OF)
			if len(parameters) < 1 {
				return nil, listing, fmt.Errorf("%d: %s: want 2 args: got %d", line, opc, len(parameters))
			}
			// set the operand to the value of the constant
			var number string
			for _, parm := range parameters {
				number = number + parm
			}
			n, err := strconv.Atoi(number)
			if err != nil {
				return nil, listing, fmt.Errorf("%d: %s %q: %w", line, opc, number, err)
			}
			word.data = WORD(n)
			vm.core[vm.pc], vm.pc = word, vm.pc+1
		case "CSS":
			vm.core[vm.pc], vm.pc = Word{op: op.CSS}, vm.pc+1
		case "DCL":
			// DCL name
			if len(parameters) < 1 {
				return nil, listing, fmt.Errorf("%d: %s: missing name", line, opc)
			}
			sym := parameters[0]
			if err := symtab.defVariable(sym, line, vm); err != nil {
				return nil, listing, fmt.Errorf("%d: %s: %w", line, opc, err)
			}
		case "EQU":
			// EQU V,V
			if len(parameters) < 2 {
				return nil, listing, fmt.Errorf("%d: %s: want 2 args: got %d", line, opc, len(parameters))
			}
			alias, sym := parameters[0], parameters[1]
			// add the alias to the symbol table.
			if err := symtab.defAlias(alias, sym, line); err != nil {
				return nil, listing, fmt.Errorf("%d: %s %q: %w", line, opc, alias, err)
			}
		case "EXIT":
			word := Word{op: op.EXIT}
			// EXIT * N,subroutine name
			if len(parameters) < 2 {
				return nil, listing, fmt.Errorf("%d: %s: want 2 args: got %d", line, opc, len(parameters))
			}
			exitNumber, name := parameters[0], parameters[1]
			fmt.Printf("exit %q %q\n", exitNumber, name)
			if n, err := strconv.Atoi(exitNumber); err != nil {
				return nil, listing, fmt.Errorf("%d: %s exit wants number: got %q", line, opc, exitNumber)
			} else if !(0 < n && n < 256) {
				return nil, listing, fmt.Errorf("%d: %s exit wants number 0<n&&n<256: got %d", line, opc, n)
			} else {
				// take one from the number since our return is zero-based.
				word.data = WORD(n - 1)
			}
			fmt.Printf("%4d: %s must clear the return stack\n", line, opc)
			vm.core[vm.pc], vm.pc = word, vm.pc+1
		case "FMOVE":
			vm.core[vm.pc], vm.pc = Word{op: op.FMOVE}, vm.pc+1
		case "FSTK":
			vm.core[vm.pc], vm.pc = Word{op: op.FSTK}, vm.pc+1
		case "GO", "GOEQ", "GOGE", "GOGR", "GOLE", "GOLT", "GOND", "GONE", "GOPC":
			var word Word
			switch opc {
			case "GO":
				word.op = op.GO
			case "GOEQ":
				word.op = op.GOEQ
			case "GOGE":
				word.op = op.GOGE
			case "GOGR":
				word.op = op.GOGR
			case "GOLE":
				word.op = op.GOLE
			case "GOLT":
				word.op = op.GOLT
			case "GOND":
				word.op = op.GOND
			case "GONE":
				word.op = op.GONE
			case "GOPC":
				word.op = op.GOPC
			default:
				panic(fmt.Sprintf("assert(op != %q)", opc))
			}
			// GO label,distance,(E|X),(C|T|X)
			if len(parameters) < 4 {
				return nil, listing, fmt.Errorf("%d: %s: want 4 args: got %d", line, opc, len(parameters))
			}
			label, distance, exFlag, ctxFlag := parameters[0], parameters[1], parameters[2], parameters[3]
			if _, err := strconv.Atoi(distance); err != nil {
				return nil, listing, fmt.Errorf("%d: %s: distance want int: got %q", line, opc, distance)
			}
			switch exFlag {
			case "E": // exits subroutine
				fmt.Printf("%4d: %s must clear the return stack\n", line, opc)
				switch opc {
				case "GO":
					word.op = op.EXIT
				case "GOEQ":
					word.op = op.EXITEQ
				case "GOGE":
					word.op = op.EXITGE
				case "GOGR":
					word.op = op.EXITGR
				case "GOLE":
					word.op = op.EXITLE
				case "GOLT":
					word.op = op.EXITLT
				case "GOND":
					word.op = op.EXITND
				case "GONE":
					word.op = op.EXITNE
				case "GOPC":
					word.op = op.EXITPC
				default:
					panic(fmt.Sprintf("assert(op != %q)", opc))
				}
			case "X": // nothing special
			default:
				return nil, listing, fmt.Errorf("%d: %s: exFlag want E|X: got %q", line, opc, exFlag)
			}
			switch ctxFlag {
			case "C": // exit following gosub
			case "T": // GOADD branch
			case "X": // nothing special
			default:
				return nil, listing, fmt.Errorf("%d: %s: ctxFlag want C|T|X: got %q", line, opc, ctxFlag)
			}
			if sym, ok := symtab.getUnaliasedSymbol(name); !ok || sym.kind == SymIsBackfill {
				symtab.refVariable(label, vm.pc) // operand to be back-filled
			} else if sym.kind == SymIsAddress {
				word.data = WORD(sym.address) // operand is the label's address
			} else {
				return nil, listing, fmt.Errorf("%d: %s %q: invalid", line, opc, label)
			}
			vm.core[vm.pc], vm.pc = word, vm.pc+1
		case "GOADD":
			vm.core[vm.pc], vm.pc = Word{op: op.GOADD}, vm.pc+1
		case "GOSUB":
			word := Word{op: op.GOSUB}
			// GOSUB subroutine name,(distance)
			// GOSUB subroutine name,(X)
			if len(parameters) < 2 {
				return nil, listing, fmt.Errorf("%d: %s: want 2 args: got %d", line, opc, len(parameters))
			}
			vName, flag := parameters[0], parameters[1]
			switch flag {
			case "X": // reference a routine in the MD-logic
				switch vName {
				case "MDERCH": // output register C to message stream
					word = Word{op: op.MDERCH}
				case "MDQUIT": // quit the application
					word = Word{op: op.MDQUIT}
				default:
					return nil, listing, fmt.Errorf("%d: %s: unknown MD routine %q", line, opc, vName)
				}
			default:
				// flag is unused, but verify that is a number
				if flag == "-" { // might be a negative number
					for _, parm := range parameters[3:] {
						flag = flag + parm
					}
				}
				if _, err := strconv.Atoi(flag); err != nil {
					return nil, listing, fmt.Errorf("%d: %s: unknown flag %q", line, opc, flag)
				}
				if sym, ok := symtab.getUnaliasedSymbol(vName); !ok || sym.kind == SymIsBackfill {
					symtab.refVariable(sym.name, vm.pc) // operand to be back-filled
				} else if sym.kind == SymIsSubroutine {
					word.data = WORD(sym.address) // operand is the subroutine address
				} else {
					return nil, listing, fmt.Errorf("%d: %s %q: invalid", line, opc, vName)
				}
			}
			vm.core[vm.pc], vm.pc = word, vm.pc+1
		case "IDENT":
			// IDENT V,decimal integer
			if len(parameters) < 2 {
				return nil, listing, fmt.Errorf("%d: %s: want 2 args: got %d", line, opc, len(parameters))
			}
			vName, number := parameters[0], parameters[1]
			if sym, ok := symtab.getSymbol(vName); ok && sym.kind != SymIsUnknown {
				return nil, listing, fmt.Errorf("%d: %s %q: redefined", line, opc, vName)
			}
			n, err := strconv.Atoi(number)
			if err != nil {
				return nil, listing, fmt.Errorf("%d: %s %q: %w", line, opc, vName, err)
			}
			// add to the symbol table as a new constant
			if err := symtab.defConstant(vName, line, WORD(n)); err != nil {
				return nil, listing, fmt.Errorf("%d: %s %q: %w", line, opc, vName, err)
			}
		case "LAA":
			vm.core[vm.pc], vm.pc = Word{op: op.LAA}, vm.pc+1
		case "LAI":
			word := Word{op: op.LAI}
			// LAI V,(R)
			// LAI V,(X)
			if len(parameters) < 2 {
				return nil, listing, fmt.Errorf("%d: %s: want 2 args: got %d", line, opc, len(parameters))
			}
			vName, flag := parameters[0], parameters[1]
			switch flag {
			case "R":
				symtab.refVariable(vName, vm.pc) // operand to be back-filled
			case "X":
				symtab.refVariable(vName, vm.pc) // operand to be back-filled
			default:
				return nil, listing, fmt.Errorf("%d: %s: unknown flag %q", line, opc, flag)
			}
			vm.core[vm.pc], vm.pc = word, vm.pc+1
		case "LAL":
			word := Word{op: op.LAL}
			// LAL N-OF
			if len(parameters) < 1 {
				return nil, listing, fmt.Errorf("%d: %s: want 1 args: got %d", line, opc, len(parameters))
			}
			// set the operand to the literal value
			if n, err := strconv.Atoi(parameters[0]); err != nil {
				return nil, listing, fmt.Errorf("%d: %s %w", line, opc, err)
			} else {
				word.data = WORD(n)
			}
			vm.core[vm.pc], vm.pc = word, vm.pc+1
		case "LAM":
			vm.core[vm.pc], vm.pc = Word{op: op.LAM}, vm.pc+1
		case "LAV":
			word := Word{op: op.LAV}
			// LAV V,(X)
			// LAV V,(R)
			if len(parameters) < 2 {
				return nil, listing, fmt.Errorf("%d: %s: want 2 args: got %d", line, opc, len(parameters))
			}
			vName, flag := parameters[0], parameters[1]
			switch flag {
			case "R":
				symtab.refVariable(vName, vm.pc) // operand to be back-filled
			case "X":
				symtab.refVariable(vName, vm.pc) // operand to be back-filled
			default:
				return nil, listing, fmt.Errorf("%d: %s: unknown flag %q", line, opc, flag)
			}
			vm.core[vm.pc], vm.pc = word, vm.pc+1
		case "LBV":
			word := Word{op: op.LBV}
			// LBV V
			if len(parameters) < 1 {
				return nil, listing, fmt.Errorf("%d: %s: want 1 args: got %d", line, opc, len(parameters))
			}
			vName := parameters[0]
			symtab.refVariable(vName, vm.pc) // operand to be back-filled
			vm.core[vm.pc], vm.pc = word, vm.pc+1
		case "LCI":
			word := Word{op: op.LCI}
			// LCI V,(R)
			// LCI V,(X)
			if len(parameters) < 2 {
				return nil, listing, fmt.Errorf("%d: %s: want 2 args: got %d", line, opc, len(parameters))
			}
			vName, flag := parameters[0], parameters[1]
			switch flag {
			case "R":
				symtab.refVariable(vName, vm.pc) // operand to be back-filled
			case "X":
				symtab.refVariable(vName, vm.pc) // operand to be back-filled
			default:
				return nil, listing, fmt.Errorf("%d: %s: unknown flag %q", line, opc, flag)
			}
			vm.core[vm.pc], vm.pc = word, vm.pc+1
		case "LCM":
			vm.core[vm.pc], vm.pc = Word{op: op.LCM}, vm.pc+1
		case "LCN":
			word := Word{op: op.LCN}
			// LCN charname
			if len(parameters) < 1 {
				return nil, listing, fmt.Errorf("%d: %s: want 1 args: got %d", line, opc, len(parameters))
			}
			if ch, ok := symtab.getConstant(parameters[0]); !ok {
				return nil, listing, fmt.Errorf("%d: %s: unknown constant %q", line, opc, parameters[0])
			} else {
				word.data = WORD(ch) // set operand to the character's value
			}
			// add the instruction with the operand set to the constant
			vm.core[vm.pc], vm.pc = word, vm.pc+1
		case "MESS": // write a message to a stream
			word := Word{op: op.MESS}
			// MESS 'characters'
			if len(parameters) < 1 {
				return nil, listing, fmt.Errorf("%d: %s: want 1 arg: got %d", line, opc, len(parameters))
			}
			msg := parameters[0]
			if !strings.HasPrefix(msg, "'") {
				return nil, listing, fmt.Errorf("%d: %s: missing open quote", line, opc)
			} else if !strings.HasSuffix(msg, "'") {
				return nil, listing, fmt.Errorf("%d: %s: missing close quote", line, opc)
			} else if len(msg) < 3 {
				return nil, listing, fmt.Errorf("%d: %s: malformed quotes", line, opc)
			}
			// remove the quotes and translate the $ in messages to new-lines
			msg = msg[1 : len(msg)-1]
			msg = strings.ReplaceAll(msg, "$", "\n")
			// add the message to the heap
			vm.heaps.msg[line] = msg
			word.data = WORD(line) // set the operand to the heap index
			//fmt.Printf("assm: %06d: %s\n", vm.pc, op.MESS.String())
			vm.core[vm.pc], vm.pc = word, vm.pc+1
		case "MULTL":
			vm.core[vm.pc], vm.pc = Word{op: op.MULTL}, vm.pc+1
		case "NB":
			word := Word{op: op.NB}
			// NB comment
			var comment, sep string
			for _, arg := range parameters {
				comment = comment + sep + arg
				sep = " "
			}
			vm.core[vm.pc], vm.pc = word, vm.pc+1
		case "NCH": // character constant
			word := Word{op: op.NCH}
			// NCH charname
			if len(parameters) < 1 {
				return nil, listing, fmt.Errorf("%d: %s: want 1 arg: got %d", line, opc, len(parameters))
			}
			if ch, ok := symtab.getConstant(parameters[0]); !ok {
				return nil, listing, fmt.Errorf("%d: %s: unknown constant %q", line, opc, parameters[0])
			} else {
				word.data = WORD(ch) // set the operand to the constant
			}
			vm.core[vm.pc], vm.pc = word, vm.pc+1
		case "PANIC":
			vm.core[vm.pc], vm.pc = Word{op: op.PANIC}, vm.pc+1
		case "PRGEN":
			vm.core[vm.pc], vm.pc = Word{op: op.PRGEN}, vm.pc+1
		case "PRGST":
			word := Word{op: op.PRGST}
			// PRGST 'characters'
			if len(parameters) < 1 {
				return nil, listing, fmt.Errorf("%d: %s: want 1 arg: got %d", line, opc, len(parameters))
			}
			word.comment = parameters[0]
			vm.core[vm.pc], vm.pc = word, vm.pc+1
		case "SAL":
			word := Word{op: op.SAL}
			// SAL N-OF
			if len(parameters) < 1 {
				return nil, listing, fmt.Errorf("%d: %s: want 1 args: got %d", line, opc, len(parameters))
			}
			// set the operand to the literal value
			if n, err := strconv.Atoi(parameters[0]); err != nil {
				return nil, listing, fmt.Errorf("%d: %s %w", line, opc, err)
			} else {
				word.data = WORD(n)
			}
			vm.core[vm.pc], vm.pc = word, vm.pc+1
		case "SAV":
			word := Word{op: op.SAV}
			// SAV V
			if len(parameters) < 1 {
				return nil, listing, fmt.Errorf("%d: %s: want 1 args: got %d", line, opc, len(parameters))
			}
			vName := parameters[0]
			if sym, ok := symtab.getUnaliasedSymbol(vName); !ok || sym.kind == SymIsBackfill {
				symtab.refVariable(sym.name, vm.pc) // operand to be back-filled
			} else if sym.kind == SymIsOnHeap {
				word.data = WORD(sym.address) // operand is the variable address
			} else {
				return nil, listing, fmt.Errorf("%d: %s %q: invalid", line, opc, vName)
			}
			vm.core[vm.pc], vm.pc = word, vm.pc+1
		case "SBL":
			vm.core[vm.pc], vm.pc = Word{op: op.SBL}, vm.pc+1
		case "SBV":
			vm.core[vm.pc], vm.pc = Word{op: op.SBV}, vm.pc+1
		case "STI":
			vm.core[vm.pc], vm.pc = Word{op: op.STI}, vm.pc+1
		case "STR": // character string constant
			// STR 'characters'
			if len(parameters) < 1 {
				return nil, listing, fmt.Errorf("%d: %s: want 1 arg: got %d", line, opc, len(parameters))
			}
			str := parameters[0]
			if !strings.HasPrefix(str, "'") {
				return nil, listing, fmt.Errorf("%d: %s: missing open quote", line, opc)
			} else if !strings.HasSuffix(str, "'") {
				return nil, listing, fmt.Errorf("%d: %s: missing close quote", line, opc)
			} else if len(str) < 3 {
				return nil, listing, fmt.Errorf("%d: %s: malformed quotes", line, opc)
			}
			// remove the quotes and translate the $ in messages to new-lines
			str = str[1 : len(str)-1]
			str = strings.ReplaceAll(str, "$", "\n")
			// add the instruction with one word per character of the string
			for _, ch := range str {
				vm.core[vm.pc], vm.pc = Word{op: op.STR, data: WORD(ch)}, vm.pc+1
			}
		case "STV":
			word := Word{op: op.STV}
			// STV V,(P)
			// STV V,(X)
			if len(parameters) < 2 {
				return nil, listing, fmt.Errorf("%d: %s: want 2 args: got %d", line, opc, len(parameters))
			}
			vName, flag := parameters[0], parameters[1]
			switch flag {
			case "P":
				if sym, ok := symtab.getUnaliasedSymbol(vName); !ok || sym.kind == SymIsBackfill {
					symtab.refVariable(sym.name, vm.pc) // operand to be back-filled
				} else if sym.kind == SymIsOnHeap {
					word.data = WORD(sym.address) // operand is the variable address
				} else {
					return nil, listing, fmt.Errorf("%d: %s %q: invalid", line, opc, vName)
				}
			case "X":
				if sym, ok := symtab.getUnaliasedSymbol(vName); !ok || sym.kind == SymIsBackfill {
					symtab.refVariable(sym.name, vm.pc) // operand to be back-filled
				} else if sym.kind == SymIsOnHeap {
					word.data = WORD(sym.address) // operand is the variable address
				} else {
					return nil, listing, fmt.Errorf("%d: %s %q: invalid", line, opc, vName)
				}
			default:
				return nil, listing, fmt.Errorf("%d: %s: unknown flag %q", line, opc, flag)
			}
			vm.core[vm.pc], vm.pc = word, vm.pc+1
		case "SUBR":
			// SUBR  subroutine name,(PARNM),N  // declare subroutine with parameter and/or exits
			// SUBR  subroutine name,(X    ),N  // declare subroutine no parameters, maybe exits
			if len(parameters) < 3 {
				return nil, listing, fmt.Errorf("%d: %s: want 3 args: got %d", line, opc, len(parameters))
			}
			name, pName := parameters[0], parameters[1]
			numberOfExits := 0
			if n, err := strconv.Atoi(parameters[2]); err != nil {
				return nil, listing, fmt.Errorf("%d: %s: numberOfExits wants int: got %q", line, opc, numberOfExits)
			} else if n == 0 {
				numberOfExits = 1
			} else {
				numberOfExits = n
			}
			fmt.Printf("%s %q %q %d\n", opc, name, pName, numberOfExits)
			if pName == "X" {
				pName = "" // unset because it really isn't a parameter name in this case
			}
			// create a new subroutine label here
			if err := symtab.defSubroutine(name, line, vm.pc, pName, numberOfExits); err != nil {
				return nil, listing, fmt.Errorf("%d: %s %q: %w", line, opc, name, err)
			}
			// if we're given a parameter name, then we must cause that variable to be loaded when the routine is called
			if pName == "" { // is there a named parameter?
				// there isn't, so the first instruction of the sub-routine is a NOOP
				word := Word{op: op.NOOP}
				vm.core[vm.pc], vm.pc = word, vm.pc+1
			} else {
				// there is, so the first operation must be "STV pName" to store register A into the variable
				word := Word{op: op.STV}
				if sym, ok := symtab.getUnaliasedSymbol(pName); !ok || sym.kind == SymIsBackfill {
					symtab.refVariable(sym.name, vm.pc) // operand to be back-filled
				} else if sym.kind == SymIsOnHeap {
					word.data = WORD(sym.address) // operand is the variable address
				} else {
					return nil, listing, fmt.Errorf("%d: %s %q: invalid", line, opc, pName)
				}
				vm.core[vm.pc], vm.pc = word, vm.pc+1
			}
		case "UNSTK":
			vm.core[vm.pc], vm.pc = Word{op: op.UNSTK}, vm.pc+1
		default:
			panic(fmt.Sprintf("assert(opcode != %q)", opc))
		}
	}

	// we must back-fill all references
	for name, addresses := range symtab.backfill {
		sym, ok := symtab.table[name]
		if !ok {
			panic(fmt.Sprintf("assert(%q is defined)", name))
		} else if sym.kind == SymIsAlias {
			if sym, ok = symtab.table[sym.aliasOf]; !ok {
				panic(fmt.Sprintf("%q is bad alias", name))
			}
		}
		for _, address := range addresses {
			switch sym.kind {
			case SymIsAddress:
				vm.core[address].data = WORD(sym.address)
			case SymIsOnHeap:
				vm.core[address].data = WORD(sym.heapIndex)
			case SymIsSubroutine:
				vm.core[address].data = WORD(sym.address)
			case SymIsValue:
				vm.core[address].data = WORD(sym.value)
			}
		}
	}

	if listing != nil {
		var lines []string
		for name, value := range symtab.table {
			switch value.kind {
			case SymIsAlias:
				lines = append(lines, fmt.Sprintf("%-15s  defn %4d     alias-of %s", name, value.line, value.aliasOf))
			case SymIsAddress:
				lines = append(lines, fmt.Sprintf("%-15s  defn %4d      address %8d", name, value.line, value.address))
			case SymIsBackfill:
				lines = append(lines, fmt.Sprintf("%-15s  defn %4d    back-fill", name, value.line))
			case SymIsOnHeap:
				lines = append(lines, fmt.Sprintf("%-15s  defn %4d   heap-index %8d", name, value.line, value.heapIndex))
			case SymIsSubroutine:
				lines = append(lines, fmt.Sprintf("%-15s  defn %4d  sub-routine %8d", name, value.line, value.address))
			case SymIsValue:
				lines = append(lines, fmt.Sprintf("%-15s  defn %4d        value %8d", name, value.line, value.value))
			case SymIsUnknown:
				lines = append(lines, fmt.Sprintf("%-15s  defn %4d    undefined", name, value.line))
			}
		}
		sort.Strings(lines)
		for _, line := range lines {
			fprintf(listing, ";; %s\n", line)
		}
	}

	if listing != nil {
		fprintf(listing, ";; messages\n")
		var lines []int
		for line := range vm.heaps.msg {
			lines = append(lines, line)
		}
		sort.Ints(lines)
		for _, line := range lines {
			fprintf(listing, ";;    defn %4d %q\n", line, vm.heaps.msg[line])
		}
	}

	// the first executable statement is labelled BEGIN
	if x, ok := symtab.getSymbol("BEGIN"); ok {
		fmt.Printf("assm: %06d: BEGIN set to %06d\n", vm.pc, x.address)
		vm.pc = x.address
	} else {
		return vm, listing, fmt.Errorf("missing BEGIN")
	}

	return vm, listing, nil
}

// evaluate the of macro with the given arguments
func ofMacro(expr []string, env map[string]int) (int, error) {
	// convert the expression to postfix and evaluate it
	return postfix.EvalPostfix(postfix.FromInfix(expr), env)
}

func ParseOF(expr string) []string {
	var items []string
	var item string
	for ch, rest := expr[:1], expr[1:]; ch != ""; ch, rest = rest[:1], rest[1:] {
		switch ch {
		case "(", ")", "+", "*", "-":
			if item != "" {
				items = append(items, item)
			}
			items = append(items, ch)
			item = ""
		default:
			item = item + ch
		}

		if len(rest) == 0 {
			break
		}
	}
	return items
}

// pToCSL parses the parameters as comma separated list.
// todo: wonky workaround for quoted parameters
func pToCSL(parameters []string) []string {
	var s string
	for _, parm := range parameters {
		if len(parm) > 0 && parm[0] == '\'' {
			return parameters // dang quoted parameter found
		}
		s = s + parm
	}
	return strings.Split(s, ",")
}
