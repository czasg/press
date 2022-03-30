package main

import (
	"github.com/czasg/press/cmd"
)

func main() {
	_ = cmd.InitCobraCmd().Execute()
}
