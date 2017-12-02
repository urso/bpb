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
	IgnoreMissing bool   `config:"ignore_missing"`
	IgnoreFailure bool   `config:"ignore_failure"`
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

func (r *rename) Name() string { return "rename" }

func (r *rename) CompileIngest() ([]ingest.Processor, error) {
	params := map[string]interface{}{
		"field":        r.Field,
		"target_field": r.To,
	}
	if r.IgnoreMissing {
		params["ignore_missing"] = true
	}
	if r.IgnoreFailure {
		params["ignore_failure"] = true
	}

	return ingest.MakeSingleProcessor("rename", params), nil
}

// failure tag: none, need to generate custom tag handling
func (r *rename) CompileLogstash(ctx *generator.LogstashCtx) (generator.FilterBlock, error) {
	var failureTag string
	if !r.IgnoreFailure {
		failureTag = ctx.CreateTag("_failure_rename")
	}

	mutate := ls.Params{
		"rename": ls.Params{
			ls.NormalizeField(r.Field): ls.NormalizeField(r.To),
		},
	}
	mutate.RemoveTag(failureTag)

	blk := ls.RunWithTags(ls.MakeBlock(ls.MakeFilter("mutate", mutate)), failureTag)
	return generator.FilterBlock{
		Block:       ls.MakeVerboseBlock(ctx.Verbose, "rename", blk...),
		FailureTags: []string{failureTag},
	}, nil
}

func defaultConfig() config {
	return config{IgnoreMissing: true}
}
