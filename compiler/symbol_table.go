package compiler

// SymbolScope represents the scope of a symbol
type SymbolScope string

const (
	GlobalScope  SymbolScope = "GLOBAL"
	LocalScope   SymbolScope = "LOCAL"
	FreeScope    SymbolScope = "FREE"
	BuiltinScope SymbolScope = "BUILTIN"
)

// Symbol represents a symbol in the symbol table
type Symbol struct {
	Name      string
	Scope     SymbolScope
	Index     int
	IsMutable bool
}

// SymbolTable represents a symbol table
type SymbolTable struct {
	outer *SymbolTable

	store          map[string]Symbol
	numDefinitions int

	FreeSymbols []Symbol
}

// NewSymbolTable creates a new symbol table
func NewSymbolTable() *SymbolTable {
	s := make(map[string]Symbol)
	free := []Symbol{}
	st := &SymbolTable{store: s, FreeSymbols: free}

	// Define built-in functions (must match order in vm/builtins.go)
	st.DefineBuiltin(0, "print")
	st.DefineBuiltin(1, "len")
	st.DefineBuiltin(2, "delete")
	st.DefineBuiltin(3, "append")
	st.DefineBuiltin(4, "keys")
	st.DefineBuiltin(5, "values")
	st.DefineBuiltin(6, "copy")
	st.DefineBuiltin(7, "enumName")
	st.DefineBuiltin(8, "enumValue")

	return st
}

// NewEnclosedSymbolTable creates a new enclosed symbol table
func NewEnclosedSymbolTable(outer *SymbolTable) *SymbolTable {
	s := NewSymbolTable()
	s.outer = outer
	return s
}

// Define defines a new symbol
func (st *SymbolTable) Define(name string) Symbol {
	return st.DefineWithMutability(name, true)
}

// DefineWithMutability defines a new symbol with specific mutability
func (st *SymbolTable) DefineWithMutability(name string, isMutable bool) Symbol {
	symbol := Symbol{
		Name:      name,
		Index:     st.numDefinitions,
		IsMutable: isMutable,
	}

	if st.outer == nil {
		symbol.Scope = GlobalScope
	} else {
		symbol.Scope = LocalScope
	}

	st.store[name] = symbol
	st.numDefinitions++
	return symbol
}

// Resolve resolves a symbol
func (st *SymbolTable) Resolve(name string) (Symbol, bool) {
	obj, ok := st.store[name]
	if !ok && st.outer != nil {
		obj, ok = st.outer.Resolve(name)
		if !ok {
			return obj, ok
		}

		if obj.Scope == GlobalScope {
			return obj, ok
		}

		free := st.defineFree(obj)
		return free, true
	}

	return obj, ok
}

// defineFree defines a free symbol
func (st *SymbolTable) defineFree(original Symbol) Symbol {
	st.FreeSymbols = append(st.FreeSymbols, original)

	symbol := Symbol{
		Name:      original.Name,
		Index:     len(st.FreeSymbols) - 1,
		Scope:     FreeScope,
		IsMutable: original.IsMutable,
	}

	st.store[original.Name] = symbol
	return symbol
}

// DefineBuiltin defines a built-in function
func (st *SymbolTable) DefineBuiltin(index int, name string) Symbol {
	symbol := Symbol{
		Name:      name,
		Index:     index,
		Scope:     BuiltinScope,
		IsMutable: false, // builtins are immutable
	}
	st.store[name] = symbol
	return symbol
}
