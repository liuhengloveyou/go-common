package common

import (
	"fmt"
	"sync"
	"time"
)

type GlobalID struct {
	Type   string
	Expand string
	moment int64

	stringHole chan string
	lock       sync.Mutex
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
	}

	p.lock.Lock()
	defer p.lock.Unlock()

	p.moment += 1
	if clock >= p.moment {
		p.moment = clock + 1
	}

	return p.moment
}
