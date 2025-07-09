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

func NewControl(cnf core.Control) (*Control, error) {
	mode := &serial.Mode{
		BaudRate: cnf.BaudRate,
	}
	port, err := serial.Open(cnf.Port, mode)
	if err != nil {
		return nil, err
	}
	return &Control{
		Cl: ch9329.NewClient(port, image.Rect(0, 0, cnf.Resolution[0], cnf.Resolution[1])),
	}, nil
}
