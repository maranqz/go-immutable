package strct

func changeValue() {
	readonlyStruct := StructRO{}

	readonlyStruct.Value++ // want "try to change readonlyStruct"

	changeReadonlyValue(readonlyStruct)
	changeReadonlyValueInPtr(&readonlyStruct)
}

func changeReadonlyValue(in StructRO) {
	in.Value++ // want "try to change in"
}

func changeReadonlyValueInPtr(in *StructRO) {
	in.Value++ // want "try to change in"
}
