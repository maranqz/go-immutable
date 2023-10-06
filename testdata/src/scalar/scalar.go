package function

func Bool() {
	readonlyRO := true

	// TODO target message showing which value was attempted to change
	// try to change readonlyRO" from testdata/src/scalar/scalar.go:4:5
	readonlyRO = false // want "try to change readonlyRO"

	tmp := readonlyRO
	tmp = tmp && readonlyRO

	tmpPtr := &readonlyRO
	// TODO tmpPtr should show readonlyRO
	*tmpPtr = tmp && readonlyRO // want "try to change tmpPtr"

	tmpPtr2 := tmpPtr
	// TODO tmpPtr2 should show readonlyRO
	*tmpPtr2 = tmp && readonlyRO // want "try to change tmpPtr2"

	value := *tmpPtr2
	value = false
	_ = value
}
