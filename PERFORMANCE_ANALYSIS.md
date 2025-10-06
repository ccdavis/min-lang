# MinLang VM Performance Analysis

## Overview

This document analyzes the MinLang VM implementation for performance bottlenecks and proposes optimizations based on the measured 20.7% of Python's performance.

## Measured Performance

**Current Performance:**
- MinLang: 4.6M iterations/second
- Python: 22.2M iterations/second
- Ratio: 20.7% (MinLang is 4.8× slower than Python)

**Target:** Achieve 30-50% of Python's performance through targeted optimizations.

## Critical Performance Issues

### 1. Interface{} Boxing (MAJOR - ~30-40% overhead)

**Problem:**
```go
type Value struct {
    Type ValueType      // 1 byte + 7 padding
    Data interface{}    // 16 bytes (pointer + type)
}
// Total: 24 bytes per value
```

Every Value uses `interface{}` which causes:
- **Heap allocation** for most types (int64, float64, bool escape to heap)
- **Type assertion overhead** on every operation (`.Data.(int64)`)
- **GC pressure** - millions of allocations in benchmark
- **24-byte copies** on every push/pop operation

**Evidence:**
```go
// vm/value.go:29-34
func IntValue(i int64) Value {
    return Value{Type: IntType, Data: i}  // int64 boxes to interface{}
}

func (v Value) AsInt() int64 {
    return v.Data.(int64)  // Type assertion on every use
}
```

**Impact:** Every arithmetic operation requires:
1. Pop (24-byte copy + type assertion)
2. Pop (24-byte copy + type assertion)
3. Compute
4. Push (24-byte copy + interface box allocation)

In the Mandelbrot benchmark inner loop (x² + y² > 4.0):
- 6 pops (6× 24 bytes = 144 bytes copied)
- 6 type assertions
- 3 pushes (3 allocations + 72 bytes copied)
- **Per iteration:** ~216 bytes copied, 9 allocations

With 122M iterations: **26 GB copied, 1.1 billion allocations**

**Solution:** Tagged union instead of interface{}

```go
type Value struct {
    Type ValueType
    _    [7]byte      // explicit padding for alignment
    Data uint64       // Union: can hold int64, float64, or pointer
}

func IntValue(i int64) Value {
    return Value{
        Type: IntType,
        Data: uint64(i),  // No boxing, just bit cast
    }
}

func (v Value) AsInt() int64 {
    return int64(v.Data)  // No type assertion, just bit cast
}
```

**Benefits:**
- Zero allocations for int/float/bool
- No type assertions (compile-time known)
- Still 24 bytes but no GC pressure
- **Estimated gain: 30-40%**

### 2. currentFrame() Call Overhead (MODERATE - ~10-15%)

**Problem:**
```go
// Called 4+ times per VM loop iteration
for vm.currentFrame().ip < len(vm.currentFrame().Instructions()) {
    vm.currentFrame().ip++
    ip = vm.currentFrame().ip - 1
    ins = vm.currentFrame().Instructions()
    op = OpCode(ins[ip])
    // ...
}

func (vm *VM) currentFrame() *Frame {
    return vm.frames[vm.framesIndex-1]  // Array lookup + subtraction
}
```

**Impact:**
- 500M+ function calls in benchmark (4 per iteration × 122M)
- Slice bounds check on every call
- Cannot be inlined (returns pointer to heap object)

**Solution:** Cache frame pointer

```go
func (vm *VM) Run() error {
    var frame *Frame

    for {
        frame = vm.frames[vm.framesIndex-1]  // Cache once per outer loop
        ip := frame.ip
        ins := frame.cl.Fn.Instructions

        for ip < len(ins) {
            op := OpCode(ins[ip])
            ip++

            switch op {
            // Use local ip variable, update frame.ip only when needed
            }
        }
    }
}
```

**Benefits:**
- Eliminate 3-4 function calls per instruction
- Better CPU register usage (ip in register)
- Compiler can optimize better
- **Estimated gain: 10-15%**

### 3. String Allocation for Map Keys (MODERATE - ~10-20% for map-heavy code)

**Problem:**
```go
// vm/vm.go:343, 422, 440, 465
keyStr := key.String()  // Allocates string on every map access
mapData.Pairs[keyStr] = value
```

Every map operation calls `Value.String()` which allocates:
```go
func (v Value) String() string {
    switch v.Type {
    case IntType:
        return fmt.Sprintf("%d", v.AsInt())  // Allocates string
    case FloatType:
        return fmt.Sprintf("%f", v.AsFloat())  // Allocates string
    // ...
    }
}
```

**Impact:**
- 2 allocations per map access (key string + fmt.Sprintf)
- GC pressure
- String comparison overhead

**Solution:** Use value-based map keys

```go
type MapKey struct {
    Type ValueType
    IntKey   int64
    StrKey   string
}

type MapData struct {
    Pairs map[MapKey]Value  // No allocation for int keys
}
```

Or use separate maps:
```go
type MapData struct {
    IntPairs map[int64]Value
    StrPairs map[string]Value
}
```

**Benefits:**
- Zero allocations for integer map keys
- Faster comparison (int vs string)
- **Estimated gain: 10-20% for map operations**

### 4. Closure Allocation on Every Function Call (MINOR - ~5%)

**Problem:**
```go
// vm/vm.go:794
func (vm *VM) callFunction(fn *Function, numArgs int) error {
    cl := &Closure{Fn: fn, Free: []Value{}}  // Allocates on every call
    frame := NewFrame(cl, vm.sp-numArgs)      // Allocates Frame
    // ...
}
```

**Impact:**
- 2 allocations per function call
- In benchmark with nested loops: thousands of allocations

**Solution:** Reuse frame pool

```go
type VM struct {
    framePool []*Frame  // Pre-allocated frame pool
    // ...
}

func (vm *VM) callFunction(fn *Function, numArgs int) error {
    // Reuse frame from pool instead of allocating
    frame := vm.frames[vm.framesIndex]
    if frame == nil {
        frame = &Frame{}
        vm.frames[vm.framesIndex] = frame
    }
    frame.Reset(fn, vm.sp-numArgs)
    vm.framesIndex++
    // ...
}
```

**Benefits:**
- Zero allocations after warmup
- Better cache locality
- **Estimated gain: 5%**

### 5. Value Copying on Stack Operations (MODERATE - ~15-20%)

**Problem:**
```go
func (vm *VM) push(val Value) error {
    vm.stack[vm.sp] = val  // Copies 24 bytes
    vm.sp++
    return nil
}

func (vm *VM) pop() Value {
    val := vm.stack[vm.sp-1]  // Copies 24 bytes
    vm.sp--
    return val
}
```

**Impact:**
- Every push/pop copies 24 bytes
- In tight loop: pop left (24B), pop right (24B), push result (24B) = 72 bytes
- With 400M+ stack ops in benchmark: **28+ GB copied**

**Solution:** Already using value copies (can't avoid without major refactor), but:

1. **Reduce Value size** (interface{} → union as mentioned)
2. **Reduce stack operations** via peephole optimization
3. **Use SSA form** to eliminate redundant push/pop pairs

**Example peephole optimization:**
```
Before:
  OpPush const[x]
  OpPush const[y]
  OpMul

After:
  OpMulConst const[x] const[y]  // Inline multiplication of constants
```

**Benefits:**
- Fewer stack operations
- Less copying
- **Estimated gain: 15-20%**

### 6. Switch Statement Dispatch (MINOR - ~5-10%)

**Problem:**
```go
switch op {
case OpPush:
    // ...
case OpPop:
    // ...
// ... 30+ cases
}
```

**Impact:**
- Go compiles this as jump table, which is fast
- But still has bounds checks and indirection
- Cannot be optimized across iterations

**Solution:** Computed goto (not available in Go) or threaded interpreter

Since Go doesn't support computed goto, we could use:
1. **Function dispatch table** (worse than switch in Go)
2. **Manual loop unrolling** for common sequences
3. **Trace compilation** for hot paths (complex)

**Benefits:**
- Limited in Go
- **Estimated gain: 5-10% max**

## Stack Machine Design Issues

### 1. Stack-Based vs Register-Based

**Current (Stack):**
```
x = a + b
  OpLoadLocal 0  // push a
  OpLoadLocal 1  // push b
  OpAdd          // pop 2, push result
  OpStoreLocal 2 // pop and store
```
4 instructions, 6 stack operations (3 push, 3 pop)

**Register-Based:**
```
x = a + b
  OpAddReg 2, 0, 1  // x = a + b
```
1 instruction, 0 stack operations

**Impact:**
- Stack-based uses 4× more instructions
- 6× more memory traffic
- Studies show register VMs are 20-50% faster

**Solution:** Convert to register-based VM

**Challenges:**
- Major rewrite of compiler and VM
- More complex instruction encoding
- Larger instruction size

**Benefits:**
- **Estimated gain: 20-50%**
- Would bring MinLang to 40-100% of Python's speed

### 2. Excessive Push/Pop for Locals

**Problem:**
```go
// Every local variable access goes through stack
case OpLoadLocal:
    localIndex, _ := ReadOperand(ins, ip+1)
    frame := vm.currentFrame()
    err := vm.push(vm.stack[frame.basePointer+localIndex])  // Copy from stack to stack!
```

**Impact:**
- Locals are already on the stack
- We copy them to TOS just to operate on them
- In benchmark: millions of pointless copies

**Solution:** Direct stack addressing

```go
case OpAddLocal:  // Add TOS with local, store in TOS
    localIndex, _ := ReadOperand(ins, ip+1)
    frame := vm.currentFrame()
    top := vm.pop()
    local := vm.stack[frame.basePointer+localIndex]
    // Add without intermediate copy
    vm.push(add(top, local))
```

**Benefits:**
- Eliminate push for right operand
- **Estimated gain: 10-15%**

## GC Pressure Analysis

**Allocation Hotspots (estimated from benchmark):**

1. **Interface boxing** - 1.1 billion allocations
   - Every IntValue(), FloatValue() boxes to interface{}
   - **Impact: CRITICAL**

2. **String allocations** - ~500M allocations
   - Map key conversions
   - String concatenation
   - fmt.Sprintf in Value.String()
   - **Impact: HIGH**

3. **Closure/Frame allocations** - ~10M allocations
   - callFunction(), callClosure()
   - MakeClosure operations
   - **Impact: MODERATE**

4. **Array/Map/Struct allocations** - ~1M allocations
   - NewArrayValue, NewMapValue, NewStructValue
   - **Impact: LOW** (amortized over structure lifetime)

**GC Impact:**
- With ~1.6 billion allocations in 27 seconds
- GC runs every ~100-200ms (estimated)
- Stop-the-world pauses: ~100-200 total pauses
- Time lost to GC: ~2-4 seconds (7-15% of runtime)

**Solution:** Arena allocation

```go
type VM struct {
    arena *Arena  // Bump allocator for Values
    // ...
}

type Arena struct {
    buffer []Value
    offset int
}

func (a *Arena) Allocate() *Value {
    if a.offset >= len(a.buffer) {
        panic("arena full")
    }
    v := &a.buffer[a.offset]
    a.offset++
    return v
}

func (a *Arena) Reset() {
    a.offset = 0  // Bulk free - no GC pressure
}
```

**Benefits:**
- Near-zero GC pressure
- Faster allocation (bump pointer vs heap allocation)
- Can reset arena between program runs
- **Estimated gain: 7-15%** (eliminate GC overhead)

## Optimization Priority

### High Priority (30-50% total gain)

1. **Replace interface{} with tagged union** (30-40%)
   - Eliminate interface boxing
   - Zero allocations for primitives
   - Complexity: Medium

2. **Cache frame pointer and IP** (10-15%)
   - Eliminate currentFrame() calls
   - Keep IP in local variable
   - Complexity: Low

### Medium Priority (20-35% total gain)

3. **Value-based map keys** (10-20%)
   - Eliminate string allocations
   - Complexity: Medium

4. **Reduce Value copies via peephole opts** (15-20%)
   - Inline constant operations
   - Combine push/pop sequences
   - Complexity: Medium

### Low Priority (10-20% total gain)

5. **Frame/closure pooling** (5%)
   - Reuse allocated frames
   - Complexity: Low

6. **Direct local addressing** (10-15%)
   - Operate on locals without push/pop
   - Complexity: Medium

### Long-Term (50-100% total gain)

7. **Register-based VM** (20-50%)
   - Major architecture change
   - Complexity: Very High

8. **Trace JIT compilation** (100-1000%)
   - Compile hot loops to native code
   - Complexity: Extreme

## Summary Table

| Optimization | Estimated Gain | Complexity | Implementation Time |
|--------------|----------------|------------|---------------------|
| Tagged union Value | 30-40% | Medium | 2-3 days |
| Cache frame/IP | 10-15% | Low | 4 hours |
| Value-based map keys | 10-20% | Medium | 1 day |
| Peephole optimization | 15-20% | Medium | 2-3 days |
| Frame pooling | 5% | Low | 2 hours |
| Direct local ops | 10-15% | Medium | 1-2 days |
| **Total (all above)** | **80-125%** | - | **1-2 weeks** |
| Register-based VM | 20-50% | High | 2-3 weeks |
| Trace JIT | 100-1000% | Extreme | 2-3 months |

## Projected Performance

**Current:** 4.6M iter/s (20.7% of Python)

**With high+medium priority optimizations:**
- 4.6M × 1.8 (80% gain) = **8.3M iter/s**
- **37% of Python's performance**

**With all low-priority optimizations:**
- 4.6M × 2.25 (125% gain) = **10.4M iter/s**
- **47% of Python's performance**

**With register-based VM:**
- 10.4M × 1.35 (35% more) = **14M iter/s**
- **63% of Python's performance**

## Recommendations

For **best ROI** (return on implementation time):

1. **Tagged union Value** - Highest impact, moderate effort
2. **Cache frame/IP** - Quick win, low effort
3. **Frame pooling** - Easy win

These three alone could achieve **45-60% gain** in **3-4 days** of work.

For **matching Python's performance** (~22M iter/s):
- Would need JIT compilation or extreme optimization
- Not worth it for pedagogical language
- Better to document why interpreted code is slower

## Conclusion

MinLang's current 20.7% of Python's performance is **reasonable** for a simple stack-based interpreter, but there's significant low-hanging fruit:

- **Interface{} boxing** is the #1 bottleneck
- **Excessive frame calls** are #2
- **Stack operations** are inherently costly in stack-based VMs

With targeted optimizations, **50% of Python's speed** is achievable without major architectural changes. This would make MinLang quite competitive with interpreted languages while maintaining code clarity.
