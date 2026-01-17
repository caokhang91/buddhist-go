package object

func blobNewBuiltin(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	sizeValue, errObj := intArg(args, 0, "blob_new")
	if errObj != nil {
		return errObj
	}
	if sizeValue < 0 {
		return newError("blob_new size must be non-negative")
	}
	maxInt := int64(^uint(0) >> 1)
	if sizeValue > maxInt {
		return newError("blob_new size too large: %d", sizeValue)
	}
	return NewBlob(int(sizeValue))
}

func blobFromStringBuiltin(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	value, errObj := stringArg(args, 0, "blob_from_string")
	if errObj != nil {
		return errObj
	}
	return NewBlobFromBytes([]byte(value))
}

func blobFromFileBuiltin(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	path, errObj := stringArg(args, 0, "blob_from_file")
	if errObj != nil {
		return errObj
	}
	blob, err := NewBlobFromFile(path)
	if err != nil {
		return newError("blob_from_file failed: %s", err.Error())
	}
	return blob
}

func blobWriteFileBuiltin(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments. got=%d, want=2", len(args))
	}
	blob, errObj := blobArg(args, 0, "blob_write_file")
	if errObj != nil {
		return errObj
	}
	path, errObj := stringArg(args, 1, "blob_write_file")
	if errObj != nil {
		return errObj
	}
	if err := blob.WriteToFile(path); err != nil {
		return newError("blob_write_file failed: %s", err.Error())
	}
	return &Null{}
}

func blobSliceBuiltin(args ...Object) Object {
	if len(args) != 3 {
		return newError("wrong number of arguments. got=%d, want=3", len(args))
	}
	blob, errObj := blobArg(args, 0, "blob_slice")
	if errObj != nil {
		return errObj
	}
	startValue, errObj := intArg(args, 1, "blob_slice")
	if errObj != nil {
		return errObj
	}
	endValue, errObj := intArg(args, 2, "blob_slice")
	if errObj != nil {
		return errObj
	}
	start, end, errObj := normalizeSliceBounds(len(blob.Data), startValue, endValue)
	if errObj != nil {
		return errObj
	}
	sub, err := blob.SubBlob(start, end)
	if err != nil {
		return newError("blob_slice failed: %s", err.Error())
	}
	return sub
}

func blobReadIntBuiltin(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments. got=%d, want=2", len(args))
	}
	blob, errObj := blobArg(args, 0, "blob_read_int")
	if errObj != nil {
		return errObj
	}
	offset, errObj := offsetArg(args, 1, "blob_read_int")
	if errObj != nil {
		return errObj
	}
	value, err := blob.ReadInt64(offset)
	if err != nil {
		return newError("blob_read_int failed: %s", err.Error())
	}
	return &Integer{Value: value}
}

func blobWriteIntBuiltin(args ...Object) Object {
	if len(args) != 3 {
		return newError("wrong number of arguments. got=%d, want=3", len(args))
	}
	blob, errObj := blobArg(args, 0, "blob_write_int")
	if errObj != nil {
		return errObj
	}
	offset, errObj := offsetArg(args, 1, "blob_write_int")
	if errObj != nil {
		return errObj
	}
	value, errObj := intArg(args, 2, "blob_write_int")
	if errObj != nil {
		return errObj
	}
	if err := blob.WriteInt64(offset, value); err != nil {
		return newError("blob_write_int failed: %s", err.Error())
	}
	return blob
}

func blobReadFloatBuiltin(args ...Object) Object {
	if len(args) != 2 {
		return newError("wrong number of arguments. got=%d, want=2", len(args))
	}
	blob, errObj := blobArg(args, 0, "blob_read_float")
	if errObj != nil {
		return errObj
	}
	offset, errObj := offsetArg(args, 1, "blob_read_float")
	if errObj != nil {
		return errObj
	}
	value, err := blob.ReadFloat64(offset)
	if err != nil {
		return newError("blob_read_float failed: %s", err.Error())
	}
	return &Float{Value: value}
}

func blobWriteFloatBuiltin(args ...Object) Object {
	if len(args) != 3 {
		return newError("wrong number of arguments. got=%d, want=3", len(args))
	}
	blob, errObj := blobArg(args, 0, "blob_write_float")
	if errObj != nil {
		return errObj
	}
	offset, errObj := offsetArg(args, 1, "blob_write_float")
	if errObj != nil {
		return errObj
	}
	value, errObj := floatArg(args, 2, "blob_write_float")
	if errObj != nil {
		return errObj
	}
	if err := blob.WriteFloat64(offset, value); err != nil {
		return newError("blob_write_float failed: %s", err.Error())
	}
	return blob
}

func blobMmapBuiltin(args ...Object) Object {
	if len(args) < 1 || len(args) > 2 {
		return newError("wrong number of arguments. got=%d, want=1 or 2", len(args))
	}
	path, errObj := stringArg(args, 0, "blob_mmap")
	if errObj != nil {
		return errObj
	}
	writable := false
	if len(args) == 2 {
		boolVal, errObj := boolArg(args, 1, "blob_mmap")
		if errObj != nil {
			return errObj
		}
		writable = boolVal
	}
	blob, err := MmapFile(path, writable)
	if err != nil {
		return newError("blob_mmap failed: %s", err.Error())
	}
	return blob
}

func blobUnmapBuiltin(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	blob, errObj := blobArg(args, 0, "blob_unmap")
	if errObj != nil {
		return errObj
	}
	if err := blob.Unmap(); err != nil {
		return newError("blob_unmap failed: %s", err.Error())
	}
	return &Null{}
}

func blobReleaseBuiltin(args ...Object) Object {
	if len(args) != 1 {
		return newError("wrong number of arguments. got=%d, want=1", len(args))
	}
	blob, errObj := blobArg(args, 0, "blob_release")
	if errObj != nil {
		return errObj
	}
	if err := blob.Release(); err != nil {
		return newError("blob_release failed: %s", err.Error())
	}
	return &Null{}
}

func blobArg(args []Object, index int, fnName string) (*Blob, *Error) {
	blob, ok := args[index].(*Blob)
	if !ok {
		return nil, newError("argument %d to `%s` must be BLOB, got %s", index+1, fnName, args[index].Type())
	}
	return blob, nil
}

func stringArg(args []Object, index int, fnName string) (string, *Error) {
	str, ok := args[index].(*String)
	if !ok {
		return "", newError("argument %d to `%s` must be STRING, got %s", index+1, fnName, args[index].Type())
	}
	return str.Value, nil
}

func intArg(args []Object, index int, fnName string) (int64, *Error) {
	value, ok := args[index].(*Integer)
	if !ok {
		return 0, newError("argument %d to `%s` must be INTEGER, got %s", index+1, fnName, args[index].Type())
	}
	return value.Value, nil
}

func offsetArg(args []Object, index int, fnName string) (int, *Error) {
	value, errObj := intArg(args, index, fnName)
	if errObj != nil {
		return 0, errObj
	}
	if value < 0 {
		return 0, newError("argument %d to `%s` must be non-negative", index+1, fnName)
	}
	maxInt := int64(^uint(0) >> 1)
	if value > maxInt {
		return 0, newError("argument %d to `%s` is too large", index+1, fnName)
	}
	return int(value), nil
}

func floatArg(args []Object, index int, fnName string) (float64, *Error) {
	switch value := args[index].(type) {
	case *Float:
		return value.Value, nil
	case *Integer:
		return float64(value.Value), nil
	default:
		return 0, newError("argument %d to `%s` must be FLOAT, got %s", index+1, fnName, args[index].Type())
	}
}

func boolArg(args []Object, index int, fnName string) (bool, *Error) {
	value, ok := args[index].(*Boolean)
	if !ok {
		return false, newError("argument %d to `%s` must be BOOLEAN, got %s", index+1, fnName, args[index].Type())
	}
	return value.Value, nil
}

func normalizeSliceBounds(length int, startValue, endValue int64) (int, int, *Error) {
	start, end := int(startValue), int(endValue)
	maxInt := int64(^uint(0) >> 1)
	if startValue < -maxInt || endValue < -maxInt || startValue > maxInt || endValue > maxInt {
		return 0, 0, newError("slice bounds out of range")
	}
	if start < 0 {
		start = length + start
	}
	if end < 0 {
		end = length + end
	}
	if start < 0 {
		start = 0
	}
	if end > length {
		end = length
	}
	if start > end {
		start = end
	}
	if start < 0 || end < 0 || start > length || end > length {
		return 0, 0, newError("slice bounds out of range")
	}
	return start, end, nil
}

