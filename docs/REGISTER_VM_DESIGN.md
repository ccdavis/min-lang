# Register-Based VM Design for MinLang

## Overview

This document describes the design of an alternative register-based virtual machine for MinLang that runs alongside the existing stack-based VM. The goal is to achieve 25-35% performance improvement while maintaining compatibility with the existing language.

## Motivation

Based on Lua 5.0's transition from stack to registers:
- **26-32% faster execution** (empirical data from Lua benchmarks)
- **Fewer instructions**: No push/pop operations
- **Less memory traffic**: Reduced value copying
- **Better locality**: Locals stay in registers
- **More compact code**: Shorter instruction sequences

## Architecture Overview

```
┌─────────────────────────────────────────────────────────┐
│                    MinLang Compiler                      │
│                                                          │
│  ┌──────────────┐                ┌──────────────┐      │
│  │ Parser & AST │  ──────────>   │   Compiler   │      │
│  └──────────────┘                └───────┬──────┘      │
│                                           │              │
│                           ┌───────────────┴─────────┐   │
│                           │                         │   │
│                           ▼                         ▼   │
│                    ┌─────────────┐         ┌────────────┐
│                    │ Stack       │         │ Register   │
│                    │ Bytecode    │         │ Bytecode   │
│                    └──────┬──────┘         └─────┬──────┘
└───────────────────────────┼──────────────────────┼───────┘
                            │                      │
                            ▼                      ▼
                    ┌──────────────┐       ┌──────────────┐
                    │  Stack VM    │       │ Register VM  │
                    │  (existing)  │       │   (new)      │
                    └──────────────┘       └──────────────┘
```

## Register Instruction Set Design

### Instruction Format

**32-bit instruction encoding:**
```
┌───────────┬────────┬────────┬────────┐
│  Opcode   │   A    │   B    │   C    │
│  (8 bits) │(8 bits)│(8 bits)│(8 bits)│
└───────────┴────────┴────────┴────────┘

Or for larger constants:
┌───────────┬────────┬──────────────────┐
│  Opcode   │   A    │       Bx         │
│  (8 bits) │(8 bits)│    (16 bits)     │
└───────────┴────────┴──────────────────┘
```

- **A**: Destination register (0-255)
- **B**: Source register 1 (0-255)
- **C**: Source register 2 (0-255)
- **Bx**: Large constant index (0-65535)

### Core Instructions

#### Arithmetic (3-register format)
```
ADD    R(A) = R(B) + R(C)        // Integer or float addition
SUB    R(A) = R(B) - R(C)        // Subtraction
MUL    R(A) = R(B) * R(C)        // Multiplication
DIV    R(A) = R(B) / R(C)        // Division
MOD    R(A) = R(B) % R(C)        // Modulo

ADDF   R(A) = R(B) + R(C)        // Float-specific add
SUBF   R(A) = R(B) - R(C)        // Float-specific subtract
MULF   R(A) = R(B) * R(C)        // Float-specific multiply
DIVF   R(A) = R(B) / R(C)        // Float-specific divide
```

#### Comparison (3-register format)
```
LT     R(A) = R(B) < R(C)
LE     R(A) = R(B) <= R(C)
GT     R(A) = R(B) > R(C)
GE     R(A) = R(B) >= R(C)
EQ     R(A) = R(B) == R(C)
NE     R(A) = R(B) != R(C)
```

#### Logical (3-register format)
```
AND    R(A) = R(B) && R(C)
OR     R(A) = R(B) || R(C)
NOT    R(A) = !R(B)
```

#### Memory Operations
```
LOADK     R(A) = K(Bx)              // Load constant
LOADNIL   R(A) = nil                // Load nil
LOADBOOL  R(A) = bool(B)            // Load boolean

MOVE      R(A) = R(B)               // Copy register
```

#### Array/Map Operations
```
NEWARR    R(A) = []                 // New array
NEWMAP    R(A) = {}                 // New map
GETIDX    R(A) = R(B)[R(C)]         // Array/map index
SETIDX    R(A)[R(B)] = R(C)         // Array/map assignment
```

#### Struct Operations
```
NEWSTRUCT R(A) = new_struct(type)   // Create struct
GETFIELD  R(A) = R(B).field(C)      // Get struct field
SETFIELD  R(A).field(B) = R(C)      // Set struct field
```

#### Function Calls
```
CALL      R(A) = R(B)(R(C)...R(C+n))  // Call function
RETURN    return R(A)...R(A+n)        // Return values
CLOSURE   R(A) = closure(proto)       // Create closure
```

#### Control Flow
```
JMP       PC += offset              // Unconditional jump
JMPT      if R(A) then PC += offset // Jump if true
JMPF      if !R(A) then PC += offset // Jump if false
```

#### Built-ins
```
BUILTIN   R(A) = builtin[B](R(C)...R(C+n))
```

## Register Allocation Strategy

### Local Variables
All local variables are allocated to registers in function scope:

```javascript
func example(a: int, b: int): int {
    var x: int = a + b      // R0=a, R1=b, R2=x
    var y: int = x * 2      // R3=y
    return x + y            // result in R4
}

Compiled to:
  R0 = param a
  R1 = param b
  ADD   R2, R0, R1    // x = a + b
  LOADK R3, K0        // load constant 2
  MUL   R3, R2, R3    // y = x * 2
  ADD   R4, R2, R3    // result = x + y
  RETURN R4
```

### Register Allocation Algorithm

**Linear Scan Register Allocation** (simple, fast, good for interpreter):

1. **Pass 1**: Calculate live ranges for each variable
   - Track first def and last use
   - Variables with non-overlapping ranges can share registers

2. **Pass 2**: Assign registers
   - Sort variables by live range start
   - Assign to lowest available register
   - No need for spilling (unlimited virtual registers)

3. **Pass 3**: Emit instructions with assigned registers

### Register Reuse

```javascript
func compute(): int {
    var a: int = 10    // R0
    var b: int = 20    // R1
    var c: int = a + b // R2
    // a and b no longer used, can reuse R0, R1
    var d: int = c * 2 // R0 (reused)
    var e: int = d + 3 // R1 (reused)
    return e
}
```

## Value Representation

Use the same tagged union approach as stack VM:

```go
// Value remains the same - 16 bytes
type Value struct {
    Type ValueType
    Data uint64
}

// Register file is just an array of Values
type RegisterFile []Value
```

## Register VM Implementation

### Core Structure

```go
// vm/register_vm.go

type RegisterVM struct {
    constants    []Value            // Constant pool
    instructions []uint32           // Register bytecode
    registers    []Value            // Virtual registers (grows as needed)
    globals      map[string]Value   // Global variables
    pc           int                // Program counter

    // Function call stack
    frames       []*RegisterFrame
    frameIndex   int

    // Object pools (shared with stack VM)
    stringPool   map[string]*StringObject
    arrayPool    sync.Pool
    mapPool      sync.Pool
}

type RegisterFrame struct {
    function     *CompiledFunction
    baseRegister int                // Base of register window
    registers    []Value            // Local register window
    returnPC     int
}
```

### Execution Loop

```go
func (vm *RegisterVM) Run() error {
    for {
        instruction := vm.instructions[vm.pc]
        opcode := OpCode(instruction >> 24)

        // Decode operands
        a := int((instruction >> 16) & 0xFF)
        b := int((instruction >> 8) & 0xFF)
        c := int(instruction & 0xFF)

        switch opcode {
        case OpRAdd:
            vm.registers[a] = vm.add(vm.registers[b], vm.registers[c])

        case OpRSub:
            vm.registers[a] = vm.sub(vm.registers[b], vm.registers[c])

        case OpRMul:
            vm.registers[a] = vm.mul(vm.registers[b], vm.registers[c])

        case OpRLoadK:
            bx := int(instruction & 0xFFFF)
            vm.registers[a] = vm.constants[bx]

        case OpRCall:
            // Function call handling
            vm.callFunction(a, b, c)

        case OpRReturn:
            if vm.frameIndex == 0 {
                return nil // Program finished
            }
            vm.returnFromFunction(a)

        // ... more opcodes
        }

        vm.pc++
    }
}
```

## Compiler Modifications

### New Compiler Pass

Add register allocation pass between AST and bytecode emission:

```go
// compiler/register_compiler.go

type RegisterCompiler struct {
    *Compiler  // Embed existing compiler

    registers      map[string]int    // Variable -> register mapping
    nextRegister   int               // Next available register
    maxRegisters   int               // Max registers used
    liveRanges     map[string]*Range // Variable live ranges
}

func (rc *RegisterCompiler) CompileToRegister(node ast.Node) error {
    // Pass 1: Calculate live ranges
    rc.calculateLiveRanges(node)

    // Pass 2: Allocate registers
    rc.allocateRegisters()

    // Pass 3: Emit register instructions
    rc.emit(node)

    return nil
}
```

### Example Compilation

**Source:**
```javascript
func mandelbrot(cx: float, cy: float): int {
    var x: float = 0.0
    var y: float = 0.0
    var iter: int = 0

    for iter < 100 {
        var x2: float = x * x
        var y2: float = y * y

        if x2 + y2 > 4.0 {
            return iter
        }

        var xtemp: float = x2 - y2 + cx
        y = 2.0 * x * y + cy
        x = xtemp

        iter = iter + 1
    }

    return 100
}
```

**Register Allocation:**
```
Parameters: R0=cx, R1=cy
Locals:     R2=x, R3=y, R4=iter
Temps:      R5=x2, R6=y2, R7=xtemp, R8=(temp for calculations)
```

**Register Bytecode:**
```
        LOADK    R2, K0         // x = 0.0
        LOADK    R3, K0         // y = 0.0
        LOADK    R4, K1         // iter = 0

loop:   LOADK    R8, K2         // load 100
        LT       R8, R4, R8     // iter < 100
        JMPF     R8, end

        MULF     R5, R2, R2     // x2 = x * x
        MULF     R6, R3, R3     // y2 = y * y

        ADDF     R8, R5, R6     // x2 + y2
        LOADK    R9, K3         // load 4.0
        GT       R8, R8, R9     // > 4.0
        JMPF     R8, continue
        RETURN   R4             // return iter

continue:
        SUBF     R7, R5, R6     // x2 - y2
        ADDF     R7, R7, R0     // + cx

        LOADK    R8, K4         // load 2.0
        MULF     R8, R8, R2     // 2.0 * x
        MULF     R8, R8, R3     // * y
        ADDF     R3, R8, R1     // y = ... + cy

        MOVE     R2, R7         // x = xtemp

        LOADK    R8, K1         // load 1
        ADD      R4, R4, R8     // iter = iter + 1

        JMP      loop

end:    LOADK    R4, K2         // load 100
        RETURN   R4
```

## Performance Optimizations

### 1. Type-Specific Instructions

Separate int and float operations to avoid type checking:
```go
case OpRAdd:   // Generic add with type check
case OpRAddI:  // Integer-only add (no check)
case OpRAddF:  // Float-only add (no check)
```

### 2. Constant Folding

```javascript
var x: int = 2 + 3  // Compile to: LOADK R0, K5 (constant 5)
```

### 3. Peephole Optimization

```
MOVE R1, R0      Replace with:  ADD R2, R0, R3
ADD  R2, R1, R3
```

### 4. Register Windows for Calls

Use register windows to avoid copying on function calls:
```
Caller:  R0-R7 (uses R0-R3)
Callee:  R4-R11 (base = R4)
```

### 5. Inline Caching for Field Access

Cache struct field offsets:
```go
type FieldCache struct {
    structType *StructType
    offset     int
}
```

## Integration Plan

### Phase 1: Basic Register VM
- [x] Design instruction set
- [ ] Implement RegisterVM struct
- [ ] Implement core arithmetic/comparison operations
- [ ] Add constant loading and moves

### Phase 2: Control Flow
- [ ] Implement jumps and branches
- [ ] Add function calls and returns
- [ ] Handle closures

### Phase 3: Compiler Support
- [ ] Add RegisterCompiler
- [ ] Implement register allocation
- [ ] Emit register bytecode
- [ ] Add compiler flag: `--backend=register`

### Phase 4: Complex Operations
- [ ] Arrays and maps
- [ ] Structs
- [ ] Built-in functions

### Phase 5: Optimizations
- [ ] Type-specific instructions
- [ ] Constant folding
- [ ] Peephole optimization
- [ ] Inline caching

### Phase 6: Testing & Benchmarking
- [ ] Port all tests to register backend
- [ ] Run mandelbrot benchmark
- [ ] Compare performance with stack VM
- [ ] Validate 25-35% improvement

## Expected Performance Improvements

Based on Lua 5.0 results and our codebase analysis:

| Metric | Stack VM | Register VM | Improvement |
|--------|----------|-------------|-------------|
| Instructions/iteration | ~15 | ~8 | 47% fewer |
| Memory traffic | High | Low | ~40% less |
| Dispatch overhead | High | Medium | ~30% less |
| Overall speed | 7.7M iter/s | 10-11M iter/s | **30-40% faster** |

**Target**: Achieve **~10.5M iterations/sec** on Mandelbrot benchmark (vs current 7.7M)

This would put MinLang at **~45% of Python's speed** (up from 33%).

## Testing Strategy

```bash
# Compile with register backend
./minlang --backend=register examples/mandelbrot_heavy.min

# Compare backends
./minlang --backend=stack examples/test.min
./minlang --backend=register examples/test.min

# Benchmark comparison
time ./minlang --backend=stack examples/mandelbrot_heavy.min
time ./minlang --backend=register examples/mandelbrot_heavy.min
```

## Future Enhancements

1. **JIT Compilation**: Compile hot register bytecode to native code
2. **SIMD Instructions**: Add vector operations for arrays
3. **Escape Analysis**: Stack-allocate objects when possible
4. **Inline Expansion**: Inline small functions at call sites
5. **Speculative Optimization**: Type-specialize based on runtime profiling

## References

- Lua 5.0 Implementation: https://www.lua.org/doc/jucs05.pdf
- "Virtual Machine Showdown: Stack versus Registers" (ACM 2008)
- Crafting Interpreters - Register-based bytecode chapter
- Dalvik VM documentation (Android's register VM)

## Conclusion

A register-based VM offers significant performance improvements over the stack-based approach while maintaining the simplicity needed for an educational language. The design leverages unlimited virtual registers, type-specific operations, and efficient register allocation to achieve 30-40% better performance.
