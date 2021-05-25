package compile

type SymbolScope string

const (
	// BuiltinScope is used for builtins functions
	BuiltinScope SymbolScope = "BUILTIN"
	// GlobalScope is used to for global variables
	GlobalScope SymbolScope = "GLOBAL"
	// LocalScope is used to for local variables
	LocalScope SymbolScope = "LOCAL"
	// FreeScope is used to for free variables(used in closures)
	FreeScope SymbolScope = "FREE"
	// FunctionScope is used to for function names
	FunctionScope SymbolScope = "FUNCTION"
)

type Symbol struct {
	Name  string
	Scope SymbolScope
	Index int
}

type SymbolTable struct {
	parent *SymbolTable

	store          map[string]Symbol
	numDefinitions int

	FreeSymbols []Symbol
}

func NewSymbolTable() *SymbolTable {
	return &SymbolTable{
		store: make(map[string]Symbol),
	}
}

func NewEnclosedSymbolTable(parent *SymbolTable) *SymbolTable {
	return &SymbolTable{
		parent: parent,
		store:  make(map[string]Symbol),
	}
}

func (st *SymbolTable) Define(ident string) Symbol {
	sym := Symbol{Name: ident, Index: st.numDefinitions}

	if st.parent == nil {
		sym.Scope = GlobalScope
	} else {
		sym.Scope = LocalScope
	}

	st.store[ident] = sym
	st.numDefinitions++

	return sym
}

func (st *SymbolTable) DefineBuiltin(index int, name string) Symbol {
	sym := Symbol{Name: name, Index: index, Scope: BuiltinScope}
	st.store[name] = sym
	return sym
}

func (st *SymbolTable) defineFree(original Symbol) Symbol {
	st.FreeSymbols = append(st.FreeSymbols, original)

	sym := Symbol{
		Name:  original.Name,
		Index: len(st.FreeSymbols) - 1,
		Scope: FreeScope,
	}

	st.store[original.Name] = sym
	return sym
}

func (st *SymbolTable) DefineFunctionName(name string) Symbol {
	sym := Symbol{Name: name, Scope: FunctionScope}
	st.store[name] = sym
	return sym
}

func (st *SymbolTable) Resolve(ident string) (sym Symbol, ok bool) {
	sym, ok = st.store[ident]
	if !ok && st.parent != nil {
		outerSym, outerOk := st.parent.Resolve(ident)
		if !outerOk {
			return outerSym, outerOk
		}

		if outerSym.Scope == GlobalScope || outerSym.Scope == BuiltinScope {
			return outerSym, outerOk
		}

		free := st.defineFree(outerSym)
		return free, true
	}

	return sym, ok
}
