package main

import (
	"fmt"

	"fx.prodigy9.co/config"
)

var (
	Name1 = config.StrDef("NAME1", "Default1")
	Name2 = config.StrDef("NAME2", "Default2")
	Name3 = config.StrDef("NAME3", "Default3")
	Name4 = config.StrDef("NAME4", "Default4")
	Name5 = config.StrDef("NAME5", "Default5")
	Name6 = config.StrDef("NAME6", "Default6")
)

func main() {
	src := config.Configure()
	config.Set(src, Name1, "Forced")

	fmt.Println(config.Get(src, Name1))
	fmt.Println(config.Get(src, Name2))
	fmt.Println(config.Get(src, Name3))
	fmt.Println(config.Get(src, Name4))
	fmt.Println(config.Get(src, Name5))
	fmt.Println(config.Get(src, Name6))
}
