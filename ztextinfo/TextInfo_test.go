package ztextinfo

import (
	"testing"

	"github.com/torlangballe/zutil/zgeo"
)

func TestWrapping(t *testing.T) {
	ti := New()
	ti.Font = zgeo.FontDefault(0)
	ti.Text = "https://vs-bks400-11.tv.siminn.is/bpk-token/1ad@l3bwruaehygsckizdyjcddmdjrlpqnefldo134da/bpk-tv/NRK2/dash-wv2/index.mpd?bkm-query"
	ti.Alignment = zgeo.Left
	ti.Rect = zgeo.RectFromWH(600, 600)
	s, lines, _ := ti.GetBounds()
	if len(lines) != 3 || s.H < 40 {
		t.Error("Not 3/40:", s, len(lines))
	}
}
