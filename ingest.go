package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/spf13/cobra"
	"github.com/urso/bpb/generator"
	"github.com/urso/bpb/prog/ingest"
)

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
			return ingestRun(gen, host, verbose, inFile, eventFormat)
		}),
	}
	cmdRun.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Ingest node verbose execution mode")
	cmdRun.PersistentFlags().StringVarP(&inFile, "in", "i", "", "event input file")
	cmdRun.PersistentFlags().StringVar(&eventFormat, "format", "plain", "event format (one of plain or json)")

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
			if err := ingestInstall(host, id, gen); err != nil {
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

func ingestRun(
	gen *generator.Generator,
	host string,
	verbose bool,
	inFile string,
	eventFormat string,
) error {
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
}

func ingestInstall(
	host string,
	id string,
	gen *generator.Generator,
) error {
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
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if _, err = io.Copy(os.Stdout, resp.Body); err != nil {
		return err
	}

	return nil
}
