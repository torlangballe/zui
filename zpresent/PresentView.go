/// Original Created by Tor Langballe on /22/9/14.

//go:build zui

package zpresent

import (
	"fmt"
	"runtime"

	"github.com/torlangballe/zui/zanimation"
	"github.com/torlangballe/zui/zcanvas"
	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/zlabel"
	"github.com/torlangballe/zui/zstyle"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zui/zwindow"
	"github.com/torlangballe/zutil/zdebug"
	"github.com/torlangballe/zutil/zfloat"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zscreen"
	"github.com/torlangballe/zutil/zstr"
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
	ModalCornerSides         []zgeo.Alignment
	ModalCloseOnOutsidePress bool
	ModalDimBackground       bool
	ModalNoBlock             bool
	ModalDropShadows         []zstyle.DropShadow
	ModalDismissOnEscapeKey  bool
	ModalStrokeWidth         float64
	ModalStrokeColor         zgeo.Color
	NoMessageOnOpenFail      bool
	PlaceOverMargin          zgeo.Size
	TitledMargin             zgeo.Rect
	PlaceOverView            zview.View
	FocusView                zview.View // TODO: Use new zcontainer.InitialFocusedView instead?
	PresentedFunc            func(win *zwindow.Window)
	ClosedFunc               func(dismissed bool)
}

var (
	presentCloseFuncs  = map[zview.View]func(dismissed bool){}
	previousFocusViews []zview.View
	FirstPresented     bool
	Presenting         = true // true for first pre-present
	ShowErrorFunc      func(title, subTitle string)
)

var ModalConfirmAttributes = Attributes{
	Modal:              true,
	ModalDimBackground: true,
	ModalDropShadows:   []zstyle.DropShadow{zstyle.DropShadowDefault},
	// ModalDismissOnEscapeKey: true, // The view needs to have a special escape key do dismiss on escape
}

var ModalPopupAttributes = Attributes{
	Modal:                    true,
	ModalCloseOnOutsidePress: true,
	ModalDismissOnEscapeKey:  true,
	ModalDropShadows:         []zstyle.DropShadow{zstyle.DropShadowDefault},
}

// PresentView presents the view v either in a new window, or a modal window which might be just a view on top of the current window.
// If opening fails (on browsers it can fail for non-modal if popups are blocked), presented, and closed (if != nil) are called.
// closed (if != nil) is called when the window is closed programatically or by user interaction.
func PresentView(v zview.View, attributes Attributes) {
	if attributes.ClosedFunc != nil {
		presentCloseFuncs[v] = attributes.ClosedFunc
	}
	Presenting = true
	CallReady(v, true)
	// zlog.Info("PresentView:", v.Native().Hierarchy(), attributes.Alignment)
	win := zwindow.GetMain()
	w := zwindow.Current()
	if w != nil {
		win = w
	}
	outer := v
	if attributes.Modal {
		outer = makeEmbeddingViewAndAddToWindow(win, v, attributes)
	}
	ct, _ := v.(zcontainer.ChildrenOwner)
	if false && ct != nil { //!!!! Let's try not doing this, as will block if popup windows not allowed in browser (cause it's not run from a user action)
		zcontainer.WhenContainerLoaded(ct, func(waited bool) {
			presentLoaded(win, v, outer, attributes)
		})
	} else {
		presentLoaded(win, v, outer, attributes)
	}
}

func presentLoaded(win *zwindow.Window, v, outer zview.View, attributes Attributes) {
	fullRect := win.ContentRect()
	fullRect.Pos = zgeo.Pos{}
	rect := fullRect
	var s zgeo.Size
	if attributes.Modal {
		s = win.ContentRect().ExpandedD(-10).Size
	} else {
		s = zscreen.GetMain().UsableRect.ExpandedD(-10).Size
	}
	size, _ := v.CalculatedSize(s)
	// size.MultiplyD(win.Scale)

	if attributes.Modal || FirstPresented {
		rect = rect.Align(size, attributes.Alignment, zgeo.SizeNull)
	}
	nv := v.Native()
	oldFocus := getCurrentFocus(win.ViewsStack)
	previousFocusViews = append(previousFocusViews, oldFocus)
	if attributes.Modal {
		if oldFocus != nil {
			oldFocus.Native().Focus(false)
		}
		if nv != nil {
			r := rect
			full := fullRect
			center := full.Center()
			if attributes.PlaceOverView != nil {
				zlog.Assert(attributes.Alignment != zgeo.AlignmentNone, v.Native().Hierarchy())
				r = attributes.PlaceOverView.Native().AbsoluteRect().Align(size, attributes.Alignment, attributes.PlaceOverMargin)
			} else if attributes.Pos != nil {
				if attributes.Alignment == zgeo.AlignmentNone {
					if attributes.Pos.X > center.X {
						attributes.Alignment = zgeo.Left
					} else {
						attributes.Alignment = zgeo.Right
					}
					if attributes.Pos.Y > center.Y {
						attributes.Alignment |= zgeo.Top
					} else {
						attributes.Alignment |= zgeo.Bottom
					}
				}
				r.Pos = zgeo.Rect{Pos: *attributes.Pos}.Align(size, attributes.Alignment|zgeo.Out, zgeo.SizeNull).Pos
				if len(attributes.ModalCornerSides) == 0 && attributes.ModalCorner != 0 {
					ah := attributes.Alignment.FlippedHorizontal()
					av := attributes.Alignment.FlippedVertical()
					attributes.ModalCornerSides = []zgeo.Alignment{ah, av, attributes.Alignment}
				}
			}
			full.Size.W -= win.ScrollBarSize() // scroll bare seems to be on top of everything, let's get out of the way
			r = r.MovedInto(full)
			zfloat.Maximize(&r.Pos.X, 0) // these are needed for overflow:scroll in blocker to work???
			zfloat.Maximize(&r.Pos.Y, 0) // +
			v.SetRect(r)
			if attributes.ModalCorner != 0 {
				if len(attributes.ModalCornerSides) != 0 {
					nv.SetCorners(attributes.ModalCorner, attributes.ModalCornerSides...)
				} else {
					nv.SetCorner(attributes.ModalCorner)
				}
			}
		}
		if attributes.ModalDismissOnEscapeKey {
			w := zwindow.FromNativeView(nv)
			w.AddKeyPressHandler(v, zkeyboard.KeyMod{Key: zkeyboard.KeyEscape}, true, func() bool {
				Close(v, true, nil)
				return true
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
			zlog.Info("WinOPen:", zdebug.CallingStackString())
			win = zwindow.Open(o)
			if win == nil {
				if !attributes.NoMessageOnOpenFail && ShowErrorFunc != nil {
					sub := o.URL
					if runtime.GOOS == "js" {
						sub = zstr.Concat("\n", sub, "This might be because popup windows are blocked in browser settings.")
					}
					ShowErrorFunc("Error opening window.", sub)
				}
				if attributes.PresentedFunc != nil {
					attributes.PresentedFunc(nil)
				}
				v.Native().Focus(true)
				if attributes.FocusView != nil {
					attributes.FocusView.Native().Focus(true)
				} else {
					zcontainer.FocusNext(v, true, true)
				}
				if attributes.ClosedFunc != nil {
					attributes.ClosedFunc(false)
				}
				return
			}
			win.AddStyle()
			// v.Show(false)
			win.AddView(v)
			if attributes.Title != "" {
				win.SetTitle(attributes.Title)
			}
			if attributes.ClosedFunc != nil {
				win.HandleClosed = func() {
					CloseOverride(v, false, Attributes{}, func(dismissed bool) {})
					// attributes.ClosedFunc(true) // this is done in CloseOverride, it's set in presentCloseFuncs[]
					delete(presentCloseFuncs, v)
				}
			}
		}
		win.ResizeHandlingView = v
		s := win.ContentRect().Size
		if s.IsNull() {
			s = size
		}
		// r := zgeo.Rect{Size: }
		r := zgeo.Rect{Size: s}
		// r.Size.W--
		// r.Size.H--
		v.SetRect(r)
	}
	FirstPresented = true
	win.ViewsStack = append(win.ViewsStack, v)

	Presenting = false
	CallReady(outer, false)
	if !attributes.Modal {
		win.SetOnResizeHandling()
	}
	if attributes.PresentedFunc != nil {
		attributes.PresentedFunc(win)
	}
	if attributes.FocusView != nil {
		attributes.FocusView.Native().Focus(true)
	} else {
		// zlog.Info("Presented, focus next")
		if v.Native().GetFocusedChildView(false) == nil {
			zcontainer.FocusNext(v, true, true)
		}
	}
}

func getCurrentFocus(stack []zview.View) zview.View {
	slen := len(stack)
	if slen == 0 {
		return nil
	}
	top := stack[slen-1]
	return top.Native().GetFocusedChildView(true)
}

func Close(view zview.View, dismissed bool, done func(dismissed bool)) {
	CloseOverride(view, dismissed, Attributes{}, done)
}

func CloseOverride(view zview.View, dismissed bool, overrideAttributes Attributes, done func(dismissed bool)) {
	// TODO: Handle non-modal window too
	// zlog.Info("CloseOverride:", view.Native().Hierarchy(), zdebug.CallingStackString())
	old := presentCloseFuncs[view]
	presentCloseFuncs[view] = func(dismissed bool) {
		if done != nil {
			done(dismissed)
		}
		if old != nil {
			old(dismissed)
		}
		plen := len(previousFocusViews)
		if plen > 0 {
			oldFoc := previousFocusViews[plen-1]
			previousFocusViews = previousFocusViews[:plen-1]
			if oldFoc != nil {
				oldFoc.Native().Focus(true)
			}
		}
	}
	nv := view.Native()
	parent := nv.Parent()
	// zlog.Info("CloseOverride:", nv.Hierarchy())
	if parent != nil && parent.ObjectName() == "$titled" {
		nv = parent
		// zlog.Info("CloseOverride:", nv.Hierarchy(), parent.ObjectName())
		parent = parent.Parent()
	}
	if parent != nil && parent.ObjectName() == "$blocker" {
		nv = parent
		// zlog.Info("CloseOverride:", nv.Hierarchy())
	}
	win := zwindow.FromNativeView(nv)
	plen := len(win.ViewsStack)
	if plen > 0 {
		win.ViewsStack = win.ViewsStack[:plen-1]
	}
	if plen > 1 {
		win.ProgrammaticView = win.ViewsStack[plen-2] // stack has been tructated by 1 since plen calculated
	} else {
		win.ProgrammaticView = nil
	}
	cf := presentCloseFuncs[view]
	if cf != nil {
		// ztimer.StartIn(0.1, func() {
		cf(dismissed)
		// })
	}
	nv.RemoveFromParent(true)
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
	a.ModalDropShadows = []zstyle.DropShadow{
		zstyle.DropShadow{
			Delta: zgeo.SizeD(4, 4),
			Blur:  8,
			Color: zgeo.ColorNewGray(0.2, 1),
		},
	}
	a.ModalCorner = 5
	return a
}

func CallReady(v zview.View, beforeWindow bool) {
	nv := v.Native()
	if nv == nil {
		return
	}
	if !nv.IsPresented() {
		if !beforeWindow {
			nv.Flags |= zview.ViewPresentedFlag
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
		// zlog.Info("CallReady1:", nv.Hierarchy(), nv.IsPresented(), len(ct.GetChildren(false)))
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
	fmt.Printf(space+"Presented: %s %p: %v\n", nv.Hierarchy(), nv, nv.IsPresented())
	ct, _ := v.(zcontainer.ChildrenOwner)
	if ct != nil {
		for _, c := range ct.GetChildren(false) {
			PrintPresented(c, space+"  ")
		}
	}
}

func makeEmbeddingViewAndAddToWindow(win *zwindow.Window, v zview.View, attributes Attributes) (outer zview.View) {
	outer = v
	nv := v.Native()
	zlog.Assert(nv != nil)
	if attributes.ModalStrokeWidth != 0 {
		nv.SetStroke(attributes.ModalStrokeWidth, attributes.ModalStrokeColor, false)
	}
	nv.SetDropShadow(attributes.ModalDropShadows...)
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
		if attributes.Alignment == zgeo.AlignmentNone {
			attributes.Alignment = zgeo.Center
		}
		blocker.Add(v, attributes.Alignment) // |zgeo.Shrink
		if attributes.ModalCloseOnOutsidePress {
			blocker.SetPressedHandler("$blocker.click.away", zkeyboard.ModifierNone, func() {
				vr := v.Native().AbsoluteRect()
				pos := zview.LastPressedPos.Plus(blocker.AbsoluteRect().Pos)
				// zlog.Info("Blocker Eater Clicked", zview.LastPressedPos, pos, "r:", vr, vr.Contains(pos))
				if vr.Contains(pos) {
					return
				}
				dismissed := true
				Close(v, dismissed, attributes.ClosedFunc)
			})
		}
		// v.Native().SetPressedHandler("$modal.click.eater", zkeyboard.ModifierNone, func() {
		// 	zview.StopPropagationOfLastPressedEvent()
		// 	zlog.Info("Blocker Eater Clicked")
		// }) // so it doesn't propagate to blocker
		blocker.JSSet("className", "znoscrollbar")
		blocker.SetJSStyle("overflow", "scroll")
	}
	win.AddView(outer)
	return
}

func PresentTitledView(view zview.View, stitle string, att Attributes, barViews map[zview.View]zgeo.Alignment, ready func(stack, bar *zcontainer.StackView, title *zlabel.Label)) {
	stack := zcontainer.StackViewVert("$titled")
	stack.SetSpacing(0)
	stack.SetBGColor(zstyle.DefaultBGColor())
	//!! if att.TitledMargin.H != 0 {
	// 	stack.SetMargin(zgeo.RectFromXY2(0, 0, 0, -att.TitledMargin.H))
	// }

	a := zgeo.Left
	if len(barViews) == 0 {
		a = zgeo.HorCenter
	}
	bar, titleLabel := MakeBar(stitle, a)
	m := bar.Margin()
	if m.Pos.X == 0 {
		m.SetMinX(att.TitledMargin.Pos.X)
		bar.SetMargin(m)
	}
	stack.Add(bar, zgeo.TopCenter|zgeo.HorExpand)
	//	m := zgeo.SizeD(att.TitledMargin.W, 0)
	stack.Add(view, zgeo.TopCenter|zgeo.Expand, zgeo.SizeNull)

	// xmargin := zstyle.DefaultRowRightMargin
	for v, a := range barViews {
		if a&zgeo.Vertical == 0 {
			a |= zgeo.Vertical
		}
		// mr := zgeo.RectMarginForSizeAndAlign(zgeo.SizeD(xmargin, 0), a)
		bar.AddAdvanced(v, a, zgeo.RectNull, zgeo.SizeNull, 0, false)
		// xmargin = 0
	}
	if ready != nil {
		ready(stack, bar, titleLabel)
	}
	att.Title = stitle
	PresentView(stack, att)
}

func MakeBar(stitle string, titleAlign zgeo.Alignment) (*zcontainer.StackView, *zlabel.Label) {
	bar := zcontainer.StackViewHor("bar")
	bar.SetSpacing(2)
	bar.SetMargin(zgeo.RectFromXY2(6, 4, -6, -7))
	bar.SetDrawHandler(func(rect zgeo.Rect, canvas *zcanvas.Canvas, view zview.View) {
		colors := []zgeo.Color{zgeo.ColorNew(0.85, 0.88, 0.91, 1), zgeo.ColorNew(0.69, 0.72, 0.76, 1)}
		path := zgeo.PathNewRect(rect, zgeo.SizeNull)
		canvas.DrawGradient(path, colors, rect.Min(), rect.BottomLeft(), nil)
	})
	stitle = zstr.TruncatedMiddle(stitle, 160, "…")
	titleLabel := zlabel.New(stitle)
	titleLabel.SetTextAlignment(titleAlign)
	titleLabel.SetFont(zgeo.FontNew("Arial", zgeo.FontDefaultSize+1, zgeo.FontStyleBold))
	titleLabel.SetColor(zgeo.ColorNewGray(0.2, 1))
	bar.Add(titleLabel, titleAlign|zgeo.VertCenter|zgeo.HorExpand)

	return bar, titleLabel
}

func PopupView(view, over zview.View, att Attributes) {
	var root zview.View
	view.Native().JSSet("className", "znofocus")
	if att.Alignment == zgeo.AlignmentNone {
		att.Alignment = zgeo.TopLeft
	}
	att.Modal = true
	att.ModalDimBackground = false
	att.ModalCloseOnOutsidePress = true

	att.ModalDismissOnEscapeKey = true
	att.PlaceOverView = over
	att.PresentedFunc = func(win *zwindow.Window) {
		root = win.ViewsStack[len(win.ViewsStack)-2] // we can only do this for sure if modal is true
		root.Native().SetInteractive(false)
	}
	att.ClosedFunc = func(dismissed bool) {
		root.Native().SetInteractive(true)
	}
	PresentView(view, att)
}
