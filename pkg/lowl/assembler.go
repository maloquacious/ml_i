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
			if n, err := nOf(line, opc, parameters, symtab.getConstants()); err != nil {
				return nil, listing, err
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
			n, err := nOf(line, opc, parameters, symtab.getConstants())
			if err != nil {
				return nil, listing, err
			}
			// set the operand to the literal value
			word.data = WORD(n)
			vm.core[vm.pc], vm.pc = word, vm.pc+1
		case "CAV":
			word := Word{op: op.CAV}
			// CAV * V,(A)
			// CAV * V,(X)
			if len(parameters) < 3 {
				return nil, listing, fmt.Errorf("%d: %s: want 2 args: got %d", line, opc, len(parameters))
			}
			vName, flag := parameters[0], parameters[2]
			switch flag {
			case "A": // compare to address
				symtab.refVariable(vName, vm.pc) // operand to be back-filled
			case "X": // compare to signed value
				symtab.refVariable(vName, vm.pc) // operand to be back-filled
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
			if len(parameters) < 3 {
				return nil, listing, fmt.Errorf("%d: %s: want 2 args: got %d", line, opc, len(parameters))
			}
			alias, sym := parameters[0], parameters[2]
			// add the alias to the symbol table.
			if err := symtab.defAlias(alias, sym, line); err != nil {
				return nil, listing, fmt.Errorf("%d: %s %q: %w", line, opc, alias, err)
			}
		case "EXIT":
			vm.core[vm.pc], vm.pc = Word{op: op.EXIT}, vm.pc+1
		case "FMOVE":
			vm.core[vm.pc], vm.pc = Word{op: op.FMOVE}, vm.pc+1
		case "FSTK":
			vm.core[vm.pc], vm.pc = Word{op: op.FSTK}, vm.pc+1
		case "GO":
			word := Word{op: op.GO}
			// GO label spec
			if len(parameters) < 1 {
				return nil, listing, fmt.Errorf("%d: %s: want 7 args: got %d", line, opc, len(parameters))
			}
			vName := parameters[0]
			if sym, ok := symtab.getUnaliasedSymbol(vName); !ok || sym.kind == SymIsBackfill {
				symtab.refVariable(vName, vm.pc) // operand to be back-filled
			} else if sym.kind == SymIsAddress {
				word.data = WORD(sym.address) // operand is the variable address
			} else {
				return nil, listing, fmt.Errorf("%d: %s %q: invalid", line, opc, vName)
			}
			vm.core[vm.pc], vm.pc = word, vm.pc+1
		case "GOADD":
			vm.core[vm.pc], vm.pc = Word{op: op.GOADD}, vm.pc+1
		case "GOEQ":
			word := Word{op: op.GOEQ}
			// GOEQ * label spec
			if len(parameters) < 1 {
				return nil, listing, fmt.Errorf("%d: %s: want 2 args: got %d", line, opc, len(parameters))
			}
			vName := parameters[0]
			if sym, ok := symtab.getUnaliasedSymbol(vName); !ok || sym.kind == SymIsBackfill {
				symtab.refVariable(vName, vm.pc) // operand to be back-filled
			} else if sym.kind == SymIsAddress {
				word.data = WORD(sym.address) // operand is the variable address
			} else {
				return nil, listing, fmt.Errorf("%d: %s %q: invalid", line, opc, vName)
			}
			vm.core[vm.pc], vm.pc = word, vm.pc+1
		case "GOGE":
			word := Word{op: op.GOGE}
			// GOGE * label spec
			if len(parameters) < 1 {
				return nil, listing, fmt.Errorf("%d: %s: want 2 args: got %d", line, opc, len(parameters))
			}
			vName := parameters[0]
			if sym, ok := symtab.getUnaliasedSymbol(vName); !ok || sym.kind == SymIsBackfill {
				symtab.refVariable(sym.name, vm.pc) // operand to be back-filled
			} else if sym.kind == SymIsAddress {
				word.data = WORD(sym.address) // operand is the variable address
			} else {
				return nil, listing, fmt.Errorf("%d: %s %q: invalid", line, opc, vName)
			}
			vm.core[vm.pc], vm.pc = word, vm.pc+1
		case "GOGR":
			word := Word{op: op.GOGR}
			// GOGR * label spec
			if len(parameters) < 1 {
				return nil, listing, fmt.Errorf("%d: %s: want 2 args: got %d", line, opc, len(parameters))
			}
			vName := parameters[0]
			if sym, ok := symtab.getUnaliasedSymbol(vName); !ok || sym.kind == SymIsBackfill {
				symtab.refVariable(sym.name, vm.pc) // operand to be back-filled
			} else if sym.kind == SymIsAddress {
				word.data = WORD(sym.address) // operand is the variable address
			} else {
				return nil, listing, fmt.Errorf("%d: %s %q: invalid", line, opc, vName)
			}
			vm.core[vm.pc], vm.pc = word, vm.pc+1
		case "GOLE":
			word := Word{op: op.GOLE}
			// GOLE * label spec
			if len(parameters) < 1 {
				return nil, listing, fmt.Errorf("%d: %s: want 2 args: got %d", line, opc, len(parameters))
			}
			vName := parameters[0]
			if sym, ok := symtab.getUnaliasedSymbol(vName); !ok || sym.kind == SymIsBackfill {
				symtab.refVariable(sym.name, vm.pc) // operand to be back-filled
			} else if sym.kind == SymIsAddress {
				word.data = WORD(sym.address) // operand is the variable address
			} else {
				return nil, listing, fmt.Errorf("%d: %s %q: invalid", line, opc, vName)
			}
			vm.core[vm.pc], vm.pc = word, vm.pc+1
		case "GOLT":
			word := Word{op: op.GOLT}
			// GOLT * label spec
			if len(parameters) < 1 {
				return nil, listing, fmt.Errorf("%d: %s: want 2 args: got %d", line, opc, len(parameters))
			}
			vName := parameters[0]
			if sym, ok := symtab.getUnaliasedSymbol(vName); !ok || sym.kind == SymIsBackfill {
				symtab.refVariable(sym.name, vm.pc) // operand to be back-filled
			} else if sym.kind == SymIsAddress {
				word.data = WORD(sym.address) // operand is the variable address
			} else {
				return nil, listing, fmt.Errorf("%d: %s %q: invalid", line, opc, vName)
			}
			vm.core[vm.pc], vm.pc = word, vm.pc+1
		case "GOND":
			vm.core[vm.pc], vm.pc = Word{op: op.GOND}, vm.pc+1
		case "GONE":
			word := Word{op: op.GONE}
			// GONE * label spec
			if len(parameters) < 1 {
				return nil, listing, fmt.Errorf("%d: %s: want 2 args: got %d", line, opc, len(parameters))
			}
			vName := parameters[0]
			if sym, ok := symtab.getUnaliasedSymbol(vName); !ok || sym.kind == SymIsBackfill {
				symtab.refVariable(sym.name, vm.pc) // operand to be back-filled
			} else if sym.kind == SymIsAddress {
				word.data = WORD(sym.address) // operand is the variable address
			} else {
				return nil, listing, fmt.Errorf("%d: %s %q: invalid", line, opc, vName)
			}
			vm.core[vm.pc], vm.pc = word, vm.pc+1
		case "GOPC":
			vm.core[vm.pc], vm.pc = Word{op: op.GOPC}, vm.pc+1
		case "GOSUB":
			word := Word{op: op.GOSUB}
			// GOSUB subroutine name,(distance)
			// GOSUB subroutine name,(X)
			if len(parameters) < 3 {
				return nil, listing, fmt.Errorf("%d: %s: want 2 args: got %d", line, opc, len(parameters))
			}
			vName, flag := parameters[0], parameters[2]
			switch flag {
			case "X": // reference a routine in the MD-logic
				symtab.refVariable(vName, vm.pc) // operand to be back-filled
			default: // flag must be a number
				if flag == "-" { // might be a negative number
					for _, parm := range parameters[3:] {
						flag = flag + parm
					}
				}
				if _, err := strconv.Atoi(flag); err != nil {
					return nil, listing, fmt.Errorf("%d: %s: unknown flag %q", line, opc, flag)
				}
				symtab.refVariable(vName, vm.pc) // operand to be back-filled
			}
			vm.core[vm.pc], vm.pc = word, vm.pc+1
		case "IDENT":
			// IDENT V,decimal integer
			if len(parameters) < 3 {
				return nil, listing, fmt.Errorf("%d: %s: want 2 args: got %d", line, opc, len(parameters))
			}
			vName, number := parameters[0], parameters[2]
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
			if len(parameters) < 3 {
				return nil, listing, fmt.Errorf("%d: %s: want 2 args: got %d", line, opc, len(parameters))
			}
			vName, flag := parameters[0], parameters[2]
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
			n, err := nOf(line, opc, parameters, symtab.getConstants())
			if err != nil {
				return nil, listing, err
			}
			word.data = WORD(n)
			vm.core[vm.pc], vm.pc = word, vm.pc+1
		case "LAM":
			vm.core[vm.pc], vm.pc = Word{op: op.LAM}, vm.pc+1
		case "LAV":
			word := Word{op: op.LAV}
			// LAV V,(X)
			// LAV V,(R)
			if len(parameters) < 3 {
				return nil, listing, fmt.Errorf("%d: %s: want 2 args: got %d", line, opc, len(parameters))
			}
			vName, flag := parameters[0], parameters[2]
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
			if len(parameters) < 3 {
				return nil, listing, fmt.Errorf("%d: %s: want 2 args: got %d", line, opc, len(parameters))
			}
			vName, flag := parameters[0], parameters[2]
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
			if n, err := nOf(line, opc, parameters, symtab.getConstants()); err != nil {
				return nil, listing, err
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
			vName, flag := parameters[0], parameters[2]
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
			vm.core[vm.pc], vm.pc = Word{op: op.SUBR}, vm.pc+1
			panic("SUBR must be implemented!")
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
				lines = append(lines, fmt.Sprintf("%-15s  defn %4d    alias-of %s", name, value.line, value.aliasOf))
			case SymIsAddress:
				lines = append(lines, fmt.Sprintf("%-15s  defn %4d     address %8d", name, value.line, value.address))
			case SymIsBackfill:
				lines = append(lines, fmt.Sprintf("%-15s  defn %4d   back-fill", name, value.line))
			case SymIsOnHeap:
				lines = append(lines, fmt.Sprintf("%-15s  defn %4d  heap-index %8d", name, value.line, value.heapIndex))
			case SymIsValue:
				lines = append(lines, fmt.Sprintf("%-15s  defn %4d       value %8d", name, value.line, value.value))
			case SymIsUnknown:
				lines = append(lines, fmt.Sprintf("%-15s  defn %4d   undefined", name, value.line))
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

// nOf is a helper to evaluate a literal or OF macro
func nOf(line int, opc string, parameters []string, env map[string]int) (n int, err error) {
	if parameters[0] == "OF" {
		args := parameters[1:]
		if args[0] != "(" {
			return 0, fmt.Errorf("%d: %s: of: missing left paren", line, opc)
		} else if args[len(args)-1] != ")" {
			return 0, fmt.Errorf("%d: %s: of: missing right paren", line, opc)
		}
		return ofMacro(args, env)
	}

	// maybe a constant
	if n, ok := env[parameters[0]]; ok {
		return n, nil
	}

	var number string
	for _, parm := range parameters {
		number = number + parm
	}
	if n, err = strconv.Atoi(number); err != nil {
		return 0, fmt.Errorf("%d: %s %q: %w", line, opc, number, err)
	}
	return n, nil
}
