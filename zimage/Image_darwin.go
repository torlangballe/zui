package zimage

// #cgo LDFLAGS: -framework CoreVideo
// #cgo LDFLAGS: -framework Foundation
// #cgo LDFLAGS: -framework AppKit
// #cgo LDFLAGS: -framework CoreGraphics
// #cgo LDFLAGS: -framework CoreImage
// #include <CoreGraphics/CoreGraphics.h>
// struct Image {
// unsigned char  *JPEGData;
// int             JPEGLength;
// unsigned char  *JPEGDataScaled;
// int             JPEGLengthScaled;
// char           *error;
// };
// struct Image ImageNRGBABytesToJPEGDataNative(unsigned char *img, int w, int h, int sw, int sh, float quality);
import "C"

import (
	"errors"
	"image"
	"time"
	"unsafe"

	"github.com/torlangballe/zutil/zgeo"
	"github.com/torlangballe/zutil/zlog"
)

func GoNRGBAImageToJPEGDataNative(img *image.NRGBA, scaleTo zgeo.Size, qualityPercent int) (fullSize, scaledSize []byte, err error) {
	start := time.Now()
	s := GoImageZSize(img)
	cimage := unsafe.Pointer(&img.Pix[0])
	qual := float32(qualityPercent) / 100
	ci := C.ImageNRGBABytesToJPEGDataNative((*C.uchar)(cimage), C.int(s.W), C.int(s.H), C.int(scaleTo.W), C.int(scaleTo.H), C.float(qual))
	serr := C.GoString(ci.error)
	if serr != "" {
		return nil, nil, errors.New(serr)
	}
	jpegData := unsafe.Slice((*byte)(ci.JPEGData), ci.JPEGLength)
	jpegDataScaled := unsafe.Slice((*byte)(ci.JPEGDataScaled), ci.JPEGLengthScaled)
	zlog.Info("GoNRGBAImageToJPEGDataNative:", len(jpegData), len(jpegDataScaled), time.Since(start))
	return jpegData, jpegDataScaled, nil
}
