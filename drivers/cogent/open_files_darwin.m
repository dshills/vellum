//go:build darwin && cgo
#import <Cocoa/Cocoa.h>
#import <CoreServices/CoreServices.h>
#include <pthread.h>
extern int vellumReceiveOpenFiles(uint64_t,const char *,int,int);
static pthread_mutex_t openFilesMu=PTHREAD_MUTEX_INITIALIZER;
static uint64_t openFilesToken;
static NSMutableArray<NSData *> *earlyFiles;
static NSUInteger earlyCount;
static AEEventHandlerUPP previousOpenHandler,installedOpenHandler;
static SRefCon previousOpenRef;
static BOOL hadPreviousOpen;
static id launchObserver;

static OSErr receiveOpenDocuments(const AppleEvent *event,AppleEvent *reply,SRefCon ref) {
 @autoreleasepool {
  if(AEGetDescDataSize(event)>128*1024)return paramErr;
  AppleEvent copy={typeNull,NULL};if(AEDuplicateDesc(event,&copy)!=noErr)return paramErr;
  NSAppleEventDescriptor *message=[[NSAppleEventDescriptor alloc] initWithAEDescNoCopy:&copy];
  NSAppleEventDescriptor *list=[message paramDescriptorForKeyword:keyDirectObject];
  NSInteger count=list.numberOfItems;if(count<1||count>16)return paramErr;
  NSMutableArray<NSString *> *paths=[NSMutableArray new];
  for(NSInteger index=1;index<=count;index++) {
   NSAppleEventDescriptor *item=[[list descriptorAtIndex:index] coerceToDescriptorType:typeFileURL];
   if(!item||item.data.length>16*1024)return paramErr;
   NSURL *url=[NSURL URLWithDataRepresentation:item.data relativeToURL:nil];
   if(!url.isFileURL||!url.path.isAbsolutePath||[url.path lengthOfBytesUsingEncoding:NSUTF8StringEncoding]>4096)return paramErr;
   [paths addObject:url.path];
  }
  NSData *json=[NSJSONSerialization dataWithJSONObject:paths options:0 error:nil];if(!json||json.length>64*1024)return paramErr;
  pthread_mutex_lock(&openFilesMu);uint64_t token=openFilesToken;
  if(!token) {
   if(earlyCount+paths.count>16){pthread_mutex_unlock(&openFilesMu);return errAEEventNotHandled;}
   if(!earlyFiles)earlyFiles=[NSMutableArray new];[earlyFiles addObject:json];earlyCount+=paths.count;
  }
  pthread_mutex_unlock(&openFilesMu);
  if(token&&!vellumReceiveOpenFiles(token,json.bytes,(int)json.length,0))return errAEEventNotHandled;
  return noErr;
 }
}
static BOOL installOpenHandler(void) {
 if(installedOpenHandler)return YES;
 hadPreviousOpen=AEGetEventHandler(kCoreEventClass,kAEOpenDocuments,&previousOpenHandler,&previousOpenRef,false)==noErr;
 installedOpenHandler=NewAEEventHandlerUPP(receiveOpenDocuments);
 if(AEInstallEventHandler(kCoreEventClass,kAEOpenDocuments,installedOpenHandler,0,false)!=noErr){DisposeAEEventHandlerUPP(installedOpenHandler);installedOpenHandler=NULL;return NO;}
 return YES;
}
// GLFW completes Cocoa launch during package initialization, before Driver.Run.
// Observe that public launch notification so opt-in packaged apps retain the
// cold-launch event. No Go callback is made until the Go receiver binds later.
__attribute__((constructor)) static void observeDocumentLaunch(void) {
 @autoreleasepool {
  launchObserver=[[NSNotificationCenter defaultCenter] addObserverForName:NSApplicationWillFinishLaunchingNotification object:nil queue:nil usingBlock:^(NSNotification *note){
   if([NSBundle.mainBundle.infoDictionary[@"VellumHandlesOpenDocuments"] boolValue])installOpenHandler();
   if(launchObserver){[[NSNotificationCenter defaultCenter] removeObserver:launchObserver];launchObserver=nil;}
  }];
 }
}
int vellumBindOpenFiles(uint64_t token) {
 @autoreleasepool {
  if(!installOpenHandler())return 0;
  pthread_mutex_lock(&openFilesMu);openFilesToken=token;NSArray<NSData *> *pending=[earlyFiles copy];earlyFiles=nil;earlyCount=0;pthread_mutex_unlock(&openFilesMu);
  for(NSData *json in pending)if(!vellumReceiveOpenFiles(token,json.bytes,(int)json.length,1))return 0;
  return 1;
 }
}
void vellumUnbindOpenFiles(uint64_t token) {
 @autoreleasepool {
  pthread_mutex_lock(&openFilesMu);if(openFilesToken!=token){pthread_mutex_unlock(&openFilesMu);return;}openFilesToken=0;earlyFiles=nil;earlyCount=0;pthread_mutex_unlock(&openFilesMu);
  AEEventHandlerUPP current=NULL;SRefCon ref=0;
  if(installedOpenHandler&&AEGetEventHandler(kCoreEventClass,kAEOpenDocuments,&current,&ref,false)==noErr&&current==installedOpenHandler){
   AERemoveEventHandler(kCoreEventClass,kAEOpenDocuments,installedOpenHandler,false);
   if(hadPreviousOpen)AEInstallEventHandler(kCoreEventClass,kAEOpenDocuments,previousOpenHandler,previousOpenRef,false);
  }
  if(installedOpenHandler)DisposeAEEventHandlerUPP(installedOpenHandler);installedOpenHandler=NULL;previousOpenHandler=NULL;
 }
}
