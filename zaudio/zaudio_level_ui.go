package zaudio

import (
	"sort"
	"time"

	"github.com/torlangballe/zui/zcanvas"
	"github.com/torlangballe/zui/zcustom"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zmap"
	"github.com/torlangballe/zutil/zmath"
	"github.com/torlangballe/zutil/ztime"
)

type LevelView struct {
	zcustom.CustomView
	KeepMinMaxSecs     float64
	Colors             map[float64]zgeo.Color
	currentLevel       float64
	oldExtreme         zmath.RangeF64
	currentExtreme     zmath.RangeF64
	cellHeight         float64
	lastExtremesUpdate time.Time
}

const updateExtreamsSeconds = 10

func NewLevelView(minSize zgeo.Size) *LevelView {
	v := &LevelView{}
	v.Init(v, minSize)
	return v
}

func (v *LevelView) Init(view zview.View, minSize zgeo.Size) {
	v.CustomView.Init(view, "audio-level")
	v.SetMinSize(minSize)
	v.SetDrawHandler(v.draw)
	v.lastExtremesUpdate = time.Now().Add(-time.Second * 10)
	v.cellHeight = 10
	v.Colors = map[float64]zgeo.Color{
		0.7: zgeo.ColorGreen,
		0.8: zgeo.ColorYellow,
		0.9: zgeo.ColorOrange,
		1:   zgeo.ColorRed,
	}
}

func (v *LevelView) SetLevel(level float64) {
	v.currentLevel = level
	v.currentExtreme.Add(level)
	v.oldExtreme.Add(level)
	if ztime.Since(v.lastExtremesUpdate) > updateExtreamsSeconds {
		v.lastExtremesUpdate = time.Now()
		v.oldExtreme = v.currentExtreme
	}
}

func (v *LevelView) levelToYCol(level float64) float64 {
	r := v.LocalRect()
	return r.Max().Y - r.Size.H*level
}

func (v *LevelView) draw(rect zgeo.Rect, canvas *zcanvas.Canvas, view zview.View) {
	clen := len(v.Colors)
	colors := make([]zgeo.Color, clen*2)
	locations := make([]float64, clen*2)
	keys := zmap.Keys(v.Colors)
	sort.Float64Slice(keys).Sort()
	// for _, n := range keys {
	// 	col := v.Colors[n]
	// 	zlog.Info("Levels:", n, col)
	// }
	lr := v.LocalRect()
	r := lr
	r.SetMinY(v.levelToYCol(v.currentLevel))
	path := zgeo.PathNewRect(r, zgeo.SizeNull)
	canvas.DrawGradient(path, colors, lr.BottomLeft(), lr.TopLeft(), locations)
	canvas.SetColor(zgeo.ColorBlack)
	canvas.StrokePath(path, 1, zgeo.PathLineButt)
}
