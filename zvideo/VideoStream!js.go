//go:build !js

package zvideo

import (
	"image"

	"github.com/pion/mediadevices"
	_ "github.com/pion/mediadevices/pkg/driver/camera"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
	"github.com/torlangballe/zutil/zlog"
)

type VideoStream struct {
	stream      mediadevices.MediaStream
	videoTrack  *mediadevices.VideoTrack
	videoReader video.Reader
}

func GetStream() (*VideoStream, error) {
	vs := &VideoStream{}
	var err error
	vs.stream, err = mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
		Video: func(constraint *mediadevices.MediaTrackConstraints) {
			// Query for ideal resolutions
			constraint.Width = prop.Int(1920)
			constraint.Height = prop.Int(1080)
		},
	})
	if err != nil {
		return nil, err
	}
	track := vs.stream.GetVideoTracks()[0] // Since track can represent audio as well, we need to cast it to *mediadevices.VideoTrack to get video specific functionalities
	zlog.Info("Track:", track.ID())
	vs.videoTrack = track.(*mediadevices.VideoTrack)
	vs.videoReader = vs.videoTrack.NewReader(false)
	return vs, nil
}

func (vs *VideoStream) Close() {
	vs.videoTrack.Close()
}

func (vs *VideoStream) FrameImage() (frame image.Image, release func(), err error) {
	frame, release, err = vs.videoReader.Read()
	return frame, release, err
}
