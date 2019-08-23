package zgo

//  Created by Tor Langballe on /22/9/14.

//var forcingRotationForPortraitOnly = false

type PresentViewTransition int

const (
	PresentViewTransitionNone PresentViewTransition = iota
	PresentViewTransitionFromLeft
	PresentViewTransitionFromRight
	PresentViewTransitionFromTop
	PresentViewTransitionFromBottom
	PresentViewTransitionFade
	PresentViewTransitionReverse
)

func setTransition(view View, transition PresentViewTransition, screen Rect, fade float32) {
	var me = screen
	var out = me
	switch transition {
	case PresentViewTransitionFromLeft:
		out.Pos.X += -me.Max().X

	case PresentViewTransitionFromRight:
		out.Pos.X += screen.Size.W - me.Pos.X

	case PresentViewTransitionFromTop:
		out.Pos.Y += -me.Max().Y

	case PresentViewTransitionFromBottom:
		out.Pos.Y += screen.Size.H - me.Pos.Y

	default:
		break
	}
	view.Alpha(fade)
	view.GetView().Rect(out)
}

type PresentViewAttributes struct {
	DurationSecs  float64
	Transition    PresentViewTransition
	OldTransition PresentViewTransition
	DarkContent   bool
	FullArea      bool
	MakeFull      bool
	PortraitOnly  bool
	FadeToo       bool
}

var stack []PresentViewAttributes

func PresentViewAttributesNew() PresentViewAttributes {
	a := PresentViewAttributes{}
	a.DurationSecs = 0.5
	a.FullArea = true
	a.MakeFull = true
	a.PortraitOnly = true
	return a
}

func PresentViewShow(view View, attributes PresentViewAttributes, deleteOld bool, done func()) {
	mainRect := WindowGetCurrent().GetRect()
	if attributes.MakeFull {
		view.GetView().Rect(mainRect)
	} else {
		size := view.GetCalculatedSize(mainRect.Size)
		r := mainRect.Align(size, AlignmentCenter, Size{}, Size{})
		view.GetView().Rect(r)
	}
	v, _ := view.(*StackView)
	if v != nil {
		v.ArrangeChildren(nil)
	}
	AddViewToRoot(view)
	if v != nil {
		v.drawAllIfExposed()
	}
}

// func poptop(s  inout Attributes)  View? {
//     let win = UIApplication.shared.keyWindow
//     assert(stack.count > 0)
//     s = stack.last ?? Attributes()
//     ZScreen.SetStatusBarForLightContent(s.lightContent)
//     stack.removeLast()

//     return  win!.subviews.first
// }

func PresentViewPop(namedView string, animated bool, overrideDurationSecs float64, overrideTransition PresentViewTransition, done *func()) {
}

func PresentViewGetTopPushed() *CustomView {
	return nil
}

func PresentViewRecusivelyHandleActivation(activated bool) {
	if activated {
	}
}

// private func setFocusInView(view ZContainerView) {
//     view.setNeedsFocusUpdate()

//     view.RangeChildren(subViews true) { (view) in
//         if let v = view as? ZCustomView {
//             if v.canFocus {
//                 view.Focus()
//                 return false
//             }
//         }
//         return true
//     }
// }
