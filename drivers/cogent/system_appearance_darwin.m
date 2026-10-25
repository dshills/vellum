// go:build darwin && cgo
#import <Cocoa/Cocoa.h>
#include <math.h>
#include <stdint.h>
extern int vellumSystemAppearanceActive(uint64_t);
extern void vellumSystemAppearance(uint64_t, int, int, int, int, int, int, int);
static char appearanceContext;
@interface VellumSystemAppearanceObserver : NSObject
@property uint64_t token;
@property(strong) NSApplication *application;
@property(strong) NSMutableArray *notifications;
@property(strong) id workspaceNotification;
@property BOOL observing;
- (void)sample;
- (void)stop;
@end
@implementation VellumSystemAppearanceObserver
- (void)sample {
  if (!vellumSystemAppearanceActive(self.token))
    return;
  NSAppearance *appearance = self.application.effectiveAppearance;
  BOOL dark = [[appearance bestMatchFromAppearancesWithNames:@[
    NSAppearanceNameAqua, NSAppearanceNameDarkAqua
  ]] isEqual:NSAppearanceNameDarkAqua];
  __block NSColor *accent = nil;
  [appearance performAsCurrentDrawingAppearance:^{
    accent = [NSColor.controlAccentColor
        colorUsingColorSpace:NSColorSpace.sRGBColorSpace];
  }];
  if (!accent)
    return;
  double red = accent.redComponent, green = accent.greenComponent,
         blue = accent.blueComponent;
  if (!isfinite(red) || !isfinite(green) || !isfinite(blue))
    return;
  NSWorkspace *workspace = NSWorkspace.sharedWorkspace;
  vellumSystemAppearance(self.token, dark,
                         workspace.accessibilityDisplayShouldIncreaseContrast,
                         workspace.accessibilityDisplayShouldReduceMotion,
                         workspace.accessibilityDisplayShouldReduceTransparency,
                         (int)lround(fmax(0, fmin(1, red)) * 255),
                         (int)lround(fmax(0, fmin(1, green)) * 255),
                         (int)lround(fmax(0, fmin(1, blue)) * 255));
}
- (void)observeValueForKeyPath:(NSString *)keyPath
                      ofObject:(id)object
                        change:(NSDictionary *)change
                       context:(void *)context {
  if (context == &appearanceContext) {
    [self sample];
    return;
  }
  [super observeValueForKeyPath:keyPath
                       ofObject:object
                         change:change
                        context:context];
}
- (void)stop {
  if (self.observing) {
    [self.application removeObserver:self
                          forKeyPath:@"effectiveAppearance"
                             context:&appearanceContext];
    self.observing = NO;
  }
  for (id token in self.notifications)
    [NSNotificationCenter.defaultCenter removeObserver:token];
  [self.notifications removeAllObjects];
  if (self.workspaceNotification)
    [NSWorkspace.sharedWorkspace.notificationCenter
        removeObserver:self.workspaceNotification];
  self.workspaceNotification = nil;
  self.application = nil;
}
@end
static VellumSystemAppearanceObserver *appearanceObserver;
void vellumStartSystemAppearance(uint64_t token) {
  dispatch_async(dispatch_get_main_queue(), ^{
    @autoreleasepool {
      if (!vellumSystemAppearanceActive(token) || !NSApp)
        return;
      if (appearanceObserver)
        return;
      VellumSystemAppearanceObserver *o = [VellumSystemAppearanceObserver new];
      o.token = token;
      o.application = NSApp;
      o.notifications = [NSMutableArray new];
      appearanceObserver = o;
      __weak VellumSystemAppearanceObserver *weak = o;
      for (NSString *name in @[
             NSSystemColorsDidChangeNotification,
             NSApplicationDidBecomeActiveNotification
           ]) {
        id receipt = [NSNotificationCenter.defaultCenter
            addObserverForName:name
                        object:nil
                         queue:NSOperationQueue.mainQueue
                    usingBlock:^(NSNotification *note) {
                      [weak sample];
                    }];
        [o.notifications addObject:receipt];
      }
      o.workspaceNotification = [NSWorkspace.sharedWorkspace.notificationCenter
          addObserverForName:
              NSWorkspaceAccessibilityDisplayOptionsDidChangeNotification
                      object:nil
                       queue:NSOperationQueue.mainQueue
                  usingBlock:^(NSNotification *note) {
                    [weak sample];
                  }];
      [NSApp addObserver:o
              forKeyPath:@"effectiveAppearance"
                 options:0
                 context:&appearanceContext];
      o.observing = YES;
      [o sample];
    }
  });
}
void vellumStopSystemAppearance(uint64_t token) {
  dispatch_async(dispatch_get_main_queue(), ^{
    if (appearanceObserver.token == token) {
      [appearanceObserver stop];
      appearanceObserver = nil;
    }
  });
}

void vellumStopAllSystemAppearance(void) { [appearanceObserver stop]; appearanceObserver=nil; }
