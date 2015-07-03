package common

import (
	"fmt"
	"time"
)

type GlobalID struct {
	servID string
	moment string
	expand string
	Hole   chan string
}

func (p *GlobalID) Init(servID, expand string) {
	p.servID = servID
	p.expand = expand
	p.moment = fmt.Sprintf("%d", time.Now().Unix())[4:]
	p.Hole = make(chan string)

	go func() {
		var serial int64 = 0

		for {
			p.Hole <- fmt.Sprintf("%v%v%v%v", p.servID, p.moment, p.expand, serial)
			serial += 1
		}
	}()
}
