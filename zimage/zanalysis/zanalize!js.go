//xxxgo:build !js

package zanalysis

import (
	"fmt"
	"image"
	"slices"
	"sort"

	"github.com/torlangballe/zui/zimage"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zmath"
	"github.com/torlangballe/zutil/zslice"
)

type SimpleAnalytics struct {
	Size           zgeo.ISize
	BlurAmount     float64
	FlatsAmount    float64
	EdgesAmount    float64
	BlockFrequency zgeo.IPos
	BlockOffset    zgeo.IPos
	BlockAmount    float64
}

type counts struct {
	blurs                    map[int]int
	flats                    map[int]int
	edgePoints               int
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
	DebugImage       *image.Image
}

func NewImageInfo() *ImageInfo {
	info := &ImageInfo{}
	info.hCounts = makeCounts()
	info.vCounts = makeCounts()
	info.BlurMinThreshold = 0.001
	info.BlurMaxThreshold = 0.005
	info.EdgeMinThreshold = 0.05
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

func (info *ImageInfo) Analyze(img image.Image, debugImagePath string) {
	goRect := img.Bounds()
	var oldRow []float64 = nil
	info.Size = zgeo.RectFromGoRect(img.Bounds()).Size.ISize()
	info.hCounts = makeCounts()
	info.vCounts = makeCounts()
	var vertLengths = make([]lengths, int(goRect.Max.X))
	var debugImage *image.RGBA
	if debugImagePath != "" {
		debugImage = image.NewRGBA(goRect)
	}
	row := make([]float64, int(goRect.Max.X))
	for y := range goRect.Max.Y {
		clear(row)
		var oldCol float64 = -1
		var horLengths lengths
		for x := range goRect.Max.X {
			goCol := img.At(x, y)
			col := zgeo.ColorFromGo(goCol)
			fcol := float64(col.GrayScale())
			row[x] = fcol
			if oldCol != -1 {
				hdiff := zmath.Abs(fcol - oldCol)
				if debugImage != nil {
					debugImage.Set(x, y, zgeo.GoGrayColor(float32(hdiff)))
				}
				info.setDiff(&horLengths, x, hdiff, &info.hCounts)
				v1 := info.setEdgeToPerp(x, hdiff, &vertLengths[x], &info.vCounts)
				if v1 != 0 && debugImage != nil {
					for y1 := max(0, y-v1); y1 < y; y1++ {
						debugImage.Set(x, y1, zgeo.ColorMagenta.GoColor())
					}
				}
			}
			if oldRow != nil {
				vdiff := zmath.Abs(fcol - oldRow[x])
				info.setDiff(&vertLengths[x], x, vdiff, &info.vCounts)
				h1 := info.setEdgeToPerp(y, vdiff, &horLengths, &info.hCounts)
				// zlog.Info("Set hcounts perp:", vdiff, horLengths.perpEdge)
				if debugImage != nil && h1 != 0 {
					for x1 := max(0, x-h1); x1 < x; x1++ {
						debugImage.Set(x1, y, zgeo.ColorRed.GoColor())
					}
				}
			}
			oldCol = fcol
		}
		oldRow = slices.Clone(row)
	}
	if debugImage != nil {
		zimage.GoImageToPNGFile(debugImage, debugImagePath)
	}
}

type freqInfo struct {
	amount int
	freq   int
	offset int
}

func (c *counts) getBlockFrequencyAndOffset() (freq, offset int, amount float64) {
	var w int
	// zlog.Info("getBlockFrequencyAndOffset:", len(c.perpendicularEdgeLengths))
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
			var lines int
			var sum int
			for i := w; i < clen-(blockMax-w); i += freq {
				sum += xcounts[i]
				lines++
			}
			if lines == 0 {
				continue
			}
			var f freqInfo
			f.amount = sum / lines
			f.freq = freq
			f.offset = w
			order = append(order, f)
		}
	}
	for i := 0; i < len(order); i++ {
		for j := i + 1; j < len(order); j++ {
			if order[i].freq == order[j].freq {
				order[i].amount = max(order[i].amount, order[j].amount)
				zslice.RemoveAt(&order, j)
				j--
			}
		}
	}
	for i := 0; i < len(order); i++ {
		for j := 0; j < len(order) && i < len(order); j++ {
			if i != j && order[j].freq > order[i].freq && order[j].freq%order[i].freq == 0 {
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
	sort.Slice(order, func(i, j int) bool {
		return order[i].amount < order[j].amount
	})
	if len(order) < 2 {
		return 0, 0, 0
	}
	// for _, o := range order {
	// 	zlog.Info("Freqs:", o.freq, o.amount, o.offset)
	// }
	best := order[len(order)-1]
	next := order[len(order)-2]
	amount = float64(best.amount) / float64(next.amount)
	// zlog.Info("Best Freq:", best.amount, best.freq, "next:", next.amount, next.freq, float64(best.amount)/float64(next.amount))
	return best.freq, best.offset, amount
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

func (info *ImageInfo) BlockFrequency() (freq, offset zgeo.IPos, amount zgeo.Pos) {
	fX, oX, amountX := info.hCounts.getBlockFrequencyAndOffset()
	fY, oY, amountY := info.vCounts.getBlockFrequencyAndOffset()
	if fX == 0 || fY == 0 {
		fX = 0
		fY = 0
		amountX = 0
		amountY = 0
	}
	return zgeo.IPos{X: fX, Y: fY}, zgeo.IPos{X: oX, Y: oY}, zgeo.PosD(amountX, amountY)
}

func (info *ImageInfo) PrintInfo() {
	fmt.Print("blur:", info.BlurAmount().Average())
	fmt.Print(" flat:", info.FlatAmount().Average())
	fmt.Print(" edges:", info.EdgePointsAmount().Average())
	freq, offset, amount := info.BlockFrequency()
	fmt.Print(" bfreq:", freq, offset, amount)
}

func (info *ImageInfo) SimpleAnalytics() SimpleAnalytics {
	freq, offset, amount := info.BlockFrequency()
	return SimpleAnalytics{
		Size:           info.Size,
		BlurAmount:     info.BlurAmount().Average(),
		FlatsAmount:    info.FlatAmount().Average(),
		EdgesAmount:    info.EdgePointsAmount().Average(),
		BlockFrequency: freq,
		BlockOffset:    offset,
		BlockAmount:    amount.Average(),
	}
}
