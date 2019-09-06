package zgo

//  Created by Tor Langballe on /13/11/15.

type ScrollView struct {
	CustomView
	Margin Rect
	Child  *ContainerView
}

func ScrollViewNew() *ScrollView {
	v := &ScrollView{}
	v.CustomView.init(v, "scrollview")
	return v
}

func (v *ScrollView) SetContentOffset(offset Pos, animated bool) {
}

func (v *ScrollView) Rect(rect Rect) View {

	if v.Child != nil {
		s := Size{v.GetLocalRect().Size.W, 3000}
		size := v.Child.GetCalculatedSize(s)
		size.W = s.W
		r := Rect{Size: size}
		r.Add(v.Margin)
		v.Child.Rect(r)
		//self.contentSize = size.GetCGSize()
		v.Child.ArrangeChildren(nil)
	}
	return v
}

func (v *ScrollView) SetChild(view *ContainerView) {
	// if child != nil {
	//     child?.RemoveFromParent()
	// }
	v.Child = view
	v.AddChild(v.Child, -1)
}

func ScrollViewToMakeItVisible(view View) {
	// var s:UIView? = view.View()
	// while s != nil {
	//     s = s!.superview
	//     if s != nil {
	//         if let sv = s! as? ZScrollView {
	//             if Double(sv.frame.size.height) - sv.margin.size.h < Double(sv.contentSize.height) {
	//                 let y = float64(view.View().convert(view.View().bounds, to:sv.View()).origin.y)
	//                 sv.SetContentOffset(ZPos(0, y - 40))
	//             }
	//             break
	//         }
	//     }
	// }
}
