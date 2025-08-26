//go:build !js

package zvideo

import (
	"image"

	"github.com/pion/mediadevices"
	_ "github.com/pion/mediadevices/pkg/driver/camera"
	_ "github.com/pion/mediadevices/pkg/frame"
	"github.com/pion/mediadevices/pkg/io/video"
	"github.com/pion/mediadevices/pkg/prop"
	"github.com/torlangballe/zutil/zgeo"
)

type VideoStream struct {
	stream      mediadevices.MediaStream
	videoTrack  *mediadevices.VideoTrack
	videoReader video.Reader
}

func GetStream(sizeConstraint zgeo.Size) (*VideoStream, error) {
	vs := &VideoStream{}
	var err error
	vs.stream, err = mediadevices.GetUserMedia(mediadevices.MediaStreamConstraints{
		Video: func(constraint *mediadevices.MediaTrackConstraints) {
			// constraint.FrameFormat = prop.FrameFormatOneOf{
			// 	frame.FormatYUYV,
			// }
			// constraint.DiscardFramesOlderThan = time.Millisecond * 200
			if !sizeConstraint.IsNull() {
				constraint.Width = prop.Int(sizeConstraint.W)
				constraint.Height = prop.Int(sizeConstraint.H)
			}
		},
		// Codec: codecSelector,
	})
	if err != nil {
		return nil, err
	}
	// for _, t := range vs.stream.GetVideoTracks() {
	// 	zlog.Info("Track:", t.Kind(), t.ID())
	// }
	// vs.videoReader = vs.videoTrack.NewReader(true)
	return vs, nil
}

func (vs *VideoStream) Close() {
	vs.videoTrack.Close()
}

func (vs *VideoStream) FrameImage() (frame image.Image, release func(), err error) {
	track := vs.stream.GetVideoTracks()[0] // Since track can represent audio as well, we need to cast it to *mediadevices.VideoTrack to get video specific functionalities
	vtrack := track.(*mediadevices.VideoTrack)
	reader := vtrack.NewReader(false)
	frame, release, err = reader.Read()
	vtrack.Close()
	return frame, release, err
}
