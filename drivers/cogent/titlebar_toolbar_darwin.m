//go:build darwin && cgo

// Copyright (c) 2026 Davin Hills. SPDX-License-Identifier: MIT.
#import <Cocoa/Cocoa.h>
#import <objc/message.h>
#include <stdint.h>
extern int vellumTitlebarAction(uint64_t token,const char *identifier);
extern int vellumTitlebarCustomizationScope(uint64_t token,int customizing);
extern void vellumTitlebarItemsChanged(uint64_t token);
static char vellumTitlebarPaletteContext;
@class VellumTitlebarOwner;

// Send the macOS 15 setter dynamically so builds using older SDK headers
// retain the same behavior when run on a supported newer macOS.
static void vellumTitlebarDisplayCustomization(NSToolbar *toolbar, BOOL enabled) {
 if(@available(macOS 15.0,*)) {
  SEL setter=NSSelectorFromString(@"setAllowsDisplayModeCustomization:");
  if([toolbar respondsToSelector:setter])((void (*)(id,SEL,BOOL))objc_msgSend)(toolbar,setter,enabled);
 }
}

// Intercept AppKit's own palette entry, including the standard context menu.
// No replacement of NSWindow/GLFW's delegate, view, or responder is involved.
@interface VellumTitlebarToolbar : NSToolbar
@property(nonatomic,weak) VellumTitlebarOwner *owner;
@end

@interface VellumTitlebarOwner : NSObject <NSToolbarDelegate>
@property(nonatomic) uint64_t token;
@property(nonatomic,weak) NSWindow *window;
@property(nonatomic,strong) VellumTitlebarToolbar *toolbar;
@property(nonatomic,strong) NSDictionary<NSString *,NSDictionary *> *catalogue;
@property(nonatomic,strong) NSArray<NSString *> *order;
@property(nonatomic,strong) NSHashTable<NSToolbarItem *> *allItems;
@property(nonatomic) BOOL allowed;
@property(nonatomic) BOOL customizable;
@property(nonatomic) BOOL customizing;
@property(nonatomic) BOOL observingPalette;
@property(nonatomic) BOOL closed;
@property(nonatomic) NSWindowTitleVisibility priorTitle;
@property(nonatomic) NSWindowToolbarStyle priorStyle;
- (void)activate:(id)sender;
- (void)refreshEnabled;
- (void)updateCustomization:(BOOL)customizing;
@end

@implementation VellumTitlebarToolbar
- (void)runCustomizationPalette:(id)sender {
 VellumTitlebarOwner *owner=self.owner;
 if(!owner || owner.closed || !owner.allowed || !owner.customizable || self.customizationPaletteIsRunning)return;
 // Revoke queued Go actions before AppKit enters its native customization UI.
 if(!vellumTitlebarCustomizationScope(owner.token,1))return;
 owner.customizing=YES;[owner refreshEnabled];
 @try { [super runCustomizationPalette:sender]; }
 @finally { if(!owner.closed)[owner updateCustomization:self.customizationPaletteIsRunning]; }
}
@end

@implementation VellumTitlebarOwner
- (NSArray *)toolbarAllowedItemIdentifiers:(NSToolbar *)toolbar { return self.order; }
- (NSArray *)toolbarDefaultItemIdentifiers:(NSToolbar *)toolbar { return self.order; }
- (BOOL)toolbar:(NSToolbar *)toolbar itemIdentifier:(NSString *)identifier canBeInsertedAtIndex:(NSInteger)index {
 if(self.closed || !self.allowed)return NO;
 if(index==NSNotFound)return YES;
 NSArray *identifiers=[toolbar.items valueForKey:@"itemIdentifier"];
 return toolbar.items.count<32 || (self.catalogue[identifier] && [identifiers containsObject:identifier]);
}
- (NSToolbarItem *)toolbar:(NSToolbar *)toolbar itemForItemIdentifier:(NSString *)identifier willBeInsertedIntoToolbar:(BOOL)flag {
 NSDictionary *entry=self.catalogue[identifier]; if(!entry || self.closed)return nil;
 NSToolbarItem *item=[[NSToolbarItem alloc]initWithItemIdentifier:identifier];
 item.label=entry[@"label"]; item.paletteLabel=item.label;item.toolTip=item.label;
 item.image=entry[@"image"];item.target=self;item.action=@selector(activate:);
 item.autovalidates=NO;
 // Palette previews stay draggable even when their corresponding command is
 // disabled. Activation checks actual current toolbar membership below.
 item.enabled=!flag || (self.allowed && !self.customizing && ![entry[@"disabled"] boolValue]);
 NSMenuItem *menu=[[NSMenuItem alloc]initWithTitle:item.label action:@selector(activate:) keyEquivalent:@""];
 menu.target=self;menu.representedObject=identifier;menu.image=item.image;menu.enabled=item.enabled;
 item.menuFormRepresentation=menu;[self.allItems addObject:item];
 return item;
}
- (void)toolbarWillAddItem:(NSNotification *)notification {
 NSToolbarItem *item=notification.userInfo[NSToolbarItemKey];
 if(self.catalogue[item.itemIdentifier] && !self.closed) {
  item.target=self;item.action=@selector(activate:);item.menuFormRepresentation.target=self;item.menuFormRepresentation.action=@selector(activate:);
  item.enabled=self.allowed && !self.customizing && ![self.catalogue[item.itemIdentifier][@"disabled"] boolValue];item.menuFormRepresentation.enabled=item.enabled;
 }
 vellumTitlebarItemsChanged(self.token);
}
- (void)toolbarDidRemoveItem:(NSNotification *)notification {
 NSToolbarItem *item=notification.userInfo[NSToolbarItemKey];
 item.target=nil;item.action=NULL;item.menuFormRepresentation.target=nil;item.menuFormRepresentation.action=NULL;
 vellumTitlebarItemsChanged(self.token);
}
- (NSToolbarItem *)currentItemForSender:(id)sender {
 for(NSToolbarItem *item in self.toolbar.items) {
  if(sender==item || sender==item.menuFormRepresentation)return item;
 }
 return nil;
}
- (void)activate:(id)sender {
 if(self.closed || !self.allowed || self.customizing || self.toolbar.customizationPaletteIsRunning)return;
 NSToolbarItem *item=[self currentItemForSender:sender];if(!item)return;
 NSDictionary *entry=self.catalogue[item.itemIdentifier];
 if(!entry || [entry[@"disabled"] boolValue])return;
 vellumTitlebarAction(self.token,[entry[@"id"] UTF8String]);
}
- (BOOL)validateMenuItem:(NSMenuItem *)item {
 NSToolbarItem *current=[self currentItemForSender:item];if(!current)return NO;
 NSDictionary *entry=self.catalogue[current.itemIdentifier];
 return !self.closed && self.allowed && !self.customizing && entry && ![entry[@"disabled"] boolValue];
}
- (void)refreshEnabled {
 for(NSToolbarItem *item in self.toolbar.items) {
  NSDictionary *entry=self.catalogue[item.itemIdentifier];
  if(entry) { item.enabled=!self.closed && self.allowed && !self.customizing && ![entry[@"disabled"] boolValue];item.menuFormRepresentation.enabled=item.enabled; }
 }
 self.toolbar.allowsUserCustomization=!self.closed && self.allowed && self.customizable;
 vellumTitlebarDisplayCustomization(self.toolbar,!self.closed && self.allowed && self.customizable);
}
- (void)updateCustomization:(BOOL)customizing {
 if(self.closed || self.customizing==customizing)return;
 self.customizing=customizing;vellumTitlebarCustomizationScope(self.token,customizing);[self refreshEnabled];
}
- (void)observeValueForKeyPath:(NSString *)keyPath ofObject:(id)object change:(NSDictionary *)change context:(void *)context {
 if(context==&vellumTitlebarPaletteContext) { [self updateCustomization:self.toolbar.customizationPaletteIsRunning];return; }
 [super observeValueForKeyPath:keyPath ofObject:object change:change context:context];
}
@end

// Native main owns this bounded, per-window-token registry. Async requests resolve their token
// at execution, so they cannot retain or resurrect a retired attachment.
static NSMutableDictionary<NSNumber *,VellumTitlebarOwner *> *vellumTitlebars;
int vellumTitlebarAttach(uint64_t token,void *pointer,const char *json,const char *identifier) {
 @autoreleasepool {
  if(![NSThread isMainThread] || !pointer || !json || !identifier || !token)return 0;
  if(@available(macOS 11.0,*)) {
   NSString *settingsID=[NSString stringWithUTF8String:identifier];BOOL customizable=settingsID.length!=0;
   // Reliable palette scope revocation uses the documented macOS15+ KVO.
   if(customizable) { if(@available(macOS 15.0,*)) {} else return 0; }
   NSWindow *window=(__bridge NSWindow *)pointer;
   if(!window.contentView || window.toolbar || !(window.styleMask&NSWindowStyleMaskTitled) || vellumTitlebars.count>=32)return 0;
   NSData *data=[[NSString stringWithUTF8String:json]dataUsingEncoding:NSUTF8StringEncoding];
   NSArray *items=[NSJSONSerialization JSONObjectWithData:data options:0 error:NULL];
   if(![items isKindOfClass:[NSArray class]] || items.count==0 || items.count>16)return 0;
   NSMutableArray *order=[NSMutableArray array];NSMutableDictionary *catalogue=[NSMutableDictionary dictionary];
   for(NSDictionary *entry in items) {
    if([entry[@"kind"] isEqual:@"space"]) { [order addObject:NSToolbarSpaceItemIdentifier];continue; }
    if([entry[@"kind"] isEqual:@"flexible"]) { [order addObject:NSToolbarFlexibleSpaceItemIdentifier];continue; }
    if(![entry[@"kind"] isEqual:@"button"])return 0;
    NSData *png=[[NSData alloc]initWithBase64EncodedString:entry[@"png"] options:0];
    NSImage *image=[[NSImage alloc]initWithData:png];if(!image)return 0;
    image.template=NO;image.size=NSMakeSize(22,22);
    NSString *itemID=[@"com.vellum.titlebar.prototype." stringByAppendingString:entry[@"id"]];
    NSMutableDictionary *record=[entry mutableCopy];record[@"image"]=image;catalogue[itemID]=record;[order addObject:itemID];
   }
   VellumTitlebarOwner *owner=[VellumTitlebarOwner new];owner.token=token;owner.window=window;
   owner.catalogue=catalogue;owner.order=order;owner.customizable=customizable;owner.allItems=[NSHashTable weakObjectsHashTable];
   owner.priorTitle=window.titleVisibility;owner.priorStyle=window.toolbarStyle;
   VellumTitlebarToolbar *toolbar=[[VellumTitlebarToolbar alloc]initWithIdentifier:customizable?settingsID:@"com.vellum.native-titlebar.prototype"];
   owner.toolbar=toolbar;toolbar.owner=owner;toolbar.delegate=owner;
   toolbar.displayMode=NSToolbarDisplayModeIconOnly;
   toolbar.allowsUserCustomization=customizable;toolbar.autosavesConfiguration=customizable;
   vellumTitlebarDisplayCustomization(toolbar,customizable);
   if(customizable) {
    [toolbar addObserver:owner forKeyPath:@"customizationPaletteIsRunning" options:0 context:&vellumTitlebarPaletteContext];owner.observingPalette=YES;
   }
   NSView *view=window.contentView;NSSize client=view.bounds.size;
   if(!vellumTitlebars)vellumTitlebars=[NSMutableDictionary dictionary];vellumTitlebars[@(token)]=owner;
   window.titleVisibility=NSWindowTitleHidden;window.toolbarStyle=NSWindowToolbarStyleUnifiedCompact;window.toolbar=toolbar;
   // AppKit restores its own opaque autosave configuration during attachment.
   // Retired identifiers are refused by the delegate; cap repeated spaces.
   while(toolbar.items.count>32)[toolbar removeItemAtIndex:toolbar.items.count-1];
   [owner refreshEnabled];
   // Retain standard content geometry below the native chrome and GLFW's view.
   [window setContentSize:client];[view.superview layoutSubtreeIfNeeded];
   return window.contentView==view && window.toolbar==toolbar;
  }
  return 0;
 }
}
void vellumTitlebarAllowed(uint64_t token,int allowed) {
 dispatch_async(dispatch_get_main_queue(),^{
  VellumTitlebarOwner *owner=vellumTitlebars[@(token)];if(!owner || owner.closed)return;
  owner.allowed=allowed!=0;[owner refreshEnabled];
 });
}
void vellumTitlebarCustomize(uint64_t token) {
 dispatch_async(dispatch_get_main_queue(),^{
  VellumTitlebarOwner *owner=vellumTitlebars[@(token)];if(!owner || owner.closed || !owner.allowed || !owner.customizable)return;
  owner.toolbar.visible=YES;[owner.window makeKeyAndOrderFront:nil];[owner.toolbar runCustomizationPalette:nil];
 });
}
void vellumTitlebarDetach(uint64_t token) {
 @autoreleasepool {
  if(![NSThread isMainThread])return;
  VellumTitlebarOwner *owner=vellumTitlebars[@(token)];if(!owner)return;
  owner.closed=YES;owner.allowed=NO;[owner refreshEnabled];
  if(owner.observingPalette) { [owner.toolbar removeObserver:owner forKeyPath:@"customizationPaletteIsRunning" context:&vellumTitlebarPaletteContext];owner.observingPalette=NO; }
  if(owner.toolbar.customizationPaletteIsRunning && owner.window.attachedSheet)[owner.window endSheet:owner.window.attachedSheet returnCode:NSModalResponseCancel];
  for(NSToolbarItem *item in owner.allItems) { item.target=nil;item.action=NULL;item.menuFormRepresentation.target=nil;item.menuFormRepresentation.action=NULL; }
  owner.toolbar.delegate=nil;owner.toolbar.owner=nil;
  NSWindow *window=owner.window;
  if(window.toolbar==owner.toolbar) { window.toolbar=nil;window.titleVisibility=owner.priorTitle;window.toolbarStyle=owner.priorStyle; }
  [vellumTitlebars removeObjectForKey:@(token)];owner.catalogue=nil;owner.order=nil;owner.allItems=nil;owner.toolbar=nil;
 }
}

int vellumTitlebarCustomizationSupported(void) {
 if(@available(macOS 15.0,*))return 1;return 0;
}

extern void *vellumAXWindowForToken(uint64_t);
extern int vellumTitlebarIsLive(uint64_t);
extern void vellumTitlebarAttached(uint64_t,int);
extern char *vellumTitlebarStateJSON(uint64_t);
void vellumTitlebarRequestAttach(uint64_t token,const char *json,const char *identifier) {
 NSString *payload=json?[[NSString alloc]initWithUTF8String:json]:nil;
 NSString *settings=identifier?[[NSString alloc]initWithUTF8String:identifier]:nil;
 dispatch_async(dispatch_get_main_queue(),^{@autoreleasepool{
  if(!vellumTitlebarIsLive(token))return;
  void *window=vellumAXWindowForToken(token);
  int success=window&&payload&&settings ? vellumTitlebarAttach(token,window,payload.UTF8String,settings.UTF8String):0;
  if(!success)vellumTitlebarDetach(token);
  vellumTitlebarAttached(token,success);
 }});
}
void vellumTitlebarStateChanged(uint64_t token) {
 dispatch_async(dispatch_get_main_queue(),^{@autoreleasepool{
  char *json=vellumTitlebarStateJSON(token);if(!json)return;
  NSData *data=[[NSString stringWithUTF8String:json]dataUsingEncoding:NSUTF8StringEncoding];free(json);
  VellumTitlebarOwner *owner=vellumTitlebars[@(token)];if(!owner||owner.closed)return;
  NSArray *states=[NSJSONSerialization JSONObjectWithData:data options:0 error:NULL];
  if(![states isKindOfClass:NSArray.class]||states.count>16)return;
  NSMutableDictionary *catalogue=[owner.catalogue mutableCopy];
  for(NSDictionary *state in states){
   NSString *key=[@"com.vellum.titlebar.prototype." stringByAppendingString:state[@"id"]];
   NSMutableDictionary *entry=[catalogue[key] mutableCopy];if(!entry)continue;
   entry[@"disabled"]=state[@"disabled"];catalogue[key]=entry;
  }
  owner.catalogue=catalogue;[owner refreshEnabled];
 }});
}
