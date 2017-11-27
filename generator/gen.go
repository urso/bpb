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

type Processor interface {
	Name() string
	CompileIngest() ([]ingest.Processor, error)
	CompileLogstash(ctx *LogstashCtx) (FilterBlock, error)
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
	prog, err := g.CompileIngest()
	if err != nil {
		return err
	}

	return ingest.Serialize(out, prog)
}

func (g *Generator) MakeLogstash(out io.Writer, ctx *LogstashCtx) error {
	prog, err := g.CompileLogstash(ctx)
	if err != nil {
		return err
	}

	if ctx.Verbose {
		prog.Block = append(ls.MakeBlock(ls.MakePrintEventDebug("init")), prog.Block...)
		prog.Block = append(prog.Block, ls.MakePrintEventDebug("emit"))
	}

	if g.ID != "" {
		prog.MetaPipeline = g.ID
	}

	return ls.Serialize(out, prog)
}

func (g *Generator) CompileIngest() (ingest.Pipeline, error) {
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

func (g *Generator) CompileLogstash(ctx *LogstashCtx) (ls.Pipeline, error) {
	pipeline := ls.Pipeline{
		Description: g.Description,
	}

	onError := MakeLSErrorReporter(ctx)
	processors, err := CompileLogstashProcessors(ctx, onError, g.Processors)
	if err != nil {
		return pipeline, err
	}

	pipeline.Block = processors.Block
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
