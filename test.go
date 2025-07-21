package main

import (
	"github.com/gibgibik/go-lineage2-macros/internal/core"
	"github.com/gibgibik/go-lineage2-macros/internal/service"
)

func main() {
	cnf, _ := core.InitConfig()
	controlCL, _ := service.GetControl(cnf.Control)
	controlCL.SendKey(0, "a")
	controlCL.EndKey()
}
