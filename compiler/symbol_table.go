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
	st.DefineBuiltin(9, "abs")
	st.DefineBuiltin(10, "min")
	st.DefineBuiltin(11, "max")
	st.DefineBuiltin(12, "sqrt")
	st.DefineBuiltin(13, "pow")
	st.DefineBuiltin(14, "floor")
	st.DefineBuiltin(15, "ceil")
	st.DefineBuiltin(16, "split")
	st.DefineBuiltin(17, "substring")
	st.DefineBuiltin(18, "int")
	st.DefineBuiltin(19, "float")
	st.DefineBuiltin(20, "string")
	// String functions
	st.DefineBuiltin(21, "trim")
	st.DefineBuiltin(22, "upper")
	st.DefineBuiltin(23, "lower")
	st.DefineBuiltin(24, "contains")
	st.DefineBuiltin(25, "indexOf")
	st.DefineBuiltin(26, "replace")
	st.DefineBuiltin(27, "startsWith")
	st.DefineBuiltin(28, "endsWith")
	st.DefineBuiltin(29, "join")
	st.DefineBuiltin(30, "repeat")
	st.DefineBuiltin(31, "char")
	st.DefineBuiltin(32, "ord")
	// Console I/O
	st.DefineBuiltin(33, "write")
	st.DefineBuiltin(34, "eprint")
	st.DefineBuiltin(35, "ewrite")
	st.DefineBuiltin(36, "readln")
	// File I/O
	st.DefineBuiltin(37, "fopen")
	st.DefineBuiltin(38, "fclose")
	st.DefineBuiltin(39, "fread")
	st.DefineBuiltin(40, "fwrite")
	st.DefineBuiltin(41, "fwriteln")
	st.DefineBuiltin(42, "feof")
	st.DefineBuiltin(43, "fseek")
	st.DefineBuiltin(44, "ftell")
	st.DefineBuiltin(45, "fsize")
	st.DefineBuiltin(46, "readFile")
	st.DefineBuiltin(47, "writeFile")
	// Utility
	st.DefineBuiltin(48, "typeof")
	st.DefineBuiltin(49, "exit")

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
