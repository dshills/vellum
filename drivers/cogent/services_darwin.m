//go:build darwin && cgo

#import <Cocoa/Cocoa.h>
#import "services_native.h"
#import "native_services_checks.h"
#include <stdint.h>
#include <stdlib.h>
#include <string.h>
extern int vellumServiceActive(uint64_t token);
extern void vellumServiceResult(uint64_t token, char *text, int status);
extern int vellumMenuCommit(uint64_t token);
extern void vellumMenuAction(uint64_t token, int index);
extern int vellumCancelAllAlerts(void);
extern int vellumCancelAllDocumentPanels(void);
extern int vellumCancelAllQuickLook(void);
extern int vellumCancelAllShares(void);
extern int vellumCancelAllPaths(void);
static NSMutableDictionary<NSNumber *, NSSavePanel *> *vellumPanels;

// Restore only the exact live parent when our panel owned keyboard focus.
// Never activate another app, resurrect a closed/minimized window, or steal
// input from a different key window chosen while the panel was open.
static BOOL vellumPanelMayRestore(BOOL active, BOOL visible, BOOL minimized,
 id key, id panel, id parent, id main) {
 return parent && active && visible && !minimized &&
  (key == panel || key == parent || (!key && main == parent));
}
int vellumServicePanelRestoreFixture(void) {
 @autoreleasepool {
  id panel=[NSObject new],parent=[NSObject new],other=[NSObject new];
  NSArray *keys=@[panel,parent,other,NSNull.null];
  for(int flags=0;flags<8;flags++) for(int k=0;k<4;k++) for(int m=0;m<2;m++) {
   id key=k==3?nil:keys[k],main=m?parent:other;
   BOOL got=vellumPanelMayRestore((flags&1)!=0,(flags&2)!=0,(flags&4)!=0,key,panel,parent,main);
   BOOL want=flags==3 && (k<2 || (k==3 && m==1));
   if(got!=want)return 0;
  }
  if(vellumPanelMayRestore(YES,YES,NO,nil,panel,nil,nil))return 0;
  return 1;
 }
}

@interface VellumMenuTarget : NSObject
@property uint64_t token;
- (void)invoke:(NSMenuItem *)item;
@end
@implementation VellumMenuTarget
- (void)invoke:(NSMenuItem *)item { vellumMenuAction(self.token, (int)item.tag); }
@end
static VellumMenuTarget *vellumTarget;
static NSMenu *vellumOriginalMenu;
static int vellumCleanupStatus;
static const NSUInteger vellumPanelLimit = 8;

// A restoration comparison has meaning only after the diagnostic captures its
// original menu. Keep availability explicit; pointer equality alone is not a
// successful observation when both pointers happen to be nil.
static NSDictionary *vellumMenuObservation(BOOL owned, BOOL captured, BOOL matches, NSArray *items) {
    return @{@"Owned":@(owned), @"OriginalCaptured":@(captured),
             @"OriginalRestored":@((BOOL)(captured && matches)),
             @"TargetRetired":@((BOOL)!owned), @"Items":items};
}

// Exercise the production serializer with scalar inputs only. No NSApplication,
// menu installation, main-queue work or desktop session is involved.
char *vellumServiceMenuObservationFixture(void) {
    @autoreleasepool {
        NSMutableArray *records=[NSMutableArray new];
        for (int flags=0; flags<8; flags++) {
            [records addObject:vellumMenuObservation((flags&4)!=0, (flags&2)!=0,
                                                    (flags&1)!=0, @[])];
        }
        NSData *data=[NSJSONSerialization dataWithJSONObject:records options:0 error:nil];
        if (!data) return NULL;
        NSString *text=[[NSString alloc] initWithData:data encoding:NSUTF8StringEncoding];
        return text.UTF8String ? strdup(text.UTF8String) : NULL;
    }
}

#ifdef VELLUM_NATIVE_CHECKS
// Main-thread, read-only inventory of exactly the production-owned panels.
// Never inspect another process, select a URL, or invoke a panel action.
NSDictionary *vellumCheckFilePanels(void) {
    if (!NSThread.isMainThread || vellumPanels.count > vellumPanelLimit) return nil;
    NSUInteger visible=0, open=0, save=0;
    for (NSSavePanel *panel in vellumPanels.allValues) {
        if (![panel isKindOfClass:NSSavePanel.class]) return nil;
        if (panel.isVisible) visible++;
        if ([panel isKindOfClass:NSOpenPanel.class]) open++; else save++;
    }
    return @{@"Tracked":@(vellumPanels.count),@"Visible":@(visible),@"Open":@(open),@"Save":@(save)};
}
#endif


// Attempt both dismissal operations even if one raises. Exceptional owned panels
// remain tracked for terminal cleanup; their count has a separate fixed bound.
static int vellumDismissPanel(NSSavePanel *panel) {
    int status=0;
    @try { [panel cancel:nil]; } @catch (NSException *exception) { status=2; }
    @try { [panel orderOut:nil]; } @catch (NSException *exception) { status=2; }
    return status;
}

@interface VellumDismissFixture : NSObject
@property BOOL failCancel, failOrder;
@property int canceled, ordered;
@end
@implementation VellumDismissFixture
- (void)cancel:(id)sender { self.canceled++; if (self.failCancel) @throw [NSException exceptionWithName:@"fixture" reason:@"cancel" userInfo:nil]; }
- (void)orderOut:(id)sender { self.ordered++; if (self.failOrder) @throw [NSException exceptionWithName:@"fixture" reason:@"order" userInfo:nil]; }
@end
// Injected objects exercise the production exception fence without NSSavePanel.
int vellumServiceDismissFixture(void) {
    @autoreleasepool {
        for (int flags=0; flags<4; flags++) {
            VellumDismissFixture *panel=[VellumDismissFixture new];
            panel.failCancel=(flags&1)!=0; panel.failOrder=(flags&2)!=0;
            int status=vellumDismissPanel((NSSavePanel *)panel);
            if (panel.canceled!=1 || panel.ordered!=1 || status!=(flags ? 2 : 0)) return 0;
        }
        return 1;
    }
}

// Drain temporary producer objects on the originating Go worker. Main-queue
// consumers strongly capture detached NSData, independently of caller storage.
BOOL vellumWithServiceJSON(uint64_t token, const char *json, void (^consume)(NSData *)) {
    if (!json) { vellumServiceResult(token, "", 2); return NO; }
    size_t length = strnlen(json, VELLUM_SERVICE_REQUEST_MAX + 1);
    if (length > VELLUM_SERVICE_REQUEST_MAX) { vellumServiceResult(token, "", 3); return NO; }
    @autoreleasepool {
        @try {
            NSData *payload = [NSData dataWithBytes:json length:length];
            if (!payload) { vellumServiceResult(token, "", 2); return NO; }
            consume(payload);
            return YES;
        } @catch (NSException *exception) { vellumServiceResult(token, "", 2); return NO; }
    }
}

// Production producer, delayed injected consumer; no panel/menu/status-bar work.
int vellumServicePayloadFixture(const char *json) {
    if (!json || strnlen(json, VELLUM_SERVICE_REQUEST_MAX + 1) > VELLUM_SERVICE_REQUEST_MAX) return 0;
    @autoreleasepool {
        size_t length = strlen(json);
        char *input = strdup(json); if (!input) return 0;
        NSData *expected = [[NSData alloc] initWithBytes:input length:length];
        __block dispatch_block_t queued = nil;
        __block __weak NSData *weakPayload = nil;
        __block BOOL copied = NO;
        BOOL accepted = vellumWithServiceJSON(0, input, ^(NSData *payload) {
            weakPayload = payload;
            queued = [^{ @autoreleasepool { copied = [payload isEqualToData:expected]; } } copy];
        });
        memset(input, 'x', length); free(input);
        BOOL alive; @autoreleasepool { alive = weakPayload != nil; }
        if (queued) queued(); queued = nil;
        BOOL released; @autoreleasepool { released = weakPayload == nil; }
        return accepted ? ((alive ? 1 : 0) | (copied ? 2 : 0) | (released ? 4 : 0)) : 0;
    }
}

void vellumChooseFile(uint64_t token, const char *json) {
    vellumWithServiceJSON(token, json, ^(NSData *data) {
    dispatch_async(dispatch_get_main_queue(), ^{
      @autoreleasepool { @try {
        if (!vellumServiceActive(token)) { vellumServiceResult(token, "", 2); return; }
        NSDictionary *options = [NSJSONSerialization JSONObjectWithData:data options:0 error:nil];
        if (![options isKindOfClass:[NSDictionary class]]) { vellumServiceResult(token, "", 2); return; }
        BOOL save = [options[@"Mode"] intValue] == 1;
        if (vellumPanels.count >= vellumPanelLimit) { vellumServiceResult(token, "", 2); return; }
        NSSavePanel *panel = save ? [NSSavePanel savePanel] : [NSOpenPanel openPanel];
        if (!save) {
            NSOpenPanel *open = (NSOpenPanel *)panel;
            open.canChooseDirectories = NO; open.canChooseFiles = YES; open.allowsMultipleSelection = NO;
        }
        NSString *title = options[@"Title"];
        if (title.length) panel.title = title;
        NSString *directory = options[@"Directory"];
        if (directory.length) panel.directoryURL = [NSURL fileURLWithPath:directory isDirectory:YES];
        NSString *filename = options[@"Filename"];
        if (filename.length) panel.nameFieldStringValue = filename;
        NSArray *extensions = options[@"Extensions"];
        if ([extensions isKindOfClass:[NSArray class]] && extensions.count) {
#pragma clang diagnostic push
#pragma clang diagnostic ignored "-Wdeprecated-declarations"
            panel.allowedFileTypes = extensions;
#pragma clang diagnostic pop
        }
        if (!vellumPanels) vellumPanels = [NSMutableDictionary new];
        if (!vellumServiceActive(token)) { vellumServiceResult(token, "", 2); return; }
        vellumPanels[@(token)] = panel;
        __weak NSWindow *parent = NSApp.keyWindow;
        __weak NSResponder *responder = parent.firstResponder;
        [panel beginWithCompletionHandler:^(NSModalResponse response) {
          @autoreleasepool {
            @try {
            NSString *path = response == NSModalResponseOK ? panel.URL.path : @"";
            int status = response == NSModalResponseOK ? (path ? 0 : 2) : 1;
            NSWindow *origin = parent;
            if (vellumPanelMayRestore(NSApp.isActive, origin.isVisible, origin.isMiniaturized, NSApp.keyWindow, panel, origin, NSApp.mainWindow)) {
              [panel orderOut:nil];
              [origin makeKeyAndOrderFront:nil];
              NSResponder *saved = responder;
              if ([saved isKindOfClass:NSView.class] && ((NSView *)saved).window == origin)
                [origin makeFirstResponder:saved];
            }
            vellumServiceResult(token, (char *)(path ?: @"").UTF8String, status);
            } @catch (NSException *exception) { vellumServiceResult(token, "", 2); }
            @finally { [vellumPanels removeObjectForKey:@(token)]; }
          }
        }];
      } @catch (NSException *exception) {
        vellumServiceResult(token, "", 2);
        NSSavePanel *panel = vellumPanels[@(token)];
        if (vellumDismissPanel(panel)) vellumCleanupStatus=2;
        else [vellumPanels removeObjectForKey:@(token)];
      } }
    });
    });
}
void vellumCancelFile(uint64_t token) {
    dispatch_async(dispatch_get_main_queue(), ^{
        @autoreleasepool {
            int status=2;
            @try {
                NSSavePanel *panel = vellumPanels[@(token)];
                status=vellumDismissPanel(panel);
                if (status) vellumCleanupStatus=2;
                else [vellumPanels removeObjectForKey:@(token)];
            } @catch (NSException *exception) { vellumCleanupStatus=2; }
            @finally { vellumServiceResult(token, "", status ? 2 : 1); }
        }
    });
}
void vellumSetMenu(uint64_t token, const char *json) {
    vellumWithServiceJSON(token, json, ^(NSData *data) {
    dispatch_async(dispatch_get_main_queue(), ^{
      @autoreleasepool { @try {
        if (!vellumServiceActive(token)) { vellumServiceResult(token, "", 2); return; }
        NSArray *groups = [NSJSONSerialization JSONObjectWithData:data options:0 error:nil];
        if (![groups isKindOfClass:[NSArray class]]) { vellumServiceResult(token, "", 2); return; }
        if (!vellumOriginalMenu) vellumOriginalMenu = NSApp.mainMenu;
        NSMenu *menu = [NSMenu new];
        // Preserve the application menu supplied by GLFW (About/Hide/Quit).
        if (vellumOriginalMenu.numberOfItems > 0) [menu addItem:[vellumOriginalMenu itemAtIndex:0].copy];
        VellumMenuTarget *target = [VellumMenuTarget new]; target.token = token;
        for (NSDictionary *group in groups) {
            if (![group isKindOfClass:[NSDictionary class]] ||
                ![group[@"Label"] isKindOfClass:[NSString class]] ||
                ![group[@"Items"] isKindOfClass:[NSArray class]]) {
                vellumServiceResult(token, "", 2); return;
            }
            NSMenuItem *heading = [[NSMenuItem alloc] initWithTitle:group[@"Label"] action:nil keyEquivalent:@""];
            NSMenu *submenu = [[NSMenu alloc] initWithTitle:group[@"Label"]]; submenu.autoenablesItems = NO;
            for (NSDictionary *item in group[@"Items"]) {
                if (![item isKindOfClass:[NSDictionary class]] ||
                    ![item[@"Label"] isKindOfClass:[NSString class]] ||
                    ![item[@"Index"] isKindOfClass:[NSNumber class]] ||
                    ![item[@"Disabled"] isKindOfClass:[NSNumber class]]) {
                    vellumServiceResult(token, "", 2); return;
                }
                NSMenuItem *entry = [[NSMenuItem alloc] initWithTitle:item[@"Label"] action:@selector(invoke:) keyEquivalent:@""];
                NSString *shortcut=item[@"Shortcut"];
                if([shortcut isKindOfClass:NSString.class] && shortcut.length<=128) {
                 NSArray<NSString *> *parts=[shortcut componentsSeparatedByString:@"+"];NSString *key=parts.lastObject.lowercaseString;
                 NSEventModifierFlags flags=0;
                 for(NSString *part in parts) {if([part caseInsensitiveCompare:@"Meta"]==NSOrderedSame)flags|=NSEventModifierFlagCommand;else if([part caseInsensitiveCompare:@"Ctrl"]==NSOrderedSame)flags|=NSEventModifierFlagControl;else if([part caseInsensitiveCompare:@"Shift"]==NSOrderedSame)flags|=NSEventModifierFlagShift;else if([part caseInsensitiveCompare:@"Alt"]==NSOrderedSame)flags|=NSEventModifierFlagOption;}
                 if(key.length==1 && flags) {entry.keyEquivalent=key;entry.keyEquivalentModifierMask=flags;}
                }
                entry.target = target; entry.tag = [item[@"Index"] integerValue];
                entry.enabled = ![item[@"Disabled"] boolValue]; [submenu addItem:entry];
            }
            heading.submenu = submenu; [menu addItem:heading];
        }
        // Keep the previous target and menu untouched until ownership is accepted.
        // Cancellation during construction cannot install a menu without callbacks.
        if (!vellumMenuCommit(token)) { vellumServiceResult(token, "", 2); return; }
        NSApp.mainMenu = menu; vellumTarget = target;
        vellumServiceResult(token, "", 0);
      } @catch (NSException *exception) { vellumServiceResult(token, "", 2); } }
    });
    });
}
extern int vellumStopAllSystemAppearance(void);
extern void vellumNativeContentShutdown(void);
int vellumShutdownOnMain(void) {
    @autoreleasepool {
        int status=vellumStopAllSystemAppearance();
        vellumNativeContentShutdown();
        if(vellumCancelAllAlerts())status=2;
 if(vellumCancelAllDocumentPanels())status=2;
 if(vellumCancelAllQuickLook())status=2;
 if(vellumCancelAllShares())status=2;
 if(vellumCancelAllPaths())status=2;
        for (NSSavePanel *panel in vellumPanels.allValues.copy) {
            if (vellumDismissPanel(panel)) status=2;
        }
        [vellumPanels removeAllObjects];
        @try { if (vellumOriginalMenu) NSApp.mainMenu= vellumOriginalMenu; }
        @catch (NSException *exception) { status=2; }
        vellumOriginalMenu=nil; vellumTarget=nil;
        @try { vellumTrayShutdown(); }
        @catch (NSException *exception) { status=2; }
        if (status) vellumCleanupStatus=status;
        return vellumCleanupStatus;
    }
}
void vellumShutdown(void) {
    dispatch_async(dispatch_get_main_queue(), ^{ vellumShutdownOnMain(); });
}

#ifdef VELLUM_NATIVE_CHECKS
#import "native_services_checks.h"
static __weak NSMenu *checkOriginalMenu;
static BOOL checkOriginalCaptured;
static VellumMenuTarget *checkStaleMenuTarget;
static NSMenuItem *checkStaleMenuItem;
NSDictionary *vellumCheckMenu(int operation,int index) {
    if (!NSThread.isMainThread) return nil;
    if (operation==1) {
        if (checkOriginalCaptured) return nil;
        checkOriginalMenu=NSApp.mainMenu;checkOriginalCaptured=YES;
    }
    if (operation==8) { checkStaleMenuTarget=nil;checkStaleMenuItem=nil; }
    if (operation==4) {
        if (!checkStaleMenuTarget || !checkStaleMenuItem) return nil;
        [checkStaleMenuTarget invoke:checkStaleMenuItem];
    }
    NSMutableArray *items=[NSMutableArray new];NSMenu *foundMenu=nil;NSInteger foundIndex=-1;
    for (NSMenuItem *group in NSApp.mainMenu.itemArray) {
        for (NSMenuItem *item in group.submenu.itemArray) {
            if (!vellumTarget || item.target!=vellumTarget || item.action!=@selector(invoke:)) continue;
            [items addObject:@{@"label":item.title,@"index":@(item.tag),@"enabled":@(item.enabled),@"separator":@NO}];
            if (item.tag==index) { foundMenu=group.submenu;foundIndex=[foundMenu indexOfItem:item]; }
        }
    }
    if (operation==2 || operation==3) {
        if (!foundMenu || foundIndex<0) return nil;
        NSMenuItem *item=[foundMenu itemAtIndex:foundIndex];
        if (operation==2) [foundMenu performActionForItemAtIndex:foundIndex];
        else { checkStaleMenuTarget=vellumTarget;checkStaleMenuItem=item.copy; }
    }
    return vellumMenuObservation(vellumTarget!=nil, checkOriginalCaptured,
                                 NSApp.mainMenu==checkOriginalMenu, items);
}
#endif
