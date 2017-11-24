package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/spf13/cobra"

	"github.com/elastic/beats/libbeat/common"

	"github.com/urso/bpb/generator"
	"github.com/urso/bpb/prog/ingest"

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

func cmdIngest() *cobra.Command {
	cmdGenerate := &cobra.Command{
		Use:   "generate",
		Short: "Generate Ingest Node pipeline configuration",
		Run: runWithPipeline(func(gen *generator.Generator) error {
			return gen.MakeIngest(os.Stdout)
		}),
	}

	var (
		host        string
		eventFormat string
		inFile      string
		verbose     bool
	)
	cmdRun := &cobra.Command{
		Use:   "run",
		Short: "Run pipeline",
		Long:  "Run ingest pipeline with Elasticsearch Ingest Node and sample events. This command uses the simulate API",
		Run: runWithPipeline(func(gen *generator.Generator) error {
			prog, err := gen.CompileIngest()
			if err != nil {
				return err
			}

			docs, err := readEvents(eventFormat, inFile)
			if err != nil {
				return err
			}

			for i, doc := range docs {
				docs[i] = map[string]interface{}{"_source": doc}
			}

			simulate := struct {
				Pipeline  ingest.Pipeline          `json:"pipeline"`
				Documents []map[string]interface{} `json:"docs"`
			}{prog, docs}

			var buf bytes.Buffer
			enc := json.NewEncoder(&buf)
			if err := enc.Encode(simulate); err != nil {
				return err
			}

			url := fmt.Sprintf("%v/_ingest/pipeline/_simulate?pretty", host)
			if verbose {
				url += "&verbose"
			}

			resp, err := http.Post(url, "application/json", &buf)
			if err != nil {
				return err
			}

			defer resp.Body.Close()
			_, err = io.Copy(os.Stdout, resp.Body)
			return err
		}),
	}
	cmdRun.PersistentFlags().StringVar(&eventFormat, "format", "plain", "event format (one of plain or json)")
	cmdRun.PersistentFlags().StringVarP(&inFile, "in", "i", "", "event input file")
	cmdRun.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Ingest node verbose execution mode")

	cmdInstall := &cobra.Command{
		Use:   "install",
		Short: "install ingest pipeline",
		Args:  cobra.MinimumNArgs(2),
		Run: func(_ *cobra.Command, args []string) {
			id := args[0]
			files := args[1:]

			gen, err := loadPipeline(files)
			if err != nil {
				log.Fatal(err)
			}

			prog, err := gen.CompileIngest()
			if err != nil {
				log.Fatal(err)
			}

			var buf bytes.Buffer
			if err := ingest.Serialize(&buf, prog); err != nil {
				log.Fatal(err)
			}

			url := fmt.Sprintf("%v/_ingest/pipeline/%v?pretty", host, id)

			req, err := http.NewRequest("PUT", url, &buf)
			if err != nil {
				log.Fatal(err)
			}
			req.Header.Set("Content-Type", "application/json")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				log.Fatal(err)
			}

			defer resp.Body.Close()
			if _, err = io.Copy(os.Stdout, resp.Body); err != nil {
				log.Fatal(err)
			}
		},
	}

	cmd := &cobra.Command{
		Use:   "ingest",
		Short: "Elasticsearch Ingest Node Mode",
	}
	cmd.AddCommand(cmdGenerate, cmdRun, cmdInstall)
	cmd.PersistentFlags().StringVar(&host, "host", "http://localhost:9200", "Ingest node URL")
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

func readEvents(format, inFile string) ([]map[string]interface{}, error) {
	type reader func(io.Reader) ([]map[string]interface{}, error)
	readers := map[string]reader{
		"plain": readPlainEvents,
		"json":  readJSONEvents,
	}

	r := readers[format]
	if r == nil {
		return nil, fmt.Errorf("format '%v' not supported", format)
	}

	var eventSource io.Reader = os.Stdin
	if inFile != "" {
		f, err := os.Open(inFile)
		if err != nil {
			log.Fatal(err)
		}

		defer f.Close()
		eventSource = f
	}
	return r(eventSource)
}

func readPlainEvents(in io.Reader) ([]map[string]interface{}, error) {
	content, err := ioutil.ReadAll(in)
	if err != nil {
		return nil, err
	}

	var events []map[string]interface{}
	for _, line := range bytes.Split(content, []byte{'\n'}) {
		if len(line) > 0 {
			events = append(events, map[string]interface{}{
				"message": string(line),
			})
		}
	}
	return events, nil
}

func readJSONEvents(in io.Reader) ([]map[string]interface{}, error) {
	var events []map[string]interface{}
	dec := json.NewDecoder(in)
	for dec.More() {
		var tmp map[string]interface{}
		if err := dec.Decode(&tmp); err != nil {
			return nil, err
		}
		events = append(events, tmp)
	}
	return events, nil
}
