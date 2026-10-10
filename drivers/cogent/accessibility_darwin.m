//go:build darwin && cgo

#import <Cocoa/Cocoa.h>
#import <objc/runtime.h>
#import <objc/message.h>
#include <stdint.h>
#include <string.h>
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
@end
@implementation VellumAXState
@end
static VellumAXState *vellumAXState(id object){return objc_getAssociatedObject(object,&vellumAXKey);}
static BOOL vellumAXHas(VellumAXElement *element,NSString *action){return [element.node[@"actions"] containsObject:action];}
static NSString *vellumAXRole(NSString *role){
 if([role isEqual:@"label"])return NSAccessibilityStaticTextRole;
 if([role isEqual:@"button"])return NSAccessibilityButtonRole;
 if([role isEqual:@"checkbox"])return NSAccessibilityCheckBoxRole;
 if([role isEqual:@"text-field"])return NSAccessibilityTextFieldRole;
 if([role isEqual:@"number-field"])return NSAccessibilityIncrementorRole;
 if([role isEqual:@"slider"])return NSAccessibilitySliderRole;
 if([role isEqual:@"selector"])return NSAccessibilityPopUpButtonRole;
 if([role isEqual:@"tabs"])return NSAccessibilityTabGroupRole;
 if([role isEqual:@"option"])return NSAccessibilityRadioButtonRole;
 if([role isEqual:@"scroll"])return NSAccessibilityScrollAreaRole;
 if([role isEqual:@"table"])return NSAccessibilityTableRole;
 if([role isEqual:@"outline"])return NSAccessibilityOutlineRole;
 if([role isEqual:@"row"]||[role isEqual:@"tree-row"])return NSAccessibilityRowRole;
 return NSAccessibilityGroupRole;
}
@implementation VellumAXElement
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
- (BOOL)isAccessibilityEnabled{return [self isAccessibilityElement]&&![self.node[@"disabled"] boolValue];}
- (BOOL)isAccessibilityFocused{return [self.node[@"focused"] boolValue];}
- (BOOL)isAccessibilitySelected{return [self.node[@"selected"] boolValue];}
- (BOOL)isAccessibilityModal{return [self.node[@"role"] isEqual:@"dialog"];}
- (BOOL)isAccessibilityExpanded{return [self.node[@"expanded"] boolValue];}
- (void)setAccessibilityExpanded:(BOOL)expanded{[self request:expanded?@"expand":@"collapse" value:@""];}
- (NSInteger)accessibilityIndex{return MAX(0,[self.node[@"row_index"] integerValue]-1);}
- (NSInteger)accessibilityDisclosureLevel{return MAX(0,[self.node[@"level"] integerValue]-1);}
- (id)accessibilityParent{return self.parentElement;}
- (NSArray *)accessibilityChildren{return self.childElements?:@[];}
- (NSArray *)accessibilityVisibleChildren{return self.accessibilityChildren;}
- (NSArray *)accessibilityRows{return [self.childElements filteredArrayUsingPredicate:[NSPredicate predicateWithBlock:^BOOL(VellumAXElement *x,NSDictionary *bindings){return [x.node[@"role"] isEqual:@"row"]||[x.node[@"role"] isEqual:@"tree-row"];}]];}
- (NSInteger)accessibilityRowCount{NSInteger count=0;for(VellumAXElement *row in self.accessibilityRows)count=MAX(count,[row.node[@"row_count"] integerValue]);return count;}
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
 VellumAXState *state=[VellumAXState new];state.view=view;state.window=window;state.token=token;state.elements=[NSMutableDictionary new];state.geometry=[NSMutableDictionary new];state.roots=@[];
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
 [state.geometry removeAllObjects];
 // A later observer/subclass owns its own isa change; do not overwrite it.
 if(state.view&&object_getClass(state.view)==state.installedViewClass)object_setClass(state.view,state.viewClass);
 objc_setAssociatedObject(state.view,&vellumAXKey,nil,OBJC_ASSOCIATION_RETAIN_NONATOMIC);
 [vellumAXStates removeObjectForKey:@(token)];
}
// Validate the whole immutable payload before any retained element changes.
static BOOL vellumAXValidSnapshot(NSDictionary *snapshot){
 if(![snapshot isKindOfClass:NSDictionary.class]||![snapshot[@"generation"] isKindOfClass:NSNumber.class]||![snapshot[@"nodes"] isKindOfClass:NSArray.class])return NO;
 NSMutableDictionary *parents=[NSMutableDictionary new];
 for(id value in snapshot[@"nodes"]){
  if(![value isKindOfClass:NSDictionary.class])return NO;NSDictionary *node=value;
  NSString *identifier=node[@"id"],*parent=node[@"parent"];
  if(![identifier isKindOfClass:NSString.class]||identifier.length==0||parents[identifier]||![parent isKindOfClass:NSString.class]||![node[@"role"] isKindOfClass:NSString.class]||![node[@"actions"] isKindOfClass:NSArray.class])return NO;
  for(id action in node[@"actions"])if(![action isKindOfClass:NSString.class])return NO;
  for(NSString *key in @[@"domain_id",@"name",@"help"])if(node[key]&&![node[key] isKindOfClass:NSString.class])return NO;
  for(NSString *key in @[@"disabled",@"focusable",@"focused",@"selected",@"min",@"max",@"row_index",@"row_count",@"level",@"expanded"])if(node[key]&&![node[key] isKindOfClass:NSNumber.class])return NO;
  id accepted=node[@"value"];if(accepted&&![accepted isKindOfClass:NSString.class]&&![accepted isKindOfClass:NSNumber.class])return NO;
  parents[identifier]=parent;
 }
 for(NSString *identifier in parents){
  NSMutableSet *ancestors=[NSMutableSet new];NSString *cursor=identifier;
  while(cursor.length){if([ancestors containsObject:cursor]||!parents[cursor])return NO;[ancestors addObject:cursor];cursor=parents[cursor];}
 }
 return YES;
}
int vellumAXSnapshotValid(const char *json){
 @autoreleasepool{NSString *payload=json?[NSString stringWithUTF8String:json]:nil;if(!payload)return 0;NSDictionary *snapshot=[NSJSONSerialization JSONObjectWithData:[payload dataUsingEncoding:NSUTF8StringEncoding] options:0 error:nil];return vellumAXValidSnapshot(snapshot);}
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

void vellumAXUpdate(uint64_t token,const char *json){
 NSString *payload=json?[NSString stringWithUTF8String:json]:nil;if(!payload)return;
 dispatch_async(dispatch_get_main_queue(),^{
  VellumAXState *state=vellumAXStates[@(token)];if(!state||state.closed)return;
  NSDictionary *snapshot=[NSJSONSerialization JSONObjectWithData:[payload dataUsingEncoding:NSUTF8StringEncoding] options:0 error:nil];
  if(!vellumAXValidSnapshot(snapshot))return;
  uint64_t generation=[snapshot[@"generation"] unsignedLongLongValue];if(generation<=state.generation)return;
  id oldFocus=vellumAXFocused(state);NSString *oldFocusedID=[oldFocus isKindOfClass:VellumAXElement.class]?[oldFocus accessibilityIdentifier]:@"";
  NSArray *oldHierarchy=state.hierarchy;NSMutableArray *hierarchy=[NSMutableArray new];NSMutableArray *changedElements=[NSMutableArray new];
  NSMutableDictionary *next=[NSMutableDictionary new];NSMutableDictionary *childLists=[NSMutableDictionary new];NSMutableArray *roots=[NSMutableArray new];
  for(NSDictionary *node in snapshot[@"nodes"]){
   [hierarchy addObject:@[node[@"id"],node[@"parent"]?:@"",node[@"role"]?:@""]];
   NSString *id=node[@"id"];VellumAXElement *element=state.elements[id]?:[VellumAXElement new];
   NSObject *oldValue=element.node[@"value"],*newValue=node[@"value"];
   BOOL changed=element.node&&oldValue!=newValue&&![oldValue isEqual:newValue];
   element.state=state;element.node=node;element.live=YES;next[id]=element;childLists[id]=[NSMutableArray new];
   NSArray *rect=state.geometry[id];if(rect.count==4)element.pixelFrame=NSMakeRect([rect[0] doubleValue],[rect[1] doubleValue],[rect[2] doubleValue],[rect[3] doubleValue]);
   if(changed)[changedElements addObject:element];
  }
  // Attach in descriptor order in a single pass, preserving navigation order.
  for(NSDictionary *node in snapshot[@"nodes"]){VellumAXElement *element=next[node[@"id"]];NSString *parent=node[@"parent"];VellumAXElement *ancestor=next[parent];element.parentElement=ancestor?:state.view;if(ancestor)[childLists[parent] addObject:element];else [roots addObject:element];}
  for(NSString *identifier in next){VellumAXElement *element=next[identifier];element.childElements=childLists[identifier];}
  for(NSString *id in state.elements)if(!next[id]){VellumAXElement *old=state.elements[id];old.live=NO;old.childElements=@[];old.parentElement=nil;}
  // Geometry may arrive before semantics during demand-rendered startup. Keep
  // it until matching elements publish, then retire IDs absent from this tree.
  for(NSString *identifier in state.geometry.allKeys)if(!next[identifier])[state.geometry removeObjectForKey:identifier];
  state.elements=next;state.roots=roots;state.hierarchy=hierarchy;state.generation=generation;
  if(![oldHierarchy isEqual:hierarchy])NSAccessibilityPostNotification(state.view,NSAccessibilityLayoutChangedNotification);
  for(VellumAXElement *changed in changedElements)NSAccessibilityPostNotification(changed,NSAccessibilityValueChangedNotification);
  id focused=vellumAXFocused(state);NSString *focusedID=focused?[focused accessibilityIdentifier]:@"";if(![focusedID isEqual:oldFocusedID])NSAccessibilityPostNotification(focused?:state.view,NSAccessibilityFocusedUIElementChangedNotification);
 });
}
void vellumAXGeometry(uint64_t token,const char *json){
 NSString *payload=json?[NSString stringWithUTF8String:json]:nil;if(!payload)return;
 dispatch_async(dispatch_get_main_queue(),^{
  VellumAXState *state=vellumAXStates[@(token)];if(!state||state.closed)return;
  NSDictionary *bounds=[NSJSONSerialization JSONObjectWithData:[payload dataUsingEncoding:NSUTF8StringEncoding] options:0 error:nil];if(![bounds isKindOfClass:NSDictionary.class])return;
  for(NSString *id in bounds){NSArray *rect=bounds[id];if(![rect isKindOfClass:NSArray.class]||rect.count!=4)continue;state.geometry[id]=rect;state.elements[id].pixelFrame=NSMakeRect([rect[0] doubleValue],[rect[1] doubleValue],[rect[2] doubleValue],[rect[3] doubleValue]);}
 });
}
static void vellumAXWalk(id object,NSMutableArray *out){
 for(id child in [object accessibilityChildren]){
  if([child isKindOfClass:VellumAXElement.class]){
   VellumAXElement *x=child;NSRect frame=x.accessibilityFrame;id value=x.accessibilityValue?:NSNull.null;NSString *parent=[x.parentElement isKindOfClass:VellumAXElement.class]?[x.parentElement accessibilityIdentifier]:@"";
   id hit=[x.state.view accessibilityHitTest:NSMakePoint(NSMidX(frame),NSMidY(frame))];NSString *hitID=[hit isKindOfClass:VellumAXElement.class]?[hit accessibilityIdentifier]:@"";
   [out addObject:@{@"id":x.accessibilityIdentifier,@"domain_id":x.node[@"domain_id"]?:@"",@"role":x.accessibilityRole,@"name":x.accessibilityLabel?:@"",@"value":value,@"enabled":@([x isAccessibilityEnabled]),@"selected":@([x isAccessibilitySelected]),@"expanded":@([x isAccessibilityExpanded]),@"parent":parent,@"hit_test_id":hitID,@"frame":@[@(frame.origin.x),@(frame.origin.y),@(frame.size.width),@(frame.size.height)]}];
  }
  vellumAXWalk(child,out);
 }
}
void vellumAXProbe(uint64_t token,uint64_t request){dispatch_async(dispatch_get_main_queue(),^{
 VellumAXState *state=vellumAXStates[@(token)];if(!state||state.closed){vellumAXResult(request,"{}");return;}
 NSMutableArray *out=[NSMutableArray new];vellumAXWalk(state.window.contentView,out);
 id focused=[state.window accessibilityFocusedUIElement];NSString *id=[focused isKindOfClass:VellumAXElement.class]?[focused accessibilityIdentifier]:@"";
 NSDictionary *result=@{@"generation":@(state.generation),@"reachable":@((BOOL)(out.count==state.elements.count&&out.count>0)),@"focused_id":id,@"window_role":state.window.accessibilityRole?:@"",@"native_window":@([state.window isKindOfClass:NSWindow.class]),@"elements":out};
 NSData *data=[NSJSONSerialization dataWithJSONObject:result options:0 error:nil];NSString *json=[[NSString alloc]initWithData:data encoding:NSUTF8StringEncoding];vellumAXResult(request,(char *)json.UTF8String);
});}
void vellumAXPerform(uint64_t token,uint64_t request,const char *id,const char *action,const char *value){
 NSString *identifier=id?[NSString stringWithUTF8String:id]:@"";NSString *kind=action?[NSString stringWithUTF8String:action]:@"";NSString *text=value?[NSString stringWithUTF8String:value]:@"";
 dispatch_async(dispatch_get_main_queue(),^{
  VellumAXState *state=vellumAXStates[@(token)];VellumAXElement *element=state.elements[identifier];BOOL accepted=NO;
  if([kind isEqual:@"press"])accepted=[element accessibilityPerformPress];
  else if([kind isEqual:@"increment"])accepted=[element accessibilityPerformIncrement];
  else if([kind isEqual:@"decrement"])accepted=[element accessibilityPerformDecrement];
  else if([kind isEqual:@"cancel"])accepted=[element accessibilityPerformCancel];
  else if([kind isEqual:@"set"]){element.lastAdmission=NO;[element setAccessibilityValue:text];accepted=element.lastAdmission;}
  else if([kind isEqual:@"focus"]){element.lastAdmission=NO;[element setAccessibilityFocused:YES];accepted=element.lastAdmission;}
  else if([kind isEqual:@"expand"]||[kind isEqual:@"collapse"]){element.lastAdmission=NO;[element setAccessibilityExpanded:[kind isEqual:@"expand"]];accepted=element.lastAdmission;}
  vellumAXResult(request,accepted?"true":"false");
 });
}
