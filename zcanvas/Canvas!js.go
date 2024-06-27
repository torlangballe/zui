//go:build !js

package zcanvas

import (
	"fmt"
	"image"
	"image/color"
	"runtime"
	"sync"

	"github.com/fogleman/gg"
	"github.com/torlangballe/zui/zimage"
	"github.com/torlangballe/zui/zstyle"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

type canvasNative struct {
	context *gg.Context
	setFont zgeo.Font
}

func New() *Canvas {
	return &Canvas{}
}

func (c *Canvas) String() string {
	if c.context == nil {
		return "context==nil"
	}
	return fmt.Sprintf("context:%dx%d", c.context.Width(), c.context.Height())
}

func CanvasFromGoImage(img image.Image) *Canvas {
	zlog.Assert(img != nil)
	c := &Canvas{}
	c.context = gg.NewContextForImage(img)
	c.size = zimage.GoImageZSize(img)
	zlog.Assert(c.context != nil)
	return c
}

func (c *Canvas) SetSize(size zgeo.Size) {
	c.size = size
	c.context = gg.NewContext(int(size.W), int(size.H))
	c.Clear()
	zlog.Assert(c.context != nil)
}

func (c *Canvas) SetRect(rect zgeo.Rect) {
}

func (c *Canvas) SetColor(color zgeo.Color) {
	// if opacity != -1 {
	// 	color = color.WithOpacity(opacity)
	// }
	c.context.SetColor(color)

	//c.context.SetRGBA(float64(color.Colors.R), float64(color.Colors.G), float64(color.Colors.B), float64(color.Colors.A))
}

func (c *Canvas) FillPath(path *zgeo.Path) {
	c.setPath(path)
	c.context.Fill()
}

func (c *Canvas) FillPathEO(path *zgeo.Path) {
	zlog.Fatal("Not implemented")
}

var fontMutex sync.Mutex

func (c *Canvas) SetFont(font *zgeo.Font, matrix *zgeo.Matrix) error {
	zlog.Assert(c.context != nil, zlog.Pointer(c))
	if *font == c.setFont {
		return nil
	}
	// fmt.Printf("CANVAS SETFONT: %p %+v %+v %s\n", c, *font, c.setFont, zlog.GetCallingStackString())
	c.setFont = *font
	var err error
	name := font.Name
	if font.Style&zgeo.FontStyleBold != 0 {
		name += " Bold"
	}
	if font.Style&zgeo.FontStyleItalic != 0 {
		name += " Italic"
	}
	var paths = []string{"Fonts/"}
	if runtime.GOOS == "darwin" {
		paths = append(paths, "/System/Library/Fonts/")
	}
	for _, path := range paths {
		for _, ext := range []string{".ttf", ".ttc"} {
			p := path + name + ext
			zlog.Assert(c.context != nil)
			fontMutex.Lock()
			err = c.context.LoadFontFace(p, font.Size)
			fontMutex.Unlock()
			if err == nil {
				return nil
			}
		}
	}
	return zlog.Error("couldn't load font", font.Name, paths)
}

func (c *Canvas) SetMatrix(matrix zgeo.Matrix) {
}

func (c *Canvas) Transform(matrix zgeo.Matrix) {
}

func (c *Canvas) ClipPath(path *zgeo.Path, eofill bool) {
	c.setPath(path)
	c.context.Clip()
}

func (c *Canvas) GetClipRect() zgeo.Rect {
	return zgeo.Rect{}
}

func (c *Canvas) StrokePath(path *zgeo.Path, width float64, ltype zgeo.PathLineType) {
	c.setPath(path)
	c.setLineType(ltype)
	c.context.SetLineWidth(width)
	c.context.Stroke() // hmmm, what is alternative?
}

func (c *Canvas) DrawPath(path *zgeo.Path, strokeColor zgeo.Color, width float64, ltype zgeo.PathLineType, eofill bool) {
	c.setPath(path)
	c.context.FillPreserve()
	c.setLineType(ltype)
	c.context.SetLineWidth(width)
	c.context.Stroke()
}

func (c *Canvas) PushState() {
	c.context.Push()
}

func (c *Canvas) PopState() {
	c.context.Pop()
}

func (c *Canvas) Clear() {
	c.SetColor(zgeo.ColorClear)
	c.context.Clear()
}

func (c *Canvas) SetDropShadow(d zstyle.DropShadow) {}
func (c *Canvas) ClearDropShadow()                  {}

func (c *Canvas) DrawGradient(path *zgeo.Path, colors []zgeo.Color, pos1 zgeo.Pos, pos2 zgeo.Pos, locations []float64) {
	c.PushState()
	gradient := gg.NewLinearGradient(pos1.X, pos1.Y, pos2.X, pos2.Y)
	if gradient != nil {
		if len(locations) == 0 {
			locations = canvasCreateGradientLocations(len(colors))
		}
		for i := range colors {
			gradient.AddColorStop(locations[i], colors[i])
		}
		c.context.SetFillStyle(gradient)
		c.FillPath(path)
		c.PopState()
	}
}

func (c *Canvas) DrawRadialGradient(path *zgeo.Path, colors []zgeo.Color, center zgeo.Pos, radius float64, endCenter *zgeo.Pos, startRadius float64, locations []float32) {
	c.PushState()
	if path != nil {
		//            self.ClipPath(path!)
	}
	c.PopState()
}

func (c *Canvas) setPath(path *zgeo.Path) {
	path.ForEachPart(func(part zgeo.PathNode) {
		// zlog.Info("setPath", part.Type)
		switch part.Type {
		case zgeo.PathMove:
			//  zlog.Info("moveTo", part.Points[0].X, part.Points[0].Y)
			c.context.MoveTo(part.Points[0].X, part.Points[0].Y)
		case zgeo.PathLine:
			// zlog.Info("lineTo", part.Points[0].X, part.Points[0].Y)
			c.context.LineTo(part.Points[0].X, part.Points[0].Y)
		case zgeo.PathClose:
			// zlog.Info("pathClose")
			c.context.ClosePath()
		case zgeo.PathQuadCurve:
			c.context.QuadraticTo(part.Points[0].X, part.Points[0].Y, part.Points[1].X, part.Points[1].Y)
			// zlog.Info("quadCurve")
			break
		case zgeo.PathCurve:
			c.context.CubicTo(part.Points[0].X, part.Points[0].Y, part.Points[1].X, part.Points[1].Y, part.Points[2].X, part.Points[2].Y)
			// zlog.Info("curveTo", part.Points[0].X, part.Points[0].Y, part.Points[1].X, part.Points[1].Y, part.Points[2].X, part.Points[2].Y)
			break
		}
	})

}

func (c *Canvas) setMatrix(m zgeo.Matrix) {
	// c.currentMatrix = m
}

func (c *Canvas) setLineType(ltype zgeo.PathLineType) {

}

func (c *Canvas) DrawTextInPos(pos zgeo.Pos, text string, strokeWidth float64) {
	// fmt.Printf("CANVAS draw text: %v %p\n", c, c)
	fontMutex.Lock()
	c.context.DrawString(text, pos.X, pos.Y)
	fontMutex.Unlock()
}

func (c *Canvas) MeasureText(text string, font *zgeo.Font) zgeo.Size {
	c.SetFont(font, nil)
	fontMutex.Lock()
	w, h := c.context.MeasureString(text)
	fontMutex.Unlock()
	return zgeo.SizeD(w, h)
}

func (c *Canvas) setAlphaMask(opacity float32) {
	var b image.Rectangle
	b.Max.X = c.context.Width()
	b.Max.Y = c.context.Height()
	a := image.NewAlpha(b)
	var alpha color.Alpha
	alpha.A = uint8(opacity * 255)
	for y := 0; y < b.Max.Y; y++ {
		for x := 0; x < b.Max.X; x++ {
			a.Set(x, y, alpha)
		}
	}
	c.context.SetMask(a)
}

func (c *Canvas) drawPlainImage(image *zimage.Image, useDownsampleCache bool, destRect zgeo.Rect, opacity float32, sourceRect zgeo.Rect) bool {
	if opacity != 1 {
		// TODO: Cache this alpha mask, made each time now...
		c.setAlphaMask(opacity)
	}
	c.context.DrawImage(image.GoImage, int(destRect.Pos.X), int(destRect.Pos.Y))
	if opacity != 1 {
		c.context.ResetClip()
	}
	return true
}

func (c *Canvas) GoImage(cut zgeo.Rect) image.Image {
	i := c.context.Image()
	zlog.Assert(cut.IsNull())
	return i
}

func (c *Canvas) SetFillRuleEvenOdd(eo bool) {
	if eo {
		c.context.SetFillRule(gg.FillRuleEvenOdd)
	} else {
		c.context.SetFillRule(gg.FillRuleWinding)
	}
}

func (c *Canvas) ZImage(ensureCopy bool, got func(image *zimage.Image)) {
	ni := c.GoImage(zgeo.Rect{})
	if ensureCopy {
		ni = zimage.CloneGoImage(ni)
	}
	zimage.FromGo(ni, got)
}
