package vm

import (
	"testing"
)

func TestIntegerArithmetic(t *testing.T) {
	tests := []struct {
		input    *Bytecode
		expected interface{}
	}{
		{
			// 1 + 2
			&Bytecode{
				Instructions: concatInstructions(
					Make(OpPush, 0),
					Make(OpPush, 1),
					Make(OpAdd),
					Make(OpPop),
					Make(OpHalt),
				),
				Constants: []Value{IntValue(1), IntValue(2)},
			},
			int64(3),
		},
		{
			// 5 - 3
			&Bytecode{
				Instructions: concatInstructions(
					Make(OpPush, 0),
					Make(OpPush, 1),
					Make(OpSub),
					Make(OpPop),
					Make(OpHalt),
				),
				Constants: []Value{IntValue(5), IntValue(3)},
			},
			int64(2),
		},
		{
			// 4 * 5
			&Bytecode{
				Instructions: concatInstructions(
					Make(OpPush, 0),
					Make(OpPush, 1),
					Make(OpMul),
					Make(OpPop),
					Make(OpHalt),
				),
				Constants: []Value{IntValue(4), IntValue(5)},
			},
			int64(20),
		},
		{
			// 20 / 4
			&Bytecode{
				Instructions: concatInstructions(
					Make(OpPush, 0),
					Make(OpPush, 1),
					Make(OpDiv),
					Make(OpPop),
					Make(OpHalt),
				),
				Constants: []Value{IntValue(20), IntValue(4)},
			},
			int64(5),
		},
		{
			// 10 % 3
			&Bytecode{
				Instructions: concatInstructions(
					Make(OpPush, 0),
					Make(OpPush, 1),
					Make(OpMod),
					Make(OpPop),
					Make(OpHalt),
				),
				Constants: []Value{IntValue(10), IntValue(3)},
			},
			int64(1),
		},
		{
			// 2 + 3 * 4 (evaluated as (2 + 3) * 4 in postfix)
			&Bytecode{
				Instructions: concatInstructions(
					Make(OpPush, 0), // 2
					Make(OpPush, 1), // 3
					Make(OpAdd),     // 5
					Make(OpPush, 2), // 4
					Make(OpMul),     // 20
					Make(OpPop),
					Make(OpHalt),
				),
				Constants: []Value{IntValue(2), IntValue(3), IntValue(4)},
			},
			int64(20),
		},
		{
			// -5
			&Bytecode{
				Instructions: concatInstructions(
					Make(OpPush, 0),
					Make(OpNeg),
					Make(OpPop),
					Make(OpHalt),
				),
				Constants: []Value{IntValue(5)},
			},
			int64(-5),
		},
	}

	for _, tt := range tests {
		vm := New(tt.input)
		err := vm.Run()
		if err != nil {
			t.Fatalf("vm error: %s", err)
		}

		stackElem := vm.LastPoppedStackElem()
		testIntegerObject(t, tt.expected.(int64), stackElem)
	}
}

func TestBooleanExpressions(t *testing.T) {
	tests := []struct {
		input    *Bytecode
		expected bool
	}{
		{
			// 1 < 2
			&Bytecode{
				Instructions: concatInstructions(
					Make(OpPush, 0),
					Make(OpPush, 1),
					Make(OpLt),
					Make(OpPop),
					Make(OpHalt),
				),
				Constants: []Value{IntValue(1), IntValue(2)},
			},
			true,
		},
		{
			// 1 > 2
			&Bytecode{
				Instructions: concatInstructions(
					Make(OpPush, 0),
					Make(OpPush, 1),
					Make(OpGt),
					Make(OpPop),
					Make(OpHalt),
				),
				Constants: []Value{IntValue(1), IntValue(2)},
			},
			false,
		},
		{
			// 1 == 1
			&Bytecode{
				Instructions: concatInstructions(
					Make(OpPush, 0),
					Make(OpPush, 0),
					Make(OpEq),
					Make(OpPop),
					Make(OpHalt),
				),
				Constants: []Value{IntValue(1)},
			},
			true,
		},
		{
			// 1 != 2
			&Bytecode{
				Instructions: concatInstructions(
					Make(OpPush, 0),
					Make(OpPush, 1),
					Make(OpNe),
					Make(OpPop),
					Make(OpHalt),
				),
				Constants: []Value{IntValue(1), IntValue(2)},
			},
			true,
		},
		{
			// !true
			&Bytecode{
				Instructions: concatInstructions(
					Make(OpPush, 0),
					Make(OpNot),
					Make(OpPop),
					Make(OpHalt),
				),
				Constants: []Value{BoolValue(true)},
			},
			false,
		},
	}

	for _, tt := range tests {
		vm := New(tt.input)
		err := vm.Run()
		if err != nil {
			t.Fatalf("vm error: %s", err)
		}

		stackElem := vm.LastPoppedStackElem()
		testBooleanObject(t, tt.expected, stackElem)
	}
}

func TestConditionals(t *testing.T) {
	tests := []struct {
		input    *Bytecode
		expected interface{}
	}{
		{
			// if (true) { 10 } else { 20 }
			&Bytecode{
				Instructions: concatInstructions(
					Make(OpPush, 0),        // 0-2: true
					Make(OpJumpIfFalse, 12), // 3-5: jump to 12 if false (else branch)
					Make(OpPush, 1),        // 6-8: 10
					Make(OpJump, 15),       // 9-11: jump to 15 (after else)
					Make(OpPush, 2),        // 12-14: 20
					Make(OpPop),            // 15
					Make(OpHalt),           // 16
				),
				Constants: []Value{BoolValue(true), IntValue(10), IntValue(20)},
			},
			int64(10),
		},
		{
			// if (false) { 10 } else { 20 }
			&Bytecode{
				Instructions: concatInstructions(
					Make(OpPush, 0),        // 0-2: false
					Make(OpJumpIfFalse, 12), // 3-5: jump to 12 if false (else branch)
					Make(OpPush, 1),        // 6-8: 10
					Make(OpJump, 15),       // 9-11: jump to 15 (after else)
					Make(OpPush, 2),        // 12-14: 20
					Make(OpPop),            // 15
					Make(OpHalt),           // 16
				),
				Constants: []Value{BoolValue(false), IntValue(10), IntValue(20)},
			},
			int64(20),
		},
	}

	for _, tt := range tests {
		vm := New(tt.input)
		err := vm.Run()
		if err != nil {
			t.Fatalf("vm error: %s", err)
		}

		stackElem := vm.LastPoppedStackElem()
		testIntegerObject(t, tt.expected.(int64), stackElem)
	}
}

func TestGlobalVariables(t *testing.T) {
	tests := []struct {
		input    *Bytecode
		expected interface{}
	}{
		{
			// var x = 5; x;
			&Bytecode{
				Instructions: concatInstructions(
					Make(OpPush, 0),
					Make(OpStoreGlobal, 0),
					Make(OpLoadGlobal, 0),
					Make(OpPop),
					Make(OpHalt),
				),
				Constants: []Value{IntValue(5)},
			},
			int64(5),
		},
		{
			// var x = 5; var y = 10; x + y;
			&Bytecode{
				Instructions: concatInstructions(
					Make(OpPush, 0),
					Make(OpStoreGlobal, 0),
					Make(OpPush, 1),
					Make(OpStoreGlobal, 1),
					Make(OpLoadGlobal, 0),
					Make(OpLoadGlobal, 1),
					Make(OpAdd),
					Make(OpPop),
					Make(OpHalt),
				),
				Constants: []Value{IntValue(5), IntValue(10)},
			},
			int64(15),
		},
	}

	for _, tt := range tests {
		vm := New(tt.input)
		err := vm.Run()
		if err != nil {
			t.Fatalf("vm error: %s", err)
		}

		stackElem := vm.LastPoppedStackElem()
		testIntegerObject(t, tt.expected.(int64), stackElem)
	}
}

func testIntegerObject(t *testing.T, expected int64, actual Value) {
	if actual.Type != IntType {
		t.Fatalf("object type is not Integer. got=%T", actual)
	}

	if actual.AsInt() != expected {
		t.Fatalf("object has wrong value. got=%d, want=%d",
			actual.AsInt(), expected)
	}
}

func testBooleanObject(t *testing.T, expected bool, actual Value) {
	if actual.Type != BoolType {
		t.Fatalf("object type is not Boolean. got=%T", actual)
	}

	if actual.AsBool() != expected {
		t.Fatalf("object has wrong value. got=%t, want=%t",
			actual.AsBool(), expected)
	}
}

func concatInstructions(s ...[]byte) []byte {
	out := []byte{}
	for _, ins := range s {
		out = append(out, ins...)
	}
	return out
}
