#import <Foundation/Foundation.h>
#import <CoreVideo/CoreVideo.h>
#import <CoreImage/CoreImage.h>
#import <CoreImage/CIFilter.h>
#import <AppKit/AppKit.h>
#import <UniformTypeIdentifiers/UniformTypeIdentifiers.h>

struct Image {
    unsigned char *JPEGData;
    int            JPEGLength;
    unsigned char *JPEGDataScaled;
    int            JPEGLengthScaled;
    char          *error;
};

unsigned char *imageToJPEG(CGImageRef cgImage, float quality, int *len) {
    NSBitmapImageRep *imgRep = [[NSBitmapImageRep alloc] initWithCGImage:cgImage];
   // NSBitmapImageRep *imgRep = [[image representations] objectAtIndex: 0];
    // NSLog(@"imageToJPEG2 %d len:%d\n", imgRep != NULL, [[image representations] count]);
    NSDictionary *properties = [NSDictionary dictionaryWithObject:[NSDecimalNumber numberWithFloat:quality] forKey:NSImageCompressionFactor];
    NSData *data = [imgRep representationUsingType: NSBitmapImageFileTypeJPEG properties: properties];
    *len = [data length];
    unsigned char *buffer = (unsigned char *)malloc(*len+1);
    [data getBytes: buffer range:NSMakeRange(0, (long)*len)];

    NSURL *nsurl = [NSURL URLWithString:@"file:///Users/tor/out.jpeg"];
    [data writeToURL:nsurl atomically: true];


    return buffer;
}

CGImageRef scaleImage(CGImageRef cgImage, float scale) {

    CIImage *ci = [CIImage imageWithCGImage:cgImage];
    CIFilter *filter = [CIFilter filterWithName:@"CILanczosScaleTransform"];
    [filter setValue:ci forKey:@"inputImage"];
    [filter setValue:@(scale) forKey:@"inputScale"];
    [filter setValue:@1.0 forKey:@"inputAspectRatio"];
    CIImage *newImage = [filter outputImage];
    CIContext *context = [CIContext contextWithOptions:nil];
    CGImageRef out = [context createCGImage:newImage fromRect:[newImage extent]];

	NSLog(@"GoNRGBAImageToJPEGDataNative scaleImage3 %f %p %p\n", scale, newImage, out);
    return out;
}

struct Image ImageNRGBABytesToJPEGDataNative(unsigned char *imgBytes, int w, int h, int sw, int sh, float quality) {
    struct Image ri;
    ri.error = "";
    CGColorSpaceRef colorspace = CGColorSpaceCreateDeviceRGB();
    CFDataRef rgbData = CFDataCreate(NULL, (const unsigned char*)imgBytes, w*h*4);
    CGDataProviderRef provider = CGDataProviderCreateWithCFData(rgbData);
    CGImageRef imageRef = CGImageCreate(w, h, 8, 32, w * 4, colorspace, kCGBitmapByteOrderDefault, provider, NULL, true, kCGRenderingIntentDefault);
    CFRelease(rgbData);
    CGDataProviderRelease(provider);
    CGColorSpaceRelease(colorspace);
    if (w != 0) {
        CGImageRef scaledImage = scaleImage(imageRef, (float)sw/(float)w);
        if (scaledImage == NULL) {
            ri.error = "couldn't scale image";
        } else {
            ri.JPEGDataScaled = imageToJPEG(scaledImage, quality, &ri.JPEGLengthScaled);
            CGImageRelease(scaledImage);
        }
    }
    ri.JPEGData = imageToJPEG(imageRef, quality, &ri.JPEGLength);
    CGImageRelease(imageRef);
    return ri;
}

void ConvertARGBToRGBAOpaque(int w, int h, int stride, unsigned char *img) {
	for (int iy = 0; iy < h; iy++) {
        unsigned char *p = &img[iy*stride];
		for (int ix = 0; ix < w; ix++) {
			// ARGB => RGBA, and set A to 255
            p[0] = p[1];
            p[1] = p[2];
            p[2] = p[3];
            p[3] = 255;
            p += 4;
		}
	}
}

