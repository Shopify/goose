package profiler

import (
	"fmt"
	"os"
	"path"
	"runtime"
	"runtime/pprof"

	"github.com/Shopify/goose/logger"
)

var log = logger.New("profiler")

type Profiler interface {
	Start() error
	End() error
}

func New(cpuFile string, memoryFile string) Profiler {
	var profilers []Profiler
	if cpuFile != "" {
		profilers = append(profilers, &cpuProfiler{filePath: cpuFile})
	}
	if memoryFile != "" {
		profilers = append(profilers, &memoryProfiler{filePath: memoryFile})
	}

	if len(profilers) == 0 {
		return nil
	}

	return &compositeProfiler{profilers: profilers}
}

type compositeProfiler struct {
	profilers []Profiler
}

func (p *compositeProfiler) Start() error {
	var err error

	for _, prof := range p.profilers {
		err = prof.Start()
		if err != nil {
			log(nil, err).
				WithField("profiler", fmt.Sprintf("%T", prof)).
				Error("error while starting profiler")
		}
	}

	return err
}

func (p *compositeProfiler) End() error {
	var err error

	for _, prof := range p.profilers {
		err = prof.End()
		if err != nil {
			log(nil, err).
				WithField("profiler", fmt.Sprintf("%T", prof)).
				Error("error while stopping profiler")
		}
	}

	return err
}

type cpuProfiler struct {
	filePath string
}

func (p *cpuProfiler) Start() error {
	f, err := createFile(p.filePath)
	if err != nil {
		return err
	}

	return pprof.StartCPUProfile(f)
}

func (p *cpuProfiler) End() error {
	pprof.StopCPUProfile()
	return nil
}

type memoryProfiler struct {
	filePath string
	file     *os.File
}

func (p *memoryProfiler) Start() error {
	f, err := createFile(p.filePath)
	if err != nil {
		return err
	}

	p.file = f
	return nil
}

func (p *memoryProfiler) End() error {
	defer p.file.Close()

	runtime.GC() // get up-to-date statistics

	return pprof.WriteHeapProfile(p.file)
}

func createFile(name string) (*os.File, error) {
	err := os.MkdirAll(path.Dir(name), 0755)
	if err != nil {
		return nil, err
	}

	f, err := os.Create(name)
	if err != nil {
		return nil, err
	}

	return f, nil
}
