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

	// Comparison operations (generic - with runtime type checking)
	OpEq // Equal
	OpNe // Not equal
	OpLt // Less than
	OpGt // Greater than
	OpLe // Less than or equal
	OpGe // Greater than or equal

	// Type-specialized comparison operations (Phase 2 optimization - no runtime checks!)
	OpEqInt    // int == int → bool (no type checking)
	OpEqFloat  // float == float → bool (no type checking)
	OpEqString // string == string → bool (no type checking)
	OpEqBool   // bool == bool → bool (no type checking)
	OpNeInt    // int != int → bool (no type checking)
	OpNeFloat  // float != float → bool (no type checking)
	OpNeString // string != string → bool (no type checking)
	OpNeBool   // bool != bool → bool (no type checking)
	OpLtInt    // int < int → bool (no type checking)
	OpLtFloat  // float < float → bool (no type checking)
	OpGtInt    // int > int → bool (no type checking)
	OpGtFloat  // float > float → bool (no type checking)
	OpLeInt    // int <= int → bool (no type checking)
	OpLeFloat  // float <= float → bool (no type checking)
	OpGeInt    // int >= int → bool (no type checking)
	OpGeFloat  // float >= float → bool (no type checking)

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
	OpStruct     // Create struct (with name-based fields)
	OpGetField   // Get struct field (by name)
	OpSetField   // Set struct field (by name)

	// Struct operations (Phase 3 optimization - offset-based access)
	OpStructOrdered  // Create struct with ordered fields (faster creation)
	OpGetFieldOffset // Get struct field by offset (no map lookup!)
	OpSetFieldOffset // Set struct field by offset (no map lookup!)

	// Phase 4A: Immediate constant arithmetic operations (variable OP constant)
	OpAddConstInt   // Add constant to TOS int (TOS + immediate → TOS)
	OpSubConstInt   // Subtract constant from TOS int (TOS - immediate → TOS)
	OpMulConstInt   // Multiply TOS int by constant (TOS * immediate → TOS)
	OpDivConstInt   // Divide TOS int by constant (TOS / immediate → TOS)
	OpModConstInt   // Modulo TOS int by constant (TOS % immediate → TOS)
	OpAddConstFloat // Add constant to TOS float (TOS + immediate → TOS)
	OpSubConstFloat // Subtract constant from TOS float (TOS - immediate → TOS)
	OpMulConstFloat // Multiply TOS float by constant (TOS * immediate → TOS)
	OpDivConstFloat // Divide TOS float by constant (TOS / immediate → TOS)

	// Phase 4B: Increment/decrement operations (special case of 4A)
	OpIncGlobal // Increment global variable by immediate value
	OpDecGlobal // Decrement global variable by immediate value
	OpIncLocal  // Increment local variable by immediate value
	OpDecLocal  // Decrement local variable by immediate value

	// Phase 4C: Square operations (x * x)
	OpSquareInt   // Square TOS int (TOS * TOS → TOS)
	OpSquareFloat // Square TOS float (TOS * TOS → TOS)

	// Phase 4D: Compare with immediate constant
	OpLtConstInt    // TOS < immediate (int)
	OpGtConstInt    // TOS > immediate (int)
	OpLeConstInt    // TOS <= immediate (int)
	OpGeConstInt    // TOS >= immediate (int)
	OpEqConstInt    // TOS == immediate (int)
	OpNeConstInt    // TOS != immediate (int)
	OpLtConstFloat  // TOS < immediate (float)
	OpGtConstFloat  // TOS > immediate (float)
	OpLeConstFloat  // TOS <= immediate (float)
	OpGeConstFloat  // TOS >= immediate (float)
	OpEqConstFloat  // TOS == immediate (float)
	OpNeConstFloat  // TOS != immediate (float)

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
	case OpEqInt:
		return "EQ_INT"
	case OpEqFloat:
		return "EQ_FLOAT"
	case OpEqString:
		return "EQ_STRING"
	case OpEqBool:
		return "EQ_BOOL"
	case OpNeInt:
		return "NE_INT"
	case OpNeFloat:
		return "NE_FLOAT"
	case OpNeString:
		return "NE_STRING"
	case OpNeBool:
		return "NE_BOOL"
	case OpLtInt:
		return "LT_INT"
	case OpLtFloat:
		return "LT_FLOAT"
	case OpGtInt:
		return "GT_INT"
	case OpGtFloat:
		return "GT_FLOAT"
	case OpLeInt:
		return "LE_INT"
	case OpLeFloat:
		return "LE_FLOAT"
	case OpGeInt:
		return "GE_INT"
	case OpGeFloat:
		return "GE_FLOAT"
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
	case OpStructOrdered:
		return "STRUCT_ORDERED"
	case OpGetFieldOffset:
		return "GET_FIELD_OFFSET"
	case OpSetFieldOffset:
		return "SET_FIELD_OFFSET"
	// Phase 4A
	case OpAddConstInt:
		return "ADD_CONST_INT"
	case OpSubConstInt:
		return "SUB_CONST_INT"
	case OpMulConstInt:
		return "MUL_CONST_INT"
	case OpDivConstInt:
		return "DIV_CONST_INT"
	case OpModConstInt:
		return "MOD_CONST_INT"
	case OpAddConstFloat:
		return "ADD_CONST_FLOAT"
	case OpSubConstFloat:
		return "SUB_CONST_FLOAT"
	case OpMulConstFloat:
		return "MUL_CONST_FLOAT"
	case OpDivConstFloat:
		return "DIV_CONST_FLOAT"
	// Phase 4B
	case OpIncGlobal:
		return "INC_GLOBAL"
	case OpDecGlobal:
		return "DEC_GLOBAL"
	case OpIncLocal:
		return "INC_LOCAL"
	case OpDecLocal:
		return "DEC_LOCAL"
	// Phase 4C
	case OpSquareInt:
		return "SQUARE_INT"
	case OpSquareFloat:
		return "SQUARE_FLOAT"
	// Phase 4D
	case OpLtConstInt:
		return "LT_CONST_INT"
	case OpGtConstInt:
		return "GT_CONST_INT"
	case OpLeConstInt:
		return "LE_CONST_INT"
	case OpGeConstInt:
		return "GE_CONST_INT"
	case OpEqConstInt:
		return "EQ_CONST_INT"
	case OpNeConstInt:
		return "NE_CONST_INT"
	case OpLtConstFloat:
		return "LT_CONST_FLOAT"
	case OpGtConstFloat:
		return "GT_CONST_FLOAT"
	case OpLeConstFloat:
		return "LE_CONST_FLOAT"
	case OpGeConstFloat:
		return "GE_CONST_FLOAT"
	case OpEqConstFloat:
		return "EQ_CONST_FLOAT"
	case OpNeConstFloat:
		return "NE_CONST_FLOAT"
	case OpHalt:
		return "HALT"
	case OpPrint:
		return "PRINT"
	default:
		return "UNKNOWN"
	}
}
