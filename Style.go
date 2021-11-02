package zui

import (
	"path"

	"github.com/torlangballe/zutil/zfile"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zstr"
)

var (
	StyleDark           bool
	StyleDefaultFGColor = StyleGrayF(0.2, 0.8)
	StyleDefaultBGColor = StyleGrayF(0.8, 0.2)
)

func StyleCol(l, d zgeo.Color) zgeo.Color {
	if StyleDark {
		// zlog.Error(nil, zlog.StackAdjust(1), "StyleColDark:", d)
		return d
	}
	// zlog.Error(nil, zlog.StackAdjust(1), "StyleColLight:", l)
	return l
}

func StyleColF(l, d zgeo.Color) func() zgeo.Color {
	return func() zgeo.Color {
		return StyleCol(l, d)
	}
}

func StyleGray(l, d float32) zgeo.Color {
	return StyleCol(zgeo.ColorNewGray(l, 1), zgeo.ColorNewGray(d, 1))
}

func StyleGrayF(l, d float32) func() zgeo.Color {
	return func() zgeo.Color {
		return StyleGray(l, d)
	}
}

func StyleImagePath(spath string) string {
	if !StyleDark {
		return spath
	}
	dir, _, stub, ext := zfile.Split(spath)
	size := ""
	if zstr.HasSuffix(stub, "@2x", &stub) {
		size = "@2x"
	}
	return path.Join(dir, stub+"_dark"+size+ext)
}
