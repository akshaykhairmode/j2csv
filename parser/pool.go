package parser

import "sync"

var objectPool = sync.Pool{
	New: func() any {
		return make(map[string]any)
	},
}
