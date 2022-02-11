// This file sets up a few minimum functions to run an app nativly on macs.
// TODO: Use catalyst and make it work in iOS too?

// +build !js,zui,catalyst

#include <Cocoa/Cocoa.h>

void *SharedApplication(void) {
	return [NSApplication sharedApplication];
}

// Run starts the app and runs until quit. This function doesn't exit. 
void Run() {
    @autoreleasepool {
        NSApplication* a = (NSApplication*)SharedApplication();
        [a setActivationPolicy:NSApplicationActivationPolicyRegular];
        [a activateIgnoringOtherApps : YES];
        [a run];
	}
}


