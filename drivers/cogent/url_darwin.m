//go:build darwin && cgo
#import <Cocoa/Cocoa.h>
#include <stdint.h>
#include <string.h>
extern int vellumServiceActive(uint64_t);
extern void vellumServiceResult(uint64_t,char *,int);
static NSURL *vellumURLRecipe(const char *text){
 if(!text || strnlen(text,4097)>4096)return nil;
 NSString *s=[NSString stringWithUTF8String:text];
 if(!s.length || [s rangeOfCharacterFromSet:NSCharacterSet.controlCharacterSet].location!=NSNotFound || [s rangeOfCharacterFromSet:NSCharacterSet.whitespaceAndNewlineCharacterSet].location!=NSNotFound || [s containsString:@"\\"])return nil;
 NSURL *u=[NSURL URLWithString:s];
 if(!u.host.length || u.user || u.password || !([u.scheme isEqual:@"http"]||[u.scheme isEqual:@"https"]))return nil;
 return u;
}
int vellumURLRecipeFixture(const char *text){@autoreleasepool{return vellumURLRecipe(text)!=nil;}}
void vellumOpenURL(uint64_t token,const char *text){
 @autoreleasepool {
 NSURL *url=vellumURLRecipe(text);
 if(!url){vellumServiceResult(token,"",2);return;}
 dispatch_async(dispatch_get_main_queue(),^{@autoreleasepool{
  if(!vellumServiceActive(token)){vellumServiceResult(token,"",1);return;}
  @try{BOOL ok=[NSWorkspace.sharedWorkspace openURL:url];vellumServiceResult(token,"",ok?0:2);}
  @catch(NSException *e){vellumServiceResult(token,"",2);}
 }});
 }
}
// Opening is irreversible after native dispatch; cancellation only prevents a
// not-yet-started request. Never imply a browser window can be rolled back.
void vellumCancelOpenURL(uint64_t token){dispatch_async(dispatch_get_main_queue(),^{vellumServiceResult(token,"",1);});}
