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

func structsPtr() {
	ro := &StructRO{}
	ro = &StructRO{} // want "try to change ro"

	roPtr := ro
	*roPtr = StructRO{} // want "try to change roPtr"
	roPtr = nil
}
