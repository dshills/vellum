// Copyright 2026 Vellum contributors. SPDX-License-Identifier: MIT.
// AppKit-only, hidden, synthetic-text reference. Does not link Vellum or GLFW.
#import <AppKit/AppKit.h>
#import <CoreFoundation/CoreFoundation.h>
#import <objc/runtime.h>
#include <stdio.h>
#include <string.h>
#include <unistd.h>
#import "native_window_snapshot.h"

static const NSUInteger MaxWindows=64, MaxClasses=64, MaxChain=16, MaxPacket=512*1024;
static __weak NSWindow *ownedWindow;
static __weak NSView *ownedView;
static BOOL textMode, allOwnedExpired=YES;

static void need(BOOL ok, NSString *message) {
    if (!ok) @throw [NSException exceptionWithName:@"ReferenceRefusal" reason:message userInfo:nil];
}
static NSString *bounded(NSString *value, NSUInteger limit) {
    value=value?:@"";
    need([value lengthOfBytesUsingEncoding:NSUTF8StringEncoding]<=limit,@"metadata byte bound");
    return value;
}
static NSArray *classChain(Class cls, NSMutableDictionary *details) {
    NSMutableArray *chain=[NSMutableArray array];
    for (NSUInteger i=0;cls;i++,cls=class_getSuperclass(cls)) {
        need(i<MaxChain,@"class chain bound");
        NSString *name=bounded(NSStringFromClass(cls),256);
        need(name.length>0,@"missing runtime class");
        [chain addObject:name];
        if (!details[name]) {
            need(details.count<MaxClasses,@"class count bound");
            const char *image=class_getImageName(cls);
            NSString *imagePath=image?[[NSString alloc] initWithUTF8String:image]:nil;
            need(!image||imagePath!=nil,@"invalid runtime image UTF8");
            NSBundle *bundle=[NSBundle bundleForClass:cls];
            Class superClass=class_getSuperclass(cls);
            details[name]=@{@"Class":name,
                @"Superclass":bounded(superClass?NSStringFromClass(superClass):nil,256),
                @"RuntimeImageAvailable":@(image!=NULL),
                @"RuntimeImage":bounded(imagePath,4096),
                @"BundleIdentifier":bounded(bundle.bundleIdentifier,512),
                @"BundlePath":bounded(bundle.bundlePath,4096)};
        }
    }
    return chain;
}
static NSDictionary *inventory(NSString *phase, NSInteger cycle) {
    need(NSThread.isMainThread,@"snapshot must be on native main");
    NSMutableArray *windows=[NSMutableArray array];
    NSMutableDictionary *details=[NSMutableDictionary dictionary];
    NSArray<NSWindow *> *ordered=NSApp.orderedWindows;
    for (NSWindow *window in NSApp.windows) {
        need(windows.count<MaxWindows,@"window count bound");
        NSView *view=window.contentView;
        NSDictionary *raw=vellumWindowSnapshot(bounded(window.title,1024),window.isVisible,
            window.isMiniaturized,window.isKeyWindow,window.occlusionState,window.frame,
            NSStringFromClass(object_getClass(window)),view?NSStringFromClass(object_getClass(view)):nil,
            window.windowNumber);
        need(raw!=nil,@"production snapshot refused runtime metadata");
        NSMutableDictionary *record=[raw mutableCopy];
        NSUInteger index=[ordered indexOfObjectIdenticalTo:window];
        record[@"OrderedIndex"]=index==NSNotFound?@(-1):@(index);
        record[@"Opaque"]=@(window.isOpaque);record[@"Alpha"]=@(window.alphaValue);record[@"Level"]=@(window.level);
        record[@"OwnedFixtureWindow"]=@(window==ownedWindow);
        record[@"OwnedFixtureView"]=@(view!=nil&&view==ownedView);
        record[@"WindowClassChain"]=classChain(object_getClass(window),details);
        record[@"ContentViewClassChain"]=classChain(view?object_getClass(view):Nil,details);
        [windows addObject:record];
    }
    NSMutableArray *classes=[NSMutableArray array];
    for (NSString *name in [[details allKeys] sortedArrayUsingSelector:@selector(compare:)]) [classes addObject:details[name]];
    return @{@"Schema":@"appkit-only-hidden-window-reference-v1",@"Phase":phase,@"Cycle":@(cycle),
        @"PID":@(getpid()),@"MainThread":@YES,@"TextMode":@(textMode),
        @"SystemUptimeSeconds":@(NSProcessInfo.processInfo.systemUptime),
        @"OperatingSystem":bounded(NSProcessInfo.processInfo.operatingSystemVersionString,512),
        @"OwnedWindowWeakExpired":@(ownedWindow==nil),@"OwnedViewWeakExpired":@(ownedView==nil),
        @"Windows":windows,@"RuntimeClassDetails":classes};
}
static void emit(NSDictionary *record) {
    NSError *error=nil;
    NSData *data=[NSJSONSerialization dataWithJSONObject:record options:NSJSONWritingSortedKeys error:&error];
    need(data!=nil&&error==nil&&data.length<=MaxPacket,@"JSON packet refusal");
    need(fwrite(data.bytes,1,data.length,stdout)==data.length&&fputc('\n',stdout)!=EOF&&fflush(stdout)==0,@"stdout write failure");
}
static void sample(NSString *phase,NSInteger cycle) { @autoreleasepool { emit(inventory(phase,cycle)); } }
static void drainFor(NSTimeInterval duration) {
    // Each system callback is still synchronous; the outer process watchdog is
    // required. This fixed-duration loop neither waits for a class to disappear
    // nor filters out a window. Local pools retire observer/runloop temporaries.
    NSTimeInterval end=NSProcessInfo.processInfo.systemUptime+duration;
    while (NSProcessInfo.processInfo.systemUptime<end) {
        @autoreleasepool {
            SInt32 result=CFRunLoopRunInMode(kCFRunLoopDefaultMode,0.01,true);
            if (result==kCFRunLoopRunFinished) usleep(1000); // no busy spin without sources
        }
    }
}
static void cycle(NSInteger index) {
    @autoreleasepool {
        NSWindow *window=[[NSWindow alloc] initWithContentRect:NSMakeRect(0,0,320,160)
            styleMask:NSWindowStyleMaskTitled|NSWindowStyleMaskClosable backing:NSBackingStoreBuffered defer:NO];
        window.releasedWhenClosed=NO;
        window.title=[NSString stringWithFormat:@"AppKit-only synthetic reference %ld",(long)index];
        NSView *view=textMode?[[NSTextView alloc] initWithFrame:NSMakeRect(0,0,320,160)]:[[NSView alloc] initWithFrame:NSMakeRect(0,0,320,160)];
        window.contentView=view;ownedWindow=window;ownedView=view;
        need(!window.visible&&!window.keyWindow,@"reference unexpectedly visible/key");
        sample(@"created-hidden",index);
        if (textMode) {
            NSTextView *editor=(NSTextView *)view;
            editor.editable=YES;editor.selectable=YES;
            editor.string=@"owned synthetic reference text";
            need([window makeFirstResponder:editor],@"hidden local first responder refused");
            // Getter only: never activate an input source/context or send an OS
            // event. AppKit may initialize its own text services lazily here.
            NSTextInputContext *context=editor.inputContext;
            need(context!=nil,@"text view input context unavailable");
            [editor setSelectedRange:NSMakeRange(editor.string.length,0)];
            [editor insertText:@"." replacementRange:NSMakeRange(NSNotFound,0)];
            [editor setMarkedText:@"owned preedit" selectedRange:NSMakeRange(5,0) replacementRange:NSMakeRange(NSNotFound,0)];
            need(editor.hasMarkedText,@"local marked text not installed");
            [editor unmarkText];need(!editor.hasMarkedText,@"local marked text not removed");
            sample(@"local-text-client-exercised",index);
            [context discardMarkedText];
            context=nil;
        } else sample(@"plain-view-no-text-client",index);
        [window makeFirstResponder:nil];
        [window orderOut:nil];[window close];
        sample(@"closed-before-owned-release",index);
        window.contentView=nil;view=nil;window=nil;
    }
    sample(@"after-owned-pool-drain",index);
    drainFor(0.1);sample(@"after-fixed-100ms-drain",index);
    drainFor(0.4);sample(@"after-fixed-500ms-drain",index);
    allOwnedExpired=allOwnedExpired&&ownedWindow==nil&&ownedView==nil;
}
int main(int argc,const char **argv) {
    @autoreleasepool { @try {
        need(argc==2&&(!strcmp(argv[1],"--plain")||!strcmp(argv[1],"--text")),@"expected one fixed mode: --plain or --text");
        textMode=!strcmp(argv[1],"--text");
        need(NSThread.isMainThread,@"native main required");
        [NSApplication sharedApplication];
        need([NSApp setActivationPolicy:NSApplicationActivationPolicyProhibited],@"process-local prohibited activation policy refused");
        [NSApp finishLaunching];
        sample(@"appkit-initialized-baseline",-1);
        for (NSInteger index=0;index<2;index++) cycle(index);
        sample(@"terminal-all-observed-windows",2);
        emit(@{@"Schema":@"appkit-only-hidden-window-reference-v1",@"Phase":@"result",@"PID":@(getpid()),
            @"ProtocolCompleted":@YES,@"TextMode":@(textMode),@"Cycles":@2,@"AllOwnedWeakReferencesExpired":@(allOwnedExpired),
            @"Classification":@"observations-only-not-a-leak-or-rights-qualification"});
        return 0;
    } @catch(NSException *exception) {
        NSString *reason=exception.reason?:@"unknown exception";
        if ([reason lengthOfBytesUsingEncoding:NSUTF8StringEncoding]>1024) reason=@"exception reason exceeded byte bound";
        fprintf(stderr,"AppKit reference refused: %s\n",reason.UTF8String);
        return 1;
    } }
}
