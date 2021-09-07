// +build zui

package zui

import (
	"fmt"

	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zstr"
	"github.com/torlangballe/zutil/ztimer"
)

//  Created by Tor Langballe on /22/9/14.

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
	ModalCorner              float64
	ModalCloseOnOutsidePress bool
	ModalDimBackground       bool
	ModalNoBlock             bool
	ModalDropShadow          zgeo.DropShadow
	ModalDismissOnEscapeKey  bool
}

var (
	stack                 []PresentViewAttributes
	presentCloseFunc      func(dismissed bool)
	presentedViewStack    []View
	firstPresented        bool
	presentViewPresenting = true
)

func PresentView(v View, attributes PresentViewAttributes, presented func(win *Window), closed func(dismissed bool)) {
	presentedViewStack = append(presentedViewStack, v)
	presentCloseFunc = closed
	presentViewPresenting = true

	PresentViewCallReady(v, true)

	outer := makeEmbeddingViewAndAddToWindow(v, attributes, closed)
	ct, _ := v.(ContainerType)
	//zlog.Info("Present1:", ct != nil, reflect.ValueOf(outer).Type())
	// zlog.Info("Present1:", zlog.GetCallingStackString())
	if ct != nil {
		WhenContainerLoaded(ct, func(waited bool) {
			// zlog.Info("Present2", firstPresented)
			presentLoaded(v, outer, attributes, presented, closed)
		})
	} else {
		presentLoaded(v, outer, attributes, presented, closed)
	}
}

func presentLoaded(v, outer View, attributes PresentViewAttributes, presented func(win *Window), closed func(dismissed bool)) {
	win := WindowGetMain()
	fullRect := win.ContentRect()
	fullRect.Pos = zgeo.Pos{}
	rect := fullRect
	size := v.CalculatedSize(rect.Size)
	if attributes.Modal || firstPresented {
		rect = rect.Align(size, zgeo.Center, zgeo.Size{})
	}
	// zlog.Info("presentLoaded", firstPresented, size, rect, fullRect)
	nv := ViewGetNative(v)
	if attributes.Modal {
		if nv != nil {
			r := rect
			if attributes.Pos != nil {
				if attributes.Alignment == zgeo.AlignmentNone {
					r.Pos = *attributes.Pos
				} else {
					// zlog.Info("ALIGN1:", *attributes.Pos, size, attributes.Alignment)
					r.Pos = zgeo.Rect{Pos: *attributes.Pos}.Align(size, attributes.Alignment|zgeo.Out, zgeo.Size{}).Pos
					// zlog.Info("ALIGN2:", r.Pos)
				}
			}
			frect := fullRect.Expanded(zgeo.SizeBoth(-4))
			r = r.MovedInto(frect)
			r = r.Intersected(frect)
			v.SetRect(r)
		}
		if attributes.ModalDismissOnEscapeKey {
			win := nv.GetWindow()
			win.AddKeypressHandler(v, func(key KeyboardKey, mod KeyboardModifier) {
				if mod == KeyboardModifierNone && key == KeyboardKeyEscape {
					PresentViewClose(v, true, nil)
				}
			})
		}
	} else {
		if !firstPresented {
			win.AddView(outer)
		} else {
			size.H += WindowBarHeight
			//			o := WindowOptions{URL: "about:blank", Pos: &rect.Pos, Size: size, ID: attributes.WindowID}
			o := attributes.WindowOptions
			o.Pos = &rect.Pos
			o.Size = size
			win = WindowOpen(o)
			win.AddView(outer)
			if attributes.Title != "" {
				win.SetTitle(attributes.Title)
			}
			if closed != nil {
				win.HandleClosed = func() {
					closed(true)
					presentCloseFunc = nil
				}
			}
		}
		v.SetRect(zgeo.Rect{Size: rect.Size})
	}
	firstPresented = true

	// cvt, _ := v.(ContainerViewType)
	// if cvt != nil {
	// 	cvt.ArrangeChildren()
	// }
	// NativeViewAddToRoot(v)
	presentViewPresenting = false
	et, _ := outer.(ExposableType)
	if et != nil {
		PresentViewCallReady(outer, false)
		et.drawIfExposed()
	}
	win.setOnResizeHandling()
	if presented != nil {
		presented(win)
	}
}

func PresentViewClose(view View, dismissed bool, done func(dismissed bool)) {
	PresentViewCloseOverride(view, dismissed, PresentViewAttributes{}, done)
}

func PresentViewCloseOverride(view View, dismissed bool, overrideAttributes PresentViewAttributes, done func(dismissed bool)) {
	// TODO: Handle non-modal window too
	// zlog.Info("PresentViewCloseOverride", dismissed, view.ObjectName(), reflect.ValueOf(view).Type())

	if done != nil {
		presentCloseFunc = nil
	}
	nv := ViewGetNative(view)
	parent := nv.Parent()
	if parent != nil && parent.ObjectName() == "$titled" {
		// zlog.Info("PresentViewCloseOverride remove blocker instead", view.ObjectName())
		nv = parent
	}
	if parent != nil && parent.ObjectName() == "$blocker" {
		// zlog.Info("PresentViewCloseOverride remove blocker instead", view.ObjectName())
		nv = parent
	}
	plen := len(presentedViewStack)
	win := nv.GetWindow()
	presentedViewStack = presentedViewStack[:plen-1]
	// zlog.Info("PresentViewCloseOverride:", plen, view.ObjectName())
	if plen > 1 {
		win.ProgrammaticView = presentedViewStack[plen-2] // stack has been tructated by 1 since plen calculated
	} else {
		win.ProgrammaticView = nil
	}
	// zlog.Info("PresentViewCloseOverride:", nv.Hierarchy())
	nv.RemoveFromParent()
	if done != nil {
		done(dismissed)
	}
	if presentCloseFunc != nil {
		ztimer.StartIn(0.1, func() {
			// zlog.Info("Check PresentCloseFunc:", presentCloseFunc != nil)
			if presentCloseFunc != nil { // we do a re-check in case it was nilled in 0.1 second
				presentCloseFunc(dismissed)
			}
		})
		// presentCloseFunc = nil // can't do this, clears before StartIn
	}
}

func PresentedViewCurrentIsParent(v View) bool {
	l := len(presentedViewStack)
	if l <= 1 {
		return true
	}
	nv := ViewGetNative(v)
	p := presentedViewStack[l-1]
	// zlog.Info("PresentedViewCurrentIsParent", l, v.ObjectName(), p.ObjectName())
	if p == v {
		return true
	}
	for _, n := range nv.AllParents() {
		if n.View == p {
			return true
		}
	}
	return false
}

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

func PresentViewAttributesNew() PresentViewAttributes {
	a := PresentViewAttributes{}
	a.DurationSecs = 0.5
	a.MakeFull = false
	a.PortraitOnly = false
	a.ModalDimBackground = true
	a.ModalDropShadow = zgeo.DropShadow{
		Delta: zgeo.Size{4, 4},
		Blur:  8,
		Color: zgeo.ColorNewGray(0.2, 1),
	}
	a.ModalCorner = 5
	return a
}

func PresentViewCallReady(v View, beforeWindow bool) {
	nv := ViewGetNative(v)
	// zlog.Info(beforeWindow, "PresentViewCallReady:", nv.Hierarchy(), nv.Presented)
	if nv == nil {
		return
	}
	if !nv.Presented {
		if !beforeWindow {
			// fmt.Printf("Set Presented: %s %p\n", nv.Hierarchy(), nv)
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
		// zlog.Info("PresentViewCallReady1:", v.ObjectName(), len(ct.GetVisibleChildren()))
		for _, c := range ct.GetChildren(false) {
			PresentViewCallReady(c, beforeWindow)
		}
	}
	if !beforeWindow {
		nv.allChildrenPresented = true
	}
}

func PrintPresented(v View, space string) {
	nv := ViewGetNative(v)
	fmt.Printf(space+"Presented: %s %p: %v\n", nv.Hierarchy(), nv, nv.Presented)
	ct, _ := v.(ContainerType)
	if ct != nil {
		for _, c := range ct.GetChildren(false) {
			PrintPresented(c, space+"  ")
		}
	}
}

func makeEmbeddingViewAndAddToWindow(v View, attributes PresentViewAttributes, closed func(dismissed bool)) (outer View) {
	outer = v
	win := WindowGetMain()
	if attributes.Modal {
		ct, _ := v.(ContainerType)
		if ct != nil && attributes.ModalCorner != 0 {
			v.SetCorner(attributes.ModalCorner)
		}
		nv := ViewGetNative(v)
		if nv != nil {
			if !attributes.ModalDropShadow.Delta.IsNull() {
				nv.SetDropShadow(attributes.ModalDropShadow)
			}
			// zlog.Info("makeEmbeddingViewAndAddToWindow:", nv.Hierarchy(), attributes.ModalNoBlock)
			if !attributes.ModalNoBlock {
				blocker := ContainerViewNew("$blocker")
				outer = blocker
				fullRect := win.ContentRect()
				fullRect.Pos = zgeo.Pos{}
				// zlog.Info("blocker rect:", fullRect)
				blocker.SetRect(fullRect)
				if attributes.ModalDimBackground {
					blocker.SetBGColor(zgeo.ColorNewGray(0, 0.5))
				} else {
					blocker.SetBGColor(zgeo.ColorClear)
				}
				blocker.Add(v, zgeo.TopLeft)
				if attributes.ModalCloseOnOutsidePress {
					// lp, _ := v.(Pressable)
					// if lp != nil {
					// 	lp.SetPressedHandler(func() {
					// 		zlog.Info("LP Pressed")
					// 	})
					// }
					blocker.SetPressedHandler(func() {
						dismissed := true
						PresentViewClose(v, dismissed, closed)
					})
				}
				win.AddView(blocker)
			} else {
				win.AddView(v)
			}
		}
	}
	ct, _ := v.(ContainerType)
	if ct != nil {
		recursive := true
		ContainerTypeRangeChildren(ct, recursive, false, func(view View) bool {
			// TODO: focus something here...
			return false
		})
	}
	return
}

func PresentViewGetTopPushed() *CustomView {
	return nil
}

func PresentViewRecusivelyHandleActivation(activated bool) {
	if activated {
	}
}

func PresentTitledView(view View, stitle string, att PresentViewAttributes, barViews map[View]zgeo.Alignment, ready func(stack, bar *StackView, title *Label), presented func(*Window), closed func(dismissed bool)) {
	stack := StackViewVert("$titled")
	stack.SetSpacing(0)
	stack.Add(view, zgeo.TopCenter|zgeo.Expand)

	bar := StackViewHor("bar")
	bar.SetSpacing(2)
	bar.SetMarginS(zgeo.Size{6, 2})
	bar.SetDrawHandler(func(rect zgeo.Rect, canvas *Canvas, view View) {
		colors := []zgeo.Color{zgeo.ColorNew(0.85, 0.88, 0.91, 1), zgeo.ColorNew(0.69, 0.72, 0.76, 1)}
		path := zgeo.PathNewRect(rect, zgeo.Size{})
		canvas.DrawGradient(path, colors, rect.Min(), rect.BottomLeft(), nil)
	})
	stitle = zstr.TruncatedMiddle(stitle, 160, "…")
	titleLabel := LabelNew(stitle)
	titleLabel.SetFont(FontNew("Arial", FontDefaultSize+1, FontStyleBold))
	titleLabel.SetColor(zgeo.ColorNewGray(0.2, 1))
	a := zgeo.Left
	if len(barViews) == 0 {
		a = zgeo.HorCenter
	}
	bar.Add(titleLabel, a|zgeo.VertCenter) //, zgeo.Size{20, 0})

	xmargin := 0.0 //10.0
	for v, a := range barViews {
		if a&zgeo.Vertical == 0 {
			a |= zgeo.Vertical
		}
		// zlog.Info("Bar add:", v.ObjectName())
		bar.Add(v, 0, a, zgeo.Size{xmargin, 0})
		xmargin = 0
	}
	stack.Add(bar, 0, zgeo.TopCenter|zgeo.HorExpand)
	if ready != nil {
		ready(stack, bar, titleLabel)
	}
	att.Title = stitle
	PresentView(stack, att, presented, closed)
}
