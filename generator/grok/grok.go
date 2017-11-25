package grok

import (
	"errors"

	"github.com/urso/bpb/generator"
	"github.com/urso/bpb/prog/ingest"
	"github.com/urso/bpb/prog/ls"

	"github.com/elastic/beats/libbeat/common"
)

type grok struct {
	Field         string
	Patterns      []string
	Definitions   map[string]string
	IgnoreMissing bool
	DropField     bool
}

type config struct {
	Field         string `validate:"required"`
	Pattern       string
	Patterns      []string
	Definitions   map[string]string
	IgnoreMissing bool `config:"ignore_missing"`
	DropField     bool `config:"drop_field"`
}

func init() {
	generator.Register("grok", makeGrok)
}

func makeGrok(cfg *common.Config) (generator.Processor, error) {
	config := defaultConfig()
	if err := cfg.Unpack(&config); err != nil {
		return nil, err
	}

	patterns := config.Patterns
	if config.Pattern != "" {
		patterns = []string{config.Pattern}
	}

	return &grok{
		Field:         config.Field,
		Patterns:      patterns,
		Definitions:   config.Definitions,
		IgnoreMissing: config.IgnoreMissing,
		DropField:     config.DropField,
	}, nil
}

func (g *grok) CompileIngest() ([]ingest.Processor, error) {
	params := map[string]interface{}{
		"field":    g.Field,
		"patterns": g.Patterns,
	}
	if len(g.Definitions) > 0 {
		params["pattern_definitions"] = g.Definitions
	}
	if g.IgnoreMissing {
		params["ignore_missing"] = true
	}

	ps := ingest.MakeSingleProcessor("grok", params)
	if g.DropField {
		ps = append(ps, ingest.RemoveField(g.Field))
	}
	return ps, nil
}

func (g *grok) CompileLogstash(verbose bool) (ls.Block, error) {
	params := ls.Params{
		"match": map[string]interface{}{
			ls.NormalizeField(g.Field): g.Patterns,
		},
	}
	if len(g.Definitions) > 0 {
		params["pattern_definitions"] = g.Definitions
	}

	params.DropField(g.DropField, g.Field)

	blk := ls.MakeBlock(
		ls.MakeFilter("grok", params),

		// TODO parse and fix field names in grok pattern instead of adding the
		//      de_dot filter
		ls.MakeFilter("de_dot", ls.Params{"nested": true}),
	)
	if g.IgnoreMissing {
		blk = ls.IgnoreMissing(g.Field, blk)
	}

	return ls.MakeVerboseBlock(verbose, "grok", blk...), nil
}

func defaultConfig() config {
	return config{}
}

func (c *config) Validate() error {
	if c.Field == "" {
		return errors.New("field missing")
	}

	if c.Pattern != "" && len(c.Patterns) > 0 {
		return errors.New("set `pattern` or `patterns` setting only")
	}

	return nil
}
