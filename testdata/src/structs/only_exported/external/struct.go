package external

import strct "structs/only_exported"

func changeValue() {
	readonlyStruct := strct.StructROExt{}

	readonlyStruct.Value++ // want "try to change readonlyStruct"

	readonlyStructCopy := readonlyStruct

	changeReadonlyValue(readonlyStructCopy)
	changeReadonlyValueInPtr(&readonlyStructCopy)
}

func changeReadonlyValue(in strct.StructROExt) {
	// TODO target message showing which value was attempted to change
	// try to change readonly variable from testdata/src/struct/external_struct.go:4:5
	in.Value++ // want "try to change in"

	cp := in.Value
	cp++
}

func changeReadonlyValueInPtr(in *strct.StructROExt) {
	in.Value++ // want "try to change in"

	cp := in.Value
	cp++
}

func changePtr() {
	readonlyStruct := strct.StructROExt{}
	readonlyStructCopy := readonlyStruct

	changeReadonlyPtr(readonlyStructCopy)
	changeReadonlyPtrInPtr(&readonlyStructCopy)
}

func changeReadonlyPtr(in strct.StructROExt) {
	cp := in.Ptr

	v := 1
	in.Ptr = &v // want "try to change in"
	*in.Ptr++   // want "try to change in"
	*in.Ptr = 1 // want "try to change in"

	cp = &v
	*cp++   // TODO skiped want "try to change cp"
	*cp = 1 // TODO skiped want "try to change cp"
}

func changeReadonlyPtrInPtr(in *strct.StructROExt) {
	cp := in.Ptr

	v := 1
	in.Ptr = &v // want "try to change in"

	*in.Ptr = 1 // want "try to change in"
	*in.Ptr++   // want "try to change in"

	cp = &v
	*cp = 1 // TODO skiped want "try to change cp"
	*cp++   // TODO skiped want "try to change cp"
}
