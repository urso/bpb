package main

import (
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/elastic/beats/libbeat/common"

	"github.com/urso/bpb/generator"

	// import available processor types
	_ "github.com/urso/bpb/generator/date"
	_ "github.com/urso/bpb/generator/geoip"
	_ "github.com/urso/bpb/generator/grok"
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

func cmdLogstash() *cobra.Command {
	var pipelineID string

	cmdGenerate := &cobra.Command{
		Use:   "generate",
		Short: "Generate logstash filter configuration",
		Run: runWithPipeline(func(gen *generator.Generator) error {
			gen.ID = pipelineID
			return gen.MakeLogstash(os.Stdout)
		}),
	}

	cmd := &cobra.Command{
		Use:   "logstash",
		Short: "Logstash Mode",
	}
	cmd.PersistentFlags().StringVar(&pipelineID, "id", "", "pipeline ID")
	cmd.AddCommand(cmdGenerate)
	return cmd
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
		Description string           `config:"pipeline.description"`
		Processors  []*common.Config `config:"pipeline.processors"`
	}{}
	if err := cfg.Unpack(&pipeline); err != nil {
		log.Fatal(err)
	}

	return generator.New(pipeline.Description, pipeline.Processors)
}
