// #import <CoreGraphics/CoreGraphics.h>
// #import <CoreFoundation/CoreFoundation.h>
#import <AppKit/AppKit.h>

struct ScreenInfo {
    CGRect bounds;
    int scale;
    int ismain;
};

int GetAllScreens(struct ScreenInfo *sis, int max) {    
    NSArray *screens = [NSScreen screens];
    int i = 0;
    for (NSScreen *s in screens) {
		sis[i].bounds = s.frame;
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