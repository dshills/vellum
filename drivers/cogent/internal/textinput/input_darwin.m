#import <AppKit/AppKit.h>
#import <Carbon/Carbon.h>
#import <objc/runtime.h>
#import <objc/message.h>
#include "input.h"
#include <string.h>
#include <math.h>

extern bool goVellumManagedPreflight(uintptr_t, uint64_t, uint64_t, char *, char *, char *, int, int, int, int);
extern bool goVellumManagedPreflightDelta(uintptr_t,uint64_t,uint64_t,char *,char *,char *,int,int,int,int,int,int);
extern void goVellumManagedCommit(uintptr_t);
static char inputKey, inputBaseClassKey;
@interface VellumManagedInputState : NSObject {
@public
 NSView *view;
 Class baseClass, subclass;
 uintptr_t handle;
 uint64_t sequence, clientID, requestedLimitClient;
 NSUInteger maximum, requestedMaximum;
 BOOL deltas, active, forwarding, forwarded, closed;
 NSEvent *keyEvent;
 NSMutableString *text;
 NSString *before;
 NSRange selected, marked, beforeSelected;
 NSRect caret;
 NSUInteger caretOffset, beforeCaretOffset;
 NSData *cells, *carets; NSRect viewport; int focusLine; BOOL geometryClient;
}
@end
@implementation VellumManagedInputState
- (void)windowGeometryChanged:(NSNotification *)notification {if(!closed)[[view inputContext] invalidateCharacterCoordinates];}
- (void)dealloc { [[NSNotificationCenter defaultCenter] removeObserver:self]; [text release]; [before release]; [cells release]; [carets release]; [view release]; [super dealloc]; }
@end
static VellumManagedInputState *state(id view) { return objc_getAssociatedObject(view, &inputKey); }
// A later isa layer may outlive our attachment. Find the original parent of
// the inherited input methods rather than redispatching through our own class.
static Class inputParentClass(id view) {
 VellumManagedInputState *s=state(view);if(s)return s->baseClass;
 for(Class c=object_getClass(view);c;c=class_getSuperclass(c)){
  Class base=(Class)objc_getAssociatedObject((id)c,&inputBaseClassKey);if(base)return base;
 }
 return class_getSuperclass(object_getClass(view));
}
static struct objc_super parent(id view) { return (struct objc_super){view,inputParentClass(view)}; }
static BOOL responds(id view,SEL selector){Class base=inputParentClass(view);return base && class_getInstanceMethod(base,selector)!=NULL;}
static BOOL usable(VellumManagedInputState *s) { return s && s->active && !s->closed && !s->forwarding; }
static BOOL composing(VellumManagedInputState *s){return s && (s->before || s->marked.location!=NSNotFound);}
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
// The UTF-16 count is a lower bound on UTF-8 bytes, so refuse oversized
// incoming objects before encoding or allocating a prospective full document.
static NSString *characters(id value,NSUInteger maximum) {
 NSString *text=[value isKindOfClass:NSAttributedString.class] ? [value string] : value;
 if(![text isKindOfClass:NSString.class] || text.length>maximum)return nil;
 NSData *utf8=[text dataUsingEncoding:NSUTF8StringEncoding];
 if(!utf8 || utf8.length>maximum || memchr(utf8.bytes,0,utf8.length))return nil;
 return text;
}
static BOOL fitsReplacement(VellumManagedInputState *s,NSRange range,NSString *insert) {
 NSUInteger existing=[s->text lengthOfBytesUsingEncoding:NSUTF8StringEncoding];
 NSUInteger removed=[[s->text substringWithRange:range] lengthOfBytesUsingEncoding:NSUTF8StringEncoding];
 NSUInteger added=[insert lengthOfBytesUsingEncoding:NSUTF8StringEncoding];
 if(removed>existing || added>s->maximum || existing-removed>s->maximum-added){NSBeep();return NO;}
 return YES;
}
// Capacity is reserved against the prospective snapshot BEFORE any accepted
// native text/selection/composition mutation. Refusal leaves state unchanged.
static BOOL admitReplacement(VellumManagedInputState *s,const char *kind,NSString *next,NSRange selected,NSRange marked,NSUInteger focus,BOOL known,NSRange range,NSString *replacement) {
 if(!s || s->closed || s->sequence==UINT64_MAX)return NO;
 int start=marked.location==NSNotFound ? -1 : (int)marked.location;
 int end=start<0 ? -1 : start+(int)marked.length;
 NSUInteger anchor=focus==selected.location ? NSMaxRange(selected) : selected.location;
 BOOL accepted=NO;
 if(s->deltas){
  if([next lengthOfBytesUsingEncoding:NSUTF8StringEncoding]>s->maximum){NSBeep();return NO;}
  NSUInteger prefix=0,ae=0,be=0;NSString *inserted;
  if(known){prefix=range.location;ae=NSMaxRange(range);inserted=replacement;}
  else{
   // Unmark has no visible delta. Cancellation compares the baseline using
   // buffered CoreFoundation units, without per-character Objective-C calls.
   ae=s->text.length;be=next.length;
   if(next==s->text){ae=be=0;}
   else{
    CFStringInlineBuffer a,b;
    CFStringInitInlineBuffer((CFStringRef)s->text,&a,CFRangeMake(0,ae));
    CFStringInitInlineBuffer((CFStringRef)next,&b,CFRangeMake(0,be));
    while(prefix<ae && prefix<be && CFStringGetCharacterFromInlineBuffer(&a,prefix)==CFStringGetCharacterFromInlineBuffer(&b,prefix))prefix++;
    if(prefix>0 && prefix<ae && CFStringIsSurrogateLowCharacter(CFStringGetCharacterFromInlineBuffer(&a,prefix)))prefix--;
    while(ae>prefix && be>prefix && CFStringGetCharacterFromInlineBuffer(&a,ae-1)==CFStringGetCharacterFromInlineBuffer(&b,be-1)){ae--;be--;}
    if(ae<s->text.length && ae>prefix && CFStringIsSurrogateLowCharacter(CFStringGetCharacterFromInlineBuffer(&a,ae))){ae++;be++;}
   }
   inserted=[next substringWithRange:NSMakeRange(prefix,be-prefix)];
  }
  NSString *removed=[s->text substringWithRange:NSMakeRange(prefix,ae-prefix)];
  if([removed lengthOfBytesUsingEncoding:NSUTF8StringEncoding]>(4u<<20) || [inserted lengthOfBytesUsingEncoding:NSUTF8StringEncoding]>(4u<<20)){NSBeep();return NO;}
  accepted=goVellumManagedPreflightDelta(s->handle,s->sequence+1,s->clientID,(char *)kind,(char *)inserted.UTF8String,(char *)removed.UTF8String,(int)prefix,(int)ae,(int)anchor,(int)focus,start,end);
 }else accepted=goVellumManagedPreflight(s->handle,s->sequence+1,s->clientID,(char *)kind,(char *)next.UTF8String,(char *)s->text.UTF8String,(int)anchor,(int)focus,start,end);
 if(!accepted){
  NSBeep();return NO; // Explicit audible refusal; caller can retry after drain.
 }
 return YES;
}
static BOOL admit(VellumManagedInputState *s,const char *kind,NSString *next,NSRange selected,NSRange marked,NSUInteger focus) {
 return admitReplacement(s,kind,next,selected,marked,focus,NO,NSMakeRange(0,0),nil);
}
static void emit(VellumManagedInputState *s) {
 // A native edit precedes its shaped layout acknowledgement. Old text geometry
 // must not answer queries for the new text, even during synchronous IME calls.
 [s->cells release];s->cells=nil;[s->carets release];s->carets=nil;
 s->sequence++;goVellumManagedCommit(s->handle);
}
static BOOL finish(VellumManagedInputState *s, BOOL cancel) {
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
 VellumManagedInputState *s=state(view);
 if(!s || !s->keyEvent || s->forwarded)return;
 s->forwarded=YES;s->forwarding=YES;
 struct objc_super p=parent(view);
 ((void(*)(struct objc_super *,SEL,id))objc_msgSendSuper)(&p,@selector(keyDown:),s->keyEvent);
 s->forwarding=NO;
}
// NSIntersectionRect treats zero-width insertion/newline rectangles as empty.
static NSRect clipTextRect(NSRect r,NSRect viewport){
 if(r.size.width==0){if(r.origin.x<NSMinX(viewport) || r.origin.x>NSMaxX(viewport))return NSZeroRect;CGFloat end=MIN(NSMaxY(r),NSMaxY(viewport));r.origin.y=MAX(NSMinY(r),NSMinY(viewport));r.size.height=MAX(0,end-r.origin.y);return r;}
 return NSIntersectionRect(r,viewport);
}
@interface VellumManagedInputMethods : NSView <NSTextInputClient>
@end
@implementation VellumManagedInputMethods
- (void)keyDown:(NSEvent *)event {
 VellumManagedInputState *s=state(self);
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
 VellumManagedInputState *s=state(self);
 if(!usable(s)){if(!responds(self,_cmd)){return;}struct objc_super p=parent(self);((void(*)(struct objc_super *,SEL,SEL))objc_msgSendSuper)(&p,_cmd,selector);return;}
 if(selector==@selector(cancelOperation:) && composing(s)){if(finish(s,YES))[[self inputContext] discardMarkedText];return;}
 if(composing(s)){if(!finish(s,NO)){s->keyEvent=nil;return;}[[self inputContext] discardMarkedText];}forwardKey(self);
}
- (void)setMarkedText:(id)value selectedRange:(NSRange)selection replacementRange:(NSRange)replacement {
 VellumManagedInputState *s=state(self);
 if(!usable(s)){if(!responds(self,_cmd)){return;}struct objc_super p=parent(self);((void(*)(struct objc_super *,SEL,id,NSRange,NSRange))objc_msgSendSuper)(&p,_cmd,value,selection,replacement);return;}
 NSString *insert=characters(value,s->maximum);if(!insert || !validRange(insert,selection))return;
 NSRange range=replacement.location==NSNotFound ? (s->marked.location!=NSNotFound ? s->marked : s->selected) : replacement;
 if(!validRange(s->text,range) || !fitsReplacement(s,range,insert))return;
 NSMutableString *next=[[s->text mutableCopy] autorelease];
 [next replaceCharactersInRange:range withString:insert];
 NSRange mark=NSMakeRange(range.location,insert.length),sel=NSMakeRange(range.location+selection.location,selection.length);
 if(!admitReplacement(s,"mark",next,sel,mark,NSMaxRange(sel),YES,range,insert))return;
 if(!composing(s)){s->before=[s->text copy];s->beforeSelected=s->selected;s->beforeCaretOffset=s->caretOffset;}
 [s->text setString:next];
 s->marked=NSMakeRange(range.location,insert.length);
 s->selected=NSMakeRange(range.location+selection.location,selection.length);
 s->caretOffset=NSMaxRange(s->selected);emit(s);
}
- (void)insertText:(id)value replacementRange:(NSRange)replacement {
 VellumManagedInputState *s=state(self);
 if(!usable(s)){if(!responds(self,_cmd)){return;}struct objc_super p=parent(self);((void(*)(struct objc_super *,SEL,id,NSRange))objc_msgSendSuper)(&p,_cmd,value,replacement);return;}
 NSString *insert=characters(value,s->maximum);if(!insert)return;
 NSRange range=replacement.location==NSNotFound ? (s->marked.location!=NSNotFound ? s->marked : s->selected) : replacement;
 if(!validRange(s->text,range) || !fitsReplacement(s,range,insert))return;
 NSMutableString *next=[[s->text mutableCopy] autorelease];
 [next replaceCharactersInRange:range withString:insert];
 NSRange sel=NSMakeRange(range.location+insert.length,0);
 if(!admitReplacement(s,"insert",next,sel,NSMakeRange(NSNotFound,0),sel.location,YES,range,insert))return;
 [s->text setString:next];s->selected=sel;
 s->marked=NSMakeRange(NSNotFound,0);[s->before release];s->before=nil;
 s->caretOffset=NSMaxRange(s->selected);emit(s);
}
- (void)unmarkText {
 VellumManagedInputState *s=state(self);
 if(usable(s)){finish(s,NO);return;}
 if(!responds(self,_cmd))return;
 struct objc_super p=parent(self);((void(*)(struct objc_super *,SEL))objc_msgSendSuper)(&p,_cmd);
}
- (BOOL)hasMarkedText {VellumManagedInputState *s=state(self);if(usable(s))return s->marked.location!=NSNotFound && s->marked.length>0;if(!responds(self,_cmd))return NO;struct objc_super p=parent(self);return ((BOOL(*)(struct objc_super *,SEL))objc_msgSendSuper)(&p,_cmd);}
- (NSRange)markedRange {VellumManagedInputState *s=state(self);if(usable(s))return s->marked;if(!responds(self,_cmd))return NSMakeRange(NSNotFound,0);struct objc_super p=parent(self);return ((NSRange(*)(struct objc_super *,SEL))objc_msgSendSuper)(&p,_cmd);}
- (NSRange)selectedRange {VellumManagedInputState *s=state(self);if(usable(s))return s->selected;if(!responds(self,_cmd))return NSMakeRange(NSNotFound,0);struct objc_super p=parent(self);return ((NSRange(*)(struct objc_super *,SEL))objc_msgSendSuper)(&p,_cmd);}
- (NSArray *)validAttributesForMarkedText {if(usable(state(self)))return @[NSUnderlineStyleAttributeName,NSMarkedClauseSegmentAttributeName];if(!responds(self,_cmd))return @[];struct objc_super p=parent(self);return ((id(*)(struct objc_super *,SEL))objc_msgSendSuper)(&p,_cmd);}
- (NSAttributedString *)attributedSubstringForProposedRange:(NSRange)range actualRange:(NSRangePointer)actual {
 VellumManagedInputState *s=state(self);
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
 VellumManagedInputState *s=state(self);
 if(!usable(s)){if(!responds(self,_cmd)){if(actual)*actual=NSMakeRange(NSNotFound,0);return NSZeroRect;}struct objc_super p=parent(self);
#if defined(__x86_64__)
 return ((NSRect(*)(struct objc_super *,SEL,NSRange,NSRangePointer))objc_msgSendSuper_stret)(&p,_cmd,range,actual);
#else
 return ((NSRect(*)(struct objc_super *,SEL,NSRange,NSRangePointer))objc_msgSendSuper)(&p,_cmd,range,actual);
#endif
 }
 NSRect r=NSZeroRect;NSRange represented=NSMakeRange(NSNotFound,0);
 if(s->geometryClient){
  if(!s->cells || !s->carets || range.location==NSNotFound || range.location>s->text.length){if(actual)*actual=represented;return r;}
  const VellumTextCell *cells=s->cells.bytes;NSUInteger nc=s->cells.length/sizeof(*cells);
  if(range.length==0){
   NSUInteger offset=range.location;
   for(NSUInteger i=0;i<nc;i++)if((NSUInteger)cells[i].start<offset && offset<(NSUInteger)cells[i].end){offset=cells[i].start;break;}
   const VellumTextCaret *carets=s->carets.bytes;NSUInteger nk=s->carets.length/sizeof(*carets);const VellumTextCaret *best=NULL;
   for(NSUInteger i=0;i<nk;i++)if((NSUInteger)carets[i].offset==offset){
    if(!best || carets[i].line>best->line)best=&carets[i];
    if(offset==s->caretOffset && carets[i].line==s->focusLine){best=&carets[i];break;}
   }
   if(best){r=NSMakeRect(best->x,best->y,0,best->height);if(r.origin.x<s->viewport.origin.x || r.origin.x>NSMaxX(s->viewport) || NSMaxY(r)<=NSMinY(s->viewport) || NSMinY(r)>=NSMaxY(s->viewport))r=NSZeroRect;
    else{r.origin.y=MAX(NSMinY(r),NSMinY(s->viewport));r.size.height=MIN(best->y+best->height,NSMaxY(s->viewport))-r.origin.y;represented=NSMakeRange(offset,0);}}
  }else{
   NSUInteger end=range.location+MIN(range.length,s->text.length-range.location);const VellumTextCell *first=NULL;
   for(NSUInteger i=0;i<nc;i++)if((NSUInteger)cells[i].end>range.location && (NSUInteger)cells[i].start<end && clipTextRect(NSMakeRect(cells[i].x,cells[i].y,cells[i].width,cells[i].height),s->viewport).size.height>0 && (!first || cells[i].start<first->start))first=&cells[i];
   if(first){
    r=NSMakeRect(first->x,first->y,first->width,first->height);represented=NSMakeRange(first->start,first->end-first->start);
    // One contiguous fragment in logical order. A wrap or a physically disjoint
    // bidi fragment terminates actualRange, letting Cocoa ask for the remainder.
    for(NSUInteger i=(NSUInteger)(first-cells)+1;i<nc;i++){
     const VellumTextCell *c=&cells[i];if(c->start<(int)NSMaxRange(represented))continue;
     if(c->line!=first->line || c->start!=(int)NSMaxRange(represented) || (NSUInteger)c->start>=end)break;
     NSRect next=NSMakeRect(c->x,c->y,c->width,c->height);
     if(clipTextRect(next,s->viewport).size.height<=0)break;
     if(NSMinX(next)>NSMaxX(r)+0.01 || NSMaxX(next)<NSMinX(r)-0.01)break;
     r=NSUnionRect(r,next);represented.length=c->end-represented.location;
    }
    r=clipTextRect(r,s->viewport);if(r.size.height<=0)represented=NSMakeRange(NSNotFound,0);
   }
  }
 }else if(s->caret.size.height>0){r=s->caret;r.size.width=0;represented=NSMakeRange(s->caretOffset,0);}
 if(actual)*actual=represented;if(represented.location==NSNotFound)return NSZeroRect;
 if(!self.isFlipped)r.origin.y=self.bounds.size.height-r.origin.y-r.size.height;
 r=[self convertRect:r toView:nil];return [self.window convertRectToScreen:r];
}
- (NSUInteger)characterIndexForPoint:(NSPoint)point {
 VellumManagedInputState *s=state(self);
 if(usable(s)){
  if(!s->geometryClient || !s->cells)return NSNotFound;
  NSPoint p=[self convertPoint:[self.window convertPointFromScreen:point] fromView:nil];
  if(!self.isFlipped)p.y=self.bounds.size.height-p.y;
  if(!NSPointInRect(p,s->viewport))return NSNotFound;
  const VellumTextCell *cells=s->cells.bytes;NSUInteger n=s->cells.length/sizeof(*cells);
  for(NSUInteger i=0;i<n;i++)if(NSPointInRect(p,NSMakeRect(cells[i].x,cells[i].y,cells[i].width,cells[i].height)))return cells[i].start;
  return NSNotFound;
 }
 if(!responds(self,_cmd))return NSNotFound;struct objc_super p=parent(self);return ((NSUInteger(*)(struct objc_super *,SEL,NSPoint))objc_msgSendSuper)(&p,_cmd,point);
}
@end

// Go-to-Cocoa calls do not inherit GLFW event pools. Each bridge below drains
// local temporaries while explicit retained state and malloc replies survive.
void *vellum_managed_input_attach(void *view, uintptr_t handle) {
 @autoreleasepool {
 NSCAssert(NSThread.isMainThread,@"Input attachment requires main thread");
 if(!view || ![(NSView *)view window] || state((id)view))return NULL;
 // A foreign isa layer can keep retired input methods in its hierarchy.
 // Refuse stacking another attachment until that layer is removed by its owner.
 for(Class c=object_getClass((id)view);c;c=class_getSuperclass(c))if(objc_getAssociatedObject((id)c,&inputBaseClassKey))return NULL;
 VellumManagedInputState *s=[[VellumManagedInputState alloc] init];s->view=[(NSView *)view retain];s->handle=handle;
 s->maximum=4u<<20;s->requestedMaximum=4u<<20;
 s->baseClass=object_getClass(s->view);s->text=[[NSMutableString alloc] init];s->marked=NSMakeRange(NSNotFound,0);
 NSString *name=[NSString stringWithFormat:@"VellumManagedInput_%s",class_getName(s->baseClass)];
 s->subclass=objc_getClass(name.UTF8String);
 if(!s->subclass){
  s->subclass=objc_allocateClassPair(s->baseClass,name.UTF8String,0);
  if(!s->subclass){[s release];return NULL;}
  unsigned int count=0;Method *methods=class_copyMethodList(VellumManagedInputMethods.class,&count);
  for(unsigned int i=0;i<count;i++)class_addMethod(s->subclass,method_getName(methods[i]),method_getImplementation(methods[i]),method_getTypeEncoding(methods[i]));
  free(methods);class_addProtocol(s->subclass,@protocol(NSTextInputClient));objc_registerClassPair(s->subclass);
 }
 if(!s->subclass || class_getSuperclass(s->subclass)!=s->baseClass){[s release];return NULL;}
 objc_setAssociatedObject((id)s->subclass,&inputBaseClassKey,(id)s->baseClass,OBJC_ASSOCIATION_ASSIGN);
 objc_setAssociatedObject(s->view,&inputKey,s,OBJC_ASSOCIATION_RETAIN_NONATOMIC);
 object_setClass(s->view,s->subclass);
 for(NSNotificationName name in @[NSWindowDidMoveNotification,NSWindowDidResizeNotification,NSWindowDidChangeScreenNotification])[[NSNotificationCenter defaultCenter] addObserver:s selector:@selector(windowGeometryChanged:) name:name object:s->view.window];
 return s;

 }
}
static BOOL publish(void *input,const char *text,int anchor,int focus,int markedStart,int markedEnd,bool active,uint64_t acknowledged,double x,double y,double width,double height, NSData *cells, NSData *carets, NSRect viewport, int focusLine, BOOL geometryClient, uint64_t clientID) {
 @autoreleasepool {
 VellumManagedInputState *s=(id)input;
 NSCAssert(NSThread.isMainThread,@"Input update requires main thread");
 if(!s || s->closed)return NO;
 if(!text || strnlen(text,(4u<<20)+1)>(4u<<20))return NO;
 NSString *value=[NSString stringWithUTF8String:text];
 if(!value || !characters(value,4u<<20) || anchor<0 || focus<0 || !isfinite(x) || !isfinite(y) || !isfinite(width) || !isfinite(height) || width<0 || height<0)return NO;
 NSUInteger a=(NSUInteger)anchor,b=(NSUInteger)focus;
 NSRange selection=NSMakeRange(MIN(a,b),MAX(a,b)-MIN(a,b));
 NSRange mark=NSMakeRange(NSNotFound,0);
 if(!validRange(value,selection))return NO;
 if(markedStart==-1 && markedEnd==-1){}else{
  if(markedStart<0 || markedEnd<markedStart)return NO;
  mark=NSMakeRange((NSUInteger)markedStart,(NSUInteger)markedEnd-(NSUInteger)markedStart);
  if(!validRange(value,mark) || selection.location<mark.location || NSMaxRange(selection)>NSMaxRange(mark))return NO;
 }
 if(acknowledged<s->sequence)return NO; // Reject stale text AND its geometry/activation.
 // Validate the complete publication before changing any state. A refused
 // focus-loss commit leaves composition active until credit is available.
 if(s->active && (!active || clientID!=s->clientID)){if(!finish(s,NO))return NO;[[s->view inputContext] discardMarkedText];}
 if(acknowledged<s->sequence)return NO;
 s->clientID=clientID;s->active=active;
 if(s->requestedLimitClient==clientID)s->maximum=s->requestedMaximum;s->caret=NSMakeRect(x,y,width,height);
 // External edits invalidate the whole-document restoration baseline. Retain
 // their marked range but refuse cancellation until it is unmarked; otherwise
 // Escape could restore a document older than an accepted application edit.
 if(mark.location!=NSNotFound && s->before && (![s->text isEqualToString:value] || !NSEqualRanges(s->marked,mark))){[s->before release];s->before=nil;}
 [s->cells release];s->cells=[cells retain];[s->carets release];s->carets=[carets retain];s->viewport=viewport;s->focusLine=focusLine;s->geometryClient=geometryClient;
 [s->text setString:value];s->selected=selection;s->caretOffset=focus;
 if(mark.location==NSNotFound){BOOL discard=composing(s);s->marked=mark;[s->before release];s->before=nil;if(discard)[[s->view inputContext] discardMarkedText];}
 else{s->marked=mark;}

 [[s->view inputContext] invalidateCharacterCoordinates];
 return YES;

 }
}
void vellum_managed_input_update(void *input,const char *text,int anchor,int focus,int markedStart,int markedEnd,bool active,uint64_t acknowledged,double x,double y,double width,double height) {
 publish(input,text,anchor,focus,markedStart,markedEnd,active,acknowledged,x,y,width,height,nil,nil,NSZeroRect,-1,NO,0);
}
static int compareCells(const void *a,const void *b){const VellumTextCell *x=a,*y=b;if(x->start!=y->start)return x->start<y->start ? -1:1;return (x->line>y->line)-(x->line<y->line);}
bool vellum_managed_input_client_geometry_update(void *input,uint64_t clientID,const char *text,int anchor,int focus,int markedStart,int markedEnd,bool active,uint64_t acknowledged,double x,double y,double width,double height,const VellumTextCell *cells,int nc,const VellumTextCaret *carets,int nk,double vx,double vy,double vw,double vh,int focusLine) {
 @autoreleasepool {
 NSCAssert(NSThread.isMainThread,@"Input geometry publication requires main thread");
 if(nc<0 || nk<0 || nc>262144 || nk>262144 || (nc && !cells) || (nk && !carets) || !isfinite(vx) || !isfinite(vy) || !isfinite(vw) || !isfinite(vh) || vw<0 || vh<=0 || !isfinite(vx+vw) || !isfinite(vy+vh))return NO;
 if(!text || strnlen(text,(4u<<20)+1)>(4u<<20))return NO;
 NSString *value=[NSString stringWithUTF8String:text];if(!value)return NO;
 for(int i=0;i<nc;i++){const VellumTextCell *c=&cells[i];if(c->start<0 || c->end<=c->start || c->line<0 || !validRange(value,NSMakeRange(c->start,c->end-c->start)) || !isfinite(c->x) || !isfinite(c->y) || !isfinite(c->width) || !isfinite(c->height) || c->width<0 || c->height<=0 || !isfinite(c->x+c->width) || !isfinite(c->y+c->height))return NO;}
 for(int i=0;i<nk;i++){const VellumTextCaret *c=&carets[i];if(c->offset<0 || c->line<0 || !validRange(value,NSMakeRange(c->offset,0)) || !isfinite(c->x) || !isfinite(c->y) || !isfinite(c->height) || c->height<=0 || !isfinite(c->y+c->height))return NO;}
 NSMutableData *ownedCells=[NSMutableData dataWithBytes:cells length:(NSUInteger)nc*sizeof(*cells)];if(nc>1)qsort(ownedCells.mutableBytes,nc,sizeof(*cells),compareCells);NSData *ownedCarets=[NSData dataWithBytes:carets length:(NSUInteger)nk*sizeof(*carets)];
 return publish(input,text,anchor,focus,markedStart,markedEnd,active,acknowledged,x,y,width,height,ownedCells,ownedCarets,NSMakeRect(vx,vy,vw,vh),focusLine,YES,clientID);

 }
}

void vellum_managed_input_close(void *input) {
 @autoreleasepool {
 VellumManagedInputState *s=(id)input;if(!s)return;
 NSCAssert(NSThread.isMainThread,@"Input close requires main thread");
 s->closed=YES;s->active=NO;[[s->view inputContext] discardMarkedText];
 [[NSNotificationCenter defaultCenter] removeObserver:s];
 NSView *view=s->view;
 // Preserve a newer KVO/observer layer. Our inherited methods become inert
 // after removing state and use permanent original-parent class metadata.
 if(object_getClass(view)==s->subclass)object_setClass(view,s->baseClass);
 if(state(view)==s)objc_setAssociatedObject(view,&inputKey,nil,OBJC_ASSOCIATION_RETAIN_NONATOMIC);
 s->view=nil;[view release];[s release];

 }
}
void vellum_managed_input_caret(void *input,double x,double y,double width,double height){
 @autoreleasepool {
 VellumManagedInputState *s=(id)input;if(!s || s->closed)return;
 NSCAssert(NSThread.isMainThread,@"Input geometry requires main thread");
 s->caret=NSMakeRect(x,y,width,height);[[s->view inputContext] invalidateCharacterCoordinates];

 }
}
char *vellum_managed_input_source(void) {
 TISInputSourceRef source=TISCopyCurrentKeyboardInputSource();
 if(!source)return NULL;
 CFStringRef name=TISGetInputSourceProperty(source,kTISPropertyInputSourceID);
 char result[512]={0};if(name)CFStringGetCString(name,result,sizeof(result),kCFStringEncodingUTF8);CFRelease(source);
 return strdup(result);
}
char *vellum_managed_input_probe(void *input,int step) {
 @autoreleasepool {
 VellumManagedInputState *s=(id)input;if(!s || s->closed || !s->active)return NULL;
 id<NSTextInputClient> client=(id)s->view;
 if(step==10)[client insertText:[@"" stringByPaddingToLength:4097 withString:@"a" startingAtIndex:0] replacementRange:NSMakeRange(NSNotFound,0)];
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
 char *source=vellum_managed_input_source();
 NSDictionary *report=@{@"client_id":@(s->clientID),@"step":@(step),@"sequence":@(s->sequence),@"text":s->text,@"selected_start":@(selection.location),@"selected_length":@(selection.length),@"marked_start":marked.location==NSNotFound ? @(-1):@(marked.location),@"marked_length":@(marked.length),@"has_marked":@([client hasMarkedText]),@"caret_screen":@[@(rect.origin.x),@(rect.origin.y),@(rect.size.width),@(rect.size.height)],@"caret_view":@[@(s->caret.origin.x),@(s->caret.origin.y),@(s->caret.size.width),@(s->caret.size.height)],@"actual_caret_start":actual.location==NSNotFound ? @(-1) : @(actual.location),@"substring":substring.string ?: @"",@"active":@(s->active),@"cell_count":@(s->cells.length/sizeof(VellumTextCell)),@"caret_count":@(s->carets.length/sizeof(VellumTextCaret)),@"viewport":@[@(s->viewport.origin.x),@(s->viewport.origin.y),@(s->viewport.size.width),@(s->viewport.size.height)],@"input_source":@(source ?: "unknown")};
 free(source);
 NSData *data=[NSJSONSerialization dataWithJSONObject:report options:0 error:NULL];
 return strdup([[[NSString alloc] initWithData:data encoding:NSUTF8StringEncoding] autorelease].UTF8String);

 }
}
char *vellum_managed_input_environment(void) {
 @autoreleasepool {
 NSDictionary *filter=@{(id)kTISPropertyInputSourceCategory:(id)kTISCategoryKeyboardInputSource};
 CFArrayRef enabled=TISCreateInputSourceList((CFDictionaryRef)filter,false);
 NSMutableArray *sources=[NSMutableArray array];
 if(enabled){for(CFIndex i=0;i<CFArrayGetCount(enabled);i++){
  TISInputSourceRef source=(TISInputSourceRef)CFArrayGetValueAtIndex(enabled,i);
  id identifier=(id)TISGetInputSourceProperty(source,kTISPropertyInputSourceID);
  id name=(id)TISGetInputSourceProperty(source,kTISPropertyLocalizedName);
  [sources addObject:@{@"id":identifier ?: @"",@"name":name ?: @""}];
 }}
 char *current=vellum_managed_input_source();
 NSData *data=[NSJSONSerialization dataWithJSONObject:@{@"current":@(current ?: "unknown"),@"enabled":sources} options:0 error:NULL];
 free(current);if(enabled)CFRelease(enabled);
 return strdup([[[NSString alloc] initWithData:data encoding:NSUTF8StringEncoding] autorelease].UTF8String);

 }
}

char *vellum_managed_input_inactive_probe(void *input){
 @autoreleasepool {
 VellumManagedInputState *s=(id)input;if(!s || s->closed || s->active)return NULL;
 id<NSTextInputClient>client=(id)s->view;NSRange actual;
 [client selectedRange];[client markedRange];[client hasMarkedText];[client validAttributesForMarkedText];
 [client attributedSubstringForProposedRange:NSMakeRange(0,0) actualRange:&actual];
 [client firstRectForCharacterRange:NSMakeRange(0,0) actualRange:&actual];
 [client characterIndexForPoint:NSMakePoint(0,0)];
 [client setMarkedText:@"ignored" selectedRange:NSMakeRange(0,0) replacementRange:NSMakeRange(NSNotFound,0)];
 [client insertText:@"ignored" replacementRange:NSMakeRange(NSNotFound,0)];[client unmarkText];
 return strdup("inactive_optional_super_queries_passed");

 }
}

char *vellum_managed_input_query_probe(void *input,int start,int length,double x,double y){
 @autoreleasepool {
 NSCAssert(NSThread.isMainThread,@"Query probe requires main thread");
 VellumManagedInputState *s=(id)input;if(!usable(s) || start<0 || length<0)return NULL;
 id<NSTextInputClient> client=(id)s->view;
 NSRange requested=NSMakeRange((NSUInteger)start,(NSUInteger)length);
 NSRange substringRange=NSMakeRange(NSNotFound,0),rectRange=NSMakeRange(NSNotFound,0);
 NSAttributedString *substring=[client attributedSubstringForProposedRange:requested actualRange:&substringRange];
 NSRect rect=[client firstRectForCharacterRange:requested actualRange:&rectRange];
 NSUInteger index=[client characterIndexForPoint:NSMakePoint(x,y)];
 NSDictionary *report=@{@"requested":@[@(start),@(length)],@"substring":substring.string ?: @"",@"substring_range":@[substringRange.location==NSNotFound ? @(-1):@(substringRange.location),@(substringRange.length)],@"rect_range":@[rectRange.location==NSNotFound ? @(-1):@(rectRange.location),@(rectRange.length)],@"screen_rect":@[@(rect.origin.x),@(rect.origin.y),@(rect.size.width),@(rect.size.height)],@"point_index":index==NSNotFound ? @(-1):@(index)};
 NSData *data=[NSJSONSerialization dataWithJSONObject:report options:0 error:NULL];
 if(!data)return NULL;
 return strdup([[[NSString alloc]initWithData:data encoding:NSUTF8StringEncoding]autorelease].UTF8String);

 }
}

VellumScreenRect vellum_managed_input_rendered_rect(void *input,double x,double y,double width,double height,int pixel_width,int pixel_height){
 NSCAssert(NSThread.isMainThread,@"Rendered coordinate oracle requires main thread");
 VellumManagedInputState *s=(id)input;if(!s || s->closed || pixel_width<=0 || pixel_height<=0 || s->view.bounds.size.width<=0 || s->view.bounds.size.height<=0)return (VellumScreenRect){0};
 NSView *view=s->view;double sx=pixel_width/view.bounds.size.width,sy=pixel_height/view.bounds.size.height;
 NSRect r=NSMakeRect(x/sx,y/sy,width/sx,height/sy);
 if(!view.isFlipped)r.origin.y=view.bounds.size.height-r.origin.y-r.size.height;
 r=[view convertRect:r toView:nil];r=[view.window convertRectToScreen:r];
 return (VellumScreenRect){r.origin.x,r.origin.y,r.size.width,r.size.height};
}

void vellum_managed_input_geometry_update(void *input,const char *text,int anchor,int focus,int markedStart,int markedEnd,bool active,uint64_t acknowledged,double x,double y,double width,double height,const VellumTextCell *cells,int nc,const VellumTextCaret *carets,int nk,double vx,double vy,double vw,double vh,int focusLine) {vellum_managed_input_client_geometry_update(input,0,text,anchor,focus,markedStart,markedEnd,active,acknowledged,x,y,width,height,cells,nc,carets,nk,vx,vy,vw,vh,focusLine);}

bool vellum_managed_input_client_update(void *input,uint64_t clientID,const char *text,int anchor,int focus,int markedStart,int markedEnd,bool active,uint64_t acknowledged,double x,double y,double width,double height) {return publish(input,text,anchor,focus,markedStart,markedEnd,active,acknowledged,x,y,width,height,nil,nil,NSZeroRect,-1,NO,clientID);}

void vellum_managed_input_enable_deltas(void *input){VellumManagedInputState *s=(id)input;NSCAssert(NSThread.isMainThread,@"Input mode requires main thread");if(s && !s->closed && !s->active)s->deltas=YES;}

bool vellum_managed_input_settle(void *input,uint64_t clientID){
 @autoreleasepool {
 NSCAssert(NSThread.isMainThread,@"Input settlement requires main thread");
 VellumManagedInputState *s=(id)input;if(!s || s->closed)return false;
 if(s->clientID!=clientID)return true;
 if(!finish(s,NO))return false;[[s->view inputContext] discardMarkedText];return true;

 }
}
void vellum_managed_input_request_limit(void *input,uint64_t clientID,int maximum){
 NSCAssert(NSThread.isMainThread,@"Input budget requires main thread");
 VellumManagedInputState *s=(id)input;if(!s || s->closed || maximum<1 || maximum>(4u<<20))return;
 s->requestedLimitClient=clientID;s->requestedMaximum=(NSUInteger)maximum;
}

// Deterministic bridge fixtures deliberately never construct an NSWindow. This
// object supplies only the attachment token and inert optional superclass methods.
static int fixtureViews;
@interface VellumManagedFixtureView : NSObject { @public int baseCalls; }
@end
@implementation VellumManagedFixtureView
- (id)init { self=[super init];if(self)fixtureViews++;return self; }
- (void)dealloc { fixtureViews--;[super dealloc]; }
- (NSWindow *)window { return (NSWindow *)self; }
- (NSTextInputContext *)inputContext { return nil; }
- (NSRange)selectedRange { return NSMakeRange(17,0); }
- (void)insertText:(id)value replacementRange:(NSRange)range { baseCalls++; }
@end
@interface VellumManagedFixtureRejectedView : VellumManagedFixtureView
@end
@implementation VellumManagedFixtureRejectedView
@end
void *vellum_managed_input_fixture_view(void) { return [[VellumManagedFixtureView alloc]init]; }
void *vellum_managed_input_fixture_rejected_view(void) {
 // A preexisting incompatible name forces the real attachment cleanup path.
 if(!objc_getClass("VellumManagedInput_VellumManagedFixtureRejectedView")){
  Class collision=objc_allocateClassPair(NSObject.class,"VellumManagedInput_VellumManagedFixtureRejectedView",0);
  if(!collision)return NULL;objc_registerClassPair(collision);
 }
 return [[VellumManagedFixtureRejectedView alloc]init];
}
void vellum_managed_input_fixture_release(void *view) { [(id)view release]; }
int vellum_managed_input_fixture_views(void) { return fixtureViews; }
bool vellum_managed_input_fixture_overlay(void *view) {
 Class base=object_getClass((id)view);Class overlay=objc_getClass("VellumManagedFixtureObserver");
 if(!overlay){overlay=objc_allocateClassPair(base,"VellumManagedFixtureObserver",0);if(!overlay)return false;objc_registerClassPair(overlay);}
 if(class_getSuperclass(overlay)!=base)return false;object_setClass((id)view,overlay);return true;
}
bool vellum_managed_input_fixture_delegates(void *view) {
 if(object_getClass((id)view)!=objc_getClass("VellumManagedFixtureObserver"))return false;
 VellumManagedFixtureView *v=(id)view;int prior=v->baseCalls;
 NSRange range=[(id<NSTextInputClient>)v selectedRange];
 [(id<NSTextInputClient>)v insertText:@"ignored" replacementRange:NSMakeRange(0,0)];
 return range.location==17 && range.length==0 && v->baseCalls==prior+1;
}
char *vellum_managed_input_fixture_snapshot(void *input) {
 VellumManagedInputState *s=(id)input;if(!s || s->closed)return NULL;
 NSDictionary *v=@{@"text":s->text,@"sequence":@(s->sequence),@"client":@(s->clientID),@"selected":@[@(s->selected.location),@(s->selected.length)],@"marked":s->marked.location==NSNotFound ? @(-1):@(s->marked.location),@"baseline":s->before!=nil ? @YES:@NO};
 NSData *data=[NSJSONSerialization dataWithJSONObject:v options:0 error:NULL];
 return strdup([[[NSString alloc]initWithData:data encoding:NSUTF8StringEncoding]autorelease].UTF8String);
}
void vellum_managed_input_fixture_edit(void *input,const char *value,int bytes,int start,int length,int mode) {
 VellumManagedInputState *s=(id)input;if(!s || s->closed)return;
 id<NSTextInputClient> client=(id)s->view;
 NSString *insert=[[[NSString alloc]initWithBytes:value length:(NSUInteger)bytes encoding:NSUTF8StringEncoding]autorelease];
 NSRange range=start<0 ? NSMakeRange(NSNotFound,0):NSMakeRange(start,length);
 if(mode==0)[client insertText:insert replacementRange:range];
 if(mode==1)[client setMarkedText:insert selectedRange:NSMakeRange(insert.length,0) replacementRange:range];
 if(mode==2)[client doCommandBySelector:@selector(cancelOperation:)];
 if(mode==3)[client unmarkText];
}

void *vellum_managed_input_fixture_pool(void) { return [[NSAutoreleasePool alloc]init]; }
void vellum_managed_input_fixture_drain(void *pool) { [(NSAutoreleasePool *)pool drain]; }
