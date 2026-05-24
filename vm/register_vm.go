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
	resultReg    int      // Where to store return value in caller's frame
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

// NewRegisterVM creates a new register-based VM.
// All frames share vm.registers as a single backing array; each frame's
// .registers field is a slice view starting at its baseReg.
func NewRegisterVM(bytecode *RegisterBytecode) *RegisterVM {
	numRegs := bytecode.MainFunction.NumLocals
	if numRegs < InitialRegs {
		numRegs = InitialRegs
	}

	vm := &RegisterVM{
		constants:  bytecode.Constants,
		globals:    make([]Value, GlobalsSize),
		registers:  make([]Value, numRegs),
		frames:     make([]*RegisterFrame, MaxFrames),
		frameIndex: 0,
	}

	mainFrame := &RegisterFrame{
		function:     bytecode.MainFunction,
		instructions: bytecode.Instructions,
		pc:           0,
		baseReg:      0,
		registers:    vm.registers,
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

	// Cache frequently accessed VM fields to reduce pointer dereferences
	constants := vm.constants
	globals := vm.globals

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
		pc++

		// Optimized decode: Always decode ABC format (just bit shifts)
		// Bx-format instructions will recompute locally when needed
		op := RegisterOpCode(instruction >> 24)
		a := uint8((instruction >> 16) & 0xFF)
		b := uint8((instruction >> 8) & 0xFF)
		c := uint8(instruction & 0xFF)

		switch op {
		// Load/Move operations
		case OpRLoadK:
			// Bx format: bottom 16 bits contain constant index
			bx := uint16(instruction & 0xFFFF)
			regs[a] = constants[bx]

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
			bx := uint16(instruction & 0xFFFF)
			pc = int(bx)

		case OpRJumpT:
			if regs[a].IsTruthy() {
				bx := uint16(instruction & 0xFFFF)
				pc = int(bx)
			}

		case OpRJumpF:
			if !regs[a].IsTruthy() {
				bx := uint16(instruction & 0xFFFF)
				pc = int(bx)
			}

		case OpRJumpFBool:
			// Fast path: condition is known-bool, skip the type switch in IsTruthy.
			if regs[a].Data == 0 {
				bx := uint16(instruction & 0xFFFF)
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
			// New encoding: A=resultReg, B=builtinIndex (full uint8), C=numArgs.
			// Args live at R(A+1) ... R(A+numArgs), immediately after the result
			// slot. The old layout packed builtinIndex into 4 bits and silently
			// wrapped indices 16-20 (split/substring/int/float/string) into
			// print/len/delete/append/keys.
			builtinIndex := int(b)
			numArgs := int(c)
			argBaseReg := int(a) + 1
			if err := vm.callBuiltin(builtinIndex, argBaseReg, int(a), numArgs); err != nil {
				return err
			}

		// Array operations
		case OpRNewArray:
			bx := uint16(instruction & 0xFFFF)
			regs[a] = NewArrayValue(int(bx))

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

		// Struct operations.
		// Name-based access via the constant pool. Field-name constant index
		// must fit in a single byte; the compiler enforces this and falls back
		// to an error if exceeded.
		case OpRNewStruct:
			// R(A) = new struct with type name K(Bx)
			bx := uint16(instruction & 0xFFFF)
			typeName := constants[bx].AsString()
			regs[a] = NewStructValue(typeName, make(map[string]Value, 4))

		case OpRGetField:
			// R(A) = R(B).field(K(C)) — C is constant index for field name
			structVal := regs[b].AsStruct()
			fieldName := constants[c].AsString()
			val, ok := structVal.Fields[fieldName]
			if !ok {
				return fmt.Errorf("field %s not found in struct %s", fieldName, structVal.TypeName)
			}
			regs[a] = val

		case OpRSetField:
			// R(A).field(K(B)) = R(C) — B is constant index for field name
			structVal := regs[a].AsStruct()
			fieldName := constants[b].AsString()
			structVal.Fields[fieldName] = regs[c]

		// Global operations
		case OpRLoadGlobal:
			bx := uint16(instruction & 0xFFFF)
			regs[a] = globals[bx]

		case OpRStoreGlobal:
			bx := uint16(instruction & 0xFFFF)
			globals[bx] = regs[a]

		// String operations.
		// Use .String() (not .AsString()) so mixed types like `"x=" + 3` work:
		// the stack VM's OpAddString does the same.
		case OpRConcat:
			regs[a] = StringValue(regs[b].String() + regs[c].String())

		// Optimized operations with immediate constants (use c as const index)
		case OpRAddConstInt:
			regs[a] = IntValue(regs[b].AsInt() + constants[c].AsInt())

		case OpRAddConstFloat:
			regs[a] = FloatValue(regs[b].AsFloat() + constants[c].AsFloat())

		case OpRMulConstInt:
			regs[a] = IntValue(regs[b].AsInt() * constants[c].AsInt())

		case OpRMulConstFloat:
			regs[a] = FloatValue(regs[b].AsFloat() * constants[c].AsFloat())

		case OpRLtConstInt:
			regs[a] = BoolValue(regs[b].AsInt() < constants[c].AsInt())

		case OpRLtConstFloat:
			regs[a] = BoolValue(regs[b].AsFloat() < constants[c].AsFloat())

		case OpRGtConstInt:
			regs[a] = BoolValue(regs[b].AsInt() > constants[c].AsInt())

		case OpRGtConstFloat:
			regs[a] = BoolValue(regs[b].AsFloat() > constants[c].AsFloat())

		case OpRLeConstInt:
			regs[a] = BoolValue(regs[b].AsInt() <= constants[c].AsInt())

		case OpRLeConstFloat:
			regs[a] = BoolValue(regs[b].AsFloat() <= constants[c].AsFloat())

		case OpRGeConstInt:
			regs[a] = BoolValue(regs[b].AsInt() >= constants[c].AsInt())

		case OpRGeConstFloat:
			regs[a] = BoolValue(regs[b].AsFloat() >= constants[c].AsFloat())

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

		// Note: frame.pc is only updated before function calls/returns
		// Not updating it here saves a write on every instruction
	}
}

// growRegisters ensures vm.registers has at least `top` slots.
// On reallocation, slice views held by existing frames are refreshed so callers
// can continue to use frame.registers without re-indexing through baseReg.
func (vm *RegisterVM) growRegisters(top int) {
	if top <= len(vm.registers) {
		return
	}
	n := len(vm.registers) * 2
	if n < top {
		n = top
	}
	newRegs := make([]Value, n)
	copy(newRegs, vm.registers)
	vm.registers = newRegs
	// Refresh existing frame views to point into the new backing array.
	for i := 0; i < vm.frameIndex; i++ {
		f := vm.frames[i]
		if f != nil {
			f.registers = vm.registers[f.baseReg:]
		}
	}
}

// callFunction handles function calls in the register VM.
// Frames share a single backing register file (Lua-style register window).
// Args already sit at the caller's argReg..argReg+NumParams-1 slots, which
// become the callee's R0..R(NumParams-1) — no copy required.
func (vm *RegisterVM) callFunction(fnReg, argReg, resultReg int) error {
	function := vm.currentFrame.registers[fnReg]

	var fn *Function
	switch function.Type {
	case FunctionType:
		fn = function.AsFunction()
	case ClosureType:
		fn = function.AsClosure().Fn
	default:
		return ErrCallingNonFunction
	}

	if len(fn.RegisterInstructions) == 0 {
		return fmt.Errorf("function %s has no register bytecode", fn.Name)
	}

	if vm.frameIndex >= MaxFrames {
		return fmt.Errorf("call stack overflow")
	}

	// New base lives at caller's argReg slot in the shared register array.
	newBase := vm.currentFrame.baseReg + argReg
	// Ensure backing array can fit the callee's register window.
	vm.growRegisters(newBase + fn.NumLocals)

	newFrame := vm.frames[vm.frameIndex]
	if newFrame == nil {
		newFrame = &RegisterFrame{}
		vm.frames[vm.frameIndex] = newFrame
	}
	newFrame.function = fn
	newFrame.instructions = fn.RegisterInstructions
	newFrame.pc = 0
	newFrame.baseReg = newBase
	newFrame.resultReg = resultReg
	newFrame.registers = vm.registers[newBase:]

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

	// Save the result register location from the callee frame
	calleeResultReg := vm.currentFrame.resultReg

	// Pop frame
	vm.frameIndex--
	vm.currentFrame = vm.frames[vm.frameIndex-1]

	// Store return value in caller's result register
	if resultReg >= 0 && calleeResultReg >= 0 {
		vm.currentFrame.registers[calleeResultReg] = returnValue
	}

	return nil
}

// callBuiltin handles builtin function calls
func (vm *RegisterVM) callBuiltin(index, argReg, resultReg, numArgs int) error {
	if index >= len(Builtins) {
		return fmt.Errorf("unknown builtin: %d", index)
	}

	builtin := Builtins[index]

	// Zero-copy: pass slice view directly (optimization - avoids allocation)
	// Args are guaranteed to be in consecutive registers argReg..argReg+numArgs-1
	endReg := argReg + numArgs
	if endReg > len(vm.currentFrame.registers) {
		endReg = len(vm.currentFrame.registers)
	}

	result := builtin(vm.currentFrame.registers[argReg:endReg]...)
	vm.currentFrame.registers[resultReg] = result

	return nil
}
