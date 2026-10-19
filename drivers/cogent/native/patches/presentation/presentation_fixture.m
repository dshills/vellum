// Copyright (c) 2026, Davin Hills. SPDX-License-Identifier: MIT
// Controlled drawable stand-ins. No NSApplication/window/device/queue is used.
#import <Foundation/Foundation.h>
#import <objc/runtime.h>
#include <Block.h>
#include <stdio.h>
#ifndef VELLUM_PRESENTATION_SOURCE
#define VELLUM_PRESENTATION_SOURCE "presentation_darwin.m"
#endif
#include VELLUM_PRESENTATION_SOURCE

static unsigned acquisitions, unavailable, completions, receipts, calls, registrations;
static uint64_t lastToken,lastTag,lastGeneration,lastComposition,lastSequence,lastDrawable;
static int lastArmed, refuse;
static double lastTime;
uint64_t vellumGoDrawableAcquired(uint64_t token,uint64_t tag,uint64_t generation,uint64_t composition,int armed,double at) {
 acquisitions++;lastToken=token;lastTag=tag;lastGeneration=generation;lastComposition=composition;lastArmed=armed;lastTime=at;
 return refuse?0:acquisitions;
}
void vellumGoDrawableUnavailable(uint64_t token,uint64_t tag,uint64_t generation,uint64_t composition,int armed,double at) {
 unavailable++;lastToken=token;lastTag=tag;lastGeneration=generation;lastComposition=composition;lastArmed=armed;lastTime=at;
}
void vellumGoDrawablePresented(uint64_t token,uint64_t sequence,uint64_t drawable,double at) {
 completions++;lastToken=token;lastSequence=sequence;lastDrawable=drawable;lastTime=at;
}
void vellumGoPresentationReceipt(uint64_t token,uint64_t tag,double at) { receipts++;lastToken=token;lastTag=tag;lastTime=at; }

@interface FakeDrawable : NSObject {
 MTLDrawablePresentedHandler _handler;
}
@property(nonatomic) uint64_t drawableID;
@property(nonatomic) double presentedTime;
- (void)addPresentedHandler:(MTLDrawablePresentedHandler)handler;
- (void)fire;
@end
@implementation FakeDrawable
- (void)addPresentedHandler:(MTLDrawablePresentedHandler)handler { registrations++;_handler=Block_copy(handler); }
- (void)fire { if(_handler)_handler((id<MTLDrawable>)self); }
- (void)dealloc { if(_handler)Block_release(_handler);[super dealloc]; }
@end
static id supplied;
static id fakeNext(id receiver,SEL selector) { (void)receiver;(void)selector;calls++;return supplied; }
#define REQUIRE(condition) do { if(!(condition)){fprintf(stderr,"fixture assertion at line %d\n",__LINE__);return 1;} } while(0)
static int check(void) {
 Method method=class_getInstanceMethod(CAMetalLayer.class,@selector(nextDrawable));
 IMP original=method_setImplementation(method,(IMP)fakeNext);
 int code=0;
 FakeDrawable *first=[[FakeDrawable alloc]init];
 FakeDrawable *second=[[FakeDrawable alloc]init];
 @try {
  @autoreleasepool {
   void *layer=vellumNewPresentationLayer();
   REQUIRE(layer && [(id)layer isKindOfClass:VellumPresentationLayer.class]);
   supplied=first;
   REQUIRE([(CAMetalLayer *)layer nextDrawable]==(id)first && calls==1 && acquisitions==0 && registrations==0);
   vellumPresentationObserver(layer,7);
   vellumPresentationReceipt(layer,11);
   REQUIRE(receipts==1 && lastToken==7 && lastTag==11 && lastTime>0);
   vellumPresentationFrame(layer,11,13,1,0);
   REQUIRE([(CAMetalLayer *)layer nextDrawable]==(id)first && acquisitions==1 && registrations==1);
   REQUIRE(lastToken==7 && lastTag==11 && lastGeneration==13 && lastComposition==1 && !lastArmed && lastTime>0);
   vellumPresentationFrame(layer,11,17,2,1);
   supplied=second;
   REQUIRE([(CAMetalLayer *)layer nextDrawable]==(id)second && acquisitions==2 && registrations==2);
   REQUIRE(lastGeneration==17 && lastComposition==2 && lastArmed);
   supplied=nil;
   REQUIRE([(CAMetalLayer *)layer nextDrawable]==nil && unavailable==1 && acquisitions==2 && registrations==2);
   REQUIRE(lastToken==7 && lastComposition==2 && lastArmed && lastTime>0);
   refuse=1;supplied=second;
   REQUIRE([(CAMetalLayer *)layer nextDrawable]==(id)second && acquisitions==3 && registrations==2);
   vellumPresentationObserver(layer,0);
   vellumPresentationReceipt(layer,99);
   REQUIRE(receipts==1);
   REQUIRE([(CAMetalLayer *)layer nextDrawable]==(id)second && acquisitions==3 && registrations==2 && calls==6);
  }
  // No layer/window reference is needed to deliver the handler later.
  second.drawableID=99;second.presentedTime=20;
  [second fire];
  REQUIRE(completions==1 && lastToken==7 && lastSequence==2 && lastDrawable==99 && lastTime==20);
  first.drawableID=0;first.presentedTime=0;
  [first fire];
  REQUIRE(completions==2 && lastToken==7 && lastSequence==1 && lastDrawable==0 && lastTime==0);
 } @finally {
  method_setImplementation(method,original);supplied=nil;[first release];[second release];
 }
 printf("{\"fixture\":\"presentation\",\"super_acquires\":%u,\"tracked\":%u,\"nil_drawables\":%u,\"registered\":%u,\"late_completions\":%u,\"windows_created\":0,\"real_gpu_drawables\":0}\n",calls,acquisitions,unavailable,registrations,completions);
 return code;
}
int main(void) { @autoreleasepool { return check(); } }
