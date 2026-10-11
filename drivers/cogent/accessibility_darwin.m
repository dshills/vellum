//go:build darwin && cgo

#import <Cocoa/Cocoa.h>
#import <objc/runtime.h>
#import <objc/message.h>
#include <stdint.h>
#include <string.h>
#include <math.h>
#include <stdatomic.h>
extern int vellumAXAction(uint64_t token,uint64_t generation,char *id,char *kind,char *value);
extern void vellumAXResult(uint64_t request,char *json);
@class VellumAXState;
static char vellumAXKey;
static NSMutableDictionary<NSNumber *,VellumAXState *> *vellumAXStates;

@interface VellumAXElement : NSAccessibilityElement
@property (weak) VellumAXState *state;
@property (copy) NSDictionary *node;
@property (weak) id parentElement;
@property (copy) NSArray *childElements;
@property NSRect pixelFrame;
@property BOOL live;
@property BOOL lastAdmission;
@property uint64_t objectIdentity;
@property (copy) NSDictionary *textGeometry;
- (BOOL)request:(NSString *)action value:(NSString *)value;
@end
@interface VellumAXState : NSObject
@property (weak) NSView *view;
@property (weak) NSWindow *window;
@property Class viewClass;
@property Class installedViewClass;
@property uint64_t token;
@property uint64_t generation;
@property BOOL closed;
@property (strong) NSMutableDictionary<NSString *,VellumAXElement *> *elements;
@property (copy) NSArray *roots;
@property (copy) NSArray *hierarchy;
@property (strong) NSMutableDictionary<NSString *,NSArray *> *geometry;
@property (strong) NSMutableDictionary<NSString *,NSDictionary *> *textGeometry;
@end
@implementation VellumAXState
@end
static VellumAXState *vellumAXState(id object){return objc_getAssociatedObject(object,&vellumAXKey);}
static BOOL vellumAXPending(VellumAXElement *element){return element.node[@"grid_scope"]&&[element.node[@"grid_pending"] boolValue];}
static BOOL vellumAXHas(VellumAXElement *element,NSString *action){return element.live&&!element.state.closed&&!vellumAXPending(element)&&[element.node[@"actions"] containsObject:action];}
// IDs are process-local monotonic values, never addresses. Retired objects keep
// their identity; new same-ID objects cannot inherit a former identity.
static _Atomic(uint64_t) vellumAXNextIdentity=1;
static uint64_t vellumAXIdentity(void){
 uint64_t next=atomic_load(&vellumAXNextIdentity);
 while(next!=UINT64_MAX){if(atomic_compare_exchange_weak(&vellumAXNextIdentity,&next,next+1))return next;}return 0;
}
static NSString *vellumAXRole(NSString *role){
 if([role isEqual:@"label"])return NSAccessibilityStaticTextRole;
 if([role isEqual:@"button"])return NSAccessibilityButtonRole;
 if([role isEqual:@"checkbox"])return NSAccessibilityCheckBoxRole;
 if([role isEqual:@"text-field"])return NSAccessibilityTextFieldRole;
 if([role isEqual:@"multiline-text"])return NSAccessibilityTextAreaRole;
 if([role isEqual:@"number-field"])return NSAccessibilityIncrementorRole;
 if([role isEqual:@"slider"])return NSAccessibilitySliderRole;
 if([role isEqual:@"selector"])return NSAccessibilityPopUpButtonRole;
 if([role isEqual:@"tabs"])return NSAccessibilityTabGroupRole;
 if([role isEqual:@"option"])return NSAccessibilityRadioButtonRole;
 if([role isEqual:@"scroll"])return NSAccessibilityScrollAreaRole;
 if([role isEqual:@"table"])return NSAccessibilityTableRole;
 if([role isEqual:@"outline"])return NSAccessibilityOutlineRole;
 if([role isEqual:@"row"]||[role isEqual:@"tree-row"]||[role isEqual:@"grid-row"])return NSAccessibilityRowRole;
 if([role isEqual:@"grid-cell"])return NSAccessibilityCellRole;
 if([role isEqual:@"grid-column"])return NSAccessibilityColumnRole;
 return NSAccessibilityGroupRole;
}
static BOOL vellumAXValidRange(NSString *text,NSRange range){
 if(!text||range.location==NSNotFound||range.location>text.length||range.length>text.length-range.location)return NO;
 for(NSUInteger p=range.location,i=0;i<2;i++,p=NSMaxRange(range))if(p>0&&p<text.length&&CFStringIsSurrogateHighCharacter([text characterAtIndex:p-1])&&CFStringIsSurrogateLowCharacter([text characterAtIndex:p]))return NO;return YES;
}
static NSRect vellumAXArrayRect(NSArray *r){return r.count==4?NSMakeRect([r[0] doubleValue],[r[1] doubleValue],[r[2] doubleValue],[r[3] doubleValue]):NSZeroRect;}
static NSRect vellumAXClipRect(NSRect r,NSRect viewport){if(r.size.width==0){if(r.origin.x<NSMinX(viewport)||r.origin.x>NSMaxX(viewport))return NSZeroRect;CGFloat end=MIN(NSMaxY(r),NSMaxY(viewport));r.origin.y=MAX(NSMinY(r),NSMinY(viewport));r.size.height=MAX(0,end-r.origin.y);return r;}return NSIntersectionRect(r,viewport);}
@implementation VellumAXElement
- (instancetype)init{self=[super init];if(self){self.objectIdentity=vellumAXIdentity();if(!self.objectIdentity)return nil;}return self;}
- (NSString *)managedText {
 if(![self isAccessibilityEnabled]||![self.node[@"text_client"] unsignedLongLongValue]||![self.node[@"value"] isKindOfClass:NSString.class])return nil;
 return self.node[@"value"];
}
- (NSDictionary *)managedGeometry {
 NSString *text=[self managedText];NSDictionary *entry=self.textGeometry;
 if(!text||![entry[@"client"] isEqual:self.node[@"text_client"]]||![entry[@"text"] isEqual:text])return nil;
 return entry[@"geometry"];
}
- (NSInteger)accessibilityNumberOfCharacters{return [self managedText].length;}
- (NSRange)accessibilitySelectedTextRange {
 NSString *text=[self managedText];NSArray *range=self.node[@"text_selection"];
 if(!text||range.count!=2)return NSMakeRange(NSNotFound,0);
 NSRange r=NSMakeRange([range[0] unsignedIntegerValue],[range[1] unsignedIntegerValue]);
 return vellumAXValidRange(text,r)?r:NSMakeRange(NSNotFound,0);
}
- (NSString *)accessibilitySelectedText {return [self accessibilityStringForRange:self.accessibilitySelectedTextRange];}
- (void)setAccessibilitySelectedTextRange:(NSRange)range {
 NSString *text=[self managedText];if(!text||!vellumAXValidRange(text,range))return;
 NSData *data=[NSJSONSerialization dataWithJSONObject:@{@"text":text,@"start":@(range.location),@"length":@(range.length)} options:0 error:nil];
 [self request:@"selection" value:[[NSString alloc]initWithData:data encoding:NSUTF8StringEncoding]];
}
- (NSString *)accessibilityStringForRange:(NSRange)range {NSString *text=[self managedText];return text&&vellumAXValidRange(text,range)?[text substringWithRange:range]:nil;}
- (NSAttributedString *)accessibilityAttributedStringForRange:(NSRange)range {NSString *text=[self accessibilityStringForRange:range];return text?[[NSAttributedString alloc]initWithString:text]:nil;}
- (NSInteger)accessibilityLineForIndex:(NSInteger)index {
 NSString *text=[self managedText];if(!text||index<0||!vellumAXValidRange(text,NSMakeRange(index,0)))return NSNotFound;
 NSInteger line=0;for(NSUInteger i=0;i<(NSUInteger)index;i++){unichar c=[text characterAtIndex:i];if(c=='\r'){if(i+1<text.length&&[text characterAtIndex:i+1]=='\n'){if(i+1>=(NSUInteger)index)break;i++;}line++;}else if(c=='\n')line++;}return line;
}
- (NSInteger)accessibilityInsertionPointLineNumber {return [self accessibilityLineForIndex:[self.node[@"text_focus"] integerValue]];}
- (NSRange)accessibilityRangeForLine:(NSInteger)line {
 NSString *text=[self managedText];if(!text||line<0)return NSMakeRange(NSNotFound,0);
 NSUInteger start=0;NSInteger current=0;for(NSUInteger i=0;i<text.length;i++){unichar c=[text characterAtIndex:i];if(c!='\n'&&c!='\r')continue;NSUInteger end=i+1;if(c=='\r'&&end<text.length&&[text characterAtIndex:end]=='\n')end++;
 if(current==line)return NSMakeRange(start,end-start);current++;start=end;i=end-1;}
 return current==line?NSMakeRange(start,text.length-start):NSMakeRange(NSNotFound,0);
}
- (NSRange)accessibilityRangeForIndex:(NSInteger)index {
 NSString *text=[self managedText];if(!text||index<0||(NSUInteger)index>=text.length||!vellumAXValidRange(text,NSMakeRange(index,0)))return NSMakeRange(NSNotFound,0);
 return [text rangeOfComposedCharacterSequenceAtIndex:index];
}
- (NSRect)managedScreenRect:(NSRect)pixel {
 NSView *view=self.state.view;NSWindow *window=self.state.window;if(!view||!window)return NSZeroRect;CGFloat scale=window.backingScaleFactor;if(scale<=0)scale=1;
 NSRect local=NSMakeRect(pixel.origin.x/scale,pixel.origin.y/scale,pixel.size.width/scale,pixel.size.height/scale);if(!view.isFlipped)local.origin.y=view.bounds.size.height-NSMaxY(local);
 return [window convertRectToScreen:[view convertRect:local toView:nil]];
}
- (NSRect)accessibilityFrameForRange:(NSRange)range {
 NSString *text=[self managedText];NSDictionary *g=[self managedGeometry];if(!g||!vellumAXValidRange(text,range))return NSZeroRect;
 NSRect viewport=vellumAXArrayRect(g[@"viewport"]),result=NSZeroRect;BOOL found=NO;
 if(range.length==0){NSDictionary *chosen=nil;NSInteger focusLine=[g[@"focus_line"] integerValue];
 for(NSDictionary *caret in g[@"carets"])if([caret[@"offset"] unsignedIntegerValue]==range.location){if(!chosen||[caret[@"line"] integerValue]==focusLine)chosen=caret;if([caret[@"line"] integerValue]==focusLine)break;}
 if(chosen){result=vellumAXClipRect(vellumAXArrayRect(chosen[@"rect"]),viewport);found=result.size.height>0;}
 }else for(NSDictionary *cell in g[@"cells"]){if([cell[@"end"] unsignedIntegerValue]<=range.location||[cell[@"start"] unsignedIntegerValue]>=NSMaxRange(range))continue;
 NSRect rect=vellumAXClipRect(vellumAXArrayRect(cell[@"rect"]),viewport);if(rect.size.height<=0)continue;
 if(!found){result=rect;found=YES;}else{CGFloat x=MIN(NSMinX(result),NSMinX(rect)),y=MIN(NSMinY(result),NSMinY(rect));result=NSMakeRect(x,y,MAX(NSMaxX(result),NSMaxX(rect))-x,MAX(NSMaxY(result),NSMaxY(rect))-y);}}
 return found?[self managedScreenRect:result]:NSZeroRect;
}
- (NSRange)accessibilityRangeForPosition:(NSPoint)point {
 NSDictionary *g=[self managedGeometry];NSView *view=self.state.view;NSWindow *window=self.state.window;if(!g||!view||!window)return NSMakeRange(NSNotFound,0);
 NSPoint p=[view convertPoint:[window convertPointFromScreen:point] fromView:nil];if(!view.isFlipped)p.y=view.bounds.size.height-p.y;CGFloat scale=window.backingScaleFactor;if(scale<=0)scale=1;p.x*=scale;p.y*=scale;
 if(!NSPointInRect(p,vellumAXArrayRect(g[@"viewport"])))return NSMakeRange(NSNotFound,0);
 for(NSDictionary *cell in g[@"cells"])if(NSPointInRect(p,vellumAXArrayRect(cell[@"rect"]))){NSUInteger start=[cell[@"start"] unsignedIntegerValue],end=[cell[@"end"] unsignedIntegerValue];return NSMakeRange(start,end-start);}return NSMakeRange(NSNotFound,0);
}

- (BOOL)isAccessibilityElement{return self.live&&!self.state.closed;}
- (NSString *)accessibilityIdentifier{return self.node[@"id"];}
- (NSString *)accessibilityRole{return vellumAXRole(self.node[@"role"]);}
- (NSString *)accessibilitySubrole{return [self.node[@"role"] isEqual:@"dialog"]?NSAccessibilityDialogSubrole:nil;}
- (NSString *)accessibilityLabel{return self.node[@"name"];}
- (NSString *)accessibilityTitle{return self.node[@"name"];}
- (NSString *)accessibilityHelp{return self.node[@"help"];}
- (id)accessibilityValue {id value=self.node[@"value"];return value==NSNull.null?nil:value;}
- (id)accessibilityMinValue{return self.node[@"min"];}
- (id)accessibilityMaxValue{return self.node[@"max"];}
- (BOOL)isAccessibilityEnabled{return [self isAccessibilityElement]&&!vellumAXPending(self)&&![self.node[@"disabled"] boolValue];}
- (BOOL)isAccessibilityFocused{return [self isAccessibilityElement]&&!vellumAXPending(self)&&[self.node[@"focused"] boolValue];}
- (BOOL)isAccessibilitySelected{return [self.node[@"selected"] boolValue];}
- (BOOL)isAccessibilityModal{return [self.node[@"role"] isEqual:@"dialog"];}
- (BOOL)isAccessibilityExpanded{return [self.node[@"expanded"] boolValue];}
- (void)setAccessibilityExpanded:(BOOL)expanded{[self request:expanded?@"expand":@"collapse" value:@""];}
- (NSInteger)accessibilityIndex{return MAX(0,[self.node[@"row_index"] integerValue]-1);}
- (NSInteger)accessibilityDisclosureLevel{return MAX(0,[self.node[@"level"] integerValue]-1);}
- (id)accessibilityParent{return self.parentElement;}
- (NSArray *)accessibilityChildren{return self.childElements?:@[];}
- (NSArray *)accessibilityVisibleChildren{return self.accessibilityChildren;}
- (NSArray *)accessibilityRows{return [self.childElements filteredArrayUsingPredicate:[NSPredicate predicateWithBlock:^BOOL(VellumAXElement *x,NSDictionary *bindings){return [x.node[@"role"] isEqual:@"row"]||[x.node[@"role"] isEqual:@"tree-row"]||[x.node[@"role"] isEqual:@"grid-row"];}]];}
- (NSInteger)accessibilityRowCount{if(self.node[@"row_count"])return [self.node[@"row_count"] integerValue];NSInteger count=0;for(VellumAXElement *row in self.accessibilityRows)count=MAX(count,[row.node[@"row_count"] integerValue]);return count;}
- (NSInteger)accessibilityColumnCount{return [self.node[@"column_count"] integerValue];}
- (BOOL)isAccessibilityOrderedByRow{return YES;}
- (NSRange)accessibilityRowIndexRange{return self.node[@"row_index"]?NSMakeRange(MAX(0,[self.node[@"row_index"] integerValue]-1),1):NSMakeRange(NSNotFound,0);}
- (NSRange)accessibilityColumnIndexRange{return self.node[@"column_index"]?NSMakeRange(MAX(0,[self.node[@"column_index"] integerValue]-1),1):NSMakeRange(NSNotFound,0);}
- (NSArray *)accessibilityVisibleRows{return self.accessibilityRows;}
- (NSArray *)accessibilityColumns{return [self.childElements filteredArrayUsingPredicate:[NSPredicate predicateWithBlock:^BOOL(VellumAXElement *x,NSDictionary *bindings){return [x.node[@"role"] isEqual:@"grid-column"];}]];}
- (NSArray *)accessibilityVisibleColumns{return self.accessibilityColumns;}
- (NSArray *)accessibilityColumnHeaderUIElements{return self.accessibilityColumns;}
- (NSArray *)accessibilityVisibleCells{NSMutableArray *cells=[NSMutableArray new];for(VellumAXElement *row in self.accessibilityRows)for(VellumAXElement *cell in row.childElements)if([cell.node[@"role"] isEqual:@"grid-cell"])[cells addObject:cell];return cells;}
- (NSArray *)accessibilitySelectedCells{return [self.accessibilityVisibleCells filteredArrayUsingPredicate:[NSPredicate predicateWithBlock:^BOOL(VellumAXElement *x,NSDictionary *bindings){return [x isAccessibilitySelected];}]];}
- (id)accessibilityCellForColumn:(NSInteger)column row:(NSInteger)row{for(VellumAXElement *cell in self.accessibilityVisibleCells)if(cell.accessibilityColumnIndexRange.location==column&&cell.accessibilityRowIndexRange.location==row)return cell;return nil;}
- (NSArray *)accessibilitySelectedChildren{return [self.childElements filteredArrayUsingPredicate:[NSPredicate predicateWithBlock:^BOOL(VellumAXElement *x,NSDictionary *bindings){return [x isAccessibilitySelected];}]];}
- (id)accessibilityWindow{return self.state.window;}
- (id)accessibilityTopLevelUIElement{return self.state.window;}
- (NSRect)accessibilityFrame {
 NSView *view=self.state.view;NSWindow *window=self.state.window;if(!view||!window||self.state.closed)return NSZeroRect;
 CGFloat scale=window.backingScaleFactor;if(scale<=0)scale=1;
 NSRect local=NSMakeRect(self.pixelFrame.origin.x/scale,self.pixelFrame.origin.y/scale,self.pixelFrame.size.width/scale,self.pixelFrame.size.height/scale);
 if(!view.isFlipped)local.origin.y=view.bounds.size.height-NSMaxY(local);
 return [window convertRectToScreen:[view convertRect:local toView:nil]];
}
- (BOOL)isAccessibilitySelectorAllowed:(SEL)selector {
 if(selector==@selector(accessibilityPerformPress))return vellumAXHas(self,@"press");
 if(selector==@selector(accessibilityPerformIncrement))return vellumAXHas(self,@"increment");
 if(selector==@selector(accessibilityPerformDecrement))return vellumAXHas(self,@"decrement");
 if(selector==@selector(accessibilityPerformCancel))return vellumAXHas(self,@"cancel");
 if(selector==@selector(setAccessibilityValue:))return vellumAXHas(self,@"set");
 if(selector==@selector(setAccessibilitySelectedTextRange:))return vellumAXHas(self,@"selection");
 if(selector==@selector(setAccessibilityFocused:))return vellumAXHas(self,@"focus");
 if(selector==@selector(setAccessibilityExpanded:))return vellumAXHas(self,@"expand")||vellumAXHas(self,@"collapse");
 return [super isAccessibilitySelectorAllowed:selector];
}
- (BOOL)request:(NSString *)action value:(NSString *)value{
 VellumAXState *state=self.state;
 if(![self isAccessibilityEnabled]||!vellumAXHas(self,action)||!state)return NO;
 BOOL accepted=vellumAXAction(state.token,state.generation,(char *)[self accessibilityIdentifier].UTF8String,(char *)action.UTF8String,(char *)(value?:@"").UTF8String)!=0;self.lastAdmission=accepted;return accepted;
}
- (BOOL)accessibilityPerformPress{return [self request:@"press" value:@""];}
- (BOOL)accessibilityPerformIncrement{return [self request:@"increment" value:@""];}
- (BOOL)accessibilityPerformDecrement{return [self request:@"decrement" value:@""];}
- (BOOL)accessibilityPerformCancel{return [self request:@"cancel" value:@""];}
- (void)setAccessibilityFocused:(BOOL)value{if(value)[self request:@"focus" value:@""];}
- (void)setAccessibilityValue:(id)value{
 NSString *text=nil;
 if([value isKindOfClass:NSString.class])text=value;
 else if([value isKindOfClass:NSNumber.class])text=[value stringValue];
 if([self.node[@"role"] isEqual:@"checkbox"]&&[value isKindOfClass:NSNumber.class])text=[value boolValue]?@"true":@"false";
 if([self.node[@"role"] isEqual:@"text-field"]||[self.node[@"role"] isEqual:@"multiline-text"]){
  // Refuse before UTF8String/C.GoString can truncate NUL or queue a value that
  // the native document cannot represent. Owner model limits still apply.
  if(!text||text.length>(4u<<20)||[text lengthOfBytesUsingEncoding:NSUTF8StringEncoding]>(4u<<20)||![text dataUsingEncoding:NSUTF8StringEncoding])return;
  unichar nul=0;if([text rangeOfString:[NSString stringWithCharacters:&nul length:1]].location!=NSNotFound)return;
 }
 if(text)[self request:@"set" value:text];
}
@end
static id vellumAXFocused(VellumAXState *state){for(VellumAXElement *x in state.elements.allValues)if(x.live&&[x isAccessibilityFocused])return x;return nil;}
// Descriptor siblings retain painter/navigation order. Search later branches
// first and descendants before their parent; traverse zero-frame containers
// because synthetic groups may still contain laid-out native targets.
static VellumAXElement *vellumAXHitTree(NSArray *elements,NSPoint point){
 for(VellumAXElement *element in elements.reverseObjectEnumerator){
  if(![element isAccessibilityElement])continue;
  VellumAXElement *child=vellumAXHitTree(element.childElements,point);
  if(child)return child;
  if(NSPointInRect(point,element.accessibilityFrame))return element;
 }
 return nil;
}

@interface VellumAXViewMethods : NSView
@end
@implementation VellumAXViewMethods
- (BOOL)isAccessibilityElement{return NO;}
- (NSArray *)accessibilityChildren{return vellumAXState(self).roots?:@[];}
- (NSArray *)accessibilityChildrenInNavigationOrder{return self.accessibilityChildren;}
- (id)accessibilityFocusedUIElement{return vellumAXFocused(vellumAXState(self));}
- (id)accessibilityHitTest:(NSPoint)point{
 VellumAXState *state=vellumAXState(self);
 return vellumAXHitTree(state.roots,point)?:self;
}
@end
static Class vellumAXSubclass(Class base,Class methods,const char *prefix){
 NSString *name=[NSString stringWithFormat:@"%s_%s",prefix,class_getName(base)];Class cls=NSClassFromString(name);if(cls)return cls;
 cls=objc_allocateClassPair(base,name.UTF8String,0);if(!cls)return Nil;
 unsigned int count=0;Method *list=class_copyMethodList(methods,&count);
 for(unsigned int i=0;i<count;i++)class_addMethod(cls,method_getName(list[i]),method_getImplementation(list[i]),method_getTypeEncoding(list[i]));
 free(list);objc_registerClassPair(cls);return cls;
}
int vellumAXAttach(uint64_t token,void *viewPtr,void *windowPtr){
 NSView *view=(__bridge NSView *)viewPtr;NSWindow *window=(__bridge NSWindow *)windowPtr;
 if(![NSThread isMainThread]||!view||!window||vellumAXState(view))return 0;
 VellumAXState *state=[VellumAXState new];state.view=view;state.window=window;state.token=token;state.elements=[NSMutableDictionary new];state.geometry=[NSMutableDictionary new];state.textGeometry=[NSMutableDictionary new];state.roots=@[];
 state.viewClass=object_getClass(view);
 Class viewClass=vellumAXSubclass(state.viewClass,VellumAXViewMethods.class,"VellumAXView");
 if(!viewClass)return 0;state.installedViewClass=viewClass;
 objc_setAssociatedObject(view,&vellumAXKey,state,OBJC_ASSOCIATION_RETAIN_NONATOMIC);
 object_setClass(view,viewClass);
 if(!vellumAXStates)vellumAXStates=[NSMutableDictionary new];vellumAXStates[@(token)]=state;return 1;
}
void vellumAXDetach(uint64_t token){
 // Production NativeAttachment.Close runs before GLFW destruction on main.
 // Keep direct callers safe without exposing mutable AppKit state off main.
 if(![NSThread isMainThread]){dispatch_sync(dispatch_get_main_queue(),^{vellumAXDetach(token);});return;}
 VellumAXState *state=vellumAXStates[@(token)];if(!state||state.closed)return;state.closed=YES;
 for(VellumAXElement *x in state.elements.allValues){x.live=NO;x.parentElement=nil;x.childElements=@[];}
 [state.geometry removeAllObjects];[state.textGeometry removeAllObjects];
 // A later observer/subclass owns its own isa change; do not overwrite it.
 if(state.view&&object_getClass(state.view)==state.installedViewClass)object_setClass(state.view,state.viewClass);
 objc_setAssociatedObject(state.view,&vellumAXKey,nil,OBJC_ASSOCIATION_RETAIN_NONATOMIC);
 [vellumAXStates removeObjectForKey:@(token)];
}
// JSON bytes are copied before async publication; no NSString UTF8 roundtrip.
static NSData *vellumAXJSONData(const char *json){return json?[NSData dataWithBytes:json length:strlen(json)]:nil;}
// CGO producers may run on Go threads with no ambient Cocoa pool. Copy and
// enqueue inside a local pool; the queued block owns its captured NSData.
static void vellumAXWithJSONPayload(const char *json,void (^enqueue)(NSData *)){
 @autoreleasepool{NSData *payload=vellumAXJSONData(json);if(payload)enqueue(payload);}
}
static BOOL vellumAXUInt(id value,BOOL positive){
 if(![value isKindOfClass:NSNumber.class]||CFGetTypeID((__bridge CFTypeRef)value)==CFBooleanGetTypeID())return NO;
 NSString *digits=[value stringValue];if(digits.length==0||digits.length>20)return NO;
 for(NSUInteger i=0;i<digits.length;i++){unichar c=[digits characterAtIndex:i];if(c<'0'||c>'9')return NO;}
 if(digits.length==20&&[digits compare:@"18446744073709551615"]==NSOrderedDescending)return NO;
 return !positive||[value unsignedLongLongValue]!=0;
}
static BOOL vellumAXBool(id value){return [value isKindOfClass:NSNumber.class]&&CFGetTypeID((__bridge CFTypeRef)value)==CFBooleanGetTypeID();}
static BOOL vellumAXValidRect(id value){
 if(![value isKindOfClass:NSArray.class]||[value count]!=4)return NO;
 for(id n in value)if(![n isKindOfClass:NSNumber.class]||CFGetTypeID((__bridge CFTypeRef)n)==CFBooleanGetTypeID()||!isfinite([n doubleValue]))return NO;
 double x=[value[0] doubleValue],y=[value[1] doubleValue],w=[value[2] doubleValue],h=[value[3] doubleValue];
 return w>=0&&h>=0&&isfinite(x+w)&&isfinite(y+h);
}
// Validate the whole immutable payload before any retained element changes.
static NSDictionary *vellumAXValidatedNodes(NSDictionary *snapshot){
 if(![snapshot isKindOfClass:NSDictionary.class]||!vellumAXUInt(snapshot[@"generation"],YES)||![snapshot[@"nodes"] isKindOfClass:NSArray.class])return nil;
 NSMutableDictionary *nodes=[NSMutableDictionary new];
 for(id value in snapshot[@"nodes"]){
  if(![value isKindOfClass:NSDictionary.class])return nil;NSDictionary *node=value;
  NSString *identifier=node[@"id"],*parent=node[@"parent"];
  if(![identifier isKindOfClass:NSString.class]||identifier.length==0||nodes[identifier]||![parent isKindOfClass:NSString.class]||![node[@"role"] isKindOfClass:NSString.class]||![node[@"actions"] isKindOfClass:NSArray.class])return nil;
  for(id action in node[@"actions"])if(![action isKindOfClass:NSString.class])return nil;
  for(NSString *key in @[@"domain_id",@"name",@"help"])if(node[key]&&![node[key] isKindOfClass:NSString.class])return nil;
  for(NSString *key in @[@"disabled",@"focusable",@"focused",@"selected",@"min",@"max",@"row_index",@"row_count",@"column_count",@"column_index",@"grid_generation",@"level",@"expanded"])if(node[key]&&![node[key] isKindOfClass:NSNumber.class])return nil;
  id accepted=node[@"value"];if(accepted&&![accepted isKindOfClass:NSString.class]&&![accepted isKindOfClass:NSNumber.class])return nil;
  id scope=node[@"grid_scope"];
  if(scope){
   if(![scope isKindOfClass:NSString.class]||[scope length]==0||!vellumAXUInt(node[@"grid_incarnation"],YES)||(node[@"grid_pending"]&&!vellumAXBool(node[@"grid_pending"]))||!vellumAXUInt(node[@"grid_target_generation"],YES)||(node[@"grid_generation"]&&!vellumAXUInt(node[@"grid_generation"],NO)))return nil;
   uint64_t completed=[node[@"grid_generation"] unsignedLongLongValue],target=[node[@"grid_target_generation"] unsignedLongLongValue];BOOL pending=[node[@"grid_pending"] boolValue];
   if(completed>target||(!pending&&completed!=target))return nil;
   if(pending&&([node[@"actions"] count]!=0||!vellumAXBool(node[@"disabled"])||![node[@"disabled"] boolValue]||[node[@"focused"] boolValue]))return nil;
   if(node[@"grid_frame"]&&!vellumAXValidRect(node[@"grid_frame"]))return nil;
   if(!pending&&!node[@"grid_frame"])return nil;
  }else for(NSString *key in @[@"grid_incarnation",@"grid_pending",@"grid_target_generation",@"grid_frame"])if(node[key])return nil;
  nodes[identifier]=node;
 }
 for(NSString *identifier in nodes){
  NSMutableSet *ancestors=[NSMutableSet new];NSString *cursor=identifier;NSDictionary *node=nodes[identifier];NSString *scope=node[@"grid_scope"];
  while(cursor.length){if([ancestors containsObject:cursor]||!nodes[cursor])return nil;[ancestors addObject:cursor];cursor=nodes[cursor][@"parent"];}
  NSDictionary *parent=nodes[node[@"parent"]];
  if(scope){
   NSDictionary *host=nodes[scope];if(!host||![host[@"grid_scope"] isEqual:scope]||![host[@"role"] isEqual:@"table"])return nil;
   for(NSString *key in @[@"grid_incarnation",@"grid_target_generation"])if(![node[key] isEqual:host[key]])return nil;
   // Go's ordinary omitempty encoding omits ready false and startup zero.
   // Missing values normalize only here; present values were type-validated.
   if([node[@"grid_pending"] boolValue]!=[host[@"grid_pending"] boolValue]||[node[@"grid_generation"] unsignedLongLongValue]!=[host[@"grid_generation"] unsignedLongLongValue])return nil;
   if([identifier isEqual:scope]){if(parent[@"grid_scope"])return nil;}
   else if(![parent[@"grid_scope"] isEqual:scope]||![ancestors containsObject:scope])return nil;
  }else if(parent[@"grid_scope"])return nil;
 }
 return nodes;
}
static BOOL vellumAXValidSnapshot(NSDictionary *snapshot){return vellumAXValidatedNodes(snapshot)!=nil;}
int vellumAXSnapshotValid(const char *json){
 @autoreleasepool{NSData *data=vellumAXJSONData(json);NSDictionary *snapshot=data?[NSJSONSerialization JSONObjectWithData:data options:0 error:nil]:nil;return vellumAXValidSnapshot(snapshot);}
}

// Native unit fixtures use the production retained-tree traversal with pixel
// frames, without creating an OS window or requiring a running main queue.
@interface VellumAXHitFixtureElement : VellumAXElement
@end
@implementation VellumAXHitFixtureElement
- (NSRect)accessibilityFrame{return self.pixelFrame;}
@end
char *vellumAXHitFixture(const char *json,const char *boundsJSON,double x,double y){
 @autoreleasepool{
  NSString *payload=json?[NSString stringWithUTF8String:json]:nil;
  NSString *geometry=boundsJSON?[NSString stringWithUTF8String:boundsJSON]:nil;
  NSDictionary *snapshot=payload?[NSJSONSerialization JSONObjectWithData:[payload dataUsingEncoding:NSUTF8StringEncoding] options:0 error:nil]:nil;
  NSDictionary *bounds=geometry?[NSJSONSerialization JSONObjectWithData:[geometry dataUsingEncoding:NSUTF8StringEncoding] options:0 error:nil]:nil;
  if(!vellumAXValidSnapshot(snapshot)||![bounds isKindOfClass:NSDictionary.class])return strdup("");
  NSMutableDictionary *elements=[NSMutableDictionary new],*children=[NSMutableDictionary new];NSMutableArray *roots=[NSMutableArray new];
  for(NSDictionary *node in snapshot[@"nodes"]){
   NSString *identifier=node[@"id"];VellumAXHitFixtureElement *element=[VellumAXHitFixtureElement new];element.node=node;element.live=YES;
   NSArray *rect=bounds[identifier];
   if(rect){if(![rect isKindOfClass:NSArray.class]||rect.count!=4)return strdup("");for(id number in rect)if(![number isKindOfClass:NSNumber.class])return strdup("");element.pixelFrame=NSMakeRect([rect[0] doubleValue],[rect[1] doubleValue],[rect[2] doubleValue],[rect[3] doubleValue]);}
   elements[identifier]=element;children[identifier]=[NSMutableArray new];
  }
  for(NSDictionary *node in snapshot[@"nodes"]){VellumAXElement *element=elements[node[@"id"]],*parent=elements[node[@"parent"]];element.parentElement=parent;if(parent)[children[node[@"parent"]] addObject:element];else [roots addObject:element];}
  for(NSString *identifier in elements)((VellumAXElement *)elements[identifier]).childElements=children[identifier];
  NSString *identifier=[vellumAXHitTree(roots,NSMakePoint(x,y)) accessibilityIdentifier];return strdup((identifier?:@"").UTF8String);
 }
}

// Inert retained-tree fixture: invokes production table selectors without an
// NSWindow, native dispatcher, screen permissions, or synthetic action delivery.
char *vellumAXGridFixture(const char *json,const char *tableID,int column,int row){
 @autoreleasepool{
  NSString *payload=json?[NSString stringWithUTF8String:json]:nil;
  NSDictionary *snapshot=payload?[NSJSONSerialization JSONObjectWithData:[payload dataUsingEncoding:NSUTF8StringEncoding] options:0 error:nil]:nil;
  if(!vellumAXValidSnapshot(snapshot))return strdup("");
  NSMutableDictionary *elements=[NSMutableDictionary new],*children=[NSMutableDictionary new];
  for(NSDictionary *node in snapshot[@"nodes"]){VellumAXElement *element=[VellumAXElement new];element.node=node;element.live=YES;elements[node[@"id"]]=element;children[node[@"id"]]=[NSMutableArray new];}
  for(NSDictionary *node in snapshot[@"nodes"]){VellumAXElement *element=elements[node[@"id"]],*parent=elements[node[@"parent"]];element.parentElement=parent;if(parent)[children[node[@"parent"]] addObject:element];}
  for(NSString *identifier in elements)((VellumAXElement *)elements[identifier]).childElements=children[identifier];
  VellumAXElement *table=elements[tableID?[NSString stringWithUTF8String:tableID]:@""];
  if(!table)return strdup("");
  VellumAXElement *cell=[table accessibilityCellForColumn:column row:row];
  NSDictionary *result=@{@"role":[table accessibilityRole]?:@"",@"rows":@([table accessibilityRowCount]),@"columns":@([table accessibilityColumnCount]),@"visible_rows":@(table.accessibilityVisibleRows.count),@"visible_columns":@(table.accessibilityVisibleColumns.count),@"visible_cells":@(table.accessibilityVisibleCells.count),@"selected_cells":@(table.accessibilitySelectedCells.count),@"cell":cell.accessibilityIdentifier?:@"",@"cell_role":cell.accessibilityRole?:@"",@"row":cell?@(cell.accessibilityRowIndexRange.location):@(-1),@"column":cell?@(cell.accessibilityColumnIndexRange.location):@(-1)};
  NSData *data=[NSJSONSerialization dataWithJSONObject:result options:0 error:nil];return strdup([[NSString alloc]initWithData:data encoding:NSUTF8StringEncoding].UTF8String?:"");
 }
}

static void vellumAXRetire(VellumAXElement *element){element.live=NO;element.parentElement=nil;element.childElements=@[];element.textGeometry=nil;}
// Scope host Parent is part of an incarnation's native ownership boundary.
static BOOL vellumAXCanReuse(VellumAXElement *old,NSDictionary *node,NSDictionary *nodes,NSDictionary<NSString *,VellumAXElement *> *oldElements){
 if(!old.live)return NO;
 NSString *scope=node[@"grid_scope"],*oldScope=old.node[@"grid_scope"];
 // Ordinary controls retain their legacy ID identity and cached geometry
 // across reparenting. Parent is an ownership boundary only for grid scopes.
 if(!scope&&!oldScope)return YES;
 if(![old.node[@"parent"] isEqual:node[@"parent"]]||![scope isEqual:oldScope]||![node[@"grid_incarnation"] isEqual:old.node[@"grid_incarnation"]])return NO;
 return [nodes[scope][@"parent"] isEqual:oldElements[scope].node[@"parent"]];
}
// Called synchronously on the native owner by publication; inert fixtures use
// the same transaction without NSWindow creation or notifications.
static BOOL vellumAXCommitSnapshot(VellumAXState *state,NSDictionary *snapshot,BOOL notify){
 if(!state||state.closed)return NO;
 NSDictionary *nodes=vellumAXValidatedNodes(snapshot);if(!nodes)return NO;
 uint64_t generation=[snapshot[@"generation"] unsignedLongLongValue];if(generation<=state.generation)return NO;
 // Allocate every required object before mutating the retained tree. Identity
 // exhaustion refuses the complete transaction and cannot wrap old IDs.
 NSMutableDictionary<NSString *,VellumAXElement *> *next=[NSMutableDictionary new];NSMutableDictionary *childLists=[NSMutableDictionary new];
 for(NSDictionary *node in snapshot[@"nodes"]){NSString *identifier=node[@"id"];VellumAXElement *old=state.elements[identifier];VellumAXElement *element=vellumAXCanReuse(old,node,nodes,state.elements)?old:[VellumAXElement new];if(!element)return NO;next[identifier]=element;childLists[identifier]=[NSMutableArray new];}
 id oldFocus=vellumAXFocused(state);NSString *oldFocusedID=[oldFocus isKindOfClass:VellumAXElement.class]?[oldFocus accessibilityIdentifier]:@"";
 NSArray *oldHierarchy=state.hierarchy;NSMutableArray *hierarchy=[NSMutableArray new],*changedElements=[NSMutableArray new],*roots=[NSMutableArray new];
 for(NSDictionary *node in snapshot[@"nodes"]){
  NSString *identifier=node[@"id"];VellumAXElement *element=next[identifier],*old=state.elements[identifier];
  [hierarchy addObject:@[identifier,node[@"parent"],node[@"role"],@(element.objectIdentity)]];
  NSObject *oldValue=element.node[@"value"],*newValue=node[@"value"];BOOL changed=element.node&&oldValue!=newValue&&![oldValue isEqual:newValue];
  if(old&&old!=element){vellumAXRetire(old);[state.geometry removeObjectForKey:identifier];[state.textGeometry removeObjectForKey:identifier];}
  element.state=state;element.node=node;element.live=YES;element.textGeometry=state.textGeometry[identifier];
  if(node[@"grid_scope"]){element.pixelFrame=node[@"grid_frame"]?vellumAXArrayRect(node[@"grid_frame"]):NSZeroRect;[state.geometry removeObjectForKey:identifier];}
  else {NSArray *rect=state.geometry[identifier];element.pixelFrame=rect.count==4?vellumAXArrayRect(rect):NSZeroRect;}
  if(changed)[changedElements addObject:element];
 }
 for(NSDictionary *node in snapshot[@"nodes"]){VellumAXElement *element=next[node[@"id"]];NSString *parent=node[@"parent"];VellumAXElement *ancestor=next[parent];element.parentElement=ancestor?:state.view;if(ancestor)[childLists[parent] addObject:element];else [roots addObject:element];}
 for(NSString *identifier in next)next[identifier].childElements=childLists[identifier];
 for(NSString *identifier in state.elements)if(!next[identifier])vellumAXRetire(state.elements[identifier]);
 for(NSString *identifier in state.geometry.allKeys)if(!next[identifier])[state.geometry removeObjectForKey:identifier];
 for(NSString *identifier in state.textGeometry.allKeys)if(!next[identifier])[state.textGeometry removeObjectForKey:identifier];
 state.elements=next;state.roots=roots;state.hierarchy=hierarchy;state.generation=generation;
 if(notify){
  if(![oldHierarchy isEqual:hierarchy])NSAccessibilityPostNotification(state.view,NSAccessibilityLayoutChangedNotification);
  for(VellumAXElement *changed in changedElements)NSAccessibilityPostNotification(changed,NSAccessibilityValueChangedNotification);
  id focused=vellumAXFocused(state);NSString *focusedID=focused?[focused accessibilityIdentifier]:@"";if(![focusedID isEqual:oldFocusedID])NSAccessibilityPostNotification(focused?:state.view,NSAccessibilityFocusedUIElementChangedNotification);
 }
 return YES;
}
void vellumAXUpdate(uint64_t token,const char *json){
 vellumAXWithJSONPayload(json,^(NSData *payload){
  dispatch_async(dispatch_get_main_queue(),^{@autoreleasepool{NSDictionary *snapshot=[NSJSONSerialization JSONObjectWithData:payload options:0 error:nil];vellumAXCommitSnapshot(vellumAXStates[@(token)],snapshot,YES);}});
 });
}
static void vellumAXApplyGeometry(VellumAXState *state,NSDictionary *bounds){
 if(!state||state.closed||![bounds isKindOfClass:NSDictionary.class])return;
 for(NSString *identifier in bounds){
  // Grid frames arrive only with their semantic generation. An unversioned
  // render callback cannot overwrite the completed or pending envelope.
  if(state.elements[identifier].node[@"grid_scope"])continue;
  NSArray *rect=bounds[identifier];if(!vellumAXValidRect(rect))continue;state.geometry[identifier]=rect;state.elements[identifier].pixelFrame=vellumAXArrayRect(rect);
 }
}
void vellumAXGeometry(uint64_t token,const char *json){
 vellumAXWithJSONPayload(json,^(NSData *payload){
  dispatch_async(dispatch_get_main_queue(),^{@autoreleasepool{NSDictionary *bounds=[NSJSONSerialization JSONObjectWithData:payload options:0 error:nil];vellumAXApplyGeometry(vellumAXStates[@(token)],bounds);}});
 });
}
// Synchronous CGO transaction fixture. It never registers an attachment, creates
// an OS window, or invokes a successful owner action. Retained references expose
// retirement independently of the new same-ID dictionary entry.
static NSDictionary *vellumAXFixtureElement(VellumAXElement *element){
 NSRect frame=element.pixelFrame;
 return @{@"id":element.accessibilityIdentifier?:@"",@"object_identity":@(element.objectIdentity),@"live":@(element.live),@"enabled":@([element isAccessibilityEnabled]),@"focused":@([element isAccessibilityFocused]),@"press_allowed":@([element isAccessibilitySelectorAllowed:@selector(accessibilityPerformPress)]),@"grid_scope":element.node[@"grid_scope"]?:@"",@"grid_incarnation":element.node[@"grid_incarnation"]?:@0,@"grid_pending":@([element.node[@"grid_pending"] boolValue]),@"grid_target_generation":element.node[@"grid_target_generation"]?:@0,@"grid_generation":element.node[@"grid_generation"]?:@0,@"value":element.accessibilityValue?:NSNull.null,@"frame":@[@(frame.origin.x),@(frame.origin.y),@(frame.size.width),@(frame.size.height)]};
}
char *vellumAXScopedFixture(const char *json){
 @autoreleasepool{
  NSData *data=vellumAXJSONData(json);id steps=data?[NSJSONSerialization JSONObjectWithData:data options:0 error:nil]:nil;
  if(![steps isKindOfClass:NSArray.class]||[steps count]>512)return strdup("");
  VellumAXState *state=[VellumAXState new];state.elements=[NSMutableDictionary new];state.geometry=[NSMutableDictionary new];state.textGeometry=[NSMutableDictionary new];state.roots=@[];
  NSMutableArray *held=[NSMutableArray new],*results=[NSMutableArray new];
  for(id step in steps){
   if(![step isKindOfClass:NSDictionary.class])return strdup("");BOOL accepted=NO;
   if([step[@"kind"] isEqual:@"update"]){accepted=vellumAXCommitSnapshot(state,step[@"snapshot"],NO);if(accepted)for(VellumAXElement *element in state.elements.allValues)if(![held containsObject:element])[held addObject:element];}
   else if([step[@"kind"] isEqual:@"geometry"])vellumAXApplyGeometry(state,step[@"bounds"]);
   else if([step[@"kind"] isEqual:@"close"]){state.closed=YES;for(VellumAXElement *element in state.elements.allValues)vellumAXRetire(element);}
   else return strdup("");
   NSMutableArray *elements=[NSMutableArray new],*retired=[NSMutableArray new],*guards=[NSMutableArray new];
   for(NSString *identifier in [state.elements.allKeys sortedArrayUsingSelector:@selector(compare:)]){
    VellumAXElement *element=state.elements[identifier];[elements addObject:vellumAXFixtureElement(element)];
    if(vellumAXPending(element)){
     // Defense-in-depth check: real pending selectors refuse independently of
     // the owner's disabled/empty-actions/unfocused presentation flags.
     VellumAXElement *probe=[VellumAXElement new];NSMutableDictionary *node=[element.node mutableCopy];node[@"disabled"]=@NO;node[@"focused"]=@YES;node[@"actions"]=@[@"press",@"focus",@"set",@"selection",@"increment",@"decrement",@"cancel",@"expand",@"collapse"];
     probe.state=state;probe.node=node;probe.live=YES;
     NSMutableDictionary *guard=[vellumAXFixtureElement(probe) mutableCopy];guard[@"admitted"]=@([probe request:@"press" value:@""]);
     guard[@"set_allowed"]=@([probe isAccessibilitySelectorAllowed:@selector(setAccessibilityValue:)]);guard[@"focus_allowed"]=@([probe isAccessibilitySelectorAllowed:@selector(setAccessibilityFocused:)]);[guards addObject:guard];
    }
   }
   for(VellumAXElement *element in held)if(!element.live){NSMutableDictionary *record=[vellumAXFixtureElement(element) mutableCopy];record[@"admitted"]=@([element request:@"press" value:@""]);[retired addObject:record];}
   [results addObject:@{@"accepted":@(accepted),@"generation":@(state.generation),@"elements":elements,@"retired":retired,@"pending_guards":guards}];
  }
  NSData *result=[NSJSONSerialization dataWithJSONObject:results options:0 error:nil];if(!result)return strdup("");return strdup([[NSString alloc]initWithData:result encoding:NSUTF8StringEncoding].UTF8String?:"");
 }
}
static void vellumAXWalk(id object,NSMutableArray *out){
 for(id child in [object accessibilityChildren]){
  if([child isKindOfClass:VellumAXElement.class]){
   VellumAXElement *x=child;NSRect frame=x.accessibilityFrame;id value=x.accessibilityValue?:NSNull.null;NSString *parent=[x.parentElement isKindOfClass:VellumAXElement.class]?[x.parentElement accessibilityIdentifier]:@"";
   id hit=[x.state.view accessibilityHitTest:NSMakePoint(NSMidX(frame),NSMidY(frame))];NSString *hitID=[hit isKindOfClass:VellumAXElement.class]?[hit accessibilityIdentifier]:@"";
   [out addObject:@{@"id":x.accessibilityIdentifier,@"domain_id":x.node[@"domain_id"]?:@"",@"role":x.accessibilityRole,@"name":x.accessibilityLabel?:@"",@"value":value,@"enabled":@([x isAccessibilityEnabled]),@"object_identity":@(x.objectIdentity),@"grid_scope":x.node[@"grid_scope"]?:@"",@"grid_incarnation":x.node[@"grid_incarnation"]?:@0,@"grid_pending":@([x.node[@"grid_pending"] boolValue]),@"grid_target_generation":x.node[@"grid_target_generation"]?:@0,@"grid_generation":x.node[@"grid_generation"]?:@0,@"selected":@([x isAccessibilitySelected]),@"expanded":@([x isAccessibilityExpanded]),@"parent":parent,@"hit_test_id":hitID,@"text_selection":@[x.accessibilitySelectedTextRange.location==NSNotFound?@(-1):@(x.accessibilitySelectedTextRange.location),@(x.accessibilitySelectedTextRange.length)],@"insertion_line":@(x.accessibilityInsertionPointLineNumber),@"row_count":@(x.accessibilityRowCount),@"column_count":@(x.accessibilityColumnCount),@"row_index":@[x.accessibilityRowIndexRange.location==NSNotFound?@(-1):@(x.accessibilityRowIndexRange.location),@(x.accessibilityRowIndexRange.length)],@"column_index":@[x.accessibilityColumnIndexRange.location==NSNotFound?@(-1):@(x.accessibilityColumnIndexRange.location),@(x.accessibilityColumnIndexRange.length)],@"frame":@[@(frame.origin.x),@(frame.origin.y),@(frame.size.width),@(frame.size.height)]}];
  }
  vellumAXWalk(child,out);
 }
}
void vellumAXProbe(uint64_t token,uint64_t request){dispatch_async(dispatch_get_main_queue(),^{@autoreleasepool{
 VellumAXState *state=vellumAXStates[@(token)];if(!state||state.closed){vellumAXResult(request,"{}");return;}
 NSMutableArray *out=[NSMutableArray new];vellumAXWalk(state.window.contentView,out);
 id focused=[state.window accessibilityFocusedUIElement];NSString *id=[focused isKindOfClass:VellumAXElement.class]?[focused accessibilityIdentifier]:@"";
 NSDictionary *result=@{@"generation":@(state.generation),@"reachable":@((BOOL)(out.count==state.elements.count&&out.count>0)),@"focused_id":id,@"window_role":state.window.accessibilityRole?:@"",@"native_window":@([state.window isKindOfClass:NSWindow.class]),@"elements":out};
 NSData *data=[NSJSONSerialization dataWithJSONObject:result options:0 error:nil];NSString *json=[[NSString alloc]initWithData:data encoding:NSUTF8StringEncoding];vellumAXResult(request,(char *)json.UTF8String);
}});}
void vellumAXPerform(uint64_t token,uint64_t request,const char *id,const char *action,const char *value){
 @autoreleasepool{
 NSString *identifier=id?[NSString stringWithUTF8String:id]:@"";NSString *kind=action?[NSString stringWithUTF8String:action]:@"";NSString *text=value?[NSString stringWithUTF8String:value]:@"";
 dispatch_async(dispatch_get_main_queue(),^{@autoreleasepool{
  VellumAXState *state=vellumAXStates[@(token)];VellumAXElement *element=state.elements[identifier];BOOL accepted=NO;
  if([kind isEqual:@"press"])accepted=[element accessibilityPerformPress];
  else if([kind isEqual:@"increment"])accepted=[element accessibilityPerformIncrement];
  else if([kind isEqual:@"decrement"])accepted=[element accessibilityPerformDecrement];
  else if([kind isEqual:@"cancel"])accepted=[element accessibilityPerformCancel];
  else if([kind isEqual:@"set"]){element.lastAdmission=NO;[element setAccessibilityValue:text];accepted=element.lastAdmission;}
  else if([kind isEqual:@"selection"]){NSDictionary *r=[NSJSONSerialization JSONObjectWithData:[text dataUsingEncoding:NSUTF8StringEncoding] options:0 error:nil];if([r isKindOfClass:NSDictionary.class]&&[r[@"text"] isEqual:element.node[@"value"]]&&[r[@"start"] isKindOfClass:NSNumber.class]&&[r[@"length"] isKindOfClass:NSNumber.class]){element.lastAdmission=NO;[element setAccessibilitySelectedTextRange:NSMakeRange([r[@"start"] unsignedIntegerValue],[r[@"length"] unsignedIntegerValue])];accepted=element.lastAdmission;}}
  else if([kind isEqual:@"focus"]){element.lastAdmission=NO;[element setAccessibilityFocused:YES];accepted=element.lastAdmission;}
  else if([kind isEqual:@"expand"]||[kind isEqual:@"collapse"]){element.lastAdmission=NO;[element setAccessibilityExpanded:[kind isEqual:@"expand"]];accepted=element.lastAdmission;}
  vellumAXResult(request,accepted?"true":"false");
 }});
 }
}

// Geometry is accepted as one validated immutable payload and retained only for
// its exact control incarnation and whole text. It cannot answer stale ranges.
static BOOL vellumAXRectValid(id value){if(![value isKindOfClass:NSArray.class]||[value count]!=4)return NO;for(id n in value)if(![n isKindOfClass:NSNumber.class]||!isfinite([n doubleValue]))return NO;return [value[2] doubleValue]>=0&&[value[3] doubleValue]>=0;}
static BOOL vellumAXTextEntryValid(id entry){
 if(![entry isKindOfClass:NSDictionary.class]||![entry[@"client"] isKindOfClass:NSNumber.class]||![entry[@"text"] isKindOfClass:NSString.class])return NO;
 NSDictionary *g=entry[@"geometry"];NSString *text=entry[@"text"];if(![g isKindOfClass:NSDictionary.class]||!vellumAXRectValid(g[@"viewport"])||![g[@"focus_line"] isKindOfClass:NSNumber.class])return NO;
 for(NSString *kind in @[@"cells",@"carets"]){id list=g[kind];if(![list isKindOfClass:NSArray.class])return NO;for(id c in list){if(![c isKindOfClass:NSDictionary.class]||!vellumAXRectValid(c[@"rect"])||![c[@"line"] isKindOfClass:NSNumber.class])return NO;
 if([kind isEqual:@"cells"]){if(![c[@"start"] isKindOfClass:NSNumber.class]||![c[@"end"] isKindOfClass:NSNumber.class])return NO;NSInteger start=[c[@"start"] integerValue],end=[c[@"end"] integerValue];if(start<0||end<=start||!vellumAXValidRange(text,NSMakeRange(start,end-start)))return NO;}
 else {if(![c[@"offset"] isKindOfClass:NSNumber.class]||[c[@"offset"] integerValue]<0||!vellumAXValidRange(text,NSMakeRange([c[@"offset"] unsignedIntegerValue],0)))return NO;}}}return YES;
}
void vellumAXTextGeometry(uint64_t token,const char *json){
 vellumAXWithJSONPayload(json,^(NSData *payload){dispatch_async(dispatch_get_main_queue(),^{@autoreleasepool{
  VellumAXState *state=vellumAXStates[@(token)];if(!state||state.closed)return;NSDictionary *entries=[NSJSONSerialization JSONObjectWithData:payload options:0 error:nil];if(![entries isKindOfClass:NSDictionary.class])return;
  for(id key in entries)if(![key isKindOfClass:NSString.class]||!vellumAXTextEntryValid(entries[key]))return;
  for(NSString *key in entries){state.textGeometry[key]=entries[key];state.elements[key].textGeometry=entries[key];}
 }});});
}
void vellumAXTextProbe(uint64_t token,uint64_t request,const char *id,int start,int length,double x,double y){@autoreleasepool{NSString *identifier=id?[NSString stringWithUTF8String:id]:@"";dispatch_async(dispatch_get_main_queue(),^{@autoreleasepool{
 VellumAXState *state=vellumAXStates[@(token)];VellumAXElement *e=state.elements[identifier];NSRange r=start<0||length<0?NSMakeRange(NSNotFound,0):NSMakeRange(start,length),selected=e.accessibilitySelectedTextRange,hit=[e accessibilityRangeForPosition:NSMakePoint(x,y)];NSRect frame=[e accessibilityFrameForRange:r];
 NSDictionary *result=@{@"available":@((BOOL)([e managedGeometry]!=nil)),@"text":[e managedText]?:@"",@"substring":[e accessibilityStringForRange:r]?:@"",@"selected_start":selected.location==NSNotFound?@(-1):@(selected.location),@"selected_length":@(selected.length),@"insertion_line":@(e.accessibilityInsertionPointLineNumber),@"line":@([e accessibilityLineForIndex:start]),@"frame":@[@(frame.origin.x),@(frame.origin.y),@(frame.size.width),@(frame.size.height)],@"point_start":hit.location==NSNotFound?@(-1):@(hit.location),@"point_length":@(hit.length)};
 NSData *data=[NSJSONSerialization dataWithJSONObject:result options:0 error:nil];NSString *resultJSON=[[NSString alloc]initWithData:data encoding:NSUTF8StringEncoding];vellumAXResult(request,(char *)resultJSON.UTF8String);
 }});}}
@interface VellumAXTextFixtureElement : VellumAXElement
@end
@implementation VellumAXTextFixtureElement
- (NSRect)managedScreenRect:(NSRect)rect{return rect;}
@end
char *vellumAXTextFixture(const char *nodeJSON,const char *geometryJSON,int start,int length,int line){
 @autoreleasepool{
 NSString *nodeText=nodeJSON?[NSString stringWithUTF8String:nodeJSON]:nil,*geometryText=geometryJSON?[NSString stringWithUTF8String:geometryJSON]:nil;
 NSDictionary *node=nodeText?[NSJSONSerialization JSONObjectWithData:[nodeText dataUsingEncoding:NSUTF8StringEncoding] options:0 error:nil]:nil;
 NSDictionary *entry=geometryText?[NSJSONSerialization JSONObjectWithData:[geometryText dataUsingEncoding:NSUTF8StringEncoding] options:0 error:nil]:nil;
 if(![node isKindOfClass:NSDictionary.class]||!vellumAXTextEntryValid(entry))return strdup("{}");
 VellumAXState *state=[VellumAXState new];VellumAXTextFixtureElement *e=[VellumAXTextFixtureElement new];e.state=state;e.node=node;e.live=YES;e.textGeometry=entry;
 NSRange range=start<0||length<0?NSMakeRange(NSNotFound,0):NSMakeRange(start,length),selected=e.accessibilitySelectedTextRange,lines=[e accessibilityRangeForLine:line];NSRect frame=[e accessibilityFrameForRange:range];
 NSDictionary *out=@{@"available":@((BOOL)([e managedGeometry]!=nil)),@"substring":[e accessibilityStringForRange:range]?:@"",@"selected_start":selected.location==NSNotFound?@(-1):@(selected.location),@"selected_length":@(selected.length),@"insertion_line":@([e accessibilityInsertionPointLineNumber]),@"line_start":lines.location==NSNotFound?@(-1):@(lines.location),@"line_length":@(lines.length),@"frame":@[@(frame.origin.x),@(frame.origin.y),@(frame.size.width),@(frame.size.height)]};
 NSData *data=[NSJSONSerialization dataWithJSONObject:out options:0 error:nil];NSString *json=[[NSString alloc]initWithData:data encoding:NSUTF8StringEncoding];return strdup(json.UTF8String);
 }
}

// Exercise the actual Objective-C setter without an NSWindow or OS preferences.
// JSON keeps embedded NUL visible until the setter's pre-C-string admission.
int vellumAXSetValueFixture(uint64_t token,uint64_t generation,const char *nodeJSON,const char *valueJSON){
 @autoreleasepool{
  NSString *n=nodeJSON?[NSString stringWithUTF8String:nodeJSON]:nil,*v=valueJSON?[NSString stringWithUTF8String:valueJSON]:nil;
  NSDictionary *node=n?[NSJSONSerialization JSONObjectWithData:[n dataUsingEncoding:NSUTF8StringEncoding] options:0 error:nil]:nil;
  id value=v?[NSJSONSerialization JSONObjectWithData:[v dataUsingEncoding:NSUTF8StringEncoding] options:NSJSONReadingFragmentsAllowed error:nil]:nil;
  if(![node isKindOfClass:NSDictionary.class]||!value)return 0;
  VellumAXState *state=[VellumAXState new];state.token=token;state.generation=generation;
  VellumAXElement *e=[VellumAXElement new];e.state=state;e.node=node;e.live=YES;
  [e setAccessibilityValue:value];return e.lastAdmission?1:0;
 }
}

// Check producer ownership with a deliberately delayed consumer and an outer
// pool that remains undrained throughout the lifetime checks. This fixture has
// no native attachment, successful action, or main-queue scheduling.
char *vellumAXPayloadFixture(const char *json,int consume){
 @autoreleasepool{
  if(!json||strlen(json)==0)return strdup("{}");
  size_t length=strlen(json);char *input=strdup(json);if(!input)return strdup("{}");
  NSData *expected=[[NSData alloc]initWithBytes:input length:length];
  __block dispatch_block_t queued=nil;__block __weak NSData *weakPayload=nil;
  __block BOOL copied=NO,decoded=NO;
  vellumAXWithJSONPayload(input,^(NSData *payload){
   weakPayload=payload;
   queued=[^{@autoreleasepool{copied=[payload isEqualToData:expected];decoded=[NSJSONSerialization JSONObjectWithData:payload options:NSJSONReadingFragmentsAllowed error:nil]!=nil;}} copy];
  });
  memset(input,'x',length);free(input);
  BOOL alive;@autoreleasepool{alive=weakPayload!=nil;}
  if(consume&&queued)queued();queued=nil;
  BOOL released;@autoreleasepool{released=weakPayload==nil;}
  NSDictionary *result=@{@"alive_while_queued":@(alive),@"copied_bytes":@(copied),@"decoded":@(decoded),@"released_after_consumer":@(released)};
  NSData *data=[NSJSONSerialization dataWithJSONObject:result options:0 error:nil];
  return strdup([[NSString alloc]initWithData:data encoding:NSUTF8StringEncoding].UTF8String?:"");
 }
}
