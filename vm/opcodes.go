package vm

// OpCode represents a VM instruction
type OpCode byte

const (
	// Stack operations
	OpPush    OpCode = iota // Push constant onto stack
	OpPop                   // Pop value from stack
	OpDup                   // Duplicate top of stack
	OpSwap                  // Swap top two stack values

	// Arithmetic operations (generic - with runtime type checking)
	OpAdd // Add top two stack values (generic)
	OpSub // Subtract top two stack values (generic)
	OpMul // Multiply top two stack values (generic)
	OpDiv // Divide top two stack values (generic)
	OpMod // Modulo operation (generic)
	OpNeg // Negate top of stack (generic)

	// Type-specialized arithmetic operations (Phase 1 optimization - no runtime checks!)
	OpAddInt    // int + int → int (no type checking)
	OpAddFloat  // float + float → float (no type checking)
	OpAddString // string + string → string (concatenation, no type checking)
	OpSubInt    // int - int → int (no type checking)
	OpSubFloat  // float - float → float (no type checking)
	OpMulInt    // int * int → int (no type checking)
	OpMulFloat  // float * float → float (no type checking)
	OpDivInt    // int / int → int (no type checking)
	OpDivFloat  // float / float → float (no type checking)
	OpModInt    // int % int → int (no type checking)

	// Direct local operations (no push/pop overhead)
	OpAddLocal // Add TOS with local variable, push result
	OpSubLocal // Subtract local from TOS, push result
	OpMulLocal // Multiply TOS with local variable, push result
	OpDivLocal // Divide TOS by local variable, push result

	// Comparison operations
	OpEq // Equal
	OpNe // Not equal
	OpLt // Less than
	OpGt // Greater than
	OpLe // Less than or equal
	OpGe // Greater than or equal

	// Logical operations
	OpAnd // Logical AND
	OpOr  // Logical OR
	OpNot // Logical NOT

	// Variable operations
	OpLoadGlobal  // Load global variable onto stack
	OpStoreGlobal // Store top of stack to global variable
	OpLoadLocal   // Load local variable onto stack
	OpStoreLocal  // Store top of stack to local variable
	OpLoadFree    // Load free variable (closure) onto stack

	// Control flow
	OpJump      // Unconditional jump
	OpJumpIfFalse // Jump if top of stack is false
	OpJumpIfTrue  // Jump if top of stack is true

	// Function operations
	OpCall         // Call function
	OpReturn       // Return from function
	OpMakeClosure  // Create closure
	OpGetBuiltin   // Get built-in function

	// Array operations
	OpArray      // Create array
	OpArrayGet   // Get array element
	OpArraySet   // Set array element
	OpArrayLen   // Get array length

	// Map operations
	OpMap       // Create map
	OpMapGet    // Get map value
	OpMapSet    // Set map value

	// Struct operations
	OpStruct     // Create struct
	OpGetField   // Get struct field
	OpSetField   // Set struct field

	// Special operations
	OpHalt       // Halt execution
	OpPrint      // Built-in print (for debugging)
)

// String returns the string representation of an opcode
func (op OpCode) String() string {
	switch op {
	case OpPush:
		return "PUSH"
	case OpPop:
		return "POP"
	case OpDup:
		return "DUP"
	case OpSwap:
		return "SWAP"
	case OpAdd:
		return "ADD"
	case OpSub:
		return "SUB"
	case OpMul:
		return "MUL"
	case OpDiv:
		return "DIV"
	case OpMod:
		return "MOD"
	case OpNeg:
		return "NEG"
	case OpAddInt:
		return "ADD_INT"
	case OpAddFloat:
		return "ADD_FLOAT"
	case OpAddString:
		return "ADD_STRING"
	case OpSubInt:
		return "SUB_INT"
	case OpSubFloat:
		return "SUB_FLOAT"
	case OpMulInt:
		return "MUL_INT"
	case OpMulFloat:
		return "MUL_FLOAT"
	case OpDivInt:
		return "DIV_INT"
	case OpDivFloat:
		return "DIV_FLOAT"
	case OpModInt:
		return "MOD_INT"
	case OpAddLocal:
		return "ADD_LOCAL"
	case OpSubLocal:
		return "SUB_LOCAL"
	case OpMulLocal:
		return "MUL_LOCAL"
	case OpDivLocal:
		return "DIV_LOCAL"
	case OpEq:
		return "EQ"
	case OpNe:
		return "NE"
	case OpLt:
		return "LT"
	case OpGt:
		return "GT"
	case OpLe:
		return "LE"
	case OpGe:
		return "GE"
	case OpAnd:
		return "AND"
	case OpOr:
		return "OR"
	case OpNot:
		return "NOT"
	case OpLoadGlobal:
		return "LOAD_GLOBAL"
	case OpStoreGlobal:
		return "STORE_GLOBAL"
	case OpLoadLocal:
		return "LOAD_LOCAL"
	case OpStoreLocal:
		return "STORE_LOCAL"
	case OpLoadFree:
		return "LOAD_FREE"
	case OpJump:
		return "JUMP"
	case OpJumpIfFalse:
		return "JUMP_IF_FALSE"
	case OpJumpIfTrue:
		return "JUMP_IF_TRUE"
	case OpCall:
		return "CALL"
	case OpReturn:
		return "RETURN"
	case OpMakeClosure:
		return "MAKE_CLOSURE"
	case OpGetBuiltin:
		return "GET_BUILTIN"
	case OpArray:
		return "ARRAY"
	case OpArrayGet:
		return "ARRAY_GET"
	case OpArraySet:
		return "ARRAY_SET"
	case OpArrayLen:
		return "ARRAY_LEN"
	case OpMap:
		return "MAP"
	case OpMapGet:
		return "MAP_GET"
	case OpMapSet:
		return "MAP_SET"
	case OpStruct:
		return "STRUCT"
	case OpGetField:
		return "GET_FIELD"
	case OpSetField:
		return "SET_FIELD"
	case OpHalt:
		return "HALT"
	case OpPrint:
		return "PRINT"
	default:
		return "UNKNOWN"
	}
}
