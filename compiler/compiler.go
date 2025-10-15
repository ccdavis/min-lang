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
	Name       string
	Fields     map[string]string // field name -> field type
	FieldOrder []string          // ordered field names (Phase 3: for offset-based access)
}

// GetFieldOffset returns the offset (index) of a field, or -1 if not found
func (st *StructType) GetFieldOffset(fieldName string) int {
	for i, name := range st.FieldOrder {
		if name == fieldName {
			return i
		}
	}
	return -1
}

// Compiler represents the compiler
type Compiler struct {
	constants   []vm.Value
	symbolTable *SymbolTable

	scopes     []CompilationScope
	scopeIndex int

	typeChecker   *TypeChecker
	typeCheckMode bool

	loopStack         []LoopContext          // Stack of loop contexts
	enumTypes         map[string]*EnumType   // Tracks enum type definitions
	structTypes       map[string]*StructType // Tracks struct type definitions
	varTypes          map[string]vm.ValueType // Tracks variable types for type inference (Phase 1 optimization)
	typeInfo          map[string]Type         // Tracks detailed type information for type checking
	functionSigs      map[string]*FunctionType // Tracks function signatures for compile-time checking
	currentFunctionRT Type                    // Current function's return type (for return statement checking)
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
		constants:    []vm.Value{},
		symbolTable:  symbolTable,
		scopes:       []CompilationScope{mainScope},
		scopeIndex:   0,
		loopStack:    []LoopContext{},
		enumTypes:    make(map[string]*EnumType),
		structTypes:  make(map[string]*StructType),
		varTypes:     make(map[string]vm.ValueType),
		typeInfo:     make(map[string]Type),
		functionSigs: make(map[string]*FunctionType),
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

		// Phase 4C optimization: Detect square pattern (x * x)
		if node.Operator == "*" {
			leftIdent, leftIsIdent := node.Left.(*ast.Identifier)
			rightIdent, rightIsIdent := node.Right.(*ast.Identifier)

			if leftIsIdent && rightIsIdent && leftIdent.Value == rightIdent.Value {
				// Pattern matched: x * x
				err := c.Compile(node.Left)
				if err != nil {
					return err
				}

				// Determine type and emit appropriate square opcode
				exprType := c.inferExpressionType(node.Left)
				if exprType == vm.FloatType {
					c.emit(vm.OpSquareFloat)
				} else {
					c.emit(vm.OpSquareInt)
				}
				return nil
			}
		}

		// Phase 4A & 4D optimization: Detect operations with constant on right side
		// Check if right operand is a constant literal
		var constIndex int
		var isConstInt, isConstFloat bool

		if intLit, ok := node.Right.(*ast.IntegerLiteral); ok {
			constIndex = c.addConstant(vm.IntValue(intLit.Value))
			isConstInt = true
		} else if floatLit, ok := node.Right.(*ast.FloatLiteral); ok {
			constIndex = c.addConstant(vm.FloatValue(floatLit.Value))
			isConstFloat = true
		}

		if isConstInt || isConstFloat {
			// Compile left operand only
			err := c.Compile(node.Left)
			if err != nil {
				return err
			}

			// Emit optimized opcode based on operator
			switch node.Operator {
			// Phase 4A: Arithmetic with constant
			case "+":
				if isConstInt {
					c.emit(vm.OpAddConstInt, constIndex)
				} else {
					c.emit(vm.OpAddConstFloat, constIndex)
				}
				return nil
			case "-":
				if isConstInt {
					c.emit(vm.OpSubConstInt, constIndex)
				} else {
					c.emit(vm.OpSubConstFloat, constIndex)
				}
				return nil
			case "*":
				if isConstInt {
					c.emit(vm.OpMulConstInt, constIndex)
				} else {
					c.emit(vm.OpMulConstFloat, constIndex)
				}
				return nil
			case "/":
				if isConstInt {
					c.emit(vm.OpDivConstInt, constIndex)
				} else {
					c.emit(vm.OpDivConstFloat, constIndex)
				}
				return nil
			case "%":
				if isConstInt {
					c.emit(vm.OpModConstInt, constIndex)
					return nil
				}
			// Phase 4D: Comparison with constant
			case "==":
				if isConstInt {
					c.emit(vm.OpEqConstInt, constIndex)
				} else {
					c.emit(vm.OpEqConstFloat, constIndex)
				}
				return nil
			case "!=":
				if isConstInt {
					c.emit(vm.OpNeConstInt, constIndex)
				} else {
					c.emit(vm.OpNeConstFloat, constIndex)
				}
				return nil
			case "<":
				if isConstInt {
					c.emit(vm.OpLtConstInt, constIndex)
				} else {
					c.emit(vm.OpLtConstFloat, constIndex)
				}
				return nil
			case ">":
				if isConstInt {
					c.emit(vm.OpGtConstInt, constIndex)
				} else {
					c.emit(vm.OpGtConstFloat, constIndex)
				}
				return nil
			case "<=":
				if isConstInt {
					c.emit(vm.OpLeConstInt, constIndex)
				} else {
					c.emit(vm.OpLeConstFloat, constIndex)
				}
				return nil
			case ">=":
				if isConstInt {
					c.emit(vm.OpGeConstInt, constIndex)
				} else {
					c.emit(vm.OpGeConstFloat, constIndex)
				}
				return nil
			}
		}

		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		err = c.Compile(node.Right)
		if err != nil {
			return err
		}

		// Get operand types for type-specialized opcodes (Phase 1 optimization)
		leftType, rightType := c.getOperandTypes(node)

		switch node.Operator {
		case "+":
			c.emitTypedAdd(leftType, rightType)
		case "-":
			c.emitTypedSub(leftType, rightType)
		case "*":
			c.emitTypedMul(leftType, rightType)
		case "/":
			c.emitTypedDiv(leftType, rightType)
		case "%":
			c.emitTypedMod(leftType, rightType)
		// Phase 2: Type-specialized comparisons
		case "==":
			c.emitTypedEq(leftType, rightType)
		case "!=":
			c.emitTypedNe(leftType, rightType)
		case "<":
			c.emitTypedLt(leftType, rightType)
		case ">":
			c.emitTypedGt(leftType, rightType)
		case "<=":
			c.emitTypedLe(leftType, rightType)
		case ">=":
			c.emitTypedGe(leftType, rightType)
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

		// Track variable type for type inference (Phase 1 optimization)
		if node.Type != nil {
			c.varTypes[node.Name.Value] = typeAnnotationToValueType(node.Type)
			// Also track the full type information for type checking
			c.typeInfo[node.Name.Value] = ConvertASTType(node.Type)
		} else if node.Value != nil {
			// Infer type from value
			c.varTypes[node.Name.Value] = c.inferExpressionType(node.Value)
			c.typeInfo[node.Name.Value] = c.inferDetailedType(node.Value)
		}

		if node.Value != nil {
			// Type check the value if we have a declared type
			if node.Type != nil {
				declaredType := ConvertASTType(node.Type)

				// For arrays and maps, do deep type checking
				if err := c.checkValueType(node.Value, declaredType); err != nil {
					return err
				}
			}

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

			// Phase 4B optimization: Detect increment/decrement pattern (i = i + const)
			if infix, ok := node.Value.(*ast.InfixExpression); ok {
				if leftIdent, ok := infix.Left.(*ast.Identifier); ok {
					if leftIdent.Value == left.Value && (infix.Operator == "+" || infix.Operator == "-") {
						// Check if right side is an integer literal
						if intLit, ok := infix.Right.(*ast.IntegerLiteral); ok {
							// Pattern matched: i = i +/- constant
							amount := int(intLit.Value)
							if amount >= 0 && amount <= 65535 { // Fits in 2-byte operand
								if infix.Operator == "+" {
									if symbol.Scope == GlobalScope {
										c.emit(vm.OpIncGlobal, symbol.Index, amount)
									} else {
										c.emit(vm.OpIncLocal, symbol.Index, amount)
									}
								} else { // "-"
									if symbol.Scope == GlobalScope {
										c.emit(vm.OpDecGlobal, symbol.Index, amount)
									} else {
										c.emit(vm.OpDecLocal, symbol.Index, amount)
									}
								}
								return nil
							}
						}
						// Also handle float literals for float variables
						if floatLit, ok := infix.Right.(*ast.FloatLiteral); ok {
							// For floats, we can still use inc/dec if it's a whole number
							amount := int(floatLit.Value)
							if float64(amount) == floatLit.Value && amount >= 0 && amount <= 65535 {
								if infix.Operator == "+" {
									if symbol.Scope == GlobalScope {
										c.emit(vm.OpIncGlobal, symbol.Index, amount)
									} else {
										c.emit(vm.OpIncLocal, symbol.Index, amount)
									}
								} else { // "-"
									if symbol.Scope == GlobalScope {
										c.emit(vm.OpDecGlobal, symbol.Index, amount)
									} else {
										c.emit(vm.OpDecLocal, symbol.Index, amount)
									}
								}
								return nil
							}
						}
					}
				}
			}

			// Compile the value
			err := c.Compile(node.Value)
			if err != nil {
				return err
			}

			c.storeSymbol(symbol)

		case *ast.IndexExpression:
			// For array[index] = value or map[key] = value
			// Stack layout: array/map, index/key, value

			// Type checking for array/map assignments
			containerType := c.inferDetailedType(left.Left)
			indexType := c.inferDetailedType(left.Index)
			valueType := c.inferDetailedType(node.Value)

			if arrayType, ok := containerType.(*ArrayType); ok {
				// Array assignment: check element type
				if !IsAssignableTo(valueType, arrayType.ElementType) {
					return fmt.Errorf("cannot assign value of type %s to array element of type %s",
						valueType.String(), arrayType.ElementType.String())
				}
			} else if mapType, ok := containerType.(*MapType); ok {
				// Map assignment: check key and value types
				if !IsAssignableTo(indexType, mapType.KeyType) {
					return fmt.Errorf("cannot use key of type %s for map with key type %s",
						indexType.String(), mapType.KeyType.String())
				}
				if !IsAssignableTo(valueType, mapType.ValueType) {
					return fmt.Errorf("cannot assign value of type %s to map value of type %s",
						valueType.String(), mapType.ValueType.String())
				}
			}

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

			// Emit specialized opcode based on container type
			// The compiler knows the type, so we can avoid runtime dispatch
			if _, ok := containerType.(*MapType); ok {
				c.emit(vm.OpMapSet)
			} else {
				// Array assignment
				c.emit(vm.OpArraySet)
			}

		case *ast.FieldAccessExpression:
			// For struct.field = value
			// Stack layout: struct, [fieldName], value (or struct, value with offset)

			// Compile the struct
			err := c.Compile(left.Left)
			if err != nil {
				return err
			}

			// Phase 3 optimization: Use offset-based field access if possible
			var structTypeName string
			if structLit, ok := left.Left.(*ast.StructLiteral); ok {
				structTypeName = structLit.Name.Value
			}

			// Try offset-based access
			useOffset := false
			var offset int
			if structTypeName != "" {
				if structType, ok := c.structTypes[structTypeName]; ok {
					offset = structType.GetFieldOffset(left.Field.Value)
					if offset >= 0 {
						useOffset = true
					}
				}
			}

			if !useOffset {
				// Push field name for name-based access
				c.emit(vm.OpPush, c.addConstant(vm.StringValue(left.Field.Value)))
			}

			// Compile the value
			err = c.Compile(node.Value)
			if err != nil {
				return err
			}

			// Emit set field operation
			if useOffset {
				c.emit(vm.OpSetFieldOffset, offset)
			} else {
				c.emit(vm.OpSetField)
			}

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
				Name:       node.Name.Value,
				Fields:     make(map[string]string),
				FieldOrder: make([]string, 0, len(def.Fields)),
			}

			// Store field types and order (Phase 3: for offset-based access)
			for _, field := range def.Fields {
				structType.Fields[field.Name.Value] = field.Type.String()
				structType.FieldOrder = append(structType.FieldOrder, field.Name.Value)
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
		// Build function signature for type checking
		paramTypes := make([]Type, len(node.Parameters))
		for i, param := range node.Parameters {
			paramTypes[i] = ConvertASTType(param.Type)
		}
		returnType := ConvertASTType(node.ReturnType)

		funcType := &FunctionType{
			ParamTypes: paramTypes,
			ReturnType: returnType,
		}
		c.functionSigs[node.Name.Value] = funcType
		c.typeInfo[node.Name.Value] = funcType

		// Define the function name in the current scope BEFORE compiling the body
		// This allows recursive calls
		symbol := c.symbolTable.Define(node.Name.Value)

		c.enterScope()

		// Store the previous return type and set current one
		prevReturnType := c.currentFunctionRT
		c.currentFunctionRT = returnType

		// Define parameters in the new scope
		for i, param := range node.Parameters {
			c.symbolTable.Define(param.Name.Value)
			// Track parameter types
			c.typeInfo[param.Name.Value] = paramTypes[i]
		}

		err := c.Compile(node.Body)
		if err != nil {
			return err
		}

		// If the last instruction is not a return, add an implicit return nil
		if !c.lastInstructionIs(vm.OpReturn) {
			// Check if function expects a specific non-nil return value
			if returnType != nil && !returnType.Equals(NilType) && !returnType.Equals(AnyTypeVal) {
				return fmt.Errorf("function %s must return %s", node.Name.Value, returnType.String())
			}
			c.emit(vm.OpPush, c.addConstant(vm.NilValue()))
			c.emit(vm.OpReturn)
		}

		// Restore previous return type
		c.currentFunctionRT = prevReturnType

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
			// Type check return value
			if c.currentFunctionRT != nil {
				returnValueType := c.inferDetailedType(node.ReturnValue)
				if !IsAssignableTo(returnValueType, c.currentFunctionRT) {
					return fmt.Errorf("cannot return %s from function expecting %s",
						returnValueType.String(), c.currentFunctionRT.String())
				}
			}

			err := c.Compile(node.ReturnValue)
			if err != nil {
				return err
			}
		} else {
			// Returning nil
			if c.currentFunctionRT != nil && !c.currentFunctionRT.Equals(NilType) && !c.currentFunctionRT.Equals(AnyTypeVal) {
				return fmt.Errorf("cannot return nil from function expecting %s", c.currentFunctionRT.String())
			}
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
		// Type check function call if we know the function signature
		if ident, ok := node.Function.(*ast.Identifier); ok {
			if funcType, exists := c.functionSigs[ident.Value]; exists {
				// Check argument count
				if len(node.Arguments) != len(funcType.ParamTypes) {
					return fmt.Errorf("function %s expects %d arguments, got %d",
						ident.Value, len(funcType.ParamTypes), len(node.Arguments))
				}

				// Check argument types
				for i, arg := range node.Arguments {
					argType := c.inferDetailedType(arg)
					expectedType := funcType.ParamTypes[i]
					if !IsAssignableTo(argType, expectedType) {
						return fmt.Errorf("function %s argument %d: expected %s, got %s",
							ident.Value, i+1, expectedType.String(), argType.String())
					}
				}
			}
		}

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
		// Phase 3 optimization: Use offset-based struct creation if type is known
		if structType, ok := c.structTypes[node.Name.Value]; ok {
			// We know the struct type - use ordered creation
			// Compile fields in the correct order, with field names
			for _, fieldName := range structType.FieldOrder {
				value, exists := node.Fields[fieldName]
				if !exists {
					return fmt.Errorf("missing required field %s in struct %s", fieldName, node.Name.Value)
				}
				// Push field name first
				c.emit(vm.OpPush, c.addConstant(vm.StringValue(fieldName)))
				// Then field value
				err := c.Compile(value)
				if err != nil {
					return err
				}
			}

			// Push the type name as a string (last, will be popped first)
			c.emit(vm.OpPush, c.addConstant(vm.StringValue(node.Name.Value)))

			// Emit OpStructOrdered with number of fields
			c.emit(vm.OpStructOrdered, len(node.Fields))
		} else {
			// Fallback to name-based struct creation (for unknown types)
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
		}

	case *ast.IndexExpression:
		// Type checking for map key access
		containerType := c.inferDetailedType(node.Left)
		if mapType, ok := containerType.(*MapType); ok {
			// For map access, check that the index type matches the key type
			indexType := c.inferDetailedType(node.Index)
			if !IsAssignableTo(indexType, mapType.KeyType) {
				return fmt.Errorf("cannot use key of type %s for map with key type %s",
					indexType.String(), mapType.KeyType.String())
			}
		}

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

		// Emit specialized opcode based on container type
		// The compiler knows the type, so we can avoid runtime dispatch
		if _, ok := containerType.(*MapType); ok {
			c.emit(vm.OpMapGet)
		} else {
			// Array or string indexing
			c.emit(vm.OpArrayGet)
		}

	case *ast.FieldAccessExpression:
		// Compile the struct expression
		err := c.Compile(node.Left)
		if err != nil {
			return err
		}

		// Phase 3 optimization: Use offset-based field access if possible
		// Try to determine the struct type from the left expression
		var structTypeName string
		if ident, ok := node.Left.(*ast.Identifier); ok {
			// Check if this is a known variable with struct type
			if varType, exists := c.varTypes[ident.Value]; exists && varType == vm.StructType {
				// We'd need more detailed type tracking to know which struct type
				// For now, fall through to name-based access
			}
		}
		// Check if left is a struct literal - we know the type directly
		if structLit, ok := node.Left.(*ast.StructLiteral); ok {
			structTypeName = structLit.Name.Value
		}

		// If we know the struct type, use offset-based access
		if structTypeName != "" {
			if structType, ok := c.structTypes[structTypeName]; ok {
				offset := structType.GetFieldOffset(node.Field.Value)
				if offset >= 0 {
					// Use offset-based access - much faster!
					c.emit(vm.OpGetFieldOffset, offset)
					return nil
				}
			}
		}

		// Fallback to name-based access
		c.emit(vm.OpPush, c.addConstant(vm.StringValue(node.Field.Value)))
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

		// Check exhaustiveness for enum switches
		if node.Default == nil {
			err := c.checkSwitchExhaustiveness(node)
			if err != nil {
				return err
			}
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

// checkSwitchExhaustiveness checks if a switch statement on an enum is exhaustive
func (c *Compiler) checkSwitchExhaustiveness(node *ast.SwitchStatement) error {
	// Try to determine the enum type of the switch value
	var enumType *EnumType

	// Check if the switch value is an identifier that's an enum variant
	if ident, ok := node.Value.(*ast.Identifier); ok {
		// Check if this identifier is a known enum variant
		for _, et := range c.enumTypes {
			for _, variantName := range et.VariantNames {
				if ident.Value == variantName {
					// This is an enum variant, but we're switching on the variant itself
					// which is just an int, not useful for exhaustiveness checking
					return nil
				}
			}
		}

		// Check if this is a variable that holds an enum value
		// We need to track which enum type a variable belongs to
		// For now, we'll use a heuristic: check if all case values are from the same enum
	}

	// Try to infer enum type from case values
	// Collect all case values and check if they're all from the same enum
	caseVariants := make(map[string]bool)
	var detectedEnumType *EnumType

	for _, caseClause := range node.Cases {
		if caseIdent, ok := caseClause.Value.(*ast.Identifier); ok {
			// Check which enum this variant belongs to
			for _, et := range c.enumTypes {
				if _, exists := et.Variants[caseIdent.Value]; exists {
					if detectedEnumType == nil {
						detectedEnumType = et
					} else if detectedEnumType.Name != et.Name {
						// Mixed enums in switch - can't check exhaustiveness
						return nil
					}
					caseVariants[caseIdent.Value] = true
					break
				}
			}
		}
	}

	// If we detected an enum type, check exhaustiveness
	if detectedEnumType != nil {
		enumType = detectedEnumType

		// Check if all variants are covered
		missingVariants := []string{}
		for _, variantName := range enumType.VariantNames {
			if !caseVariants[variantName] {
				missingVariants = append(missingVariants, variantName)
			}
		}

		if len(missingVariants) > 0 {
			// Build a helpful error message
			missing := ""
			for i, v := range missingVariants {
				if i > 0 {
					if i == len(missingVariants)-1 {
						missing += " and "
					} else {
						missing += ", "
					}
				}
				missing += v
			}
			return fmt.Errorf("switch on enum %s is not exhaustive, missing cases: %s", enumType.Name, missing)
		}
	} else {
		// Not an enum switch - require a default clause for safety
		return fmt.Errorf("switch statement must have a default case (or switch on an enum with all variants covered)")
	}

	return nil
}
