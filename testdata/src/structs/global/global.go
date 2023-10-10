package global

import "structs/global/common"

func globalExportedVariable() {
	ro := common.GlobalStruct
	ro.Value++ // want "try to change ro"
	*ro.Ptr++  // want "try to change ro"

	roPtr := &ro
	roPtr.Value++ // want "try to change roPtr"
	*roPtr.Ptr++  // want "try to change roPtr"
	roPtr = nil
}

func globalExported() {
	ro := common.StructRO{}
	ro.Value++ // want "try to change ro"
	*ro.Ptr++  // want "try to change ro"

	roPtr := &common.StructRO{}
	roPtr.Value++ // want "try to change roPtr"
	*roPtr.Ptr++  // want "try to change roPtr"
	roPtr = nil
}
