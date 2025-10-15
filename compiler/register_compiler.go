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
}

// LiveRange tracks when a variable is live
type LiveRange struct {
	start int // First instruction where variable is defined
	end   int // Last instruction where variable is used
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
	}
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

// freeTempRegister marks a temporary register as available
func (rc *RegisterCompiler) freeTempRegister(reg int) {
	rc.tempRegs = append(rc.tempRegs, reg)
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
func (rc *RegisterCompiler) CompileToRegister(node ast.Node) error {
	switch node := node.(type) {
	case *ast.Program:
		for _, s := range node.Statements {
			if err := rc.CompileToRegister(s); err != nil {
				return err
			}
		}

	case *ast.ExpressionStatement:
		// Compile expression but don't store result (discard)
		if err := rc.CompileToRegister(node.Expression); err != nil {
			return err
		}

	case *ast.IntegerLiteral:
		// Load constant into temp register
		constIndex := rc.addConstant(vm.IntValue(node.Value))
		tempReg := rc.allocateTempRegister()
		rc.emitRBx(vm.OpRLoadK, uint8(tempReg), uint16(constIndex))
		return nil

	case *ast.FloatLiteral:
		constIndex := rc.addConstant(vm.FloatValue(node.Value))
		tempReg := rc.allocateTempRegister()
		rc.emitRBx(vm.OpRLoadK, uint8(tempReg), uint16(constIndex))
		return nil

	case *ast.BooleanLiteral:
		constIndex := rc.addConstant(vm.BoolValue(node.Value))
		tempReg := rc.allocateTempRegister()
		rc.emitRBx(vm.OpRLoadK, uint8(tempReg), uint16(constIndex))
		return nil

	case *ast.StringLiteral:
		constIndex := rc.addConstant(vm.StringValue(node.Value))
		tempReg := rc.allocateTempRegister()
		rc.emitRBx(vm.OpRLoadK, uint8(tempReg), uint16(constIndex))
		return nil

	case *ast.Identifier:
		// Variable reference - already in register
		if reg, exists := rc.registers[node.Value]; exists {
			// Return the register number somehow
			_ = reg
			return nil
		}
		return fmt.Errorf("undefined variable: %s", node.Value)

	case *ast.VarStatement:
		// Allocate register for variable
		reg := rc.allocateRegister(node.Name.Value)

		// Track variable type
		if node.Type != nil {
			rc.varTypes[node.Name.Value] = typeAnnotationToValueType(node.Type)
		}

		// Compile initializer value if present
		if node.Value != nil {
			// For now, compile value and assume it's in a temp register
			if err := rc.CompileToRegister(node.Value); err != nil {
				return err
			}
			// TODO: Move from temp to allocated register
			_ = reg
		}

	case *ast.InfixExpression:
		// Compile left and right operands
		// For now, simplified version
		if err := rc.CompileToRegister(node.Left); err != nil {
			return err
		}
		if err := rc.CompileToRegister(node.Right); err != nil {
			return err
		}

		// Determine types
		leftType := rc.inferExpressionType(node.Left)
		rightType := rc.inferExpressionType(node.Right)

		// Allocate result register
		resultReg := rc.allocateTempRegister()

		// Emit type-specific operation
		// TODO: Track which registers hold left/right values
		leftReg := uint8(0)  // Placeholder
		rightReg := uint8(1) // Placeholder

		switch node.Operator {
		case "+":
			if leftType == vm.IntType && rightType == vm.IntType {
				rc.emitR(vm.OpRAddInt, uint8(resultReg), leftReg, rightReg)
			} else {
				rc.emitR(vm.OpRAddFloat, uint8(resultReg), leftReg, rightReg)
			}
		case "-":
			if leftType == vm.IntType && rightType == vm.IntType {
				rc.emitR(vm.OpRSubInt, uint8(resultReg), leftReg, rightReg)
			} else {
				rc.emitR(vm.OpRSubFloat, uint8(resultReg), leftReg, rightReg)
			}
		case "*":
			// Check for square pattern (x * x)
			leftIdent, leftIsIdent := node.Left.(*ast.Identifier)
			rightIdent, rightIsIdent := node.Right.(*ast.Identifier)
			if leftIsIdent && rightIsIdent && leftIdent.Value == rightIdent.Value {
				// Square optimization
				if leftType == vm.FloatType {
					rc.emitR(vm.OpRSquareFloat, uint8(resultReg), leftReg, 0)
				} else {
					rc.emitR(vm.OpRSquareInt, uint8(resultReg), leftReg, 0)
				}
			} else {
				if leftType == vm.IntType && rightType == vm.IntType {
					rc.emitR(vm.OpRMulInt, uint8(resultReg), leftReg, rightReg)
				} else {
					rc.emitR(vm.OpRMulFloat, uint8(resultReg), leftReg, rightReg)
				}
			}
		case "/":
			if leftType == vm.IntType && rightType == vm.IntType {
				rc.emitR(vm.OpRDivInt, uint8(resultReg), leftReg, rightReg)
			} else {
				rc.emitR(vm.OpRDivFloat, uint8(resultReg), leftReg, rightReg)
			}
		case "%":
			rc.emitR(vm.OpRModInt, uint8(resultReg), leftReg, rightReg)

		// Comparisons
		case "==":
			if leftType == vm.IntType {
				rc.emitR(vm.OpREqInt, uint8(resultReg), leftReg, rightReg)
			} else if leftType == vm.FloatType {
				rc.emitR(vm.OpREqFloat, uint8(resultReg), leftReg, rightReg)
			} else if leftType == vm.StringType {
				rc.emitR(vm.OpREqString, uint8(resultReg), leftReg, rightReg)
			} else {
				rc.emitR(vm.OpREqBool, uint8(resultReg), leftReg, rightReg)
			}
		case "!=":
			if leftType == vm.IntType {
				rc.emitR(vm.OpRNeInt, uint8(resultReg), leftReg, rightReg)
			} else if leftType == vm.FloatType {
				rc.emitR(vm.OpRNeFloat, uint8(resultReg), leftReg, rightReg)
			} else if leftType == vm.StringType {
				rc.emitR(vm.OpRNeString, uint8(resultReg), leftReg, rightReg)
			} else {
				rc.emitR(vm.OpRNeBool, uint8(resultReg), leftReg, rightReg)
			}
		case "<":
			if leftType == vm.IntType {
				rc.emitR(vm.OpRLtInt, uint8(resultReg), leftReg, rightReg)
			} else {
				rc.emitR(vm.OpRLtFloat, uint8(resultReg), leftReg, rightReg)
			}
		case ">":
			if leftType == vm.IntType {
				rc.emitR(vm.OpRGtInt, uint8(resultReg), leftReg, rightReg)
			} else {
				rc.emitR(vm.OpRGtFloat, uint8(resultReg), leftReg, rightReg)
			}
		case "<=":
			if leftType == vm.IntType {
				rc.emitR(vm.OpRLeInt, uint8(resultReg), leftReg, rightReg)
			} else {
				rc.emitR(vm.OpRLeFloat, uint8(resultReg), leftReg, rightReg)
			}
		case ">=":
			if leftType == vm.IntType {
				rc.emitR(vm.OpRGeInt, uint8(resultReg), leftReg, rightReg)
			} else {
				rc.emitR(vm.OpRGeFloat, uint8(resultReg), leftReg, rightReg)
			}

		// Logical
		case "&&":
			rc.emitR(vm.OpRAnd, uint8(resultReg), leftReg, rightReg)
		case "||":
			rc.emitR(vm.OpROr, uint8(resultReg), leftReg, rightReg)

		default:
			return fmt.Errorf("unknown operator: %s", node.Operator)
		}

	case *ast.PrefixExpression:
		if err := rc.CompileToRegister(node.Right); err != nil {
			return err
		}

		resultReg := rc.allocateTempRegister()
		operandReg := uint8(0) // TODO: track actual register

		switch node.Operator {
		case "!":
			rc.emitR(vm.OpRNot, uint8(resultReg), operandReg, 0)
		case "-":
			exprType := rc.inferExpressionType(node.Right)
			if exprType == vm.IntType {
				rc.emitR(vm.OpRNegInt, uint8(resultReg), operandReg, 0)
			} else {
				rc.emitR(vm.OpRNegFloat, uint8(resultReg), operandReg, 0)
			}
		}

	case *ast.ForStatement:
		// Initialize if present
		if node.Init != nil {
			if err := rc.CompileToRegister(node.Init); err != nil {
				return err
			}
		}

		// Loop start
		loopStart := len(rc.instructions)

		// Compile condition
		if err := rc.CompileToRegister(node.Condition); err != nil {
			return err
		}

		// Jump if false (placeholder)
		condReg := uint8(0) // TODO: track condition register
		jumpToEnd := rc.emitRBx(vm.OpRJumpF, condReg, 9999)

		// Compile body
		if err := rc.CompileToRegister(node.Body); err != nil {
			return err
		}

		// Post statement
		if node.Post != nil {
			if err := rc.CompileToRegister(node.Post); err != nil {
				return err
			}
		}

		// Jump back to start
		rc.emitRBx(vm.OpRJump, 0, uint16(loopStart))

		// Patch jump to end
		loopEnd := len(rc.instructions)
		rc.instructions[jumpToEnd] = vm.EncodeRegisterInstructionBx(vm.OpRJumpF, condReg, uint16(loopEnd))

	case *ast.BlockStatement:
		for _, stmt := range node.Statements {
			if err := rc.CompileToRegister(stmt); err != nil {
				return err
			}
		}

	case *ast.ReturnStatement:
		if node.ReturnValue != nil {
			if err := rc.CompileToRegister(node.ReturnValue); err != nil {
				return err
			}
			// Return value in register
			rc.emitR(vm.OpRReturn, 0, 0, 0) // TODO: specify return register
		} else {
			rc.emitR(vm.OpRReturnN, 0, 0, 0)
		}

	default:
		return fmt.Errorf("register compilation not yet implemented for node type: %T", node)
	}

	return nil
}
