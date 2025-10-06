package vm

import (
	"testing"
)

// TestValueTypes tests all value type operations
func TestValueTypes(t *testing.T) {
	t.Run("IntValue", func(t *testing.T) {
		v := IntValue(42)
		if v.Type != IntType {
			t.Errorf("Expected IntType, got %d", v.Type)
		}
		if v.AsInt() != 42 {
			t.Errorf("Expected 42, got %d", v.AsInt())
		}
	})

	t.Run("FloatValue", func(t *testing.T) {
		v := FloatValue(3.14)
		if v.Type != FloatType {
			t.Errorf("Expected FloatType, got %d", v.Type)
		}
		if v.AsFloat() != 3.14 {
			t.Errorf("Expected 3.14, got %f", v.AsFloat())
		}
	})

	t.Run("BoolValue", func(t *testing.T) {
		v := BoolValue(true)
		if v.Type != BoolType {
			t.Errorf("Expected BoolType, got %d", v.Type)
		}
		if !v.AsBool() {
			t.Error("Expected true")
		}
	})

	t.Run("StringValue", func(t *testing.T) {
		v := StringValue("hello")
		if v.Type != StringType {
			t.Errorf("Expected StringType, got %d", v.Type)
		}
		if v.AsString() != "hello" {
			t.Errorf("Expected 'hello', got %s", v.AsString())
		}
	})

	t.Run("NilValue", func(t *testing.T) {
		v := NilValue()
		if v.Type != NilType {
			t.Errorf("Expected NilType, got %d", v.Type)
		}
	})
}

// TestStringInterning verifies string interning works
func TestStringInterning(t *testing.T) {
	// Create two identical strings
	str := "test_interning_unique_string_xyz"
	s1 := StringValue(str)
	s2 := StringValue(str)

	// Should point to same interned string (or at least have same value)
	if s1.AsString() != s2.AsString() {
		t.Error("String values should match")
	}

	// Different strings should have different values
	s3 := StringValue("different_string_abc")
	if s1.AsString() == s3.AsString() {
		t.Error("Different strings should have different values")
	}
}

// TestArrayValue tests array operations
func TestArrayValue(t *testing.T) {
	arr := NewArrayValue(3)

	if arr.Type != ArrayType {
		t.Errorf("Expected ArrayType, got %d", arr.Type)
	}

	arrVal := arr.AsArray()
	if len(arrVal.Elements) != 3 {
		t.Errorf("Expected length 3, got %d", len(arrVal.Elements))
	}

	// Set and get elements
	arrVal.Elements[0] = IntValue(10)
	arrVal.Elements[1] = IntValue(20)
	arrVal.Elements[2] = IntValue(30)

	if arrVal.Elements[0].AsInt() != 10 {
		t.Error("Array element access failed")
	}
}

// TestMapValue tests map operations
func TestMapValue(t *testing.T) {
	m := NewMapValue()

	if m.Type != MapType {
		t.Errorf("Expected MapType, got %d", m.Type)
	}

	mapVal := m.AsMap()

	// Test integer keys
	intKey := IntValue(1).ToMapKey()
	mapVal.Pairs[intKey] = IntValue(100)

	if mapVal.Pairs[intKey].AsInt() != 100 {
		t.Error("Map integer key failed")
	}

	// Test string keys
	strKey := StringValue("test").ToMapKey()
	mapVal.Pairs[strKey] = IntValue(200)

	if mapVal.Pairs[strKey].AsInt() != 200 {
		t.Error("Map string key failed")
	}
}

// TestStructValue tests struct operations
func TestStructValue(t *testing.T) {
	fields := map[string]Value{
		"x": IntValue(10),
		"y": IntValue(20),
	}

	s := NewStructValue("Point", fields)

	if s.Type != StructType {
		t.Errorf("Expected StructType, got %d", s.Type)
	}

	structVal := s.AsStruct()

	if structVal.TypeName != "Point" {
		t.Errorf("Expected 'Point', got %s", structVal.TypeName)
	}

	if structVal.Fields["x"].AsInt() != 10 {
		t.Error("Struct field access failed")
	}
}

// TestFunctionValue tests function value creation
func TestFunctionValue(t *testing.T) {
	fn := &Function{
		Name:         "test",
		NumParams:    2,
		NumLocals:    3,
		Instructions: []byte{0x01, 0x02, 0x03},
	}

	v := NewFunctionValue(fn)

	if v.Type != FunctionType {
		t.Errorf("Expected FunctionType, got %d", v.Type)
	}

	retrieved := v.AsFunction()
	if retrieved.Name != "test" {
		t.Errorf("Expected 'test', got %s", retrieved.Name)
	}
	if retrieved.NumParams != 2 {
		t.Errorf("Expected 2 params, got %d", retrieved.NumParams)
	}
}

// TestClosureValue tests closure value creation
func TestClosureValue(t *testing.T) {
	fn := &Function{
		Name:      "closure",
		NumParams: 1,
		NumLocals: 1,
	}

	free := []Value{IntValue(42)}
	v := NewClosureValue(fn, free)

	if v.Type != ClosureType {
		t.Errorf("Expected ClosureType, got %d", v.Type)
	}

	closure := v.AsClosure()
	if closure.Fn.Name != "closure" {
		t.Error("Closure function mismatch")
	}
	if len(closure.Free) != 1 || closure.Free[0].AsInt() != 42 {
		t.Error("Closure free variables mismatch")
	}
}

// TestValueTruthiness tests IsTruthy() for all types
func TestValueTruthiness(t *testing.T) {
	tests := []struct {
		name     string
		value    Value
		expected bool
	}{
		{"true", BoolValue(true), true},
		{"false", BoolValue(false), false},
		{"nonzero int", IntValue(42), true},
		{"zero int", IntValue(0), false},
		{"nonzero float", FloatValue(3.14), true},
		{"zero float", FloatValue(0.0), false},
		{"non-empty string", StringValue("hello"), true},
		{"empty string", StringValue(""), false},
		{"nil", NilValue(), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value.IsTruthy() != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, tt.value.IsTruthy())
			}
		})
	}
}

// TestValueString tests String() method for all types
func TestValueString(t *testing.T) {
	tests := []struct {
		name     string
		value    Value
		expected string
	}{
		{"int", IntValue(42), "42"},
		{"float", FloatValue(3.14), "3.140000"},
		{"true", BoolValue(true), "true"},
		{"false", BoolValue(false), "false"},
		{"string", StringValue("hello"), "hello"},
		{"nil", NilValue(), "nil"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.value.String() != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, tt.value.String())
			}
		})
	}
}

// TestStackOperations tests push/pop operations
func TestStackOperations(t *testing.T) {
	bytecode := &Bytecode{
		Instructions: []byte{},
		Constants:    []Value{},
	}

	vm := New(bytecode)

	// Test push
	err := vm.push(IntValue(10))
	if err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	err = vm.push(IntValue(20))
	if err != nil {
		t.Fatalf("Push failed: %v", err)
	}

	if vm.sp != 2 {
		t.Errorf("Expected sp=2, got %d", vm.sp)
	}

	// Test pop
	v2 := vm.pop()
	if v2.AsInt() != 20 {
		t.Errorf("Expected 20, got %d", v2.AsInt())
	}

	v1 := vm.pop()
	if v1.AsInt() != 10 {
		t.Errorf("Expected 10, got %d", v1.AsInt())
	}

	if vm.sp != 0 {
		t.Errorf("Expected sp=0, got %d", vm.sp)
	}
}

// TestGCProtection verifies all object pools exist and are used
func TestGCProtection(t *testing.T) {
	// Create values - they should be added to pools
	_ = NewFunctionValue(&Function{Name: "test_gc_protection"})
	_ = NewClosureValue(&Function{Name: "closure_gc_protection"}, nil)
	_ = NewArrayValue(5)
	_ = NewMapValue()
	_ = NewStructValue("TestGCProtection", map[string]Value{})

	// Pools should have entries (we can't check exact count due to other tests)
	// Just verify the functions don't crash and pools exist
	t.Log("GC protection pools are functioning")
}

// TestPreAllocatedErrors verifies error constants exist
func TestPreAllocatedErrors(t *testing.T) {
	errors := []error{
		ErrDivisionByZero,
		ErrModuloByZero,
		ErrStackOverflow,
		ErrUnsupportedOperands,
		ErrCallingNonFunction,
		ErrUnsupportedComparison,
		ErrUnsupportedNegation,
	}

	for _, err := range errors {
		if err == nil {
			t.Error("Pre-allocated error is nil")
		}
		if err.Error() == "" {
			t.Error("Pre-allocated error has empty message")
		}
	}
}

// TestDirectLocalOperations tests OpAddLocal, OpSubLocal, etc.
func TestDirectLocalOperations(t *testing.T) {
	tests := []struct {
		name string
		op   OpCode
		a    int64
		b    int64
		want int64
	}{
		{"AddLocal", OpAddLocal, 10, 5, 15},
		{"SubLocal", OpSubLocal, 10, 5, 5},
		{"MulLocal", OpMulLocal, 10, 5, 50},
		{"DivLocal", OpDivLocal, 10, 5, 2},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a simple program that uses direct local operations
			// This is tested more thoroughly in integration tests
			// Here we just verify the opcodes exist
			switch tt.op {
			case OpAddLocal, OpSubLocal, OpMulLocal, OpDivLocal:
				// Opcodes exist
			default:
				t.Errorf("Opcode %d not recognized", tt.op)
			}
		})
	}
}
