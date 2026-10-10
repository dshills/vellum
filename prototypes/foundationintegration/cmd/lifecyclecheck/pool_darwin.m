#import <Foundation/Foundation.h>
void *vellum_check_pool_begin(void) { return [[NSAutoreleasePool alloc] init]; }
void vellum_check_pool_end(void *pool) { [(NSAutoreleasePool *)pool drain]; }
#import <AppKit/AppKit.h>
#import <objc/runtime.h>
static char bareWindowKey;
void *vellum_bare_view_new(void){
 NSView *view=[[NSView alloc] initWithFrame:NSMakeRect(0,0,400,250)];
 NSWindow *window=[[NSWindow alloc] initWithContentRect:NSMakeRect(0,0,400,250) styleMask:NSWindowStyleMaskBorderless backing:NSBackingStoreBuffered defer:NO];
 [window setReleasedWhenClosed:NO];[window setContentView:view];
 objc_setAssociatedObject(view,&bareWindowKey,window,OBJC_ASSOCIATION_RETAIN_NONATOMIC);[window release];
 return view;
}
void vellum_bare_view_free(void *ptr){NSView *view=ptr;NSWindow *window=objc_getAssociatedObject(view,&bareWindowKey);[window close];objc_setAssociatedObject(view,&bareWindowKey,nil,OBJC_ASSOCIATION_RETAIN_NONATOMIC);[view release];}
