# MinLang VM Optimization Analysis

**Date**: October 5, 2025
**Analysis Type**: GC Pressure & Performance Review

## Executive Summary

Found **5 major optimization opportunities** in the VM after implementing bugfixes:

1. ✅ **CRITICAL BUG**: Direct local operations defined but not implemented
2. 🔴 **HIGH IMPACT**: Closure allocation on every function call
3. 🟡 **MEDIUM IMPACT**: 34 error string allocations
4. 🟢 **LOW IMPACT**: Redundant float64 initializations
5. 🟢 **MEMORY**: Excessive globals array pre-allocation

## Critical Issues Found & Fixed

### 1. ✅ Missing OpAddLocal/OpSubLocal/OpMulLocal/OpDivLocal Handlers

**Status**: FIXED
**Impact**: CRITICAL - Caused crashes when peephole optimization triggered

**Problem**:
```go
// Compiler emits these:
0015  ADD_LOCAL 1

// But VM had no handlers!
// Result: panic: index out of range
```

**Solution**: Implemented full handlers in vm/vm.go:156-223

**Test**:
```bash
$ ./cmd/minlang/minlang test_local_ops.min
15  # Correct! (was crashing before)
```

## High-Impact Optimizations

### 2. 🔴 Closure Allocation on Every Function Call

**Location**: vm/vm.go:845
**Impact**: HIGH - Allocates on EVERY non-closure function call
**GC Pressure**: Significant for recursive/loop-heavy code

**Current Code**:
```go
func (vm *VM) callFunction(fn *Function, numArgs int) error {
    // Allocates NEW closure every time!
    cl := &Closure{Fn: fn, Free: []Value{}}  // 2 allocations
    basePointer := vm.sp - numArgs
    // ...
}
```

**Problem**:
- Allocates `&Closure{}` heap object
- Allocates `[]Value{}` empty slice
- Happens for every function call (factorial(5) = 6 allocations)

**Proposed Solution A - Shared Empty Closure**:
```go
var emptyFreeVars []Value = nil  // Use nil instead of []Value{}

func (vm *VM) callFunction(fn *Function, numArgs int) error {
    // Reuse frame's closure if possible, or create lightweight one
    cl := &Closure{Fn: fn, Free: emptyFreeVars}  // 1 allocation
    // ...
}
```

**Proposed Solution B - Frame-Embedded Closure** (Better):
```go
type Frame struct {
    cl          *Closure
    tempClosure Closure  // Embedded for non-closure calls
    // ...
}

func (vm *VM) callFunction(fn *Function, numArgs int) error {
    // Reuse frame's embedded closure (zero allocations!)
    frame := vm.frames[vm.framesIndex]
    frame.tempClosure.Fn = fn
    frame.tempClosure.Free = nil
    frame.cl = &frame.tempClosure
    // ...
}
```

**Estimated Improvement**: 5-10% for function-heavy code

### 3. 🟡 Error String Allocations

**Location**: Throughout vm/vm.go
**Count**: 34 `fmt.Errorf()` calls
**Impact**: MEDIUM - Allocates on error paths

**Current Pattern**:
```go
return fmt.Errorf("division by zero")  // Allocates string
return fmt.Errorf("wrong number of arguments: want=%d, got=%d", want, got)
```

**Proposed Solution - Pre-allocated Errors**:
```go
// Package level
var (
    ErrDivisionByZero       = errors.New("division by zero")
    ErrModuloByZero         = errors.New("modulo by zero")
    ErrStackOverflow        = errors.New("stack overflow")
    ErrStackUnderflow       = errors.New("stack underflow")
    ErrUnsupportedOperands  = errors.New("unsupported operand types")
    // ... etc
)

// In code
if right == 0 {
    return ErrDivisionByZero  // No allocation!
}
```

**For formatted errors**, use errorf only when necessary:
```go
// Keep these as-is (need formatting):
return fmt.Errorf("wrong number of arguments: want=%d, got=%d", want, got)

// Convert these to constants:
return fmt.Errorf("division by zero") → return ErrDivisionByZero
```

**Estimated Improvement**: 1-2% (errors are cold paths, but still worthwhile)

## Low-Impact Optimizations

### 4. 🟢 Redundant Float Initializations

**Location**: vm/vm.go:609-610, 694-695, 188-189 (direct local ops)
**Impact**: LOW - Minor cleanup

**Current Code**:
```go
leftVal := float64(0)   // Unnecessary
rightVal := float64(0)  // Unnecessary

if left.Type == FloatType {
    leftVal = left.AsFloat()  // Immediately reassigned
} else {
    leftVal = float64(left.AsInt())  // Immediately reassigned
}
```

**Proposed Solution**:
```go
var leftVal, rightVal float64

if left.Type == FloatType {
    leftVal = left.AsFloat()
} else {
    leftVal = float64(left.AsInt())
}
```

**Estimated Improvement**: <1% (compiler probably optimizes this anyway)

### 5. 🟢 Excessive Globals Pre-allocation

**Location**: vm/vm.go:64
**Impact**: MEMORY - 512KB allocated but rarely used

**Current Code**:
```go
globals: make([]Value, GlobalsSize),  // 65536 * 8 bytes = 512KB
```

**Observation**:
- Most programs use <100 globals
- Pre-allocating 65536 wastes memory
- But doesn't cause GC pressure (no pointers in Value array)

**Proposed Solution** (if memory is concern):
```go
const InitialGlobalsSize = 256

globals: make([]Value, InitialGlobalsSize),

// Add growth logic in OpStoreGlobal:
if globalIndex >= len(vm.globals) {
    // Grow globals array
    newSize := globalIndex * 2
    newGlobals := make([]Value, newSize)
    copy(newGlobals, vm.globals)
    vm.globals = newGlobals
}
```

**Trade-off**: Adds branch check to OpStoreGlobal, may slow writes
**Recommendation**: Keep current approach (memory is cheap, speed matters more)

## Additional Observations

### Already Well-Optimized

✅ **Stack/Frame pre-allocation**: Good, avoids allocation in hot path
✅ **Frame pooling**: Reuses frames, excellent
✅ **Value-based operations**: No boxing/unboxing, very good
✅ **Local IP caching**: Reduces indirection, good

### Opportunities for Future Work

1. **Inline critical operations**: `push()`, `pop()` could be inlined
2. **Computed goto**: Replace switch with jump table (advanced)
3. **Constant folding**: Compiler could pre-compute constant expressions
4. **Loop unrolling**: For tight loops (advanced)

## Implementation Priority

### Priority 1 - Implement Now ✅
- [x] Fix missing direct local op handlers (CRITICAL - DONE)

### Priority 2 - High Value
- [ ] Eliminate closure allocation in callFunction()
- [ ] Pre-allocate common errors

### Priority 3 - Nice to Have
- [ ] Clean up redundant float64 initializations
- [ ] Consider globals growth strategy (only if memory is issue)

## Performance Impact Estimate

| Optimization | Estimated Improvement | Implementation Effort |
|--------------|----------------------|----------------------|
| Direct local ops (fixed) | Baseline correction | Done ✅ |
| Eliminate closure alloc | 5-10% | Low (30 mins) |
| Pre-allocated errors | 1-2% | Low (20 mins) |
| Float cleanup | <1% | Trivial (5 mins) |
| **TOTAL** | **6-13%** | **~1 hour** |

## Next Steps

1. ✅ Test current state with direct local ops
2. Implement closure allocation fix
3. Add pre-allocated errors
4. Clean up minor issues
5. Final benchmark run
6. Document results

---

**Analysis Complete**: VM is in good shape with clear optimization path forward.
