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

func (t *sel) CompileIngest() ([]ingest.Processor, error) {
	return generator.CompileIngestProcessors(t.ingest)
}

func (t *sel) CompileLogstash(verbose bool) (ls.Block, error) {
	return generator.CompileLogstashProcessors(t.logstash, verbose)
}

func defaultConfig() config {
	return config{}
}
