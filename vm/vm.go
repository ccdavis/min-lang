package vm

import (
	"errors"
	"fmt"
)

// Pre-allocated errors to avoid string allocation on error paths
var (
	ErrDivisionByZero        = errors.New("division by zero")
	ErrModuloByZero          = errors.New("modulo by zero")
	ErrStackOverflow         = errors.New("stack overflow")
	ErrUnsupportedOperands   = errors.New("unsupported operand types")
	ErrCallingNonFunction    = errors.New("calling non-function")
	ErrUnsupportedComparison = errors.New("unsupported operand types for comparison")
	ErrUnsupportedNegation   = errors.New("unsupported operand type for negation")
)

const (
	StackSize      = 2048
	GlobalsSize    = 65536
	MaxFrames      = 1024
)

// Frame represents a call frame
type Frame struct {
	cl          *Closure
	ip          int      // instruction pointer
	basePointer int      // base pointer for this frame
	tempClosure Closure  // Embedded closure for non-closure function calls (avoids allocation)
}

// NewFrame creates a new frame
func NewFrame(cl *Closure, basePointer int) *Frame {
	return &Frame{
		cl:          cl,
		ip:          0,
		basePointer: basePointer,
	}
}

// Instructions returns the instructions for this frame
func (f *Frame) Instructions() []byte {
	return f.cl.Fn.Instructions
}

// VM represents the virtual machine
type VM struct {
	constants []Value

	stack []Value
	sp    int // stack pointer (points to next free slot)

	globals []Value

	frames      []*Frame
	framesIndex int
}

// New creates a new VM
func New(bytecode *Bytecode) *VM {
	mainFn := &Function{
		Instructions: bytecode.Instructions,
		NumLocals:    0,
		NumParams:    0,
	}
	mainClosure := &Closure{Fn: mainFn, Free: nil}  // Use nil instead of empty slice
	mainFrame := NewFrame(mainClosure, 0)

	frames := make([]*Frame, MaxFrames)
	frames[0] = mainFrame

	return &VM{
		constants:   bytecode.Constants,
		stack:       make([]Value, StackSize),
		sp:          0,
		globals:     make([]Value, GlobalsSize),
		frames:      frames,
		framesIndex: 1,
	}
}

// Bytecode represents compiled bytecode
type Bytecode struct {
	Instructions []byte
	Constants    []Value
}

// currentFrame returns the current frame
func (vm *VM) currentFrame() *Frame {
	return vm.frames[vm.framesIndex-1]
}

// push pushes a value onto the stack
func (vm *VM) push(val Value) error {
	if vm.sp >= StackSize {
		return ErrStackOverflow
	}
	vm.stack[vm.sp] = val
	vm.sp++
	return nil
}

// pop pops a value from the stack
func (vm *VM) pop() Value {
	if vm.sp <= 0 {
		panic(fmt.Sprintf("stack underflow: sp=%d", vm.sp))
	}
	val := vm.stack[vm.sp-1]
	vm.sp--
	return val
}

// LastPoppedStackElem returns the last popped stack element
func (vm *VM) LastPoppedStackElem() Value {
	return vm.stack[vm.sp]
}

// Run executes the bytecode
func (vm *VM) Run() error {
	// Outer loop - manages frames
	for {
		if vm.framesIndex == 0 {
			break
		}

		// Cache frame and instructions to avoid repeated lookups
		frame := vm.frames[vm.framesIndex-1]
		ins := frame.Instructions()
		ip := frame.ip

		// fmt.Printf("DEBUG: Starting frame, ip=%d, insLen=%d\n", ip, len(ins))

	innerLoop:
		// Inner loop - executes instructions until frame change
		for ip < len(ins) {
			op := OpCode(ins[ip])
			ip++

			switch op {
			case OpPush:
				constIndex, _ := ReadOperand(ins, ip)
				ip += 2

				err := vm.push(vm.constants[constIndex])
				if err != nil {
					return err
				}

			case OpPop:
				vm.pop()

			case OpDup:
				if vm.sp <= 0 {
					return fmt.Errorf("stack underflow on OpDup: sp=%d", vm.sp)
				}
				val := vm.stack[vm.sp-1]
				err := vm.push(val)
				if err != nil {
					return err
				}

			case OpAdd, OpSub, OpMul, OpDiv, OpMod:
				err := vm.executeBinaryOperation(op)
				if err != nil {
					return err
				}

			case OpAddLocal, OpSubLocal, OpMulLocal, OpDivLocal:
				localIndex, _ := ReadOperand(ins, ip)
				ip += 2

				// Get TOS and local value
				tos := vm.pop()
				local := vm.stack[frame.basePointer+localIndex]

				// Perform operation directly without type checking overhead
				// Handle integer operations (fast path)
				if tos.Type == IntType && local.Type == IntType {
					var result int64
					switch op {
					case OpAddLocal:
						result = tos.AsInt() + local.AsInt()
					case OpSubLocal:
						result = tos.AsInt() - local.AsInt()
					case OpMulLocal:
						result = tos.AsInt() * local.AsInt()
					case OpDivLocal:
						if local.AsInt() == 0 {
							return ErrDivisionByZero
						}
						result = tos.AsInt() / local.AsInt()
					}
					err := vm.push(IntValue(result))
					if err != nil {
						return err
					}
				} else if (tos.Type == FloatType || tos.Type == IntType) &&
					(local.Type == FloatType || local.Type == IntType) {
					// Handle float operations
					var tosVal, localVal float64

					if tos.Type == FloatType {
						tosVal = tos.AsFloat()
					} else {
						tosVal = float64(tos.AsInt())
					}

					if local.Type == FloatType {
						localVal = local.AsFloat()
					} else {
						localVal = float64(local.AsInt())
					}

					var result float64
					switch op {
					case OpAddLocal:
						result = tosVal + localVal
					case OpSubLocal:
						result = tosVal - localVal
					case OpMulLocal:
						result = tosVal * localVal
					case OpDivLocal:
						if localVal == 0 {
							return ErrDivisionByZero
						}
						result = tosVal / localVal
					}
					err := vm.push(FloatValue(result))
					if err != nil {
						return err
					}
				} else {
					return fmt.Errorf("unsupported operand types for local operation")
				}

			case OpNeg:
				operand := vm.pop()
				switch operand.Type {
				case IntType:
					err := vm.push(IntValue(-operand.AsInt()))
					if err != nil {
						return err
					}
				case FloatType:
					err := vm.push(FloatValue(-operand.AsFloat()))
					if err != nil {
						return err
					}
				default:
					return fmt.Errorf("unsupported operand type for negation: %d", operand.Type)
				}

			case OpEq, OpNe, OpLt, OpGt, OpLe, OpGe:
				err := vm.executeComparison(op)
				if err != nil {
					return err
				}

			case OpAnd, OpOr:
				err := vm.executeLogicalOperation(op)
				if err != nil {
					return err
				}

			case OpNot:
				operand := vm.pop()
				err := vm.push(BoolValue(!operand.IsTruthy()))
				if err != nil {
					return err
				}

			case OpLoadGlobal:
				globalIndex, _ := ReadOperand(ins, ip)
				ip += 2

				value := vm.globals[globalIndex]
				// DEBUG
				// fmt.Printf("DEBUG: LoadGlobal[%d] = %v\n", globalIndex, value)
				err := vm.push(value)
				if err != nil {
					return err
				}

			case OpStoreGlobal:
				globalIndex, _ := ReadOperand(ins, ip)
				ip += 2

				value := vm.pop()
				vm.globals[globalIndex] = value
				// DEBUG
				// fmt.Printf("DEBUG: StoreGlobal[%d] = %v\n", globalIndex, value)

			case OpLoadLocal:
				localIndex, _ := ReadOperand(ins, ip)
				ip += 2

				err := vm.push(vm.stack[frame.basePointer+localIndex])
				if err != nil {
					return err
				}

			case OpStoreLocal:
				localIndex, _ := ReadOperand(ins, ip)
				ip += 2

				vm.stack[frame.basePointer+localIndex] = vm.pop()

			case OpJump:
				pos, _ := ReadOperand(ins, ip)
				ip = pos
				frame.ip = ip
				break innerLoop // Break inner loop to reload frame

			case OpJumpIfFalse:
				pos, _ := ReadOperand(ins, ip)
				ip += 2

				condition := vm.pop()
				if !condition.IsTruthy() {
					ip = pos
					frame.ip = ip
					break innerLoop // Break inner loop to reload frame
				}

			case OpJumpIfTrue:
				pos, _ := ReadOperand(ins, ip)
				ip += 2

				condition := vm.pop()
				if condition.IsTruthy() {
					ip = pos
					frame.ip = ip
					break innerLoop // Break inner loop to reload frame
				}

			case OpCall:
				numArgs, _ := ReadOperand(ins, ip)
				ip += 2

				// fmt.Printf("DEBUG: OpCall with %d args\n", numArgs)
				frame.ip = ip // Sync before call
				err := vm.executeCall(numArgs)
				if err != nil {
					return err
				}
				// fmt.Printf("DEBUG: OpCall completed, breaking to reload frame\n")
				break innerLoop // Break to reload new frame

			case OpReturn:
				returnValue := vm.pop()
				// fmt.Printf("DEBUG: OpReturn with value %v\n", returnValue)

				// Set sp to where the function was (one before the first argument)
				// This removes the function and all arguments from the stack
				vm.sp = frame.basePointer - 1

				vm.framesIndex--

				err := vm.push(returnValue)
				if err != nil {
					return err
				}
				// fmt.Printf("DEBUG: OpReturn pushed value, returning to previous frame\n")
				break innerLoop // Break to reload previous frame

			case OpMakeClosure:
				fnIndex, _ := ReadOperand(ins, ip)
				numFree, _ := ReadOperand(ins, ip+2)
				ip += 4

				fn := vm.constants[fnIndex].AsFunction()

				free := make([]Value, numFree)
				for i := numFree - 1; i >= 0; i-- {
					free[i] = vm.pop()
				}

				closure := NewClosureValue(fn, free)
				err := vm.push(closure)
				if err != nil {
					return err
				}

			case OpLoadFree:
				freeIndex, _ := ReadOperand(ins, ip)
				ip += 2

				currentClosure := frame.cl
				err := vm.push(currentClosure.Free[freeIndex])
				if err != nil {
					return err
				}

			case OpGetBuiltin:
				builtinIndex, _ := ReadOperand(ins, ip)
				ip += 2

				builtin := vm.getBuiltin(builtinIndex)
				err := vm.push(builtin)
				if err != nil {
					return err
				}

			case OpArray:
				size, _ := ReadOperand(ins, ip)
				ip += 2

				array := NewArrayValue(size)
				arrayVal := array.AsArray()

				// Pop elements from stack in reverse order
				for i := size - 1; i >= 0; i-- {
					arrayVal.Elements[i] = vm.pop()
				}

				err := vm.push(array)
				if err != nil {
					return err
				}

			case OpArrayGet:
				index := vm.pop()
				container := vm.pop()

				switch container.Type {
				case ArrayType:
					if index.Type != IntType {
						return fmt.Errorf("array index must be integer, got %d", index.Type)
					}

					idx := int(index.AsInt())
					arrayVal := container.AsArray()

					if idx < 0 || idx >= len(arrayVal.Elements) {
						return fmt.Errorf("array index out of bounds: %d", idx)
					}

					err := vm.push(arrayVal.Elements[idx])
					if err != nil {
						return err
					}

				case MapType:
					mapKey := index.ToMapKey()
					mapData := container.AsMap()

					val, ok := mapData.Pairs[mapKey]
					if !ok {
						err := vm.push(NilValue())
						if err != nil {
							return err
						}
					} else {
						err := vm.push(val)
						if err != nil {
							return err
						}
					}

				case StringType:
					if index.Type != IntType {
						return fmt.Errorf("string index must be integer, got %d", index.Type)
					}

					idx := int(index.AsInt())
					str := container.AsString()

					if idx < 0 || idx >= len(str) {
						return fmt.Errorf("string index out of bounds: %d", idx)
					}

					// Return a single-character string
					err := vm.push(StringValue(string(str[idx])))
					if err != nil {
						return err
					}

				default:
					return fmt.Errorf("index operator not supported for type %d", container.Type)
				}

			case OpArraySet:
				value := vm.pop()
				index := vm.pop()
				container := vm.pop()

				switch container.Type {
				case ArrayType:
					if index.Type != IntType {
						return fmt.Errorf("array index must be integer, got %d", index.Type)
					}

					idx := int(index.AsInt())
					arrayVal := container.AsArray()

					if idx < 0 || idx >= len(arrayVal.Elements) {
						return fmt.Errorf("array index out of bounds: %d", idx)
					}

					arrayVal.Elements[idx] = value

				case MapType:
					mapKey := index.ToMapKey()
					mapData := container.AsMap()
					mapData.Pairs[mapKey] = value

				default:
					return fmt.Errorf("index assignment not supported for type %d", container.Type)
				}

			case OpMap:
				size, _ := ReadOperand(ins, ip)
				ip += 2

				mapVal := NewMapValue()
				mapData := mapVal.AsMap()

				// Pop key-value pairs from stack
				for i := 0; i < size; i++ {
					value := vm.pop()
					key := vm.pop()

					// Use optimized map key (no allocation for ints)
					mapKey := key.ToMapKey()
					mapData.Pairs[mapKey] = value
				}

				err := vm.push(mapVal)
				if err != nil {
					return err
				}

			case OpMapGet:
				key := vm.pop()
				mapVal := vm.pop()

				if mapVal.Type != MapType {
					return fmt.Errorf("index operator not supported for type %d", mapVal.Type)
				}

				mapKey := key.ToMapKey()
				mapData := mapVal.AsMap()

				val, ok := mapData.Pairs[mapKey]
				if !ok {
					err := vm.push(NilValue())
					if err != nil {
						return err
					}
				} else {
					err := vm.push(val)
					if err != nil {
						return err
					}
				}

			case OpMapSet:
				value := vm.pop()
				key := vm.pop()
				mapVal := vm.pop()

				if mapVal.Type != MapType {
					return fmt.Errorf("index operator not supported for type %d", mapVal.Type)
				}

				mapKey := key.ToMapKey()
				mapData := mapVal.AsMap()
				mapData.Pairs[mapKey] = value

			case OpStruct:
				numFields, _ := ReadOperand(ins, ip)
				ip += 2

				typeNameVal := vm.pop()
				if typeNameVal.Type != StringType {
					return fmt.Errorf("struct type name must be string")
				}

				fields := make(map[string]Value)
				for i := 0; i < numFields; i++ {
					value := vm.pop()
					fieldName := vm.pop()

					if fieldName.Type != StringType {
						return fmt.Errorf("struct field name must be string")
					}

					fields[fieldName.AsString()] = value
				}

				structVal := NewStructValue(typeNameVal.AsString(), fields)
				err := vm.push(structVal)
				if err != nil {
					return err
				}

			case OpGetField:
				fieldNameVal := vm.pop()
				structVal := vm.pop()

				if structVal.Type != StructType {
					return fmt.Errorf("field access not supported for type %d", structVal.Type)
				}

				if fieldNameVal.Type != StringType {
					return fmt.Errorf("field name must be string")
				}

				structData := structVal.AsStruct()
				fieldName := fieldNameVal.AsString()

				val, ok := structData.Fields[fieldName]
				if !ok {
					return fmt.Errorf("field %s not found in struct %s", fieldName, structData.TypeName)
				}

				err := vm.push(val)
				if err != nil {
					return err
				}

			case OpSetField:
				value := vm.pop()
				fieldNameVal := vm.pop()
				structVal := vm.pop()

				if structVal.Type != StructType {
					return fmt.Errorf("field access not supported for type %d", structVal.Type)
				}

				if fieldNameVal.Type != StringType {
					return fmt.Errorf("field name must be string")
				}

				structData := structVal.AsStruct()
				fieldName := fieldNameVal.AsString()

				structData.Fields[fieldName] = value

			case OpPrint:
				val := vm.pop()
				fmt.Println(val.String())

			case OpHalt:
				return nil
			}
		}

		// Sync IP after inner loop completes normally
		// NOTE: If we broke out of innerLoop (e.g. from OpReturn or OpCall),
		// this syncs the IP for the frame we just left, which is fine.
		frame.ip = ip

		// Reload frame variables after potential frame change
		// (OpCall, OpReturn, OpJump can all change frames or IP)
		frame = vm.frames[vm.framesIndex-1]
		ins = frame.Instructions()
		ip = frame.ip

		// If we're in the main frame and completed all instructions, exit
		if vm.framesIndex == 1 && ip >= len(ins) {
			return nil
		}
	}

	return nil
}

// executeBinaryOperation executes a binary operation
func (vm *VM) executeBinaryOperation(op OpCode) error {
	right := vm.pop()
	left := vm.pop()

	// Handle string concatenation
	if op == OpAdd && (left.Type == StringType || right.Type == StringType) {
		leftStr := left.String()
		rightStr := right.String()
		return vm.push(StringValue(leftStr + rightStr))
	}

	// Handle integer operations
	if left.Type == IntType && right.Type == IntType {
		return vm.executeBinaryIntegerOperation(op, left.AsInt(), right.AsInt())
	}

	// Handle float operations
	if (left.Type == FloatType || left.Type == IntType) &&
		(right.Type == FloatType || right.Type == IntType) {
		var leftVal, rightVal float64

		if left.Type == FloatType {
			leftVal = left.AsFloat()
		} else {
			leftVal = float64(left.AsInt())
		}

		if right.Type == FloatType {
			rightVal = right.AsFloat()
		} else {
			rightVal = float64(right.AsInt())
		}

		return vm.executeBinaryFloatOperation(op, leftVal, rightVal)
	}

	return ErrUnsupportedOperands
}

// executeBinaryIntegerOperation executes a binary integer operation
func (vm *VM) executeBinaryIntegerOperation(op OpCode, left, right int64) error {
	var result int64

	switch op {
	case OpAdd:
		result = left + right
	case OpSub:
		result = left - right
	case OpMul:
		result = left * right
	case OpDiv:
		if right == 0 {
			return ErrDivisionByZero
		}
		result = left / right
	case OpMod:
		if right == 0 {
			return ErrModuloByZero
		}
		result = left % right
	default:
		return fmt.Errorf("unknown integer operator: %d", op)
	}

	return vm.push(IntValue(result))
}

// executeBinaryFloatOperation executes a binary float operation
func (vm *VM) executeBinaryFloatOperation(op OpCode, left, right float64) error {
	var result float64

	switch op {
	case OpAdd:
		result = left + right
	case OpSub:
		result = left - right
	case OpMul:
		result = left * right
	case OpDiv:
		if right == 0 {
			return ErrDivisionByZero
		}
		result = left / right
	default:
		return fmt.Errorf("unknown float operator: %d", op)
	}

	return vm.push(FloatValue(result))
}

// executeComparison executes a comparison operation
func (vm *VM) executeComparison(op OpCode) error {
	right := vm.pop()
	left := vm.pop()

	// Handle integer comparisons
	if left.Type == IntType && right.Type == IntType {
		return vm.executeIntegerComparison(op, left.AsInt(), right.AsInt())
	}

	// Handle float comparisons
	if (left.Type == FloatType || left.Type == IntType) &&
		(right.Type == FloatType || right.Type == IntType) {
		var leftVal, rightVal float64

		if left.Type == FloatType {
			leftVal = left.AsFloat()
		} else {
			leftVal = float64(left.AsInt())
		}

		if right.Type == FloatType {
			rightVal = right.AsFloat()
		} else {
			rightVal = float64(right.AsInt())
		}

		return vm.executeFloatComparison(op, leftVal, rightVal)
	}

	// Handle boolean comparisons
	if left.Type == BoolType && right.Type == BoolType {
		switch op {
		case OpEq:
			return vm.push(BoolValue(left.AsBool() == right.AsBool()))
		case OpNe:
			return vm.push(BoolValue(left.AsBool() != right.AsBool()))
		default:
			return fmt.Errorf("unknown boolean comparison operator: %d", op)
		}
	}

	return ErrUnsupportedComparison
}

// executeIntegerComparison executes an integer comparison
func (vm *VM) executeIntegerComparison(op OpCode, left, right int64) error {
	var result bool

	switch op {
	case OpEq:
		result = left == right
	case OpNe:
		result = left != right
	case OpLt:
		result = left < right
	case OpGt:
		result = left > right
	case OpLe:
		result = left <= right
	case OpGe:
		result = left >= right
	default:
		return fmt.Errorf("unknown integer comparison operator: %d", op)
	}

	return vm.push(BoolValue(result))
}

// executeFloatComparison executes a float comparison
func (vm *VM) executeFloatComparison(op OpCode, left, right float64) error {
	var result bool

	switch op {
	case OpEq:
		result = left == right
	case OpNe:
		result = left != right
	case OpLt:
		result = left < right
	case OpGt:
		result = left > right
	case OpLe:
		result = left <= right
	case OpGe:
		result = left >= right
	default:
		return fmt.Errorf("unknown float comparison operator: %d", op)
	}

	return vm.push(BoolValue(result))
}

// executeLogicalOperation executes a logical operation
func (vm *VM) executeLogicalOperation(op OpCode) error {
	right := vm.pop()
	left := vm.pop()

	switch op {
	case OpAnd:
		return vm.push(BoolValue(left.IsTruthy() && right.IsTruthy()))
	case OpOr:
		return vm.push(BoolValue(left.IsTruthy() || right.IsTruthy()))
	default:
		return fmt.Errorf("unknown logical operator: %d", op)
	}
}

// executeCall executes a function call
func (vm *VM) executeCall(numArgs int) error {
	callee := vm.stack[vm.sp-1-numArgs]

	switch callee.Type {
	case ClosureType:
		return vm.callClosure(callee.AsClosure(), numArgs)
	case FunctionType:
		return vm.callFunction(callee.AsFunction(), numArgs)
	case BuiltinFunctionType:
		return vm.executeBuiltin(callee.AsBuiltinFunction(), numArgs)
	default:
		return ErrCallingNonFunction
	}
}

// callClosure calls a closure
func (vm *VM) callClosure(cl *Closure, numArgs int) error {
	if numArgs != cl.Fn.NumParams {
		return fmt.Errorf("wrong number of arguments: want=%d, got=%d",
			cl.Fn.NumParams, numArgs)
	}

	// basePointer points to the first argument
	// Stack layout: [... function arg1 arg2 ...]
	// We want basePointer to point to arg1
	basePointer := vm.sp - numArgs

	// Reuse existing frame if available, otherwise allocate new
	frame := vm.frames[vm.framesIndex]
	if frame == nil {
		frame = &Frame{}
		vm.frames[vm.framesIndex] = frame
	}

	// Reset frame fields
	frame.cl = cl
	frame.ip = 0
	frame.basePointer = basePointer

	vm.framesIndex++
	vm.sp = basePointer + cl.Fn.NumLocals

	return nil
}

// callFunction calls a function
func (vm *VM) callFunction(fn *Function, numArgs int) error {
	if numArgs != fn.NumParams {
		return fmt.Errorf("wrong number of arguments: want=%d, got=%d",
			fn.NumParams, numArgs)
	}

	basePointer := vm.sp - numArgs

	// Reuse existing frame if available, otherwise allocate new
	frame := vm.frames[vm.framesIndex]
	if frame == nil {
		frame = &Frame{}
		vm.frames[vm.framesIndex] = frame
	}

	// Use embedded closure to avoid heap allocation
	frame.tempClosure.Fn = fn
	frame.tempClosure.Free = nil  // No free variables for regular functions
	frame.cl = &frame.tempClosure

	// Reset frame fields
	frame.ip = 0
	frame.basePointer = basePointer

	vm.framesIndex++
	vm.sp = basePointer + fn.NumLocals

	return nil
}
