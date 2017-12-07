package useragent

import (
	"github.com/urso/bpb/generator"
	"github.com/urso/bpb/prog/ingest"
	"github.com/urso/bpb/prog/ls"

	"github.com/elastic/beats/libbeat/common"
)

type useragent struct {
	config
}

type config struct {
	Field         string `validate:"required"`
	To            string `config:"target_field"`
	DropField     bool   `config:"drop_field"`
	IgnoreFailure bool   `config:"ignore_failure"`
}

func init() {
	generator.Register("user_agent", makeUserAgent)
}

func makeUserAgent(cfg *common.Config) (generator.Processor, error) {
	config := defaultConfig()
	if err := cfg.Unpack(&config); err != nil {
		return nil, err
	}

	return &useragent{config}, nil
}

func (u *useragent) Name() string { return "useragent" }

func (u *useragent) CompileIngest() ([]ingest.Processor, error) {
	params := map[string]interface{}{
		"field": u.Field,
	}
	if u.To != "" {
		params["target_field"] = u.To
	}
	if u.IgnoreFailure {
		params["ignore_failure"] = true
	}

	ps := ingest.MakeSingleProcessor("user_agent", params)
	if u.DropField {
		ps = append(ps, ingest.RemoveField(u.Field))
	}
	return ps, nil
}

// failure tag: none, need to generate custom tag handling
func (u *useragent) CompileLogstash(ctx *generator.LogstashCtx) (generator.FilterBlock, error) {
	var failureTag string
	if !u.IgnoreFailure {
		failureTag = ctx.CreateTag("_failure_useragent")
	}

	params := ls.Params{
		"source": ls.NormalizeField(u.Field),
	}
	params.Target(u.To)
	params.DropField(u.DropField, u.Field)
	params.RemoveTag(failureTag)

	blk := ls.MakeBlock(ls.MakeFilter("useragent", params))
	blk = ls.RunWithTags(blk, failureTag)
	blk = ls.MakeVerboseBlock(ctx.Verbose, "useragent", blk)

	return generator.FilterBlock{
		Block:       blk,
		FailureTags: []string{failureTag},
	}, nil
}

func defaultConfig() config {
	return config{}
}
