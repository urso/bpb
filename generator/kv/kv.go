package kv

import (
	"errors"
	"fmt"

	"github.com/urso/bpb/generator"
	"github.com/urso/bpb/prog/ingest"

	"github.com/elastic/beats/libbeat/common"
)

type kv struct {
	config
}

type config struct {
	Field         string      `validate:"required"`
	To            string      `validate:"required"`
	FieldSplit    splitConfig `config:"split.field"`
	ValueSplit    splitConfig `config:"split.value"`
	IgnoreMissing bool        `config:"ignore_missing"`
	IgnoreFailure bool        `config:"ignore_failure"`
}

type splitConfig struct {
	mode    splitMode
	pattern string
}

type splitMode uint8

const (
	noSplitMode splitMode = iota
	classSplitMode
	regexSplitMode
	// TODO: more split modes?
)

func init() {
	generator.Register("kv", makeKV)
}

func makeKV(cfg *common.Config) (generator.Processor, error) {
	config := defaultConfig()
	if err := cfg.Unpack(&config); err != nil {
		return nil, err
	}

	return &kv{config}, nil
}

func (k *kv) Name() string {
	return "kv"
}

func (k *kv) CompileIngest() ([]ingest.Processor, error) {
	fieldSplit, err := ingestPattern(k.FieldSplit)
	if err != nil {
		return nil, fmt.Errorf("%v on field", err)
	}

	valueSplit, err := ingestPattern(k.ValueSplit)
	if err != nil {
		return nil, fmt.Errorf("%v on value", err)
	}

	params := map[string]interface{}{
		"field":       k.Field,
		"field_split": fieldSplit,
		"value_split": valueSplit,
	}
	if k.To != "" {
		params["target_field"] = k.To
	}
	if k.IgnoreMissing {
		params["ignore_missing"] = true
	}
	if k.IgnoreFailure {
		params["ignore_failure"] = true
	}

	return ingest.MakeSingleProcessor("kv", params), nil
}

func (k *kv) CompileLogstash(ctx *generator.LogstashCtx) (generator.FilterBlock, error) {
	return generator.FilterBlock{}, errors.New("TODO (logstash kv filter)")
}

func defaultConfig() config {
	return config{}
}

func (c *splitConfig) Unpack(cfg *common.Config) error {
	cfg.PrintDebugf("splitConfig: ")

	fields := cfg.GetFields()
	switch len(fields) {
	case 0:
		return errors.New("no split option given")
	case 1:
		break
	default:
		return errors.New("more then 1 split option")
	}

	mode := fields[0]
	value, err := cfg.String(mode, -1)
	if err != nil {
		return err
	}

	switch mode {
	case "class":
		c.mode = classSplitMode
	case "regex":
		c.mode = regexSplitMode
	default:
		return fmt.Errorf("'%v' is no valid split mode", mode)
	}

	c.pattern = value
	return nil
}

func ingestPattern(c splitConfig) (string, error) {
	switch c.mode {
	case classSplitMode:
		return fmt.Sprintf("[%v]+", c.pattern), nil
	case regexSplitMode:
		return c.pattern, nil
	default:
		return "", errors.New("no split mode configured")
	}
}
