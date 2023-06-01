package zwidgets

import (
	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zlabel"
	"github.com/torlangballe/zui/zstyle"
	"github.com/torlangballe/zutil/zbool"
	"github.com/torlangballe/zutil/zgeo"
)

var DefaultFrameStyling = zstyle.Styling{
	StrokeWidth:   2,
	StrokeColor:   zstyle.DefaultFGColor().WithOpacity(0.5),
	Corner:        5,
	StrokeIsInset: zbool.True,
	Margin:        zgeo.RectFromXY2(8, 13, -8, -8),
}

var DefaultFrameTitleStyling = zstyle.Styling{
	FGColor: zstyle.DefaultFGColor().WithOpacity(0.7),
	Font:    *zgeo.FontNice(zgeo.FontDefaultSize, zgeo.FontStyleBold),
}

func MakeStackATitledFrame(stack *zcontainer.StackView, title string, titleOnFrame bool, styling, titleStyling zstyle.Styling) (header *zcontainer.StackView) {
	s := DefaultFrameStyling.MergeWith(styling)
	fs := s
	fs.Font = zgeo.Font{}
	stack.SetStyling(fs)
	if title != "" {
		header = zcontainer.StackViewHor("header")
		header.SetSpacing(2)
		h := -8.0
		if titleOnFrame {
			h = -(s.Margin.Min().Y + zgeo.FontDefaultSize - 4)
			header.SetCorner(4)
			header.SetBGColor(zgeo.ColorWhite)
		}
		stack.AddAdvanced(header, zgeo.TopLeft|zgeo.HorExpand, zgeo.Size{0, h}, zgeo.Size{}, 0, false)
		label := zlabel.New(title)
		label.SetObjectName("title")
		ts := DefaultFrameTitleStyling.MergeWith(titleStyling)
		label.SetStyling(ts)
		header.Add(label, zgeo.CenterLeft, zgeo.Size{})
	}
	return header
}
