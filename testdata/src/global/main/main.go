package main

import (
	"global"
	alias "global/aliasnested"
	"global/nested"
	nested2 "global/nested2"
)

func main() {
	localPtr := &nested.GlobalRO
	*localPtr++ // want "try to change localPtr"

	nested.GlobalRO++  // want "try to change GlobalRO"
	alias.GlobalRO++   // want "try to change GlobalRO"
	global.GlobalRO++  // want "try to change GlobalRO"
	*global.Global++   // want "try to change Global"
	nested2.GlobalRO++ // want "try to change GlobalRO"
}
