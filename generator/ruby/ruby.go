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

func (r *ruby) CompileIngest() ([]ingest.Processor, error) {
	return nil, errors.New("ruby not supported on 'ingest' target")
}

// failure tag: config via `tag_on_exception` (default: `_rubyexception`)
func (r *ruby) CompileLogstash(ctx *generator.LogstashCtx) (ls.Block, error) {
	code := strings.Replace(r.Code, "\n", "; ", -1)
	return ls.MakeVerboseBlock(ctx.Verbose, "ruby",
		ls.MakeFilter("ruby", ls.Params{
			"code": code,
		}),
	), nil
}

func defaultConfig() config {
	return config{}
}
