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
	DebugImage     image.Image `json:"-"`
}

type count struct {
	blur        int
	flats       int
	edgePoints  int
	perpLengths []zmath.Range[int]
}

// type lengths struct {
// 	blur     int
// 	flat     int
// 	perpEdge int
// }

type ImageInfo struct {
	Size                     zgeo.ISize
	hCounts                  []count
	vCounts                  []count
	BlurMinThreshold         float64
	BlurMaxThreshold         float64
	EdgeMinThreshold         float64
	LimitFrame               zgeo.IRect
	DebugImageBackgroundOnly bool
	DebugImage               *image.RGBA
	DebugImageBlockFreq      zgeo.IRect
}

const blockMax = 32

func (s *SimpleAnalytics) PrintInfo() {
	fmt.Print("zanalize: blur:", s.BlurAmount)
	fmt.Print(" flat:", s.FlatsAmount)
	fmt.Print(" edges:", s.EdgesAmount)
	fmt.Println(" bfreq:", s.BlockFrequency.Size, "boff:", s.BlockFrequency.Pos, "bamount:", s.BlockAmount, "bbetter:", s.BlockBetter)
}

func NewImageInfo() *ImageInfo {
	info := &ImageInfo{}
	info.BlurMinThreshold = 0.001
	info.BlurMaxThreshold = 0.005
	info.EdgeMinThreshold = 0.074 // 0.074 gave best results of 81 test image pairs
	return info
}

func (info *ImageInfo) setHorEdge(x int, vdiff float64, c *count, endX int) int {
	return info.setHorVertEdge(x, vdiff, c, endX)
}

func (info *ImageInfo) setVertEdge(y int, hdiff float64, c *count, endY int) int {
	return info.setHorVertEdge(y, hdiff, c, endY)
}

// For an x/y i, setEdgeToPerp adds 1 to an existing perpLens.perpEdge if still contrast
func (info *ImageInfo) setHorVertEdge(i int, diff float64, c *count, end int) int {
	var lastRange *zmath.Range[int]
	if len(c.perpLengths) > 0 {
		lastRange = &c.perpLengths[len(c.perpLengths)-1]
	}
	edgy := (diff > info.EdgeMinThreshold)
	if edgy {
		if lastRange != nil {
			if lastRange.Max == i {
				lastRange.Max++
				if i == end-1 {
					return lastRange.Length()
				}
				return 0
			}
		}
		r := zmath.MakeRange(i, i+1)
		c.perpLengths = append(c.perpLengths, r)
		if i == end-1 {
			return 1
		}
		return 0
	}
	if lastRange != nil && lastRange.Max == i {
		return lastRange.Length()
	}
	return 0
}

func (info *ImageInfo) setDiff(i int, diff float64, counts []count) {
	if diff > info.EdgeMinThreshold {
		counts[i].edgePoints++
	}
	if diff > info.BlurMinThreshold {
		if diff < info.BlurMaxThreshold {
			counts[i].blur++
		}
	} else {
		counts[i].flats++
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

type freqInfo struct {
	amount float64
	better float64
	freq   int
	offset int
}

func getBlockFrequencyAndOffset(print bool, counts []count, frame zgeo.IRect) (freq, offset int, amount, better float64) {
	clen := len(counts)
	xcounts := make([]int, clen)
	// maxX := frame.Pos.X + frame.Size.W
	// maxY := frame.Pos.Y + frame.Size.H
	// for i := frame.Pos.X; i < maxX; i++ {
	// 	for _, r := range counts[i].perpLengths {
	// 		if r.Max >= frame.Pos.Y && r.Min <= maxY {
	// 			xcounts[i] += r.Length() // we do whole for now, even if partially outside frame
	// 		}
	// 	}
	// }
	maxX := frame.Pos.X + frame.Size.W
	maxY := frame.Pos.Y + frame.Size.H
	for i := frame.Pos.X; i < maxX; i++ {
		for _, r := range counts[i].perpLengths {
			if r.Max >= frame.Pos.Y && r.Min <= maxY {
				xcounts[i] += r.Length() // we do whole for now, even if partially outside frame
			}
		}
	}
	var order []freqInfo
	count := 0
	for freq := 8; freq <= blockMax; freq++ {
		for w := range blockMax {
			if w != 0 && w%freq == 0 {
				continue
			}
			var sum, lines int
			start := zmath.RoundToMod(frame.Pos.X, blockMax) + w
			end := min(maxX, clen)
			for i := start; i < end; i += freq {
				// if w == 0 && xcounts[i] != 0 {
				// 	zlog.Info("XCount:", i, xcounts[i], "freq:", freq)
				// }
				// if i >= frame.Pos.X {
				sum += xcounts[i]
				lines += frame.Size.H
				count++
				// }
			}
			var f freqInfo
			f.amount = float64(sum) / float64(lines)
			// if f.amount != 0 {
			// 	zlog.Info("AMOUNT:", clen, f.amount, frame.Rect().Size.Area(), f.amount/frame.Rect().Size.Area())
			// }
			f.freq = freq
			f.offset = w
			order = append(order, f)
		}
	}
	// zlog.Info("Count:", count, frame)
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
	if print {
		// for _, o := range order {
		// 	zlog.Info("freqs:", o.freq, o.amount, o.offset)
		// }
	}
	best := order[0]
	next := order[1]
	better = float64(best.amount) / float64(next.amount)
	// zlog.Info("Best Freq Count:", frame.Pos, len(order), better)
	// zlog.Info("Best Freq:", frame.Pos, int(frame.Rect().Size.Area()*best.amount))
	return best.freq, best.offset, best.amount, better
}

func (info *ImageInfo) Analyze(img image.Image) {
	goRect := img.Bounds()
	var oldRow []zgeo.Color = nil
	info.Size = zgeo.RectFromGoRect(img.Bounds()).Size.ISize()

	info.vCounts = make([]count, info.Size.H)
	info.hCounts = make([]count, info.Size.W)

	blue := zgeo.ColorBlue.GoColor()
	magenta := zgeo.ColorMagenta.GoColor()
	yellow := zgeo.ColorYellow.GoColor()
	orange := zgeo.ColorOrange.GoColor()
	darkYellow := zgeo.ColorYellow.Mixed(zgeo.ColorBlack, 0.5).GoColor()
	row := make([]zgeo.Color, int(goRect.Max.X))
	sx := 0
	sy := 0
	ex := goRect.Max.X
	ey := goRect.Max.Y
	if info.LimitFrame.Size.W != 0 {
		sx = info.LimitFrame.Pos.X
		sy = info.LimitFrame.Pos.Y
		ex = sx + info.LimitFrame.Size.W
		ey = sy + info.LimitFrame.Size.H
	}
	// if !info.DebugImageBackgroundOnly {
	// zlog.Info("Analyse:", sx, sy, ex, ey")
	// }
	for y := sy; y < ey; y++ {
		// zlog.Info("Y:", y)
		clear(row)
		var oldCol zgeo.Color
		for x := sx; x < ex; x++ {
			goCol := img.At(x, y)
			col := zgeo.ColorFromGo(goCol)
			dark := col.Mixed(zgeo.ColorBlack, 0.6)
			dotCol := dark.GoColor()
			if info.DebugImage != nil {
				if isXYOnBlockFrequency(x, y, info.DebugImageBlockFreq) {
					dotCol = darkYellow
				}
				info.DebugImage.Set(x, y, dotCol)
			}
			row[x] = col
			if oldCol.Valid {
				hdiff := float64(col.Difference(oldCol))
				info.setDiff(x, hdiff, info.hCounts)
				v1 := info.setVertEdge(y, hdiff, &info.hCounts[x], info.Size.H)
				if v1 != 0 && info.DebugImage != nil {
					col := blue
					if isXOnBlockFrequency(x, info.DebugImageBlockFreq) {
						col = yellow
					} else {
						// continue
					}
					if !info.DebugImageBackgroundOnly {
						for y1 := max(0, y-v1); y1 < y; y1++ {
							info.DebugImage.Set(x, y1, col)
						}
					}
				}
			}
			if oldRow != nil {
				vdiff := float64(col.Difference(oldRow[x]))
				// zlog.Info("SetVDIFF:", y)
				info.setDiff(y, vdiff, info.vCounts)
				h1 := info.setHorEdge(x, vdiff, &info.vCounts[y], info.Size.W)
				if info.DebugImage != nil && h1 != 0 {
					// zlog.Info("Set hcounts perp:", x, vdiff, h1, isYOnBlockFrequency(y, info.DebugImageBlockFreq))
					col := magenta
					if isYOnBlockFrequency(y, info.DebugImageBlockFreq) {
						col = orange
					} else {
						// continue
					}
					if !info.DebugImageBackgroundOnly {
						for x1 := max(0, x-h1); x1 < x; x1++ {
							info.DebugImage.Set(x1, y, col)
						}
					}
				}
			}
			oldCol = col
		}
		oldRow = slices.Clone(row)
	}

}

func (info *ImageInfo) AnalyzeSquares(img image.Image, squareSize int, minArea, minBlockAmount, minBlockBetter float64) (debugImage image.Image, areaCoverage, amount, better float64, freq zgeo.IRect, blocky bool) {
	type squares struct {
		freq   zgeo.IRect
		amount float64
		better float64
	}
	s := zgeo.RectFromGoRect(img.Bounds()).Size.ISize()
	sx := s.W % int(squareSize) / 2
	sy := s.H % int(squareSize) / 2
	sx = zmath.RoundToMod(sx, blockMax)
	sy = zmath.RoundToMod(sy, blockMax)
	dbImage := image.NewRGBA(img.Bounds())
	var squareTotal, squareCount float64
	var amountAdd, betterAdd float64
	var freqs []zgeo.IRect
	yc := 0
	for y := sy; y <= s.H-squareSize; y += squareSize {
		// zlog.Info("Y:", yc)
		xc := 0
		for x := sx; x <= s.W-squareSize; x += squareSize {
			squareTotal++
			info.LimitFrame = zgeo.RectFromXYWH(float64(x), float64(y), float64(squareSize), float64(squareSize)).IRect()
			info.DebugImage = nil
			// now := time.Now()
			info.Analyze(img)
			// next := time.Since(now)
			a := info.SimpleAnalytics(false)
			// zlog.Info("Time:", next, time.Since(now))
			info.DebugImage = dbImage
			if a.BlockFrequency.Size.W != 0 && a.BlockBetter > minBlockBetter && a.BlockAmount > minBlockAmount {
				// zlog.Info("A:", xc, yc, a.BlockAmount, a.BlockBetter, a.BlockFrequency)
				info.DebugImageBackgroundOnly = false
				info.DebugImageBlockFreq = a.BlockFrequency
				amountAdd += a.BlockAmount
				betterAdd += a.BlockBetter
				freqs = append(freqs, a.BlockFrequency)
				squareCount++
			} else {
				// zlog.Info("B: [", xc, yc, "]", a.BlockAmount, a.BlockBetter, a.BlockFrequency)
				info.DebugImageBlockFreq = zgeo.IRect{}
				info.DebugImageBackgroundOnly = true
			}
			info.Analyze(img)
			xc++
		}
		yc++
	}
	sorted, _ := zslice.SortByFrequency(freqs)
	if len(sorted) != 0 {
		freq = sorted[0]
	}
	area := squareCount / squareTotal
	amount = amountAdd / squareCount
	better = betterAdd / squareCount
	blocky = amount > minBlockAmount && better > minBlockBetter && area > minArea
	// zlog.Info("Squares:", amount, better, area, blocky)
	return dbImage, area, amount, better, freq, blocky
}

func (info *ImageInfo) BlurAmount() zgeo.Pos {
	var w, h int
	sum := info.Size.W * info.Size.H
	for _, c := range info.hCounts {
		// zlog.Info("HBlur:", len, count, i)
		w += c.blur
	}
	for _, c := range info.vCounts {
		h += c.blur
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
	for _, c := range info.hCounts {
		w += c.flats
	}
	for _, c := range info.vCounts {
		h += c.flats
	}
	flat := zgeo.PosD(float64(w)/float64(sum), float64(h)/float64(sum))
	return flat
}

func (info *ImageInfo) EdgePointsAmount() zgeo.Pos {
	sum := info.Size.W * info.Size.H
	var w, h int
	for _, c := range info.hCounts {
		w += c.edgePoints
	}
	for _, c := range info.vCounts {
		h += c.edgePoints
	}
	edges := zgeo.PosD(float64(h)/float64(sum), float64(w)/float64(sum))
	return edges
}

func (info *ImageInfo) BlockFrequency(print bool, frame zgeo.IRect) (freq zgeo.IRect, amount, better zgeo.Size) {
	if frame.Size.W == 0 {
		frame.Size = info.Size
	}
	fX, oX, amountX, betterX := getBlockFrequencyAndOffset(print, info.hCounts, frame)
	var yf zgeo.IRect
	yf.Pos.X = frame.Pos.Y
	yf.Pos.Y = frame.Pos.X
	yf.Size.W = frame.Size.H
	yf.Size.H = frame.Size.W
	fY, oY, amountY, betterY := getBlockFrequencyAndOffset(print, info.vCounts, yf)
	if fX == 0 || fY == 0 {
		fX = 0
		fY = 0
		amountX = 0
		amountY = 0
	}
	// zlog.Info("better:", frame, betterX, betterY)
	return zgeo.IRectForXYWH(oX, oY, fX, fY), zgeo.SizeD(amountX, amountY), zgeo.SizeD(betterX, betterY)
}

func (info *ImageInfo) SimpleAnalytics(print bool) SimpleAnalytics {
	freq, amount, better := info.BlockFrequency(print, info.LimitFrame)
	// zlog.Info("Simple:", info.LimitFrame, area)
	return SimpleAnalytics{
		Size:           info.Size,
		BlurAmount:     info.BlurAmount().Average(),
		FlatsAmount:    info.FlatAmount().Average(),
		EdgesAmount:    info.EdgePointsAmount().Average(),
		BlockFrequency: freq,
		BlockAmount:    amount.Average(),
		BlockBetter:    better.Average(),
	}
}
