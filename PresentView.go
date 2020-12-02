package zui

import (
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zstr"
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
	WindowOptions
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

func presentViewCallReady(v View, beforeWindow bool) {
	nv := ViewGetNative(v)
	if nv == nil {
		return
	}
	if !nv.Presented {
		if !beforeWindow {
			nv.Presented = true
		}
		r, _ := v.(ReadyToShowType)
		if r != nil {
			r.ReadyToShow(beforeWindow)
		}
	}
	if nv.allChildrenPresented {
		return
	}
	ct, _ := v.(ContainerType)
	if ct != nil {
		// zlog.Info("presentViewCallReady1:", v.ObjectName(), len(ct.GetChildren()))
		for _, c := range ct.GetChildren() {
			presentViewCallReady(c, beforeWindow)
		}
	}
	if !beforeWindow {
		nv.allChildrenPresented = true
	}
}

var presentViewPresenting = true

func PresentView(v View, attributes PresentViewAttributes, presented func(win *Window), closed func()) {
	presentViewPresenting = true
	presentViewCallReady(v, true)

	// ct, _ := v.(ContainerType)
	// if ct != nil {
	// 	WhenContainerLoaded(ct, func(waited bool) {
	// 		presentLoaded(v, attributes, presented, closed)
	// 	})
	// } else {
	presentLoaded(v, attributes, presented, closed)
	// }
}

var firstPresented bool

func presentLoaded(v View, attributes PresentViewAttributes, presented func(win *Window), closed func()) {
	// zlog.Info("PresentView", v.ObjectName())
	win := WindowGetMain()

	fullRect := win.ContentRect()
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
				lp, _ := v.(Pressable)
				zlog.Info("LP:", lp != nil, v.ObjectName())
				if lp != nil {
					lp.SetPressedHandler(func() {
						zlog.Info("LP Pressed")
					})
				}
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
			//			o := WindowOptions{URL: "about:blank", Pos: &rect.Pos, Size: size, ID: attributes.WindowID}
			o := attributes.WindowOptions
			o.Pos = &rect.Pos
			o.Size = size
			// zlog.Info("PresentView:", rect.Pos, size, attributes.ID)
			win = WindowOpen(o)
			if attributes.Title != "" {
				win.SetTitle(attributes.Title)
			}
			if closed != nil {
				win.HandleClosed = closed
			}
		}
		v.SetRect(zgeo.Rect{Size: rect.Size})
		win.AddView(v)
		win.setOnResize()
	}
	firstPresented = true

	// cvt, _ := v.(ContainerViewType)
	// if cvt != nil {
	// 	cvt.ArrangeChildren(nil)
	// }
	// NativeViewAddToRoot(v)
	presentViewCallReady(v, false)
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
	nv := ViewGetNative(view)
	if view.ObjectName() == "$blocker" {
		ct, _ := view.(ContainerType)
		if ct != nil {
			for _, c := range ct.GetChildren() {
				zlog.Info("PresentViewPopOverride child", c.ObjectName())
			}
		}
		// zlog.Info("PresentViewPopOverride", view.ObjectName(), parent.ObjectName())
	}
	nv.StopStoppers()
	nv.RemoveFromParent()
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

func PresentTitledView(view View, stitle string, winOptions WindowOptions, barViews map[View]zgeo.Alignment, ready func(stack, bar *StackView), presented func(*Window), closed func()) {
	stack, _ := view.(*StackView)
	if stack == nil {
		stack = StackViewVert("present-titled-stack")
		stack.SetSpacing(0)
		stack.Add(zgeo.TopCenter|zgeo.Expand, view)
	}
	bar := StackViewHor("bar")
	bar.SetSpacing(2)
	// bar.SetMarginS(zgeo.Size{6, 2})
	bar.SetDrawHandler(func(rect zgeo.Rect, canvas *Canvas, view View) {
		colors := []zgeo.Color{zgeo.ColorNew(0.85, 0.88, 0.91, 1), zgeo.ColorNew(0.69, 0.72, 0.76, 1)}
		path := zgeo.PathNewRect(rect, zgeo.Size{})
		canvas.DrawGradient(path, colors, rect.Min(), rect.BottomLeft(), nil)
	})
	stitle = zstr.TruncatedMiddle(stitle, 80, "…")
	titleLabel := LabelNew(stitle)
	titleLabel.SetFont(FontNew("Arial", 16, FontStyleBold))
	titleLabel.SetColor(zgeo.ColorNewGray(0.3, 1))
	bar.Add(zgeo.Left|zgeo.VertCenter, titleLabel) //, zgeo.Size{20, 0})

	xmargin := 0.0 //10.0
	for v, a := range barViews {
		if a&zgeo.Vertical == 0 {
			a |= zgeo.VertCenter
		}
		// zlog.Info("Bar add:", v.ObjectName())
		bar.Add(a, 0, v, zgeo.Size{xmargin, 0})
		xmargin = 0
	}
	stack.Add(zgeo.TopCenter|zgeo.HorExpand, 0, bar)
	if ready != nil {
		ready(stack, bar)
	}
	att := PresentViewAttributesNew()
	att.Title = stitle
	att.WindowOptions = winOptions
	PresentView(stack, att, presented, closed)
}
