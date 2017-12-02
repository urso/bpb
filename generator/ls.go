package generator

import (
	"fmt"
	"strings"

	"github.com/urso/bpb/prog/ls"
)

type LogstashCtx struct {
	Verbose       bool
	DisableErrors bool

	tagCount uint
}

type FilterBlock struct {
	Block       ls.Block
	FailureTags []string
}

func (ctx *LogstashCtx) CreateTag(name string) string {
	ctx.tagCount++
	if name == "" {
		name = "_logstash_tag"
	}
	return fmt.Sprintf("%v_%v", name, ctx.tagCount)
}

func (b *FilterBlock) Append(b2 FilterBlock) {
	b.AppendBlock(b2.Block)
	b.AddTags(b2.FailureTags...)
}

func (b *FilterBlock) AppendBlock(blk ls.Block) {
	b.Block = append(b.Block, blk...)
}

func (b *FilterBlock) AddTags(tags ...string) {
outter:
	for _, t := range tags {
		// ignore duplicate tags
		for _, active := range b.FailureTags {
			if active == t {
				continue outter
			}
		}

		b.FailureTags = append(b.FailureTags, t)
	}
}

func (b *FilterBlock) AddFilter(f ls.Filter) {
	b.Block = append(b.Block, f)
}

func CompileLogstashProcessors(
	ctx *LogstashCtx,
	onError func(processorName string, tags []string) FilterBlock,
	input []Processor,
) (FilterBlock, error) {
	if len(input) == 0 {
		return FilterBlock{}, nil
	}

	// compile processors to individual blocks
	blks := make([]FilterBlock, len(input))
	for i, gen := range input {
		blk, err := gen.CompileLogstash(ctx)
		if err != nil {
			return FilterBlock{}, err
		}

		// filter empty tags
		if tags := blk.FailureTags; len(tags) > 0 {
			tmp := make([]string, 0, len(tags))
			for _, t := range tags {
				if t != "" {
					tmp = append(tmp, t)
				}
			}
			blk.FailureTags = tmp
		}

		blks[i] = blk
	}

	// create conditional per block that can fail
	conds := make([]ls.Conditional, len(blks))
	var failTags []string
	if !ctx.DisableErrors {
		for i, blk := range blks {
			failCond := makeLSFailTagsCondition(blk.FailureTags)
			if failCond == "" {
				continue
			}

			errBlk := onError(input[i].Name(), blk.FailureTags)
			failTags = append(failTags, errBlk.FailureTags...)

			onFail := ls.MakeBlock(ls.MakeFilter("mutate", ls.Params{
				"remove_tag": blk.FailureTags,
			}))
			onFail = append(onFail, errBlk.Block...)

			conds[i] = ls.Conditional{
				Cond: []ls.Case{
					{
						Cond:  ls.Expression(failCond),
						Block: onFail,
					},
				},
			}
		}
	}

	// link conditionals and execution blocks bottom-up
	var active ls.Block
	for i := len(blks) - 1; i >= 0; i-- {
		if len(conds[i].Cond) == 0 {
			active = append(blks[i].Block, active...)
			continue
		}

		cond := conds[i]
		cond.Else = active

		active = ls.MakeBlock(blks[i].Block...)
		active = append(active, cond)
	}

	return FilterBlock{
		Block:       active,
		FailureTags: failTags,
	}, nil
}

func makeLSFailTagsCondition(tags []string) string {
	if len(tags) == 0 {
		return ""
	}

	cmps := make([]string, len(tags))
	for i, tag := range tags {
		cmps[i] = fmt.Sprintf(`("%v" in [tags])`, tag)
	}

	return strings.Join(cmps, `or`)
}

// TODO: add support for versioning, tag_on_exception not available in 6.0 yet
func MakeRuby(ctx *LogstashCtx, code, failureTag string, extra ls.Params) ls.Block {
	params := ls.Params{
		"code": code,
	}

	params.RemoveTag(failureTag)
	for k, v := range extra {
		params[k] = v
	}

	blk := ls.MakeBlock()
	if failureTag != "" {
		blk = append(blk, ls.MakeFilter("mutate", ls.Params{
			"add_tag": []string{failureTag},
		}))
	}

	return append(blk, ls.MakeFilter("ruby", params))
}

func MakeLSErrorReporter(ctx *LogstashCtx) func(string, []string) FilterBlock {
	return func(filter string, tags []string) FilterBlock {
		msg := fmt.Sprintf(`filter %v (tags: %v) failed`, filter, tags)
		code := fmt.Sprintf(`msg='%v'; field='[error][message]'; old=event.get(field); event.set(field, old ? [event.get(field), msg].join(' : ') : msg)`, msg)
		return FilterBlock{
			Block: MakeRuby(ctx, code, "", nil),
		}
	}
}
