package global

import "github.com/maranqz/go-immutable/testdata/src/structs/local"

func globalExportedValue() {
	ro := local.GlobalStruct
	ro.Value++ // want "try to change roPtr"
	*ro.Ptr++  // want "try to change roPtr"
}
