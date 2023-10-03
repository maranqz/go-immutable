package main

import alias "global/nested"

func main() {
	alias.GlobalRO++ // want "try to change GlobalRO"
}
