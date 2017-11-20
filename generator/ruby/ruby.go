package ruby

import (
	"errors"

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

func (r *ruby) CompileLogstash() (ls.Block, error) {
	return ls.MakeBlock(ls.MakeFilter("ruby", ls.Params{
		"code": r.Code,
	})), nil
}

func defaultConfig() config {
	return config{}
}
