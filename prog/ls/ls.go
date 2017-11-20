package ls

import (
	"fmt"
	"io"
	"strings"
)

type Pipeline struct {
	MetaPipeline string
	Description  string
	Block        Block
}

func IgnoreMissing(field string, blk Block) Block {
	expr := Expression(NormalizeField(field))
	return MakeBlock(Conditional{
		Cond: []Case{
			{expr, blk},
		},
	})
}

func NormalizeField(field string) string {
	sub := strings.Split(field, ".")
	for i := range sub {
		sub[i] = "[" + sub[i] + "]"
	}
	return strings.Join(sub, "")
}

func Serialize(out io.Writer, p Pipeline) error {
	ctx := &formatCtx{
		out:    out,
		indent: "    ",
	}

	blk := p.Block
	if p.MetaPipeline != "" {
		params := Params{}
		params.RemoveField("@metadate.pipeline")
		blk = append(MakeBlock(MakeFilter("mutate", params)), blk...)

		expr := fmt.Sprintf("[@metadate][pipeline] == \"%v\"", p.MetaPipeline)
		blk = MakeBlock(
			Conditional{
				Cond: []Case{
					{Expression(expr), blk},
				},
			},
		)
	}

	if p.Description != "" {
		for _, line := range strings.Split(p.Description, "\n") {
			ctx.Println("# " + line)
		}
	}

	ctx.Println("filter {")
	ctx.withIndent(blk.format)
	return ctx.Println("}")
}
