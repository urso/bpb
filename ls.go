package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/urso/bpb/generator"
)

func cmdLogstash() *cobra.Command {
	var (
		pipelineID string
		verbose    bool
		noError    bool
	)

	cmdGenerate := &cobra.Command{
		Use:   "generate",
		Short: "Generate logstash filter configuration",
		Run: runWithPipeline(func(gen *generator.Generator) error {
			gen.ID = pipelineID
			ctx := &generator.LogstashCtx{
				Verbose:       verbose,
				DisableErrors: noError,
			}
			return gen.MakeLogstash(os.Stdout, ctx)
		}),
	}

	var (
		lsHome      string
		inFile      string
		eventFormat string
	)
	cmdRun := &cobra.Command{
		Use:   "run",
		Short: "Run pipeline",
		Run: runWithPipeline(func(gen *generator.Generator) error {
			ctx := &generator.LogstashCtx{
				Verbose:       verbose,
				DisableErrors: noError,
			}
			return lsRun(lsHome, gen, ctx, inFile, eventFormat)
		}),
	}
	cmdRun.PersistentFlags().StringVar(&lsHome, "lshome", "", "logstash home path")
	cmdRun.PersistentFlags().StringVarP(&inFile, "in", "i", "", "event input file")
	cmdRun.PersistentFlags().StringVar(&eventFormat, "format", "plain", "event format (one of plain or json)")

	cmd := &cobra.Command{
		Use:   "logstash",
		Short: "Logstash Mode",
	}
	cmd.PersistentFlags().StringVar(&pipelineID, "id", "", "pipeline ID")
	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose mode - create debug prints on each filter")
	cmd.PersistentFlags().BoolVar(&noError, "noerr", false, "disable filter error handling")
	cmd.AddCommand(cmdGenerate, cmdRun)
	return cmd
}

func lsRun(
	lsHome string,
	gen *generator.Generator,
	ctx *generator.LogstashCtx,
	inFile, eventFormat string,
) error {
	var err error
	var lsBin string

	eventReader, err := findEventReader(eventFormat)
	if err != nil {
		return err
	}

	if lsHome == "" {
		lsBin, err = exec.LookPath("logstash")
		if err != nil {
			return err
		}
	} else {
		lsBin = filepath.Join(lsHome, "bin", "logstash")
	}

	eventsFile, err := os.Open(inFile)
	if err != nil {
		return err
	}
	defer eventsFile.Close()

	confFile, err := ioutil.TempFile("", "lstestconf")
	if err != nil {
		return err
	}
	confFileName := confFile.Name()
	defer os.Remove(confFileName)
	defer confFile.Close()

	// serialize test config:
	if _, err := io.WriteString(confFile, `input { stdin { codec => json } }`); err != nil {
		return err
	}
	if err := gen.MakeLogstash(confFile, ctx); err != nil {
		return err
	}
	if _, err := io.WriteString(confFile, `output { stdout { codec => rubydebug { metadata => true } } }`); err != nil {
		return err
	}
	if err := confFile.Sync(); err != nil {
		return err
	}

	// start logstash:
	cmd := exec.Command(lsBin, "-f", confFileName)
	eventRead, eventWrite, err := os.Pipe()
	if err != nil {
		return err
	}
	defer eventRead.Close()

	// setup ls IO
	cmd.Stdin = eventRead
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// start event writer:
	go func() {
		defer eventWrite.Close()

		enc := json.NewEncoder(eventWrite)
		err := eventReader(eventsFile, func(event map[string]interface{}) error {
			return enc.Encode(event)
		})
		if err != nil {
			log.Printf("Copying event error: %v", err)
		}
	}()

	// run logstash + wait for exit
	return cmd.Run()
}
