package date

import (
	"errors"

	"github.com/urso/bpb/generator"
	"github.com/urso/bpb/prog/ingest"
	"github.com/urso/bpb/prog/ls"

	"github.com/elastic/beats/libbeat/common"
)

type date struct {
	Field, To     string
	Formats       []string
	Locale        string
	Timezone      string
	DropField     bool
	IgnoreFailure bool
}

type config struct {
	Field         string `validate:"required"`
	To            string `config:"target_field"`
	Format        string
	Formats       []string
	Locale        string
	Timezone      string
	DropField     bool `config:"drop_field"`
	IgnoreFailure bool `config:"ignore_failure"`
}

func init() {
	generator.Register("date", makeDate)
}

func makeDate(cfg *common.Config) (generator.Processor, error) {
	config := defaultConfig()
	if err := cfg.Unpack(&config); err != nil {
		return nil, err
	}

	formats := config.Formats
	if config.Format != "" {
		formats = []string{config.Format}
	}

	return &date{
		Field:         config.Field,
		To:            config.To,
		Formats:       formats,
		Locale:        config.Locale,
		Timezone:      config.Timezone,
		DropField:     config.DropField,
		IgnoreFailure: config.IgnoreFailure,
	}, nil
}

func (d *date) Name() string { return "date" }

func (d *date) CompileIngest() ([]ingest.Processor, error) {
	params := map[string]interface{}{
		"field":   d.Field,
		"formats": d.Formats,
	}
	if d.To != "" {
		params["target_field"] = d.To
	}
	if d.Timezone != "" {
		params["timezone"] = d.Timezone
	}
	if d.Locale != "" {
		params["locale"] = d.Locale
	}
	if d.IgnoreFailure {
		params["ignore_failure"] = d.IgnoreFailure
	}

	ps := ingest.MakeSingleProcessor("date", params)
	if d.DropField {
		ps = append(ps, ingest.RemoveField(d.Field))
	}
	return ps, nil
}

func (d *date) CompileLogstash(ctx *generator.LogstashCtx) (generator.FilterBlock, error) {
	var failureTag string
	if !d.IgnoreFailure {
		failureTag = ctx.CreateTag("_failure_date")
	}

	params := ls.Params{
		"match": append([]string{ls.NormalizeField(d.Field)}, d.Formats...),
	}
	if failureTag != "" {
		params["tag_on_failure"] = failureTag
	}

	params.Target(d.To)
	params.DropField(d.DropField, d.Field)

	if d.Timezone != "" {
		params["timezone"] = d.Timezone
	}
	if d.Locale != "" {
		params["locale"] = d.Locale
	}

	return generator.FilterBlock{
		Block: ls.MakeVerboseBlock(ctx.Verbose, "date",
			ls.MakeFilter("date", params),
		),
		FailureTags: []string{failureTag},
	}, nil
}

func defaultConfig() config {
	return config{}
}

func (c *config) Validate() error {
	if len(c.Formats) > 0 && c.Format != "" {
		return errors.New("format and formats is configured")
	}

	return nil
}
