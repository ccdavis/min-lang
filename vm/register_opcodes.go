package vm

// RegisterOpCode represents a register VM instruction
type RegisterOpCode byte

const (
	// Load/Move operations
	OpRLoadK RegisterOpCode = iota // R(A) = K(Bx) - Load constant
	OpRMove                         // R(A) = R(B) - Copy register

	// Arithmetic operations (type-specialized, no checks)
	OpRAddInt   // R(A) = R(B) + R(C) - int
	OpRAddFloat // R(A) = R(B) + R(C) - float
	OpRSubInt   // R(A) = R(B) - R(C) - int
	OpRSubFloat // R(A) = R(B) - R(C) - float
	OpRMulInt   // R(A) = R(B) * R(C) - int
	OpRMulFloat // R(A) = R(B) * R(C) - float
	OpRDivInt   // R(A) = R(B) / R(C) - int
	OpRDivFloat // R(A) = R(B) / R(C) - float
	OpRModInt   // R(A) = R(B) % R(C) - int
	OpRNegInt   // R(A) = -R(B) - int
	OpRNegFloat // R(A) = -R(B) - float

	// Comparison operations (type-specialized, no checks)
	OpREqInt    // R(A) = R(B) == R(C) - int
	OpREqFloat  // R(A) = R(B) == R(C) - float
	OpREqBool   // R(A) = R(B) == R(C) - bool
	OpREqString // R(A) = R(B) == R(C) - string
	OpRNeInt    // R(A) = R(B) != R(C) - int
	OpRNeFloat  // R(A) = R(B) != R(C) - float
	OpRNeBool   // R(A) = R(B) != R(C) - bool
	OpRNeString // R(A) = R(B) != R(C) - string
	OpRLtInt    // R(A) = R(B) < R(C) - int
	OpRLtFloat  // R(A) = R(B) < R(C) - float
	OpRGtInt    // R(A) = R(B) > R(C) - int
	OpRGtFloat  // R(A) = R(B) > R(C) - float
	OpRLeInt    // R(A) = R(B) <= R(C) - int
	OpRLeFloat  // R(A) = R(B) <= R(C) - float
	OpRGeInt    // R(A) = R(B) >= R(C) - int
	OpRGeFloat  // R(A) = R(B) >= R(C) - float

	// Logical operations
	OpRAnd // R(A) = R(B) && R(C)
	OpROr  // R(A) = R(B) || R(C)
	OpRNot // R(A) = !R(B)

	// Control flow
	OpRJump    // PC = offset
	OpRJumpT   // if R(A) then PC = offset
	OpRJumpF   // if !R(A) then PC = offset
	OpRReturn  // return R(A)...R(A+n)
	OpRReturnN // return (no value)

	// Function calls
	OpRCall    // R(A) = call R(B)(R(C)...R(C+n))
	OpRBuiltin // R(A) = builtin[Bx](R(C)...R(C+n))

	// Array operations
	OpRNewArray // R(A) = new array[Bx]
	OpRGetIdx   // R(A) = R(B)[R(C)]
	OpRSetIdx   // R(A)[R(B)] = R(C)

	// Map operations
	OpRNewMap // R(A) = new map
	OpRMapGet // R(A) = R(B)[R(C)]
	OpRMapSet // R(A)[R(B)] = R(C)

	// Struct operations
	OpRNewStruct   // R(A) = new struct
	OpRGetField    // R(A) = R(B).field(C) - by offset
	OpRSetField    // R(A).field(B) = R(C) - by offset
	OpRLoadGlobal  // R(A) = global[Bx]
	OpRStoreGlobal // global[Bx] = R(A)

	// String operations
	OpRConcat // R(A) = R(B) + R(C) - string concatenation

	// Optimized operations (immediate constants)
	OpRAddConstInt   // R(A) = R(B) + K(C) - int
	OpRAddConstFloat // R(A) = R(B) + K(C) - float
	OpRMulConstInt   // R(A) = R(B) * K(C) - int
	OpRMulConstFloat // R(A) = R(B) * K(C) - float

	// Special optimizations
	OpRSquareInt   // R(A) = R(B) * R(B) - int (for Mandelbrot)
	OpRSquareFloat // R(A) = R(B) * R(B) - float (for Mandelbrot)

	OpRHalt // Halt execution
)

// RegisterInstruction represents a 32-bit register instruction
type RegisterInstruction uint32

// Encode creates a register instruction
func EncodeRegisterInstruction(op RegisterOpCode, a, b, c uint8) RegisterInstruction {
	return RegisterInstruction(uint32(op)<<24 | uint32(a)<<16 | uint32(b)<<8 | uint32(c))
}

// EncodeRegisterInstructionBx creates a register instruction with large immediate
func EncodeRegisterInstructionBx(op RegisterOpCode, a uint8, bx uint16) RegisterInstruction {
	return RegisterInstruction(uint32(op)<<24 | uint32(a)<<16 | uint32(bx))
}

// Decode extracts fields from a register instruction
func (ins RegisterInstruction) Decode() (op RegisterOpCode, a, b, c uint8) {
	op = RegisterOpCode(ins >> 24)
	a = uint8((ins >> 16) & 0xFF)
	b = uint8((ins >> 8) & 0xFF)
	c = uint8(ins & 0xFF)
	return
}

// DecodeBx extracts fields from a register instruction with large immediate
func (ins RegisterInstruction) DecodeBx() (op RegisterOpCode, a uint8, bx uint16) {
	op = RegisterOpCode(ins >> 24)
	a = uint8((ins >> 16) & 0xFF)
	bx = uint16(ins & 0xFFFF)
	return
}

// String returns the string representation of a register opcode
func (op RegisterOpCode) String() string {
	switch op {
	case OpRLoadK:
		return "LOADK"
	case OpRMove:
		return "MOVE"
	case OpRAddInt:
		return "ADD_INT"
	case OpRAddFloat:
		return "ADD_FLOAT"
	case OpRSubInt:
		return "SUB_INT"
	case OpRSubFloat:
		return "SUB_FLOAT"
	case OpRMulInt:
		return "MUL_INT"
	case OpRMulFloat:
		return "MUL_FLOAT"
	case OpRDivInt:
		return "DIV_INT"
	case OpRDivFloat:
		return "DIV_FLOAT"
	case OpRModInt:
		return "MOD_INT"
	case OpRNegInt:
		return "NEG_INT"
	case OpRNegFloat:
		return "NEG_FLOAT"
	case OpREqInt:
		return "EQ_INT"
	case OpREqFloat:
		return "EQ_FLOAT"
	case OpREqBool:
		return "EQ_BOOL"
	case OpREqString:
		return "EQ_STRING"
	case OpRNeInt:
		return "NE_INT"
	case OpRNeFloat:
		return "NE_FLOAT"
	case OpRNeBool:
		return "NE_BOOL"
	case OpRNeString:
		return "NE_STRING"
	case OpRLtInt:
		return "LT_INT"
	case OpRLtFloat:
		return "LT_FLOAT"
	case OpRGtInt:
		return "GT_INT"
	case OpRGtFloat:
		return "GT_FLOAT"
	case OpRLeInt:
		return "LE_INT"
	case OpRLeFloat:
		return "LE_FLOAT"
	case OpRGeInt:
		return "GE_INT"
	case OpRGeFloat:
		return "GE_FLOAT"
	case OpRAnd:
		return "AND"
	case OpROr:
		return "OR"
	case OpRNot:
		return "NOT"
	case OpRJump:
		return "JUMP"
	case OpRJumpT:
		return "JUMPT"
	case OpRJumpF:
		return "JUMPF"
	case OpRReturn:
		return "RETURN"
	case OpRReturnN:
		return "RETURNN"
	case OpRCall:
		return "CALL"
	case OpRBuiltin:
		return "BUILTIN"
	case OpRNewArray:
		return "NEWARRAY"
	case OpRGetIdx:
		return "GETIDX"
	case OpRSetIdx:
		return "SETIDX"
	case OpRNewMap:
		return "NEWMAP"
	case OpRMapGet:
		return "MAPGET"
	case OpRMapSet:
		return "MAPSET"
	case OpRNewStruct:
		return "NEWSTRUCT"
	case OpRGetField:
		return "GETFIELD"
	case OpRSetField:
		return "SETFIELD"
	case OpRLoadGlobal:
		return "LOADGLOBAL"
	case OpRStoreGlobal:
		return "STOREGLOBAL"
	case OpRConcat:
		return "CONCAT"
	case OpRAddConstInt:
		return "ADDCONST_INT"
	case OpRAddConstFloat:
		return "ADDCONST_FLOAT"
	case OpRMulConstInt:
		return "MULCONST_INT"
	case OpRMulConstFloat:
		return "MULCONST_FLOAT"
	case OpRSquareInt:
		return "SQUARE_INT"
	case OpRSquareFloat:
		return "SQUARE_FLOAT"
	case OpRHalt:
		return "HALT"
	default:
		return "UNKNOWN"
	}
}
