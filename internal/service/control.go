package service

import (
	"image"

	"github.com/gibgibik/go-ch9329/pkg/ch9329"
	"github.com/gibgibik/go-lineage2-macros/internal/core"
	"go.bug.st/serial"
)

type Control struct {
	Cl *ch9329.Client
}

var (
	AssistPartyMemberMap = map[string]image.Point{
		"1": {40, 147},
		"2": {40, 201},
		"3": {40, 255},
		"4": {40, 309},
		"5": {40, 364},
		"6": {40, 417},
		"7": {40, 472},
		"8": {40, 525},
	}

	control *Control
)

func GetControl(cnf core.Control) (*Control, error) {
	if control != nil {
		return control, nil
	}
	mode := &serial.Mode{
		BaudRate: cnf.BaudRate,
	}
	port, err := serial.Open(cnf.Port, mode)
	if err != nil {
		return nil, err
	}
	control = &Control{
		Cl: ch9329.NewClient(port, image.Rect(0, 0, cnf.Resolution[0], cnf.Resolution[1])),
	}
	return control, nil
}
