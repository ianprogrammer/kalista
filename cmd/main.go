package main

import (
	"github.com/ianprogrammer/kalista/internal/kalista"
)

func main() {

	k := kalista.NewKalista("../contracts")
	k.StartContracTest()
}
