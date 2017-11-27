package remove

import (
	"github.com/urso/bpb/generator"
	"github.com/urso/bpb/prog/ingest"
	"github.com/urso/bpb/prog/ls"

	"github.com/elastic/beats/libbeat/common"
)

type remove struct {
	Field string
}

type config struct {
	Field string `validate:"required"`
}

func init() {
	generator.Register("remove", makeRemove)
}

func makeRemove(cfg *common.Config) (generator.Processor, error) {
	config := defaultConfig()
	if err := cfg.Unpack(&config); err != nil {
		return nil, err
	}

	return &remove{Field: config.Field}, nil
}

func (r *remove) Name() string { return "remove" }

func (r *remove) CompileIngest() ([]ingest.Processor, error) {
	return ingest.MakeSingleProcessor("remove", map[string]interface{}{
		"field": r.Field,
	}), nil
}

// failure tag: none, need to generate custom tag handling
func (r *remove) CompileLogstash(ctx *generator.LogstashCtx) (generator.FilterBlock, error) {
	failureTag := ctx.CreateTag("_failure_remove")

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
