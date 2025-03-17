//go:build zui

package zlabel

import (
	"fmt"

	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zstyle"
	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zstr"
)

type StyledTextBuilder struct {
	Default  zstyle.TextStyle
	Singles  bool
	Stack    *zcontainer.StackView
	info     info
	hasAdded bool
}

type info struct {
	reset zstyle.TextStyle
	cs    zstyle.TextStyle
	text  string
}

func NewStyledTextBuilder() *StyledTextBuilder {
	b := &StyledTextBuilder{}
	b.Singles = true
	b.Default.Size = zgeo.FontDefaultSize
	b.Default.Name = zgeo.FontDefaultName
	b.Default.Style = zgeo.FontStyleNormal
	b.Default.Color = zstyle.DefaultFGColor()
	b.Default.Gap = 4
	return b
}

func (b *StyledTextBuilder) AddLabelsToHorStack(parameters ...any) {
	b.info.reset = b.Default
	b.info.cs = b.info.reset
	for _, p := range parameters {
		// zlog.Info("AddLabelsToStack", p)
		tstart, is := p.(zstyle.Control)
		if is {
			switch tstart {
			case zstyle.NoOp:
				continue
			case zstyle.Start:
				b.info.reset = b.info.cs
			}
			continue
		}
		tstyle, is := p.(zstyle.TextStyle)
		if is {
			b.outputLabel()
			b.info.cs.Set(tstyle)
			continue
		}
		fsize, is := p.(zstyle.Size)
		if is {
			b.outputLabel()
			b.info.cs.Size = float64(fsize)
			continue
		}
		fname, is := p.(zstyle.Name)
		if is {
			b.outputLabel()
			b.info.cs.Name = string(fname)
			continue
		}
		finc, is := p.(zstyle.SizeInc)
		if is {
			b.outputLabel()
			b.info.cs.Size = zgeo.FontDefaultSize + float64(finc)
			continue
		}
		fgap, is := p.(zstyle.Gap)
		if is {
			b.outputLabel()
			b.info.cs.Gap = float64(fgap)
			continue
		}
		fstyle, is := p.(zgeo.FontStyle)
		if is {
			b.outputLabel()
			b.info.cs.Style = fstyle
			continue
		}
		col, is := p.(zgeo.Color)
		if is {
			b.outputLabel()
			b.info.cs.Color = col
			continue
		}
		if b.Singles && b.info.text != "" {
			b.outputLabel()
		}
		b.info.text = zstr.Spaced(b.info.text, fmt.Sprint(p))
		// zlog.Info("AddLabelsToStack2", b.info.text)
	}
	b.outputLabel()
}

func (b *StyledTextBuilder) outputLabel() {
	if b.info.text != "" {
		// zlog.Info("AddLabelsToStack3", i)
		label := New(b.info.text)
		font := zgeo.FontNew(b.info.cs.Name, b.info.cs.Size, b.info.cs.Style)
		label.SetFont(font)
		label.SetColor(b.info.cs.Color)
		gap := 0.0
		if b.hasAdded {
			if b.info.cs.Gap != zfloat.Undefined {
				gap = b.info.cs.Gap
			}
		}
		// zlog.Info("Add", b.info.text, "gap:", gap, b.Stack.Spacing())
		b.Stack.Add(label, zgeo.CenterLeft, zgeo.SizeD(gap, 0))
		b.info.cs = b.info.reset
		b.info.text = ""
	}
}

func (b *StyledTextBuilder) AddLabelsRowToVertStack(vstack *zcontainer.StackView, parameters ...any) *zcontainer.StackView {
	bRow := *b
	bRow.Stack = zcontainer.StackViewHor("row")
	bRow.Stack.SetSpacing(0)
	bRow.AddLabelsToHorStack(parameters...)
	vstack.Add(bRow.Stack, zgeo.TopLeft|zgeo.HorExpand)
	return bRow.Stack
}
