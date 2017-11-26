package script

import (
	"errors"
	"strings"

	"github.com/urso/bpb/generator"
	"github.com/urso/bpb/prog/ingest"
	"github.com/urso/bpb/prog/ls"

	"github.com/elastic/beats/libbeat/common"
)

type script struct {
	config
}

type config struct {
	Code string
	ID   string
}

func init() {
	generator.Register("script", makeScript)
}

func makeScript(cfg *common.Config) (generator.Processor, error) {
	config := defaultConfig()
	if err := cfg.Unpack(&config); err != nil {
		return nil, err
	}

	if code := config.Code; code != "" {
		// remove newlines
		config.Code = strings.Replace(code, "\n", "", -1)
	}

	return &script{config: config}, nil
}

func (s *script) CompileIngest() ([]ingest.Processor, error) {
	params := map[string]interface{}{
		"lang": "painless",
	}
	if s.Code != "" {
		params["source"] = s.Code
	}
	if s.ID != "" {
		params["id"] = s.ID
	}

	return ingest.MakeSingleProcessor("script", params), nil
}

func (s *script) CompileLogstash(_ *generator.LogstashCtx) (ls.Block, error) {
	return nil, errors.New("script not supported on 'logstash' target")
}

func defaultConfig() config {
	return config{}
}

func (c *config) Validate() error {
	if c.Code == "" && c.ID == "" {
		return errors.New("code or script id required")
	}

	if c.Code != "" && c.ID != "" {
		return errors.New("only code or id allowed")
	}

	return nil
}
