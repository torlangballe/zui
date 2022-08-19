//go:build zui && !js

package zvideo

// Check out: https://github.com/pion/mediadevices

import "github.com/torlangballe/zutil/zgeo"

type baseVideoView struct {
}

func (v *VideoView) Init(view zview.View, minSize zgeo.Size) {}
func (v *VideoView) CreateStream(withAudio bool)             {}
