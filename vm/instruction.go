package vm

import (
	"encoding/binary"
	"fmt"
)

// Instruction represents a single bytecode instruction with its operands
type Instruction []byte

// Make creates an instruction from an opcode and operands
func Make(op OpCode, operands ...int) []byte {
	ins := []byte{byte(op)}

	for _, operand := range operands {
		// Encode operands as 2-byte values (big-endian)
		buf := make([]byte, 2)
		binary.BigEndian.PutUint16(buf, uint16(operand))
		ins = append(ins, buf...)
	}

	return ins
}

// ReadOperand reads a 2-byte operand from the instruction stream
func ReadOperand(ins []byte, offset int) (int, int) {
	if offset+2 > len(ins) {
		return 0, offset
	}
	operand := int(binary.BigEndian.Uint16(ins[offset:]))
	return operand, offset + 2
}

// Disassemble converts bytecode to a human-readable format
func Disassemble(bytecode []byte) string {
	result := ""
	i := 0

	for i < len(bytecode) {
		op := OpCode(bytecode[i])
		result += fmt.Sprintf("%04d  %s", i, op.String())

		switch op {
		case OpMakeClosure:
			if i+4 < len(bytecode) {
				fnIndex, _ := ReadOperand(bytecode, i+1)
				numFree, _ := ReadOperand(bytecode, i+3)
				result += fmt.Sprintf(" %d %d", fnIndex, numFree)
				i += 5
			} else {
				i++
			}
		// Phase 4B: Inc/Dec have 2 operands (variable index and amount)
		case OpIncGlobal, OpDecGlobal, OpIncLocal, OpDecLocal:
			if i+4 < len(bytecode) {
				varIndex, _ := ReadOperand(bytecode, i+1)
				amount, _ := ReadOperand(bytecode, i+3)
				result += fmt.Sprintf(" %d %d", varIndex, amount)
				i += 5
			} else {
				i++
			}
		case OpPush, OpLoadGlobal, OpStoreGlobal, OpLoadLocal, OpStoreLocal,
			OpLoadFree, OpJump, OpJumpIfFalse, OpJumpIfTrue, OpCall,
			OpGetBuiltin, OpArray, OpMap, OpStruct, OpGetField, OpSetField,
			OpAddLocal, OpSubLocal, OpMulLocal, OpDivLocal,
			OpGetFieldOffset, OpSetFieldOffset,
			// Phase 4A: Const ops have 1 operand (constant value)
			OpAddConstInt, OpSubConstInt, OpMulConstInt, OpDivConstInt, OpModConstInt,
			OpAddConstFloat, OpSubConstFloat, OpMulConstFloat, OpDivConstFloat,
			// Phase 4D: Compare with const have 1 operand (constant value)
			OpLtConstInt, OpGtConstInt, OpLeConstInt, OpGeConstInt, OpEqConstInt, OpNeConstInt,
			OpLtConstFloat, OpGtConstFloat, OpLeConstFloat, OpGeConstFloat, OpEqConstFloat, OpNeConstFloat:
			if i+2 < len(bytecode) {
				operand, _ := ReadOperand(bytecode, i+1)
				result += fmt.Sprintf(" %d", operand)
				i += 3
			} else {
				i++
			}
		default:
			i++
		}

		result += "\n"
	}

	return result
}
