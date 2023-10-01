package external

import strct "github.com/maranqz/go-immutable/testdata/src/testlintdata/struct"

func changeValue() {
	readonlyStruct := strct.ExtReadonlyStruct{}
	readonlyStructCopy := readonlyStruct

	changeReadonlyValue(readonlyStructCopy)
	changeReadonlyValueInPtr(&readonlyStructCopy)
}

func changeReadonlyValue(in strct.ExtReadonlyStruct) {
	// TODO target message showing which value was attempted to change
	in.Value++ // want "try to change readonly variable from testdata/src/testlintdata/struct/external_struct.go:4:5"

	cp := in.Value
	cp++
}

func changeReadonlyValueInPtr(in *strct.ExtReadonlyStruct) {
	in.Value++ // want "try to change readonly"

	cp := in.Value
	cp++
}

func changePtr() {
	readonlyStruct := strct.ExtReadonlyStruct{}
	readonlyStructCopy := readonlyStruct

	changeReadonlyPtr(readonlyStructCopy)
	changeReadonlyPtrInPtr(&readonlyStructCopy)
}

func changeReadonlyPtr(in strct.ExtReadonlyStruct) {
	cp := in.Ptr

	v := 1
	in.Ptr = &v // want "try to change readonly"
	cp = &v

	*in.Ptr++ // want "try to change readonly"
	*cp++     // want "try to change readonly"

	*in.Ptr = 1 // want "try to change readonly"
	*cp = 1     // want "try to change readonly"
}

func changeReadonlyPtrInPtr(in *strct.ExtReadonlyStruct) {
	cp := in.Ptr

	v := 1
	in.Ptr = &v // want "try to change readonly"
	cp = &v

	*in.Ptr++ // want "try to change readonly"
	*cp++     // want "try to change readonly"

	*in.Ptr = 1 // want "try to change readonly"
	*cp = 1     // want "try to change readonly"
}
