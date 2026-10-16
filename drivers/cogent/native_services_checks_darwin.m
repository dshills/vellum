//go:build darwin && cgo && !offscreen && vellum_native_checks

#import "native_services_checks.h"
#import "services_native.h"
#import "native_window_snapshot.h"
#import <objc/runtime.h>
#include <stdint.h>
#include <stdlib.h>
extern int vellumServiceActive(uint64_t token);
extern void vellumServiceResult(uint64_t token, char *text, int status);

static NSDictionary *checkSnapshot(BOOL cleanup) {
    NSDictionary *menu=vellumCheckMenu(0,0), *tray=vellumCheckTray(0,0), *panels=vellumCheckFilePanels();
    if (!menu || !tray || !panels) return nil;
    NSMutableArray *titles=[NSMutableArray array], *windows=[NSMutableArray array];
    NSArray<NSWindow *> *ordered=NSApp.orderedWindows;
    for (NSWindow *window in NSApp.windows) {
        if (windows.count>=64) return nil;
        NSRect f=window.frame;
        NSView *contentView=window.contentView;
        NSDictionary *sample=vellumWindowSnapshot(window.title,window.isVisible,window.isMiniaturized,window.isKeyWindow,window.occlusionState,f,NSStringFromClass(object_getClass(window)),contentView?NSStringFromClass(object_getClass(contentView)):nil,window.windowNumber);
        if (!sample) return nil;
        NSMutableDictionary *state=[NSMutableDictionary dictionaryWithDictionary:sample];
        NSUInteger index=[ordered indexOfObjectIdenticalTo:window];
        state[@"OrderedIndex"]=index==NSNotFound?@(-1):@(index);
        state[@"Opaque"]=@(window.isOpaque);
        state[@"Alpha"]=@(window.alphaValue);
        state[@"Level"]=@(window.level);
        [windows addObject:state];
        if (window.isVisible) [titles addObject:window.title ?: @""];
    }
    return @{@"MainThread":@(NSThread.isMainThread),@"CleanupObserved":@(cleanup),@"WindowTitles":titles,@"Windows":windows,@"Menu":menu,@"Tray":tray,@"FilePanels":panels};
}
static char *checkWire(NSDictionary *snapshot) {
    NSData *data=snapshot ? [NSJSONSerialization dataWithJSONObject:snapshot options:0 error:nil] : nil;
    if (!data || data.length>VELLUM_SERVICE_REQUEST_MAX) return NULL;
    NSString *text=[[NSString alloc] initWithData:data encoding:NSUTF8StringEncoding];
    const char *utf8=text.UTF8String;
    return utf8 ? strdup(utf8) : NULL;
}
void vellumNativeServicesCheck(uint64_t token,const char *json) {
    vellumWithServiceJSON(token,json,^(NSData *wire) {
        dispatch_async(dispatch_get_main_queue(),^{ @autoreleasepool { @try {
            if (!vellumServiceActive(token)) { vellumServiceResult(token,"",2);return; }
            NSDictionary *request=[NSJSONSerialization JSONObjectWithData:wire options:0 error:nil];
            if (![request isKindOfClass:NSDictionary.class] || request.count!=2) { vellumServiceResult(token,"",2);return; }
            for (id key in @[@"Operation",@"Index"]) {
                id n=request[key];
                if (![n isKindOfClass:NSNumber.class] || CFGetTypeID((__bridge CFTypeRef)n)==CFBooleanGetTypeID() || [n doubleValue]!=[n intValue]) { vellumServiceResult(token,"",2);return; }
            }
            int op=[request[@"Operation"] intValue],index=[request[@"Index"] intValue];
            if (op<0 || op>7 || index<0 || index>=128) { vellumServiceResult(token,"",2);return; }
            NSDictionary *changed=nil;
            if (op>=1 && op<=4) changed=vellumCheckMenu(op,index);
            else if (op>=5) changed=vellumCheckTray(op,index);
            if (op && !changed) { vellumServiceResult(token,"",2);return; }
            char *result=checkWire(checkSnapshot(NO));
            vellumServiceResult(token,result,result?0:2);free(result);
        } @catch(NSException *exception) { vellumServiceResult(token,"",2); } }});
    });
}
char *vellumNativeServicesCleanupCheck(void) {
    @autoreleasepool { @try {
        // Retained stale targets/items belong to this test, not production.
        // Release them before the observer records terminal ownership.
        vellumCheckMenu(8,0);vellumCheckTray(8,0);
        return checkWire(checkSnapshot(YES));
    } @catch(NSException *exception) { return NULL; } }
}
