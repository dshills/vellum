// Copyright (c) 2026 Davin Hills. SPDX-License-Identifier: MIT.
#define DC_TEST 1
#ifndef DC_IMPL
#define DC_IMPL "display_clock_impl.m"
#endif
#import DC_IMPL
#include <assert.h>
static int made,freed,invalidated,factory_nil;static double fake_timestamp;static long testWindow=42;static unsigned int testDisplay=7;
@interface FakeDCLink : NSObject {
 id _target;
}
- (instancetype)initWithTarget:(id)target;
- (void)addToRunLoop:(NSRunLoop*)loop forMode:(NSRunLoopMode)mode;
- (void)invalidate;
@end
@implementation FakeDCLink
- (instancetype)initWithTarget:(id)target {if((self=[super init])){_target=[target retain];made++;}return self;}
- (void)addToRunLoop:(NSRunLoop*)loop forMode:(NSRunLoopMode)mode {(void)loop;(void)mode;}
- (void)invalidate {invalidated++;[_target release];_target=nil;}
- (CFTimeInterval)timestamp{return fake_timestamp;}
- (CFTimeInterval)targetTimestamp{return fake_timestamp+1.0/60.0;}
- (CFTimeInterval)duration{return 1.0/60.0;}
- (void)dealloc {assert(!_target);freed++;[super dealloc];}
@end
static CADisplayLink *dc_make_link(id target){if(factory_nil)return nil;return (CADisplayLink*)[[[FakeDCLink alloc]initWithTarget:target]autorelease];}
static void dc_identity(long *window,unsigned int *display){*window=testWindow;*display=testDisplay;}
static int json_flags(void){
 char *raw=dc_snapshot();if(!raw)return 0;
 NSData *data=[NSData dataWithBytes:raw length:strlen(raw)];free(raw);
 NSDictionary *doc=[NSJSONSerialization JSONObjectWithData:data options:0 error:nil];
 for(NSString *key in @[@"configured",@"enabled",@"started",@"stopped",@"invalid",@"overflow"]){
  if(CFGetTypeID((CFTypeRef)doc[key])!=CFBooleanGetTypeID())return 0;
 }
 return 1;
}
static void reset(void){assert(!dc_link && !dc_target);memset(&dc,0,sizeof(dc));}
int main(void){@autoreleasepool {
 assert(dc_configure(0)==1 && dc_configure(1)==0);assert(dc_start()==1 && dc_start()==0);assert(!dc_link && !dc_target);assert(dc_stop()==1 && dc_stop()==0);assert(dc.count==0);if(!json_flags()){fputs("numeric boolean detected in native JSON\n",stderr);return 2;}reset();
 assert(dc_configure(1)==1);assert(dc_start()==1);for(int i=0;i<DC_CAP;i++)dc_append(10+i*.02,9+i*.02,9.01+i*.02,.02);assert(dc.count==DC_CAP && !dc.overflow && !dc.invalid);dc_append(10000,10000,10000.01,.02);assert(dc.overflow && dc.count==DC_CAP);assert(dc_stop()==1);assert(!dc_link && !dc_target);reset();
 assert(dc_configure(1)==1 && dc_start()==1);dc_append(1,NAN,2,.01);assert(dc.invalid && dc.count==0);assert(dc_stop()==1);reset();
 assert(dc_configure(1)==1 && dc_start()==1);dc_append(2,1,1.01,.01);dc_append(3,1,1.01,.01);assert(dc.invalid && dc.count==1);assert(dc_stop()==1);reset();
 assert(dc_configure(1)==1 && dc_start()==1);testWindow++;[(DCObserver*)dc_target tick:nil];assert(dc.invalid && dc.count==0);assert(dc_stop()==1);reset();
 assert(dc_configure(1)==1 && dc_start()==1);fake_timestamp=CACurrentMediaTime()-.016;[(DCObserver*)dc_target tick:dc_link];assert(dc.count==1 && !dc.invalid);assert(dc_stop()==1);assert(json_flags());char *s=dc_snapshot();assert(s && strlen(s)<(2u<<20));free(s);reset();
 testDisplay=0;assert(dc_configure(1)==1 && dc_start()==0);assert(dc.started && dc.invalid && !dc_link && !dc_target);assert(dc_stop()==1 && dc.stopped);reset();testDisplay=7;
 factory_nil=1;assert(dc_configure(1)==1 && dc_start()==0);assert(dc.started && dc.invalid && !dc_link && !dc_target);assert(dc_stop()==1 && dc.stopped);reset();factory_nil=0;
 }assert(made==freed && made==invalidated);puts("PASS off/on, single-use, capacity/no-eviction, invalid clocks/order, identity drift, native JSON boolean types, actual callback selector, start-failure cleanup, snapshot bound, invalidate/release ownership; no NSApplication or window created");return 0;}
