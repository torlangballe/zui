// This file sets up a few minimum functions to run an app nativly on macs.
// TODO: Use catalyst and make it work in iOS too?

// +build !js,zui,catalyst

#include <Cocoa/Cocoa.h>

void *SharedApplication(void) {
	return [NSApplication sharedApplication];
}

// Run starts the app in the background, but on main thread
void Run() {
    static bool running = false;
    if (running) {
       return;
    }
    running = true;
    @autoreleasepool {
        NSApplication* a = (NSApplication*)SharedApplication();
        [a setActivationPolicy:NSApplicationActivationPolicyRegular];
        [a activateIgnoringOtherApps : YES];
        dispatch_async(dispatch_get_main_queue(), ^{
            NSLog(@"Dispatch Run1\n");
            [a run];
            NSLog(@"Dispatch run() done\n");
       });
       NSLog(@"Dispatch Run exited\n");
	}
}


