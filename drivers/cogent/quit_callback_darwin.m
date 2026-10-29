//go:build darwin && cgo
#import <Cocoa/Cocoa.h>
#import <objc/runtime.h>
#include <pthread.h>
#include <string.h>
extern void vellumNativeQuitCallback(uint64_t);
static pthread_mutex_t quitMu=PTHREAD_MUTEX_INITIALIZER;
static uint64_t quitToken;
static Method quitMethod;
static IMP previousQuit;
static NSApplicationTerminateReply routeQuit(id self,SEL selector,NSApplication *sender) {
 pthread_mutex_lock(&quitMu);
 uint64_t token=quitToken; IMP previous=previousQuit;
 pthread_mutex_unlock(&quitMu);
 if(token){vellumNativeQuitCallback(token);return NSTerminateCancel;}
 if(previous)return ((NSApplicationTerminateReply(*)(id,SEL,NSApplication *))previous)(self,selector,sender);
 return NSTerminateCancel;
}
static int bindDelegate(id delegate,uint64_t token) {
 Method method=class_getInstanceMethod([delegate class],@selector(applicationShouldTerminate:));
 if(!method||method_getNumberOfArguments(method)!=3||!token)return 0;
 char *result=method_copyReturnType(method);
 BOOL valid=result&&strcmp(result,@encode(NSApplicationTerminateReply))==0;
 free(result);if(!valid)return 0;
 pthread_mutex_lock(&quitMu);
 if(quitToken){pthread_mutex_unlock(&quitMu);return 0;}
 quitMethod=method;previousQuit=method_getImplementation(method);quitToken=token;
 method_setImplementation(method,(IMP)routeQuit);
 pthread_mutex_unlock(&quitMu);return 1;
}
int vellumBindQuit(uint64_t token) { return bindDelegate(NSApp.delegate,token); }
void vellumUnbindQuit(uint64_t token) {
 pthread_mutex_lock(&quitMu);
 if(quitToken==token){
  if(quitMethod&&method_getImplementation(quitMethod)==(IMP)routeQuit)method_setImplementation(quitMethod,previousQuit);
  quitToken=0;quitMethod=NULL;previousQuit=NULL;
 }
 pthread_mutex_unlock(&quitMu);
}
// Inert fixture exercises selector interception/restoration without NSApplication,
// windows, input events, settings or an AppKit event loop.
@interface VellumQuitFixture:NSObject
- (NSApplicationTerminateReply)applicationShouldTerminate:(NSApplication *)sender;
@end
@implementation VellumQuitFixture
- (NSApplicationTerminateReply)applicationShouldTerminate:(NSApplication *)sender { return NSTerminateNow; }
@end
int vellumTestQuitDelegate(void) {
 @autoreleasepool {
  VellumQuitFixture *fixture=[VellumQuitFixture new];
  if([fixture applicationShouldTerminate:nil]!=NSTerminateNow||!bindDelegate(fixture,UINT64_MAX))return 0;
  BOOL intercepted=[fixture applicationShouldTerminate:nil]==NSTerminateCancel;
  vellumUnbindQuit(UINT64_MAX);
  return intercepted&&[fixture applicationShouldTerminate:nil]==NSTerminateNow;
 }
}
