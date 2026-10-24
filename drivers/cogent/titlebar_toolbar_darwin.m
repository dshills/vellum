//go:build darwin && cgo

// Copyright (c) 2026 Davin Hills. SPDX-License-Identifier: MIT.
#import <Cocoa/Cocoa.h>
#include <stdint.h>
extern int vellumTitlebarAction(uint64_t token,const char *identifier);

@interface VellumTitlebarOwner : NSObject <NSToolbarDelegate>
@property(nonatomic) uint64_t token;
@property(nonatomic,weak) NSWindow *window;
@property(nonatomic,strong) NSToolbar *toolbar;
@property(nonatomic,strong) NSDictionary<NSString *,NSDictionary *> *catalogue;
@property(nonatomic,strong) NSArray<NSString *> *order;
@property(nonatomic) BOOL allowed;
@property(nonatomic) BOOL closed;
@property(nonatomic) NSWindowTitleVisibility priorTitle;
@property(nonatomic) NSWindowToolbarStyle priorStyle;
- (void)activate:(id)sender;
- (void)refreshEnabled;
@end

@implementation VellumTitlebarOwner
- (NSArray *)toolbarAllowedItemIdentifiers:(NSToolbar *)toolbar { return self.order; }
- (NSArray *)toolbarDefaultItemIdentifiers:(NSToolbar *)toolbar { return self.order; }
- (NSToolbarItem *)toolbar:(NSToolbar *)toolbar itemForItemIdentifier:(NSString *)identifier willBeInsertedIntoToolbar:(BOOL)flag {
 NSDictionary *entry=self.catalogue[identifier]; if(!entry || self.closed)return nil;
 NSToolbarItem *item=[[NSToolbarItem alloc]initWithItemIdentifier:identifier];
 item.label=entry[@"label"]; item.paletteLabel=item.label;item.toolTip=item.label;
 item.image=entry[@"image"];item.target=self;item.action=@selector(activate:);
 item.autovalidates=NO;item.enabled=self.allowed && ![entry[@"disabled"] boolValue];
 NSMenuItem *menu=[[NSMenuItem alloc]initWithTitle:item.label action:@selector(activate:) keyEquivalent:@""];
 menu.target=self;menu.representedObject=identifier;menu.image=item.image;menu.enabled=item.enabled;
 item.menuFormRepresentation=menu;
 return item;
}
- (void)activate:(id)sender {
 if(self.closed || !self.allowed)return;
 NSString *identifier=nil;
 if([sender isKindOfClass:NSToolbarItem.class])identifier=[sender itemIdentifier];
 else if([sender isKindOfClass:NSMenuItem.class])identifier=[sender representedObject];
 if(![identifier isKindOfClass:NSString.class])return;
 NSDictionary *entry=self.catalogue[identifier];
 if(!entry || [entry[@"disabled"] boolValue])return;
 vellumTitlebarAction(self.token,[entry[@"id"] UTF8String]);
}
- (BOOL)validateMenuItem:(NSMenuItem *)item {
 NSDictionary *entry=self.catalogue[item.representedObject];
 return !self.closed && self.allowed && entry && ![entry[@"disabled"] boolValue];
}
- (void)refreshEnabled {
 for(NSToolbarItem *item in self.toolbar.items) {
  NSDictionary *entry=self.catalogue[item.itemIdentifier];
  if(entry) { item.enabled=!self.closed && self.allowed && ![entry[@"disabled"] boolValue];item.menuFormRepresentation.enabled=item.enabled; }
 }
}
@end

// Native main owns this bounded registry. Queued scope updates resolve tokens
// at execution time, so they cannot retain or resurrect a retired attachment.
static NSMutableDictionary<NSNumber *,VellumTitlebarOwner *> *vellumTitlebars;

int vellumTitlebarAttach(uint64_t token,void *pointer,const char *json) {
 @autoreleasepool {
  if(![NSThread isMainThread] || !pointer || !json || !token)return 0;
  if(@available(macOS 11.0,*)) {
   NSWindow *window=(__bridge NSWindow *)pointer;
   if(!window.contentView || window.toolbar || !(window.styleMask&NSWindowStyleMaskTitled) || vellumTitlebars.count>=1)return 0;
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
    NSString *identifier=[@"com.vellum.titlebar.prototype." stringByAppendingString:entry[@"id"]];
    NSMutableDictionary *record=[entry mutableCopy];record[@"image"]=image;
    catalogue[identifier]=record;[order addObject:identifier];
   }
   VellumTitlebarOwner *owner=[VellumTitlebarOwner new];owner.token=token;owner.window=window;
   owner.catalogue=catalogue;owner.order=order;owner.priorTitle=window.titleVisibility;owner.priorStyle=window.toolbarStyle;
   NSToolbar *toolbar=[[NSToolbar alloc]initWithIdentifier:@"com.vellum.native-titlebar.prototype"];
   owner.toolbar=toolbar;toolbar.delegate=owner;toolbar.allowsUserCustomization=NO;toolbar.autosavesConfiguration=NO;
   if(@available(macOS 15.0,*))toolbar.allowsDisplayModeCustomization=NO;
   toolbar.displayMode=NSToolbarDisplayModeIconOnly;
   NSView *view=window.contentView;NSSize client=view.bounds.size;
   if(!vellumTitlebars)vellumTitlebars=[NSMutableDictionary dictionary];vellumTitlebars[@(token)]=owner;
   window.titleVisibility=NSWindowTitleHidden;window.toolbarStyle=NSWindowToolbarStyleUnifiedCompact;window.toolbar=toolbar;
   // Standard content remains below AppKit chrome; never replace GLFW's view,
   // window delegate, first responder, Metal layer, or frame-coordinate policy.
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
void vellumTitlebarDetach(uint64_t token) {
 @autoreleasepool {
  if(![NSThread isMainThread])return;
  VellumTitlebarOwner *owner=vellumTitlebars[@(token)];if(!owner)return;
  owner.closed=YES;owner.allowed=NO;[owner refreshEnabled];
  for(NSToolbarItem *item in owner.toolbar.items) { item.target=nil;item.action=NULL;item.menuFormRepresentation.target=nil;item.menuFormRepresentation.action=NULL; }
  owner.toolbar.delegate=nil;
  NSWindow *window=owner.window;
  if(window.toolbar==owner.toolbar) { window.toolbar=nil;window.titleVisibility=owner.priorTitle;window.toolbarStyle=owner.priorStyle; }
  [vellumTitlebars removeObjectForKey:@(token)];owner.catalogue=nil;owner.order=nil;owner.toolbar=nil;
 }
}
