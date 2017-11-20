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
	To        string
	DropField bool `config:"drop_field"`
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

func (g *geoip) CompileLogstash() (ls.Block, error) {
	params := ls.Params{
		"source": ls.NormalizeField(g.Field),
	}
	params.Target(g.To)
	params.DropField(g.DropField, g.Field)
	return ls.MakeBlock(ls.MakeFilter("geoip", params)), nil
}

func defaultConfig() config {
	return config{}
}
