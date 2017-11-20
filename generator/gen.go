package generator

import (
	"errors"
	"io"

	"github.com/urso/bpb/prog/ingest"
	"github.com/urso/bpb/prog/ls"

	"github.com/elastic/beats/libbeat/common"
)

type Generator struct {
	ID          string
	Description string
	Processors  []Processor
}

func New(descr string, processors []*common.Config) (*Generator, error) {
	if len(processors) == 0 {
		return nil, errors.New("no processors")
	}

	ps, err := LoadAll(processors)
	if err != nil {
		return nil, err
	}

	return &Generator{Description: descr, Processors: ps}, nil
}

func (g *Generator) MakeIngest(out io.Writer) error {
	prog, err := g.compileIngest()
	if err != nil {
		return err
	}

	return ingest.Serialize(out, prog)
}

func (g *Generator) MakeLogstash(out io.Writer) error {
	prog, err := g.compileLogstash()
	if err != nil {
		return err
	}

	if g.ID != "" {
		prog.MetaPipeline = g.ID
	}

	return ls.Serialize(out, prog)
}

func (g *Generator) compileIngest() (ingest.Pipeline, error) {
	pipeline := ingest.Pipeline{
		Description: g.Description,
	}

	processors, err := CompileIngestProcessors(g.Processors)
	if err != nil {
		return pipeline, err
	}

	pipeline.Processors = processors
	if len(pipeline.OnFailure) == 0 {
		pipeline.OnFailure = ingest.MakeSingleProcessor("set", map[string]interface{}{
			"field": "error.message",
			"value": "{{ _ingest.on_failure_message }}",
		})
	}

	return pipeline, nil
}

func (g *Generator) compileLogstash() (ls.Pipeline, error) {
	pipeline := ls.Pipeline{
		Description: g.Description,
	}

	processors, err := CompileLogstashProcessors(g.Processors)
	if err != nil {
		return pipeline, err
	}

	pipeline.Block = processors
	return pipeline, nil
}

func CompileIngestProcessors(input []Processor) ([]ingest.Processor, error) {
	if len(input) == 0 {
		return nil, nil
	}

	var processors []ingest.Processor
	for _, gen := range input {
		ps, err := gen.CompileIngest()
		if err != nil {
			return nil, err
		}

		processors = append(processors, ps...)
	}

	return processors, nil
}

func CompileLogstashProcessors(input []Processor) (ls.Block, error) {
	if len(input) == 0 {
		return nil, nil
	}

	var blk ls.Block
	for _, gen := range input {
		sub, err := gen.CompileLogstash()
		if err != nil {
			return nil, err
		}

		blk = append(blk, sub...)
	}

	return blk, nil
}
