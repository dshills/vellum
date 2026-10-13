// Copyright (c) 2026, Davin Hills.
// SPDX-License-Identifier: MIT
#import <AppKit/AppKit.h>
#import <objc/runtime.h>
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <pthread.h>

// Silence refusal feedback in this headless fixture only. Production source
// remains unchanged; no system sound or preference is changed.
static void silentBeep(void) {}
#define NSBeep silentBeep
#ifndef VELLUM_INPUT_SOURCE
#define VELLUM_INPUT_SOURCE "input_darwin.m"
#endif
#include VELLUM_INPUT_SOURCE

bool goVellumManagedPreflight(uintptr_t a,uint64_t b,uint64_t c,char *d,char *e,char *f,int g,int h,int i,int j) {return true;}
bool goVellumManagedPreflightDelta(uintptr_t a,uint64_t b,uint64_t c,char *d,char *e,char *f,int g,int h,int i,int j,int k,int l) {return true;}
void goVellumManagedCommit(uintptr_t handle) {}

static int temporaries, factoryCalls, completed;
@interface InputPoolWitness : NSObject @end
@implementation InputPoolWitness
- (id)init { if((self=[super init]))temporaries++;return self; }
- (void)discardMarkedText {}
- (void)invalidateCharacterCoordinates {}
- (void)dealloc { temporaries--;[super dealloc]; }
@end
@interface PoolInputView : VellumManagedFixtureView @end
@implementation PoolInputView
- (NSTextInputContext *)inputContext { return (id)[[[InputPoolWitness alloc]init]autorelease]; }
@end
static IMP stringOriginal, arrayOriginal, dataOriginal;
static void witness(void) { if(pthread_main_np()){factoryCalls++;[[[InputPoolWitness alloc]init]autorelease];} }
static id stringFactory(id c,SEL s,const char *text) { id value=((id(*)(id,SEL,const char *))stringOriginal)(c,s,text);witness();return value; }
static id arrayFactory(id c,SEL s,const id *values,NSUInteger count) { id value=((id(*)(id,SEL,const id *,NSUInteger))arrayOriginal)(c,s,values,count);witness();return value; }
static id dataFactory(id c,SEL s,const void *bytes,NSUInteger count) { id value=((id(*)(id,SEL,const void *,NSUInteger))dataOriginal)(c,s,bytes,count);witness();return value; }
#define REQUIRE(condition,message) do { if(!(condition)){fprintf(stderr,"input ownership: %s temporary=%d\n",message,temporaries);return 1;} }while(0)
static int runChecks(void *view,void **owned) {
 *owned=vellum_managed_input_attach(view,1);
 REQUIRE(*owned && temporaries==0 && factoryCalls>0,"attach pool/refused ownership");
 VellumManagedInputState *s=*owned;
 for(int i=0;i<100;i++) {
  REQUIRE(vellum_managed_input_client_update(s,3,"ABCD",0,4,-1,-1,true,s->sequence,0,0,10,20),"valid text refused");
  REQUIRE(temporaries==0 && [s->text isEqualToString:@"ABCD"],"text not owned after pool drain");
  REQUIRE(!vellum_managed_input_client_update(s,3,"ABCD",0,99,-1,-1,true,s->sequence,0,0,10,20),"invalid text accepted");
  REQUIRE(temporaries==0 && s->selected.length==4,"early text refusal leaked or changed state");
  VellumTextCell cells[2]={{.start=1,.end=4,.line=0,.x=10,.y=0,.width=10,.height=20},{.start=0,.end=1,.line=0,.x=0,.y=0,.width=10,.height=20}};
  VellumTextCaret caret={.offset=4,.line=0,.x=20,.y=0,.height=20};
  REQUIRE(vellum_managed_input_client_geometry_update(s,3,"ABCD",0,4,-1,-1,true,s->sequence,0,0,10,20,cells,2,&caret,1,0,0,20,20,0),"valid geometry refused");
  REQUIRE(temporaries==0 && s->cells.length==sizeof(cells) && s->carets.length==sizeof(caret),"geometry not retained after pool drain");
  cells[0].end=99;caret.offset=99;
  REQUIRE(((const VellumTextCell *)s->cells.bytes)[0].start==0 && ((const VellumTextCaret *)s->carets.bytes)[0].offset==4,"caller edits changed copied geometry");
  REQUIRE(!vellum_managed_input_client_geometry_update(s,3,"ABCD",0,4,-1,-1,true,s->sequence,0,0,10,20,cells,2,&caret,1,0,0,20,20,0),"bad geometry accepted");
  REQUIRE(temporaries==0 && s->cells.length==sizeof(cells),"early geometry refusal leaked/retired accepted state");
  NSData *held=[s->cells retain];
  REQUIRE(vellum_managed_input_client_update(s,3,"ABCD",0,4,-1,-1,true,s->sequence,0,0,10,20),"geometry retirement publication refused");
  BOOL preserved=held.length==sizeof(cells) && ((const VellumTextCell *)held.bytes)[0].start==0;[held release];
  REQUIRE(preserved && !s->cells && !s->carets && temporaries==0,"caller ownership or state retirement failed");
  vellum_managed_input_caret(s,0,0,10,20);REQUIRE(temporaries==0,"caret pool missing");
  REQUIRE(vellum_managed_input_settle(s,3) && temporaries==0,"settle pool missing");
  REQUIRE(vellum_managed_input_client_update(s,3,"ABCD",0,4,-1,-1,false,s->sequence,0,0,10,20),"inactive publication refused");
  char *reply=vellum_managed_input_inactive_probe(s);
  BOOL copied=reply && strcmp(reply,"inactive_optional_super_queries_passed")==0;free(reply);
  REQUIRE(copied && temporaries==0,"malloc diagnostic reply did not outlive pool");
  completed++;
 }
 return 0;
}
static int check(void) {
 Method sm=class_getClassMethod(NSString.class,@selector(stringWithUTF8String:));
 Method am=class_getClassMethod(NSArray.class,@selector(arrayWithObjects:count:));
 Method dm=class_getClassMethod(NSData.class,@selector(dataWithBytes:length:));
 stringOriginal=method_getImplementation(sm);arrayOriginal=method_getImplementation(am);dataOriginal=method_getImplementation(dm);
 method_setImplementation(sm,(IMP)stringFactory);method_setImplementation(am,(IMP)arrayFactory);method_setImplementation(dm,(IMP)dataFactory);
 id view;@autoreleasepool {view=[[PoolInputView alloc]init];}void *input=NULL;int code=0;
 @try {code=runChecks(view,&input);}
 @finally {
  if(input)vellum_managed_input_close(input);
  if(!code && temporaries!=0)code=1;
  method_setImplementation(sm,stringOriginal);method_setImplementation(am,arrayOriginal);method_setImplementation(dm,dataOriginal);
  [view release];
 }
 if(!code && vellum_managed_input_fixture_views()!=0)code=1;
 printf("{\"fixture\":\"input_pool\",\"exit\":%d,\"temporaries\":%d,\"live_views\":%d,\"factory_calls\":%d,\"requested_cycles\":100,\"completed_cycles\":%d,\"windows_created\":0,\"input_sources_changed\":false}\n",code,temporaries,vellum_managed_input_fixture_views(),factoryCalls,completed);
 return code;
}
int main(void) {int code;@autoreleasepool {code=check();}return code;}
