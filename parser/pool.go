package parser

import "sync"

func (p *parser) setPools() {

	p.objPool = &sync.Pool{
		New: func() any {
			return make(map[string]any, len(p.headers))
		},
	}

	p.strPool = &sync.Pool{
		New: func() any {
			return make([]string, 0, len(p.headers))
		},
	}

}
