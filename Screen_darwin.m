// +build !js,zui

#import <AppKit/AppKit.h>

// https://github.com/jakehilborn/displayplacer/blob/master/displayplacer.c

struct ScreenInfo {
    CGRect frame, visibleFrame;
    int scale;
    int ismain;
    long sid;
};

int GetAllScreens(struct ScreenInfo *sis, int max) {    
    NSArray *screens = [NSScreen screens];
    int i = 0;
    for (NSScreen *s in screens) {
        sis[i].sid = [[ s.deviceDescription objectForKey: @"NSScreenNumber"] longValue]; 
        // NSLog(@"screen id: %ld %f %f", sis[i].sid, s.frame.origin.x, s.frame.origin.y);
		sis[i].frame = s.frame;
		sis[i].visibleFrame = s.visibleFrame;
		sis[i].scale = (int)s.backingScaleFactor;
        sis[i].ismain = (i == 0) ? 1 : 0;
        i++;
        if (i >= max) {
            break;
        }
	}
    [screens release];
	return i;
}

void SetMainScreenResolutionWithinWidths(long minw, long minh, long maxw, long maxh) {
    // https://developer.apple.com/library/archive/documentation/GraphicsImaging/Conceptual/QuartzDisplayServicesConceptual/Articles/DisplayModes.html#//apple_ref/doc/uid/TP40004234-SW1
    // https://developer.apple.com/documentation/coregraphics/1456259-cgdisplaycapture?language=objc
    const int MAX = 100;
    CGDirectDisplayID displays[MAX];
    uint32_t numDisplays;
 
    CGGetOnlineDisplayList(MAX, displays, &numDisplays); 
    CGDirectDisplayID mainID = CGMainDisplayID();
 
    for (int i = 0; i < numDisplays; i++) // 2
    {
        long             bestWidth = 0, bestHeight = 0;
        CGDisplayModeRef bestMode;

        if (displays[i] != mainID) {
            continue;
        }
        CFIndex          count;
        CFArrayRef       modeList;

        modeList = CGDisplayCopyAllDisplayModes (displays[i], NULL); 
        count = CFArrayGetCount (modeList);
    
        for (int mi = 0; mi < count; mi++) 
        {
            CGDisplayModeRef mode = (CGDisplayModeRef)CFArrayGetValueAtIndex(modeList, mi);
            long width = CGDisplayModeGetWidth(mode);
            long height = CGDisplayModeGetHeight(mode);
            NSLog(@"SetRez mode: %ld x %ld", width, height);
 
            if (width >= minw && width <= maxw && height >= minh && height <= maxh && (bestWidth == 0.0 || bestWidth > width)) {
                bestWidth = width;
                bestHeight = height;
                bestMode = (CGDisplayModeRef)CFArrayGetValueAtIndex(modeList, mi);
            }
        }
        if (bestWidth != 0) {
            NSLog(@"SetMode: %ldx%ld", bestWidth, bestHeight);
            CGDisplayConfigRef config;
            CGError err = CGBeginDisplayConfiguration(&config);
            CGConfigureDisplayWithDisplayMode(config, mainID, bestMode, NULL);
            err = CGCompleteDisplayConfiguration(config, kCGConfigurePermanently);
            // CGDisplaySetDisplayMode(mainID, bestMode, NULL); 
        }
        CFRelease(modeList);
        break;
    }
}

