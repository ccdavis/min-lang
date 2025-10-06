package vm

import (
	"fmt"
	"math"
	"unsafe"
)

// String pool to keep strings alive for GC
// The GC doesn't track uint64, so we need to keep references to heap-allocated data
// No locking needed - VM is single-threaded
var stringPool []*string

// ValueType represents the type of a value
type ValueType byte

const (
	IntType ValueType = iota
	FloatType
	BoolType
	StringType
	ArrayType
	MapType
	StructType
	FunctionType
	ClosureType
	BuiltinFunctionType
	NilType
)

// Value represents a runtime value in the VM
// Uses a tagged union to avoid interface{} boxing overhead
type Value struct {
	Type ValueType
	_    [7]byte // Explicit padding for 8-byte alignment
	Data uint64  // Union: holds int64, float64, bool, or pointer
}

// Integer values
func IntValue(i int64) Value {
	return Value{Type: IntType, Data: uint64(i)}
}

func (v Value) AsInt() int64 {
	return int64(v.Data)
}

// Float values
func FloatValue(f float64) Value {
	return Value{Type: FloatType, Data: math.Float64bits(f)}
}

func (v Value) AsFloat() float64 {
	return math.Float64frombits(v.Data)
}

// Boolean values
func BoolValue(b bool) Value {
	var data uint64
	if b {
		data = 1
	}
	return Value{Type: BoolType, Data: data}
}

func (v Value) AsBool() bool {
	return v.Data != 0
}

// String values
func StringValue(s string) Value {
	// Allocate string on heap and add to pool to keep it alive for GC
	ptr := new(string)
	*ptr = s

	// Add to pool so GC doesn't collect it (no locking needed - VM is single-threaded)
	stringPool = append(stringPool, ptr)

	return Value{Type: StringType, Data: uint64(uintptr(unsafe.Pointer(ptr)))}
}

func (v Value) AsString() string {
	ptr := (*string)(unsafe.Pointer(uintptr(v.Data)))
	return *ptr
}

// Nil value
func NilValue() Value {
	return Value{Type: NilType, Data: 0}
}

// IsTruthy returns whether a value is considered true
func (v Value) IsTruthy() bool {
	switch v.Type {
	case BoolType:
		return v.AsBool()
	case IntType:
		return v.AsInt() != 0
	case FloatType:
		return v.AsFloat() != 0
	case StringType:
		return v.AsString() != ""
	case NilType:
		return false
	default:
		return true
	}
}

// String returns a string representation of the value
func (v Value) String() string {
	switch v.Type {
	case IntType:
		return fmt.Sprintf("%d", v.AsInt())
	case FloatType:
		return fmt.Sprintf("%f", v.AsFloat())
	case BoolType:
		return fmt.Sprintf("%t", v.AsBool())
	case StringType:
		return v.AsString()
	case NilType:
		return "nil"
	case ArrayType:
		return fmt.Sprintf("%v", v.AsArray())
	case MapType:
		return fmt.Sprintf("%v", v.AsMap())
	case StructType:
		return fmt.Sprintf("%v", v.AsStruct())
	case FunctionType:
		return "<function>"
	case ClosureType:
		return "<closure>"
	case BuiltinFunctionType:
		return "<builtin>"
	default:
		return "<unknown>"
	}
}

// ArrayValue represents an array
type ArrayValue struct {
	Elements []Value
}

func NewArrayValue(size int) Value {
	arr := &ArrayValue{Elements: make([]Value, size)}
	return Value{
		Type: ArrayType,
		Data: uint64(uintptr(unsafe.Pointer(arr))),
	}
}

func (v Value) AsArray() *ArrayValue {
	return (*ArrayValue)(unsafe.Pointer(uintptr(v.Data)))
}

// MapKey represents a map key that can be int or string without allocation
type MapKey struct {
	IsInt bool
	IntVal int64
	StrVal string
}

// MapValue represents a map
type MapValue struct {
	Pairs map[MapKey]Value
}

func NewMapValue() Value {
	m := &MapValue{Pairs: make(map[MapKey]Value)}
	return Value{
		Type: MapType,
		Data: uint64(uintptr(unsafe.Pointer(m))),
	}
}

func (v Value) AsMap() *MapValue {
	return (*MapValue)(unsafe.Pointer(uintptr(v.Data)))
}

// ToMapKey converts a Value to a MapKey without allocation for ints
func (v Value) ToMapKey() MapKey {
	if v.Type == IntType {
		return MapKey{IsInt: true, IntVal: v.AsInt()}
	}
	return MapKey{IsInt: false, StrVal: v.String()}
}

// StructValue represents a struct instance
type StructValue struct {
	TypeName string
	Fields   map[string]Value
}

func NewStructValue(typeName string, fields map[string]Value) Value {
	s := &StructValue{
		TypeName: typeName,
		Fields:   fields,
	}
	return Value{
		Type: StructType,
		Data: uint64(uintptr(unsafe.Pointer(s))),
	}
}

func (v Value) AsStruct() *StructValue {
	return (*StructValue)(unsafe.Pointer(uintptr(v.Data)))
}

// Function represents a compiled function
type Function struct {
	Name          string
	NumParams     int
	NumLocals     int
	Instructions  []byte
	Constants     []Value
}

func NewFunctionValue(fn *Function) Value {
	return Value{Type: FunctionType, Data: uint64(uintptr(unsafe.Pointer(fn)))}
}

func (v Value) AsFunction() *Function {
	return (*Function)(unsafe.Pointer(uintptr(v.Data)))
}

// Closure represents a closure (function + captured variables)
type Closure struct {
	Fn   *Function
	Free []Value
}

func NewClosureValue(fn *Function, free []Value) Value {
	cl := &Closure{Fn: fn, Free: free}
	return Value{
		Type: ClosureType,
		Data: uint64(uintptr(unsafe.Pointer(cl))),
	}
}

func (v Value) AsClosure() *Closure {
	return (*Closure)(unsafe.Pointer(uintptr(v.Data)))
}

// AsBuiltinFunction extracts a builtin function from a Value
// Note: BuiltinFunction is defined in builtins.go
func (v Value) AsBuiltinFunction() func(args ...Value) Value {
	fnPtr := (*func(args ...Value) Value)(unsafe.Pointer(uintptr(v.Data)))
	return *fnPtr
}
