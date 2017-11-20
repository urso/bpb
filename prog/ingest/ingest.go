package ingest

import (
	"encoding/json"
	"io"
)

type Pipeline struct {
	Description string      `json:"description"`
	Processors  []Processor `json:"processors"`
	OnFailure   []Processor `json:"on_failure"`
}

type Processor map[string]map[string]interface{}

func MakeProcessor(name string, params map[string]interface{}) Processor {
	return Processor{name: params}
}

func Single(p Processor) []Processor {
	return []Processor{p}
}

func MakeSingleProcessor(name string, params map[string]interface{}) []Processor {
	return Single(MakeProcessor(name, params))
}

func Serialize(out io.Writer, p Pipeline) error {
	enc := json.NewEncoder(out)
	enc.SetIndent("", "    ")
	return enc.Encode(p)
}

func RemoveField(name string) Processor {
	return MakeProcessor("remove", map[string]interface{}{
		"field": name,
	})
}
