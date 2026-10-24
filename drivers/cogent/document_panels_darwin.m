//go:build darwin && cgo
#import <Cocoa/Cocoa.h>
#include <string.h>
#include <stdlib.h>
#include <math.h>
#import "services_native.h"
extern int vellumServiceActive(uint64_t token);
extern void vellumServiceResult(uint64_t token,char *text,int status);
extern void vellumDocumentPanelPreview(uint64_t token,char *text);
extern void *vellumAXNativeWindow(uint64_t token);
@class VellumDocumentPanelSession;
static NSMutableDictionary<NSNumber *,VellumDocumentPanelSession *> *vellumDocumentPanels;
static BOOL vellumOwnsColorPanel;
// Foundation-only bounded parser: never creates AppKit panels in offscreen tests.
static NSDictionary *vellumDocumentPanelRecipe(NSData *data){
 NSDictionary *request=[NSJSONSerialization JSONObjectWithData:data options:0 error:nil];if(![request isKindOfClass:NSDictionary.class]||![request[@"Kind"] isEqual:@"color"])return nil;
 NSDictionary *o=request[@"Options"];if(![o isKindOfClass:NSDictionary.class]||![o[@"Title"] isKindOfClass:NSString.class]||![o[@"Title"] length]||[o[@"Title"] length]>256||![o[@"ShowsAlpha"] isKindOfClass:NSNumber.class])return nil;
 NSDictionary *c=o[@"Color"];if(![c isKindOfClass:NSDictionary.class])return nil;for(NSString *key in @[@"R",@"G",@"B",@"A"]){id v=c[key];if(![v isKindOfClass:NSNumber.class]||!isfinite([v doubleValue])||[v doubleValue]<0||[v doubleValue]>255||floor([v doubleValue])!=[v doubleValue])return nil;}return request;
}
static NSString *vellumPanelJSON(NSDictionary *value){NSData *data=value?[NSJSONSerialization dataWithJSONObject:value options:0 error:nil]:nil;return data?[[NSString alloc]initWithData:data encoding:NSUTF8StringEncoding]:nil;}
char *vellumDocumentPanelRecipeFixture(const char *json){@autoreleasepool{@try{if(!json||strnlen(json,VELLUM_SERVICE_REQUEST_MAX+1)>VELLUM_SERVICE_REQUEST_MAX)return NULL;NSDictionary *recipe=vellumDocumentPanelRecipe([NSData dataWithBytes:json length:strlen(json)]);NSString *result=vellumPanelJSON(recipe);return result?strdup(result.UTF8String):NULL;}@catch(NSException *e){return NULL;}}}
@interface VellumDocumentPanelSession:NSObject<NSWindowDelegate>
@property uint64_t token;
@property BOOL finished;
@property BOOL alpha;
@property(strong) NSColorPanel *panel;
@property(weak) NSWindow *parent;
@property(weak) id<NSWindowDelegate> savedDelegate;
@property(strong) NSView *savedAccessory;
@property(copy) NSString *savedTitle;
@property(copy) NSColor *savedColor;
@property BOOL savedContinuous,savedAlpha,savedHides;
@property(strong) id parentObserver;
-(void)finish:(int)status;
-(int)cleanup;
@end
@implementation VellumDocumentPanelSession
-(NSDictionary *)value{NSColor *c=[self.panel.color colorUsingColorSpace:NSColorSpace.sRGBColorSpace];if(!c)return nil;CGFloat red=c.redComponent,green=c.greenComponent,blue=c.blueComponent,alpha=self.alpha?c.alphaComponent:1;if(!isfinite(red)||!isfinite(green)||!isfinite(blue)||!isfinite(alpha))return nil;return @{@"R":@(lround(fmax(0,fmin(1,red))*255)),@"G":@(lround(fmax(0,fmin(1,green))*255)),@"B":@(lround(fmax(0,fmin(1,blue))*255)),@"A":@(lround(fmax(0,fmin(1,alpha))*255))};}
-(void)changed:(id)sender{@try{if(self.finished||!vellumServiceActive(self.token))return;NSString *wire=vellumPanelJSON([self value]);if(wire)vellumDocumentPanelPreview(self.token,(char *)wire.UTF8String);}@catch(NSException *e){[self finish:2];}}
-(void)apply:(id)sender{[self finish:0];}
-(void)cancel:(id)sender{[self finish:1];}
-(BOOL)windowShouldClose:(NSWindow *)sender{[self finish:1];return NO;}
-(int)cleanup{int result=0;
 @try{if(self.parentObserver){[NSNotificationCenter.defaultCenter removeObserver:self.parentObserver];self.parentObserver=nil;}}@catch(NSException *e){result=2;}
 @try{[self.parent removeChildWindow:self.panel];[self.panel orderOut:nil];}@catch(NSException *e){result=2;}
 @try{[self.panel setTarget:nil];[self.panel setAction:NULL];self.panel.delegate=self.savedDelegate;self.panel.accessoryView=self.savedAccessory;self.panel.title=self.savedTitle?:@"Colors";self.panel.continuous=self.savedContinuous;self.panel.showsAlpha=self.savedAlpha;self.panel.hidesOnDeactivate=self.savedHides;if(self.savedColor)self.panel.color=self.savedColor;}@catch(NSException *e){result=2;}
 if(!result){[vellumDocumentPanels removeObjectForKey:@(self.token)];self.panel=nil;self.savedAccessory=nil;self.savedColor=nil;self.parent=nil;}return result;
}
-(void)finish:(int)status{if(self.finished)return;self.finished=YES;NSString *wire=nil;@try{if(status==0){wire=vellumPanelJSON([self value]);if(!wire)status=2;}}@catch(NSException *e){status=2;}if([self cleanup])status=2;vellumServiceResult(self.token,(char *)(wire?:@"").UTF8String,status);}
@end
static NSView *vellumColorActions(VellumDocumentPanelSession *session){NSView *view=[[NSView alloc]initWithFrame:NSMakeRect(0,0,300,42)];NSButton *cancel=[NSButton buttonWithTitle:@"Cancel" target:session action:@selector(cancel:)];cancel.frame=NSMakeRect(106,8,86,28);cancel.bezelStyle=NSBezelStyleRounded;cancel.keyEquivalent=@"\033";NSButton *apply=[NSButton buttonWithTitle:@"Use Color" target:session action:@selector(apply:)];apply.frame=NSMakeRect(198,8,96,28);apply.bezelStyle=NSBezelStyleRounded;apply.keyEquivalent=@"\r";[view addSubview:cancel];[view addSubview:apply];return view;}
void vellumChooseDocumentPanel(uint64_t token,uint64_t parentToken,const char *json){vellumWithServiceJSON(token,json,^(NSData *data){dispatch_async(dispatch_get_main_queue(),^{@autoreleasepool{@try{
 if(!vellumServiceActive(token)){vellumServiceResult(token,"",1);return;}NSDictionary *recipe=vellumDocumentPanelRecipe(data);if(!recipe){vellumServiceResult(token,"",2);return;}
 NSWindow *parent=(__bridge NSWindow *)vellumAXNativeWindow(parentToken);if(!parent){vellumServiceResult(token,"",6);return;}if(!parent.isVisible||parent.isMiniaturized){vellumServiceResult(token,"",2);return;}
 if(vellumDocumentPanels.count||parent.attachedSheet||NSApp.modalWindow||(!vellumOwnsColorPanel&&NSColorPanel.sharedColorPanelExists)){vellumServiceResult(token,"",5);return;}
 NSColorPanel *panel=NSColorPanel.sharedColorPanel;vellumOwnsColorPanel=YES;if(panel.isVisible){vellumServiceResult(token,"",5);return;}
 NSDictionary *o=recipe[@"Options"];VellumDocumentPanelSession *session=[VellumDocumentPanelSession new];session.token=token;session.parent=parent;session.panel=panel;session.alpha=[o[@"ShowsAlpha"] boolValue];session.savedDelegate=panel.delegate;session.savedAccessory=panel.accessoryView;session.savedTitle=panel.title;session.savedColor=panel.color;session.savedContinuous=panel.continuous;session.savedAlpha=panel.showsAlpha;session.savedHides=panel.hidesOnDeactivate;
 if(!vellumDocumentPanels)vellumDocumentPanels=[NSMutableDictionary new];vellumDocumentPanels[@(token)]=session;
 panel.title=o[@"Title"];panel.delegate=session;panel.showsAlpha=session.alpha;panel.continuous=YES;NSDictionary *c=o[@"Color"];panel.color=[NSColor colorWithSRGBRed:[c[@"R"] doubleValue]/255 green:[c[@"G"] doubleValue]/255 blue:[c[@"B"] doubleValue]/255 alpha:session.alpha?[c[@"A"] doubleValue]/255:1];[panel setTarget:session];[panel setAction:@selector(changed:)];panel.accessoryView=vellumColorActions(session);panel.hidesOnDeactivate=YES;
 __weak VellumDocumentPanelSession *weakSession=session;session.parentObserver=[NSNotificationCenter.defaultCenter addObserverForName:NSWindowWillCloseNotification object:parent queue:nil usingBlock:^(NSNotification *note){[weakSession finish:6];}];
 if(!vellumServiceActive(token)){[session finish:1];return;}[parent addChildWindow:panel ordered:NSWindowAbove];[panel makeKeyAndOrderFront:nil];
 }@catch(NSException *e){VellumDocumentPanelSession *session=vellumDocumentPanels[@(token)];if(session)[session finish:2];else vellumServiceResult(token,"",2);}}});});}
void vellumCancelDocumentPanel(uint64_t token){dispatch_async(dispatch_get_main_queue(),^{@autoreleasepool{VellumDocumentPanelSession *session=vellumDocumentPanels[@(token)];if(session)[session finish:1];else vellumServiceResult(token,"",1);}});}
int vellumCancelAllDocumentPanels(void){int status=0;for(NSNumber *token in vellumDocumentPanels.allKeys.copy){VellumDocumentPanelSession *session=vellumDocumentPanels[token];if(session.finished){if([session cleanup])status=2;}else[session finish:1];}return vellumDocumentPanels.count?2:status;}
