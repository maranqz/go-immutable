package only_exported

// StructROExt is mutable structure inside the package but immutable when it is imported.
type StructROExt struct {
	Value int
	Ptr   *int
}

func extChangeValue() {
	extReadonlyStruct := StructROExt{}

	extReadonlyStruct = StructROExt{}

	extReadonlyStruct.Value++
	extReadonlyStruct.Ptr = &extReadonlyStruct.Value
	*extReadonlyStruct.Ptr++
}
