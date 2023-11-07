//go:build !js && zui && catalyst

#include <Cocoa/Cocoa.h>


@interface ZWindow : NSWindow
{
@public
//    ZWindow *window;
}
- (void)keyDown:(NSEvent *)theEvent;
@end

@implementation ZWindow
- (void)keyDown:(NSEvent *)theEvent {
    NSLog(@"keyDown\n");
}
- (BOOL)becomeFirstResponder {
    return YES;
}
@end

void *NewWindow(int x, int y, int width, int height)
{
    ZWindow* window = [[ZWindow alloc] initWithContentRect:NSMakeRect(x, y, width, height)
                                              styleMask: NSWindowStyleMaskBorderless  // NSWindowStyleMaskTitled
                                              backing:NSBackingStoreBuffered 
                                              defer:NO];

    [window setLevel:NSMainMenuWindowLevel];
//    [window setLevel:NSMainMenuWindowLevel+1];
    [window setOpaque:YES];
    NSLog(@"NewWindow\n");
    return window;
}

void MakeKeyAndOrderFront(void *self) {
    ZWindow *window = self;
    [window makeKeyAndOrderFront:nil];
}

void SetTitle(void *self, char *title) {
    ZWindow *window = self;
    NSString *nsTitle =  [NSString stringWithUTF8String:title];

    [window setTitle:nsTitle];
    free(title);
}

void AddView(void *win, void *view) {
    [(ZWindow *)win setContentView: view];
}

