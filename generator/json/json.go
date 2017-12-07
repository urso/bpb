package main

import (
	"errors"

	"github.com/urso/bpb/generator"
	"github.com/urso/bpb/prog/ingest"

	"github.com/elastic/beats/libbeat/common"
)

type processor struct {
	config
}

type config struct {
	Field         string `validate:"required"`
	To            string
	IgnoreFailure bool `config:"ignore_failure"`
	DropField     bool `config:"drop_field"`
}

func init() {
	generator.Register("json", makeProcessor)
}

func makeProcessor(cfg *common.Config) (generator.Processor, error) {
	config := defaultConfig()
	if err := cfg.Unpack(&config); err != nil {
		return nil, err
	}

	return &processor{config}, nil
}

func defaultConfig() config {
	return config{}
}

func (p *processor) Name() string { return "json" }

func (p *processor) CompileIngest() ([]ingest.Processor, error) {
	params := map[string]interface{}{
		"field": p.Field,
	}
	if p.To != "" {
		params["target_field"] = p.To
	} else {
		params["add_to_root"] = true
	}
	if p.IgnoreFailure {
		params["ignore_failure"] = true
	}

	ps := ingest.MakeSingleProcessor("json", params)
	if p.DropField {
		ps = append(ps, ingest.RemoveField(p.Field))
	}
	return ps, nil
}

func (p *processor) CompileLogstash(ctx *generator.LogstashCtx) (generator.FilterBlock, error) {
	return generator.FilterBlock{}, errors.New("TODO (logstash json)")
}
