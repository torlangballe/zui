/// Original Created by Tor Langballe on /22/9/14.

//go:build zui

package zpresent

import (
	"fmt"
	"runtime"

	"github.com/torlangballe/zui/zanimation"
	"github.com/torlangballe/zui/zcanvas"
	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zcustom"
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/zlabel"
	"github.com/torlangballe/zui/zstyle"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zui/zwindow"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zstr"
	"github.com/torlangballe/zutil/ztimer"
)

type Attributes struct {
	zwindow.Options
	DurationSecs             float64
	Transition               zanimation.Transition
	OldTransition            zanimation.Transition
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
	ModalDropShadow          zstyle.DropShadow
	ModalDismissOnEscapeKey  bool
	NoMessageOnOpenFail      bool
	FocusView                zview.View
}

var (
	presentCloseFuncs = map[zview.View]func(dismissed bool){}
	FirstPresented    bool
	Presenting        = true // true for first pre-present
	ShowErrorFunc     func(title, subTitle string)
)

// PresentView presents the view v either in a new window, or a modal window which might be just a view on top of the current window.
// If opening fails (on browsers it can fail for non-modal if popups are blocked), presented, and closed (if != nil) are called.
// closed (if != nil) is called when the window is closed programatically or by user interaction.
func PresentView(v zview.View, attributes Attributes, presented func(win *zwindow.Window), closed func(dismissed bool)) {
	if closed != nil {
		presentCloseFuncs[v] = closed
	}
	Presenting = true

	CallReady(v, true)

	win := zwindow.GetMain()
	w := zwindow.Current()
	if w != nil {
		win = w
	}

	outer := v
	if attributes.Modal {
		outer = makeEmbeddingViewAndAddToWindow(win, v, attributes, closed)
	}
	ct, _ := v.(zcontainer.ChildrenOwner)
	if ct != nil {
		zcontainer.WhenContainerLoaded(ct, func(waited bool) {
			presentLoaded(win, v, outer, attributes, presented, closed)
		})
	} else {
		presentLoaded(win, v, outer, attributes, presented, closed)
	}
}

func presentLoaded(win *zwindow.Window, v, outer zview.View, attributes Attributes, presented func(win *zwindow.Window), closed func(dismissed bool)) {
	fullRect := win.ContentRect()
	fullRect.Pos = zgeo.Pos{}
	rect := fullRect
	size := v.CalculatedSize(rect.Size)
	// zlog.Info("Present:", v.Native().Hierarchy(), size, rect, attributes.Modal, FirstPresented, zlog.CallingStackString())
	if attributes.Modal || FirstPresented {
		rect = rect.Align(size, zgeo.Center, zgeo.Size{})
	}
	nv := v.Native()
	if attributes.Modal {
		// zlog.Info("Present:", len(presentCloseFuncs))
		if nv != nil {
			r := rect
			if attributes.Pos != nil {
				if attributes.Alignment == zgeo.AlignmentNone {
					r.Pos = *attributes.Pos
				} else {
					r.Pos = zgeo.Rect{Pos: *attributes.Pos}.Align(size, attributes.Alignment|zgeo.Out, zgeo.Size{}).Pos
					// zlog.Info("ALIGN2:", r.Pos)
				}
			}
			frect := fullRect //.Expanded(zgeo.SizeBoth(-4))
			r = r.MovedInto(frect)
			r = r.Intersected(frect)
			v.SetRect(r)
			// zlog.Info("Present:", v.Native().Hierarchy(), r, win.ContentRect())
		}
		if attributes.ModalDismissOnEscapeKey {
			w := zwindow.FromNativeView(nv)
			w.AddKeypressHandler(v, func(km zkeyboard.KeyMod, down bool) bool {
				if km.Modifier == zkeyboard.ModifierNone && km.Key == zkeyboard.KeyEscape && down {
					Close(v, true, nil)
					return true
				}
				return false
			})
		}
	} else {
		if !FirstPresented {
			win.AddView(v)
			win.AddStyle()
		} else {
			size.H += zwindow.BarHeight()
			//			o := WindowOptions{URL: "about:blank", Pos: &rect.Pos, Size: size, ID: attributes.WindowID}
			o := attributes.Options
			o.Pos = &rect.Pos
			o.Size = size
			win = zwindow.Open(o)
			if win == nil {
				if !attributes.NoMessageOnOpenFail && ShowErrorFunc != nil {
					sub := o.URL
					if runtime.GOOS == "js" {
						sub = zstr.Concat("\n", sub, "This might be because popup windows are blocked in browser settings.")
					}
					ShowErrorFunc("Error opening window.", sub)
				}
				if presented != nil {
					presented(nil)
					if attributes.FocusView != nil {
						attributes.FocusView.Native().Focus(true)
					}
				}
				if closed != nil {
					closed(false)
				}
				return
			}
			win.AddStyle()
			win.AddView(v)
			if attributes.Title != "" {
				win.SetTitle(attributes.Title)
			}
			if closed != nil {
				win.HandleClosed = func() {
					CloseOverride(v, false, Attributes{}, func(dismissed bool) {})
					closed(true)
					delete(presentCloseFuncs, v)
				}
			}
		}
		win.ResizeHandlingView = v
		v.SetRect(zgeo.Rect{Size: rect.Size})
	}
	FirstPresented = true
	win.SetOnKeyEvents()
	win.ViewsStack = append(win.ViewsStack, v)

	Presenting = false
	CallReady(outer, false)
	if !attributes.Modal {
		win.SetOnResizeHandling()
	}
	if presented != nil {
		presented(win)
		if attributes.FocusView != nil {
			attributes.FocusView.Native().Focus(true)
		}
	}
}

func Close(view zview.View, dismissed bool, done func(dismissed bool)) {
	CloseOverride(view, dismissed, Attributes{}, done)
}

func CloseOverride(view zview.View, dismissed bool, overrideAttributes Attributes, done func(dismissed bool)) {
	// zlog.Info("CloseOverride", dismissed, zlog.CallingStackString())
	// TODO: Handle non-modal window too
	// zlog.Info("CloseOverride", dismissed, view.ObjectName(), reflect.ValueOf(view).Type())
	if done != nil {
		delete(presentCloseFuncs, view)
	}
	nv := view.Native()
	parent := nv.Parent()
	if parent != nil && parent.ObjectName() == "$titled" {
		// zlog.Info("CloseOverride remove blocker instead", view.ObjectName())
		nv = parent
	}
	if parent != nil && parent.ObjectName() == "$blocker" {
		// zlog.Info("CloseOverride remove blocker instead", view.ObjectName())
		nv = parent
	}
	win := zwindow.FromNativeView(nv)
	plen := len(win.ViewsStack)
	if plen > 0 {
		win.ViewsStack = win.ViewsStack[:plen-1]
	}
	// zlog.Info("CloseOverride:", plen, view != nil, win != nil)
	if plen > 1 {
		win.ProgrammaticView = win.ViewsStack[plen-2] // stack has been tructated by 1 since plen calculated
	} else {
		win.ProgrammaticView = nil
	}
	nv.RemoveFromParent()
	if done != nil {
		done(dismissed)
	}
	cf := presentCloseFuncs[view]
	if cf != nil {
		ztimer.StartIn(0.1, func() {
			// zlog.Info("Check PresentCloseFunc:", presentCloseFunc != nil)
			cf(dismissed)
		})
		// presentCloseFunc = nil // can't do this, clears before StartIn
	}
}

func CurrentIsParent(v zview.View) bool {
	nv := v.Native()
	win := zwindow.FromNativeView(nv)
	l := len(win.ViewsStack)
	if l <= 1 {
		return true
	}
	p := win.ViewsStack[l-1]
	// zlog.Info("PresentedViewCurrentIsParent", l, v.ObjectName(), p.ObjectName())
	if p == v {
		return true
	}
	for _, n := range nv.AllParents() {
		// fmt.Printf("CIP: %p %p\n", v, n.View)
		if n.View == p {
			return true
		}
	}
	return false
}

func setTransition(n *zview.NativeView, transition zanimation.Transition, screen zgeo.Rect, fade float32) {
	var me = screen
	var out = me
	switch transition {
	case zanimation.TransitionFromLeft:
		out.Pos.X += -me.Max().X

	case zanimation.TransitionFromRight:
		out.Pos.X += screen.Size.W - me.Pos.X

	case zanimation.TransitionFromTop:
		out.Pos.Y += -me.Max().Y

	case zanimation.TransitionFromBottom:
		out.Pos.Y += screen.Size.H - me.Pos.Y

	default:
		break
	}
	n.SetAlpha(fade)
	n.SetRect(out)
}

func AttributesNew() Attributes {
	a := Attributes{}
	a.DurationSecs = 0.5
	a.MakeFull = false
	a.PortraitOnly = false
	a.ModalDimBackground = true
	a.ModalDropShadow = zstyle.DropShadow{
		Delta: zgeo.Size{4, 4},
		Blur:  8,
		Color: zgeo.ColorNewGray(0.2, 1),
	}
	a.ModalCorner = 5
	return a
}

func CallReady(v zview.View, beforeWindow bool) {
	nv := v.Native()
	if nv == nil {
		return
	}
	if !nv.Presented {
		if !beforeWindow {
			nv.Presented = true
		}
		r, _ := v.(zview.ReadyToShowType)
		if r != nil {
			r.ReadyToShow(beforeWindow)
		}
	}
	// if nv.allChildrenPresented {
	// 	return
	// }
	ct, _ := v.(zcontainer.ChildrenOwner)
	if ct != nil {
		// zlog.Info("CallReady1:", nv.Hierarchy(), nv.Presented, len(ct.GetChildren(false)))
		for _, c := range ct.GetChildren(false) {
			CallReady(c, beforeWindow)
		}
	}
	// if !beforeWindow {
	// 	nv.allChildrenPresented = true
	// }
}

func PrintPresented(v zview.View, space string) {
	nv := v.Native()
	fmt.Printf(space+"Presented: %s %p: %v\n", nv.Hierarchy(), nv, nv.Presented)
	ct, _ := v.(zcontainer.ChildrenOwner)
	if ct != nil {
		for _, c := range ct.GetChildren(false) {
			PrintPresented(c, space+"  ")
		}
	}
}

func makeEmbeddingViewAndAddToWindow(win *zwindow.Window, v zview.View, attributes Attributes, closed func(dismissed bool)) (outer zview.View) {
	outer = v
	nv := v.Native()
	ct, _ := v.(zcontainer.ChildrenOwner)
	if ct != nil && attributes.ModalCorner != 0 {
		nv.SetCorner(attributes.ModalCorner)
	}
	zlog.Assert(nv != nil)
	if !attributes.ModalDropShadow.Delta.IsNull() {
		nv.SetDropShadow(attributes.ModalDropShadow)
	}
	if !attributes.ModalNoBlock {
		blocker := zcontainer.New("$blocker")
		outer = blocker
		fullRect := win.ContentRect()
		fullRect.Pos = zgeo.Pos{}
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
				Close(v, dismissed, closed)
			})
		}
	}
	win.AddView(outer)
	return
}

func GetTopPushed() *zcustom.CustomView {
	return nil
}

func RecusivelyHandleActivation(activated bool) {
	if activated {
	}
}

func PresentTitledView(view zview.View, stitle string, att Attributes, barViews map[zview.View]zgeo.Alignment, ready func(stack, bar *zcontainer.StackView, title *zlabel.Label), presented func(*zwindow.Window), closed func(dismissed bool)) {
	stack := zcontainer.StackViewVert("$titled")
	stack.SetSpacing(0)
	stack.Add(view, zgeo.TopCenter|zgeo.Expand)
	stack.SetBGColor(zstyle.DefaultBGColor())

	bar := zcontainer.StackViewHor("bar")
	bar.SetSpacing(2)
	bar.SetMarginS(zgeo.Size{6, 2})
	bar.SetDrawHandler(func(rect zgeo.Rect, canvas *zcanvas.Canvas, view zview.View) {
		colors := []zgeo.Color{zgeo.ColorNew(0.85, 0.88, 0.91, 1), zgeo.ColorNew(0.69, 0.72, 0.76, 1)}
		path := zgeo.PathNewRect(rect, zgeo.Size{})
		canvas.DrawGradient(path, colors, rect.Min(), rect.BottomLeft(), nil)
	})
	stitle = zstr.TruncatedMiddle(stitle, 160, "…")
	titleLabel := zlabel.New(stitle)
	titleLabel.SetFont(zgeo.FontNew("Arial", zgeo.FontDefaultSize+1, zgeo.FontStyleBold))
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
		bar.AddAdvanced(v, a, zgeo.Size{xmargin, 0}, zgeo.Size{}, 0, false)
		xmargin = 0
	}
	stack.AddAdvanced(bar, zgeo.TopCenter|zgeo.HorExpand, zgeo.Size{}, zgeo.Size{}, 0, false)
	if ready != nil {
		ready(stack, bar, titleLabel)
	}
	att.Title = stitle
	PresentView(stack, att, presented, closed)
}

func PopupView(view, over zview.View, align zgeo.Alignment, marg zgeo.Size) {
	view.Native().JSSet("className", "znofocus")
	att := AttributesNew()
	att.Modal = true
	att.ModalDimBackground = false
	att.ModalCloseOnOutsidePress = true
	att.ModalDropShadow.Delta = zgeo.SizeBoth(1)
	att.ModalDropShadow.Blur = 2
	att.ModalDismissOnEscapeKey = true
	var root zview.View
	over.Native().RootParent()
	pos := over.Native().AbsoluteRect().Align(zgeo.Size{1, 1}, align, marg).Pos
	att.Pos = &pos
	PresentView(view, att, func(win *zwindow.Window) {
		root = win.ViewsStack[len(win.ViewsStack)-2] // we can only do this for sure if modal is true
		view.Native().Focus(true)
		zcontainer.FocusNext(view, true, true)
		root.Native().SetInteractive(false)
	}, func(dismissed bool) {
		root.Native().SetInteractive(true)
	})

}
