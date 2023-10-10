package local

type StructRO struct {
	Value int
	Ptr   *int
}

func structsValue() {
	ro := StructRO{}
	ro = StructRO{} // want "try to change ro"

	roPtr := &ro
	*roPtr = StructRO{} // want "try to change roPtr"
	roPtr = nil
}

// TODO check what is readonly struct and what is RO variable
func structsPtr() {
	ro := &StructRO{}
	ro = &StructRO{}

	roPtr := ro
	*roPtr = StructRO{} // want "try to change roPtr"
	roPtr = nil
}
