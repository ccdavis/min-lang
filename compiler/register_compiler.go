package compiler

import (
	"fmt"
	"minlang/ast"
	"minlang/vm"
)

// RegisterCompiler extends Compiler to emit register bytecode
type RegisterCompiler struct {
	*Compiler // Embed stack compiler for reuse

	// Register allocation
	registers      map[string]int        // Variable name -> register
	nextReg        int                   // Next available register
	MaxRegs        int                   // Max registers used in current function (exported)
	tempRegs       []int                 // Available temporary registers
	liveRanges     map[string]*LiveRange // Variable live ranges
	instructions   []vm.RegisterInstruction

	// Register scope stack
	regScopes      []map[string]int
	regScopeIndex  int

	// Loop context stack
	loopStack      []LoopContext
}

// LiveRange tracks when a variable is live
type LiveRange struct {
	start int // First instruction where variable is defined
	end   int // Last instruction where variable is used
}

// constOpMatch describes a const-aware register opcode that fits a particular
// `expr op literal` pattern. constVal is the Value to add to the constant pool
// and feed via the instruction's C field.
type constOpMatch struct {
	op       vm.RegisterOpCode
	constVal vm.Value
}

// constOpForLiteralRight inspects an InfixExpression and returns the const-aware
// opcode and constant Value if the right operand is an integer or float literal
// AND the operator has a const-aware variant. Otherwise returns (_, false).
func constOpForLiteralRight(node *ast.InfixExpression, rc *RegisterCompiler) (constOpMatch, bool) {
	var rightIsInt, rightIsFloat bool
	var intVal int64
	var floatVal float64
	if il, ok := node.Right.(*ast.IntegerLiteral); ok {
		rightIsInt = true
		intVal = il.Value
	} else if fl, ok := node.Right.(*ast.FloatLiteral); ok {
		rightIsFloat = true
		floatVal = fl.Value
	} else {
		return constOpMatch{}, false
	}

	leftType := rc.inferExpressionType(node.Left)
	useFloat := rightIsFloat || leftType == vm.FloatType

	var op vm.RegisterOpCode
	switch node.Operator {
	case "+":
		if useFloat {
			op = vm.OpRAddConstFloat
		} else {
			op = vm.OpRAddConstInt
		}
	case "*":
		if useFloat {
			op = vm.OpRMulConstFloat
		} else {
			op = vm.OpRMulConstInt
		}
	case "<":
		if useFloat {
			op = vm.OpRLtConstFloat
		} else {
			op = vm.OpRLtConstInt
		}
	case ">":
		if useFloat {
			op = vm.OpRGtConstFloat
		} else {
			op = vm.OpRGtConstInt
		}
	case "<=":
		if useFloat {
			op = vm.OpRLeConstFloat
		} else {
			op = vm.OpRLeConstInt
		}
	case ">=":
		if useFloat {
			op = vm.OpRGeConstFloat
		} else {
			op = vm.OpRGeConstInt
		}
	default:
		return constOpMatch{}, false
	}

	var v vm.Value
	if useFloat {
		if rightIsInt {
			v = vm.FloatValue(float64(intVal))
		} else {
			v = vm.FloatValue(floatVal)
		}
	} else {
		v = vm.IntValue(intVal)
	}
	return constOpMatch{op: op, constVal: v}, true
}

// NewRegisterCompiler creates a new register compiler
func NewRegisterCompiler() *RegisterCompiler {
	return &RegisterCompiler{
		Compiler:      New(),
		registers:     make(map[string]int),
		nextReg:       0,
		MaxRegs:       0,
		tempRegs:      []int{},
		liveRanges:    make(map[string]*LiveRange),
		instructions:  []vm.RegisterInstruction{},
		regScopes:     []map[string]int{},
		regScopeIndex: 0,
		loopStack:     []LoopContext{},
	}
}

// enterLoop pushes a new loop context
func (rc *RegisterCompiler) enterRegisterLoop() {
	rc.loopStack = append(rc.loopStack, LoopContext{
		breakJumps:    []int{},
		continueJumps: []int{},
	})
}

// leaveLoop pops a loop context
func (rc *RegisterCompiler) leaveRegisterLoop() {
	rc.loopStack = rc.loopStack[:len(rc.loopStack)-1]
}

// currentLoop returns the current loop context
func (rc *RegisterCompiler) currentRegisterLoop() *LoopContext {
	if len(rc.loopStack) == 0 {
		return nil
	}
	return &rc.loopStack[len(rc.loopStack)-1]
}

// allocateRegister allocates a register for a variable
func (rc *RegisterCompiler) allocateRegister(name string) int {
	// Check if already allocated
	if reg, exists := rc.registers[name]; exists {
		return reg
	}

	// Allocate new register
	reg := rc.nextReg
	rc.registers[name] = reg
	rc.nextReg++

	if rc.nextReg > rc.MaxRegs {
		rc.MaxRegs = rc.nextReg
	}

	return reg
}

// allocateConsecutiveBlock reserves n contiguous register slots starting at
// nextReg. Always at the top: callees in the register VM occupy a contiguous
// window starting at the caller's argBaseReg and extending for NumLocals slots,
// so any in-use register below argBaseReg risks getting clobbered. We rely on
// freeTempRegister's stack-discipline reclaim (see below) to keep MaxRegs
// bounded across many calls.
func (rc *RegisterCompiler) allocateConsecutiveBlock(n int) int {
	if n <= 0 {
		return rc.nextReg
	}
	base := rc.nextReg
	rc.nextReg += n
	if rc.nextReg > rc.MaxRegs {
		rc.MaxRegs = rc.nextReg
	}
	// Remove any in-block slots that lingered in tempRegs (rare; defensive).
	if len(rc.tempRegs) > 0 {
		pruned := rc.tempRegs[:0]
		for _, r := range rc.tempRegs {
			if r < base || r >= base+n {
				pruned = append(pruned, r)
			}
		}
		rc.tempRegs = pruned
	}
	return base
}

// allocateTempRegister allocates a temporary register
func (rc *RegisterCompiler) allocateTempRegister() int {
	// Reuse freed temps if available
	if len(rc.tempRegs) > 0 {
		reg := rc.tempRegs[len(rc.tempRegs)-1]
		rc.tempRegs = rc.tempRegs[:len(rc.tempRegs)-1]
		return reg
	}

	// Allocate new temp
	reg := rc.nextReg
	rc.nextReg++

	if rc.nextReg > rc.MaxRegs {
		rc.MaxRegs = rc.nextReg
	}

	return reg
}

// freeTempRegister returns a temporary register to the pool.
// Permanent variable registers (parameters, locals declared with var) are
// rejected — adding them lets the next temp allocation clobber the variable.
// After enqueueing, we compact the top of the stack: any contiguous freed
// slots at nextReg-1 collapse back into the unallocated zone. This keeps the
// allocator at stack discipline so repeated call sites don't grow nextReg
// past 256 (the 8-bit register-field limit).
func (rc *RegisterCompiler) freeTempRegister(reg int) {
	for _, permReg := range rc.registers {
		if permReg == reg {
			return
		}
	}
	for _, r := range rc.tempRegs {
		if r == reg {
			return
		}
	}
	rc.tempRegs = append(rc.tempRegs, reg)
	// Compact: if the just-freed range now reaches nextReg-1, shrink nextReg.
	for rc.nextReg > 0 {
		top := rc.nextReg - 1
		found := -1
		for i, r := range rc.tempRegs {
			if r == top {
				found = i
				break
			}
		}
		if found < 0 {
			break
		}
		rc.tempRegs = append(rc.tempRegs[:found], rc.tempRegs[found+1:]...)
		rc.nextReg--
	}
}

// emitR emits a register instruction
func (rc *RegisterCompiler) emitR(op vm.RegisterOpCode, a, b, c uint8) int {
	ins := vm.EncodeRegisterInstruction(op, a, b, c)
	rc.instructions = append(rc.instructions, ins)
	return len(rc.instructions) - 1
}

// emitRBx emits a register instruction with large immediate
func (rc *RegisterCompiler) emitRBx(op vm.RegisterOpCode, a uint8, bx uint16) int {
	ins := vm.EncodeRegisterInstructionBx(op, a, bx)
	rc.instructions = append(rc.instructions, ins)
	return len(rc.instructions) - 1
}

// RegisterBytecode returns the compiled register bytecode
func (rc *RegisterCompiler) RegisterBytecode() *vm.RegisterBytecode {
	return &vm.RegisterBytecode{
		Instructions: rc.instructions,
		Constants:    rc.constants,
		MainFunction: &vm.Function{
			Name:         "main",
			NumParams:    0,
			NumLocals:    rc.MaxRegs,
			Instructions: nil, // Register bytecode is stored separately
		},
	}
}

// CompileToRegister compiles an AST node to register bytecode
// Returns the register number containing the result (or -1 for statements)
func (rc *RegisterCompiler) CompileToRegister(node ast.Node) (int, error) {
	switch node := node.(type) {
	case *ast.Program:
		for _, s := range node.Statements {
			_, err := rc.CompileToRegister(s)
			if err != nil {
				return -1, err
			}
		}
		return -1, nil

	case *ast.ExpressionStatement:
		// Compile expression and discard result
		resultReg, err := rc.CompileToRegister(node.Expression)
		if err != nil {
			return -1, err
		}
		// Free the result register if it's a temp
		if resultReg >= 0 {
			rc.freeTempRegister(resultReg)
		}
		return -1, nil

	case *ast.TypeStatement:
		// Type declarations are compile-time metadata only.
		// Structs: register name+field-order; emit no bytecode.
		// Enums: register variants and emit init code to bind each variant
		// to a global/local (mirroring the stack compiler).
		switch def := node.Definition.(type) {
		case *ast.StructStatement:
			def.Name = node.Name
			structType := &StructType{
				Name:       node.Name.Value,
				Fields:     make(map[string]string),
				FieldOrder: make([]string, 0, len(def.Fields)),
			}
			for _, field := range def.Fields {
				structType.Fields[field.Name.Value] = field.Type.String()
				structType.FieldOrder = append(structType.FieldOrder, field.Name.Value)
			}
			rc.structTypes[node.Name.Value] = structType
			return -1, nil

		case *ast.EnumStatement:
			def.Name = node.Name
			return rc.CompileToRegister(def)
		}
		return -1, nil

	case *ast.EnumStatement:
		// Register enum metadata; bind each variant to its integer value in
		// the enclosing scope. Mirrors the stack compiler's behaviour but
		// emits register opcodes.
		enumType := &EnumType{
			Name:         node.Name.Value,
			Variants:     make(map[string]int),
			VariantNames: make([]string, len(node.Variants)),
		}
		for i, variant := range node.Variants {
			enumType.Variants[variant.Value] = i
			enumType.VariantNames[i] = variant.Value

			symbol := rc.symbolTable.DefineWithMutability(variant.Value, false)
			constIdx := rc.addConstant(vm.IntValue(int64(i)))
			if symbol.Scope == GlobalScope {
				tempReg := rc.allocateTempRegister()
				rc.emitRBx(vm.OpRLoadK, uint8(tempReg), uint16(constIdx))
				rc.emitRBx(vm.OpRStoreGlobal, uint8(tempReg), uint16(symbol.Index))
				rc.freeTempRegister(tempReg)
			} else {
				varReg := rc.allocateRegister(variant.Value)
				rc.emitRBx(vm.OpRLoadK, uint8(varReg), uint16(constIdx))
			}
		}
		rc.enumTypes[node.Name.Value] = enumType
		vm.EnumRegistry[node.Name.Value] = make(map[int]string)
		for value, name := range enumType.VariantNames {
			vm.EnumRegistry[node.Name.Value][value] = name
		}
		return -1, nil

	case *ast.IntegerLiteral:
		// Load constant into temp register
		constIndex := rc.addConstant(vm.IntValue(node.Value))
		tempReg := rc.allocateTempRegister()
		rc.emitRBx(vm.OpRLoadK, uint8(tempReg), uint16(constIndex))
		return tempReg, nil

	case *ast.FloatLiteral:
		constIndex := rc.addConstant(vm.FloatValue(node.Value))
		tempReg := rc.allocateTempRegister()
		rc.emitRBx(vm.OpRLoadK, uint8(tempReg), uint16(constIndex))
		return tempReg, nil

	case *ast.BooleanLiteral:
		constIndex := rc.addConstant(vm.BoolValue(node.Value))
		tempReg := rc.allocateTempRegister()
		rc.emitRBx(vm.OpRLoadK, uint8(tempReg), uint16(constIndex))
		return tempReg, nil

	case *ast.StringLiteral:
		constIndex := rc.addConstant(vm.StringValue(node.Value))
		tempReg := rc.allocateTempRegister()
		rc.emitRBx(vm.OpRLoadK, uint8(tempReg), uint16(constIndex))
		return tempReg, nil

	case *ast.NilLiteral:
		constIndex := rc.addConstant(vm.NilValue())
		tempReg := rc.allocateTempRegister()
		rc.emitRBx(vm.OpRLoadK, uint8(tempReg), uint16(constIndex))
		return tempReg, nil

	case *ast.Identifier:
		// Check symbol table first (for builtins and scope tracking)
		symbol, ok := rc.symbolTable.Resolve(node.Value)
		if !ok {
			return -1, fmt.Errorf("undefined variable: %s", node.Value)
		}

		// For builtins, we'll handle them in CallExpression
		// Just return a marker value for now
		if symbol.Scope == BuiltinScope {
			// We can't really "load" a builtin - it's handled at call time
			// Return negative value encoding the builtin index
			// This is a bit of a hack, but it works
			return -(symbol.Index + 100), nil
		}

		// Check if it's a global variable
		if symbol.Scope == GlobalScope {
			// Load from globals array into temp register
			tempReg := rc.allocateTempRegister()
			rc.emitRBx(vm.OpRLoadGlobal, uint8(tempReg), uint16(symbol.Index))
			return tempReg, nil
		}

		// Local variable reference - should be in a register
		if reg, exists := rc.registers[node.Value]; exists {
			return reg, nil
		}
		return -1, fmt.Errorf("variable %s not in register (symbol scope: %v)", node.Value, symbol.Scope)

	case *ast.VarStatement:
		// Define in symbol table
		symbol := rc.symbolTable.DefineWithMutability(node.Name.Value, node.IsMutable)

		// Track variable type
		if node.Type != nil {
			rc.varTypes[node.Name.Value] = typeAnnotationToValueType(node.Type)
			rc.typeInfo[node.Name.Value] = ConvertASTType(node.Type)
		} else if node.Value != nil {
			rc.varTypes[node.Name.Value] = rc.inferExpressionType(node.Value)
			rc.typeInfo[node.Name.Value] = rc.inferDetailedType(node.Value)
		}

		// Check if this is a global or local variable
		if symbol.Scope == GlobalScope {
			// Global variable - use OpRStoreGlobal
			if node.Value != nil {
				valueReg, err := rc.CompileToRegister(node.Value)
				if err != nil {
					return -1, err
				}
				rc.emitRBx(vm.OpRStoreGlobal, uint8(valueReg), uint16(symbol.Index))
				rc.freeTempRegister(valueReg)
			}
		} else {
			// Local variable - allocate register
			reg := rc.allocateRegister(node.Name.Value)

			// Compile initializer value if present
			if node.Value != nil {
				valueReg, err := rc.CompileToRegister(node.Value)
				if err != nil {
					return -1, err
				}
				// Move value to variable register
				if valueReg != reg {
					rc.emitR(vm.OpRMove, uint8(reg), uint8(valueReg), 0)
					rc.freeTempRegister(valueReg)
				}
			}
		}

		return -1, nil

	case *ast.AssignmentStatement:
		switch left := node.Left.(type) {
		case *ast.Identifier:
			// Variable assignment
			valueReg, err := rc.CompileToRegister(node.Value)
			if err != nil {
				return -1, err
			}

			// Check if this is a global variable
			symbol, ok := rc.symbolTable.Resolve(left.Value)
			if !ok {
				return -1, fmt.Errorf("undefined variable: %s", left.Value)
			}

			if symbol.Scope == GlobalScope {
				// Global variable assignment
				rc.emitRBx(vm.OpRStoreGlobal, uint8(valueReg), uint16(symbol.Index))
				rc.freeTempRegister(valueReg)
			} else {
				// Local variable assignment
				varReg, exists := rc.registers[left.Value]
				if !exists {
					return -1, fmt.Errorf("undefined local variable: %s", left.Value)
				}

				// Move value to variable register
				if valueReg != varReg {
					rc.emitR(vm.OpRMove, uint8(varReg), uint8(valueReg), 0)
					rc.freeTempRegister(valueReg)
				}
			}

		case *ast.IndexExpression:
			// Array/map assignment: arr[i] = value
			containerReg, err := rc.CompileToRegister(left.Left)
			if err != nil {
				return -1, err
			}

			indexReg, err := rc.CompileToRegister(left.Index)
			if err != nil {
				return -1, err
			}

			valueReg, err := rc.CompileToRegister(node.Value)
			if err != nil {
				return -1, err
			}

			// OpRSetIdx unconditionally casts to *ArrayValue; maps need OpRMapSet.
			setOp := vm.OpRSetIdx
			if rc.inferExpressionType(left.Left) == vm.MapType {
				setOp = vm.OpRMapSet
			}
			rc.emitR(setOp, uint8(containerReg), uint8(indexReg), uint8(valueReg))

			rc.freeTempRegister(containerReg)
			rc.freeTempRegister(indexReg)
			rc.freeTempRegister(valueReg)

		case *ast.FieldAccessExpression:
			// Struct field assignment: obj.field = value
			objReg, err := rc.CompileToRegister(left.Left)
			if err != nil {
				return -1, err
			}

			valueReg, err := rc.CompileToRegister(node.Value)
			if err != nil {
				return -1, err
			}

			fieldIdx := rc.addConstant(vm.StringValue(left.Field.Value))
			if fieldIdx >= 256 {
				return -1, fmt.Errorf("too many constants: field name %q exceeds register-VM uint8 limit", left.Field.Value)
			}
			// OpRSetField: R(A).field(K(B)) = R(C)
			rc.emitR(vm.OpRSetField, uint8(objReg), uint8(fieldIdx), uint8(valueReg))

			rc.freeTempRegister(objReg)
			rc.freeTempRegister(valueReg)
		}
		return -1, nil

	case *ast.InfixExpression:
		// Peephole: `expr OP literal` -> emit const-aware opcode that reads
		// the constant inline instead of materializing it into a register first.
		// The register instruction's C field is uint8 so this only fires when
		// the constant index fits in 256.
		if constOp, ok := constOpForLiteralRight(node, rc); ok {
			constIdx := rc.addConstant(constOp.constVal)
			// Register instruction's C field is uint8, so the const op variant
			// only applies when the (deduped) constant index fits in 256.
			if constIdx < 256 {
				leftReg, err := rc.CompileToRegister(node.Left)
				if err != nil {
					return -1, err
				}
				resultReg := rc.allocateTempRegister()
				rc.emitR(constOp.op, uint8(resultReg), uint8(leftReg), uint8(constIdx))
				rc.freeTempRegister(leftReg)
				return resultReg, nil
			}
		}

		// Compile left and right operands
		leftReg, err := rc.CompileToRegister(node.Left)
		if err != nil {
			return -1, err
		}
		rightReg, err := rc.CompileToRegister(node.Right)
		if err != nil {
			return -1, err
		}

		// Determine types
		leftType := rc.inferExpressionType(node.Left)
		rightType := rc.inferExpressionType(node.Right)

		// Allocate result register
		resultReg := rc.allocateTempRegister()

		switch node.Operator {
		case "+":
			if leftType == vm.StringType || rightType == vm.StringType {
				rc.emitR(vm.OpRConcat, uint8(resultReg), uint8(leftReg), uint8(rightReg))
			} else if leftType == vm.IntType && rightType == vm.IntType {
				rc.emitR(vm.OpRAddInt, uint8(resultReg), uint8(leftReg), uint8(rightReg))
			} else {
				rc.emitR(vm.OpRAddFloat, uint8(resultReg), uint8(leftReg), uint8(rightReg))
			}
		case "-":
			if leftType == vm.IntType && rightType == vm.IntType {
				rc.emitR(vm.OpRSubInt, uint8(resultReg), uint8(leftReg), uint8(rightReg))
			} else {
				rc.emitR(vm.OpRSubFloat, uint8(resultReg), uint8(leftReg), uint8(rightReg))
			}
		case "*":
			// Check for square pattern (x * x)
			leftIdent, leftIsIdent := node.Left.(*ast.Identifier)
			rightIdent, rightIsIdent := node.Right.(*ast.Identifier)
			if leftIsIdent && rightIsIdent && leftIdent.Value == rightIdent.Value {
				// Square optimization
				if leftType == vm.FloatType {
					rc.emitR(vm.OpRSquareFloat, uint8(resultReg), uint8(leftReg), 0)
				} else {
					rc.emitR(vm.OpRSquareInt, uint8(resultReg), uint8(leftReg), 0)
				}
			} else {
				if leftType == vm.IntType && rightType == vm.IntType {
					rc.emitR(vm.OpRMulInt, uint8(resultReg), uint8(leftReg), uint8(rightReg))
				} else {
					rc.emitR(vm.OpRMulFloat, uint8(resultReg), uint8(leftReg), uint8(rightReg))
				}
			}
		case "/":
			if leftType == vm.IntType && rightType == vm.IntType {
				rc.emitR(vm.OpRDivInt, uint8(resultReg), uint8(leftReg), uint8(rightReg))
			} else {
				rc.emitR(vm.OpRDivFloat, uint8(resultReg), uint8(leftReg), uint8(rightReg))
			}
		case "%":
			rc.emitR(vm.OpRModInt, uint8(resultReg), uint8(leftReg), uint8(rightReg))

		// Comparisons
		case "==":
			if leftType == vm.IntType {
				rc.emitR(vm.OpREqInt, uint8(resultReg), uint8(leftReg), uint8(rightReg))
			} else if leftType == vm.FloatType {
				rc.emitR(vm.OpREqFloat, uint8(resultReg), uint8(leftReg), uint8(rightReg))
			} else if leftType == vm.StringType {
				rc.emitR(vm.OpREqString, uint8(resultReg), uint8(leftReg), uint8(rightReg))
			} else {
				rc.emitR(vm.OpREqBool, uint8(resultReg), uint8(leftReg), uint8(rightReg))
			}
		case "!=":
			if leftType == vm.IntType {
				rc.emitR(vm.OpRNeInt, uint8(resultReg), uint8(leftReg), uint8(rightReg))
			} else if leftType == vm.FloatType {
				rc.emitR(vm.OpRNeFloat, uint8(resultReg), uint8(leftReg), uint8(rightReg))
			} else if leftType == vm.StringType {
				rc.emitR(vm.OpRNeString, uint8(resultReg), uint8(leftReg), uint8(rightReg))
			} else {
				rc.emitR(vm.OpRNeBool, uint8(resultReg), uint8(leftReg), uint8(rightReg))
			}
		case "<":
			if leftType == vm.IntType {
				rc.emitR(vm.OpRLtInt, uint8(resultReg), uint8(leftReg), uint8(rightReg))
			} else {
				rc.emitR(vm.OpRLtFloat, uint8(resultReg), uint8(leftReg), uint8(rightReg))
			}
		case ">":
			if leftType == vm.IntType {
				rc.emitR(vm.OpRGtInt, uint8(resultReg), uint8(leftReg), uint8(rightReg))
			} else {
				rc.emitR(vm.OpRGtFloat, uint8(resultReg), uint8(leftReg), uint8(rightReg))
			}
		case "<=":
			if leftType == vm.IntType {
				rc.emitR(vm.OpRLeInt, uint8(resultReg), uint8(leftReg), uint8(rightReg))
			} else {
				rc.emitR(vm.OpRLeFloat, uint8(resultReg), uint8(leftReg), uint8(rightReg))
			}
		case ">=":
			if leftType == vm.IntType {
				rc.emitR(vm.OpRGeInt, uint8(resultReg), uint8(leftReg), uint8(rightReg))
			} else {
				rc.emitR(vm.OpRGeFloat, uint8(resultReg), uint8(leftReg), uint8(rightReg))
			}

		// Logical
		case "&&":
			rc.emitR(vm.OpRAnd, uint8(resultReg), uint8(leftReg), uint8(rightReg))
		case "||":
			rc.emitR(vm.OpROr, uint8(resultReg), uint8(leftReg), uint8(rightReg))

		default:
			return -1, fmt.Errorf("unknown operator: %s", node.Operator)
		}

		// freeTempRegister itself rejects permanent variable registers, so the
		// callers no longer need the linear scan they used to do here.
		rc.freeTempRegister(leftReg)
		rc.freeTempRegister(rightReg)

		return resultReg, nil

	case *ast.PrefixExpression:
		operandReg, err := rc.CompileToRegister(node.Right)
		if err != nil {
			return -1, err
		}

		resultReg := rc.allocateTempRegister()

		switch node.Operator {
		case "!":
			rc.emitR(vm.OpRNot, uint8(resultReg), uint8(operandReg), 0)
		case "-":
			exprType := rc.inferExpressionType(node.Right)
			if exprType == vm.IntType {
				rc.emitR(vm.OpRNegInt, uint8(resultReg), uint8(operandReg), 0)
			} else {
				rc.emitR(vm.OpRNegFloat, uint8(resultReg), uint8(operandReg), 0)
			}
		}

		rc.freeTempRegister(operandReg)
		return resultReg, nil

	case *ast.IfStatement:
		// Compile condition
		condReg, err := rc.CompileToRegister(node.Condition)
		if err != nil {
			return -1, err
		}

		// Pick fast-path jump if the condition's static type is bool.
		jumpOp := vm.OpRJumpF
		if rc.inferExpressionType(node.Condition) == vm.BoolType {
			jumpOp = vm.OpRJumpFBool
		}

		// Jump if false (placeholder)
		jumpIfFalse := rc.emitRBx(jumpOp, uint8(condReg), 9999)
		rc.freeTempRegister(condReg)

		// Compile consequence
		_, err = rc.CompileToRegister(node.Consequence)
		if err != nil {
			return -1, err
		}

		// Jump over alternative
		jumpOverAlt := rc.emitRBx(vm.OpRJump, 0, 9999)

		// Patch first jump
		afterConsequence := len(rc.instructions)
		rc.instructions[jumpIfFalse] = vm.EncodeRegisterInstructionBx(
			jumpOp, uint8(condReg), uint16(afterConsequence))

		// Compile alternative if present
		if node.Alternative != nil {
			_, err = rc.CompileToRegister(node.Alternative)
			if err != nil {
				return -1, err
			}
		}

		// Patch second jump
		afterAlternative := len(rc.instructions)
		rc.instructions[jumpOverAlt] = vm.EncodeRegisterInstructionBx(
			vm.OpRJump, 0, uint16(afterAlternative))

		return -1, nil

	case *ast.SwitchStatement:
		// Switch is compiled as a chain of test-and-skip: for each case, eq-compare
		// switch value with case value; if not equal, jump past the body; if equal,
		// run the body then jump to the end.
		switchReg, err := rc.CompileToRegister(node.Value)
		if err != nil {
			return -1, err
		}

		// Pick the type-specialized eq opcode for the switch value's type.
		var eqOp vm.RegisterOpCode
		switch rc.inferExpressionType(node.Value) {
		case vm.FloatType:
			eqOp = vm.OpREqFloat
		case vm.StringType:
			eqOp = vm.OpREqString
		case vm.BoolType:
			eqOp = vm.OpREqBool
		default:
			// IntType plus anything else (enums resolve to int).
			eqOp = vm.OpREqInt
		}

		caseEndJumps := []int{}

		for _, caseClause := range node.Cases {
			caseValReg, err := rc.CompileToRegister(caseClause.Value)
			if err != nil {
				return -1, err
			}
			eqReg := rc.allocateTempRegister()
			rc.emitR(eqOp, uint8(eqReg), uint8(switchReg), uint8(caseValReg))
			rc.freeTempRegister(caseValReg)

			// Skip the body if the comparison was false.
			skipJump := rc.emitRBx(vm.OpRJumpFBool, uint8(eqReg), 9999)
			rc.freeTempRegister(eqReg)

			if _, err := rc.CompileToRegister(caseClause.Body); err != nil {
				return -1, err
			}

			// After the body, jump over the remaining cases / default.
			caseEndJumps = append(caseEndJumps, rc.emitRBx(vm.OpRJump, 0, 9999))

			// Patch the skip-target to the instruction right after the case-end jump.
			afterCase := len(rc.instructions)
			rc.instructions[skipJump] = vm.EncodeRegisterInstructionBx(
				vm.OpRJumpFBool, uint8(eqReg), uint16(afterCase))
		}

		if node.Default == nil {
			// Reuse the exhaustiveness check from the stack compiler (embedded).
			if err := rc.checkSwitchExhaustiveness(node); err != nil {
				return -1, err
			}
		} else {
			if _, err := rc.CompileToRegister(node.Default); err != nil {
				return -1, err
			}
		}

		endPos := len(rc.instructions)
		for _, jp := range caseEndJumps {
			rc.instructions[jp] = vm.EncodeRegisterInstructionBx(
				vm.OpRJump, 0, uint16(endPos))
		}

		rc.freeTempRegister(switchReg)
		return -1, nil

	case *ast.ForStatement:
		// Enter loop context for break/continue
		rc.enterRegisterLoop()
		defer rc.leaveRegisterLoop()

		// Initialize if present
		if node.Init != nil {
			_, err := rc.CompileToRegister(node.Init)
			if err != nil {
				return -1, err
			}
		}

		// Loop start
		loopStart := len(rc.instructions)

		// Compile condition
		condReg, err := rc.CompileToRegister(node.Condition)
		if err != nil {
			return -1, err
		}

		// Pick fast-path jump if the condition's static type is bool.
		jumpOp := vm.OpRJumpF
		if rc.inferExpressionType(node.Condition) == vm.BoolType {
			jumpOp = vm.OpRJumpFBool
		}

		// Jump if false (placeholder)
		jumpToEnd := rc.emitRBx(jumpOp, uint8(condReg), 9999)
		rc.freeTempRegister(condReg)

		// Compile body
		_, err = rc.CompileToRegister(node.Body)
		if err != nil {
			return -1, err
		}

		// Mark where continue should jump (before post statement)
		continuePos := len(rc.instructions)

		// Post statement
		if node.Post != nil {
			_, err = rc.CompileToRegister(node.Post)
			if err != nil {
				return -1, err
			}
		}

		// Jump back to start
		rc.emitRBx(vm.OpRJump, 0, uint16(loopStart))

		// Patch jump to end
		loopEnd := len(rc.instructions)
		rc.instructions[jumpToEnd] = vm.EncodeRegisterInstructionBx(jumpOp, uint8(condReg), uint16(loopEnd))

		// Patch all break jumps to jump to loopEnd
		loop := rc.currentRegisterLoop()
		for _, breakPos := range loop.breakJumps {
			rc.instructions[breakPos] = vm.EncodeRegisterInstructionBx(vm.OpRJump, 0, uint16(loopEnd))
		}

		// Patch all continue jumps to jump to continuePos
		for _, contPos := range loop.continueJumps {
			rc.instructions[contPos] = vm.EncodeRegisterInstructionBx(vm.OpRJump, 0, uint16(continuePos))
		}

		return -1, nil

	case *ast.BreakStatement:
		loop := rc.currentRegisterLoop()
		if loop == nil {
			return -1, fmt.Errorf("break statement outside of loop")
		}
		// Emit a jump with placeholder address
		pos := rc.emitRBx(vm.OpRJump, 0, 9999)
		// Record this position so we can patch it later
		loop.breakJumps = append(loop.breakJumps, pos)
		return -1, nil

	case *ast.ContinueStatement:
		loop := rc.currentRegisterLoop()
		if loop == nil {
			return -1, fmt.Errorf("continue statement outside of loop")
		}
		// Emit a jump with placeholder address
		pos := rc.emitRBx(vm.OpRJump, 0, 9999)
		// Record this position so we can patch it later
		loop.continueJumps = append(loop.continueJumps, pos)
		return -1, nil

	case *ast.BlockStatement:
		for _, stmt := range node.Statements {
			_, err := rc.CompileToRegister(stmt)
			if err != nil {
				return -1, err
			}
		}
		return -1, nil

	case *ast.ReturnStatement:
		if node.ReturnValue != nil {
			valueReg, err := rc.CompileToRegister(node.ReturnValue)
			if err != nil {
				return -1, err
			}
			// Return value in register
			rc.emitR(vm.OpRReturn, uint8(valueReg), 0, 0)
			rc.freeTempRegister(valueReg)
		} else {
			rc.emitR(vm.OpRReturnN, 0, 0, 0)
		}
		return -1, nil

	case *ast.CallExpression:
		// Check if this is a builtin call
		isBuiltin := false
		builtinIndex := 0
		if ident, ok := node.Function.(*ast.Identifier); ok {
			if symbol, ok := rc.symbolTable.Resolve(ident.Value); ok && symbol.Scope == BuiltinScope {
				isBuiltin = true
				builtinIndex = symbol.Index
			}
		}

		if isBuiltin {
			numArgs := len(node.Arguments)

			// New OpRBuiltin encoding: A=resultReg, B=builtinIndex, C=numArgs.
			// Args must sit at R(resultReg+1) ... R(resultReg+numArgs). We
			// reserve a contiguous block (resultReg + numArgs args) — recycling
			// freed slots when possible so register pressure stays bounded.
			baseReg := rc.allocateConsecutiveBlock(1 + numArgs)
			resultReg := baseReg
			argRegs := make([]int, numArgs)
			for i := 0; i < numArgs; i++ {
				argRegs[i] = baseReg + 1 + i
			}

			// Compile each argument into its slot.
			for i, arg := range node.Arguments {
				argReg, err := rc.CompileToRegister(arg)
				if err != nil {
					return -1, err
				}
				if argReg != argRegs[i] {
					rc.emitR(vm.OpRMove, uint8(argRegs[i]), uint8(argReg), 0)
					rc.freeTempRegister(argReg)
				}
			}

			if builtinIndex > 255 {
				return -1, fmt.Errorf("builtin index %d does not fit in uint8", builtinIndex)
			}
			if numArgs > 255 {
				return -1, fmt.Errorf("builtin call with %d args exceeds uint8 limit", numArgs)
			}
			rc.emitR(vm.OpRBuiltin, uint8(resultReg), uint8(builtinIndex), uint8(numArgs))

			// Return arg slots to the temp pool. Otherwise nextReg grows by
			// numArgs on every builtin call and pushes total register use past
			// the 8-bit instruction-field limit (256), silently wrapping.
			for _, ar := range argRegs {
				rc.freeTempRegister(ar)
			}

			return resultReg, nil
		}

		// Regular function call
		// Compile function expression to get function register
		fnReg, err := rc.CompileToRegister(node.Function)
		if err != nil {
			return -1, err
		}

		// OpRCall layout: A=resultReg, B=fnReg, C=argBaseReg (args at C..C+numArgs-1).
		// Allocate a contiguous block of numArgs slots for the args; recycle
		// freed slots so we don't grow nextReg unbounded across many calls.
		numArgs := len(node.Arguments)
		argBaseReg := rc.allocateConsecutiveBlock(numArgs)
		argRegs := make([]int, numArgs)
		for i := 0; i < numArgs; i++ {
			argRegs[i] = argBaseReg + i
		}

		// Compile each argument and move to its designated register
		for i, arg := range node.Arguments {
			argReg, err := rc.CompileToRegister(arg)
			if err != nil {
				return -1, err
			}
			if argReg != argRegs[i] {
				rc.emitR(vm.OpRMove, uint8(argRegs[i]), uint8(argReg), 0)
				rc.freeTempRegister(argReg)
			}
		}

		// Allocate result register
		resultReg := rc.allocateTempRegister()

		// Emit call instruction
		// OpRCall: R(A) = call R(B)(args starting at R(C))
		rc.emitR(vm.OpRCall, uint8(resultReg), uint8(fnReg), uint8(argBaseReg))

		// Free function register and arg slots — see builtin path for why.
		rc.freeTempRegister(fnReg)
		for _, ar := range argRegs {
			rc.freeTempRegister(ar)
		}

		return resultReg, nil

	case *ast.ArrayLiteral:
		// Create array
		arrayReg := rc.allocateTempRegister()
		rc.emitRBx(vm.OpRNewArray, uint8(arrayReg), uint16(len(node.Elements)))

		// Compile and store elements
		for i, elem := range node.Elements {
			elemReg, err := rc.CompileToRegister(elem)
			if err != nil {
				return -1, err
			}

			// Store element at index i
			idxReg := rc.allocateTempRegister()
			constIdx := rc.addConstant(vm.IntValue(int64(i)))
			rc.emitRBx(vm.OpRLoadK, uint8(idxReg), uint16(constIdx))

			rc.emitR(vm.OpRSetIdx, uint8(arrayReg), uint8(idxReg), uint8(elemReg))

			rc.freeTempRegister(idxReg)
			rc.freeTempRegister(elemReg)
		}

		return arrayReg, nil

	case *ast.IndexExpression:
		// Array/map access: container[index]
		containerReg, err := rc.CompileToRegister(node.Left)
		if err != nil {
			return -1, err
		}

		indexReg, err := rc.CompileToRegister(node.Index)
		if err != nil {
			return -1, err
		}

		// OpRGetIdx handles array and string; maps need OpRMapGet.
		getOp := vm.OpRGetIdx
		if rc.inferExpressionType(node.Left) == vm.MapType {
			getOp = vm.OpRMapGet
		}
		resultReg := rc.allocateTempRegister()
		rc.emitR(getOp, uint8(resultReg), uint8(containerReg), uint8(indexReg))

		rc.freeTempRegister(containerReg)
		rc.freeTempRegister(indexReg)

		return resultReg, nil

	case *ast.MapLiteral:
		// Create map
		mapReg := rc.allocateTempRegister()
		rc.emitR(vm.OpRNewMap, uint8(mapReg), 0, 0)

		// Compile and store key-value pairs. Use OpRMapSet (not OpRSetIdx,
		// which assumes an *ArrayValue and would segfault on a map pointer).
		for key, value := range node.Pairs {
			keyReg, err := rc.CompileToRegister(key)
			if err != nil {
				return -1, err
			}

			valueReg, err := rc.CompileToRegister(value)
			if err != nil {
				return -1, err
			}

			rc.emitR(vm.OpRMapSet, uint8(mapReg), uint8(keyReg), uint8(valueReg))

			rc.freeTempRegister(keyReg)
			rc.freeTempRegister(valueReg)
		}

		return mapReg, nil

	case *ast.StructLiteral:
		// Create struct instance and populate each field by name.
		structReg := rc.allocateTempRegister()
		typeIdx := rc.addConstant(vm.StringValue(node.Name.Value))
		rc.emitRBx(vm.OpRNewStruct, uint8(structReg), uint16(typeIdx))

		for fieldName, fieldValue := range node.Fields {
			valueReg, err := rc.CompileToRegister(fieldValue)
			if err != nil {
				return -1, err
			}
			fieldIdx := rc.addConstant(vm.StringValue(fieldName))
			if fieldIdx >= 256 {
				return -1, fmt.Errorf("too many constants: field name %q exceeds register-VM uint8 limit", fieldName)
			}
			// OpRSetField: R(A).field(K(B)) = R(C)
			rc.emitR(vm.OpRSetField, uint8(structReg), uint8(fieldIdx), uint8(valueReg))
			rc.freeTempRegister(valueReg)
		}

		return structReg, nil

	case *ast.FieldAccessExpression:
		// Struct field access: obj.field
		objReg, err := rc.CompileToRegister(node.Left)
		if err != nil {
			return -1, err
		}

		fieldIdx := rc.addConstant(vm.StringValue(node.Field.Value))
		if fieldIdx >= 256 {
			return -1, fmt.Errorf("too many constants: field name %q exceeds register-VM uint8 limit", node.Field.Value)
		}

		resultReg := rc.allocateTempRegister()
		// OpRGetField: R(A) = R(B).field(K(C))
		rc.emitR(vm.OpRGetField, uint8(resultReg), uint8(objReg), uint8(fieldIdx))

		rc.freeTempRegister(objReg)

		return resultReg, nil

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
		rc.functionSigs[node.Name.Value] = funcType
		rc.typeInfo[node.Name.Value] = funcType

		// Define the function name in the current scope BEFORE compiling the body
		// This allows recursive calls
		symbol := rc.symbolTable.Define(node.Name.Value)

		// Save current compiler state
		savedInstructions := rc.instructions
		savedRegisters := rc.registers
		savedNextReg := rc.nextReg
		savedMaxRegs := rc.MaxRegs
		savedTempRegs := rc.tempRegs

		// Create new state for function body
		rc.instructions = []vm.RegisterInstruction{}
		rc.registers = make(map[string]int)
		rc.nextReg = 0
		rc.MaxRegs = 0
		rc.tempRegs = []int{}

		// Enter scope for symbol table (uses embedded Compiler's method)
		rc.Compiler.enterScope()

		// Store the previous return type and set current one
		prevReturnType := rc.currentFunctionRT
		rc.currentFunctionRT = returnType

		// Define parameters in the new scope - parameters occupy first registers
		for i, param := range node.Parameters {
			// Define in symbol table
			rc.symbolTable.Define(param.Name.Value)
			// Allocate register
			rc.allocateRegister(param.Name.Value)
			// Track parameter types
			rc.typeInfo[param.Name.Value] = paramTypes[i]
		}

		// Compile function body
		_, err := rc.CompileToRegister(node.Body)
		if err != nil {
			return -1, err
		}

		// If the last instruction is not a return, add an implicit return nil
		needsReturn := len(rc.instructions) == 0
		if !needsReturn {
			lastOp, _, _, _ := rc.instructions[len(rc.instructions)-1].Decode()
			needsReturn = (lastOp != vm.OpRReturn && lastOp != vm.OpRReturnN)
		}
		if needsReturn {
			// Check if function expects a specific non-nil return value
			if returnType != nil && !returnType.Equals(NilType) && !returnType.Equals(AnyTypeVal) {
				return -1, fmt.Errorf("function %s must return %s", node.Name.Value, returnType.String())
			}
			rc.emitR(vm.OpRReturnN, 0, 0, 0)
		}

		// Restore previous return type
		rc.currentFunctionRT = prevReturnType

		// Get the compiled instructions
		numLocals := rc.MaxRegs
		functionInstructions := rc.instructions

		// Leave scope for symbol table (uses embedded Compiler's method)
		rc.Compiler.leaveScope()

		// Restore compiler state
		rc.instructions = savedInstructions
		rc.registers = savedRegisters
		rc.nextReg = savedNextReg
		rc.MaxRegs = savedMaxRegs
		rc.tempRegs = savedTempRegs

		// Create the function object with register bytecode
		compiledFn := &vm.Function{
			Name:                 node.Name.Value,
			NumParams:            len(node.Parameters),
			NumLocals:            numLocals,
			RegisterInstructions: functionInstructions,
			Instructions:         nil, // No stack bytecode
			Constants:            rc.constants, // Share constants with parent
		}

		// Add function to constant pool
		fnIndex := rc.addConstant(vm.NewFunctionValue(compiledFn))

		// Load function constant into a register
		if symbol.Scope == GlobalScope {
			// Global function - load into temp then store to global
			tempReg := rc.allocateTempRegister()
			rc.emitRBx(vm.OpRLoadK, uint8(tempReg), uint16(fnIndex))
			rc.emitRBx(vm.OpRStoreGlobal, uint8(tempReg), uint16(symbol.Index))
			rc.freeTempRegister(tempReg)
		} else {
			// Local function - load into variable register
			varReg := rc.allocateRegister(node.Name.Value)
			rc.emitRBx(vm.OpRLoadK, uint8(varReg), uint16(fnIndex))
		}

		return -1, nil

	default:
		return -1, fmt.Errorf("register compilation not yet implemented for node type: %T", node)
	}
}
