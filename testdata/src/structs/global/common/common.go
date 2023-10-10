package common

var GlobalStruct = StructRO{}

type StructRO struct {
	Value int
	Ptr   *int
}
