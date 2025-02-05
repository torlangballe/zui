//xxxgo:build !js

package zimage

import (
	"fmt"
	"image"

	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
	"github.com/torlangballe/zutil/zmath"
)

type counts struct {
	blurs                    map[int]int
	flats                    map[int]int
	edgePoints               int
	perpendicularEdgeLengths map[int]map[int]int
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
	blurMinThreshold float64
	blurMaxThreshold float64
	edgeMinThreshold float64
}

func NewImageInfo() *ImageInfo {
	info := &ImageInfo{}
	info.hCounts = makeCounts()
	info.vCounts = makeCounts()
	info.blurMinThreshold = 0.001
	info.blurMaxThreshold = 0.005
	info.edgeMinThreshold = 0.05
	return info
}

func makeCounts() counts {
	var c counts
	c.blurs = map[int]int{}
	c.flats = map[int]int{}
	c.perpendicularEdgeLengths = map[int]map[int]int{}
	return c
}

func (info *ImageInfo) setEdgeToPerp(i int, diff float64, perpLens *lengths, counts *counts) int {
	if diff > info.edgeMinThreshold {
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
	if diff > info.edgeMinThreshold {
		counts.edgePoints++
	}
	if diff > info.blurMinThreshold {
		if l.flat != 0 {
			(*counts).flats[l.flat]++
			l.flat = 0
		}
		if diff < info.blurMaxThreshold {
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

func (info *ImageInfo) Analyze(img image.Image) {
	goRect := img.Bounds()
	var oldRow []float64 = nil
	info.Size = zgeo.RectFromGoRect(img.Bounds()).Size.ISize()
	info.hCounts = makeCounts()
	info.vCounts = makeCounts()
	var vertLengths = make([]lengths, int(goRect.Max.X))
	// outImage := image.NewRGBA(goRect)
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
				// outImage.Set(x, y, zgeo.GoGrayColor(float32(hdiff)))
				info.setDiff(&horLengths, x, hdiff, &info.hCounts)
				v1 := info.setEdgeToPerp(x, hdiff, &vertLengths[x], &info.vCounts)
				if v1 != 0 {
					// for y1 := max(0, y-v1); y1 < y; y1++ {
					// outImage.Set(x, y1, zgeo.ColorMagenta.GoColor())
					// }
				}
			}
			if oldRow != nil {
				vdiff := zmath.Abs(fcol - oldRow[x])
				info.setDiff(&vertLengths[x], x, vdiff, &info.vCounts)
				// h1 := info.setEdgeToPerp(y, vdiff, &horLengths, &info.hCounts)
				// if h1 != 0 {
				// 	for x1 := max(0, x-h1); x1 < x; x1++ {
				// 		outImage.Set(x1, y, zgeo.ColorRed.GoColor())
				// 	}
				// }
			}
			oldCol = fcol
		}
		oldRow = row
	}
}

func (c *counts) getBlockFrequencyAndOffset() (freq, offset int, nextRatio float64) {
	var w int
	for x, _ := range c.perpendicularEdgeLengths {
		w = max(w, x)
	}
	w += 1
	xcounts := make([]int, w)
	for x, cs := range c.perpendicularEdgeLengths {
		// zlog.Info("HEdges:", clen, len(counts))
		for elen, c := range cs {
			xcounts[x] += elen * c
		}
	}
	const blockMax = 32
	clen := len(xcounts)
	var best, bestFreq, bestOffset, nextBest int // , nextBestFreq
	for freq := 8; freq <= blockMax; freq *= 2 {
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
			n := sum / lines
			if n > best {
				if best != 0 && 100*n/best < 105 {
					continue
				}
				nextBest = best
				// nextBestFreq = bestFreq
				best = n
				bestFreq = freq
				bestOffset = w
			}
		}
	}
	if bestFreq == 0 {
		return 0, 0, 0
	}
	// zlog.Info("Next Best Freq:", nextBestFreq)
	return bestFreq, bestOffset, float64(nextBest) / float64(best)
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

func (info *ImageInfo) BlockFrequency() (freq, offset zgeo.IPos, nextRatio zgeo.Pos) {
	fX, oX, nX := info.hCounts.getBlockFrequencyAndOffset()
	fY, oY, nY := info.vCounts.getBlockFrequencyAndOffset()
	if fX == 0 || fY == 0 || (fX/fY != 4 && fX/fY != 2 && fY/fX != 4 && fY/fX != 2) {
		fX = 0
		fY = 0
	}
	return zgeo.IPos{X: fX, Y: fY}, zgeo.IPos{X: oX, Y: oY}, zgeo.PosD(nX, nY)
}

func (info *ImageInfo) PrintInfo() {
	fmt.Print("blur:", info.BlurAmount().Average())
	fmt.Print(" flat:", info.FlatAmount().Average())
	fmt.Print(" edges:", info.EdgePointsAmount().Average())
	freq, offset, nextRatio := info.BlockFrequency()
	fmt.Print("bfreq:", freq, offset, nextRatio)
}

func (info *ImageInfo) SimpleAnalytics() SimpleAnalytics {
	freq, offset, _ := info.BlockFrequency()
	return SimpleAnalytics{
		Size:           info.Size,
		BlurAmount:     info.BlurAmount().Average(),
		FlatsAmount:    info.FlatAmount().Average(),
		EdgesAmount:    info.EdgePointsAmount().Average(),
		BlockFrequency: freq,
		BlockOffset:    offset,
	}
}
