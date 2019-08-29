package zgo

import (
	"strconv"

	"github.com/torlangballe/zutil/ustr"
)

//  Created by Tor Langballe on /20/10/15.

type Image struct {
	imageBase
	scale   int
	path    string
	loading bool
}

type ImageOwner interface {
	GetImage() *Image
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

func (i *Image) CapInsetsCorner(c Size) *Image {
	r := RectFromMinMax(c.Pos(), c.Pos().Negative())
	return i.CapInsets(r)
}

func imageGetScaleFromPath(path string) int {
	var n string
	_, _, m, _ := FilePathGetPathParts(path)
	if ustr.SplitN(m, "@", &n, &m) {
		if ustr.HasSuffix(m, "x", &m) {
			scale, err := strconv.ParseInt(m, 10, 32)
			if err == nil && scale >= 1 && scale <= 3 {
				return int(scale)
			}
		}
	}
	return 1
}
