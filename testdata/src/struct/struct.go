package strct

type ReadonlyStruct struct {
	Value int
	Ptr   *int
}

func changeValue() {
	readonlyStruct := ReadonlyStruct{}

	changeReadonlyValue(readonlyStruct)
	changeReadonlyValueInPtr(&readonlyStruct)
}

func changeReadonlyValue(in ReadonlyStruct) {
	in.Value++ // want "try to change readonly"
}

func changeReadonlyValueInPtr(in *ReadonlyStruct) {
	in.Value++ // want "try to change readonly"
}

func changePtr() {
	readonlyStruct := ReadonlyStruct{}

	changeReadonlyPtr(readonlyStruct)
	changeReadonlyPtrInPtr(&readonlyStruct)
}

func changeReadonlyPtr(in ReadonlyStruct) {
	v := 1
	in.Ptr = &v // want "try to change readonly"

	*in.Ptr++ // want "try to change readonly"

	*in.Ptr = 1 // want "try to change readonly"
}

func changeReadonlyPtrInPtr(in *ReadonlyStruct) {
	v := 1
	in.Ptr = &v // want "try to change readonly"

	*in.Ptr++ // want "try to change readonly"

	*in.Ptr = 1 // want "try to change readonly"
}
