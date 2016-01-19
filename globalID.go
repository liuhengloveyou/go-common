package common

import (
	"fmt"
	"time"
)

type GlobalID struct {
	Type   string
	Expand string
	moment int64

	stringHole chan string
	int64Hole  chan int64
}

func (p *GlobalID) ID() string {
	if p.moment == 0 {
		p.moment = time.Now().Unix()
		p.stringHole = make(chan string)

		go func() {
			var serial int64 = 0

			for {
				p.stringHole <- fmt.Sprintf("%v%v%v%v", p.Type, p.Expand, p.moment, serial)
				serial += 1
			}
		}()
	}

	return <-p.stringHole
}

func (p *GlobalID) LogicClock(clock int64) int64 {
	if p.moment == 0 {
		p.moment = time.Now().Unix()
		p.int64Hole = make(chan int64)

		go func() {
			for {
				p.int64Hole <- p.moment
				p.moment += 1
			}
		}()
	}

	return <-p.int64Hole
}
