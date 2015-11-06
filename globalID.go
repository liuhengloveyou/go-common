package common

import (
	"fmt"
	"time"
)

type GlobalID struct {
	Type   string
	Expand string
	moment string

	hole chan string
}

func (p *GlobalID) ID() string {
	if p.moment == "" {
		p.moment = fmt.Sprintf("%v", time.Now().Unix())
		p.hole = make(chan string)

		go func() {
			var serial int64 = 0

			for {
				p.hole <- fmt.Sprintf("%v%v%v%v", p.Type, p.moment, p.Expand, serial)
				serial += 1
			}
		}()
	}

	return <-p.hole
}
