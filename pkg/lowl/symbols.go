package lowl

import "fmt"

type SymbolTable struct {
	table    map[string]SYMBOL
	backfill map[string][]ADDR
	env      map[string]int
}

type SYMBOL struct {
	name string     // name of the symbol
	kind SYMBOLKIND // kind of the symbol
	line int        // line in listing with definition
	// address in memory for labels
	address ADDR
	// alias information
	aliasOf string
	// INDEX in the heap for variables
	heapIndex int
	// value for constants
	value WORD
}

// SYMBOLKIND is an enum for the kind of symbol in the table
type SYMBOLKIND int

// enums for SYMBOLKIND
const (
	SymIsUnknown  SYMBOLKIND = iota // never seen, never defined
	SymIsAddress                    // address in memory
	SymIsAlias                      // alias for another symbol
	SymIsBackfill                   // used but not yet defined
	SymIsOnHeap                     // address on the heap
	SymIsValue                      // a constant value
)

func newSymbolTable() *SymbolTable {
	st := &SymbolTable{}
	st.table = make(map[string]SYMBOL)
	st.backfill = make(map[string][]ADDR)
	st.env = make(map[string]int)

	return st
}

// defAlias defines a new alias for an existing symbol.
// the symbol must be defined in the symbol table.
// the new alias must not be defined.
// it is okay for it to be in the back-fill table.
func (st *SymbolTable) defAlias(aliasName, symName string, line int) error {
	// alias must not be in the table
	if _, ok := st.table[aliasName]; ok {
		return fmt.Errorf("%q redefined", aliasName)
	}
	// symbol must be in the table as a variable
	if sym, ok := st.table[symName]; !ok {
		return fmt.Errorf("%q undefined", symName)
	} else if sym.kind != SymIsOnHeap {
		return fmt.Errorf("%q not variable", symName)
	}

	// add the alias to the table
	st.table[aliasName] = SYMBOL{
		name:    aliasName,
		kind:    SymIsAlias,
		aliasOf: symName,
	}

	return nil
}

// defConstant defines a new constant
func (st *SymbolTable) defConstant(name string, line int, value WORD) error {
	if _, ok := st.table[name]; ok {
		return fmt.Errorf("redefined")
	}
	st.table[name] = SYMBOL{
		name:  name,
		line:  line,
		kind:  SymIsValue,
		value: value,
	}
	st.env[name] = int(value)
	return nil
}

// defLabel defines a new label
func (st *SymbolTable) defLabel(name string, line int, addr ADDR) error {
	if _, ok := st.table[name]; ok {
		return fmt.Errorf("redefined")
	}
	st.table[name] = SYMBOL{
		name:    name,
		line:    line,
		kind:    SymIsAddress,
		address: addr,
	}
	return nil
}

// defVariable defines a new variable
func (st *SymbolTable) defVariable(name string, line int, vm *VM) error {
	if _, ok := st.table[name]; ok {
		return fmt.Errorf("redefined")
	}

	// allocate space for the variable on the heap
	heapIndex := ADDR(len(vm.heaps.vars))
	vm.heaps.vars = append(vm.heaps.vars, 0)

	st.table[name] = SYMBOL{
		name:    name,
		line:    line,
		kind:    SymIsOnHeap,
		address: heapIndex,
	}

	return nil
}

// getConstant returns the current value of a constant value
func (st *SymbolTable) getConstant(name string) (int, bool) {
	v, ok := st.env[name]
	return v, ok
}

// getConstants returns the current set of constant values
func (st *SymbolTable) getConstants() map[string]int {
	return st.env
}

// getSymbol returns the symbol from the table
func (st *SymbolTable) getSymbol(name string) (SYMBOL, bool) {
	if sym, ok := st.table[name]; ok {
		return sym, ok
	} else if _, ok = st.backfill[name]; ok {
		return SYMBOL{
			name: name,
			kind: SymIsBackfill,
		}, ok
	}
	// we haven't seen that name yet
	return SYMBOL{
		name: name,
		kind: SymIsUnknown,
	}, false
}

// getUnaliasedSymbol returns the symbol from the table.
// if the symbol is an alias, it returns the alias-of symbol.
func (st *SymbolTable) getUnaliasedSymbol(name string) (SYMBOL, bool) {
	sym, ok := st.getSymbol(name)
	if sym.kind == SymIsAlias {
		sym, ok = st.getSymbol(sym.aliasOf)
		if sym.kind == SymIsAlias {
			sym, ok = st.getSymbol(sym.aliasOf)
			if sym.kind == SymIsAlias {
				panic(fmt.Sprintf("%q: too much recursion", name))
			}
		}
	}
	return sym, ok
}

// refVariable creates a back-fill entry for a variable in the symbol table.
func (st *SymbolTable) refVariable(name string, addr ADDR) {
	st.backfill[name] = append(st.backfill[name], addr)
}
