package strct

// ExtReadonlyStruct is mutable structure inside the package but immutable when it is imported.
type ExtReadonlyStruct struct {
	Value int
	Ptr   *int
}

func extChangeValue() {
	extReadonlyStruct := ExtReadonlyStruct{}

	extReadonlyStruct = ExtReadonlyStruct{}
	extReadonlyStruct.Value++
	extReadonlyStruct.Ptr = &extReadonlyStruct.Value
	*extReadonlyStruct.Ptr++
}
