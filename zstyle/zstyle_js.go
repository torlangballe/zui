package zstyle

import (
	"fmt"
	"strings"

	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zkeyvalue"
	"github.com/torlangballe/zutil/zstr"
)

func init() {
	Dark, _ = zkeyvalue.DefaultStore.GetBool("zstyle.DarkMode", true)
}

func SetStyling(v zview.View, style Styling) {
	nv := v.Native()
	if style.DropShadow.Color.Valid {
		nv.SetDropShadow(style.DropShadow)
	}
	if style.BGColor.Valid {
		nv.SetBGColor(style.BGColor)
	}
	if style.Corner != -1 {
		nv.SetCorner(style.Corner)
	}
	if style.StrokeColor.Valid {
		// zlog.Info("SetStyling:", nv.Hierarchy(), style.StrokeWidth, style.StrokeColor)
		nv.SetStroke(style.StrokeWidth, style.StrokeColor, style.StrokeIsInset.IsTrue())
	}
	if style.OutlineColor.Valid {
		nv.SetOutline(style.OutlineWidth, style.OutlineColor, style.OutlineOffset)
	}
	if style.FGColor.Valid {
		nv.SetColor(style.FGColor)
	}
	if style.Font.Name != "" {
		nv.SetFont(&style.Font)
	}
	if !style.Margin.IsUndef() {
		m, _ := v.(zview.Marginalizer)
		if m != nil {
			m.SetMargin(style.Margin)
		}
	}
}

func ChangedImagePathBasedOnDarkTags(path string) string {
	if !strings.Contains(path, "{") {
		return path
	}
	path = zstr.ReplaceInSquigglyBrackets(path, func(s string) string {
		var pre string
		if zstr.HasSuffix(s, "dark", &pre) {
			if Dark {
				return pre + "4dark"
			}
			return pre + "4light"
		}
		if zstr.HasSuffix(s, "dark?", &pre) {
			if Dark {
				return pre + "4dark"
			} else {
				return ""
			}
		}
		return "<?>"
	})
	return path
}

func SetStyleAs9Slice(v *zview.NativeView, imagePath string, capInset zgeo.Size) {
	path := ChangedImagePathBasedOnDarkTags(imagePath)
	v.SetJSStyle("box-sizing", "border-box")
	v.SetJSStyle("border-style", "solid")
	v.SetJSStyle("border-width", "0")                                                    // need to set this to 0 and border-image-source to not have children offset
	v.SetJSStyle("border-image-width", fmt.Sprintf("%gpx %gpx", capInset.H, capInset.W)) // top/bottom, left/right
	v.SetJSStyle("border-image-source", fmt.Sprintf(`url("%s")`, path))
	v.SetJSStyle("border-image-slice", fmt.Sprintf("%g %g fill", capInset.H, capInset.W))
	// v.SetJSStyle("border-image-repeat", "round") //  maybe this is better sometimes?
	// v.SetJSStyle("image-rendering", "auto") // and this?
}
