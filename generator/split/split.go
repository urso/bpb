package split

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/urso/bpb/generator"
	"github.com/urso/bpb/prog/ingest"
	"github.com/urso/bpb/prog/ls"

	"github.com/elastic/beats/libbeat/common"
)

type split struct {
	config
}

type config struct {
	Field     string `validate:"required"`
	Separator string
	Regex     string
	To        string
	DropField bool `config:"drop_field"`
}

func init() {
	generator.Register("split", makeSplit)
}

func makeSplit(cfg *common.Config) (generator.Processor, error) {
	config := defaultConfig()
	if err := cfg.Unpack(&config); err != nil {
		return nil, err
	}

	return &split{config}, nil
}

func (s *split) CompileIngest() ([]ingest.Processor, error) {
	var split ingest.Processor
	if s.Regex != "" {
		split = s.compileIngestRegex()
	} else {
		split = s.compileIngestSeparator()
	}

	ps := ingest.Single(split)
	if s.DropField {
		ps = append(ps, ingest.RemoveField(s.Field))
	}
	return ps, nil
}

func (s *split) compileIngestRegex() ingest.Processor {
	params := map[string]interface{}{
		"field":     s.Field,
		"separator": s.Regex,
	}
	if s.To != "" {
		params["target_field"] = s.To
	}

	return ingest.MakeProcessor("split", params)
}

func (s *split) compileIngestSeparator() ingest.Processor {
	source, target := s.Field, s.To
	if target == "" {
		target = s.Field
	}

	code := fmt.Sprintf(`ctx.%v = ctx.%v.split(Patter.quote("%v")`, target, source, strconv.Quote(s.Separator))
	return ingest.MakeProcessor("script", map[string]interface{}{
		"lang":   "painless",
		"source": code,
	})
}

func (s *split) CompileLogstash(ctx *generator.LogstashCtx) (ls.Block, error) {
	var split ls.Filter
	if s.Regex != "" {
		split = s.compileLogstashRegex()
	} else {
		split = s.compileLogstashSeparator()
	}

	split.Params.DropField(s.DropField, s.Field)
	return ls.MakeVerboseBlock(ctx.Verbose, "split", split), nil
}

// failure tag: config via `tag_on_exception` (default: `_rubyexception`)
func (s *split) compileLogstashRegex() ls.Filter {
	source, target := s.Field, s.To
	if target == "" {
		target = source
	}

	source = ls.NormalizeField(source)
	target = ls.NormalizeField(target)

	code := fmt.Sprintf(`event.set('%v', event.get('%v').split(/%v/))`, target, source, s.Regex)
	return ls.MakeFilter("ruby", ls.Params{
		"code": code,
	})
}

// failure tag: not configurable... potentially multiple (_split_type_failure and on exception?)
func (s *split) compileLogstashSeparator() ls.Filter {
	params := ls.Params{"terminator": s.Separator}
	if s.To != "" {
		params.Target(s.To)
	}
	return ls.MakeFilter("split", params)
}

func defaultConfig() config {
	return config{}
}

func (c *config) Validate() error {
	if c.Separator == "" && c.Regex == "" {
		return errors.New("split requires separator or regex setting")
	}

	if c.Separator != "" && c.Regex != "" {
		return errors.New("separator and regex set")
	}

	return nil
}
