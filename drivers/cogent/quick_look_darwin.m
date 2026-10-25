// go:build darwin && cgo
#import "services_native.h"
#import <Cocoa/Cocoa.h>
#import <QuickLookUI/QuickLookUI.h>
#include <stdlib.h>
#include <string.h>
extern int vellumServiceActive(uint64_t);
extern void vellumServiceResult(uint64_t, char *, int);
extern void *vellumAXNativeWindow(uint64_t);
extern int vellumDocumentPanelBusy(void);
extern void vellumPositionOwnedPanel(NSPanel *, NSWindow *);
static BOOL vellumOwnsPreview;
@class VellumQuickLookSession;
static VellumQuickLookSession *vellumPreviewSession;
int vellumQuickLookBusy(void) { return vellumPreviewSession != nil; }
static NSDictionary *vellumQuickLookRecipe(NSData *data) {
  id o = [NSJSONSerialization JSONObjectWithData:data options:0 error:nil];
  if (![o isKindOfClass:NSDictionary.class] ||
      ![o[@"Items"] isKindOfClass:NSArray.class] || [o[@"Items"] count] < 1 ||
      [o[@"Items"] count] > 16)
    return nil;
  NSMutableSet *seen = [NSMutableSet new];
  BOOL selected = NO;
  for (id v in o[@"Items"]) {
    if (![v isKindOfClass:NSDictionary.class])
      return nil;
    for (NSString *key in @[ @"ID", @"Path", @"Title" ]) {
      id s = v[key];
      if (![s isKindOfClass:NSString.class] || ![s length] ||
          [s length] > ([key isEqual:@"Path"] ? 4096 : 256) ||
          [s rangeOfString:[NSString stringWithFormat:@"%C", (unichar)0]]
                  .location != NSNotFound)
        return nil;
    }
    if (![v[@"Path"] isAbsolutePath] || [seen containsObject:v[@"ID"]])
      return nil;
    [seen addObject:v[@"ID"]];
    selected = selected || [v[@"ID"] isEqual:o[@"SelectedID"]];
  }
  return selected ? o : nil;
}
char *vellumQuickLookRecipeFixture(const char *json) {
  @autoreleasepool {
    @try {
      if (!json || strnlen(json, VELLUM_SERVICE_REQUEST_MAX + 1) >
                       VELLUM_SERVICE_REQUEST_MAX)
        return NULL;
      NSDictionary *o =
          vellumQuickLookRecipe([NSData dataWithBytes:json
                                               length:strlen(json)]);
      return o ? strdup("valid") : NULL;
    } @catch (NSException *e) {
      return NULL;
    }
  }
}
@interface VellumQuickLookItem : NSObject <QLPreviewItem>
@property(strong) NSURL *previewItemURL;
@property(strong) NSString *previewItemTitle;
@end
@implementation VellumQuickLookItem
@end
@interface VellumQuickLookSession
    : NSResponder <QLPreviewPanelDataSource, QLPreviewPanelDelegate,
                   NSWindowDelegate>
@property uint64_t token;
@property BOOL finished;
@property(strong) QLPreviewPanel *panel;
@property(weak) NSWindow *parent;
@property(strong) NSResponder *savedNext;
@property(strong) NSArray<VellumQuickLookItem *> *items;
@property NSInteger selected;
@property(strong) id parentObserver;
- (void)finish:(int)status;
- (int)cleanup;
@end
@implementation VellumQuickLookSession
- (BOOL)acceptsPreviewPanelControl:(QLPreviewPanel *)panel {
  return !self.finished && panel == self.panel;
}
- (void)beginPreviewPanelControl:(QLPreviewPanel *)panel {
  if (self.finished)
    return;
  panel.dataSource = self;
  panel.delegate = self;
  [panel reloadData];
  panel.currentPreviewItemIndex = self.selected;
}
- (void)endPreviewPanelControl:(QLPreviewPanel *)panel {
  if (panel.dataSource == self)
    panel.dataSource = nil;
  if (panel.delegate == self)
    panel.delegate = nil;
}
- (NSInteger)numberOfPreviewItemsInPreviewPanel:(QLPreviewPanel *)panel {
  return self.finished ? 0 : self.items.count;
}
- (id<QLPreviewItem>)previewPanel:(QLPreviewPanel *)panel
               previewItemAtIndex:(NSInteger)index {
  return !self.finished && index >= 0 && index < self.items.count
             ? self.items[index]
             : nil;
}
- (BOOL)previewPanel:(QLPreviewPanel *)panel handleEvent:(NSEvent *)event {
  if (event.type == NSEventTypeKeyDown &&
      (event.keyCode == 53 || event.keyCode == 49)) {
    [self finish:0];
    return YES;
  }
  return NO;
}
- (BOOL)windowShouldClose:(NSWindow *)window {
  [self finish:0];
  return NO;
}
- (int)cleanup {
  int result = 0;
  @try {
    if (self.parentObserver) {
      [NSNotificationCenter.defaultCenter removeObserver:self.parentObserver];
      self.parentObserver = nil;
    }
  } @catch (NSException *e) {
    result = 2;
  } @
  try {
    if (self.panel.dataSource == self)
      self.panel.dataSource = nil;
    if (self.panel.delegate == self)
      self.panel.delegate = nil;
  } @catch (NSException *e) {
    result = 2;
  } @
  try {
    if (self.parent.nextResponder == self)
      self.parent.nextResponder = self.savedNext;
    self.nextResponder = nil;
    [self.parent removeChildWindow:self.panel];
    if (self.panel.inFullScreenMode)
      [self.panel exitFullScreenModeWithOptions:nil];
    [self.panel orderOut:nil];
    [self.panel updateController];
  } @catch (NSException *e) {
    result = 2;
  }
  if (!result) {
    self.items = nil;
    self.parent = nil;
    self.savedNext = nil;
    self.panel = nil;
    vellumPreviewSession = nil;
  }
  return result;
}
- (void)finish:(int)status {
  if (self.finished)
    return;
  self.finished = YES;
  if ([self cleanup])
    status = 2;
  vellumServiceResult(self.token, "", status);
}

@end
void vellumQuickLook(uint64_t token, uint64_t parentToken, const char *json) {
  vellumWithServiceJSON(token, json, ^(NSData *data) {
    dispatch_async(dispatch_get_main_queue(), ^{
      @autoreleasepool {
        @try {
          if (!vellumServiceActive(token)) {
            vellumServiceResult(token, "", 1);
            return;
          }
          NSDictionary *o = vellumQuickLookRecipe(data);
          if (!o) {
            vellumServiceResult(token, "", 2);
            return;
          }
          NSWindow *parent =
              (__bridge NSWindow *)vellumAXNativeWindow(parentToken);
          if (!parent) {
            vellumServiceResult(token, "", 6);
            return;
          }
          if (!parent.isVisible || parent.isMiniaturized) {
            vellumServiceResult(token, "", 2);
            return;
          }
          if (vellumPreviewSession || vellumDocumentPanelBusy() ||
              parent.attachedSheet || NSApp.modalWindow ||
              (!vellumOwnsPreview && QLPreviewPanel.sharedPreviewPanelExists)) {
            vellumServiceResult(token, "", 5);
            return;
          }
          QLPreviewPanel *panel = QLPreviewPanel.sharedPreviewPanel;
          if (vellumOwnsPreview && panel.isVisible) {
            vellumServiceResult(token, "", 5);
            return;
          }
          vellumOwnsPreview = YES;
          VellumQuickLookSession *session = [VellumQuickLookSession new];
          session.token = token;
          session.parent = parent;
          session.panel = panel;
          session.savedNext = parent.nextResponder;
          NSMutableArray *items = [NSMutableArray new];
          for (NSDictionary *entry in o[@"Items"]) {
            VellumQuickLookItem *item = [VellumQuickLookItem new];
            item.previewItemURL = [NSURL fileURLWithPath:entry[@"Path"]];
            item.previewItemTitle = entry[@"Title"];
            if ([entry[@"ID"] isEqual:o[@"SelectedID"]])
              session.selected = items.count;
            [items addObject:item];
          }
          session.items = items;
          vellumPreviewSession = session;
          session.nextResponder = parent.nextResponder;
          parent.nextResponder = session;
          panel.restorable = NO;
          panel.appearance = parent.effectiveAppearance;
          __weak VellumQuickLookSession *weak = session;
          session.parentObserver = [NSNotificationCenter.defaultCenter
              addObserverForName:NSWindowWillCloseNotification
                          object:parent
                           queue:nil
                      usingBlock:^(NSNotification *note) {
                        [weak finish:6];
                      }];
          if (!vellumServiceActive(token)) {
            [session finish:1];
            return;
          }
          [parent addChildWindow:panel ordered:NSWindowAbove];
          vellumPositionOwnedPanel(panel, parent);
          [panel makeKeyAndOrderFront:nil];
          [panel updateController];
          if (panel.currentController != session) {
            [session finish:5];
            return;
          }
          vellumPositionOwnedPanel(panel, parent);
        } @catch (NSException *e) {
          if (vellumPreviewSession && vellumPreviewSession.token == token)
            [vellumPreviewSession finish:2];
          else
            vellumServiceResult(token, "", 2);
        }
      }
    });
  });
}
void vellumCancelQuickLook(uint64_t token) {
  dispatch_async(dispatch_get_main_queue(), ^{
    if (vellumPreviewSession.token == token)
      [vellumPreviewSession finish:1];
    else
      vellumServiceResult(token, "", 1);
  });
}
int vellumCancelAllQuickLook(void) {
  @try {
    if (vellumPreviewSession.finished)
      [vellumPreviewSession cleanup];
    else if (vellumPreviewSession)
      [vellumPreviewSession finish:1];
    return vellumPreviewSession ? 2 : 0;
  } @catch (NSException *e) {
    return 2;
  }
}
