package zgo

//  Created by Tor Langballe on /2/11/15.

type Label struct {
	TextBaseHandler
	ViewBaseHandler
	MinWidth  float64
	MaxWidth  float64
	MaxHeight *float64
	Margin    Rect
	pressed   func(pos Pos)
}

func (v *Label) GetCalculatedSize(total Size) Size {
	return v.TextBaseHandler.GetCalculatedSize(total)
}
