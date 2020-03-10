package zui

import (
	"github.com/torlangballe/zutil/zgeo"
)

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

func setTransition(n *NativeView, transition PresentViewTransition, screen zgeo.Rect, fade float32) {
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
	n.SetAlpha(fade)
	n.SetRect(out)
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
	DeleteOld     bool
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

func presentViewCallReady(v View) {
	o := v.(NativeViewOwner)
	if o != nil {
		nv := o.GetNative()
		if nv.presented {
			return
		}
		nv.presented = true
	}
	r, got := v.(ReadyToShowType)
	if got {
		r.ReadyToShow()
	}
	ct, got := v.(ContainerType)
	if got {
		for _, c := range ct.GetChildren() {
			presentViewCallReady(c)
		}
	}
}

var presentViewPresenting = true

func PresentViewShow(v View, attributes PresentViewAttributes, done func()) {
	presentViewPresenting = true
	mainRect := WindowGetCurrent().Rect()
	presentViewCallReady(v)
	ct := v.(ContainerType)

	// fmt.Println("PresentViewShow", v.ObjectName())

	WhenContainerLoaded(ct, func(waited bool) {
		// fmt.Println("PresentViewShow loaded", v.ObjectName())
		if attributes.MakeFull {
			// fmt.Println("Present:", mainRect, presentViewPresenting)
			v.SetRect(mainRect)
		} else {
			size := v.CalculatedSize(mainRect.Size)
			r := mainRect.Align(size, zgeo.Center, zgeo.Size{}, zgeo.Size{})
			v.SetRect(r)
			v.SetBGColor(zgeo.ColorNewGray(0.8, 1))
			v.SetCorner(10)
			no := v.(NativeViewOwner)
			if no != nil {
				no.GetNative().SetDropShadow(zgeo.Size{4, 4}, 8, zgeo.ColorBlack)
			}
		}
		// cvt, _ := v.(ContainerViewType)
		// if cvt != nil {
		// 	cvt.ArrangeChildren(nil)
		// }
		NativeViewAddToRoot(v)
		presentViewPresenting = false
		et, _ := v.(ExposableType)
		if et != nil {
			et.drawIfExposed()
		}
		if done != nil {
			done()
		}
	})
}

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
