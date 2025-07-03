package main

import (
	"github.com/gibgibik/go-lineage2-macros/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		panic(err)
	}
	return
}
