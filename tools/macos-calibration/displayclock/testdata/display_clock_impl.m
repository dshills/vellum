//go:build darwin && !offscreen
// Copyright (c) 2026 Davin Hills. SPDX-License-Identifier: MIT.
#import <Cocoa/Cocoa.h>
#import <QuartzCore/CADisplayLink.h>
#import <QuartzCore/CAMediaTiming.h>
#include <math.h>
#include <stdlib.h>
#include <string.h>
#include "display_clock.h"
#define DC_CAP 8192
struct DCTick { double observed, timestamp, target, duration; };
struct DCState {
 int configured, enabled, started, stopped, invalid, overflow;
 size_t count; long window; unsigned int display;
 double start, stop; struct DCTick rows[DC_CAP];
};
static struct DCState dc;
static id dc_target;
static NSNumber *dc_flag(int value){return value ? @YES : @NO;}
static CADisplayLink *dc_link;
static void dc_main(void (^work)(void)) {
 if(NSThread.isMainThread)work();else dispatch_sync(dispatch_get_main_queue(),work);
}
static void dc_append(double now,double timestamp,double target,double duration) {
 if(!dc.started || dc.stopped || !dc.enabled){dc.invalid=1;return;}
 if(!isfinite(now)||!isfinite(timestamp)||!isfinite(target)||!isfinite(duration)||now<=0||timestamp<=0||target<=timestamp||duration<=0){dc.invalid=1;return;}
 if(dc.count && (now<dc.rows[dc.count-1].observed || timestamp<=dc.rows[dc.count-1].timestamp || target<=dc.rows[dc.count-1].target)){dc.invalid=1;return;}
 if(dc.count==DC_CAP){dc.overflow=1;return;}
 dc.rows[dc.count++]=(struct DCTick){now,timestamp,target,duration};
}
#ifdef DC_TEST
static CADisplayLink *dc_make_link(id target);
static void dc_identity(long *window,unsigned int *display);
#else
static void dc_identity(long *window,unsigned int *display) {
 NSWindow *w=NSApp.keyWindow;*window=w.windowNumber;
 *display=[w.screen.deviceDescription[@"NSScreenNumber"] unsignedIntValue];
}
#endif
API_AVAILABLE(macos(14.0))
@interface DCObserver : NSObject
- (void)tick:(CADisplayLink *)link;
@end
@implementation DCObserver
- (void)tick:(CADisplayLink *)link {
 long window=0;unsigned int display=0;dc_identity(&window,&display);
 if(window!=dc.window || display!=dc.display){dc.invalid=1;return;}
 dc_append(CACurrentMediaTime(),link.timestamp,link.targetTimestamp,link.duration);
}
@end
#ifndef DC_TEST
static CADisplayLink *dc_make_link(id target) API_AVAILABLE(macos(14.0));
static CADisplayLink *dc_make_link(id target) {
 return [NSApp.keyWindow.contentView displayLinkWithTarget:target selector:@selector(tick:)];
}
#endif
int dc_configure(int enabled) {
 // One producer, before runtime/start; no callbacks exist yet.
 if(dc.configured || (enabled!=0 && enabled!=1))return 0;
 dc.configured=1;dc.enabled=enabled;return 1;
}
int dc_start(void) {
 __block int ok=0;
 dc_main(^{@autoreleasepool {
  if(!dc.configured || dc.started || dc.stopped)return;
  dc.started=1;dc.start=CACurrentMediaTime();dc_identity(&dc.window,&dc.display);
  if(dc.window<=0 || !dc.display){dc.invalid=1;return;}
  if(!dc.enabled){ok=1;return;}
  if(@available(macOS 14.0,*)) {
   dc_target=[DCObserver new];dc_link=[dc_make_link(dc_target) retain];
   if(!dc_link){[dc_target release];dc_target=nil;dc.invalid=1;return;}
   [dc_link addToRunLoop:NSRunLoop.mainRunLoop forMode:NSRunLoopCommonModes];ok=1;
  } else {dc.invalid=1;}
 }});return ok;
}
int dc_stop(void) {
 __block int ok=0;
 dc_main(^{@autoreleasepool {
  if(!dc.configured || dc.stopped)return;
  if(dc_link){[dc_link invalidate];[dc_link release];dc_link=nil;}
  [dc_target release];dc_target=nil;
  dc.stopped=1;dc.stop=CACurrentMediaTime();ok=1;
 }});return ok;
}
char *dc_snapshot(void) {
 __block char *out=NULL;
 dc_main(^{@autoreleasepool {
  NSMutableArray *rows=[NSMutableArray arrayWithCapacity:dc.count];
  for(size_t i=0;i<dc.count;i++){
   struct DCTick r=dc.rows[i];[rows addObject:@{@"ordinal":@(i),@"observed":@(r.observed),@"timestamp":@(r.timestamp),@"target":@(r.target),@"duration":@(r.duration)}];
  }
  NSDictionary *d=@{@"configured":dc_flag(dc.configured),@"enabled":dc_flag(dc.enabled),@"started":dc_flag(dc.started),@"stopped":dc_flag(dc.stopped),@"invalid":dc_flag(dc.invalid),@"overflow":dc_flag(dc.overflow),@"count":@(dc.count),@"window":@(dc.window),@"display":@(dc.display),@"start":@(dc.start),@"stop":@(dc.stop),@"records":rows};
  NSData *data=[NSJSONSerialization dataWithJSONObject:d options:NSJSONWritingSortedKeys error:nil];
  if(data && data.length<(2u<<20))out=strndup(data.bytes,data.length);
 }});return out;
}
