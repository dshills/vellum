//go:build darwin && cgo

#import <Cocoa/Cocoa.h>
#import <objc/runtime.h>
#include <stdint.h>
extern int vellumDropFixtureActive(uint64_t token);
extern void vellumDropFixtureResult(uint64_t token, int status);

// Only selectors consumed by GLFW's performDragOperation are supplied. This is
// a receive-path fixture, not a simulation of an OS drag session or Finder.
@interface VellumFileURLFixture : NSObject
@property NSPoint draggingLocation;
@property(strong) NSPasteboard *draggingPasteboard;
@end
@implementation VellumFileURLFixture
@end

// GLFW rereads NSWindow's live cursor during DropEvent. Supply the same fixture
// coordinate for that synchronous read, without moving the user's pointer.
// A temporary subclass adds no fields and restores the prior class immediately.
static NSPoint vellumFixtureCursor;
static NSMutableDictionary<NSString *,NSNumber *> *vellumFixtureWindows;
@interface VellumDropWindowMethods : NSWindow
@end
@implementation VellumDropWindowMethods
- (NSPoint)mouseLocationOutsideOfEventStream { return vellumFixtureCursor; }
@end
static Class vellumDropWindowClass(Class base) {
 NSString *name=[@"VellumDropFixture_" stringByAppendingString:NSStringFromClass(base)];
 Class cls=NSClassFromString(name);if(cls)return cls;
 cls=objc_allocateClassPair(base,name.UTF8String,0);if(!cls)return Nil;
 Method method=class_getInstanceMethod(VellumDropWindowMethods.class,@selector(mouseLocationOutsideOfEventStream));
 if(!class_addMethod(cls,method_getName(method),method_getImplementation(method),method_getTypeEncoding(method))){objc_disposeClassPair(cls);return Nil;}
 objc_registerClassPair(cls);return cls;
}

void vellumDropFixture(uint64_t token, const char *payload) {
 NSString *text=payload?[NSString stringWithUTF8String:payload]:nil;
 dispatch_async(dispatch_get_main_queue(), ^{
  if(!vellumDropFixtureActive(token))return;
  NSDictionary *data=text?[NSJSONSerialization JSONObjectWithData:[text dataUsingEncoding:NSUTF8StringEncoding] options:0 error:nil]:nil;
  if(![data isKindOfClass:NSDictionary.class]){vellumDropFixtureResult(token,2);return;}
  NSWindow *window=nil;
  if(!vellumFixtureWindows)vellumFixtureWindows=[NSMutableDictionary new];
  NSString *identity=data[@"Title"];NSNumber *number=vellumFixtureWindows[identity];
  for(NSWindow *candidate in NSApp.windows) {
   if(number?candidate.windowNumber==number.integerValue:[candidate.title isEqual:identity]){window=candidate;break;}
  }
  if(window&&!number)vellumFixtureWindows[identity]=@(window.windowNumber);
  NSView *view=window.contentView;
  if(!view||![view.registeredDraggedTypes containsObject:NSPasteboardTypeURL]||![view respondsToSelector:@selector(performDragOperation:)]){vellumDropFixtureResult(token,3);return;}
  NSPasteboard *pasteboard=[NSPasteboard pasteboardWithUniqueName];
  NSMutableArray *urls=[NSMutableArray new];
  for(NSString *path in data[@"Paths"])[urls addObject:[NSURL fileURLWithPath:path]];
  if(![pasteboard writeObjects:urls]){[pasteboard releaseGlobally];vellumDropFixtureResult(token,4);return;}
  VellumFileURLFixture *fixture=[VellumFileURLFixture new];fixture.draggingPasteboard=pasteboard;
  fixture.draggingLocation=[window convertPointFromScreen:NSMakePoint([data[@"X"] doubleValue],[data[@"Y"] doubleValue])];
  Class original=object_getClass(window), cursorClass=vellumDropWindowClass(original);
  if(!cursorClass){[pasteboard releaseGlobally];vellumDropFixtureResult(token,6);return;}
  vellumFixtureCursor=fixture.draggingLocation;
  BOOL accepted=NO;
  object_setClass(window,cursorClass);
  @try { accepted=[view performDragOperation:(id<NSDraggingInfo>)fixture]; }
  @finally { object_setClass(window,original); }
  [pasteboard releaseGlobally];vellumDropFixtureResult(token,accepted?1:5);
 });
}
