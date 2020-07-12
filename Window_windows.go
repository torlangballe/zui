package zui

import (
	"github.com/torlangballe/zutil/zcommand"
	"github.com/torlangballe/zutil/zfile"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

func WindowsCloseWindowWithTitle(title string) error {
	eval := `"WINDOWTITLE eq ` + title + `"`
	str, err := zcommand.RunCommand("killtask", 4, "/FI", eval)
	zlog.Info(str, err)
	return err
}

func WindowsResizeWindowWithTitle(title string, size zgeo.Size) error {
	return nil
}

func WindowGetImageForTitle(title string, crop zgeo.Rect) (*Image, error) {
	_, filepath, err := zfile.CreateTempFile("win.png")
	zlog.Assert(err == nil)

	_, err = zcommand.RunCommand("screencapture.exe", 2, filepath, title)
	if err != nil {
		return nil, err
	}
	image := ImageFromPath(filepath, nil)
	if image == nil {
		return nil, zlog.Error(nil, "image from path")
	}
	image = image.Cropped(crop, false)
	return image, nil
}