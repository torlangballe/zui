//go:build !js

package zaudio

import (
	"io"
)

type audioNative struct {
}

func MP4DurationSecs(reader io.ReadSeeker, size int64) (float64, error) {
	return float64(size*8) / float64(128000), nil
	// meta, err := audiotag.ReadFrom(reader)
	// if err != nil {
	// 	zlog.Error(err, "read audio tag")
	// 	return 0, err
	// }
	// return float64(meta.Duration()), nil

	// mp4, err := mp4.OpenFromReader(reader, size)
	// zlog.Info("DUR:", mp4.Moov.Mvhd.Duration, mp4.Moov.Mvhd.Timescale)
	// durationSec := float64(mp4.Moov.Mvhd.Duration) / float64(mp4.Moov.Mvhd.Timescale)
	// return durationSec, nil
}
