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
	PresentViewTransitionSame
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
	DurationSecs             float64
	Transition               PresentViewTransition
	OldTransition            PresentViewTransition
	DarkContent              bool
	MakeFull                 bool
	PortraitOnly             bool
	FadeToo                  bool
	DeleteOld                bool
	Modal                    bool
	Title                    string
	Pos                      *zgeo.Pos
	ModalCloseOnOutsidePress bool
}

var stack []PresentViewAttributes

func PresentViewAttributesNew() PresentViewAttributes {
	a := PresentViewAttributes{}
	a.DurationSecs = 0.5
	a.MakeFull = false
	a.PortraitOnly = false
	return a
}

func presentViewCallReady(v View) {
	o := v.(NativeViewOwner)
	if o != nil {
		nv := o.GetNative()
		if nv.Presented {
			return
		}
		nv.Presented = true
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

func PresentViewShow(v View, attributes PresentViewAttributes, presented func(win *Window), closed func()) {
	presentViewPresenting = true
	presentViewCallReady(v)
	ct, _ := v.(ContainerType)
	if ct != nil {
		WhenContainerLoaded(ct, func(waited bool) {
			presentLoaded(v, attributes, presented, closed)
		})
	} else {
		presentLoaded(v, attributes, presented, closed)
	}
}

var firstPresented bool

func presentLoaded(v View, attributes PresentViewAttributes, presented func(win *Window), closed func()) {
	// zlog.Info("PresentViewShow", v.ObjectName())
	win := WindowGetCurrent()

	fullRect := win.Rect()
	rect := fullRect

	size := v.CalculatedSize(rect.Size)
	if attributes.Modal || firstPresented {
		rect = rect.Align(size, zgeo.Center, zgeo.Size{}, zgeo.Size{})
	}
	if attributes.Modal {
		ct, _ := v.(ContainerType)
		if ct != nil {
			v.SetBGColor(zgeo.ColorNewGray(0.95, 1))
			v.SetCorner(5)
		}
		nv := ViewGetNative(v)
		if nv != nil {
			nv.SetDropShadow(zgeo.DropShadow{Delta: zgeo.Size{4, 4}, Blur: 8, Color: zgeo.ColorNewGray(0.2, 1)})
			g := ContainerViewNew(nil, "$blocker")
			fullRect.Pos = zgeo.Pos{}
			g.SetRect(fullRect)
			g.SetBGColor(zgeo.ColorNewGray(0, 0.5))
			if attributes.Pos != nil {
				g.Add(zgeo.TopLeft, v, attributes.Pos.Size())
			} else {
				g.Add(zgeo.Center, v)
			}
			g.ArrangeChildren(nil)
			if attributes.ModalCloseOnOutsidePress {
				g.SetPressedHandler(func() {
					PresentViewPop(v, closed)
				})
			}
			v = g
			win.AddView(g)
		}
	} else {
		if firstPresented {
			size.H += WindowBarHeight
			win = WindowOpenWithURL("about:blank", size, &rect.Pos)
			if attributes.Title != "" {
				win.SetTitle(attributes.Title)
			}
			if closed != nil {
				win.SetHandleClosed(closed)
			}
		}
		v.SetRect(zgeo.RectFromSize(rect.Size))
		win.AddView(v)
	}
	firstPresented = true

	// cvt, _ := v.(ContainerViewType)
	// if cvt != nil {
	// 	cvt.ArrangeChildren(nil)
	// }
	// NativeViewAddToRoot(v)
	presentViewPresenting = false
	et, _ := v.(ExposableType)
	if et != nil {
		et.drawIfExposed()
	}
	if presented != nil {
		presented(win)
	}
}

func PresentViewPop(view View, done func()) {
	PresentViewPopOverride(view, PresentViewAttributes{}, done)
}

func PresentViewPopOverride(view View, overrideAttributes PresentViewAttributes, done func()) {
	// TODO: Handle non-modal window too
	parent := ViewGetNative(view).Parent()
	if parent.ObjectName() == "$blocker" {
		view = parent
	}
	ViewGetNative(view).RemoveFromParent()
	if done != nil {
		done()
	}
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
