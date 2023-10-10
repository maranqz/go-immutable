package local

var GlobalStruct = StructRO{}

func globalValue() {
	ro := GlobalStruct
	ro = StructRO{}
	ro.Value++ // want "try to change ro"

	roPtr := &GlobalStruct
	*roPtr = StructRO{} // want "try to change roPtr"
	roPtr = nil
}
