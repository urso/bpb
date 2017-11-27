package sel

import (
	"github.com/urso/bpb/generator"
	"github.com/urso/bpb/prog/ingest"
	"github.com/urso/bpb/prog/ls"

	"github.com/elastic/beats/libbeat/common"
)

type sel struct {
	ingest   []generator.Processor
	logstash []generator.Processor
}

type config struct {
	Ingest   []*common.Config
	Logstash []*common.Config
}

func init() {
	generator.Register("select", makeSelect)
}

func makeSelect(cfg *common.Config) (generator.Processor, error) {
	config := defaultConfig()
	if err := cfg.Unpack(&config); err != nil {
		return nil, err
	}

	ingest, err := generator.LoadAll(config.Ingest)
	if err != nil {
		return nil, err
	}

	logstash, err := generator.LoadAll(config.Logstash)
	if err != nil {
		return nil, err
	}

	return &sel{ingest: ingest, logstash: logstash}, nil
}

func (s *sel) Name() string { return "select" }

func (t *sel) CompileIngest() ([]ingest.Processor, error) {
	return generator.CompileIngestProcessors(t.ingest)
}

func (t *sel) CompileLogstash(ctx *generator.LogstashCtx) (generator.FilterBlock, error) {
	failureTags := []string{ctx.CreateTag("_failure_select")}

	onError := func(filter string, tags []string) generator.FilterBlock {
		return generator.FilterBlock{
			Block: ls.MakeBlock(
				ls.MakeFilter("mutate", ls.Params{
					"add_tag": failureTags,
				}),
			),
			FailureTags: failureTags,
		}
	}
	return generator.CompileLogstashProcessors(ctx, onError, t.logstash)
}

func defaultConfig() config {
	return config{}
}
