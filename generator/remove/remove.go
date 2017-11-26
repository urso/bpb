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

func (r *remove) CompileIngest() ([]ingest.Processor, error) {
	return ingest.MakeSingleProcessor("remove", map[string]interface{}{
		"field": r.Field,
	}), nil
}

// failure tag: none, need to generate custom tag handling
func (r *remove) CompileLogstash(ctx *generator.LogstashCtx) (ls.Block, error) {
	params := ls.Params{}
	params.RemoveField(r.Field)
	return ls.MakeVerboseBlock(ctx.Verbose, "remove", ls.MakeFilter("mutate", params)), nil
}

func defaultConfig() config {
	return config{}
}
