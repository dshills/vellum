//go:build darwin && cgo

#import <Cocoa/Cocoa.h>
#import "services_native.h"
#include <stdint.h>
#include <string.h>
extern int vellumServiceActive(uint64_t token);
extern void vellumServiceResult(uint64_t token, char *text, int status);
extern int vellumTrayCommit(uint64_t token, int remove);
extern void vellumTrayAction(uint64_t token, int index);

@interface VellumTrayTarget : NSObject
@property uint64_t token;
- (void)invoke:(NSMenuItem *)item;
@end
@implementation VellumTrayTarget
- (void)invoke:(NSMenuItem *)item { vellumTrayAction(self.token, (int)item.tag); }
@end
static NSStatusItem *vellumTrayItem;
static VellumTrayTarget *vellumTrayTarget;

static BOOL trayString(id value, NSUInteger maxBytes) {
    if (![value isKindOfClass:[NSString class]] || [value lengthOfBytesUsingEncoding:NSUTF8StringEncoding] > maxBytes) return NO;
    for (NSUInteger i=0; i<[value length]; i++) if ([value characterAtIndex:i]==0) return NO;
    return YES;
}
static BOOL trayBool(id value) { return [value isKindOfClass:[NSNumber class]] && CFGetTypeID((__bridge CFTypeRef)value)==CFBooleanGetTypeID(); }
static BOOL trayInt(id value, NSInteger min, NSInteger max) {
    if (![value isKindOfClass:[NSNumber class]] || trayBool(value)) return NO;
    double n = [value doubleValue];
    return n>=min && n<=max && n==[value integerValue];
}
// Decode and validate without constructing status items, menus or images. The
// trusted producer always emits the complete nine-field versionless envelope.
static NSDictionary *trayParse(NSData *wire) {
    if (!wire || wire.length > VELLUM_SERVICE_REQUEST_MAX) return nil;
    id value = [NSJSONSerialization JSONObjectWithData:wire options:0 error:nil];
    if (![value isKindOfClass:[NSDictionary class]] || [value count]!=9) return nil;
    NSDictionary *d=value;
    if (!trayString(d[@"Title"],256) || !trayString(d[@"Tooltip"],1024) || !trayString(d[@"AccessibilityLabel"],1024) || !trayBool(d[@"Template"]) || !trayBool(d[@"Remove"]) ||
        !trayInt(d[@"Width"],0,256) || !trayInt(d[@"Height"],0,256)) return nil;
    NSArray *items=d[@"Items"]==NSNull.null ? @[] : d[@"Items"];
    if (![items isKindOfClass:[NSArray class]] || items.count>128) return nil;
    NSInteger next=0;
    for (id obj in items) {
        if (![obj isKindOfClass:[NSDictionary class]] || [obj count]!=4) return nil;
        NSDictionary *item=obj;
        if (!trayString(item[@"Label"],1024) || !trayBool(item[@"Disabled"]) || !trayBool(item[@"Separator"]) || !trayInt(item[@"Index"],-1,127)) return nil;
        if ([item[@"Separator"] boolValue]) {
            if ([item[@"Label"] length] || [item[@"Index"] integerValue]!=-1 || ![item[@"Disabled"] boolValue]) return nil;
        } else if (![item[@"Label"] length] || [item[@"Index"] integerValue]!=next++) return nil;
    }
    NSInteger w=[d[@"Width"] integerValue],h=[d[@"Height"] integerValue];
    NSData *rgba=nil;
    id encoded=d[@"RGBA"];
    if (encoded!=NSNull.null) {
        if (![encoded isKindOfClass:[NSString class]] || [encoded length]>349528) return nil;
        rgba=[[NSData alloc] initWithBase64EncodedString:encoded options:0];
        if (!rgba) return nil;
    }
    if ([d[@"Remove"] boolValue]) {
        if ([d[@"Title"] length] || [d[@"Tooltip"] length] || [d[@"AccessibilityLabel"] length] || [d[@"Template"] boolValue] || w || h || rgba.length || items.count) return nil;
    } else {
        if ((!w)!=(!h) || (w && rgba.length!=(NSUInteger)(w*h*4)) || (!w && rgba.length)) return nil;
        if (![d[@"Title"] length] && !w) return nil;
        if (![[d[@"AccessibilityLabel"] stringByTrimmingCharactersInSet:NSCharacterSet.whitespaceAndNewlineCharacterSet] length]) return nil;
        if ([d[@"Template"] boolValue] && !w) return nil;
    }
    return @{@"Title":d[@"Title"],@"Tooltip":d[@"Tooltip"],@"AccessibilityLabel":d[@"AccessibilityLabel"],@"Template":d[@"Template"],@"Remove":d[@"Remove"],
             @"Width":d[@"Width"],@"Height":d[@"Height"],@"Pixels":rgba ?: NSData.data,@"Items":items};
}
int vellumTrayValidate(const char *json) {
    if (!json || strnlen(json,VELLUM_SERVICE_REQUEST_MAX+1)>VELLUM_SERVICE_REQUEST_MAX) return 0;
    @autoreleasepool { @try {
        NSData *wire=[NSData dataWithBytesNoCopy:(void *)json length:strlen(json) freeWhenDone:NO];
        return trayParse(wire)!=nil;
    } @catch (NSException *exception) { return 0; } }
}

static NSImage *trayImage(NSDictionary *d) {
    NSInteger w=[d[@"Width"] integerValue],h=[d[@"Height"] integerValue];
    if (!w) return nil;
    NSData *pixels=d[@"Pixels"];
    NSBitmapImageRep *rep=[[NSBitmapImageRep alloc] initWithBitmapDataPlanes:NULL pixelsWide:w pixelsHigh:h bitsPerSample:8
        samplesPerPixel:4 hasAlpha:YES isPlanar:NO colorSpaceName:NSDeviceRGBColorSpace bitmapFormat:NSBitmapFormatAlphaNonpremultiplied
        bytesPerRow:w*4 bitsPerPixel:32];
    if (!rep || !rep.bitmapData || rep.bytesPerRow<w*4) return nil;
    for (NSInteger y=0;y<h;y++) memcpy(rep.bitmapData+y*rep.bytesPerRow,(const uint8_t *)pixels.bytes+y*w*4,(size_t)w*4);
    CGFloat scale=18.0/MAX(w,h);
    NSImage *image=[[NSImage alloc] initWithSize:NSMakeSize(w*scale,h*scale)];
    [image addRepresentation:rep]; image.template=[d[@"Template"] boolValue];
    return image;
}

// Exercise the real bitmap/image conversion offscreen, without a status item.
int vellumTrayImageFixture(const char *json) {
    if (!json || strnlen(json,VELLUM_SERVICE_REQUEST_MAX+1)>VELLUM_SERVICE_REQUEST_MAX) return 0;
    @autoreleasepool { @try {
        NSDictionary *d=trayParse([NSData dataWithBytesNoCopy:(void *)json length:strlen(json) freeWhenDone:NO]);
        NSImage *image=d ? trayImage(d) : nil;
        NSBitmapImageRep *rep=(NSBitmapImageRep *)image.representations.firstObject;
        if (!rep) return 0;
        NSInteger w=[d[@"Width"] integerValue],h=[d[@"Height"] integerValue];
        NSData *pixels=d[@"Pixels"]; BOOL same=YES;
        for (NSInteger y=0;y<h;y++) if (memcmp(rep.bitmapData+y*rep.bytesPerRow,(const uint8_t *)pixels.bytes+y*w*4,(size_t)w*4)) same=NO;
        BOOL alpha=(rep.bitmapFormat&NSBitmapFormatAlphaNonpremultiplied)!=0 && (rep.bitmapFormat&NSBitmapFormatAlphaFirst)==0;
        CGFloat scale=18.0/MAX(w,h); BOOL geometry=image.size.width==w*scale && image.size.height==h*scale;
        return (same?1:0)|(alpha?2:0)|(geometry?4:0);
    } @catch (NSException *exception) { return 0; } }
}

void vellumSetTray(uint64_t token,const char *json) {
    vellumWithServiceJSON(token,json,^(NSData *wire) {
        dispatch_async(dispatch_get_main_queue(),^{ @autoreleasepool {
            NSStatusItem *candidate=nil;
            BOOL created=NO;
            BOOL committed=NO;
            @try {
                if (!vellumServiceActive(token)) { vellumServiceResult(token,"",2); return; }
                NSDictionary *d=trayParse(wire);
                if (!d) { vellumServiceResult(token,"",2); return; }
                if ([d[@"Remove"] boolValue]) {
                    if (!vellumTrayCommit(token,1)) { vellumServiceResult(token,"",2); return; }
                    if (vellumTrayItem) [NSStatusBar.systemStatusBar removeStatusItem:vellumTrayItem];
                    vellumTrayItem=nil; vellumTrayTarget=nil;
                    vellumServiceResult(token,"",0); return;
                }
                NSImage *icon=trayImage(d);
                if ([d[@"Width"] integerValue] && !icon) { vellumServiceResult(token,"",2); return; }
                VellumTrayTarget *target=[VellumTrayTarget new]; target.token=token;
                NSMenu *menu=[NSMenu new]; menu.autoenablesItems=NO;
                for (NSDictionary *item in d[@"Items"]) {
                    if ([item[@"Separator"] boolValue]) { [menu addItem:NSMenuItem.separatorItem]; continue; }
                    NSMenuItem *entry=[[NSMenuItem alloc] initWithTitle:item[@"Label"] action:@selector(invoke:) keyEquivalent:@""];
                    entry.target=target; entry.tag=[item[@"Index"] integerValue]; entry.enabled=![item[@"Disabled"] boolValue];
                    [menu addItem:entry];
                }
                if (!vellumServiceActive(token)) { vellumServiceResult(token,"",2); return; }
                candidate=vellumTrayItem;
                if (!candidate) { candidate=[NSStatusBar.systemStatusBar statusItemWithLength:NSVariableStatusItemLength]; created=YES; candidate.visible=NO; }
                if (!candidate.button) { if (created && candidate) [NSStatusBar.systemStatusBar removeStatusItem:candidate]; vellumServiceResult(token,"",2); return; }
                if (!vellumTrayCommit(token,0)) {
                    if (created) [NSStatusBar.systemStatusBar removeStatusItem:candidate]; vellumServiceResult(token,"",2); return;
                }
                committed=YES;
                // Keep accepted native ownership visible to retries/teardown even
                // if a later AppKit presentation setter fails.
                vellumTrayItem=candidate; vellumTrayTarget=target;
                // Reuse the installed native object; updates preserve bar position
                // and mutate presentation only after callback ownership transfers.
                candidate.button.title=d[@"Title"]; candidate.button.toolTip=d[@"Tooltip"];
                candidate.button.accessibilityLabel=d[@"AccessibilityLabel"];
                candidate.button.image=icon; candidate.button.imagePosition=NSImageLeft;
                candidate.menu=[d[@"Items"] count] ? menu : nil;
                candidate.visible=YES;
                vellumServiceResult(token,"",0);
            } @catch (NSException *exception) {
                if (created && candidate && !committed) { @try { [NSStatusBar.systemStatusBar removeStatusItem:candidate]; } @catch (NSException *cleanupException) {} }
                vellumServiceResult(token,"",2);
            }
        }});
    });
}
// Called on the native main queue; only this adapter's item is removed.
void vellumTrayShutdown(void) {
    if (vellumTrayItem) [NSStatusBar.systemStatusBar removeStatusItem:vellumTrayItem];
    vellumTrayItem=nil; vellumTrayTarget=nil;
}

#ifdef VELLUM_NATIVE_CHECKS
#import "native_services_checks.h"
static __weak NSStatusItem *checkLastTrayItem;
static uint64_t checkTrayIdentity;
static VellumTrayTarget *checkStaleTrayTarget;
static NSMenuItem *checkStaleTrayItem;
NSDictionary *vellumCheckTray(int operation,int index) {
    if (!NSThread.isMainThread) return nil;
    if (operation==8) { checkStaleTrayTarget=nil;checkStaleTrayItem=nil; }
    if (operation==7) {
        if (!checkStaleTrayTarget || !checkStaleTrayItem) return nil;
        [checkStaleTrayTarget invoke:checkStaleTrayItem];
    }
    if (vellumTrayItem && checkLastTrayItem!=vellumTrayItem) {
        if (checkTrayIdentity==UINT64_MAX) return nil;
        checkTrayIdentity++;checkLastTrayItem=vellumTrayItem;
    }
    NSStatusBarButton *button=vellumTrayItem.button;NSImage *image=button.image;
    NSBitmapImageRep *rep=(NSBitmapImageRep *)image.representations.firstObject;
    NSMutableArray *items=[NSMutableArray new];NSInteger foundIndex=-1;
    for (NSMenuItem *item in vellumTrayItem.menu.itemArray) {
        [items addObject:@{@"label":item.title?:@"",@"index":item.separatorItem?@(-1):@(item.tag),@"enabled":@(item.enabled),@"separator":@(item.separatorItem)}];
        if (!item.separatorItem && item.tag==index && item.target==vellumTrayTarget) foundIndex=[vellumTrayItem.menu indexOfItem:item];
    }
    if (operation==5 || operation==6) {
        if (!vellumTrayTarget || foundIndex<0) return nil;
        NSMenuItem *item=[vellumTrayItem.menu itemAtIndex:foundIndex];
        if (operation==5) [vellumTrayItem.menu performActionForItemAtIndex:foundIndex];
        else { checkStaleTrayTarget=vellumTrayTarget;checkStaleTrayItem=item.copy; }
    }
    return @{@"Present":@((BOOL)(vellumTrayItem!=nil)),@"Visible":@(vellumTrayItem.visible),@"TargetRetired":@((BOOL)(vellumTrayTarget==nil)),
             @"Identity":vellumTrayItem?@(checkTrayIdentity):@0,@"Title":button.title?:@"",@"Tooltip":button.toolTip?:@"",
             @"AccessibilityLabel":button.accessibilityLabel?:@"",@"Template":@(image.template),@"ImageWidth":@(image.size.width),@"ImageHeight":@(image.size.height),
             @"PixelWidth":@(rep.pixelsWide),@"PixelHeight":@(rep.pixelsHigh),@"Items":items};
}
#endif
