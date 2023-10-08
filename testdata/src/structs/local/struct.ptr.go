package local

func changePtr() {
	readonlyStruct := StructRO{}

	*readonlyStruct.Ptr++ // want "try to change readonlyStruct"

	changeReadonlyPtr(readonlyStruct)
	changeReadonlyPtrInPtr(&readonlyStruct)
}

func changeReadonlyPtr(in StructRO) {
	v := 1
	in.Ptr = &v // want "try to change in"

	*in.Ptr++ // want "try to change in"

	*in.Ptr = 1 // want "try to change in"
}

func changeReadonlyPtrInPtr(in *StructRO) {
	v := 1
	in.Ptr = &v // want "try to change in"

	*in.Ptr++ // want "try to change in"

	*in.Ptr = 1 // want "try to change in"
}
