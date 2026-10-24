//go:build darwin && cgo
#import <Cocoa/Cocoa.h>
#include <string.h>
#include <stdlib.h>
#import "services_native.h"
extern int vellumServiceActive(uint64_t token);
extern void vellumServiceResult(uint64_t token,char *text,int status);
extern void *vellumAXNativeWindow(uint64_t token);
static NSMutableDictionary<NSNumber *,NSAlert *> *vellumAlerts;

// Foundation-only recipe parser shared by native presentation and offscreen
// fixtures. It never instantiates NSAlert/NSWindow or dispatches to the desktop.
static NSDictionary *vellumAlertRecipe(NSData *data) {
 NSDictionary *o=[NSJSONSerialization JSONObjectWithData:data options:0 error:nil];
 if (![o isKindOfClass:NSDictionary.class]||![o[@"Title"] isKindOfClass:NSString.class]||![o[@"Message"] isKindOfClass:NSString.class]) return nil;
 NSArray *buttons=o[@"Buttons"];if (![buttons isKindOfClass:NSArray.class]||buttons.count<1||buttons.count>3) return nil;
 int defaults=0,cancels=0;NSMutableSet *ids=[NSMutableSet new];
 for (NSDictionary *b in buttons) {
  if (![b isKindOfClass:NSDictionary.class]||![b[@"ID"] isKindOfClass:NSString.class]||![b[@"Label"] isKindOfClass:NSString.class]||![b[@"Role"] isKindOfClass:NSString.class]) return nil;
  if (![b[@"ID"] length]||![b[@"Label"] length]||[ids containsObject:b[@"ID"]]) return nil;[ids addObject:b[@"ID"]];
  NSString *role=b[@"Role"];if ([role isEqual:@"default"]) defaults++;else if ([role isEqual:@"cancel"]) cancels++;else if (role.length&&![role isEqual:@"destructive"])return nil;
 }
 if (defaults!=1||cancels>1) return nil;return o;
}
char *vellumAlertRecipeFixture(const char *json) {
 @autoreleasepool {@try {if (!json||strnlen(json,VELLUM_SERVICE_REQUEST_MAX+1)>VELLUM_SERVICE_REQUEST_MAX)return NULL;NSData *data=[NSData dataWithBytes:json length:strlen(json)];NSDictionary *o=vellumAlertRecipe(data);if (!o)return NULL;NSData *result=[NSJSONSerialization dataWithJSONObject:@{@"buttons":o[@"Buttons"],@"title":o[@"Title"]} options:0 error:nil];NSString *text=[[NSString alloc]initWithData:result encoding:NSUTF8StringEncoding];return text.UTF8String?strdup(text.UTF8String):NULL;}@catch(NSException *e){return NULL;}}
}
static int vellumDismissAlert(NSAlert *alert) {
 int status=0;@try{NSWindow *parent=alert.window.sheetParent;if(parent)[parent endSheet:alert.window returnCode:NSModalResponseCancel];}@catch(NSException *e){status=2;}
 @try{[alert.window orderOut:nil];}@catch(NSException *e){status=2;}return status;
}
void vellumShowAlert(uint64_t token,uint64_t parentToken,const char *json) {
 vellumWithServiceJSON(token,json,^(NSData *data){dispatch_async(dispatch_get_main_queue(),^{@autoreleasepool {@try{
  if(!vellumServiceActive(token)){vellumServiceResult(token,"",1);return;}
  NSDictionary *o=vellumAlertRecipe(data);NSWindow *parent=(__bridge NSWindow *)vellumAXNativeWindow(parentToken);
  if(!o){vellumServiceResult(token,"",2);return;}
  if(!parent){vellumServiceResult(token,"",6);return;}
  if(!parent.isVisible){vellumServiceResult(token,"",2);return;}
  if(vellumAlerts.count||parent.attachedSheet||NSApp.modalWindow){vellumServiceResult(token,"",5);return;}
  NSAlert *alert=[NSAlert new];alert.messageText=o[@"Title"];alert.informativeText=o[@"Message"];
  NSString *style=o[@"Style"];alert.alertStyle=[style isEqual:@"critical"]?NSAlertStyleCritical:([style isEqual:@"warning"]?NSAlertStyleWarning:NSAlertStyleInformational);
  NSString *encoded=o[@"IconPNG"];if([encoded isKindOfClass:NSString.class]&&encoded.length){NSData *png=[[NSData alloc]initWithBase64EncodedString:encoded options:0];NSImage *image=[[NSImage alloc]initWithData:png];if(!image){vellumServiceResult(token,"",2);return;}image.template=NO;alert.icon=image;}
  NSArray *buttons=o[@"Buttons"];for(NSDictionary *b in buttons){NSButton *button=[alert addButtonWithTitle:b[@"Label"]];button.keyEquivalent=@"";if([b[@"Role"] isEqual:@"default"])button.keyEquivalent=@"\r";if([b[@"Role"] isEqual:@"cancel"])button.keyEquivalent=@"\033";if(@available(macOS 11.0,*)){button.hasDestructiveAction=[b[@"Role"] isEqual:@"destructive"];}}
  NSString *suppression=o[@"SuppressionLabel"];if(suppression.length){alert.showsSuppressionButton=YES;alert.suppressionButton.title=suppression;alert.suppressionButton.state=[o[@"InitiallySuppressed"] boolValue]?NSControlStateValueOn:NSControlStateValueOff;}
  if(!vellumAlerts)vellumAlerts=[NSMutableDictionary new];if(!vellumServiceActive(token)){vellumServiceResult(token,"",1);return;}vellumAlerts[@(token)]=alert;
  [alert beginSheetModalForWindow:parent completionHandler:^(NSModalResponse response){@autoreleasepool {@try{NSInteger index=response-NSAlertFirstButtonReturn;if(index<0||index>=(NSInteger)buttons.count){vellumServiceResult(token,"",1);}else{NSDictionary *result=@{@"ButtonID":buttons[index][@"ID"],@"Suppressed":@(alert.suppressionButton.state==NSControlStateValueOn)};NSData *reply=[NSJSONSerialization dataWithJSONObject:result options:0 error:nil];NSString *text=[[NSString alloc]initWithData:reply encoding:NSUTF8StringEncoding];vellumServiceResult(token,(char *)(text?:@"").UTF8String,text?0:2);}}@catch(NSException *e){vellumServiceResult(token,"",2);}@finally{[vellumAlerts removeObjectForKey:@(token)];}}}];
 }@catch(NSException *e){NSAlert *alert=vellumAlerts[@(token)];if(alert&&!vellumDismissAlert(alert))[vellumAlerts removeObjectForKey:@(token)];vellumServiceResult(token,"",2);}}});});
}
void vellumCancelAlert(uint64_t token){dispatch_async(dispatch_get_main_queue(),^{@autoreleasepool{int status=0;@try{NSAlert *alert=vellumAlerts[@(token)];if(alert)status=vellumDismissAlert(alert);if(!status)[vellumAlerts removeObjectForKey:@(token)];}@catch(NSException *e){status=2;}vellumServiceResult(token,"",status?2:1);}});}
int vellumCancelAllAlerts(void){int status=0;for(NSNumber *token in vellumAlerts.allKeys.copy){NSAlert *alert=vellumAlerts[token];if(vellumDismissAlert(alert))status=2;else[vellumAlerts removeObjectForKey:token];}return status;}
