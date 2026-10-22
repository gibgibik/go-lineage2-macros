package service

import (
	"image"
	"sync"
	"time"

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
	AssistPartyMemberMap = map[string]image.Point{}

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
	// Bound serial reads so a silent/unplugged CH9329 device causes the
	// ch9329 client call to return an error instead of blocking forever
	// while holding Control.Mutex (which would freeze every control call
	// and stall the macro loop with no further log output).
	if err := port.SetReadTimeout(2 * time.Second); err != nil {
		_ = port.Close()
		return nil, err
	}
	control = &Control{
		cl: ch9329.NewClient(port, image.Rect(0, 0, cnf.Resolution[0], cnf.Resolution[1])),
	}
	return control, nil
}
