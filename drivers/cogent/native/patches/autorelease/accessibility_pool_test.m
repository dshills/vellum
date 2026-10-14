// Copyright (c) 2026, Davin Hills.
// SPDX-License-Identifier: MIT
#import <Cocoa/Cocoa.h>
#include <stdio.h>
#include <stdlib.h>
#ifndef VELLUM_AX_SOURCE
#define VELLUM_AX_SOURCE "../../../accessibility_darwin.m"
#endif
#include VELLUM_AX_SOURCE
int vellumAXAction(uint64_t a,uint64_t b,char *c,char *d,char *e){return 0;}
void vellumAXResult(uint64_t a,char *b){}
static int check(void){
 // Keep the enclosing pool alive throughout: teardown must own its drainage.
 NSObject *window=[NSObject new];NSView *view=[[NSView alloc]initWithFrame:NSZeroRect];
 Class original=object_getClass(view);
 for(uint64_t cycle=1;cycle<=100;cycle++){
  if(!vellumAXAttach(cycle,(__bridge void *)view,(__bridge void *)window))return 2;
  __weak VellumAXElement *observed;
  __weak VellumAXState *retired;
  @autoreleasepool{
   VellumAXState *state=vellumAXState(view);retired=state;
   for(int i=0;i<1000;i++){
    VellumAXElement *element=[VellumAXElement new];element.live=YES;element.state=state;
    element.node=@{@"id":[NSString stringWithFormat:@"node-%d",i],@"role":@"label",@"name":@"owned node payload"};
    state.elements[element.node[@"id"]]=element;if(i==0)observed=element;
   }
   state.roots=state.elements.allValues;
  }
  vellumAXDetach(cycle);
  if(observed||retired||vellumAXState(view)||vellumAXStates[@(cycle)]||object_getClass(view)!=original){
   fprintf(stderr,"native AX objects retained at cycle %llu before enclosing pool drain\n",(unsigned long long)cycle);return 1;
  }
  vellumAXDetach(cycle); // repeated/missing close is safe
 }
 if(vellumAXAttach(200,NULL,(__bridge void *)window))return 3;
 vellumAXDetach(UINT64_MAX);
 printf("{\"fixture\":\"ax_teardown_pool\",\"cycles\":100,\"nodes_per_cycle\":1000,\"window_created\":false,\"retained_before_outer_pool_drain\":0}\n");
 return 0;
}
int main(void){@autoreleasepool{return check();}}
