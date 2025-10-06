package compiler

import (
	"minlang/ast"
	"minlang/vm"
)

// convertToValueType converts compiler.Type to vm.ValueType
// This is used for type-specialized opcode generation
func convertToValueType(t Type) vm.ValueType {
	switch typ := t.(type) {
	case *BasicType:
		switch typ.Name {
		case "int":
			return vm.IntType
		case "float":
			return vm.FloatType
		case "bool":
			return vm.BoolType
		case "string":
			return vm.StringType
		default:
			// For struct types defined as BasicType with custom names
			// we don't know the exact type, so default to IntType
			return vm.IntType
		}
	case *ArrayType:
		return vm.ArrayType
	case *MapType:
		return vm.MapType
	case *FunctionType:
		return vm.FunctionType
	}
	// Default to IntType for unknown types
	return vm.IntType
}

// typeAnnotationToValueType converts AST TypeAnnotation to vm.ValueType
func typeAnnotationToValueType(ta *ast.TypeAnnotation) vm.ValueType {
	if ta == nil {
		return vm.IntType // Default
	}

	switch ta.Name {
	case "int":
		return vm.IntType
	case "float":
		return vm.FloatType
	case "bool":
		return vm.BoolType
	case "string":
		return vm.StringType
	}

	// Check if it's an array type
	if ta.ElementType != nil {
		return vm.ArrayType
	}

	// Check if it's a map type
	if ta.KeyType != nil && ta.ValueType != nil {
		return vm.MapType
	}

	// Default to IntType for unknown types (including struct types)
	return vm.IntType
}

// inferExpressionType determines the type of an expression at compile time
// This allows us to emit type-specialized opcodes
func (c *Compiler) inferExpressionType(node ast.Expression) vm.ValueType {
	switch n := node.(type) {
	case *ast.IntegerLiteral:
		return vm.IntType

	case *ast.FloatLiteral:
		return vm.FloatType

	case *ast.BooleanLiteral:
		return vm.BoolType

	case *ast.StringLiteral:
		return vm.StringType

	case *ast.Identifier:
		// Check if we have type information from our type tracking
		if t, ok := c.varTypes[n.Value]; ok {
			return t
		}

		// Check for enum values (they're integers)
		for _, enumType := range c.enumTypes {
			if _, ok := enumType.Variants[n.Value]; ok {
				return vm.IntType
			}
		}

		// Default: we don't know, return IntType as a safe default
		return vm.IntType

	case *ast.InfixExpression:
		return c.inferInfixType(n)

	case *ast.PrefixExpression:
		// -x has the same type as x
		// !x is always bool
		if n.Operator == "!" {
			return vm.BoolType
		}
		return c.inferExpressionType(n.Right)

	case *ast.CallExpression:
		// Function calls - we'd need to track return types
		// For now, default to int
		return vm.IntType

	case *ast.ArrayLiteral:
		return vm.ArrayType

	case *ast.MapLiteral:
		return vm.MapType

	case *ast.StructLiteral:
		return vm.StructType

	default:
		// Unknown type - default to int
		return vm.IntType
	}
}

// inferInfixType determines the result type of an infix expression
func (c *Compiler) inferInfixType(node *ast.InfixExpression) vm.ValueType {
	leftType := c.inferExpressionType(node.Left)
	rightType := c.inferExpressionType(node.Right)

	switch node.Operator {
	case "+":
		// String concatenation takes precedence
		if leftType == vm.StringType || rightType == vm.StringType {
			return vm.StringType
		}
		// Float promotion
		if leftType == vm.FloatType || rightType == vm.FloatType {
			return vm.FloatType
		}
		return vm.IntType

	case "-", "*", "/":
		// Float promotion
		if leftType == vm.FloatType || rightType == vm.FloatType {
			return vm.FloatType
		}
		return vm.IntType

	case "%":
		// Modulo is integer-only
		return vm.IntType

	case "==", "!=", "<", ">", "<=", ">=":
		// Comparisons always return bool
		return vm.BoolType

	case "&&", "||":
		// Logical operations return bool
		return vm.BoolType

	default:
		return vm.IntType
	}
}

// getOperandTypes returns the types of both operands in an infix expression
// This is used to determine which specialized opcode to emit
func (c *Compiler) getOperandTypes(node *ast.InfixExpression) (vm.ValueType, vm.ValueType) {
	leftType := c.inferExpressionType(node.Left)
	rightType := c.inferExpressionType(node.Right)
	return leftType, rightType
}
