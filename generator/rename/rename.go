package grok

import (
	"github.com/urso/bpb/generator"
	"github.com/urso/bpb/prog/ingest"
	"github.com/urso/bpb/prog/ls"

	"github.com/elastic/beats/libbeat/common"
)

type rename struct {
	config
}

type config struct {
	Field         string `validate:"required"`
	To            string `validate:"required"`
	IgnoreMissing bool
}

func init() {
	generator.Register("rename", makeRename)
}

func makeRename(cfg *common.Config) (generator.Processor, error) {
	config := defaultConfig()
	if err := cfg.Unpack(&config); err != nil {
		return nil, err
	}

	return &rename{config}, nil
}

func (r *rename) CompileIngest() ([]ingest.Processor, error) {
	params := map[string]interface{}{
		"field":        r.Field,
		"target_field": r.To,
	}
	if r.IgnoreMissing {
		params["ignore_missing"] = true
	}

	return ingest.MakeSingleProcessor("rename", params), nil
}

func (r *rename) CompileLogstash() (ls.Block, error) {
	return ls.MakeBlock(ls.MakeFilter("mutate", ls.Params{
		"rename": ls.Params{
			ls.NormalizeField(r.Field): ls.NormalizeField(r.To),
		},
	})), nil
}

func defaultConfig() config {
	return config{IgnoreMissing: true}
}
