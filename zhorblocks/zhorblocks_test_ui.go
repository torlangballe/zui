//go:build zui && js

package zhorblocks

import (
	"strconv"

	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zlabel"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zint"
)

func MakeHorBlocksTestView(id string, delete bool) zview.View {
	if delete {
		return nil
	}
	scroller := NewHorBlocksView(3, 14)
	scroller.GetViewFunc = func(index int) zview.View {
		c := zcontainer.New(strconv.Itoa(index))
		label := zlabel.New(strconv.Itoa(index))
		label.SetTextAlignment(zgeo.Center)
		label.SetColor(zgeo.ColorWhite)
		label.SetFont(zgeo.FontNice(100, zgeo.FontStyleBold))
		col := zgeo.ColorGreen
		if zint.Abs(index)%2 == 1 {
			col = zgeo.ColorRed
		}
		c.SetBGColor(col)
		c.Add(label, zgeo.Center)
		c.SetLongPressedHandler("jump", 0, func() {
			d := 20.0
			x := zview.LastPressedPos.X
			if x < scroller.viewSize.W/2 {
				d = -33
			}
			index := scroller.CurrentIndex()
			scroller.SetCurrentIndex(index - d)
		})
		return c
	}
	return scroller
}

/*
var lanes = []Lane{
	Lane{ID: "dog", Name: "Dog",
		Rows: []Row{
			Row{Name: "German Shepard", Height: 22},
			Row{Name: "Poodle", Height: 40},
			Row{Name: "Snauzer", Height: 50},
			Row{Name: "Dashund", Height: 30},
		},
	},
}
var events = makeEvents()

func MakeHorEventsTestView(id string, delete bool) zview.View {
	if delete {
		return nil
	}
	v := NewEventsView(nil, "test-horevents-store", time.Now(), 3, 10, time.Second*20)
	v.TimeAxisHeight = 22
	v.GetEventViewsFunc = func(blockIndex int, start, end time.Time, got func(childView zview.View, x float64, laneID, rowID string)) {
		// si := -1
		for _, e := range events {
			if !e.Start.Before(end) {
				break
			}
			if e.End.After(start) {
				view := makeEventView(v, e)
				x := v.TimeToXInCorrectBlock(e.Start)
				got(view, x, e.LaneID, e.LaneRowID)
			}
		}
	}
	v.SetLanes(lanes)
	return v
}

func makeEventView(v *HorEventsView, e Event) zview.View {
	// zlog.Info("makeEventView:", id, size)
	str := fmt.Sprint(e.ID, "\n", ztime.GetNiceSubSecs(e.Start, 3), "-", ztime.GetNiceSubSecs(e.End, 3))
	view := zcontainer.New(str)
	view.SetBGColor(zgeo.ColorRandom())
	w := v.DurationToWidth(e.End.Sub(e.Start))
	_, row := v.FindLaneAndRow(e.LaneID, e.LaneRowID)
	view.SetMinSize(zgeo.SizeD(w, row.Height))

	label := zlabel.New(str)
	label.SetTextAlignment(zgeo.Center)
	label.SetColor(zgeo.ColorWhite)
	label.SetFont(zgeo.FontNice(7, zgeo.FontStyleNormal))

	view.Add(label, zgeo.Center)
	return view
}

func makeEvents() []Event {
	var es []Event
	now := time.Now()
	for t := now.Add(-time.Hour); t.Before(now.Add(time.Hour * 2)); t = t.Add(time.Second * 3) {
		for _, lane := range lanes {
			if rand.Int31n(10) == 5 {
				continue
			}
			for _, row := range lane.Rows {
				if rand.Int31n(4) == 2 {
					continue
				}
				var e Event
				e.ID = rand.Int63()
				e.LaneID = lane.ID
				e.LaneRowID = row.ID
				e.Start = t.Add(time.Duration(rand.Float64()) * time.Second)
				e.End = t.Add(time.Second*3 + time.Duration(rand.Float64())*time.Second)
				es = append(es, e)
			}
		}
	}
	return es
}
*/
