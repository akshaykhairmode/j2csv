package parser

import "sync"

type pool struct {
	mapAnyPool, stringSlicePool *sync.Pool //Map pool for decoding the object. string slice pool for csvWriter.
	length                      int        //Size of the map and slice which we will create
	enabled                     bool       //should enable pooling or not
}

func (po *pool) GetMapStringAny() map[string]any {
	if po.mapAnyPool == nil {
		return make(map[string]any, po.length)
	}
	return po.mapAnyPool.Get().(map[string]any)
}

func (po *pool) PutMapStringAny(m map[string]any) {
	if po.mapAnyPool == nil {
		return
	}

	for k := range m {
		delete(m, k)
	}

	po.mapAnyPool.Put(m)
}

func (po *pool) GetStringSlice() []string {
	if po.stringSlicePool == nil {
		return make([]string, 0, po.length)
	}

	return po.stringSlicePool.Get().([]string)
}

func (po *pool) PutStringSlice(s []string) {
	if po.stringSlicePool == nil {
		return
	}

	s = s[:0]
	po.stringSlicePool.Put(s)
}

func (po *pool) SetPools(size int) {

	po.length = size

	if !po.enabled {
		return
	}

	po.mapAnyPool = &sync.Pool{
		New: func() any {
			return make(map[string]any, po.length)
		},
	}

	po.stringSlicePool = &sync.Pool{
		New: func() any {
			return make([]string, 0, po.length)
		},
	}

}
