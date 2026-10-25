//go:build darwin && cgo
#import <Foundation/Foundation.h>
static void vellumGalleryFreshSession(NSUserDefaults *defaults){
 NSMutableDictionary *args=[[defaults volatileDomainForName:NSArgumentDomain] mutableCopy]?:[NSMutableDictionary new];
 args[@"ApplePersistenceIgnoreState"]=@YES;
 [defaults setVolatileDomain:args forName:NSArgumentDomain];
}
// Cogent initializes GLFW in a dependency init, before Go main. Bootstrap the
// packaged gallery before Go initialization; never alter embedding apps.
static BOOL vellumGalleryBundle(NSString *bundle){return [bundle isEqual:@"com.dshills.vellum.widgets.light.native"]||[bundle isEqual:@"com.dshills.vellum.widgets.dark.native"];}
__attribute__((constructor)) static void vellumGalleryFreshCocoaSession(void){@autoreleasepool{if(vellumGalleryBundle(NSBundle.mainBundle.bundleIdentifier))vellumGalleryFreshSession(NSUserDefaults.standardUserDefaults);}}
int vellumGalleryRestorationFixture(void){@autoreleasepool{
 NSString *name=[@"vellum-restoration-test-" stringByAppendingString:NSUUID.UUID.UUIDString];
 NSUserDefaults *defaults=[[NSUserDefaults alloc]initWithSuiteName:name];
 [defaults setVolatileDomain:@{@"unrelated":@"preserved"} forName:NSArgumentDomain];
 vellumGalleryFreshSession(defaults);
 NSDictionary *args=[defaults volatileDomainForName:NSArgumentDomain];
 return vellumGalleryBundle(@"com.dshills.vellum.widgets.light.native")&&vellumGalleryBundle(@"com.dshills.vellum.widgets.dark.native")&&!vellumGalleryBundle(@"other.app")&&[args[@"ApplePersistenceIgnoreState"] boolValue]&&[args[@"unrelated"] isEqual:@"preserved"]&&[defaults persistentDomainForName:name].count==0;
}}
