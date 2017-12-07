package convert

import (
	"errors"
	"fmt"

	"github.com/urso/bpb/generator"
	"github.com/urso/bpb/prog/ingest"

	"github.com/elastic/beats/libbeat/common"
)

type convert struct {
	config
}

type config struct {
	Field         string   `validate:"required"`
	To            string   `config:"target_field"`
	Type          convType `validate:"required"`
	IgnoreMissing bool     `config:"ignore_missing"`
	IgnoreFailure bool     `config:"ignore_failure"`
	DropField     bool     `config:"drop_field"`
}

type convType uint8

const (
	invalidConv convType = iota
	convBool
	convInt
	convFloat
	convString
)

func init() {
	generator.Register("convert", makeConvert)
}

func makeConvert(cfg *common.Config) (generator.Processor, error) {
	config := defaultConfig()
	if err := cfg.Unpack(&config); err != nil {
		return nil, err
	}

	return &convert{config}, nil
}

func (c *convert) Name() string { return "convert" }

func (c *convert) CompileIngest() ([]ingest.Processor, error) {
	params := map[string]interface{}{
		"field": c.Field,
		"type":  c.Type.String(),
	}
	if c.To != "" {
		params["target_field"] = c.To
	}
	if c.IgnoreMissing {
		params["ignore_missing"] = true
	}
	if c.IgnoreFailure {
		params["ignore_failure"] = true
	}

	ps := ingest.MakeSingleProcessor("convert", params)
	if c.DropField {
		ps = append(ps, ingest.RemoveField(c.Field))
	}
	return ps, nil
}

func (C *convert) CompileLogstash(ctx *generator.LogstashCtx) (generator.FilterBlock, error) {
	return generator.FilterBlock{}, errors.New("TODO (logstash convert)")
}

func defaultConfig() config {
	return config{
		IgnoreMissing: true,
	}
}

func (t *convType) Unpack(name string) error {
	v := getConvType(name)
	if v == invalidConv {
		return fmt.Errorf("type '%v' not supported", name)
	}

	*t = v
	return nil
}

func (t convType) String() string {
	return map[convType]string{
		convBool:   "bool",
		convInt:    "integer",
		convFloat:  "float",
		convString: "string",
	}[t]
}

func getConvType(name string) convType {
	return map[string]convType{
		"bool":    convBool,
		"integer": convInt,
		"int":     convInt,
		"float":   convFloat,
		"string":  convString,
	}[name]
}
