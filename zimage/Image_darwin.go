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
// void ConvertARGBToRGBAOpaque(int w, int h, int stride, unsigned char *img);
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

func createColorspace() C.CGColorSpaceRef {
	return C.CGColorSpaceCreateWithName(C.kCGColorSpaceSRGB)
}

func createBitmapContext(width int, height int, data *C.uint32_t, bytesPerRow int) C.CGContextRef {
	colorSpace := createColorspace()
	if colorSpace == 0 {
		return 0
	}
	defer C.CGColorSpaceRelease(colorSpace)

	return C.CGBitmapContextCreate(unsafe.Pointer(data),
		C.size_t(width),
		C.size_t(height),
		8, // bits per component
		C.size_t(bytesPerRow),
		colorSpace,
		C.kCGImageAlphaNoneSkipFirst)
}

func CGImageToGoImage(imageRef unsafe.Pointer, inRect zgeo.Rect, scale float64) (image.Image, error) {
	cgimage := C.CGImageRef(imageRef)

	iw := int(C.CGImageGetWidth(cgimage))
	ih := int(C.CGImageGetHeight(cgimage))
	if inRect.IsNull() {
		inRect = zgeo.RectFromWH(float64(iw), float64(ih))
	}
	osize := zgeo.SizeI(iw, ih).TimesD(scale)
	inRect.Pos.MultiplyD(scale)
	sw := int(inRect.Size.W * scale)
	sh := int(inRect.Size.H * scale)

	// zlog.Info("CGImageToGoImage:", inRect, iw, ih)
	img := image.NewNRGBA(image.Rect(0, 0, sw, sh))
	if img == nil {
		return nil, zlog.Error("NewRGBA returned nil", sw, sh)
	}
	ctx := createBitmapContext(sw, sh, (*C.uint32_t)(unsafe.Pointer(&img.Pix[0])), img.Stride)
	diff := float64(ih - int(inRect.Max().Y))
	x := C.CGFloat(-inRect.Pos.X)
	y := -C.CGFloat(diff) // origo is at the bottom, so we subtract difference between full snap size and inset Max().Y
	cgDrawRect := C.CGRectMake(x, y, C.CGFloat(osize.W), C.CGFloat(osize.H))
	C.CGContextDrawImage(ctx, cgDrawRect, cgimage)
	C.ConvertARGBToRGBAOpaque(C.int(sw), C.int(sh), C.int(img.Stride), (*C.uchar)(unsafe.Pointer(&img.Pix[0])))
	C.CGContextRelease(ctx)

	return img, nil
}
