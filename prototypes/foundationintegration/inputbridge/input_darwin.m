#import <AppKit/AppKit.h>
#import <Carbon/Carbon.h>
#import <objc/runtime.h>
#import <objc/message.h>
#include "input.h"
#include <string.h>

extern bool goVellumPreflight(uintptr_t, uint64_t, char *, char *, int, int, int, int);
extern void goVellumCommit(uintptr_t);
static char inputKey;
@interface VellumInputState : NSObject {
@public
 NSView *view;
 Class baseClass, subclass;
 uintptr_t handle;
 uint64_t sequence;
 BOOL active, forwarding, forwarded, closed;
 NSEvent *keyEvent;
 NSMutableString *text;
 NSString *before;
 NSRange selected, marked, beforeSelected;
 NSRect caret;
 NSUInteger caretOffset, beforeCaretOffset;
}
@end
@implementation VellumInputState
- (void)windowGeometryChanged:(NSNotification *)notification {if(!closed)[[view inputContext] invalidateCharacterCoordinates];}
- (void)dealloc { [[NSNotificationCenter defaultCenter] removeObserver:self]; [text release]; [before release]; [super dealloc]; }
@end
static VellumInputState *state(id view) { return objc_getAssociatedObject(view, &inputKey); }
static struct objc_super parent(id view) {
 VellumInputState *s=state(view);
 return (struct objc_super){view,s ? s->baseClass : class_getSuperclass(object_getClass(view))};
}
static BOOL responds(id view,SEL selector){VellumInputState *s=state(view);return s && class_getInstanceMethod(s->baseClass,selector)!=NULL;}
static BOOL usable(VellumInputState *s) { return s && s->active && !s->closed && !s->forwarding; }
static BOOL composing(VellumInputState *s){return s && (s->before || s->marked.location!=NSNotFound);}
static BOOL validRange(NSString *text, NSRange range) {
 NSUInteger count=text.length;
 if(range.location==NSNotFound || range.location>count || range.length>count-range.location)return NO;
 NSUInteger end=range.location+range.length;
 for(NSUInteger i=0;i<2;i++){
  NSUInteger p=i ? end : range.location;
  if(p>0 && p<count && CFStringIsSurrogateHighCharacter([text characterAtIndex:p-1]) && CFStringIsSurrogateLowCharacter([text characterAtIndex:p]))return NO;
 }
 return YES;
}
static NSString *characters(id value) {
 NSString *text=[value isKindOfClass:NSAttributedString.class] ? [value string] : value;
 if(![text isKindOfClass:NSString.class] || ![text dataUsingEncoding:NSUTF8StringEncoding])return nil;
 for(NSUInteger i=0;i<text.length;i++)if([text characterAtIndex:i]==0)return nil;
 return text;
}
// Capacity is reserved against the prospective snapshot BEFORE any accepted
// native text/selection/composition mutation. Refusal leaves state unchanged.
static BOOL admit(VellumInputState *s,const char *kind,NSString *next,NSRange selected,NSRange marked,NSUInteger focus) {
 if(!s || s->closed || s->sequence==UINT64_MAX)return NO;
 int start=marked.location==NSNotFound ? -1 : (int)marked.location;
 int end=start<0 ? -1 : start+(int)marked.length;
 NSUInteger anchor=focus==selected.location ? NSMaxRange(selected) : selected.location;
 if(!goVellumPreflight(s->handle,s->sequence+1,(char *)kind,(char *)next.UTF8String,(int)anchor,(int)focus,start,end)){
  NSBeep();return NO; // Explicit audible refusal; caller can retry after drain.
 }
 return YES;
}
static void emit(VellumInputState *s) {s->sequence++;goVellumCommit(s->handle);}
static BOOL finish(VellumInputState *s, BOOL cancel) {
 if(!s->before && s->marked.location==NSNotFound)return YES;
 // Externally published marked text has no restoration baseline. Cancel must
 // refuse explicitly; unmark can still commit the current state transactionally.
 if(cancel && !s->before){NSBeep();return NO;}
 NSString *next=cancel ? s->before : s->text;
 NSRange selection=cancel ? s->beforeSelected : s->selected;
 if(!admit(s,cancel ? "cancel" : "unmark",next,selection,NSMakeRange(NSNotFound,0),cancel ? s->beforeCaretOffset : s->caretOffset))return NO;
 if(cancel){[s->text setString:next];s->selected=selection;s->caretOffset=s->beforeCaretOffset;}
 s->marked=NSMakeRange(NSNotFound,0);
 [s->before release];s->before=nil;
 emit(s);return YES;
}
static void forwardKey(id view) {
 VellumInputState *s=state(view);
 if(!s || !s->keyEvent || s->forwarded)return;
 s->forwarded=YES;s->forwarding=YES;
 struct objc_super p=parent(view);
 ((void(*)(struct objc_super *,SEL,id))objc_msgSendSuper)(&p,@selector(keyDown:),s->keyEvent);
 s->forwarding=NO;
}
@interface VellumInputMethods : NSView <NSTextInputClient>
@end
@implementation VellumInputMethods
- (void)keyDown:(NSEvent *)event {
 VellumInputState *s=state(self);
 if(!usable(s)){struct objc_super p=parent(self);((void(*)(struct objc_super *,SEL,id))objc_msgSendSuper)(&p,_cmd,event);return;}
 s->keyEvent=event;s->forwarded=NO;
 if(event.keyCode==53 && composing(s)){if(finish(s,YES))[[self inputContext] discardMarkedText];s->keyEvent=nil;return;}
 // Preserve application shortcuts; ordinary input belongs to the input context.
 if(event.modifierFlags & (NSEventModifierFlagCommand)){
  if(composing(s)){if(!finish(s,NO)){s->keyEvent=nil;return;}[[self inputContext] discardMarkedText];}forwardKey(self);
 }else{[self interpretKeyEvents:@[event]];}
 s->keyEvent=nil;
}
- (void)doCommandBySelector:(SEL)selector {
 VellumInputState *s=state(self);
 if(!usable(s)){if(!responds(self,_cmd)){return;}struct objc_super p=parent(self);((void(*)(struct objc_super *,SEL,SEL))objc_msgSendSuper)(&p,_cmd,selector);return;}
 if(selector==@selector(cancelOperation:) && composing(s)){if(finish(s,YES))[[self inputContext] discardMarkedText];return;}
 if(composing(s)){if(!finish(s,NO)){s->keyEvent=nil;return;}[[self inputContext] discardMarkedText];}forwardKey(self);
}
- (void)setMarkedText:(id)value selectedRange:(NSRange)selection replacementRange:(NSRange)replacement {
 VellumInputState *s=state(self);
 if(!usable(s)){if(!responds(self,_cmd)){return;}struct objc_super p=parent(self);((void(*)(struct objc_super *,SEL,id,NSRange,NSRange))objc_msgSendSuper)(&p,_cmd,value,selection,replacement);return;}
 NSString *insert=characters(value);if(!insert || !validRange(insert,selection))return;
 NSRange range=replacement.location==NSNotFound ? (s->marked.location!=NSNotFound ? s->marked : s->selected) : replacement;
 if(!validRange(s->text,range))return;
 NSMutableString *next=[[s->text mutableCopy] autorelease];
 [next replaceCharactersInRange:range withString:insert];
 NSRange mark=NSMakeRange(range.location,insert.length),sel=NSMakeRange(range.location+selection.location,selection.length);
 if(!admit(s,"mark",next,sel,mark,NSMaxRange(sel)))return;
 if(!composing(s)){s->before=[s->text copy];s->beforeSelected=s->selected;s->beforeCaretOffset=s->caretOffset;}
 [s->text setString:next];
 s->marked=NSMakeRange(range.location,insert.length);
 s->selected=NSMakeRange(range.location+selection.location,selection.length);
 s->caretOffset=NSMaxRange(s->selected);emit(s);
}
- (void)insertText:(id)value replacementRange:(NSRange)replacement {
 VellumInputState *s=state(self);
 if(!usable(s)){if(!responds(self,_cmd)){return;}struct objc_super p=parent(self);((void(*)(struct objc_super *,SEL,id,NSRange))objc_msgSendSuper)(&p,_cmd,value,replacement);return;}
 NSString *insert=characters(value);if(!insert)return;
 NSRange range=replacement.location==NSNotFound ? (s->marked.location!=NSNotFound ? s->marked : s->selected) : replacement;
 if(!validRange(s->text,range))return;
 NSMutableString *next=[[s->text mutableCopy] autorelease];
 [next replaceCharactersInRange:range withString:insert];
 NSRange sel=NSMakeRange(range.location+insert.length,0);
 if(!admit(s,"insert",next,sel,NSMakeRange(NSNotFound,0),sel.location))return;
 [s->text setString:next];s->selected=sel;
 s->marked=NSMakeRange(NSNotFound,0);[s->before release];s->before=nil;
 s->caretOffset=NSMaxRange(s->selected);emit(s);
}
- (void)unmarkText {
 VellumInputState *s=state(self);
 if(usable(s)){finish(s,NO);return;}
 if(!responds(self,_cmd))return;
 struct objc_super p=parent(self);((void(*)(struct objc_super *,SEL))objc_msgSendSuper)(&p,_cmd);
}
- (BOOL)hasMarkedText {VellumInputState *s=state(self);if(usable(s))return s->marked.location!=NSNotFound && s->marked.length>0;if(!responds(self,_cmd))return NO;struct objc_super p=parent(self);return ((BOOL(*)(struct objc_super *,SEL))objc_msgSendSuper)(&p,_cmd);}
- (NSRange)markedRange {VellumInputState *s=state(self);if(usable(s))return s->marked;if(!responds(self,_cmd))return NSMakeRange(NSNotFound,0);struct objc_super p=parent(self);return ((NSRange(*)(struct objc_super *,SEL))objc_msgSendSuper)(&p,_cmd);}
- (NSRange)selectedRange {VellumInputState *s=state(self);if(usable(s))return s->selected;if(!responds(self,_cmd))return NSMakeRange(NSNotFound,0);struct objc_super p=parent(self);return ((NSRange(*)(struct objc_super *,SEL))objc_msgSendSuper)(&p,_cmd);}
- (NSArray *)validAttributesForMarkedText {if(usable(state(self)))return @[NSUnderlineStyleAttributeName,NSMarkedClauseSegmentAttributeName];if(!responds(self,_cmd))return @[];struct objc_super p=parent(self);return ((id(*)(struct objc_super *,SEL))objc_msgSendSuper)(&p,_cmd);}
- (NSAttributedString *)attributedSubstringForProposedRange:(NSRange)range actualRange:(NSRangePointer)actual {
 VellumInputState *s=state(self);
 if(!usable(s)){if(!responds(self,_cmd)){if(actual)*actual=NSMakeRange(NSNotFound,0);return nil;}struct objc_super p=parent(self);return ((id(*)(struct objc_super *,SEL,NSRange,NSRangePointer))objc_msgSendSuper)(&p,_cmd,range,actual);}
 if(actual)*actual=NSMakeRange(NSNotFound,0);
 // Context requests may extend beyond the document. Clamp without adding
 // untrusted NSUInteger values, then expand surrogate boundaries atomically.
 NSUInteger count=s->text.length;
 if(range.location==NSNotFound || range.location>=count || range.length==0)return nil;
 NSUInteger end=range.location+MIN(range.length,count-range.location);
 if(range.location>0 && CFStringIsSurrogateLowCharacter([s->text characterAtIndex:range.location]) && CFStringIsSurrogateHighCharacter([s->text characterAtIndex:range.location-1]))range.location--;
 if(end<count && CFStringIsSurrogateLowCharacter([s->text characterAtIndex:end]) && CFStringIsSurrogateHighCharacter([s->text characterAtIndex:end-1]))end++;
 range.length=end-range.location;
 if(actual)*actual=range;
 return [[[NSAttributedString alloc] initWithString:[s->text substringWithRange:range]] autorelease];
}
- (NSRect)firstRectForCharacterRange:(NSRange)range actualRange:(NSRangePointer)actual {
 VellumInputState *s=state(self);
 if(!usable(s)){if(!responds(self,_cmd)){if(actual)*actual=NSMakeRange(NSNotFound,0);return NSZeroRect;}struct objc_super p=parent(self);
#if defined(__x86_64__)
 return ((NSRect(*)(struct objc_super *,SEL,NSRange,NSRangePointer))objc_msgSendSuper_stret)(&p,_cmd,range,actual);
#else
 return ((NSRect(*)(struct objc_super *,SEL,NSRange,NSRangePointer))objc_msgSendSuper)(&p,_cmd,range,actual);
#endif
 }
 // Only the visible current insertion caret is represented; advertise that range.
 if(s->caret.size.height<=0){if(actual)*actual=NSMakeRange(NSNotFound,0);return NSZeroRect;}
 if(actual)*actual=NSMakeRange(s->caretOffset,0);
 NSRect r=s->caret;
 r.size.width=0; // A zero-length actual range represents the insertion point.
 if(!self.isFlipped)r.origin.y=self.bounds.size.height-r.origin.y-r.size.height;
 r=[self convertRect:r toView:nil];
 return [self.window convertRectToScreen:r];
}
- (NSUInteger)characterIndexForPoint:(NSPoint)point {if(usable(state(self)))return NSNotFound;if(!responds(self,_cmd))return NSNotFound;struct objc_super p=parent(self);return ((NSUInteger(*)(struct objc_super *,SEL,NSPoint))objc_msgSendSuper)(&p,_cmd,point);}
@end

void *vellum_input_attach(void *view, uintptr_t handle) {
 NSCAssert(NSThread.isMainThread,@"Input attachment requires main thread");
 if(!view || ![(NSView *)view window] || state((id)view))return NULL;
 VellumInputState *s=[[VellumInputState alloc] init];s->view=(NSView *)view;s->handle=handle;
 s->baseClass=object_getClass(s->view);s->text=[[NSMutableString alloc] init];s->marked=NSMakeRange(NSNotFound,0);
 NSString *name=[NSString stringWithFormat:@"VellumInput_%s",class_getName(s->baseClass)];
 s->subclass=objc_getClass(name.UTF8String);
 if(!s->subclass){
  s->subclass=objc_allocateClassPair(s->baseClass,name.UTF8String,0);
  if(!s->subclass){[s release];return NULL;}
  unsigned int count=0;Method *methods=class_copyMethodList(VellumInputMethods.class,&count);
  for(unsigned int i=0;i<count;i++)class_addMethod(s->subclass,method_getName(methods[i]),method_getImplementation(methods[i]),method_getTypeEncoding(methods[i]));
  free(methods);class_addProtocol(s->subclass,@protocol(NSTextInputClient));objc_registerClassPair(s->subclass);
 }
 if(class_getSuperclass(s->subclass)!=s->baseClass){[s release];return NULL;}
 objc_setAssociatedObject(s->view,&inputKey,s,OBJC_ASSOCIATION_RETAIN_NONATOMIC);
 object_setClass(s->view,s->subclass);
 for(NSNotificationName name in @[NSWindowDidMoveNotification,NSWindowDidResizeNotification,NSWindowDidChangeScreenNotification])[[NSNotificationCenter defaultCenter] addObserver:s selector:@selector(windowGeometryChanged:) name:name object:s->view.window];
 return s;
}
void vellum_input_update(void *input,const char *text,int anchor,int focus,int markedStart,int markedEnd,bool active,uint64_t acknowledged,double x,double y,double width,double height) {
 VellumInputState *s=(id)input;
 NSCAssert(NSThread.isMainThread,@"Input update requires main thread");
 if(!s || s->closed)return;
 NSString *value=[NSString stringWithUTF8String:text];
 if(!value || !characters(value) || anchor<0 || focus<0)return;
 NSUInteger a=(NSUInteger)anchor,b=(NSUInteger)focus;
 NSRange selection=NSMakeRange(MIN(a,b),MAX(a,b)-MIN(a,b));
 NSRange mark=NSMakeRange(NSNotFound,0);
 if(!validRange(value,selection))return;
 if(markedStart==-1 && markedEnd==-1){}else{
  if(markedStart<0 || markedEnd<markedStart)return;
  mark=NSMakeRange((NSUInteger)markedStart,(NSUInteger)markedEnd-(NSUInteger)markedStart);
  if(!validRange(value,mark) || selection.location<mark.location || NSMaxRange(selection)>NSMaxRange(mark))return;
 }
 if(acknowledged<s->sequence)return; // Reject stale text AND its geometry/activation.
 // Validate the complete publication before changing any state. A refused
 // focus-loss commit leaves composition active until credit is available.
 if(s->active && !active){if(!finish(s,NO))return;[[s->view inputContext] discardMarkedText];}
 s->active=active;s->caret=NSMakeRect(x,y,width,height);
 [[s->view inputContext] invalidateCharacterCoordinates];
 if(acknowledged<s->sequence)return;
 // External edits invalidate the whole-document restoration baseline. Retain
 // their marked range but refuse cancellation until it is unmarked; otherwise
 // Escape could restore a document older than an accepted application edit.
 if(mark.location!=NSNotFound && s->before && (![s->text isEqualToString:value] || !NSEqualRanges(s->marked,mark))){[s->before release];s->before=nil;}
 [s->text setString:value];s->selected=selection;s->caretOffset=focus;
 if(mark.location==NSNotFound){BOOL discard=composing(s);s->marked=mark;[s->before release];s->before=nil;if(discard)[[s->view inputContext] discardMarkedText];}
 else{s->marked=mark;}

 [[s->view inputContext] invalidateCharacterCoordinates];
}
void vellum_input_close(void *input) {
 VellumInputState *s=(id)input;if(!s)return;
 NSCAssert(NSThread.isMainThread,@"Input close requires main thread");
 s->closed=YES;s->active=NO;[[s->view inputContext] discardMarkedText];
 NSCAssert(object_getClass(s->view)==s->subclass,@"Unexpected native subclass changed attachment order");
 object_setClass(s->view,s->baseClass);objc_setAssociatedObject(s->view,&inputKey,nil,OBJC_ASSOCIATION_RETAIN_NONATOMIC);[s release];
}
void vellum_input_caret(void *input,double x,double y,double width,double height){
 VellumInputState *s=(id)input;if(!s || s->closed)return;
 NSCAssert(NSThread.isMainThread,@"Input geometry requires main thread");
 s->caret=NSMakeRect(x,y,width,height);[[s->view inputContext] invalidateCharacterCoordinates];
}
char *vellum_input_source(void) {
 TISInputSourceRef source=TISCopyCurrentKeyboardInputSource();
 if(!source)return NULL;
 CFStringRef name=TISGetInputSourceProperty(source,kTISPropertyInputSourceID);
 char result[512]={0};if(name)CFStringGetCString(name,result,sizeof(result),kCFStringEncodingUTF8);CFRelease(source);
 return strdup(result);
}
char *vellum_input_probe(void *input,int step) {
 VellumInputState *s=(id)input;if(!s || !s->active)return NULL;
 id<NSTextInputClient> client=(id)s->view;
 if(step==1)[client setMarkedText:@"に" selectedRange:NSMakeRange(1,0) replacementRange:NSMakeRange(NSNotFound,0)];
 if(step==2)[client setMarkedText:@"日本" selectedRange:NSMakeRange(0,2) replacementRange:NSMakeRange(NSNotFound,0)];
 if(step==3)[client insertText:@"日本" replacementRange:NSMakeRange(NSNotFound,0)];
 if(step==4)[client setMarkedText:@"取消" selectedRange:NSMakeRange(2,0) replacementRange:NSMakeRange(NSNotFound,0)];
 if(step==5)[client doCommandBySelector:@selector(cancelOperation:)];
 if(step==6)[client setMarkedText:@"中文" selectedRange:NSMakeRange(2,0) replacementRange:NSMakeRange(1,2)];
 if(step==7)[client unmarkText];
 if(step==8)[client insertText:@"BAD" replacementRange:NSMakeRange(4,0)];
 if(step==9){[client setMarkedText:@"字" selectedRange:NSMakeRange(1,0) replacementRange:NSMakeRange(NSNotFound,0)];[client setMarkedText:@"汉" selectedRange:NSMakeRange(1,0) replacementRange:NSMakeRange(NSNotFound,0)];[client insertText:@"漢" replacementRange:NSMakeRange(NSNotFound,0)];}
 NSRange selection=[client selectedRange],marked=[client markedRange],actual;
 NSRect rect=[client firstRectForCharacterRange:selection actualRange:&actual];
 NSRange all=NSMakeRange(0,s->text.length),substringRange;
 NSAttributedString *substring=[client attributedSubstringForProposedRange:all actualRange:&substringRange];
 char *source=vellum_input_source();
 NSDictionary *report=@{@"step":@(step),@"sequence":@(s->sequence),@"text":s->text,@"selected_start":@(selection.location),@"selected_length":@(selection.length),@"marked_start":marked.location==NSNotFound ? @(-1):@(marked.location),@"marked_length":@(marked.length),@"has_marked":@([client hasMarkedText]),@"caret_screen":@[@(rect.origin.x),@(rect.origin.y),@(rect.size.width),@(rect.size.height)],@"caret_view":@[@(s->caret.origin.x),@(s->caret.origin.y),@(s->caret.size.width),@(s->caret.size.height)],@"actual_caret_start":actual.location==NSNotFound ? @(-1) : @(actual.location),@"substring":substring.string ?: @"",@"input_source":@(source ?: "unknown")};
 free(source);
 NSData *data=[NSJSONSerialization dataWithJSONObject:report options:0 error:NULL];
 return strdup([[[NSString alloc] initWithData:data encoding:NSUTF8StringEncoding] autorelease].UTF8String);
}
char *vellum_input_environment(void) {
 NSDictionary *filter=@{(id)kTISPropertyInputSourceCategory:(id)kTISCategoryKeyboardInputSource};
 CFArrayRef enabled=TISCreateInputSourceList((CFDictionaryRef)filter,false);
 NSMutableArray *sources=[NSMutableArray array];
 if(enabled){for(CFIndex i=0;i<CFArrayGetCount(enabled);i++){
  TISInputSourceRef source=(TISInputSourceRef)CFArrayGetValueAtIndex(enabled,i);
  id identifier=(id)TISGetInputSourceProperty(source,kTISPropertyInputSourceID);
  id name=(id)TISGetInputSourceProperty(source,kTISPropertyLocalizedName);
  [sources addObject:@{@"id":identifier ?: @"",@"name":name ?: @""}];
 }}
 char *current=vellum_input_source();
 NSData *data=[NSJSONSerialization dataWithJSONObject:@{@"current":@(current ?: "unknown"),@"enabled":sources} options:0 error:NULL];
 free(current);if(enabled)CFRelease(enabled);
 return strdup([[[NSString alloc] initWithData:data encoding:NSUTF8StringEncoding] autorelease].UTF8String);
}

char *vellum_input_inactive_probe(void *input){
 VellumInputState *s=(id)input;if(!s || s->active)return NULL;
 id<NSTextInputClient>client=(id)s->view;NSRange actual;
 [client selectedRange];[client markedRange];[client hasMarkedText];[client validAttributesForMarkedText];
 [client attributedSubstringForProposedRange:NSMakeRange(0,0) actualRange:&actual];
 [client firstRectForCharacterRange:NSMakeRange(0,0) actualRange:&actual];
 [client characterIndexForPoint:NSMakePoint(0,0)];
 [client setMarkedText:@"ignored" selectedRange:NSMakeRange(0,0) replacementRange:NSMakeRange(NSNotFound,0)];
 [client insertText:@"ignored" replacementRange:NSMakeRange(NSNotFound,0)];[client unmarkText];
 return strdup("inactive_optional_super_queries_passed");
}
