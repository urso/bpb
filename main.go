package main

import (
	"log"

	"github.com/spf13/cobra"

	"github.com/elastic/beats/libbeat/common"

	"github.com/urso/bpb/generator"

	// import available processor types
	_ "github.com/urso/bpb/generator/convert"
	_ "github.com/urso/bpb/generator/date"
	_ "github.com/urso/bpb/generator/geoip"
	_ "github.com/urso/bpb/generator/grok"
	_ "github.com/urso/bpb/generator/gsub"
	_ "github.com/urso/bpb/generator/json"
	_ "github.com/urso/bpb/generator/kv"
	_ "github.com/urso/bpb/generator/remove"
	_ "github.com/urso/bpb/generator/rename"
	_ "github.com/urso/bpb/generator/ruby"
	_ "github.com/urso/bpb/generator/script"
	_ "github.com/urso/bpb/generator/sel"
	_ "github.com/urso/bpb/generator/split"
	_ "github.com/urso/bpb/generator/useragent"
)

func main() {
	main := cobra.Command{Short: "beats pipeline builder"}
	main.AddCommand(cmdLogstash(), cmdIngest())
	main.Execute()
}

func runWithPipeline(
	fn func(gen *generator.Generator) error,
) func(*cobra.Command, []string) {
	return func(_ *cobra.Command, args []string) {
		gen, err := loadPipeline(args)
		if err != nil {
			log.Fatal(err)
		}

		if err := fn(gen); err != nil {
			log.Fatal(err)
		}
	}
}

func loadPipeline(files []string) (*generator.Generator, error) {
	cfg, err := common.LoadFiles(files...)
	if err != nil {
		return nil, err
	}

	pipeline := struct {
		Description string           `config:"description"`
		Processors  []*common.Config `config:"processors"`
	}{}
	if err := cfg.Unpack(&pipeline); err != nil {
		log.Fatal(err)
	}

	return generator.New(pipeline.Description, pipeline.Processors)
}
