package service

import (
	"image"
	"sync"

	"github.com/gibgibik/go-ch9329/pkg/ch9329"
	"github.com/gibgibik/go-lineage2-macros/internal/core"
	"go.bug.st/serial"
)

type Control struct {
	sync.Mutex
	cl *ch9329.Client
}

func (c *Control) SendKey(modifier byte, key string) (n int, err error) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	return c.cl.SendKey(modifier, key)
}

func (c *Control) MouseActionAbsolute(pressButton byte, point image.Point, wheel byte) (n int, err error) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	return c.cl.MouseActionAbsolute(pressButton, point, wheel)
}

func (c *Control) MouseAbsoluteEnd() (n int, err error) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	return c.cl.MouseAbsoluteEnd()
}
func (c *Control) EndKey() (n int, err error) {
	c.Mutex.Lock()
	defer c.Mutex.Unlock()
	return c.cl.EndKey()
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
		cl: ch9329.NewClient(port, image.Rect(0, 0, cnf.Resolution[0], cnf.Resolution[1])),
	}
	return control, nil
}
