package compiler

import (
	"fmt"
	"minlang/ast"
	"minlang/vm"
)

// LoopContext tracks information about the current loop
type LoopContext struct {
	breakJumps    []int // Positions of break jumps to patch
	continueJumps []int // Positions of continue jumps to patch
}

// EnumType tracks enum type information
type EnumType struct {
	Name         string
	Variants     map[string]int // variant name -> integer value
	VariantNames []string       // ordered variant names
}

// StructType tracks struct type information
type StructType struct {
	Name   string
	Fields map[string]string // field name -> field type
}

// Compiler represents the compiler
type Compiler struct {
	constants   []vm.Value
	symbolTable *SymbolTable

	scopes     []CompilationScope
	scopeIndex int

	typeChecker   *TypeChecker
	typeCheckMode bool

	loopStack   []LoopContext         // Stack of loop contexts
	enumTypes   map[string]*EnumType  // Tracks enum type definitions
	structTypes map[string]*StructType // Tracks struct type definitions
}

// CompilationScope represents a compilation scope
type CompilationScope struct {
	instructions vm.Instruction
	lastInstruction EmittedInstruction
	previousInstruction EmittedInstruction
}

// EmittedInstruction tracks the last emitted instruction
type EmittedInstruction struct {
	Opcode   vm.OpCode
	Position int
}

// New creates a new compiler
func New() *Compiler {
	mainScope := CompilationScope{
		instructions: vm.Instruction{},
		lastInstruction: EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}

	symbolTable := NewSymbolTable()

	return &Compiler{
		constants:   []vm.Value{},
		symbolTable: symbolTable,
		scopes:      []CompilationScope{mainScope},
		scopeIndex:  0,
		loopStack:   []LoopContext{},
		enumTypes:   make(map[string]*EnumType),
		structTypes: make(map[string]*StructType),
	}
}

// enterLoop pushes a new loop context
func (c *Compiler) enterLoop() {
	c.loopStack = append(c.loopStack, LoopContext{
		breakJumps:    []int{},
		continueJumps: []int{},
	})
}

// leaveLoop pops a loop context
func (c *Compiler) leaveLoop() {
	c.loopStack = c.loopStack[:len(c.loopStack)-1]
}

// currentLoop returns the current loop context
func (c *Compiler) currentLoop() *LoopContext {
	if len(c.loopStack) == 0 {
		return nil
	}
	return &c.loopStack[len(c.loopStack)-1]
}

// Bytecode returns the compiled bytecode
func (c *Compiler) Bytecode() *vm.Bytecode {
	return &vm.Bytecode{
		Instructions: c.currentInstructions(),
		Constants:    c.constants,
	}
}

func (c *Compiler) currentInstructions() vm.Instruction {
	return c.scopes[c.scopeIndex].instructions
}

func (c *Compiler) emit(op vm.OpCode, operands ...int) int {
	ins := vm.Make(op, operands...)
	pos := c.addInstruction(ins)

	c.setLastInstruction(op, pos)

	return pos
}

func (c *Compiler) addInstruction(ins []byte) int {
	posNewInstruction := len(c.currentInstructions())
	updatedInstructions := append(c.currentInstructions(), ins...)

	c.scopes[c.scopeIndex].instructions = updatedInstructions

	return posNewInstruction
}

func (c *Compiler) setLastInstruction(op vm.OpCode, pos int) {
	previous := c.scopes[c.scopeIndex].lastInstruction
	last := EmittedInstruction{Opcode: op, Position: pos}

	c.scopes[c.scopeIndex].previousInstruction = previous
	c.scopes[c.scopeIndex].lastInstruction = last
}

func (c *Compiler) lastInstructionIs(op vm.OpCode) bool {
	if len(c.currentInstructions()) == 0 {
		return false
	}

	return c.scopes[c.scopeIndex].lastInstruction.Opcode == op
}

func (c *Compiler) removeLastPop() {
	last := c.scopes[c.scopeIndex].lastInstruction
	previous := c.scopes[c.scopeIndex].previousInstruction

	old := c.currentInstructions()
	new := old[:last.Position]

	c.scopes[c.scopeIndex].instructions = new
	c.scopes[c.scopeIndex].lastInstruction = previous
}

func (c *Compiler) replaceInstruction(pos int, newInstruction []byte) {
	ins := c.currentInstructions()

	for i := 0; i < len(newInstruction); i++ {
		ins[pos+i] = newInstruction[i]
	}
}

func (c *Compiler) changeOperand(opPos int, operand int) {
	op := vm.OpCode(c.currentInstructions()[opPos])
	newInstruction := vm.Make(op, operand)

	c.replaceInstruction(opPos, newInstruction)
}

func (c *Compiler) addConstant(obj vm.Value) int {
	c.constants = append(c.constants, obj)
	return len(c.constants) - 1
}

// tryEmitDirectLocalOp attempts to optimize binary operations with local variables
// If the last instruction was OpLoadLocal, it replaces it with a direct local operation
func (c *Compiler) tryEmitDirectLocalOp(normalOp, directLocalOp vm.OpCode) {
	// Check if last instruction was OpLoadLocal
	if c.lastInstructionIs(vm.OpLoadLocal) {
		// Get the position and extract the local index
		lastPos := c.scopes[c.scopeIndex].lastInstruction.Position
		ins := c.currentInstructions()

		// Extract the local index from the OpLoadLocal instruction
		localIndex, _ := vm.ReadOperand(ins, lastPos+1)

		// Replace OpLoadLocal with the direct local operation in place
		newIns := vm.Make(directLocalOp, localIndex)
		for i := 0; i < len(newIns); i++ {
			ins[lastPos+i] = newIns[i]
		}

		// Update the last instruction opcode
		c.scopes[c.scopeIndex].lastInstruction.Opcode = directLocalOp
	} else {
		// No optimization possible, emit normal operation
		c.emit(normalOp)
	}
}

func (c *Compiler) enterScope() {
	scope := CompilationScope{
		instructions:        vm.Instruction{},
		lastInstruction:     EmittedInstruction{},
		previousInstruction: EmittedInstruction{},
	}
	c.scopes = append(c.scopes, scope)
	c.scopeIndex++

	c.symbolTable = NewEnclosedSymbolTable(c.symbolTable)
}

func (c *Compiler) leaveScope() vm.Instruction {
	instructions := c.currentInstructions()

	c.scopes = c.scopes[:len(c.scopes)-1]
	c.scopeIndex--

	c.symbolTable = c.symbolTable.outer

	return instructions
}

// Compile compiles an AST node
func (c *Compiler) Compile(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}

	case *ast.ExpressionStatement:
		err := c.Compile(node.Expression)
		if err != nil {
			return err
		}
		c.emit(vm.OpPop)

	case *ast.InfixExpression:
		// Handle comparison operators with special ordering
		if node.Operator == "<" {
			err := c.Compile(node.Right)
			if err != nil {
				return err
			}

			err = c.Compile(node.Left)
			if err != nil {
				return err
			}

			c.emit(vm.OpGt)
			return nil
		}

		if node.Operator == "<=" {
			err := c.Compile(node.Right)
			if err != nil {
				return err
			}

			err = c.Compile(node.Left)
			if err != nil {
				return err
			}

			c.emit(vm.OpGe)
			return nil
		}

		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		err = c.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "+":
			c.tryEmitDirectLocalOp(vm.OpAdd, vm.OpAddLocal)
		case "-":
			c.tryEmitDirectLocalOp(vm.OpSub, vm.OpSubLocal)
		case "*":
			c.tryEmitDirectLocalOp(vm.OpMul, vm.OpMulLocal)
		case "/":
			c.tryEmitDirectLocalOp(vm.OpDiv, vm.OpDivLocal)
		case "%":
			c.emit(vm.OpMod)
		case ">":
			c.emit(vm.OpGt)
		case ">=":
			c.emit(vm.OpGe)
		case "==":
			c.emit(vm.OpEq)
		case "!=":
			c.emit(vm.OpNe)
		case "&&":
			c.emit(vm.OpAnd)
		case "||":
			c.emit(vm.OpOr)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}

	case *ast.PrefixExpression:
		err := c.Compile(node.Right)
		if err != nil {
			return err
		}

		switch node.Operator {
		case "!":
			c.emit(vm.OpNot)
		case "-":
			c.emit(vm.OpNeg)
		default:
			return fmt.Errorf("unknown operator %s", node.Operator)
		}

	case *ast.IntegerLiteral:
		integer := vm.IntValue(node.Value)
		c.emit(vm.OpPush, c.addConstant(integer))

	case *ast.FloatLiteral:
		float := vm.FloatValue(node.Value)
		c.emit(vm.OpPush, c.addConstant(float))

	case *ast.BooleanLiteral:
		if node.Value {
			c.emit(vm.OpPush, c.addConstant(vm.BoolValue(true)))
		} else {
			c.emit(vm.OpPush, c.addConstant(vm.BoolValue(false)))
		}

	case *ast.StringLiteral:
		str := vm.StringValue(node.Value)
		c.emit(vm.OpPush, c.addConstant(str))

	case *ast.NilLiteral:
		c.emit(vm.OpPush, c.addConstant(vm.NilValue()))

	case *ast.IfStatement:
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}

		// Emit jump instruction with placeholder
		jumpNotTruthyPos := c.emit(vm.OpJumpIfFalse, 9999)

		err = c.Compile(node.Consequence)
		if err != nil {
			return err
		}

		// Emit jump to skip alternative
		jumpPos := c.emit(vm.OpJump, 9999)

		afterConsequencePos := len(c.currentInstructions())
		c.changeOperand(jumpNotTruthyPos, afterConsequencePos)

		if node.Alternative != nil {
			err := c.Compile(node.Alternative)
			if err != nil {
				return err
			}
		}

		afterAlternativePos := len(c.currentInstructions())
		c.changeOperand(jumpPos, afterAlternativePos)

	case *ast.BlockStatement:
		for _, s := range node.Statements {
			err := c.Compile(s)
			if err != nil {
				return err
			}
		}

	case *ast.VarStatement:
		symbol := c.symbolTable.DefineWithMutability(node.Name.Value, node.IsMutable)

		if node.Value != nil {
			err := c.Compile(node.Value)
			if err != nil {
				return err
			}
		} else {
			// Default to nil if no value provided
			c.emit(vm.OpPush, c.addConstant(vm.NilValue()))
		}

		if symbol.Scope == GlobalScope {
			c.emit(vm.OpStoreGlobal, symbol.Index)
		} else {
			c.emit(vm.OpStoreLocal, symbol.Index)
		}

	case *ast.Identifier:
		symbol, ok := c.symbolTable.Resolve(node.Value)
		if !ok {
			return fmt.Errorf("undefined variable %s", node.Value)
		}

		c.loadSymbol(symbol)

	case *ast.AssignmentStatement:
		// Handle different types of left-hand sides
		switch left := node.Left.(type) {
		case *ast.Identifier:
			// Check if variable exists and is mutable
			symbol, ok := c.symbolTable.Resolve(left.Value)
			if !ok {
				return fmt.Errorf("undefined variable %s", left.Value)
			}

			if !symbol.IsMutable {
				return fmt.Errorf("cannot assign to const variable %s", left.Value)
			}

			// Compile the value
			err := c.Compile(node.Value)
			if err != nil {
				return err
			}

			c.storeSymbol(symbol)

		case *ast.IndexExpression:
			// For array[index] = value
			// Stack layout: array, index, value

			// Compile the array/map
			err := c.Compile(left.Left)
			if err != nil {
				return err
			}

			// Compile the index
			err = c.Compile(left.Index)
			if err != nil {
				return err
			}

			// Compile the value
			err = c.Compile(node.Value)
			if err != nil {
				return err
			}

			// Emit set operation (runtime will determine if array or map)
			c.emit(vm.OpArraySet)

		case *ast.FieldAccessExpression:
			// For struct.field = value
			// Stack layout: struct, fieldName, value

			// Compile the struct
			err := c.Compile(left.Left)
			if err != nil {
				return err
			}

			// Push field name
			c.emit(vm.OpPush, c.addConstant(vm.StringValue(left.Field.Value)))

			// Compile the value
			err = c.Compile(node.Value)
			if err != nil {
				return err
			}

			// Emit set field operation
			c.emit(vm.OpSetField)

		default:
			return fmt.Errorf("unsupported assignment target")
		}

	case *ast.TypeStatement:
		// Handle type definitions
		switch def := node.Definition.(type) {
		case *ast.EnumStatement:
			// Set the name from the TypeStatement
			def.Name = node.Name
			return c.Compile(def)
		case *ast.StructStatement:
			// Set the name from the TypeStatement
			def.Name = node.Name

			// Register struct type
			structType := &StructType{
				Name:   node.Name.Value,
				Fields: make(map[string]string),
			}

			// Store field types
			for _, field := range def.Fields {
				structType.Fields[field.Name.Value] = field.Type.String()
			}

			c.structTypes[node.Name.Value] = structType

			// Structs don't need runtime code generation
			return nil
		}

	case *ast.EnumStatement:
		// Register enum type
		enumType := &EnumType{
			Name:         node.Name.Value,
			Variants:     make(map[string]int),
			VariantNames: make([]string, len(node.Variants)),
		}

		// Assign integer values to variants (0, 1, 2, ...)
		for i, variant := range node.Variants {
			enumType.Variants[variant.Value] = i
			enumType.VariantNames[i] = variant.Value

			// Define variant as a constant in the symbol table
			symbol := c.symbolTable.DefineWithMutability(variant.Value, false)

			// Push the integer value
			c.emit(vm.OpPush, c.addConstant(vm.IntValue(int64(i))))

			// Store it
			if symbol.Scope == GlobalScope {
				c.emit(vm.OpStoreGlobal, symbol.Index)
			} else {
				c.emit(vm.OpStoreLocal, symbol.Index)
			}
		}

		// Store enum type info
		c.enumTypes[node.Name.Value] = enumType

		// Register in VM runtime registry
		vm.EnumRegistry[node.Name.Value] = make(map[int]string)
		for value, name := range enumType.VariantNames {
			vm.EnumRegistry[node.Name.Value][value] = name
		}

	case *ast.FunctionStatement:
		// Define the function name in the current scope BEFORE compiling the body
		// This allows recursive calls
		symbol := c.symbolTable.Define(node.Name.Value)

		c.enterScope()

		// Define parameters in the new scope
		for _, param := range node.Parameters {
			c.symbolTable.Define(param.Name.Value)
		}

		err := c.Compile(node.Body)
		if err != nil {
			return err
		}

		// If the last instruction is not a return, add an implicit return nil
		if !c.lastInstructionIs(vm.OpReturn) {
			c.emit(vm.OpPush, c.addConstant(vm.NilValue()))
			c.emit(vm.OpReturn)
		}

		// Get the compiled instructions
		freeSymbols := c.symbolTable.FreeSymbols
		numLocals := c.symbolTable.numDefinitions
		instructions := c.leaveScope()

		// Create the function object
		compiledFn := &vm.Function{
			Name:         node.Name.Value,
			NumParams:    len(node.Parameters),
			NumLocals:    numLocals,
			Instructions: instructions,
		}

		// If there are free variables, create a closure
		if len(freeSymbols) > 0 {
			for _, s := range freeSymbols {
				c.loadSymbol(s)
			}
			fnIndex := c.addConstant(vm.NewFunctionValue(compiledFn))
			c.emit(vm.OpMakeClosure, fnIndex, len(freeSymbols))
		} else {
			fnIndex := c.addConstant(vm.NewFunctionValue(compiledFn))
			c.emit(vm.OpPush, fnIndex)
		}

		// Store the function value
		c.storeSymbol(symbol)

	case *ast.ReturnStatement:
		if node.ReturnValue != nil {
			err := c.Compile(node.ReturnValue)
			if err != nil {
				return err
			}
		} else {
			c.emit(vm.OpPush, c.addConstant(vm.NilValue()))
		}

		c.emit(vm.OpReturn)

	case *ast.BreakStatement:
		loop := c.currentLoop()
		if loop == nil {
			return fmt.Errorf("break statement outside of loop")
		}
		// Emit a jump with placeholder address
		pos := c.emit(vm.OpJump, 9999)
		// Record this position so we can patch it later
		loop.breakJumps = append(loop.breakJumps, pos)

	case *ast.ContinueStatement:
		loop := c.currentLoop()
		if loop == nil {
			return fmt.Errorf("continue statement outside of loop")
		}
		// Emit a jump with placeholder address
		pos := c.emit(vm.OpJump, 9999)
		// Record this position so we can patch it later
		loop.continueJumps = append(loop.continueJumps, pos)

	case *ast.CallExpression:
		err := c.Compile(node.Function)
		if err != nil {
			return err
		}

		for _, arg := range node.Arguments {
			err := c.Compile(arg)
			if err != nil {
				return err
			}
		}

		c.emit(vm.OpCall, len(node.Arguments))

	case *ast.ArrayLiteral:
		// Compile each element
		for _, el := range node.Elements {
			err := c.Compile(el)
			if err != nil {
				return err
			}
		}

		// Emit OpArray with number of elements
		c.emit(vm.OpArray, len(node.Elements))

	case *ast.MapLiteral:
		// Compile each key-value pair
		for key, value := range node.Pairs {
			err := c.Compile(key)
			if err != nil {
				return err
			}
			err = c.Compile(value)
			if err != nil {
				return err
			}
		}

		// Emit OpMap with number of pairs
		c.emit(vm.OpMap, len(node.Pairs))

	case *ast.StructLiteral:
		// Compile each field first (they'll be popped in reverse order)
		for fieldName, value := range node.Fields {
			// Push field name
			c.emit(vm.OpPush, c.addConstant(vm.StringValue(fieldName)))
			// Push field value
			err := c.Compile(value)
			if err != nil {
				return err
			}
		}

		// Push the type name as a string (last, will be popped first)
		c.emit(vm.OpPush, c.addConstant(vm.StringValue(node.Name.Value)))

		// Emit OpStruct with number of fields
		c.emit(vm.OpStruct, len(node.Fields))

	case *ast.IndexExpression:
		// Compile the array/map expression
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		// Compile the index expression
		err = c.Compile(node.Index)
		if err != nil {
			return err
		}

		// Emit get operation
		// We'll determine at runtime if it's array or map
		c.emit(vm.OpArrayGet)

	case *ast.FieldAccessExpression:
		// Compile the struct expression
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		// Push the field name as a string
		c.emit(vm.OpPush, c.addConstant(vm.StringValue(node.Field.Value)))

		// Emit get field operation
		c.emit(vm.OpGetField)

	case *ast.SwitchStatement:
		// Compile the switch value
		err := c.Compile(node.Value)
		if err != nil {
			return err
		}

		// We'll compile switch as a series of comparisons and jumps
		// For each case:
		//   1. Duplicate switch value on stack
		//   2. Push case value
		//   3. Compare (OpEq)
		//   4. Jump to case body if true
		//   5. Otherwise continue to next case

		jumpToEnd := []int{}        // Collect jumps to end of switch
		jumpToCaseBody := []int{}  // Jumps to case bodies

		for _, caseClause := range node.Cases {
			// Duplicate switch value for comparison
			c.emit(vm.OpDup)

			// Compile case value
			err := c.Compile(caseClause.Value)
			if err != nil {
				return err
			}

			// Compare
			c.emit(vm.OpEq)

			// Jump to case body if equal (placeholder)
			// OpJumpIfTrue will pop the comparison result
			jumpIfTrue := c.emit(vm.OpJumpIfTrue, 9999)
			jumpToCaseBody = append(jumpToCaseBody, jumpIfTrue)

			// Note: OpJumpIfTrue already popped the comparison result
		}

		// If no cases matched, jump to default or end
		jumpToDefaultOrEnd := c.emit(vm.OpJump, 9999)

		// Compile case bodies
		caseBodyPositions := []int{}
		for i, caseClause := range node.Cases {
			// Record position of this case body
			caseBodyPos := len(c.currentInstructions())
			caseBodyPositions = append(caseBodyPositions, caseBodyPos)

			// Patch the jump for this case
			c.changeOperand(jumpToCaseBody[i], caseBodyPos)

			// Pop the switch value (OpJumpIfTrue already popped the comparison result)
			c.emit(vm.OpPop)

			// Compile case body
			err := c.Compile(caseClause.Body)
			if err != nil {
				return err
			}

			// Jump to end after case body
			jumpToEnd = append(jumpToEnd, c.emit(vm.OpJump, 9999))
		}

		// Default case
		defaultPos := len(c.currentInstructions())
		c.changeOperand(jumpToDefaultOrEnd, defaultPos)

		if node.Default != nil {
			// Pop the switch value
			c.emit(vm.OpPop)

			err := c.Compile(node.Default)
			if err != nil {
				return err
			}
		} else {
			// No default, just pop the switch value
			c.emit(vm.OpPop)
		}

		// Patch all jumps to end
		endPos := len(c.currentInstructions())
		for _, jumpPos := range jumpToEnd {
			c.changeOperand(jumpPos, endPos)
		}

	case *ast.ForStatement:
		// Enter loop context for break/continue
		c.enterLoop()
		defer c.leaveLoop()

		// Compile initialization if present
		if node.Init != nil {
			err := c.Compile(node.Init)
			if err != nil {
				return err
			}
		}

		// Mark the start of the loop (where continue jumps to)
		loopStart := len(c.currentInstructions())

		// Compile the condition
		if node.Condition == nil {
			return fmt.Errorf("for loop must have a condition")
		}
		err := c.Compile(node.Condition)
		if err != nil {
			return err
		}

		// Jump to end if condition is false (placeholder address)
		jumpToEnd := c.emit(vm.OpJumpIfFalse, 9999)

		// Compile the loop body
		err = c.Compile(node.Body)
		if err != nil {
			return err
		}

		// Mark where continue should jump (before post statement)
		continuePos := len(c.currentInstructions())

		// Compile the post statement if present
		if node.Post != nil {
			err = c.Compile(node.Post)
			if err != nil {
				return err
			}
		}

		// Jump back to the start of the loop
		c.emit(vm.OpJump, loopStart)

		// Patch the jump to end address (where break jumps to)
		loopEnd := len(c.currentInstructions())
		c.changeOperand(jumpToEnd, loopEnd)

		// Patch all break jumps to jump to loopEnd
		loop := c.currentLoop()
		for _, breakPos := range loop.breakJumps {
			c.changeOperand(breakPos, loopEnd)
		}

		// Patch all continue jumps to jump to continuePos
		for _, contPos := range loop.continueJumps {
			c.changeOperand(contPos, continuePos)
		}

	}

	return nil
}

func (c *Compiler) loadSymbol(s Symbol) {
	switch s.Scope {
	case GlobalScope:
		c.emit(vm.OpLoadGlobal, s.Index)
	case LocalScope:
		c.emit(vm.OpLoadLocal, s.Index)
	case FreeScope:
		c.emit(vm.OpLoadFree, s.Index)
	case BuiltinScope:
		c.emit(vm.OpGetBuiltin, s.Index)
	}
}

func (c *Compiler) storeSymbol(s Symbol) {
	switch s.Scope {
	case GlobalScope:
		c.emit(vm.OpStoreGlobal, s.Index)
	case LocalScope:
		c.emit(vm.OpStoreLocal, s.Index)
	}
}
