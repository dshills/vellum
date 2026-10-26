//go:build darwin && cgo
#import <Cocoa/Cocoa.h>
#import "services_native.h"
#include <string.h>
extern int vellumServiceActive(uint64_t);
extern void vellumServiceResult(uint64_t,char *,int);
extern void *vellumAXNativeWindow(uint64_t);
extern BOOL vellumAXShareAnchor(uint64_t,NSString *,NSView **,NSRect *);
extern int vellumDocumentPanelBusy(void);
extern int vellumQuickLookBusy(void);
static NSDictionary *vellumShareRecipe(NSData *data){
 id o=[NSJSONSerialization JSONObjectWithData:data options:0 error:nil];
 if(![o isKindOfClass:NSDictionary.class]||![o[@"AnchorID"] isKindOfClass:NSString.class]||![o[@"AnchorID"] length]||[o[@"AnchorID"] length]>256||![o[@"Items"] isKindOfClass:NSArray.class]||[o[@"Items"] count]<1||[o[@"Items"] count]>16)return nil;
 if([o[@"AnchorID"] rangeOfString:[NSString stringWithFormat:@"%C",(unichar)0]].location!=NSNotFound)return nil;
 NSUInteger total=0;
 for(id item in o[@"Items"]){
  if(![item isKindOfClass:NSDictionary.class]||![item[@"Value"] isKindOfClass:NSString.class]||![item[@"Kind"] isKindOfClass:NSString.class])return nil;
  NSString *v=item[@"Value"],*kind=item[@"Kind"];
  total+=[v lengthOfBytesUsingEncoding:NSUTF8StringEncoding];if(total>65536||![v stringByTrimmingCharactersInSet:NSCharacterSet.whitespaceAndNewlineCharacterSet].length||[v rangeOfString:[NSString stringWithFormat:@"%C",(unichar)0]].location!=NSNotFound)return nil;
  if([kind isEqual:@"file"]){if(!v.isAbsolutePath||[v lengthOfBytesUsingEncoding:NSUTF8StringEncoding]>4096)return nil;}
  else if([kind isEqual:@"url"]){NSURL *u=[NSURL URLWithString:v];if(!u.host.length||u.user||u.password||!([u.scheme isEqual:@"http"]||[u.scheme isEqual:@"https"]))return nil;}
  else if(![kind isEqual:@"text"])return nil;
 }
 return o;
}
int vellumShareRecipeFixture(const char *json){@autoreleasepool{@try{if(!json||strnlen(json,VELLUM_SERVICE_REQUEST_MAX+1)>VELLUM_SERVICE_REQUEST_MAX)return 0;return vellumShareRecipe([NSData dataWithBytes:json length:strlen(json)])!=nil;}@catch(NSException *e){return 0;}}}
@interface VellumShareSession:NSObject<NSSharingServicePickerDelegate>
@property uint64_t token,parentToken;
@property BOOL finished,selected;
@property(strong)NSSharingServicePicker *picker;
@property(weak)NSWindow *parent;
@property(copy)NSString *anchor,*choice;
@property NSRect anchorRect;
@property(strong)id closeObserver;
-(void)finish:(int)status;
-(BOOL)cleanup;
@end
static VellumShareSession *vellumShareSession;
@implementation VellumShareSession
-(BOOL)cleanup{
 @try{
  self.picker.delegate=nil;
  if(!self.selected){if(@available(macOS 13.0,*)){[self.picker close];}}
  if(self.closeObserver)[NSNotificationCenter.defaultCenter removeObserver:self.closeObserver];
  self.closeObserver=nil;self.picker=nil;self.parent=nil;self.anchor=nil;
  if(vellumShareSession==self)vellumShareSession=nil;
  return YES;
 }@catch(NSException *e){return NO;}
}
-(void)finish:(int)status{
 if(self.finished)return;self.finished=YES;
 uint64_t token=self.token;
 NSString *choice=self.choice?:@"";
 if(![self cleanup])status=2;
 vellumServiceResult(token,(char *)choice.UTF8String,status);
}
-(void)sharingServicePicker:(NSSharingServicePicker *)picker didChooseSharingService:(NSSharingService *)service{
 if(self.finished)return;
 self.selected=service!=nil;self.choice=service.title;
 // The picker hands items to the selected service after this delegate returns.
 // Release only on the next main turn; never close a committed service UI.
 dispatch_async(dispatch_get_main_queue(),^{[self finish:self.selected?0:1];});
}
@end
void vellumShareAnchorChanged(uint64_t parent){
 VellumShareSession *s=vellumShareSession;if(!s||s.finished||s.selected||s.parentToken!=parent)return;
 NSView *view=nil;NSRect rect=NSZeroRect;
 if(!vellumAXShareAnchor(parent,s.anchor,&view,&rect)||!NSEqualRects(rect,s.anchorRect))[s finish:1];
}
void vellumShare(uint64_t token,uint64_t parentToken,const char *json){
 vellumWithServiceJSON(token,json,^(NSData *data){dispatch_async(dispatch_get_main_queue(),^{@autoreleasepool{@try{
  if(!vellumServiceActive(token)){vellumServiceResult(token,"",1);return;}
  if(@available(macOS 13.0,*)){}else{vellumServiceResult(token,"",2);return;}
  NSDictionary *o=vellumShareRecipe(data);if(!o){vellumServiceResult(token,"",2);return;}
  NSWindow *parent=(__bridge NSWindow *)vellumAXNativeWindow(parentToken);
  if(!parent){vellumServiceResult(token,"",6);return;}
  if(!parent.isVisible||parent.isMiniaturized||!NSApp.isActive||NSApp.keyWindow!=parent){vellumServiceResult(token,"",2);return;}
  if(vellumShareSession||vellumDocumentPanelBusy()||vellumQuickLookBusy()||parent.attachedSheet||NSApp.modalWindow){vellumServiceResult(token,"",5);return;}
  NSView *view=nil;NSRect anchor=NSZeroRect;
  if(!vellumAXShareAnchor(parentToken,o[@"AnchorID"],&view,&anchor)){vellumServiceResult(token,"",2);return;}
  NSMutableArray *items=[NSMutableArray new];for(NSDictionary *v in o[@"Items"]){NSString *kind=v[@"Kind"],*value=v[@"Value"];[items addObject:[kind isEqual:@"file"]?[NSURL fileURLWithPath:value]:[kind isEqual:@"url"]?[NSURL URLWithString:value]:value];}
  VellumShareSession *s=[VellumShareSession new];s.token=token;s.parentToken=parentToken;s.parent=parent;s.anchor=o[@"AnchorID"];s.anchorRect=anchor;
  s.picker=[[NSSharingServicePicker alloc]initWithItems:items];s.picker.delegate=s;vellumShareSession=s;
  __weak VellumShareSession *weak=s;
  s.closeObserver=[NSNotificationCenter.defaultCenter addObserverForName:NSWindowWillCloseNotification object:parent queue:nil usingBlock:^(NSNotification *note){[weak finish:6];}];
  if(!vellumServiceActive(token)){[s finish:1];return;}
  [s.picker showRelativeToRect:anchor ofView:view preferredEdge:NSMinYEdge];
 }@catch(NSException *e){if(vellumShareSession.token==token)[vellumShareSession finish:2];else vellumServiceResult(token,"",2);}}});});
}
void vellumCancelShare(uint64_t token){dispatch_async(dispatch_get_main_queue(),^{if(vellumShareSession.token==token)[vellumShareSession finish:1];else vellumServiceResult(token,"",1);});}
int vellumCancelAllShares(void){@try{if(vellumShareSession.finished){if(![vellumShareSession cleanup])return 2;}else [vellumShareSession finish:1];return vellumShareSession?2:0;}@catch(NSException *e){return 2;}}
// Exercise the real session cleanup and native terminal callback without an OS
// picker, window, item transfer, sharing service, timer or main-loop dispatch.
@interface VellumSharePickerFixture:NSObject
@property(weak)id delegate;
@property int closes;
-(void)close;
@end
@implementation VellumSharePickerFixture
-(void)close{self.closes++;}
@end
int vellumShareCleanupFixture(uint64_t token,int selected){
 @autoreleasepool{
  if(vellumShareSession)return 0;
  VellumSharePickerFixture *p=[VellumSharePickerFixture new];VellumShareSession *s=[VellumShareSession new];s.token=token;s.selected=selected!=0;s.picker=(NSSharingServicePicker *)p;p.delegate=s;vellumShareSession=s;
  [s finish:1];[s finish:0];return s.finished&&!s.picker&&!vellumShareSession&&!p.delegate&&p.closes==(selected?0:1);
 }
}
