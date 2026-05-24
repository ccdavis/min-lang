package vm

import (
	"bufio"
	"encoding/binary"
	"io"
	"os"
)

func fopenBuiltin(args ...Value) Value {
	if len(args) < 2 || len(args) > 3 {
		printError("fopen: wrong number of arguments. got=%d, want=2 or 3", len(args))
		return NilValue()
	}
	if args[0].Type != StringType {
		printError("fopen: first argument (path) must be string")
		return NilValue()
	}
	if args[1].Type != StringType {
		printError("fopen: second argument (mode) must be string")
		return NilValue()
	}

	path := args[0].AsString()
	mode := args[1].AsString()
	elemType := "string"
	if len(args) == 3 {
		if args[2].Type != StringType {
			printError("fopen: third argument (elemType) must be string")
			return NilValue()
		}
		elemType = args[2].AsString()
	}

	var recSize int
	switch elemType {
	case "string":
		recSize = 0
	case "byte":
		recSize = 1
	case "int":
		recSize = 8
	case "float":
		recSize = 8
	default:
		printError("fopen: unsupported element type '%s' (use string, byte, int, or float)", elemType)
		return NilValue()
	}

	var file *os.File
	var err error
	switch mode {
	case "r":
		file, err = os.Open(path)
	case "w":
		file, err = os.Create(path)
	case "a":
		file, err = os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	default:
		printError("fopen: invalid mode '%s' (use r, w, or a)", mode)
		return NilValue()
	}

	if err != nil {
		printError("fopen: %s", err.Error())
		return NilValue()
	}

	fh := &FileHandle{
		File:     file,
		Mode:     mode,
		ElemType: elemType,
		RecSize:  recSize,
	}
	if elemType == "string" && mode == "r" {
		fh.Scanner = bufio.NewScanner(file)
	}

	return NewFileHandleValue(fh)
}

func fcloseBuiltin(args ...Value) Value {
	if len(args) != 1 {
		printError("fclose: wrong number of arguments. got=%d, want=1", len(args))
		return NilValue()
	}
	if args[0].Type != FileType {
		printError("fclose: argument must be a file handle")
		return NilValue()
	}
	fh := args[0].AsFileHandle()
	if fh.File != nil {
		if err := fh.File.Close(); err != nil {
			printError("fclose: %s", err.Error())
		}
		fh.File = nil
	}
	return NilValue()
}

func freadBuiltin(args ...Value) Value {
	if len(args) != 1 {
		printError("fread: wrong number of arguments. got=%d, want=1", len(args))
		return NilValue()
	}
	if args[0].Type != FileType {
		printError("fread: argument must be a file handle")
		return NilValue()
	}
	fh := args[0].AsFileHandle()
	if fh.File == nil {
		printError("fread: file is closed")
		return NilValue()
	}
	if fh.Mode != "r" {
		printError("fread: file not opened for reading")
		return NilValue()
	}

	switch fh.ElemType {
	case "string":
		if fh.Scanner == nil {
			fh.AtEOF = true
			return NilValue()
		}
		if fh.Scanner.Scan() {
			return StringValue(fh.Scanner.Text())
		}
		fh.AtEOF = true
		if err := fh.Scanner.Err(); err != nil {
			printError("fread: %s", err.Error())
		}
		return NilValue()

	case "byte":
		var b [1]byte
		_, err := io.ReadFull(fh.File, b[:])
		if err != nil {
			fh.AtEOF = true
			if err != io.EOF && err != io.ErrUnexpectedEOF {
				printError("fread: %s", err.Error())
			}
			return NilValue()
		}
		return IntValue(int64(b[0]))

	case "int":
		var val int64
		err := binary.Read(fh.File, binary.LittleEndian, &val)
		if err != nil {
			fh.AtEOF = true
			if err != io.EOF && err != io.ErrUnexpectedEOF {
				printError("fread: %s", err.Error())
			}
			return NilValue()
		}
		return IntValue(val)

	case "float":
		var val float64
		err := binary.Read(fh.File, binary.LittleEndian, &val)
		if err != nil {
			fh.AtEOF = true
			if err != io.EOF && err != io.ErrUnexpectedEOF {
				printError("fread: %s", err.Error())
			}
			return NilValue()
		}
		return FloatValue(val)
	}

	return NilValue()
}

func fwriteBuiltin(args ...Value) Value {
	if len(args) != 2 {
		printError("fwrite: wrong number of arguments. got=%d, want=2", len(args))
		return NilValue()
	}
	if args[0].Type != FileType {
		printError("fwrite: first argument must be a file handle")
		return NilValue()
	}
	fh := args[0].AsFileHandle()
	if fh.File == nil {
		printError("fwrite: file is closed")
		return NilValue()
	}
	if fh.Mode != "w" && fh.Mode != "a" {
		printError("fwrite: file not opened for writing")
		return NilValue()
	}

	val := args[1]

	switch fh.ElemType {
	case "string":
		var s string
		if val.Type == StringType {
			s = val.AsString()
		} else {
			s = val.String()
		}
		if _, err := fh.File.WriteString(s); err != nil {
			printError("fwrite: %s", err.Error())
		}

	case "byte":
		if val.Type != IntType {
			printError("fwrite: byte file expects int value (0-255)")
			return NilValue()
		}
		if _, err := fh.File.Write([]byte{byte(val.AsInt())}); err != nil {
			printError("fwrite: %s", err.Error())
		}

	case "int":
		if val.Type != IntType {
			printError("fwrite: int file expects int value")
			return NilValue()
		}
		if err := binary.Write(fh.File, binary.LittleEndian, val.AsInt()); err != nil {
			printError("fwrite: %s", err.Error())
		}

	case "float":
		var f float64
		if val.Type == FloatType {
			f = val.AsFloat()
		} else if val.Type == IntType {
			f = float64(val.AsInt())
		} else {
			printError("fwrite: float file expects float or int value")
			return NilValue()
		}
		if err := binary.Write(fh.File, binary.LittleEndian, f); err != nil {
			printError("fwrite: %s", err.Error())
		}
	}

	return NilValue()
}

func fwritelnBuiltin(args ...Value) Value {
	if len(args) != 2 {
		printError("fwriteln: wrong number of arguments. got=%d, want=2", len(args))
		return NilValue()
	}
	if args[0].Type != FileType {
		printError("fwriteln: first argument must be a file handle")
		return NilValue()
	}
	fh := args[0].AsFileHandle()
	if fh.File == nil {
		printError("fwriteln: file is closed")
		return NilValue()
	}
	if fh.ElemType != "string" {
		printError("fwriteln: only supported for text files")
		return NilValue()
	}
	if fh.Mode != "w" && fh.Mode != "a" {
		printError("fwriteln: file not opened for writing")
		return NilValue()
	}

	val := args[1]
	var s string
	if val.Type == StringType {
		s = val.AsString()
	} else {
		s = val.String()
	}
	if _, err := fh.File.WriteString(s + "\n"); err != nil {
		printError("fwriteln: %s", err.Error())
	}
	return NilValue()
}

func feofBuiltin(args ...Value) Value {
	if len(args) != 1 {
		printError("feof: wrong number of arguments. got=%d, want=1", len(args))
		return NilValue()
	}
	if args[0].Type != FileType {
		printError("feof: argument must be a file handle")
		return NilValue()
	}
	fh := args[0].AsFileHandle()
	return BoolValue(fh.AtEOF)
}

func fseekBuiltin(args ...Value) Value {
	if len(args) != 2 {
		printError("fseek: wrong number of arguments. got=%d, want=2", len(args))
		return NilValue()
	}
	if args[0].Type != FileType {
		printError("fseek: first argument must be a file handle")
		return NilValue()
	}
	if args[1].Type != IntType {
		printError("fseek: second argument must be int")
		return NilValue()
	}
	fh := args[0].AsFileHandle()
	if fh.File == nil {
		printError("fseek: file is closed")
		return NilValue()
	}
	if fh.ElemType == "string" {
		printError("fseek: not supported for text files")
		return NilValue()
	}

	pos := args[1].AsInt()
	byteOffset := pos * int64(fh.RecSize)
	_, err := fh.File.Seek(byteOffset, io.SeekStart)
	if err != nil {
		printError("fseek: %s", err.Error())
		return NilValue()
	}
	fh.AtEOF = false
	return NilValue()
}

func ftellBuiltin(args ...Value) Value {
	if len(args) != 1 {
		printError("ftell: wrong number of arguments. got=%d, want=1", len(args))
		return NilValue()
	}
	if args[0].Type != FileType {
		printError("ftell: argument must be a file handle")
		return NilValue()
	}
	fh := args[0].AsFileHandle()
	if fh.File == nil {
		printError("ftell: file is closed")
		return NilValue()
	}
	if fh.ElemType == "string" {
		printError("ftell: not supported for text files")
		return NilValue()
	}

	bytePos, err := fh.File.Seek(0, io.SeekCurrent)
	if err != nil {
		printError("ftell: %s", err.Error())
		return NilValue()
	}
	return IntValue(bytePos / int64(fh.RecSize))
}

func fsizeBuiltin(args ...Value) Value {
	if len(args) != 1 {
		printError("fsize: wrong number of arguments. got=%d, want=1", len(args))
		return NilValue()
	}
	if args[0].Type != FileType {
		printError("fsize: argument must be a file handle")
		return NilValue()
	}
	fh := args[0].AsFileHandle()
	if fh.File == nil {
		printError("fsize: file is closed")
		return NilValue()
	}
	if fh.ElemType == "string" {
		printError("fsize: not supported for text files")
		return NilValue()
	}

	info, err := fh.File.Stat()
	if err != nil {
		printError("fsize: %s", err.Error())
		return NilValue()
	}
	return IntValue(info.Size() / int64(fh.RecSize))
}

func readFileBuiltin(args ...Value) Value {
	if len(args) != 1 {
		printError("readFile: wrong number of arguments. got=%d, want=1", len(args))
		return NilValue()
	}
	if args[0].Type != StringType {
		printError("readFile: argument must be string (file path)")
		return NilValue()
	}
	data, err := os.ReadFile(args[0].AsString())
	if err != nil {
		printError("readFile: %s", err.Error())
		return NilValue()
	}
	return StringValue(string(data))
}

func writeFileBuiltin(args ...Value) Value {
	if len(args) != 2 {
		printError("writeFile: wrong number of arguments. got=%d, want=2", len(args))
		return NilValue()
	}
	if args[0].Type != StringType {
		printError("writeFile: first argument must be string (file path)")
		return NilValue()
	}
	if args[1].Type != StringType {
		printError("writeFile: second argument must be string (content)")
		return NilValue()
	}
	err := os.WriteFile(args[0].AsString(), []byte(args[1].AsString()), 0644)
	if err != nil {
		printError("writeFile: %s", err.Error())
	}
	return NilValue()
}
