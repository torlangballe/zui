//go:build zui && js

package zhorblocks

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zlabel"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zint"
	"github.com/torlangballe/zutil/ztime"
)

type Event struct {
	ID        int64
	Start     time.Time
	End       time.Time
	LaneID    int64
	LaneRowID int64
}

const (
	dogRowType = 55
	dogLaneID  = 512
)

var testLanes = []Lane{
	Lane{ID: dogLaneID, Name: "Dog",
		Rows: []Row{
			Row{Name: "German Shepard", Height: 22},
			Row{Name: "Poodle", Height: 40},
			Row{Name: "Snauzer", Height: 50},
			Row{Name: "Dashund", Height: 30},
		},
	},
}
var testEvents = makeEvents()

func MakeHorBlocksTestView(id string, delete bool) zview.View {
	if delete {
		return nil
	}
	v := NewHorBlocksView(3, 14)
	v.SetContentHeight(2200)
	v.GetViewFunc = func(index int) zview.View {
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
			if x < v.viewSize.W/2 {
				d = -33
			}
			index := v.CurrentIndex()
			v.SetCurrentIndex(index - d)
		})
		return c
	}
	for _, a := range []zgeo.Alignment{zgeo.Right} { //, zgeo.Right} {
		name := a.String()
		pole := zcontainer.StackViewVert(name + "-pole")
		pole.SetMinSize(zgeo.SizeD(10, 200))
		pole.SetBGColor(zgeo.ColorBlue)
		pole.SetZIndex(5555)
		v.Overlay.Add(pole, zgeo.Top|a).Free = true // |zgeo.VertExpand
	}
	return v
}

func MakeHorEventsTestView(id string, delete bool) zview.View {
	if delete {
		return nil
	}
	opts := Options{
		StoreKey:             "test-horevents-store",
		BlocksIndexGetWidth:  3,
		BlockIndexCacheDelta: 10,
		BlockDuration:        time.Second * 20,
		StartTime:            time.Now(),
		ShowNowPole:          true,
		TimeAxisHeight:       30,
	}
	v := NewEventsView(nil, opts)
	v.GetEventViewsFunc = func(blockIndex int, isNewView bool, got func(childView zview.View, x int, cellBox zgeo.Size, laneID, rowType int64)) {
		// si := -1
		start := v.IndexToTime(float64(blockIndex))
		end := start.Add(v.BlockDuration)
		for _, e := range testEvents {
			if !e.Start.Before(end) {
				break
			}
			if e.End.After(start) {
				view := makeEventView(v, e)
				x := v.TimeToXInCorrectBlock(e.Start)
				cellBox := zgeo.SizeD(40, 30)
				got(view, int(x), cellBox, e.LaneID, dogRowType)
			}
		}
	}
	v.SetLanes(testLanes)
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
		for _, lane := range testLanes {
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
