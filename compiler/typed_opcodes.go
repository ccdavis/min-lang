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

// emitTypedEq emits type-specialized equality opcode (Phase 2)
func (c *Compiler) emitTypedEq(leftType, rightType vm.ValueType) {
	// For equality, both operands should be the same type
	// (type checker should ensure this)
	switch leftType {
	case vm.IntType:
		c.emit(vm.OpEqInt)
	case vm.FloatType:
		c.emit(vm.OpEqFloat)
	case vm.StringType:
		c.emit(vm.OpEqString)
	case vm.BoolType:
		c.emit(vm.OpEqBool)
	default:
		// Fall back to generic comparison for complex types
		c.emit(vm.OpEq)
	}
}

// emitTypedNe emits type-specialized inequality opcode (Phase 2)
func (c *Compiler) emitTypedNe(leftType, rightType vm.ValueType) {
	switch leftType {
	case vm.IntType:
		c.emit(vm.OpNeInt)
	case vm.FloatType:
		c.emit(vm.OpNeFloat)
	case vm.StringType:
		c.emit(vm.OpNeString)
	case vm.BoolType:
		c.emit(vm.OpNeBool)
	default:
		c.emit(vm.OpNe)
	}
}

// emitTypedLt emits type-specialized less-than opcode (Phase 2)
func (c *Compiler) emitTypedLt(leftType, rightType vm.ValueType) {
	// Float promotion for mixed types
	if leftType == vm.FloatType || rightType == vm.FloatType {
		c.emit(vm.OpLtFloat)
		return
	}

	if leftType == vm.IntType {
		c.emit(vm.OpLtInt)
		return
	}

	// Fall back to generic
	c.emit(vm.OpLt)
}

// emitTypedGt emits type-specialized greater-than opcode (Phase 2)
func (c *Compiler) emitTypedGt(leftType, rightType vm.ValueType) {
	// Float promotion for mixed types
	if leftType == vm.FloatType || rightType == vm.FloatType {
		c.emit(vm.OpGtFloat)
		return
	}

	if leftType == vm.IntType {
		c.emit(vm.OpGtInt)
		return
	}

	// Fall back to generic
	c.emit(vm.OpGt)
}

// emitTypedLe emits type-specialized less-than-or-equal opcode (Phase 2)
func (c *Compiler) emitTypedLe(leftType, rightType vm.ValueType) {
	// Float promotion for mixed types
	if leftType == vm.FloatType || rightType == vm.FloatType {
		c.emit(vm.OpLeFloat)
		return
	}

	if leftType == vm.IntType {
		c.emit(vm.OpLeInt)
		return
	}

	// Fall back to generic
	c.emit(vm.OpLe)
}

// emitTypedGe emits type-specialized greater-than-or-equal opcode (Phase 2)
func (c *Compiler) emitTypedGe(leftType, rightType vm.ValueType) {
	// Float promotion for mixed types
	if leftType == vm.FloatType || rightType == vm.FloatType {
		c.emit(vm.OpGeFloat)
		return
	}

	if leftType == vm.IntType {
		c.emit(vm.OpGeInt)
		return
	}

	// Fall back to generic
	c.emit(vm.OpGe)
}
