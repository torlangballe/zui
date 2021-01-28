// +build !js

package zui

import "github.com/torlangballe/zutil/zgeo"

type baseVideoView struct {
}

func (v *VideoView) Init(view View, minSize zgeo.Size) {}
func (v *VideoView) CreateStream(withAudio bool)       {}
