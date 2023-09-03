package function

func main() {
	readonly := 1

	changeReadonly(&readonly)
}

func changeReadonly(in *int) {
	newIn := in

	*newIn++ // want "try to change readonly"
}

func changeReturnReadonly() {
	readonly := returnReadonly

	*readonly()++ // want "try to change readonly"
}

func returnReadonly() *int {
	readonly := 1

	return &readonly
}
