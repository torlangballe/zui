//go:build !js,zui,catalyst

#include <Cocoa/Cocoa.h>

void *NewWindow(int x, int y, int width, int height)
{
    NSWindow* window = [[NSWindow alloc] initWithContentRect:NSMakeRect(x, y, width, height)
                                              styleMask: NSWindowStyleMaskBorderless  // NSWindowStyleMaskTitled
                                              backing:NSBackingStoreBuffered 
                                              defer:NO];

    [window setLevel:NSMainMenuWindowLevel+1];
    [window setOpaque:YES];
    return window;
}

void MakeKeyAndOrderFront(void *self) {
    NSWindow *window = self;
    [window makeKeyAndOrderFront:nil];
}

void SetTitle(void *self, char *title) {
    NSWindow *window = self;
    NSString *nsTitle =  [NSString stringWithUTF8String:title];

    [window setTitle:nsTitle];
    free(title);
}

void AddView(void *win, void *view) {
    [(NSWindow *)win setContentView: view];
}

