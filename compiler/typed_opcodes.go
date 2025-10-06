package compiler

import "minlang/vm"

// emitTypedAdd emits type-specialized addition opcode
func (c *Compiler) emitTypedAdd(leftType, rightType vm.ValueType) {
	// String concatenation
	if leftType == vm.StringType || rightType == vm.StringType {
		c.emit(vm.OpAddString)
		return
	}

	// Float addition (with promotion)
	if leftType == vm.FloatType || rightType == vm.FloatType {
		c.emit(vm.OpAddFloat)
		return
	}

	// Integer addition
	c.emit(vm.OpAddInt)
}

// emitTypedSub emits type-specialized subtraction opcode
func (c *Compiler) emitTypedSub(leftType, rightType vm.ValueType) {
	// Float subtraction (with promotion)
	if leftType == vm.FloatType || rightType == vm.FloatType {
		c.emit(vm.OpSubFloat)
		return
	}

	// Integer subtraction
	c.emit(vm.OpSubInt)
}

// emitTypedMul emits type-specialized multiplication opcode
func (c *Compiler) emitTypedMul(leftType, rightType vm.ValueType) {
	// Float multiplication (with promotion)
	if leftType == vm.FloatType || rightType == vm.FloatType {
		c.emit(vm.OpMulFloat)
		return
	}

	// Integer multiplication
	c.emit(vm.OpMulInt)
}

// emitTypedDiv emits type-specialized division opcode
func (c *Compiler) emitTypedDiv(leftType, rightType vm.ValueType) {
	// Float division (with promotion)
	if leftType == vm.FloatType || rightType == vm.FloatType {
		c.emit(vm.OpDivFloat)
		return
	}

	// Integer division
	c.emit(vm.OpDivInt)
}

// emitTypedMod emits type-specialized modulo opcode
func (c *Compiler) emitTypedMod(leftType, rightType vm.ValueType) {
	// Modulo is integer-only
	c.emit(vm.OpModInt)
}
