//xxxgo:build !js

package zanalysis

import (
	"fmt"
	"image"
	"slices"
	"sort"

	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zmath"
	"github.com/torlangballe/zutil/zslice"
)

type SimpleAnalytics struct {
	Size           zgeo.ISize
	BlurAmount     float64
	FlatsAmount    float64
	EdgesAmount    float64
	BlockFrequency zgeo.IRect
	BlockAmount    float64
	BlockBetter    float64
	Blockiness     float64
	DebugImage     image.Image `json:"-"`
}

type counts struct {
	blurs                    map[int]int
	flats                    map[int]int
	edgePoints               int
	oppositeLength           int
	perpendicularEdgeLengths map[int]map[int]int // map of x/y coordinate to histogram of continuous edge-length counts for that x/y.

}

type lengths struct {
	blur     int
	flat     int
	perpEdge int
}

type ImageInfo struct {
	Size             zgeo.ISize
	hCounts          counts
	vCounts          counts
	BlurMinThreshold float64
	BlurMaxThreshold float64
	EdgeMinThreshold float64
	WorkingFrame     zgeo.Rect

	DebugImageBackgroundOnly bool
	DebugImage               *image.RGBA
	DebugImageBlockFreq      zgeo.IRect
}

func (s *SimpleAnalytics) PrintInfo() {
	fmt.Print("zanalize: blur:", s.BlurAmount)
	fmt.Print(" flat:", s.FlatsAmount)
	fmt.Print(" edges:", s.EdgesAmount)
	fmt.Println(" bfreq:", s.BlockFrequency.Size, "boff:", s.BlockFrequency.Pos, "bamount:", s.BlockAmount, "bbetter:", s.BlockBetter)
}

func NewImageInfo() *ImageInfo {
	info := &ImageInfo{}
	info.hCounts = makeCounts()
	info.vCounts = makeCounts()
	info.BlurMinThreshold = 0.001
	info.BlurMaxThreshold = 0.005
	info.EdgeMinThreshold = 0.09 // 0.05
	return info
}

func makeCounts() counts {
	var c counts
	c.blurs = map[int]int{}
	c.flats = map[int]int{}
	c.perpendicularEdgeLengths = map[int]map[int]int{}
	return c
}

// For an x/y i, setEdgeToPerp adds 1 to an existing perpLens.perpEdge if still contrast
// If low contrast, and existing edge count exists, its count is used to create a histogram count of lengths for that coordinate
func (info *ImageInfo) setEdgeToPerp(i int, diff float64, perpLens *lengths, counts *counts) int {
	if diff > info.EdgeMinThreshold {
		perpLens.perpEdge++
		return 0
	}
	if perpLens.perpEdge > 0 {
		c := counts.perpendicularEdgeLengths[i]
		if c == nil {
			c = map[int]int{}
			counts.perpendicularEdgeLengths[i] = c
		}
		c[perpLens.perpEdge]++
		l := perpLens.perpEdge
		perpLens.perpEdge = 0
		return l
	}
	return 0
}

func (info *ImageInfo) setDiff(l *lengths, i int, diff float64, counts *counts) {
	if diff > info.EdgeMinThreshold {
		counts.edgePoints++
	}
	if diff > info.BlurMinThreshold {
		if l.flat != 0 {
			(*counts).flats[l.flat]++
			l.flat = 0
		}
		if diff < info.BlurMaxThreshold {
			l.blur++
		} else {
			if l.blur > 1 {
				(*counts).blurs[l.blur]++
			}
			l.blur = 0
		}
	} else {
		l.flat++
	}
}

func isXYOnBlockFrequency(x, y int, bf zgeo.IRect) bool {
	if !isXOnBlockFrequency(x, bf) {
		return false
	}
	return isYOnBlockFrequency(y, bf)
}

func isXOnBlockFrequency(x int, bf zgeo.IRect) bool {
	if bf.Size.W == 0 {
		return false
	}
	if (x-bf.Pos.X+1)%bf.Size.W != 0 {
		return false
	}
	return true
}

func isYOnBlockFrequency(y int, bf zgeo.IRect) bool {
	if bf.Size.W == 0 {
		return false
	}
	if (y-bf.Pos.Y+1)%bf.Size.H != 0 {
		return false
	}
	return true
}

func (info *ImageInfo) Analyze(img image.Image) {
	goRect := img.Bounds()
	var oldRow []float64 = nil
	info.Size = zgeo.RectFromGoRect(img.Bounds()).Size.ISize()
	info.hCounts = makeCounts()
	info.vCounts = makeCounts()
	var vertLengths = make([]lengths, int(goRect.Max.X))
	info.hCounts.oppositeLength = goRect.Dy()
	info.vCounts.oppositeLength = goRect.Dx()

	blue := zgeo.ColorBlue.GoColor()
	magenta := zgeo.ColorMagenta.GoColor()
	yellow := zgeo.ColorYellow.GoColor()
	orange := zgeo.ColorOrange.GoColor()
	darkYellow := zgeo.ColorYellow.Mixed(zgeo.ColorBlack, 0.7).GoColor()
	row := make([]float64, int(goRect.Max.X))
	sy := 0
	sx := 0
	ex := goRect.Max.X
	ey := goRect.Max.Y
	if !info.WorkingFrame.IsNull() {
		sy = int(info.WorkingFrame.Pos.Y)
		ey = int(info.WorkingFrame.Max().Y)
		sx = int(info.WorkingFrame.Pos.X)
		ex = int(info.WorkingFrame.Max().X)
	}
	for y := sy; y < ey; y++ {
		clear(row)
		var oldCol float64 = -1
		var horLengths lengths
		for x := sx; x < ex; x++ {
			goCol := img.At(x, y)
			col := zgeo.ColorFromGo(goCol)
			dark := col.Mixed(zgeo.ColorBlack, 0.6)
			if info.DebugImage != nil {
				info.DebugImage.Set(x, y, dark.GoColor())
				if isXYOnBlockFrequency(x, y, info.DebugImageBlockFreq) {
					info.DebugImage.Set(x, y, darkYellow)
				}
			}
			fcol := float64(col.GrayScale())
			row[x] = fcol
			if oldCol != -1 {
				hdiff := zmath.Abs(fcol - oldCol)
				// if info.DebugImage != nil {
				// 	info.DebugImage.Set(x, y, zgeo.GoGrayColor(float32(hdiff)))
				// }
				info.setDiff(&horLengths, x, hdiff, &info.hCounts)
				v1 := info.setEdgeToPerp(x, hdiff, &vertLengths[x], &info.vCounts)
				if v1 != 0 && info.DebugImage != nil {
					col := blue
					if isXOnBlockFrequency(x, info.DebugImageBlockFreq) {
						col = yellow
					} else {
						// continue
					}
					for y1 := max(0, y-v1); y1 < y; y1++ {
						info.DebugImage.Set(x, y1, col)
					}
				}
			}
			if oldRow != nil {
				vdiff := zmath.Abs(fcol - oldRow[x])
				info.setDiff(&vertLengths[x], x, vdiff, &info.vCounts)
				h1 := info.setEdgeToPerp(y, vdiff, &horLengths, &info.hCounts)
				// zlog.Info("Set hcounts perp:", vdiff, horLengths.perpEdge)
				if info.DebugImage != nil && h1 != 0 {
					col := magenta
					if isYOnBlockFrequency(y, info.DebugImageBlockFreq) {
						col = orange
					} else {
						// continue
					}
					for x1 := max(0, x-h1); x1 < x; x1++ {
						info.DebugImage.Set(x1, y, col)
						// zlog.Info("Yellow2", x1, y)
					}
				}
			}
			oldCol = fcol
		}
		oldRow = slices.Clone(row)
	}

}

func SimpleAnalysesOfSquares(img image.Image, squareSize int, blockinessBetterCutOff float64) (debugImage image.Image, squareCount, squareTotal int) {
	type squares struct {
		freq   zgeo.IRect
		amount float64
		better float64
	}
	s := zgeo.RectFromGoRect(img.Bounds()).Size.ISize()
	sx := s.W % int(squareSize) / 2
	sy := s.H % int(squareSize) / 2
	dbImage := image.NewRGBA(img.Bounds())
	for x := sx; x < s.W; x += squareSize {
		for y := sy; y < s.H; y += squareSize {
			squareTotal++
			info := NewImageInfo()
			info.WorkingFrame = zgeo.RectFromXYWH(float64(x), float64(y), float64(squareSize), float64(squareSize))
			info.Analyze(img)
			a := info.SimpleAnalytics(false)
			info.DebugImage = dbImage
			if a.BlockBetter < blockinessBetterCutOff {
				info.DebugImageBackgroundOnly = true
			} else {
				info.DebugImageBlockFreq = a.BlockFrequency
				squareCount++
			}
		}
	}
	return dbImage, squareCount, squareSize
}

type freqInfo struct {
	amount float64
	better float64
	freq   int
	offset int
}

func (c *counts) getBlockFrequencyAndOffset(print bool) (freq, offset int, amount, better float64) {
	var w int
	for x, _ := range c.perpendicularEdgeLengths {
		w = max(w, x)
	}
	w += 1
	xcounts := make([]int, w)
	for x, cs := range c.perpendicularEdgeLengths {
		// zlog.Info("HEdges:", clen, len(counts))
		for elen, c := range cs {
			xcounts[x] += elen * c // length x count of that length
		}
	}
	// for x, c := range xcounts {
	// 	zlog.Info("XCount:", x, c)
	// }
	const blockMax = 32
	clen := len(xcounts)
	// var best, bestFreq, bestOffset, nextBestFreq, nextBest int
	var order []freqInfo
	for freq := 8; freq <= blockMax; freq++ {
		for w := range blockMax {
			if w != 0 && w%freq == 0 {
				continue
			}
			var lines int
			var sum int
			for i := w; i < clen-(blockMax-w); i += freq {
				// if w == 0 && xcounts[i] != 0 {
				// 	zlog.Info("XCount:", i, xcounts[i], "freq:", freq)
				// }
				sum += xcounts[i]
				lines++
			}
			if lines == 0 {
				continue
			}
			var f freqInfo
			f.amount = float64(sum) / float64(lines) / float64(c.oppositeLength)
			// zlog.Info("AMOUNT:", clen, f.amount, "freq:", freq, "x:", w, sum, lines)
			f.freq = freq
			f.offset = w
			order = append(order, f)
		}
	}

	// for i := 0; i < len(order); i++ {
	// 	for j := i + 1; j < len(order); j++ {
	// 		if order[i].freq == order[j].freq {
	// 			if order[i].amount < order[j].amount {
	// 				order[i] = order[j]
	// 				zslice.RemoveAt(&order, j)
	// 				j--
	// 			}
	// 		}
	// 	}
	// }
	for i := 0; i < len(order); i++ {
		for j := 0; j < len(order) && i < len(order); j++ {
			if i != j && order[j].freq >= order[i].freq && order[j].freq%order[i].freq == 0 && order[j].offset%order[i].freq == 0 {
				order[i].amount += order[j].amount
				zslice.RemoveAt(&order, j)
				j--
			}
		}
	}
	for i := 0; i < len(order); i++ {
		if order[i].amount == 0 {
			zslice.RemoveAt(&order, i)
			i--
		}
	}
	if len(order) < 2 {
		return 0, 0, 0, 0
	}
	sort.Slice(order, func(i, j int) bool {
		return order[i].amount > order[j].amount
	})
	// bestFreq := order[0].freq
	// order = slices.DeleteFunc(order, func(f freqInfo) bool {
	// 	zlog.Info("==", f.freq, bestFreq, f.freq == 0)
	// 	return f.freq > bestFreq || f.freq == 0
	// })
	// for _, o := range order {
	// 	zlog.Info("Penultimate2:", o.freq, o.amount)
	// }
	if len(order) < 2 {
		return 0, 0, 0, 0
	}
	if print {
		// for _, o := range order {
		// 	zlog.Info("freqs:", o.freq, o.amount, o.offset)
		// }
	}
	best := order[0]
	next := order[1]
	better = float64(best.amount) / float64(next.amount)
	// zlog.Info("Best Freq:", best.amount, best.freq, "next:", next.amount, next.freq, float64(best.amount)/float64(next.amount))
	return best.freq, best.offset, best.amount, better
}

func (info *ImageInfo) BlurAmount() zgeo.Pos {
	var w, h int
	sum := info.Size.W * info.Size.H
	for len, count := range info.hCounts.blurs {
		// zlog.Info("HBlur:", len, count, i)
		w += len * count
	}
	for len, count := range info.vCounts.blurs {
		h += len * count
	}
	blur := zgeo.PosD(float64(w)/float64(sum), float64(h)/float64(sum))
	return blur
}

func (info *ImageInfo) IsBlurry() bool {
	a := info.BlurAmount()
	return a.X > 0.1 && a.Y > 0.1
}

func (info *ImageInfo) FlatAmount() zgeo.Pos {
	var w, h int
	sum := info.Size.W * info.Size.H
	for len, count := range info.hCounts.flats {
		w += len * count
	}
	for len, count := range info.vCounts.flats {
		h += len * count
	}
	flat := zgeo.PosD(float64(w)/float64(sum), float64(h)/float64(sum))
	return flat
}

func (info *ImageInfo) EdgePointsAmount() zgeo.Pos {
	sum := info.Size.W * info.Size.H
	edges := zgeo.PosD(float64(info.hCounts.edgePoints)/float64(sum), float64(info.vCounts.edgePoints)/float64(sum))
	return edges
}

func (info *ImageInfo) BlockFrequency(print bool) (offset zgeo.IPos, freq zgeo.ISize, amount, better zgeo.Pos) {
	fX, oX, amountX, betterX := info.hCounts.getBlockFrequencyAndOffset(print)
	fY, oY, amountY, betterY := info.vCounts.getBlockFrequencyAndOffset(print)
	if fX == 0 || fY == 0 {
		fX = 0
		fY = 0
		amountX = 0
		amountY = 0
	}
	return zgeo.IPos{X: oX, Y: oY}, zgeo.ISize{W: fX, H: fY}, zgeo.PosD(amountX, amountY), zgeo.PosD(betterX, betterY)
}

func (info *ImageInfo) SimpleAnalytics(print bool) SimpleAnalytics {
	offset, freq, amount, better := info.BlockFrequency(print)
	return SimpleAnalytics{
		Size:           info.Size,
		BlurAmount:     info.BlurAmount().Average(),
		FlatsAmount:    info.FlatAmount().Average(),
		EdgesAmount:    info.EdgePointsAmount().Average(),
		BlockFrequency: zgeo.IRect{Pos: offset, Size: freq},
		BlockAmount:    amount.Average(),
		BlockBetter:    better.Average(),
		Blockiness:     amount.Average() * (better.Average() - 1),
	}
}
