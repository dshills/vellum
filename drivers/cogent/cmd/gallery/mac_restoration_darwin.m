//go:build darwin && cgo
#import <Foundation/Foundation.h>
static void vellumGalleryFreshSession(NSUserDefaults *defaults){
 NSMutableDictionary *args=[[defaults volatileDomainForName:NSArgumentDomain] mutableCopy]?:[NSMutableDictionary new];
 args[@"ApplePersistenceIgnoreState"]=@YES;
 [defaults setVolatileDomain:args forName:NSArgumentDomain];
}
void vellumGalleryFreshCocoaSession(void){@autoreleasepool{vellumGalleryFreshSession(NSUserDefaults.standardUserDefaults);}}
int vellumGalleryRestorationFixture(void){@autoreleasepool{
 NSString *name=[@"vellum-restoration-test-" stringByAppendingString:NSUUID.UUID.UUIDString];
 NSUserDefaults *defaults=[[NSUserDefaults alloc]initWithSuiteName:name];
 [defaults setVolatileDomain:@{@"unrelated":@"preserved"} forName:NSArgumentDomain];
 vellumGalleryFreshSession(defaults);
 NSDictionary *args=[defaults volatileDomainForName:NSArgumentDomain];
 return [args[@"ApplePersistenceIgnoreState"] boolValue]&&[args[@"unrelated"] isEqual:@"preserved"]&&[defaults persistentDomainForName:name].count==0;
}}
