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
	To        string `config:"target_field"`
	DropField bool   `config:"drop_field"`
}

func init() {
	generator.Register("split_by", makeSplit)
}

func makeSplit(cfg *common.Config) (generator.Processor, error) {
	config := defaultConfig()
	if err := cfg.Unpack(&config); err != nil {
		return nil, err
	}

	return &split{config}, nil
}

func (s *split) Name() string { return "split" }

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

func (s *split) CompileLogstash(ctx *generator.LogstashCtx) (generator.FilterBlock, error) {
	failureTag := ctx.CreateTag("_failure_split")

	var split ls.Block
	if s.Regex != "" {
		split = s.compileLogstashRegex(ctx, failureTag)
	} else {
		split = s.compileLogstashSeparator(failureTag)
	}

	return generator.FilterBlock{
		Block:       ls.MakeVerboseBlock(ctx.Verbose, "split", split),
		FailureTags: []string{failureTag},
	}, nil
}

func (s *split) compileLogstashRegex(ctx *generator.LogstashCtx, failureTag string) ls.Block {
	source, target := s.Field, s.To
	if target == "" {
		target = source
	}

	source = ls.NormalizeField(source)
	target = ls.NormalizeField(target)

	code := fmt.Sprintf(`event.set('%v', event.get('%v').split(/%v/))`, target, source, s.Regex)

	params := ls.Params{}
	params.DropField(s.DropField, s.Field)
	return generator.MakeRuby(ctx, code, failureTag, params)
}

// failure tag: not configurable... potentially multiple (_split_type_failure and on exception?)
func (s *split) compileLogstashSeparator(failureTag string) ls.Block {
	params := ls.Params{"terminator": s.Separator}
	if s.To != "" {
		params.Target(s.To)
	}
	params.DropField(s.DropField, s.Field)
	params.RemoveTag(failureTag)

	blk := ls.MakeBlock(ls.MakeFilter("split", params))
	ls.RunWithTags(blk, failureTag)
	return blk
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
