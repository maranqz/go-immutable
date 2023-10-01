package slice_and_array

var slice = []string{"0", "1"}
var sliceCopy = slice

func SliceGet() {
	get := slice[0]
	_ = get

	getFromCopy := sliceCopy[0]
	_ = getFromCopy
}

func SliceChangeIndex() {
	slice[0] = "1"     // want "try to change readonly"
	sliceCopy[0] = "1" // want "try to change readonly"
}

func SliceAppend() {
	slice = append(slice, "2")         // want "try to change readonly"
	sliceCopy = append(sliceCopy, "2") // want "try to change readonly"
}

func SliceReset() {
	slice = []string{"0", "1"}     // want "try to change readonly"
	sliceCopy = []string{"0", "1"} // want "try to change readonly"
}

func SliceResize() {
	slice = slice[1:] // want "try to change readonly"
	slice = slice[:]  // want "try to change readonly"

	sliceCopy = sliceCopy[1:] // want "try to change readonly"
}

func SliceReference() {
	ref := &slice

	get := (*ref)[0]
	_ = get

	*ref = (*ref)[:] // want "try to change readonly"
	(*ref)[0] = "2"  // want "try to change readonly"

	ref = nil
}

func SliceLocalVariable() {
	slice := []string{"1", "2"}
	slice = append(slice, "3")

	sliceCopy := []string{"1", "2"}
	sliceCopy = append(sliceCopy, "3")
}
