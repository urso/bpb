package generator

import (
	"errors"
	"fmt"

	"github.com/urso/bpb/prog/ingest"
	"github.com/urso/bpb/prog/ls"

	"github.com/elastic/beats/libbeat/common"
)

var processors = map[string]Factory{}

type Factory func(config *common.Config) (Processor, error)

type LogstashCtx struct {
	Verbose bool
}

type Processor interface {
	CompileIngest() ([]ingest.Processor, error)
	CompileLogstash(ctx *LogstashCtx) (ls.Block, error)
}

func Register(name string, f Factory) {
	if processors[name] != nil {
		panic(fmt.Errorf("Generator %v already registered", name))
	}
	processors[name] = f
}

func Find(name string) Factory {
	return processors[name]
}

func LoadAll(configs []*common.Config) ([]Processor, error) {
	if len(configs) == 0 {
		return nil, nil
	}

	ps := make([]Processor, len(configs))
	for i, cfg := range configs {
		var err error
		if ps[i], err = Load(cfg); err != nil {
			return nil, err
		}
	}
	return ps, nil
}

func Load(config *common.Config) (Processor, error) {
	processor := map[string]*common.Config{}
	err := config.Unpack(&processor)
	if err != nil {
		return nil, err
	}

	if len(processor) == 0 {
		return nil, errors.New("can not load empty processor")
	}
	if len(processor) > 1 {
		return nil, errors.New("multiple processors")
	}

	var name string
	var params *common.Config
	for n, p := range processor {
		name, params = n, p
	}

	return LoadNamed(name, params)
}

func LoadNamed(name string, config *common.Config) (Processor, error) {
	factory := Find(name)
	if factory == nil {
		return nil, fmt.Errorf("processor '%v' not available", name)
	}

	return factory(config)
}
