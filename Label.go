package zgo

import "fmt"

//  Created by Tor Langballe on /2/11/15.

type Label struct {
	TextBase
	View
	MinWidth  float64
	MaxWidth  float64
	MaxHeight *float64
	Margin    Rect
	//    var touchInfo = ZTouchInfo()
}

func (v *Label) GetCalculatedSize(total Size) Size {
	fmt.Println("label calc size")
	b := v.TextBase.(*TextBaseHandler)
	return b.GetCalculatedSize(total)
}
