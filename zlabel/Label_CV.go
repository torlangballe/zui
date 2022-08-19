//go:build zui

package zlabel

//  Created by Tor Langballe on /2/11/15.
/*
type LabelCV struct {
	// NativeView
	// LongPresser
	CustomView

	minWidth  float64
	maxWidth  float64
	maxLines  int
	margin    zgeo.Rect
	alignment zgeo.Alignment
	text      string

	// pressed     func()
	// longPressed func()

	Columns int
}

func LabelCVNew(text string) *LabelCV {
	v := &LabelCV{}
	v.Init(v, zstr.Head(text, 100))
	v.text = text
	v.SetObjectName(text)
	v.SetMaxLines(1)
	v.alignment = zgeo.CenterLeft
	v.SetFont(zgeo.FontNice(zgeo.FontDefaultSize, zgeo.FontStyleNormal))
	v.SetDrawHandler(func(rect zgeo.Rect, canvas *zcanvas.Canvas, view zview.View) {
		ti := v.GetTextInfo()
		ti.Rect = v.LocalRect()
		ti.Draw(canvas)
	})
	return v
}

func (v *LabelCV) SetText(text string) {
	v.text = text
	v.Expose()
}

func (v *LabelCV) Text() string {
	return v.text
}

func (v *LabelCV) GetTextInfo() ztextinfo.Info {
	t := ztextinfo.New()
	t.Alignment = v.alignment
	t.Font = v.Font()
	t.Color = v.Color()
	t.Text = v.Text()
	if v.maxWidth != 0 {
		t.SetWidthFreeHight(v.maxWidth)
	}
	t.MaxLines = v.maxLines
	return *t
}

func (v *LabelCV) CalculatedSize(total zgeo.Size) zgeo.Size {
	var s zgeo.Size
	to := v.View.(ztextinfo.Owner)
	ti := to.GetTextInfo()
	if v.Columns != 0 {
		s = ti.GetColumnsSize(v.Columns)
	} else {
		s, _, _ = ti.GetBounds()
	}
	s.Add(v.margin.Size.Negative())
	zfloat.Maximize(&s.W, v.minWidth)
	if v.maxWidth != 0 {
		zfloat.Minimize(&s.W, v.maxWidth)
	}
	s.Maximize(v.minSize)
	s = s.Ceil()
	// if v.ObjectName() == "thumbSetTime" {
	// 	zlog.Info("CS:", s, v.Columns, v.maxWidth, v.minWidth, ti.Text)
	// }
	return s
}

func (v *LabelCV) IsMinimumOneLineHight() bool {
	return v.maxLines > 0
}

func (v *LabelCV) TextAlignment() zgeo.Alignment {
	return v.alignment
}

func (v *LabelCV) MinWidth() float64 {
	return v.minWidth
}

func (v *LabelCV) MaxWidth() float64 {
	return v.maxWidth
}

func (v *LabelCV) MaxLines() int {
	return v.maxLines
}

func (v *LabelCV) SetMinWidth(min float64) {
	v.minWidth = min
}

func (v *LabelCV) SetMaxWidth(max float64) {
	v.maxWidth = max
}

func LabelizeCV(view zview.View, prefix string, minWidth float64, alignment zgeo.Alignment) (label *LabelCV, stack *zcontainer.StackView, viewCell *zcontainer.Cell) {
	font := zgeo.FontNice(zgeo.FontDefaultSize, zgeo.FontStyleBold)
	to, _ := view.(ztextinfo.Owner)
	if to != nil {
		ti := to.GetTextInfo()
		font = ti.Font
		font.Style = zgeo.FontStyleBold
	}
	title := prefix
	checkBox, isCheck := view.(*CheckBox)
	if checkBox != nil && alignment&zgeo.Right != 0 {
		title = ""
		_, cstack := checkBox.Labelize(prefix)
		view = cstack
		alignment = alignment.FlippedHorizontal()
	}
	label = LabelCVNew(title)
	label.SetObjectName("$labelize.label " + prefix)
	label.SetTextAlignment(zgeo.Right)
	label.SetFont(font)
	label.SetColor(zstyle.DefaultFGColor().WithOpacity(0.7))
	stack = StackViewHor("$labelize." + prefix) // give it special name so not easy to mis-search for in recursive search

	stack.AddView(label, zgeo.CenterLeft).MinSize.W = minWidth
	marg := zgeo.Size{}
	if isCheck {
		marg.W = -6 // in html cell has a box around it of 20 pixels
	}
	viewCell = stack.Add(view, alignment, marg)
	return
}

func (v *LabelCV) SetMaxLines(max int) {
	// zlog.Info("Label.SetMaxLines:", max, v.ObjectName())
	v.maxLines = max
	v.Expose()
}

func (v *LabelCV) SetTextAlignment(a zgeo.Alignment) {
	v.alignment = a
	v.Expose()
}

func (v *LabelCV) SetMargin(m zgeo.Rect) {
	v.margin = m
	v.Expose()
}
*/
