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
static BOOL vellumOwnsColorPanel,vellumOwnsFontPanel;
static BOOL vellumDocumentPanelMayUse(BOOL existed,BOOL owned,BOOL visible){return !existed||(owned&&!visible);}
int vellumDocumentPanelAdmissionFixture(int existed,int owned,int visible){return vellumDocumentPanelMayUse(existed,owned,visible);}

static NSPoint vellumDocumentPanelPosition(NSRect parent,NSRect visible,NSSize size){
 CGFloat x=NSMidX(parent)-size.width/2,y=NSMidY(parent)-size.height/2;
 if(size.height>NSHeight(visible))y=NSMaxY(visible)-size.height;else y=fmax(NSMinY(visible),fmin(y,NSMaxY(visible)-size.height));
 return NSMakePoint(fmax(NSMinX(visible),fmin(x,NSMaxX(visible)-size.width)),y);
}
// Pure geometry fixture; no screen/window/panel is constructed offscreen.
double vellumDocumentPanelPositionFixture(int axis,double px,double py,double pw,double ph,double vx,double vy,double vw,double vh,double width,double height){NSPoint p=vellumDocumentPanelPosition(NSMakeRect(px,py,pw,ph),NSMakeRect(vx,vy,vw,vh),NSMakeSize(width,height));return axis==0?p.x:p.y;}

static NSString *vellumPanelApplyTitle(NSString *kind){return [kind isEqual:@"font"]?@"Use Font":@"Use Color";}
static NSView *vellumPanelAccessory(NSPanel *panel,BOOL font){return font?((NSFontPanel *)panel).accessoryView:((NSColorPanel *)panel).accessoryView;}
static void vellumSetPanelAccessory(NSPanel *panel,BOOL font,NSView *view){if(font)((NSFontPanel *)panel).accessoryView=view;else ((NSColorPanel *)panel).accessoryView=view;}

// Foundation-only bounded parser: never creates AppKit panels in offscreen tests.
static NSDictionary *vellumDocumentPanelRecipe(NSData *data){
 NSDictionary *request=[NSJSONSerialization JSONObjectWithData:data options:0 error:nil];if(![request isKindOfClass:NSDictionary.class]||(![request[@"Kind"] isEqual:@"color"]&&![request[@"Kind"] isEqual:@"font"]))return nil;
 NSDictionary *o=request[@"Options"];if(![o isKindOfClass:NSDictionary.class]||![o[@"Title"] isKindOfClass:NSString.class]||![o[@"Title"] length]||[o[@"Title"] length]>256)return nil;
 if([request[@"Kind"] isEqual:@"font"]){NSString *name=o[@"Name"];NSNumber *size=o[@"Size"];if(![name isKindOfClass:NSString.class]||!name.length||name.length>256||![size isKindOfClass:NSNumber.class]||!isfinite(size.doubleValue)||size.doubleValue<6||size.doubleValue>144)return nil;return request;}
 if(![o[@"ShowsAlpha"] isKindOfClass:NSNumber.class])return nil;NSDictionary *c=o[@"Color"];if(![c isKindOfClass:NSDictionary.class])return nil;for(NSString *key in @[@"R",@"G",@"B",@"A"]){id v=c[key];if(![v isKindOfClass:NSNumber.class]||!isfinite([v doubleValue])||[v doubleValue]<0||[v doubleValue]>255||floor([v doubleValue])!=[v doubleValue])return nil;}return request;
}
static NSString *vellumPanelJSON(NSDictionary *value){NSData *data=value?[NSJSONSerialization dataWithJSONObject:value options:0 error:nil]:nil;return data?[[NSString alloc]initWithData:data encoding:NSUTF8StringEncoding]:nil;}
char *vellumDocumentPanelRecipeFixture(const char *json){@autoreleasepool{@try{if(!json||strnlen(json,VELLUM_SERVICE_REQUEST_MAX+1)>VELLUM_SERVICE_REQUEST_MAX)return NULL;NSDictionary *recipe=vellumDocumentPanelRecipe([NSData dataWithBytes:json length:strlen(json)]);NSString *result=recipe?vellumPanelJSON(@{@"Kind":recipe[@"Kind"],@"Options":recipe[@"Options"],@"ApplyTitle":vellumPanelApplyTitle(recipe[@"Kind"])}):nil;return result?strdup(result.UTF8String):NULL;}@catch(NSException *e){return NULL;}}}
@interface VellumDocumentPanelSession:NSObject<NSWindowDelegate>
@property uint64_t token;
@property BOOL finished;
@property BOOL alpha;
@property(strong) NSPanel *panel;
@property(weak) NSWindow *parent;
@property(weak) id<NSWindowDelegate> savedDelegate;
@property(strong) NSView *savedAccessory;
@property(copy) NSString *savedTitle;
@property(copy) NSColor *savedColor;
@property BOOL savedContinuous,savedAlpha,savedHides,savedFontMultiple,savedFontEnabled;
@property(copy) NSString *kind;
@property(strong) NSAppearance *savedAppearance;
@property(strong) NSFont *font;
@property(strong) NSFont *savedFont;
@property(weak) id savedFontTarget;
@property SEL savedFontAction;
@property(strong) id parentObserver;
-(void)finish:(int)status;
-(int)cleanup;
@end
@implementation VellumDocumentPanelSession
-(NSDictionary *)value{if([self.kind isEqual:@"font"]){NSFont *font=self.font;if(!font||font.pointSize<6||font.pointSize>144||!isfinite(font.pointSize))return nil;return @{@"Name":font.fontName,@"Family":font.familyName?:font.fontName,@"Size":@(font.pointSize),@"Bold":((font.fontDescriptor.symbolicTraits&NSFontDescriptorTraitBold)!=0?@YES:@NO),@"Italic":((font.fontDescriptor.symbolicTraits&NSFontDescriptorTraitItalic)!=0?@YES:@NO)};}NSColor *c=[((NSColorPanel *)self.panel).color colorUsingColorSpace:NSColorSpace.sRGBColorSpace];if(!c)return nil;CGFloat red=c.redComponent,green=c.greenComponent,blue=c.blueComponent,alpha=self.alpha?c.alphaComponent:1;if(!isfinite(red)||!isfinite(green)||!isfinite(blue)||!isfinite(alpha))return nil;return @{@"R":@(lround(fmax(0,fmin(1,red))*255)),@"G":@(lround(fmax(0,fmin(1,green))*255)),@"B":@(lround(fmax(0,fmin(1,blue))*255)),@"A":@(lround(fmax(0,fmin(1,alpha))*255))};}
-(void)changed:(id)sender{@try{if(self.finished||!vellumServiceActive(self.token))return;NSString *wire=vellumPanelJSON([self value]);if(wire)vellumDocumentPanelPreview(self.token,(char *)wire.UTF8String);}@catch(NSException *e){[self finish:2];}}
-(NSFontPanelModeMask)validModesForFontPanel:(NSFontPanel *)panel{return NSFontPanelModeMaskFace|NSFontPanelModeMaskSize|NSFontPanelModeMaskCollection;}
-(void)changeFont:(NSFontManager *)manager{@try{if(self.finished||!vellumServiceActive(self.token))return;NSFont *font=[manager convertFont:self.font];CGFloat size=fmax(6,fmin(144,font.pointSize));self.font=[NSFont fontWithDescriptor:font.fontDescriptor size:size];[manager setSelectedFont:self.font isMultiple:NO];[self changed:manager];}@catch(NSException *e){[self finish:2];}}
-(void)apply:(id)sender{@try{if(![self.panel makeFirstResponder:nil]){NSBeep();return;}[self finish:0];}@catch(NSException *e){[self finish:2];}}
-(void)cancel:(id)sender{[self finish:1];}
-(BOOL)windowShouldClose:(NSWindow *)sender{[self finish:1];return NO;}
-(int)cleanup{int result=0;
 @try{if(self.parentObserver){[NSNotificationCenter.defaultCenter removeObserver:self.parentObserver];self.parentObserver=nil;}}@catch(NSException *e){result=2;}
 @try{[self.parent removeChildWindow:self.panel];[self.panel orderOut:nil];}@catch(NSException *e){result=2;}
 @try{self.panel.delegate=self.savedDelegate;vellumSetPanelAccessory(self.panel,[self.kind isEqual:@"font"],self.savedAccessory);self.panel.title=self.savedTitle?:@"Panel";self.panel.hidesOnDeactivate=self.savedHides;self.panel.appearance=self.savedAppearance;
 if([self.kind isEqual:@"font"]){NSFontManager *manager=NSFontManager.sharedFontManager;manager.target=self.savedFontTarget;manager.action=self.savedFontAction;if(self.savedFont)[manager setSelectedFont:self.savedFont isMultiple:self.savedFontMultiple];((NSFontPanel *)self.panel).enabled=self.savedFontEnabled;}else{NSColorPanel *panel=(NSColorPanel *)self.panel;[panel setTarget:nil];[panel setAction:NULL];panel.continuous=self.savedContinuous;panel.showsAlpha=self.savedAlpha;if(self.savedColor)panel.color=self.savedColor;}}@catch(NSException *e){result=2;}
 if(!result){[vellumDocumentPanels removeObjectForKey:@(self.token)];self.panel=nil;self.savedAccessory=nil;self.savedColor=nil;self.savedFont=nil;self.font=nil;self.savedFontTarget=nil;self.parent=nil;}return result;
}
-(void)finish:(int)status{if(self.finished)return;self.finished=YES;NSString *wire=nil;@try{if(status==0){wire=vellumPanelJSON([self value]);if(!wire)status=2;}}@catch(NSException *e){status=2;}if([self cleanup])status=2;vellumServiceResult(self.token,(char *)(wire?:@"").UTF8String,status);}
@end
static NSView *vellumPanelActions(VellumDocumentPanelSession *session){NSView *view=[[NSView alloc]initWithFrame:NSMakeRect(0,0,300,42)];NSButton *cancel=[NSButton buttonWithTitle:@"Cancel" target:session action:@selector(cancel:)];cancel.frame=NSMakeRect(106,8,86,28);cancel.bezelStyle=NSBezelStyleRounded;cancel.keyEquivalent=@"\033";NSButton *apply=[NSButton buttonWithTitle:vellumPanelApplyTitle(session.kind) target:session action:@selector(apply:)];apply.frame=NSMakeRect(198,8,96,28);apply.bezelStyle=NSBezelStyleRounded;apply.keyEquivalent=@"\r";cancel.autoresizingMask=NSViewMinXMargin;apply.autoresizingMask=NSViewMinXMargin;view.autoresizingMask=NSViewWidthSizable;[view addSubview:cancel];[view addSubview:apply];return view;}
void vellumChooseDocumentPanel(uint64_t token,uint64_t parentToken,const char *json){vellumWithServiceJSON(token,json,^(NSData *data){dispatch_async(dispatch_get_main_queue(),^{@autoreleasepool{NSPanel *createdPanel=nil;@try{
 if(!vellumServiceActive(token)){vellumServiceResult(token,"",1);return;}NSDictionary *recipe=vellumDocumentPanelRecipe(data);if(!recipe){vellumServiceResult(token,"",2);return;}
 NSWindow *parent=(__bridge NSWindow *)vellumAXNativeWindow(parentToken);if(!parent){vellumServiceResult(token,"",6);return;}if(!parent.isVisible||parent.isMiniaturized){vellumServiceResult(token,"",2);return;}
 NSString *kind=recipe[@"Kind"];BOOL fontMode=[kind isEqual:@"font"];
 BOOL existed=fontMode?NSFontPanel.sharedFontPanelExists:NSColorPanel.sharedColorPanelExists;BOOL owned=fontMode?vellumOwnsFontPanel:vellumOwnsColorPanel;
 if(vellumDocumentPanels.count||parent.attachedSheet||NSApp.modalWindow||(existed&&!owned)){vellumServiceResult(token,"",5);return;}
 NSDictionary *o=recipe[@"Options"];NSFont *font=fontMode?[NSFont fontWithName:o[@"Name"] size:[o[@"Size"] doubleValue]]:nil;if(fontMode&&!font){vellumServiceResult(token,"",2);return;}
 NSPanel *panel=fontMode?NSFontPanel.sharedFontPanel:NSColorPanel.sharedColorPanel;if(!existed)createdPanel=panel;
 // AppKit may immediately show a newly-created shared panel using its saved
 // visibility. Only a panel that existed before this call has another owner.
 if(!vellumDocumentPanelMayUse(existed,owned,panel.isVisible)){vellumServiceResult(token,"",5);return;}
 if(fontMode)vellumOwnsFontPanel=YES;else vellumOwnsColorPanel=YES;
 VellumDocumentPanelSession *session=[VellumDocumentPanelSession new];session.token=token;session.kind=kind;session.parent=parent;session.panel=panel;session.savedDelegate=panel.delegate;session.savedAccessory=vellumPanelAccessory(panel,fontMode);session.savedTitle=panel.title;session.savedHides=panel.hidesOnDeactivate;session.savedAppearance=panel.appearance;
 if(fontMode){NSFontManager *manager=NSFontManager.sharedFontManager;session.savedFont=manager.selectedFont;session.savedFontMultiple=manager.isMultiple;session.savedFontTarget=manager.target;session.savedFontAction=manager.action;session.savedFontEnabled=((NSFontPanel *)panel).enabled;session.font=font;}
 else{NSColorPanel *colors=(NSColorPanel *)panel;session.alpha=[o[@"ShowsAlpha"] boolValue];session.savedColor=colors.color;session.savedContinuous=colors.continuous;session.savedAlpha=colors.showsAlpha;}
 if(!vellumDocumentPanels)vellumDocumentPanels=[NSMutableDictionary new];vellumDocumentPanels[@(token)]=session;
 panel.restorable=NO;panel.title=o[@"Title"];panel.delegate=session;panel.hidesOnDeactivate=YES;panel.appearance=parent.effectiveAppearance;
 if(fontMode){NSFontManager *manager=NSFontManager.sharedFontManager;manager.target=session;manager.action=@selector(changeFont:);[manager setSelectedFont:font isMultiple:NO];[(NSFontPanel *)panel setPanelFont:font isMultiple:NO];((NSFontPanel *)panel).enabled=YES;// Rebuilding the shared panel on every reuse invalidates AppKit table views on macOS 26.
 if(!existed)[(NSFontPanel *)panel reloadDefaultFontFamilies];}
 else{NSColorPanel *colors=(NSColorPanel *)panel;colors.showsAlpha=session.alpha;colors.continuous=YES;NSDictionary *c=o[@"Color"];colors.color=[NSColor colorWithSRGBRed:[c[@"R"] doubleValue]/255 green:[c[@"G"] doubleValue]/255 blue:[c[@"B"] doubleValue]/255 alpha:session.alpha?[c[@"A"] doubleValue]/255:1];[colors setTarget:session];[colors setAction:@selector(changed:)];}
 vellumSetPanelAccessory(panel,fontMode,vellumPanelActions(session));
 NSScreen *screen=parent.screen?:NSScreen.mainScreen;if(screen)[panel setFrameOrigin:vellumDocumentPanelPosition(parent.frame,screen.visibleFrame,panel.frame.size)];
 __weak VellumDocumentPanelSession *weakSession=session;session.parentObserver=[NSNotificationCenter.defaultCenter addObserverForName:NSWindowWillCloseNotification object:parent queue:nil usingBlock:^(NSNotification *note){[weakSession finish:6];}];
 if(!vellumServiceActive(token)){[session finish:1];return;}[parent addChildWindow:panel ordered:NSWindowAbove];[panel makeKeyAndOrderFront:nil];
 }@catch(NSException *e){VellumDocumentPanelSession *session=vellumDocumentPanels[@(token)];if(session)[session finish:2];else{if(createdPanel){@try{[createdPanel orderOut:nil];}@catch(NSException *ignored){}}vellumServiceResult(token,"",2);}}}});});}
void vellumCancelDocumentPanel(uint64_t token){dispatch_async(dispatch_get_main_queue(),^{@autoreleasepool{VellumDocumentPanelSession *session=vellumDocumentPanels[@(token)];if(session)[session finish:1];else vellumServiceResult(token,"",1);}});}
int vellumCancelAllDocumentPanels(void){int status=0;for(NSNumber *token in vellumDocumentPanels.allKeys.copy){VellumDocumentPanelSession *session=vellumDocumentPanels[token];if(session.finished){if([session cleanup])status=2;}else[session finish:1];}return vellumDocumentPanels.count?2:status;}
