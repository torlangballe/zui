//  Image.go
//
//  Created by Tor Langballe on /20/10/15.

package zgo

import "fmt"

type Image struct {
	size      Size    `json:"size"`
	scale     float32 `json:"scale"`
	capInsets Rect    `json:"capInsets"`
	hasAlpha  bool    `json:"hasAlpha"`
}

func ImageNamed(name string) *Image {
	return nil
}

func (i *Image) Colored(color Color, size Size) *Image {
	//	rect := NewRect(0, 0, size.W, size.H)
	return i
}

func (i *Image) Size() Size {
	return i.size
}

func (i *Image) Make9PatchImage(capInsets Rect) *Image {
	return i
}

func (i *Image) TintedWithColor(color Color) *Image {
	return i
}

func (i *Image) GetScaledInSize(size Size, proportional bool) *Image {
	var vsize = size
	if proportional {
		vsize = RectFromSize(size).Align(i.Size(), AlignmentCenter|AlignmentShrink|AlignmentScaleToFitProportionally, Size{0, 0}, Size{0, 0}).Size
	}
	width := int(vsize.W) / int(i.scale)
	height := int(vsize.H) / int(i.scale)
	fmt.Println("GetScaledInSize not made yet:", width, height)
	return nil
}

func (i *Image) GetCropped(crop Rect) *Image {
	return i
}

func (i *Image) HasAlpha() bool {
	return i.hasAlpha
}

func (i *Image) GetLeftRightFlipped() *Image {
	return i
}

func (i *Image) Normalized() *Image {
	return i
}

func (i *Image) GetRotationAdjusted(flip bool) *Image { // adjustForScreenOrientation:bool
	return i
}

func (i *Image) Rotated(deg float64, around *Pos) *Image {
	var pos = i.Size().Pos().DividedByD(2)
	if around != nil {
		pos = *around
	}
	transform := MatrixForRotatingAroundPoint(pos, deg)
	fmt.Println("Image.Rotated not made yet:", transform)
	return i
}

func (i *Image) FixedOrientation() *Image {
	return i
}

/*
 var MainCache = ImageCache()

     func DownloadFromUrl(url string, cache bool, maxSize *Size, done func(image *Image)) *URLSessionTask {
        if cache {
            return MainCache.DownloadFromUrl(url, done)
        }
        //        let start = ZTime.Now()
        req := UrlRequest.Make(UrlRequestGet, url)
        return UrlSession.Send(req, false, true, func() { (resp, data, err) in
            if err != nil {
                ZDebug.Print("Image.DownloadFromUrl error:", err!.localizedDescription, url)
                ZMainQue.async {
                    done?(nil)
                }
                return
            }
            if data == nil {
                ZDebug.Print("Image.DownloadFromUrl data=null:", url)
                ZMainQue.async {
                    done?(nil)
                }
                return
            }
            var scale:CGFloat = 1.0
            let name = req.url!.deletingPathExtension().lastPathComponent
            if name.hasSuffix("@2x") {
                scale = 2
            } else if name.hasSuffix("@3x") {
                scale = 3
            }
            if var image = Image(data:data!, scale:scale) {
                if maxSize != nil && (image.Size.w > maxSize!.w || image.Size.h > maxSize!.h) {
                    if let small = image.GetScaledInSize(maxSize!) {
//                        ZDebug.Print("Image.Download: Scaling too big image:", image.Size, "max:", maxSize!, url)
                        image = small
                    } else {
                        ZDebug.Print("Image.Download: Failing too big image not scaleable:", image.Size, "max:", maxSize!, url)
                        ZMainQue.async {
                            done?(nil)
                        }
                        return
                    }
                }
                ZMainQue.async {
                    done?(image)
                }
            } else {
                ZMainQue.async {
                    done?(nil)
                }
            }
        }
    }

    func SaveToPng(_ file:ZFileUrl)  ZError? {
      let data:ZData? = self.pngData() as ZData?
        if data != nil {
            if data!.SaveToFile(file) == nil {
                return nil
            }
        }
        return ZNewError("error storing image as png")
    }

    func SaveToJpeg(_ file:ZFileUrl, quality:float32 = 0.8)  ZError? {
        let data:ZData? = self.jpegData(compressionQuality:CGFloat(quality)) as ZData?
        if data != nil {
            if data!.SaveToFile(file) != nil {
                return nil
            }
        }
        return ZNewError("error storing image as png")
    }

    class func FromFile(_ file:ZFileUrl)  Image? {
        if file.url != nil {
            do {
                let data = try ZData(contentsOf:file.url! as URL)
                return Image(data:data as Data)
            } catch {
                return nil
            }
        }
        return nil
    }
}
*/

func MakeImageFromDrawFunction(size Size, scale float32, draw func(size Size, canvas Canvas)) *Image {
	return nil
}

func (i *Image) ForPixels(got func(pos Pos, color Color)) {
}

/*
    func ClipToCircle(fit:Size = Size(0, 0))  Image? {
        var si = Size(size)
        var s = fit

        if fit.IsNull() {
            let w = (si.w > si.h) ? si.h : si.w
            s = Size(w, w)
        } else {
            let scale = max(fit.w / float32(size.width), fit.h / float32(size.height))
            si *= scale
        }
        var ir = Rect(size:si)
        var r = ir.Align(s, align:.Center | .Shrink)
        return ZMakeImageFromDrawFunction(s) { (size, canvas) in
            let path = ZPath()
            ir.pos -= r.pos
            r.pos = Pos(0, 0)
            path.AddOval(inrect:r)
            canvas.ClipPath(path)
            canvas.DrawImage(self, destRect:ir)
        }
    }

    static func GetNamedImagesFromWildcard(_ wild:string)  [Image] {
        var images = [Image]()
        let folder = ZGetResourceFileUrl("")
        folder.Walk(wildcard:wild) { (furl, finfo) in
            if let image = Image.Named(furl.GetName()) {
                images.append(image)
            }
            return true
        }
        return images
    }
}
*/
