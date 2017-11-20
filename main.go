package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/urso/bpb/generator"

	"github.com/elastic/beats/libbeat/common"

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
	var (
		id     string
		format string
	)
	flag.StringVar(&id, "id", "", "pipeline ID")
	flag.StringVar(&format, "format", "ingest", "pipeline output format")
	flag.Parse()
	files := flag.Args()
	if len(files) == 0 {
		log.Fatal("No files given")
	}

	cfg, err := common.LoadFiles(files...)
	if err != nil {
		log.Fatal(err)
	}

	pipeline := struct {
		Description string           `config:"pipeline.description"`
		Processors  []*common.Config `config:"pipeline.processors"`
	}{}
	if err := cfg.Unpack(&pipeline); err != nil {
		log.Fatal(err)
	}

	gen, err := generator.New(pipeline.Description, pipeline.Processors)
	gen.ID = id
	if err != nil {
		log.Fatal(err)
	}

	switch format {
	case "ingest":
		err = gen.MakeIngest(os.Stdout)
	case "logstash":
		err = gen.MakeLogstash(os.Stdout)
	default:
		err = fmt.Errorf("unkown output format '%v'", format)
	}
	if err != nil {
		log.Fatal(err)
	}
}
