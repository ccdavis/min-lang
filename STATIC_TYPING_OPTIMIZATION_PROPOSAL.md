# Static Typing Optimization Proposal

**Date**: October 6, 2025
**Status**: Analysis & Proposal
**Estimated Impact**: 15-25% performance improvement

## Executive Summary

MinLang is statically typed with full type checking at compile-time, yet the VM performs extensive runtime type checks on every operation. By leveraging compile-time type information to emit type-specialized bytecode, we can eliminate most runtime type checks and achieve significant performance gains.

## Current Architecture Analysis

### Runtime Type Checking Overhead

**Current VM Behavior:**
```go
// executeBinaryOperation - called for EVERY arithmetic operation
func (vm *VM) executeBinaryOperation(op OpCode) error {
    right := vm.pop()
    left := vm.pop()

    // Runtime type check #1
    if left.Type == IntType && right.Type == IntType {
        return vm.executeBinaryIntegerOperation(...)
    }

    // Runtime type check #2
    if (left.Type == FloatType || left.Type == IntType) && ... {
        // More type conversion checks
        if left.Type == FloatType {
            leftVal = left.AsFloat()
        } else {
            leftVal = float64(left.AsInt())  // Conversion
        }
        return vm.executeBinaryFloatOperation(...)
    }
}
```

**Measured Overhead:**
- 13+ type checks in VM hot path
- Every OpAdd, OpSub, OpMul, OpDiv, OpMod checks types
- Every comparison operation checks types
- String concatenation checks if either operand is string
- Type conversions (int→float) happen at runtime

**Current Compiler:**
```go
// Emits generic opcode without type information
case "+":
    c.emit(vm.OpAdd)  // No type info!
```

### Type Information Available at Compile-Time

The compiler has complete type information:

1. **Variable declarations**: `var x: int = 42`
   - Type is known: int

2. **Binary expressions**: `x + y`
   - Symbol table knows x and y types
   - Result type is determinable

3. **Function parameters**: `func add(x: int, y: int): int`
   - Parameter types known
   - Return type known

4. **Struct fields**: `type Point struct { x: int, y: int }`
   - Field types and offsets known
   - Total struct size known

## Optimization Opportunities

### 1. Type-Specialized Opcodes (HIGH IMPACT - 15-20%)

**Problem:** Generic opcodes require runtime type checking

**Solution:** Emit type-specific opcodes

#### Current:
```
OpAdd         // Generic, checks types at runtime
OpSub
OpMul
OpDiv
```

#### Proposed:
```
OpAddInt      // int + int → int (no type check!)
OpAddFloat    // float + float → float (no type check!)
OpAddString   // string + string → string (no type check!)
OpSubInt      // int - int → int
OpSubFloat    // float - float → float
OpMulInt
OpMulFloat
OpDivInt
OpDivFloat
OpModInt      // Only for integers
```

**VM Implementation:**
```go
case OpAddInt:
    right := vm.pop()
    left := vm.pop()
    // No type checking - guaranteed to be ints!
    result := left.AsInt() + right.AsInt()
    vm.push(IntValue(result))

case OpAddFloat:
    right := vm.pop()
    left := vm.pop()
    // No type checking - guaranteed to be floats!
    result := left.AsFloat() + right.AsFloat()
    vm.push(FloatValue(result))
```

**Compiler Changes:**
```go
case "+":
    // Determine operand types from symbol table
    leftType := c.getExpressionType(node.Left)
    rightType := c.getExpressionType(node.Right)

    if leftType == IntType && rightType == IntType {
        c.emit(vm.OpAddInt)
    } else if leftType == FloatType || rightType == FloatType {
        c.emit(vm.OpAddFloat)
    } else if leftType == StringType || rightType == StringType {
        c.emit(vm.OpAddString)  // Concatenation
    }
```

**Benefits:**
- ✅ Eliminates 2 type checks per operation
- ✅ Eliminates type conversion code (int→float)
- ✅ Smaller, faster VM dispatch
- ✅ Better CPU branch prediction
- ✅ 15-20% estimated improvement

### 2. Specialized Comparison Opcodes (MEDIUM IMPACT - 5-10%)

**Current:**
```
OpEq, OpNe, OpLt, OpGt, OpLe, OpGe  // All check types
```

**Proposed:**
```
OpEqInt, OpEqFloat, OpEqString, OpEqBool
OpLtInt, OpLtFloat
OpGtInt, OpGtFloat
// etc.
```

**Benefits:**
- ✅ Faster comparisons in loops
- ✅ No type checking overhead
- ✅ Cleaner code

### 3. Struct Field Access Optimization (MEDIUM IMPACT - 5-8%)

**Current:**
```go
// OpGetField - looks up field by name at runtime
case OpGetField:
    fieldName := vm.constants[fieldIndex].AsString()
    obj := vm.pop()
    structVal := obj.AsStruct()

    // Runtime hash map lookup!
    value, exists := structVal.Fields[fieldName]
```

**Proposed:**
```go
// OpGetFieldOffset - direct array access
case OpGetFieldOffset:
    offset, _ := ReadOperand(ins, ip)
    obj := vm.pop()
    structVal := obj.AsStruct()

    // Direct access - compiler knows offset!
    value := structVal.FieldsArray[offset]
```

**Compiler Change:**
```go
// When compiling struct access: point.x
structDef := c.getStructDefinition("Point")
fieldOffset := structDef.GetFieldOffset("x")  // 0 for x, 1 for y
c.emit(vm.OpGetFieldOffset, fieldOffset)
```

**Benefits:**
- ✅ O(1) array access instead of O(log n) map lookup
- ✅ Better cache locality
- ✅ Smaller bytecode (offsets are smaller than strings)

### 4. Inline Type Conversion (LOW IMPACT - 2-3%)

When mixing int and float:
```
var x: int = 5
var y: float = 3.14
var z: float = x + y
```

**Current:** Runtime conversion

**Proposed:** Compile-time conversion
```go
// Compiler emits:
OpLoadLocal 0      // x (int)
OpIntToFloat       // Convert at the point we know it's needed
OpLoadLocal 1      // y (float)
OpAddFloat         // Now both are floats
```

### 5. Array Element Type Specialization (FUTURE - 10-15%)

**Observation:** Arrays are homogeneous - all elements same type

**Current:**
```go
// Array stores []Value - each element has Type field
type ArrayValue struct {
    Elements []Value
}
```

**Proposed:**
```go
// Specialized arrays
type IntArray struct {
    Elements []int64  // Direct int storage, no boxing!
}

type FloatArray struct {
    Elements []float64
}
```

**Benefits:**
- ✅ 50% memory reduction (no Type field per element)
- ✅ Better cache locality
- ✅ Faster access
- ⚠️ Requires significant refactoring

## Implementation Priority

### Phase 1: Type-Specialized Arithmetic (IMPLEMENT NOW)
**Effort:** 2-3 hours
**Impact:** 15-20%

- Add OpAddInt, OpAddFloat, OpSubInt, OpSubFloat, OpMulInt, OpMulFloat, OpDivInt, OpDivFloat
- Implement type inference in compiler
- Emit specialized opcodes based on operand types
- Update VM to handle new opcodes

### Phase 2: Specialized Comparisons
**Effort:** 1-2 hours
**Impact:** 5-10%

- Add OpEqInt, OpEqFloat, OpLtInt, OpLtFloat, etc.
- Update compiler to emit based on types
- Simplify VM comparison logic

### Phase 3: Struct Field Offsets
**Effort:** 2-3 hours
**Impact:** 5-8%

- Track struct definitions with field offsets
- Change struct representation to use array instead of map
- Emit offset-based field access

### Phase 4: Specialized Arrays (FUTURE)
**Effort:** 8-10 hours
**Impact:** 10-15%

- Major refactoring
- New specialized array types
- Type-safe array operations

## Example: Before vs After

### Before (Current):
```
Source: var x: int = 10 + 20

Bytecode:
  PUSH 0          // constant 10
  PUSH 1          // constant 20
  ADD             // Generic add - checks types at runtime!
  STORE_LOCAL 0

VM Execution:
  1. Pop 20 (check type)
  2. Pop 10 (check type)
  3. if IntType && IntType → call executeBinaryIntegerOperation
  4. Inside: switch OpAdd, compute result
  5. Push result
```

### After (Optimized):
```
Source: var x: int = 10 + 20

Bytecode:
  PUSH 0          // constant 10
  PUSH 1          // constant 20
  ADD_INT         // Type-specific - NO runtime checks!
  STORE_LOCAL 0

VM Execution:
  1. Pop 20 (no check - guaranteed int)
  2. Pop 10 (no check - guaranteed int)
  3. result = left + right  // Direct computation
  4. Push IntValue(result)
```

**Savings:**
- Eliminated: 2 type checks
- Eliminated: 1 function call (executeBinaryIntegerOperation)
- Eliminated: 1 switch statement
- Result: 3-4x faster for arithmetic operations

## Estimated Performance Gains

### Conservative Estimate:
| Optimization | Improvement |
|--------------|-------------|
| Typed arithmetic ops | 15-20% |
| Typed comparisons | 5-10% |
| Struct field offsets | 5-8% |
| **Total (Phases 1-3)** | **25-38%** |

### On Mandelbrot Benchmark:
- Current: 18.36s
- After Phase 1: ~15.5s (-15-20%)
- After Phases 1-3: ~13.5s (-25-30%)

**Combined with existing optimizations:**
- Original: 27.68s
- After all optimizations: ~13.5s
- **Total improvement: 51% faster than original**
- **vs Python: ~4.3× faster** (up from 3.16×)

## Implementation Challenges

### 1. Type Inference
**Challenge:** Compiler needs to track expression types

**Solution:** Add type inference pass
```go
func (c *Compiler) inferType(node ast.Expression) ValueType {
    switch n := node.(type) {
    case *ast.IntegerLiteral:
        return IntType
    case *ast.Identifier:
        symbol := c.symbolTable.Resolve(n.Value)
        return symbol.Type
    case *ast.InfixExpression:
        leftType := c.inferType(n.Left)
        rightType := c.inferType(n.Right)
        return c.resultType(n.Operator, leftType, rightType)
    }
}
```

### 2. Type Coercion
**Challenge:** int + float → float (need conversion)

**Solution:** Emit conversion opcodes
```go
if leftType == IntType && rightType == FloatType {
    c.emit(OpIntToFloat)  // Convert TOS to float
    c.emit(OpAddFloat)
}
```

### 3. Backward Compatibility
**Challenge:** Existing bytecode won't work

**Solution:**
- Keep old opcodes for compatibility
- Add version to bytecode
- Or: just accept breaking change (it's early stage)

## Risks & Mitigations

**Risk:** Increased bytecode size (more opcodes)
- **Mitigation:** Offset by removing string constants for field names
- **Net:** Likely smaller or same size

**Risk:** Compiler complexity
- **Mitigation:** Type inference is well-understood
- **Benefit:** Better error messages for type mismatches

**Risk:** Testing burden
- **Mitigation:** Existing test suite catches regressions
- **Benefit:** More specialized tests

## Conclusion

MinLang's static typing is currently underutilized. By emitting type-specialized bytecode, we can:

1. **Eliminate 90% of runtime type checks**
2. **Achieve 25-38% performance improvement** (Phases 1-3)
3. **Simplify VM code** (fewer branches, clearer logic)
4. **Enable future optimizations** (JIT compilation becomes easier)

**Recommendation:** Implement Phase 1 (type-specialized arithmetic) immediately for maximum impact with minimal effort.

---

**Next Steps:**
1. Implement type inference in compiler
2. Add typed arithmetic opcodes
3. Benchmark results
4. Document findings
5. Proceed to Phase 2 if results are positive
