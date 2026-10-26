//go:build darwin && cgo
#import <Cocoa/Cocoa.h>
#import "services_native.h"
#include <string.h>
extern int vellumServiceActive(uint64_t);
extern void vellumServiceResult(uint64_t,char *,int);
extern void *vellumAXNativeWindow(uint64_t);
extern int vellumDocumentPanelBusy(void);
extern int vellumQuickLookBusy(void);
static NSDictionary *vellumPathRecipe(NSData *data){id o=[NSJSONSerialization JSONObjectWithData:data options:0 error:nil];if(![o isKindOfClass:NSDictionary.class]||!([o[@"Kind"] isEqual:@"file"]||[o[@"Kind"] isEqual:@"folder"]))return nil;for(NSString *k in @[@"Title",@"InitialPath"]){id v=o[k];if(![v isKindOfClass:NSString.class]||[v lengthOfBytesUsingEncoding:NSUTF8StringEncoding]>([k isEqual:@"Title"]?256:4096)||[v rangeOfString:[NSString stringWithFormat:@"%C",(unichar)0]].location!=NSNotFound)return nil;}NSString *initial=o[@"InitialPath"];if(initial.length&&!initial.isAbsolutePath)return nil;id extensions=o[@"Extensions"];if(extensions==NSNull.null)extensions=@[];if(![extensions isKindOfClass:NSArray.class]||[extensions count]>32||([o[@"Kind"] isEqual:@"folder"]&&[extensions count]))return nil;for(id e in extensions){if(![e isKindOfClass:NSString.class]||![e length]||[e length]>32||[e rangeOfCharacterFromSet:[NSCharacterSet characterSetWithCharactersInString:@". /\\\n\r"]].location!=NSNotFound)return nil;}return o;}
int vellumPathRecipeFixture(const char *json){@autoreleasepool{if(!json||strnlen(json,VELLUM_SERVICE_REQUEST_MAX+1)>VELLUM_SERVICE_REQUEST_MAX)return 0;@try{return vellumPathRecipe([NSData dataWithBytes:json length:strlen(json)])!=nil;}@catch(NSException *e){return 0;}}}
@interface VellumPathSession:NSObject
@property uint64_t token;
@property BOOL finished;
@property(strong)NSOpenPanel *panel;
@property(weak)NSWindow *parent;
@property(strong)id observer;
-(void)finish:(int)status value:(NSString *)value;
-(BOOL)cleanup;
@end
static NSMutableDictionary<NSNumber *,VellumPathSession *> *vellumPathSessions;
@implementation VellumPathSession
-(BOOL)cleanup{
 @try{if(self.observer)[NSNotificationCenter.defaultCenter removeObserver:self.observer];self.observer=nil;if(self.parent.attachedSheet==self.panel)[self.parent endSheet:self.panel returnCode:NSModalResponseCancel];[self.panel orderOut:nil];self.panel=nil;self.parent=nil;[vellumPathSessions removeObjectForKey:@(self.token)];return YES;}@catch(NSException *e){return NO;}
}
-(void)finish:(int)status value:(NSString *)value{
 if(self.finished)return;self.finished=YES;uint64_t token=self.token;
 if(![self cleanup])status=2;
 vellumServiceResult(token,(char *)(value?:@"").UTF8String,status);
}
@end
void vellumChoosePath(uint64_t token,uint64_t parentToken,const char *json){vellumWithServiceJSON(token,json,^(NSData *data){dispatch_async(dispatch_get_main_queue(),^{@autoreleasepool{@try{
 if(!vellumServiceActive(token)){vellumServiceResult(token,"",1);return;};NSDictionary *o=vellumPathRecipe(data);if(!o){vellumServiceResult(token,"",2);return;}
 NSWindow *parent=(__bridge NSWindow *)vellumAXNativeWindow(parentToken);if(!parent){vellumServiceResult(token,"",6);return;}if(!parent.isVisible||parent.isMiniaturized||!NSApp.isActive||NSApp.keyWindow!=parent){vellumServiceResult(token,"",2);return;}
 if(vellumPathSessions.count>=8||parent.attachedSheet||NSApp.modalWindow||vellumDocumentPanelBusy()||vellumQuickLookBusy()){vellumServiceResult(token,"",5);return;}
 NSOpenPanel *panel=[NSOpenPanel openPanel];BOOL folder=[o[@"Kind"] isEqual:@"folder"];panel.canChooseDirectories=folder;panel.canChooseFiles=!folder;panel.allowsMultipleSelection=NO;panel.resolvesAliases=YES;panel.canCreateDirectories=folder;panel.title=o[@"Title"];
 NSString *initial=o[@"InitialPath"];if(initial.length)panel.directoryURL=[NSURL fileURLWithPath:folder?initial:initial.stringByDeletingLastPathComponent isDirectory:YES];
 NSArray *extensions=o[@"Extensions"];if([extensions isKindOfClass:NSArray.class]&&extensions.count){
#pragma clang diagnostic push
#pragma clang diagnostic ignored "-Wdeprecated-declarations"
 panel.allowedFileTypes=extensions;
#pragma clang diagnostic pop
 }
 if(!vellumPathSessions)vellumPathSessions=[NSMutableDictionary new];VellumPathSession *s=[VellumPathSession new];s.token=token;s.panel=panel;s.parent=parent;vellumPathSessions[@(token)]=s;__weak VellumPathSession *weak=s;s.observer=[NSNotificationCenter.defaultCenter addObserverForName:NSWindowWillCloseNotification object:parent queue:nil usingBlock:^(NSNotification *note){[weak finish:6 value:nil];}];
 if(!vellumServiceActive(token)){[s finish:1 value:nil];return;}
 [panel beginSheetModalForWindow:parent completionHandler:^(NSModalResponse result){[s finish:result==NSModalResponseOK?0:1 value:result==NSModalResponseOK?panel.URL.path:nil];}];
 }@catch(NSException *e){VellumPathSession *s=vellumPathSessions[@(token)];if(s)[s finish:2 value:nil];else vellumServiceResult(token,"",2);}}});});}
void vellumCancelPath(uint64_t token){dispatch_async(dispatch_get_main_queue(),^{VellumPathSession *s=vellumPathSessions[@(token)];if(s)[s finish:1 value:nil];else vellumServiceResult(token,"",1);});}
int vellumCancelAllPaths(void){@try{for(VellumPathSession *s in vellumPathSessions.allValues.copy){if(s.finished)[s cleanup];else [s finish:1 value:nil];}return vellumPathSessions.count?2:0;}@catch(NSException *e){return 2;}}
void vellumRevealPath(uint64_t token,const char *path){@autoreleasepool{NSString *p=path?[NSString stringWithUTF8String:path]:nil;if(!p.isAbsolutePath||[p lengthOfBytesUsingEncoding:NSUTF8StringEncoding]>4096){vellumServiceResult(token,"",2);return;}NSURL *url=[NSURL fileURLWithPath:p];dispatch_async(dispatch_get_main_queue(),^{@try{if(!vellumServiceActive(token)){vellumServiceResult(token,"",1);return;}[NSWorkspace.sharedWorkspace activateFileViewerSelectingURLs:@[url]];vellumServiceResult(token,"",0);}@catch(NSException *e){vellumServiceResult(token,"",2);}});}}

@interface VellumPathPanelFixture:NSObject
@property int closes;
-(void)orderOut:(id)sender;
@end
@implementation VellumPathPanelFixture
-(void)orderOut:(id)sender{self.closes++;}
@end
int vellumPathCleanupFixture(uint64_t token){@autoreleasepool{
 if(vellumPathSessions.count)return 0;if(!vellumPathSessions)vellumPathSessions=[NSMutableDictionary new];
 VellumPathSession *s=[VellumPathSession new];VellumPathPanelFixture *p=[VellumPathPanelFixture new];s.token=token;s.panel=(NSOpenPanel *)p;vellumPathSessions[@(token)]=s;
 [s finish:1 value:nil];[s finish:0 value:@"ignored"];return s.finished&&!s.panel&&!s.observer&&!vellumPathSessions.count&&p.closes==1;
}}
