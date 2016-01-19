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
	int64Hole  chan int64

	lock *sync.Mutex
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
		p.lock = new(sync.Mutex)

		go func() {
			for {
				p.moment += 1
				p.int64Hole <- p.moment
			}
		}()
	}

	p.lock.Lock()
	defer p.lock.Unlock()

	current := <-p.int64Hole
	for clock >= current {
		current = <-p.int64Hole
	}

	return current
}
