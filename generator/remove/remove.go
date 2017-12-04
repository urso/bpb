package remove

import (
	"github.com/urso/bpb/generator"
	"github.com/urso/bpb/prog/ingest"
	"github.com/urso/bpb/prog/ls"

	"github.com/elastic/beats/libbeat/common"
)

type remove struct {
	config
}

type config struct {
	Field         string `validate:"required"`
	IgnoreFailure bool   `config:"ignore_failure"`
}

func init() {
	generator.Register("remove", makeRemove)
	generator.Register("try_remove", makeTryRemove)
}

func makeRemove(cfg *common.Config) (generator.Processor, error) {
	config := defaultConfig()
	if err := cfg.Unpack(&config); err != nil {
		return nil, err
	}

	return &remove{config}, nil
}

func makeTryRemove(cfg *common.Config) (generator.Processor, error) {
	tryConfig := struct {
		Field string `validate:"required"`
	}{}
	if err := cfg.Unpack(&tryConfig); err != nil {
		return nil, err
	}

	return &remove{config{
		Field:         tryConfig.Field,
		IgnoreFailure: true,
	}}, nil
}

func (r *remove) Name() string { return "remove" }

func (r *remove) CompileIngest() ([]ingest.Processor, error) {
	params := map[string]interface{}{
		"field": r.Field,
	}
	if r.IgnoreFailure {
		params["ignore_failure"] = true
	}

	return ingest.MakeSingleProcessor("remove", params), nil
}

// failure tag: none, need to generate custom tag handling
func (r *remove) CompileLogstash(ctx *generator.LogstashCtx) (generator.FilterBlock, error) {
	var failureTag string

	if !r.IgnoreFailure {
		failureTag = ctx.CreateTag("_failure_remove")
	}

	params := ls.Params{}
	params.RemoveField(r.Field)
	params.RemoveTag(failureTag)

	blk := ls.RunWithTags(
		ls.MakeBlock(
			ls.MakeFilter("mutate", params),
		),
		failureTag,
	)

	return generator.FilterBlock{
		Block:       ls.MakeVerboseBlock(ctx.Verbose, "remove", blk...),
		FailureTags: []string{failureTag},
	}, nil
}

func defaultConfig() config {
	return config{}
}
