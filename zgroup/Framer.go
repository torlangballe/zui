package zgroup

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
	Margin:        zgeo.RectFromXY2(8, 9, -8, -8),
}

var DefaultFrameTitleStyling = zstyle.Styling{
	FGColor: zstyle.DefaultFGColor().WithOpacity(0.7),
	Font:    *zgeo.FontNice(zgeo.FontDefaultSize, zgeo.FontStyleBold),
}

func MakeStackTitledFrame(stack *zcontainer.StackView, title string, titleOnFrame bool, styling, titleStyling zstyle.Styling) {
	s := DefaultFrameStyling.MergeWith(styling)
	// zlog.Info("NewTitledFrame1:", title, s.Margin, styling.Margin, styling.StrokeWidth, styling.StrokeColor)
	fs := s
	fs.Font = zgeo.Font{}
	stack.SetStyling(fs)
	if title != "" {
		header := zcontainer.StackViewHor("header")
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
}
