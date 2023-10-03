package mp_and_array

var mp = map[string]int{}
var mpCopy = mp

func MapGet() {
	get := mp["0"]
	_ = get

	getFromCopy := mpCopy["0"]
	_ = getFromCopy
}

func mpChangeIndex() {
	mp["0"] = 1     // want "try to change readonly"
	mpCopy["0"] = 1 // want "try to change readonly"
}

func mpDelete() {
	delete(mp, "2")     // want "try to change readonly"
	delete(mpCopy, "2") // want "try to change readonly"
}

func mpReset() {
	mp = make(map[string]int) // want "try to change readonly"
	mpCopy = map[string]int{} // want "try to change readonly"
}

func mpReference() {
	ref := &mp

	get := (*ref)["0"]
	_ = get

	*ref = map[string]int{} // want "try to change readonly"
	(*ref)["0"] = 1         // want "try to change readonly"

	ref = nil
}

func mpLocalVariable() {
	mp := map[string]int{}
	mp["0"] = 1

	mpCopy := map[string]int{"1": 3}
	mpCopy["1"] = 2
}
