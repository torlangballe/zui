import Foundation
import AppKit
import CoreGraphics
import Cocoa
import IOKit
//import AVFoundation

@_cdecl("GetAppContentPath") public func GetAppContentPath() -> UnsafePointer<CChar> {
    let str = Bundle.main.bundleURL.path + "/Contents/"
    let s = (str as NSString).utf8String
    let buffer = UnsafePointer<Int8>(s)
    return buffer!
}

@_cdecl("GetNetworkTrafficBytes") public func GetNetworkTrafficBytes() -> Int64 {
    // let info = DataUsage.getDataUsage()
    // return Int64(info.lanReceived) + Int64(info.lanSent) + Int64(info.wirelessWanDataReceived) + Int64(info.wirelessWanDataSent)
    return 0
}

@_cdecl("GetMainScreenScale") public func GetMainScreenScale() -> Int {
    return Int(NSScreen.main!.backingScaleFactor)
}

@_cdecl("GetDeviceUniqueID") public func GetDeviceUniqueID() -> UnsafeMutablePointer<CChar> {

    var hwUUIDBytes: [UInt8] = [0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0]
    var ts = timespec(tv_sec: 0,tv_nsec: 0)
    gethostuuid(&hwUUIDBytes, &ts)
    var str = ""
    for c in hwUUIDBytes {
        str += String(c)
    }
    return strdup(str)
    
    // let matchingDict = IOServiceMatching("IOPlatformExpertDevice")
    // let platformExpert = IOServiceGetMatchingService(kIOMasterPortDefault, matchingDict)
    // defer{ IOObjectRelease(platformExpert) }

    // var s = ""
    // if platformExpert != 0 {
    //      s = IORegistryEntryCreateCFProperty(platformExpert, kIOPlatformUUIDKey as CFString, kCFAllocatorDefault, 0).takeRetainedValue() as! String
    // }
    // // let s = NSDevice.current.identifierForVendor!.uuidString
    // let buffer = UnsafePointer<Int8>(s)
    // return buffer
}

// @_cdecl("runOnMain") public func runOnMain(index Int32) {
//     DispatchQueue.main.async() {
//         runFuncFromQue(index)
//     }
// }

var forceScreenRecording = true
func getWindowIDs(got: (_ wid: Int, _ pid: Int, _ title: String, _ bounds: CGRect)->Bool) {
    // invoking a screen shot prompts user to allow app to record screen, which allows WindowName (title) to be gotten
    // https://stackoverflow.com/questions/56597221/detecting-screen-recording-settings-on-macos-catalina
    if forceScreenRecording { 
        let _ = CGDisplayStream.init(display: CGMainDisplayID(), outputWidth: 1, outputHeight:1, pixelFormat: Int32(kCVPixelFormatType_32BGRA), properties: nil) { (CGDisplayStreamFrameStatus, UInt64, IOSurfaceRef, CGDisplayStreamUpdate) in
            print("screen capture!!\n")       
        }
        forceScreenRecording = false
    }
    // stream.start() is not called, so it never starts recording, and goes out of scope
    let options = CGWindowListOption(rawValue: CGWindowListOption.optionOnScreenOnly.rawValue | CGWindowListOption.excludeDesktopElements.rawValue)
    let windows = CGWindowListCopyWindowInfo(options, kCGNullWindowID) as NSArray? as? [[String: AnyObject]]
    for win in windows! {
        var app = ""
        var title = ""
        if let ownerName = win[kCGWindowOwnerName as String] as? String {
            app = ownerName
        }
        if let windowName = win[kCGWindowName as String] as? String {// not part of dictionary if screen recording not allowed
            title = windowName
        }
        let windowNum = win[kCGWindowNumber as String] as? Int ?? 0
        let windowPid = win[kCGWindowOwnerPID as String] as? Int ?? 0
        let frameDict = win[kCGWindowBounds as String] as! NSDictionary
        var bounds = CGRect()
        bounds.origin.x = frameDict["X"] as! CGFloat
        bounds.origin.y = frameDict["Y"] as! CGFloat
        bounds.size.width = frameDict["Width"] as! CGFloat
        bounds.size.height = frameDict["Height"] as! CGFloat
        if !got(windowNum, windowPid, title, app, bounds) {
            return
        }
    }
}

func getBestScreenForBounds(bounds: CGRect) -> NSScreen {
    var bestScreen = NSScreen()
    var bestArea: CGFloat = 0
    for s in NSScreen.screens {
        let inter = s.frame.intersection(bounds)
        let a = inter.size.width * inter.size.height
        if a > bestArea {
            bestArea = a
            bestScreen = s
        }
        // print("union:", inter, a, s.frame, bounds)
    }
    return bestScreen
}

@_cdecl("DoesWindowWithTitleExists") public func DoesWindowWithTitleExists(title: UnsafeMutablePointer<Int8>) -> Int {
    var found = 0
    let stitle = String(cString: title)
    getWindowIDs() { (id, pid, t, a, bounds) in 
        if t == stitle {
            found = 1
            return false
        }
        return true
    }
    return found
}

@_cdecl("GetWindowIDAndScaleForTitle") public func GetWindowIDAndScaleForTitle(app: UnsafeMutablePointer<Int8>, title: UnsafeMutablePointer<Int8>) -> UnsafeMutablePointer<Int8> {
    var str = ""
    let stitle = String(cString: title)
    let sapp = String(cString: app)
    getWindowIDs() { (id, pid, t, a, bounds) in 
        // print("windowID:", id, pid, t, a, bounds)
        if stitle == t {
            let screen = getBestScreenForBounds(bounds: bounds)
            str = String(format: "%d@%d", id, Int(screen.backingScaleFactor))
            return false
        }
        return true
    }
    return strdup(str)
    // return str.withCString { (cstr) -> UnsafePointer<Int8> in
    //     strdup(cstr)! as! UnsafePointer<Int8>;
    // }
}

@_cdecl("GetFirstWindowIDForPID") public func GetFirstWindowIDForPID(pid: Int) -> Int {
    var wid = 0
    getWindowIDs() { (id, p, t, a, scale) in 
        print(id, p, pid, t, a, scale)
        if pid == p {
            wid = id
            return false
        }
        return true
    }
    return wid
}

