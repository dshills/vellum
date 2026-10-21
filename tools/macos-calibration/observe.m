// Copyright (c) 2026 Davin Hills. SPDX-License-Identifier: MIT.
#import <Cocoa/Cocoa.h>
#import <QuartzCore/CAMetalLayer.h>
#import <Metal/Metal.h>
#include <stdatomic.h>
#include <unistd.h>
#include "observe.h"
#import "flags.h"
static _Atomic(int) losses;
static NSArray *tokens;
static id workspaceToken;
static CAMetalLayer *metalLayer(CALayer *root,int depth) {
 if(!root || depth>16)return nil;
 if([root isKindOfClass:CAMetalLayer.class])return (CAMetalLayer*)root;
 for(CALayer *child in root.sublayers){CAMetalLayer *found=metalLayer(child,depth+1);if(found)return found;}return nil;
}
static CAMetalLayer *viewLayer(NSView *view,int depth){
 if(!view || depth>16)return nil;CAMetalLayer *layer=metalLayer(view.layer,0);if(layer)return layer;
 for(NSView *child in view.subviews){layer=viewLayer(child,depth+1);if(layer)return layer;}return nil;
}
double cal_clock(void){return CACurrentMediaTime();}
char *cal_observe(int phase) {
 __block char *out=NULL;
 void (^work)(void)=^{
  NSWindow *w=NSApp.keyWindow;CAMetalLayer *layer=viewLayer(w.contentView,0);
  if(phase==1 && !tokens && w && w.keyWindow && NSApp.active && (w.occlusionState&NSWindowOcclusionStateVisible) && NSWorkspace.sharedWorkspace.frontmostApplication.processIdentifier==getpid()){
   NSNotificationCenter *c=NSNotificationCenter.defaultCenter;
   id a=[c addObserverForName:NSApplicationDidResignActiveNotification object:NSApp queue:NSOperationQueue.mainQueue usingBlock:^(NSNotification *n){(void)n;atomic_fetch_add(&losses,1);}];
   id k=[c addObserverForName:NSWindowDidResignKeyNotification object:w queue:NSOperationQueue.mainQueue usingBlock:^(NSNotification *n){(void)n;atomic_fetch_add(&losses,1);}];tokens=[[NSArray alloc] initWithObjects:a,k,nil];
   workspaceToken=[NSWorkspace.sharedWorkspace.notificationCenter addObserverForName:NSWorkspaceDidActivateApplicationNotification object:nil queue:NSOperationQueue.mainQueue usingBlock:^(NSNotification *n){NSRunningApplication *app=n.userInfo[NSWorkspaceApplicationKey];if(app.processIdentifier!=getpid())atomic_fetch_add(&losses,1);}];
  }
#if !__has_feature(objc_arc)
  if(phase==1 && workspaceToken)[workspaceToken retain];
#endif
  NSDictionary *d=@{@"clock":@(cal_clock()),@"pid":@(getpid()),@"front_pid":@(NSWorkspace.sharedWorkspace.frontmostApplication.processIdentifier),@"active":cal_flag(NSApp.active),@"key":cal_flag(w.keyWindow),@"visible":cal_flag((w.occlusionState&NSWindowOcclusionStateVisible)!=0),@"losses":@(atomic_load(&losses)),@"width":@(layer.drawableSize.width),@"height":@(layer.drawableSize.height),@"scale":@(w.backingScaleFactor),@"pixel_format":@(layer.pixelFormat),@"drawables":@(layer.maximumDrawableCount),@"vsync":cal_flag(layer.displaySyncEnabled),@"transactions":cal_flag(layer.presentsWithTransaction),@"timeout":cal_flag(layer.allowsNextDrawableTimeout)};
  NSData *data=[NSJSONSerialization dataWithJSONObject:d options:NSJSONWritingSortedKeys error:nil];out=strndup(data.bytes,data.length);
  if(phase==2){for(id token in tokens)[NSNotificationCenter.defaultCenter removeObserver:token];if(workspaceToken)[NSWorkspace.sharedWorkspace.notificationCenter removeObserver:workspaceToken];
#if !__has_feature(objc_arc)
  [tokens release];[workspaceToken release];
#endif
  tokens=nil;workspaceToken=nil;}

 };
 if(NSThread.isMainThread)work();else dispatch_sync(dispatch_get_main_queue(),work);return out;
}
