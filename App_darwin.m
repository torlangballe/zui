// +build !js,zui,catalyst

#include <Cocoa/Cocoa.h>

void *SharedApplication(void) {
	return [NSApplication sharedApplication];
}

void Run(void *app) {
    @autoreleasepool {
        NSApplication* a = (NSApplication*)app;
        [a setActivationPolicy:NSApplicationActivationPolicyRegular];
        [a activateIgnoringOtherApps : YES];
        [a run];
	}
}


