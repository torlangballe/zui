//go:build !js && zui && catalyst

#include <Cocoa/Cocoa.h>
#include "WebKit/WebKit.h"

@interface ZWKWebView : WKWebView 
<WKScriptMessageHandler> {
@public
    NSFileHandle *logFile;
}
- (void)keyDown:(NSEvent *)theEvent;
//- (void)userContentController:(WKUserContentController *)userContentController didReceiveScriptMessage:(WKScriptMessage *)message;
@end

@implementation ZWKWebView
- (void)keyDown:(NSEvent *)theEvent {
    NSLog(@"web keyDown\n");
}

- (void)userContentController:(WKUserContentController *)userContentController didReceiveScriptMessage:(WKScriptMessage *)message
{
    static NSDateFormatter *formatter = NULL;
    
    if (formatter == NULL) {
        formatter = [[NSDateFormatter alloc] init];
        [formatter setDateFormat:@"HH:mm:ss"];
    }
    NSString *dstr = [formatter stringFromDate:[NSDate date]];

    NSString *nsStr = [NSString stringWithFormat:@"%@ %@\n", dstr, message.body];
    // NSLog(@"log: %@ %@", self->logFile, message.body);
    if (self->logFile != NULL) {
        NSError *err;
        [self->logFile seekToEndOfFile];
        [self->logFile writeData:[nsStr dataUsingEncoding:NSUTF8StringEncoding]];
        [self->logFile synchronizeAndReturnError: &err];        
        // NSLog(@"log2file: %@", err);
    }
    // what ever were logged with console.log() in wkwebview arrives here in message.body property
}
@end

void WebViewClearAllCaches() {
    NSSet *dataTypes = [NSSet setWithArray:@[WKWebsiteDataTypeDiskCache,
                                         WKWebsiteDataTypeMemoryCache,
                                         ]];
    [[WKWebsiteDataStore defaultDataStore] removeDataOfTypes:dataTypes
                                           modifiedSince:[NSDate dateWithTimeIntervalSince1970:0]
                                           completionHandler:^{}];
}

void *NewWKWebView(int width, int height) {
    ZWKWebView *w = [[ZWKWebView alloc] initWithFrame: NSMakeRect(0, 0, width, height)];    
    w->logFile = NULL;
    return w;
}

void WebViewSetLogPath(void *v, const char *logPath) {
    ZWKWebView *w = (ZWKWebView *)v;

    NSString *nsPath = [NSString stringWithUTF8String:logPath];
    w->logFile = [NSFileHandle fileHandleForUpdatingAtPath:nsPath];
    if (w->logFile == NULL) {
        [[NSFileManager defaultManager] createFileAtPath:nsPath contents:nil attributes:nil];
        w->logFile = [NSFileHandle fileHandleForUpdatingAtPath:nsPath];
    }

    // NSLog(@"logfile: %@ at %s", w->logFile, logPath);
    NSString * js = @"var console = { log: function(msg){window.webkit.messageHandlers.logging.postMessage(msg) } };";
    [w evaluateJavaScript:js completionHandler:^(id _Nullable ignored, NSError * _Nullable error) {
        if (error != nil)
            NSLog(@"installation of console.log() failed: %@", error);
    }];   
    WKUserContentController *ucc = w.configuration.userContentController;
    [ucc addScriptMessageHandler:w name:@"logging"];
}

void WebViewSetURL(void *view, const char *surl) {
    NSString *nsStr = [NSString stringWithUTF8String:surl];
    NSURL *nsURL = [NSURL URLWithString: nsStr];
    NSURLRequest *nsReq = [NSURLRequest requestWithURL: nsURL];
    [ (ZWKWebView *)view loadRequest: nsReq];   
}

