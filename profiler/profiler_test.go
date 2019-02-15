package profiler_test

import "github.com/Shopify/goose/profiler"

func ExampleNewProfiler() {
	cpuFile := "cpu.prof"
	memoryFile := "memory.prof"

	p := profiler.NewProfiler(cpuFile, memoryFile)
	if err := p.Start(); err != nil {
		panic(err)
	}

	// Do stuff

	if err := p.End(); err != nil {
		panic(err)
	}
}
