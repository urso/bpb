package ls

import "github.com/urso/go-structform/gotype"

// statement types
type (
	Statement interface {
		format(ctx *formatCtx) error
	}

	Block []Statement

	Filter struct {
		Name   string
		Params Params
	}

	Conditional struct {
		Cond []Case
		Else Block
	}
)

type Expression string

type Case struct {
	Cond  Expression
	Block Block
}

type Params map[string]interface{}

func MakeBlock(stmts ...Statement) Block {
	return stmts
}

func MakeVerboseBlock(verbose bool, name string, stmts ...Statement) Block {
	blk := Block(stmts)
	if verbose {
		blk = append(blk, MakePrintEventDebug(name))
	}
	return blk
}

func MakeFilter(name string, params Params) Filter {
	return Filter{name, params}
}

func (p Params) Target(field string) {
	if field != "" {
		p["target"] = NormalizeField(field)
	}
}

func (p Params) DropField(drop bool, name string) {
	if drop {
		p.RemoveField(name)
	}
}

func (p Params) RemoveField(name string) {
	p["remove_field"] = []string{NormalizeField(name)}
}

func (b Block) format(ctx *formatCtx) error {
	for _, stmt := range b {
		if err := stmt.format(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (c Conditional) format(ctx *formatCtx) error {
	ifClause := c.Cond[0]
	elifClauses := c.Cond[1:]

	ctx.Printf("if %v {\n", ifClause.Cond)
	ctx.withIndent(ifClause.Block.format)

	for _, clause := range elifClauses {
		ctx.Printf("} elif %v {\n", clause.Cond)
		ctx.withIndent(clause.Block.format)
	}

	if len(c.Else) > 0 {
		ctx.Println("} else {")
		ctx.withIndent(c.Else.format)
	}

	return ctx.Println("}")
}

func (f Filter) format(ctx *formatCtx) error {
	if len(f.Params) == 0 {
		return ctx.Printf("%v {}", f.Name)
	}

	ctx.Printf("%v ", f.Name)
	return gotype.Fold(f.Params, newParamPrinter(ctx))
}
