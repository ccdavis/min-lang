package vm

import (
	"fmt"
)

const (
	MaxRegisters = 256  // Max registers per function
	InitialRegs  = 32   // Initial register allocation
)

// RegisterFrame represents a function call frame in the register VM
type RegisterFrame struct {
	function     *Function
	instructions []RegisterInstruction
	pc           int      // Program counter
	baseReg      int      // Base register for this frame
	registers    []Value  // Local register window
}

// RegisterVM is a register-based virtual machine
type RegisterVM struct {
	constants []Value               // Constant pool
	globals   []Value               // Global variables

	// Register file - grows as needed
	registers []Value

	// Function call stack
	frames      []*RegisterFrame
	frameIndex  int

	// Current frame cache (for performance)
	currentFrame *RegisterFrame
}

// NewRegisterVM creates a new register-based VM
func NewRegisterVM(bytecode *RegisterBytecode) *RegisterVM {
	vm := &RegisterVM{
		constants:  bytecode.Constants,
		globals:    make([]Value, GlobalsSize),
		registers:  make([]Value, InitialRegs),
		frames:     make([]*RegisterFrame, MaxFrames),
		frameIndex: 0,
	}

	// Create main frame
	mainFrame := &RegisterFrame{
		function:     bytecode.MainFunction,
		instructions: bytecode.Instructions,
		pc:           0,
		baseReg:      0,
		registers:    vm.registers, // Main frame uses full register file
	}

	vm.frames[0] = mainFrame
	vm.frameIndex = 1
	vm.currentFrame = mainFrame

	return vm
}

// RegisterBytecode represents compiled register bytecode
type RegisterBytecode struct {
	Instructions []RegisterInstruction
	Constants    []Value
	MainFunction *Function
}

// Run executes the register bytecode
func (vm *RegisterVM) Run() error {
	frame := vm.currentFrame
	ins := frame.instructions
	pc := frame.pc
	regs := frame.registers

	// Main execution loop
	for {
		if pc >= len(ins) {
			// Check if we're in main frame
			if vm.frameIndex <= 1 {
				return nil
			}
			// Return from function with no value
			if err := vm.returnFromFunction(0); err != nil {
				return err
			}
			// Reload frame
			frame = vm.currentFrame
			ins = frame.instructions
			pc = frame.pc
			regs = frame.registers
			continue
		}

		instruction := ins[pc]
		op, a, b, c := instruction.Decode()
		pc++

		switch op {
		// Load/Move operations
		case OpRLoadK:
			_, a, bx := ins[pc-1].DecodeBx()
			regs[a] = vm.constants[bx]

		case OpRMove:
			regs[a] = regs[b]

		// Arithmetic operations (NO TYPE CHECKS - compiler guarantees)
		case OpRAddInt:
			regs[a] = IntValue(regs[b].AsInt() + regs[c].AsInt())

		case OpRAddFloat:
			regs[a] = FloatValue(regs[b].AsFloat() + regs[c].AsFloat())

		case OpRSubInt:
			regs[a] = IntValue(regs[b].AsInt() - regs[c].AsInt())

		case OpRSubFloat:
			regs[a] = FloatValue(regs[b].AsFloat() - regs[c].AsFloat())

		case OpRMulInt:
			regs[a] = IntValue(regs[b].AsInt() * regs[c].AsInt())

		case OpRMulFloat:
			regs[a] = FloatValue(regs[b].AsFloat() * regs[c].AsFloat())

		case OpRDivInt:
			divisor := regs[c].AsInt()
			if divisor == 0 {
				return ErrDivisionByZero
			}
			regs[a] = IntValue(regs[b].AsInt() / divisor)

		case OpRDivFloat:
			divisor := regs[c].AsFloat()
			if divisor == 0.0 {
				return ErrDivisionByZero
			}
			regs[a] = FloatValue(regs[b].AsFloat() / divisor)

		case OpRModInt:
			divisor := regs[c].AsInt()
			if divisor == 0 {
				return ErrModuloByZero
			}
			regs[a] = IntValue(regs[b].AsInt() % divisor)

		case OpRNegInt:
			regs[a] = IntValue(-regs[b].AsInt())

		case OpRNegFloat:
			regs[a] = FloatValue(-regs[b].AsFloat())

		// Comparison operations (NO TYPE CHECKS)
		case OpREqInt:
			regs[a] = BoolValue(regs[b].AsInt() == regs[c].AsInt())

		case OpREqFloat:
			regs[a] = BoolValue(regs[b].AsFloat() == regs[c].AsFloat())

		case OpREqBool:
			regs[a] = BoolValue(regs[b].AsBool() == regs[c].AsBool())

		case OpREqString:
			regs[a] = BoolValue(regs[b].AsString() == regs[c].AsString())

		case OpRNeInt:
			regs[a] = BoolValue(regs[b].AsInt() != regs[c].AsInt())

		case OpRNeFloat:
			regs[a] = BoolValue(regs[b].AsFloat() != regs[c].AsFloat())

		case OpRNeBool:
			regs[a] = BoolValue(regs[b].AsBool() != regs[c].AsBool())

		case OpRNeString:
			regs[a] = BoolValue(regs[b].AsString() != regs[c].AsString())

		case OpRLtInt:
			regs[a] = BoolValue(regs[b].AsInt() < regs[c].AsInt())

		case OpRLtFloat:
			regs[a] = BoolValue(regs[b].AsFloat() < regs[c].AsFloat())

		case OpRGtInt:
			regs[a] = BoolValue(regs[b].AsInt() > regs[c].AsInt())

		case OpRGtFloat:
			regs[a] = BoolValue(regs[b].AsFloat() > regs[c].AsFloat())

		case OpRLeInt:
			regs[a] = BoolValue(regs[b].AsInt() <= regs[c].AsInt())

		case OpRLeFloat:
			regs[a] = BoolValue(regs[b].AsFloat() <= regs[c].AsFloat())

		case OpRGeInt:
			regs[a] = BoolValue(regs[b].AsInt() >= regs[c].AsInt())

		case OpRGeFloat:
			regs[a] = BoolValue(regs[b].AsFloat() >= regs[c].AsFloat())

		// Logical operations
		case OpRAnd:
			regs[a] = BoolValue(regs[b].IsTruthy() && regs[c].IsTruthy())

		case OpROr:
			regs[a] = BoolValue(regs[b].IsTruthy() || regs[c].IsTruthy())

		case OpRNot:
			regs[a] = BoolValue(!regs[b].IsTruthy())

		// Control flow
		case OpRJump:
			_, _, bx := ins[pc-1].DecodeBx()
			pc = int(bx)

		case OpRJumpT:
			_, _, bx := ins[pc-1].DecodeBx()
			if regs[a].IsTruthy() {
				pc = int(bx)
			}

		case OpRJumpF:
			_, _, bx := ins[pc-1].DecodeBx()
			if !regs[a].IsTruthy() {
				pc = int(bx)
			}

		case OpRReturn:
			// Save PC before calling returnFromFunction
			frame.pc = pc
			if err := vm.returnFromFunction(int(a)); err != nil {
				return err
			}
			// Reload frame after return
			frame = vm.currentFrame
			ins = frame.instructions
			pc = frame.pc
			regs = frame.registers

		case OpRReturnN:
			frame.pc = pc
			if err := vm.returnFromFunction(-1); err != nil {
				return err
			}
			frame = vm.currentFrame
			ins = frame.instructions
			pc = frame.pc
			regs = frame.registers

		// Function calls
		case OpRCall:
			// R(A) = call R(B)(R(C)...R(C+numArgs))
			// B = function register
			// C = first arg register
			// a = result register
			// Decode number of arguments from next byte
			frame.pc = pc
			if err := vm.callFunction(int(b), int(c), int(a)); err != nil {
				return err
			}
			// Reload frame
			frame = vm.currentFrame
			ins = frame.instructions
			pc = frame.pc
			regs = frame.registers

		case OpRBuiltin:
			_, a, bx := ins[pc-1].DecodeBx()
			builtinIndex := int(bx)
			// Builtin calls: args in R(C)...R(C+n), result in R(A)
			// For now, use c as first arg register
			if err := vm.callBuiltin(builtinIndex, int(c), int(a)); err != nil {
				return err
			}

		// Array operations
		case OpRNewArray:
			_, a, bx := ins[pc-1].DecodeBx()
			size := int(bx)
			regs[a] = NewArrayValue(size)

		case OpRGetIdx:
			// R(A) = R(B)[R(C)]
			container := regs[b]
			index := regs[c]

			switch container.Type {
			case ArrayType:
				idx := int(index.AsInt())
				arrayVal := container.AsArray()
				if idx < 0 || idx >= len(arrayVal.Elements) {
					return fmt.Errorf("array index out of bounds: %d", idx)
				}
				regs[a] = arrayVal.Elements[idx]

			case StringType:
				idx := int(index.AsInt())
				str := container.AsString()
				if idx < 0 || idx >= len(str) {
					return fmt.Errorf("string index out of bounds: %d", idx)
				}
				regs[a] = StringValue(string(str[idx]))
			}

		case OpRSetIdx:
			// R(A)[R(B)] = R(C)
			container := regs[a]
			index := regs[b]
			value := regs[c]

			idx := int(index.AsInt())
			arrayVal := container.AsArray()
			if idx < 0 || idx >= len(arrayVal.Elements) {
				return fmt.Errorf("array index out of bounds: %d", idx)
			}
			arrayVal.Elements[idx] = value

		// Map operations
		case OpRNewMap:
			regs[a] = NewMapValue()

		case OpRMapGet:
			// R(A) = R(B)[R(C)]
			mapVal := regs[b].AsMap()
			key := regs[c].ToMapKey()
			if val, ok := mapVal.Pairs[key]; ok {
				regs[a] = val
			} else {
				regs[a] = NilValue()
			}

		case OpRMapSet:
			// R(A)[R(B)] = R(C)
			mapVal := regs[a].AsMap()
			key := regs[b].ToMapKey()
			value := regs[c]
			mapVal.Pairs[key] = value

		// Struct operations
		case OpRNewStruct:
			// TODO: Implement struct creation
			regs[a] = NilValue()

		case OpRGetField:
			// R(A) = R(B).field(C) - C is field offset
			structVal := regs[b].AsStruct()
			if int(c) >= len(structVal.FieldsArray) {
				return fmt.Errorf("field offset out of bounds: %d", c)
			}
			regs[a] = structVal.FieldsArray[c]

		case OpRSetField:
			// R(A).field(B) = R(C) - B is field offset
			structVal := regs[a].AsStruct()
			if int(b) >= len(structVal.FieldsArray) {
				return fmt.Errorf("field offset out of bounds: %d", b)
			}
			structVal.FieldsArray[b] = regs[c]

		// Global operations
		case OpRLoadGlobal:
			_, a, bx := ins[pc-1].DecodeBx()
			regs[a] = vm.globals[bx]

		case OpRStoreGlobal:
			_, a, bx := ins[pc-1].DecodeBx()
			vm.globals[bx] = regs[a]

		// String operations
		case OpRConcat:
			regs[a] = StringValue(regs[b].AsString() + regs[c].AsString())

		// Optimized operations with immediate constants
		case OpRAddConstInt:
			_, a, bx := ins[pc-1].DecodeBx()
			regs[a] = IntValue(regs[b].AsInt() + vm.constants[bx].AsInt())

		case OpRAddConstFloat:
			_, a, bx := ins[pc-1].DecodeBx()
			regs[a] = FloatValue(regs[b].AsFloat() + vm.constants[bx].AsFloat())

		case OpRMulConstInt:
			_, a, bx := ins[pc-1].DecodeBx()
			regs[a] = IntValue(regs[b].AsInt() * vm.constants[bx].AsInt())

		case OpRMulConstFloat:
			_, a, bx := ins[pc-1].DecodeBx()
			regs[a] = FloatValue(regs[b].AsFloat() * vm.constants[bx].AsFloat())

		// Special optimizations
		case OpRSquareInt:
			val := regs[b].AsInt()
			regs[a] = IntValue(val * val)

		case OpRSquareFloat:
			val := regs[b].AsFloat()
			regs[a] = FloatValue(val * val)

		case OpRHalt:
			return nil

		default:
			return fmt.Errorf("unknown register opcode: %d", op)
		}

		// Update PC in frame
		frame.pc = pc
	}
}

// callFunction handles function calls in the register VM
func (vm *RegisterVM) callFunction(fnReg, argReg, resultReg int) error {
	function := vm.currentFrame.registers[fnReg]

	// Only handle Function and Closure types
	var fn *Function
	switch function.Type {
	case FunctionType:
		fn = function.AsFunction()
	case ClosureType:
		fn = function.AsClosure().Fn
	default:
		return ErrCallingNonFunction
	}

	// Allocate new frame
	if vm.frameIndex >= MaxFrames {
		return fmt.Errorf("call stack overflow")
	}

	newFrame := vm.frames[vm.frameIndex]
	if newFrame == nil {
		newFrame = &RegisterFrame{}
		vm.frames[vm.frameIndex] = newFrame
	}

	// Set up new frame
	newFrame.function = fn
	newFrame.instructions = convertToRegisterInstructions(fn.Instructions) // TODO: implement
	newFrame.pc = 0
	newFrame.baseReg = argReg

	// Allocate registers for new frame if needed
	numRegs := fn.NumLocals + fn.NumParams + 16 // Extra for temporaries
	if len(vm.registers) < vm.currentFrame.baseReg + argReg + numRegs {
		// Grow register file
		newRegs := make([]Value, vm.currentFrame.baseReg + argReg + numRegs + 32)
		copy(newRegs, vm.registers)
		vm.registers = newRegs
	}

	newFrame.registers = vm.registers[argReg:argReg+numRegs]

	vm.frameIndex++
	vm.currentFrame = newFrame

	return nil
}

// returnFromFunction handles returns in the register VM
func (vm *RegisterVM) returnFromFunction(resultReg int) error {
	if vm.frameIndex <= 1 {
		return nil // Main frame, exit program
	}

	// Save return value if any
	var returnValue Value
	if resultReg >= 0 {
		returnValue = vm.currentFrame.registers[resultReg]
	}

	// Pop frame
	vm.frameIndex--
	vm.currentFrame = vm.frames[vm.frameIndex-1]

	// Store return value
	if resultReg >= 0 {
		// TODO: Store in caller's result register
		_ = returnValue
	}

	return nil
}

// callBuiltin handles builtin function calls
func (vm *RegisterVM) callBuiltin(index, argReg, resultReg int) error {
	if index >= len(Builtins) {
		return fmt.Errorf("unknown builtin: %d", index)
	}

	builtin := Builtins[index]

	// Collect arguments from registers
	// For now, assume max 4 args
	args := make([]Value, 4)
	for i := 0; i < 4; i++ {
		if argReg+i < len(vm.currentFrame.registers) {
			args[i] = vm.currentFrame.registers[argReg+i]
		}
	}

	result := builtin.Fn(args...)
	vm.currentFrame.registers[resultReg] = result

	return nil
}

// Helper function to convert stack bytecode to register bytecode
// This is a placeholder - real implementation would be in compiler
func convertToRegisterInstructions(stackBytecode []byte) []RegisterInstruction {
	// TODO: Implement proper conversion or have compiler generate directly
	return []RegisterInstruction{}
}
