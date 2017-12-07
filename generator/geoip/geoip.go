package geoip

import (
	"github.com/urso/bpb/generator"
	"github.com/urso/bpb/prog/ingest"
	"github.com/urso/bpb/prog/ls"

	"github.com/elastic/beats/libbeat/common"
)

type geoip struct {
	config
}

type config struct {
	Field     string `validate:"required"`
	To        string `config:"target_field"`
	DropField bool   `config:"drop_field"`
}

func init() {
	generator.Register("geoip", makeGeoip)
}

func makeGeoip(cfg *common.Config) (generator.Processor, error) {
	config := defaultConfig()
	if err := cfg.Unpack(&config); err != nil {
		return nil, err
	}

	return &geoip{config}, nil
}

func (g *geoip) Name() string { return "geoip" }

func (u *geoip) CompileIngest() ([]ingest.Processor, error) {
	params := map[string]interface{}{
		"field": u.Field,
	}
	if u.To != "" {
		params["target_field"] = u.To
	}

	ps := ingest.MakeSingleProcessor("geoip", params)
	if u.DropField {
		ps = append(ps, ingest.RemoveField(u.Field))
	}
	return ps, nil
}

// failure tag: config via `tag_on_failure` (default: `_geoip_lookup_failure`)
func (g *geoip) CompileLogstash(ctx *generator.LogstashCtx) (generator.FilterBlock, error) {
	failureTag := ctx.CreateTag("_failure_geoip")

	params := ls.Params{
		"source":         ls.NormalizeField(g.Field),
		"tag_on_failure": failureTag,
	}
	params.Target(g.To)
	params.DropField(g.DropField, g.Field)
	return generator.FilterBlock{
		Block: ls.MakeVerboseBlock(ctx.Verbose, "geoip",
			ls.MakeFilter("geoip", params),
		),
		FailureTags: []string{failureTag},
	}, nil
}

func defaultConfig() config {
	return config{}
}
