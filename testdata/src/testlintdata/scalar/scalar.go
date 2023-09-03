package function

func Bool() {
	readonly := true

	// TODO target message showing which value was attempted to change
	readonly = false // want "try to change readonly from testdata/src/testlintdata/scalar/scalar.go:4:5"

	tmp := readonly

	tmp = false           // want "try to change readonly"
	tmp = tmp && readonly // want "try to change readonly"
}

func Int() {
	readonly := 1

	readonly = 2 // want "try to change readonly"

	tmp := readonly

	tmp++ // want "try to change readonly"
	tmp-- // want "try to change readonly"

	tmp += 1 // want "try to change readonly"
	tmp -= 1 // want "try to change readonly"

	tmp |= 1  // want "try to change readonly"
	tmp ^= 1  // want "try to change readonly"
	tmp &= 1  // want "try to change readonly"
	tmp <<= 1 // want "try to change readonly"
	tmp >>= 1 // want "try to change readonly"
}

func String() {
	readonly := "1"

	var tmp string

	tmp = readonly

	tmp = "1"            // want "try to change readonly"
	tmp = "1" + readonly // want "try to change readonly"

	_ = tmp[0]
}
