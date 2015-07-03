package common

import (
	"fmt"
	"time"
)

type GlobalID struct {
	servID string
	monent int64
	expand string
	Hole   chan string
}

func (p *GlobalID) Init(servID, expand string) {
	p.servID = servID
	p.expand = expand
	p.monent = time.Now().Unix()[4:]
	p.Hole = make(chan string)

	go func() {
		var serial int64 = 0

		for {
			p.Hole <- fmt.Sprintf("%v%v%v%v", p.servID, p.monent, p.expand, serial)
			serial += 1
		}
	}()
}
