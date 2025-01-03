//go:build zui

package zaudio

import (
	"github.com/torlangballe/zui/zcontainer"
	"github.com/torlangballe/zui/zimageview"
	"github.com/torlangballe/zui/zkeyboard"
	"github.com/torlangballe/zui/zview"
	"github.com/torlangballe/zui/zwidgets"
	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

type AudioIconView struct {
	zcontainer.ContainerView

	imageView *zimageview.ImageView
	activity  *zwidgets.ActivityView
	path      string
	audio     *Audio
}

func NewAudioIconView(size zgeo.Size, path string) *AudioIconView {
	zlog.Info("NewAudioIconView")
	v := &AudioIconView{}
	v.Init(v, "audio")
	// v.SetInteractive(false)
	v.path = path
	if size.IsNull() {
		size = zgeo.SizeBoth(20)
	}
	v.SetMinSize(size)

	v.activity = zwidgets.NewActivityView(size, zgeo.ColorBlack)
	v.Add(v.activity, zgeo.Center)

	v.imageView = zimageview.NewWithCachedPath("images/zcore/speaker.png", size)
	v.imageView.DownsampleImages = true
	v.imageView.SetObjectName("image")
	v.Add(v.imageView, zgeo.Center)

	v.imageView.SetLongPressedHandler("", zkeyboard.ModifierNone, func() {
		zview.DownloadURI(v.path, "")
	})
	v.imageView.SetPressedHandler("", zkeyboard.ModifierNone, func() {
		zlog.Info("Pressed", v.path)
		if v.audio == nil {
			v.audio = AudioNew(v.path)
			v.audio.SetHandleFinished(func() {
				zlog.Info("finished")
				v.stop()
			})
			v.imageView.Show(false)
			v.activity.Start()
			v.audio.Play(func(error) {
				v.stop()
			})
		}
	})
	v.activity.SetPressedHandler("", zkeyboard.ModifierNone, func() {
		v.stop()
	})
	return v
}

func (v *AudioIconView) stop() {
	v.audio.Stop()
	v.imageView.Show(true)
	v.activity.Stop()
	v.audio = nil
}
