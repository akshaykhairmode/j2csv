package parser

import (
	"testing"
)

func BenchmarkPoolEnabled(b *testing.B) {
	p := pool{
		enabled: true,
	}
	p.SetPools(5)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sl := p.GetMapStringAny()
		sl["k1"] = "v"
		sl["k2"] = "v"
		sl["k3"] = "v"
		p.PutMapStringAny(sl)

		ss := p.GetStringSlice()
		ss = append(ss, "test1", "test2", "test3", "test4", "test5", "test1", "test2", "test3", "test4", "test5")
		p.PutStringSlice(ss)
	}
}

func BenchmarkPoolDisabled(b *testing.B) {
	p := pool{}
	p.SetPools(5)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sl := p.GetMapStringAny()
		sl["k1"] = "v"
		sl["k2"] = "v"
		sl["k3"] = "v"
		p.PutMapStringAny(sl)

		ss := p.GetStringSlice()
		ss = append(ss, "test1", "test2", "test3", "test4", "test5", "test1", "test2", "test3", "test4", "test5")
		p.PutStringSlice(ss)
	}
}
