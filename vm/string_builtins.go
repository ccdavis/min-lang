package vm

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

func trimBuiltin(args ...Value) Value {
	if len(args) != 1 {
		printError("trim: wrong number of arguments. got=%d, want=1", len(args))
		return NilValue()
	}
	if args[0].Type != StringType {
		printError("trim: argument must be string")
		return NilValue()
	}
	return StringValue(strings.TrimSpace(args[0].AsString()))
}

func upperBuiltin(args ...Value) Value {
	if len(args) != 1 {
		printError("upper: wrong number of arguments. got=%d, want=1", len(args))
		return NilValue()
	}
	if args[0].Type != StringType {
		printError("upper: argument must be string")
		return NilValue()
	}
	return StringValue(strings.ToUpper(args[0].AsString()))
}

func lowerBuiltin(args ...Value) Value {
	if len(args) != 1 {
		printError("lower: wrong number of arguments. got=%d, want=1", len(args))
		return NilValue()
	}
	if args[0].Type != StringType {
		printError("lower: argument must be string")
		return NilValue()
	}
	return StringValue(strings.ToLower(args[0].AsString()))
}

func containsBuiltin(args ...Value) Value {
	if len(args) != 2 {
		printError("contains: wrong number of arguments. got=%d, want=2", len(args))
		return NilValue()
	}
	if args[0].Type != StringType {
		printError("contains: first argument must be string")
		return NilValue()
	}
	if args[1].Type != StringType {
		printError("contains: second argument must be string")
		return NilValue()
	}
	return BoolValue(strings.Contains(args[0].AsString(), args[1].AsString()))
}

func indexOfBuiltin(args ...Value) Value {
	if len(args) != 2 {
		printError("indexOf: wrong number of arguments. got=%d, want=2", len(args))
		return NilValue()
	}
	if args[0].Type != StringType {
		printError("indexOf: first argument must be string")
		return NilValue()
	}
	if args[1].Type != StringType {
		printError("indexOf: second argument must be string")
		return NilValue()
	}
	return IntValue(int64(strings.Index(args[0].AsString(), args[1].AsString())))
}

func replaceBuiltin(args ...Value) Value {
	if len(args) != 3 {
		printError("replace: wrong number of arguments. got=%d, want=3", len(args))
		return NilValue()
	}
	if args[0].Type != StringType {
		printError("replace: first argument must be string")
		return NilValue()
	}
	if args[1].Type != StringType {
		printError("replace: second argument must be string")
		return NilValue()
	}
	if args[2].Type != StringType {
		printError("replace: third argument must be string")
		return NilValue()
	}
	return StringValue(strings.ReplaceAll(args[0].AsString(), args[1].AsString(), args[2].AsString()))
}

func startsWithBuiltin(args ...Value) Value {
	if len(args) != 2 {
		printError("startsWith: wrong number of arguments. got=%d, want=2", len(args))
		return NilValue()
	}
	if args[0].Type != StringType {
		printError("startsWith: first argument must be string")
		return NilValue()
	}
	if args[1].Type != StringType {
		printError("startsWith: second argument must be string")
		return NilValue()
	}
	return BoolValue(strings.HasPrefix(args[0].AsString(), args[1].AsString()))
}

func endsWithBuiltin(args ...Value) Value {
	if len(args) != 2 {
		printError("endsWith: wrong number of arguments. got=%d, want=2", len(args))
		return NilValue()
	}
	if args[0].Type != StringType {
		printError("endsWith: first argument must be string")
		return NilValue()
	}
	if args[1].Type != StringType {
		printError("endsWith: second argument must be string")
		return NilValue()
	}
	return BoolValue(strings.HasSuffix(args[0].AsString(), args[1].AsString()))
}

func joinBuiltin(args ...Value) Value {
	if len(args) != 2 {
		printError("join: wrong number of arguments. got=%d, want=2", len(args))
		return NilValue()
	}
	if args[0].Type != ArrayType {
		printError("join: first argument must be array")
		return NilValue()
	}
	if args[1].Type != StringType {
		printError("join: second argument must be string")
		return NilValue()
	}

	arr := args[0].AsArray()
	sep := args[1].AsString()
	parts := make([]string, len(arr.Elements))
	for i, elem := range arr.Elements {
		parts[i] = elem.String()
	}
	return StringValue(strings.Join(parts, sep))
}

func repeatBuiltin(args ...Value) Value {
	if len(args) != 2 {
		printError("repeat: wrong number of arguments. got=%d, want=2", len(args))
		return NilValue()
	}
	if args[0].Type != StringType {
		printError("repeat: first argument must be string")
		return NilValue()
	}
	if args[1].Type != IntType {
		printError("repeat: second argument must be int")
		return NilValue()
	}
	count := int(args[1].AsInt())
	if count < 0 {
		count = 0
	}
	return StringValue(strings.Repeat(args[0].AsString(), count))
}

func charBuiltin(args ...Value) Value {
	if len(args) != 1 {
		printError("char: wrong number of arguments. got=%d, want=1", len(args))
		return NilValue()
	}
	if args[0].Type != IntType {
		printError("char: argument must be int")
		return NilValue()
	}
	code := args[0].AsInt()
	if code < 0 || code > unicode.MaxRune {
		printError("char: code point %d out of range", code)
		return NilValue()
	}
	return StringValue(string(rune(code)))
}

func ordBuiltin(args ...Value) Value {
	if len(args) != 1 {
		printError("ord: wrong number of arguments. got=%d, want=1", len(args))
		return NilValue()
	}
	if args[0].Type != StringType {
		printError("ord: argument must be string")
		return NilValue()
	}
	s := args[0].AsString()
	if len(s) == 0 {
		printError("ord: string must not be empty")
		return NilValue()
	}
	r, _ := utf8.DecodeRuneInString(s)
	return IntValue(int64(r))
}
