package ruby

import (
	"errors"
	"strings"

	"github.com/urso/bpb/generator"
	"github.com/urso/bpb/prog/ingest"
	"github.com/urso/bpb/prog/ls"

	"github.com/elastic/beats/libbeat/common"
)

type ruby struct {
	config
}

type config struct {
	Code string
}

func init() {
	generator.Register("ruby", makeRuby)
}

func makeRuby(cfg *common.Config) (generator.Processor, error) {
	config := defaultConfig()
	if err := cfg.Unpack(&config); err != nil {
		return nil, err
	}

	return &ruby{config}, nil
}

func (r *ruby) Name() string { return "ruby" }

func (r *ruby) CompileIngest() ([]ingest.Processor, error) {
	return nil, errors.New("ruby not supported on 'ingest' target")
}

// failure tag: config via `tag_on_exception` (default: `_rubyexception`)
func (r *ruby) CompileLogstash(ctx *generator.LogstashCtx) (generator.FilterBlock, error) {
	failureTag := ctx.CreateTag("_failure_ruby")
	code := strings.Replace(r.Code, "\n", "; ", -1)

	blk := generator.MakeRuby(ctx, code, failureTag, nil)
	return generator.FilterBlock{
		Block:       ls.MakeVerboseBlock(ctx.Verbose, "ruby", blk...),
		FailureTags: []string{failureTag},
	}, nil
}

func defaultConfig() config {
	return config{}
}
