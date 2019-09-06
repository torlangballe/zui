package zgo

//  Created by Tor Langballe on /4/12/15.

type ListViewDelegate interface {
	ListViewGetRowCount() int
	ListViewGetHeightOfItem(index int) float64
	ListViewSetupCell(cellSize Size, index int) *CustomView
	HandleRowSelected(index int)
	//    GetAccessibilityForCell( index int, prefixString)  [ZAccessibilty]
}

//typealias ZListViewRowAnimation = UIListView.RowAnimation

type ListView struct {
	NativeView
	First                   bool
	TableRowBackgroundColor Color
	Scrolling               bool
	DrawHandler             *func(rect Rect, canvas Canvas)
	Margins                 Size
	Spacing                 float64
	FocusedRow              *int

	selectionIndex int
	owner          ListViewDelegate
	selectable     bool
	selectedColor  Color
}

func ListViewNew() *ListView {
	v := &ListView{}
	v.selectable = true
	v.selectionIndex = -1
	return v
	//        allowsSelection = true // selectable
}

//     override func layoutSubviews() {
// //        let tvOsInsetCGFloat = ZIsTVBox() ? 174  0
// //        let tvOsInsetCGFloat = ZIsTVBox() ? 87  0
//         if first {
//             allowsSelection = true // selectable
//             if selectionIndex != -1 {
//                 Select(selectionIndex);
//             }
//   //        contentInset = UIEdgeInsets(top CGFloat(margins.h), left -tvOsInset, bottom CGFloat(margins.h), right tvOsInset)
//             first = false
//         }
//         super.layoutSubviews()
//     }

//     override func draw( rect CGRect) {
//         drawHandler?(Rect(rect), Canvas(context UIGraphicsGetCurrentContext()!))
//     }

func (v *ListView) ExposeRows() {
	// for i in indexPathsForVisibleRows ?? [] {
	//     if let c = self.cellForRow(at i) {
	//         exposeAll(c.contentView)
	//     }
	// }
}

func (v *ListView) UpdateVisibleRows(animate bool) {
	//  reloadRows(at indexPathsForVisibleRows ?? [], withanimate ? UIListView.RowAnimation.automatic  UIListView.RowAnimation.none)
}

func (v *ListView) ScrollToMakeRowVisible(row int, animate bool) {
}

func (v *ListView) UpdateRow(row int) {
}

func (v *ListView) ReloadData(animate bool) {
}

func (v *ListView) MoveRow(fromIndex int, toIndex int) {
}

// private func getZViewChild( vUIView)  ZView? {
//     for c in v.subviews {
//         if let z = c as? ZView {
//             return z
//         }
//     }
//     for c in v.subviews {
//         if let z = getZViewChild(c) {
//             return z
//         }
//     }
//     return nil
// }

func (v *ListView) GetRowViewFromIndex(index int) *View {
	return nil
}

func (v *ListView) GetIndexFromRowView(view View) *int {
	return nil
}

func LiistViewGetParentListViewFromRow(child *ContainerView) *ListView {
	return nil
}

func ListViewGetIndexFromRowView(view *ContainerView) int {
	return -1
}

func (v *ListView) Select(row int) {
}

func (v *ListView) DeleteChildRow(index int, transition PresentViewTransition) { // call this after removing data
}

// func scrollViewWillBeginDragging( scrollView UIScrollView) {
//     scrolling = true
// }

// func scrollViewDidEndDragging( scrollView UIScrollView, willDecelerate decelerate bool) {
//     if !decelerate {
//         scrolling = false
//     }
// }

// func scrollViewDidEndDecelerating( scrollView UIScrollView) {
//     scrolling = false
// }

func (v *ListView) IsFocused(row *CustomView) bool {
	return false
}

// func ListView( ListView UIListView, didSelectRowAt indexPath IndexPath) {
//     let index = pathToRow(indexPath)
//     owner!.HandleRowSelected(index)
//     selectionIndex = index
// }

// func ListView( ListView UIListView, heightForRowAt indexPath IndexPath)  CGFloat {
//     let index = pathToRow(indexPath)
//     return CGFloat(owner!.ListViewGetHeightOfItem(index))
// }

// func numberOfSections(in ListView UIListView)  int {
//     return 1
// }

// func ListView( ListViewUIListView, numberOfRowsInSection sectionint)  int {
//     let c = owner!.ListViewGetRowCount()
//     return c
// }

// func ListView( ListView UIListView, cellForRowAt indexPath IndexPath)  UIListViewCell {
//     //        let cell  UIListViewCell = self.dequeueReusableCellWithIdentifier("ZListView", forIndexPathindexPath) as UIListViewCell
//     let cell = zUIListViewCell()
//     cell.isEditing = true
//     let index = pathToRow(indexPath)
//     var r = Rect(sizeSize(Rect.size.w, owner!.ListViewGetHeightOfItem(index)))
//     var m = margins.w
//     if ZIsTVBox() {
//         m = 87
//     }
//     r = r.Expanded(Size(-m, 0))
//     if ZIsTVBox() {
//         cell.focusStyle = UIListViewCell.FocusStyle.custom
//     }
//     cell.frame = r.GetCGRect()
//     cell.backgroundColor = UIColor.clear
//     let s = Size(cell.frame.size)
//     let customView = owner!.ListViewSetupCell(s, indexindex)
//     customView?.frame = Rect(sizes).GetCGRect()
//     if !ZIsTVBox() {
//         customView!.minSize.h -= spacing
//     }
//     customView?.frame.size.height = CGFloat(customView!.minSize.h)
//     if let cv = customView as? ZContainerView {
//         cv.ArrangeChildren()
//     }
//     cell.isUserinteractionEnabled = true //cell.Usable
//     if selectable {
//         if !selectedColor.undefined {
//             let bgColorView = UIView()
//             bgColorView.backgroundColor = selectedColor.rawColor
//             cell.selectedBackgroundView = bgColorView
//         } else {
//             cell.selectedBackgroundView = UIView()
//         }
//     }
//     if customView != nil {
//         cell.contentView.addSubview(customView!)
//         cell.isOpaque = customView!.isOpaque
//         cell.backgroundColor = customView!.backgroundColor
//     }
//     if cell.backgroundColor != nil && cell.backgroundView != nil && ZColor(colorcell.backgroundColor!).Opacity == 0.0 {
//         cell.backgroundView!.backgroundColor = UIColor.clear
//         cell.contentView.backgroundColor = UIColor.clear
//         cell.backgroundColor = UIColor.clear
//     }
//     return cell
// }

// func ListView( ListView UIListView, willDisplay willDisplayCellUIListViewCell, forRowAt forRowAtIndexPathIndexPath) {
//     //        ZDebug.Print("willDisplayCell", forRowAtIndexPath.row)
// }

// func ListView( ListView UIListView, shouldHighlightRowAt indexPath IndexPath)  bool {
//     return true
// }

// func ListView( ListView UIListView, willSelectRowAt indexPath IndexPath)  IndexPath? {
//     return indexPath
// }

// func ListView( ListView UIListView, canEditRowAt indexPath IndexPath)  bool {
//     return false
// }

// func ListView( ListView UIListView, canMoveRowAt indexPath IndexPath)  bool {
//     return false
// }

// fileprivate func makeIndexPathFromIndex( indexint)  IndexPath {
//     let indexes[int] = [ 0, index]
//     return (NSIndexPath(indexesindexes, length2) as IndexPath)
// }
