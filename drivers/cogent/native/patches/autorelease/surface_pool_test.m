// Copyright (c) 2026, Davin Hills.
// SPDX-License-Identifier: MIT
#import <Cocoa/Cocoa.h>
#import <QuartzCore/CAMetalLayer.h>
#import <objc/runtime.h>
#include <stdio.h>
#include <pthread.h>

static int temporaries, layers, creations, retirements;
@interface PoolWitness : NSObject @end
@implementation PoolWitness
- (id)init { if((self=[super init]))temporaries++;return self; }
- (void)dealloc { temporaries--;[super dealloc]; }
@end
// Counted layer stand-ins isolate reference ownership from CoreAnimation's
// deferred native cleanup. This fixture does not qualify actual layer resources.
@interface CountingLayer : NSObject @end
@implementation CountingLayer
- (id)init { if((self=[super init])){layers++;creations++;}return self; }
- (void)dealloc { layers--;retirements++;[super dealloc]; }
@end
static IMP factoryOriginal;
static id factory(id receiver,SEL selector) {
 if(!pthread_main_np())return ((id(*)(id,SEL))factoryOriginal)(receiver,selector);
 (void)receiver;(void)selector;
 [[[PoolWitness alloc]init]autorelease];
 return [[[CountingLayer alloc]init]autorelease];
}
// Faithful strong-property stand-ins. No NSWindow, NSApplication, display loop,
// input source, GPU device/drawable or desktop interaction is constructed.
@interface RetainingView : NSObject { id _layer; BOOL _wants; }
- (void)setWantsLayer:(BOOL)value;
- (void)setLayer:(id)value;
- (id)layer;
@end
@implementation RetainingView
- (void)setWantsLayer:(BOOL)value { _wants=value; }
- (void)setLayer:(id)value { [value retain];[_layer release];_layer=value; }
- (id)layer { return _layer; }
- (void)dealloc { [_layer release];[super dealloc]; }
@end
@interface RetainingWindow : NSObject { RetainingView *_view; }
- (RetainingView *)contentView;
@end
@implementation RetainingWindow
- (id)init { if((self=[super init]))_view=[[RetainingView alloc]init];return self; }
- (RetainingView *)contentView { return _view; }
- (void)dealloc { [_view release];[super dealloc]; }
@end
@interface EmptyWindow : NSObject
- (id)contentView;
@end
@implementation EmptyWindow
- (id)contentView { return nil; }
@end
#ifndef VELLUM_LAYER_SOURCE
#define VELLUM_LAYER_SOURCE "surface_layer_darwin.h"
#endif
#include VELLUM_LAYER_SOURCE

// Link the actual owned layer override; these callbacks never record data in
// this separate property-ownership fixture.
uint64_t vellumGoDrawableAcquired(uint64_t a,uint64_t b,uint64_t c,uint64_t d,int e,double f) { (void)a;(void)b;(void)c;(void)d;(void)e;(void)f;return 0; }
void vellumGoDrawableUnavailable(uint64_t a,uint64_t b,uint64_t c,uint64_t d,int e,double f) { (void)a;(void)b;(void)c;(void)d;(void)e;(void)f; }
void vellumGoDrawablePresented(uint64_t a,uint64_t b,uint64_t c,double d) { (void)a;(void)b;(void)c;(void)d; }
void vellumGoPresentationReceipt(uint64_t a,uint64_t b,double c) { (void)a;(void)b;(void)c; }
static int check(void) {
 Method method=class_getClassMethod(CAMetalLayer.class,@selector(layer));
 IMP original=method_getImplementation(method);factoryOriginal=original;
 method_setImplementation(method,(IMP)factory);
 RetainingWindow *window=[[RetainingWindow alloc]init];
 id held=nil;
 int code=0;BOOL nilCases=NO;
 @try {
  id empty=[[EmptyWindow alloc]init];
  void *nilWindow=vellumMetalLayerFromWindow(NULL);
  void *nilView=vellumMetalLayerFromWindow(empty);
  nilCases=nilWindow==NULL && nilView==NULL && creations==0 && temporaries==0;
  [empty release];
  if(!nilCases){fprintf(stderr,"nil window/view returned an unowned layer\n");code=1;}
  for(int i=0;!code && i<64;i++) {
   id value=(id)vellumMetalLayerFromWindow(window);
   if(!value || value!=window.contentView.layer || temporaries!=0 || layers!=1+(held!=nil)) {
    fprintf(stderr,"layer factory failed scoped ownership at iteration %d: temporary=%d layers=%d\n",i,temporaries,layers);code=1;break;
   }
   if(i==0)held=[value retain];
   if(i==1){[held release];held=nil;if(layers!=1){code=1;break;}}
  }
  [window.contentView setLayer:nil];
  if(!code && (layers!=0 || temporaries!=0 || creations!=64 || retirements!=64))code=1;
 } @finally {
  method_setImplementation(method,original);
  [held release];[window release];
 }
 printf("{\"fixture\":\"surface_pool\",\"exit\":%d,\"temporaries\":%d,\"live_layer_standins\":%d,\"created\":%d,\"retired\":%d,\"nil_window_and_view_passed\":%s,\"windows_created\":0,\"gpu_drawables_acquired\":0}\n",code,temporaries,layers,creations,retirements,nilCases ? "true":"false");
 return code;
}
int main(void) { int code;@autoreleasepool { code=check(); }return code; }
