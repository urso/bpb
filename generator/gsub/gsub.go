package gsub

import (
	"errors"

	"github.com/urso/bpb/generator"
	"github.com/urso/bpb/prog/ingest"

	"github.com/elastic/beats/libbeat/common"
)

type gsub struct {
	config
}

type config struct {
	Field         string `validate:"required"`
	Pattern       string `validate:"required"`
	Replacement   string `validate:"required"`
	To            string `config:"target_field"`
	IgnoreMissing bool   `config:"ignore_missing"`
	IgnoreFailure bool   `config:"ignore_failure"`
	DropField     bool   `config:"drop_field"`
}

func init() {
	generator.Register("gsub", makeGsub)
}

func makeGsub(cfg *common.Config) (generator.Processor, error) {
	config := defaultConfig()
	if err := cfg.Unpack(&config); err != nil {
		return nil, err
	}

	return &gsub{config}, nil
}

func defaultConfig() config {
	return config{}
}

func (g *gsub) Name() string { return "gsub" }

func (g *gsub) CompileIngest() ([]ingest.Processor, error) {
	params := map[string]interface{}{
		"field":       g.Field,
		"pattern":     g.Pattern,
		"replacement": g.Replacement,
	}
	if g.To != "" {
		params["target_field"] = g.To
	}
	if g.IgnoreMissing {
		params["ignore_missing"] = true
	}
	if g.IgnoreFailure {
		params["ignore_failure"] = true
	}

	ps := ingest.MakeSingleProcessor("gsub", params)
	if g.DropField {
		ps = append(ps, ingest.RemoveField(g.Field))
	}

	return ps, nil
}

func (g *gsub) CompileLogstash(ctx *generator.LogstashCtx) (generator.FilterBlock, error) {
	return generator.FilterBlock{}, errors.New("TODO (logstash gsub)")
}
