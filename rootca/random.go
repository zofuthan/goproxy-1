// http://stackoverflow.com/questions/12771930/what-is-the-fastest-way-to-generate-a-long-random-string-in-go

package rootca

import (
	"math/rand"
)

type randomDataMaker struct {
	src rand.Source
}

func (r *randomDataMaker) Read(p []byte) (n int, err error) {
	for i := range p {
		p[i] = byte(r.src.Int63() & 0xff)
	}
	return len(p), nil
}
