//go:build darwin && cgo

#import <Cocoa/Cocoa.h>
#include <stdint.h>
extern int vellumServiceActive(uint64_t token);
extern void vellumServiceResult(uint64_t token, char *text, int status);
extern int vellumMenuCommit(uint64_t token);
extern void vellumMenuAction(uint64_t token, int index);
static NSMutableDictionary<NSNumber *, NSSavePanel *> *vellumPanels;
@interface VellumMenuTarget : NSObject
@property uint64_t token;
- (void)invoke:(NSMenuItem *)item;
@end
@implementation VellumMenuTarget
- (void)invoke:(NSMenuItem *)item { vellumMenuAction(self.token, (int)item.tag); }
@end
static VellumMenuTarget *vellumTarget;
static NSMenu *vellumOriginalMenu;

void vellumChooseFile(uint64_t token, const char *json) {
    NSString *payload = json ? [NSString stringWithUTF8String:json] : nil;
    if (!payload) { vellumServiceResult(token, "", 2); return; }
    dispatch_async(dispatch_get_main_queue(), ^{
        if (!vellumServiceActive(token)) return;
        NSData *data = [payload dataUsingEncoding:NSUTF8StringEncoding];
        NSDictionary *options = [NSJSONSerialization JSONObjectWithData:data options:0 error:nil];
        if (![options isKindOfClass:[NSDictionary class]]) { vellumServiceResult(token, "", 2); return; }
        BOOL save = [options[@"Mode"] intValue] == 1;
        NSSavePanel *panel = save ? [NSSavePanel savePanel] : [NSOpenPanel openPanel];
        if (!save) {
            NSOpenPanel *open = (NSOpenPanel *)panel;
            open.canChooseDirectories = NO; open.canChooseFiles = YES; open.allowsMultipleSelection = NO;
        }
        NSString *title = options[@"Title"];
        if (title.length) panel.title = title;
        NSString *directory = options[@"Directory"];
        if (directory.length) panel.directoryURL = [NSURL fileURLWithPath:directory isDirectory:YES];
        NSString *filename = options[@"Filename"];
        if (filename.length) panel.nameFieldStringValue = filename;
        NSArray *extensions = options[@"Extensions"];
        if ([extensions isKindOfClass:[NSArray class]] && extensions.count) {
#pragma clang diagnostic push
#pragma clang diagnostic ignored "-Wdeprecated-declarations"
            panel.allowedFileTypes = extensions;
#pragma clang diagnostic pop
        }
        if (!vellumPanels) vellumPanels = [NSMutableDictionary new];
        vellumPanels[@(token)] = panel;
        [panel beginWithCompletionHandler:^(NSModalResponse response) {
            NSString *path = response == NSModalResponseOK ? panel.URL.path : @"";
            vellumServiceResult(token, (char *)path.UTF8String, response == NSModalResponseOK ? 0 : 1);
            [vellumPanels removeObjectForKey:@(token)];
        }];
    });
}
void vellumCancelFile(uint64_t token) {
    dispatch_async(dispatch_get_main_queue(), ^{
        NSSavePanel *panel = vellumPanels[@(token)];
        [panel cancel:nil]; [panel orderOut:nil];
        [vellumPanels removeObjectForKey:@(token)];
    });
}
void vellumSetMenu(uint64_t token, const char *json) {
    NSString *payload = json ? [NSString stringWithUTF8String:json] : nil;
    if (!payload) { vellumServiceResult(token, "", 2); return; }
    dispatch_async(dispatch_get_main_queue(), ^{
        if (!vellumServiceActive(token)) return;
        NSData *data = [payload dataUsingEncoding:NSUTF8StringEncoding];
        NSArray *groups = [NSJSONSerialization JSONObjectWithData:data options:0 error:nil];
        if (![groups isKindOfClass:[NSArray class]]) { vellumServiceResult(token, "", 2); return; }
        if (!vellumOriginalMenu) vellumOriginalMenu = NSApp.mainMenu;
        NSMenu *menu = [NSMenu new];
        // Preserve the application menu supplied by GLFW (About/Hide/Quit).
        if (vellumOriginalMenu.numberOfItems > 0) [menu addItem:[vellumOriginalMenu itemAtIndex:0].copy];
        VellumMenuTarget *target = [VellumMenuTarget new]; target.token = token;
        for (NSDictionary *group in groups) {
            if (![group isKindOfClass:[NSDictionary class]] ||
                ![group[@"Label"] isKindOfClass:[NSString class]] ||
                ![group[@"Items"] isKindOfClass:[NSArray class]]) {
                vellumServiceResult(token, "", 2); return;
            }
            NSMenuItem *heading = [[NSMenuItem alloc] initWithTitle:group[@"Label"] action:nil keyEquivalent:@""];
            NSMenu *submenu = [[NSMenu alloc] initWithTitle:group[@"Label"]]; submenu.autoenablesItems = NO;
            for (NSDictionary *item in group[@"Items"]) {
                if (![item isKindOfClass:[NSDictionary class]] ||
                    ![item[@"Label"] isKindOfClass:[NSString class]] ||
                    ![item[@"Index"] isKindOfClass:[NSNumber class]] ||
                    ![item[@"Disabled"] isKindOfClass:[NSNumber class]]) {
                    vellumServiceResult(token, "", 2); return;
                }
                NSMenuItem *entry = [[NSMenuItem alloc] initWithTitle:item[@"Label"] action:@selector(invoke:) keyEquivalent:@""];
                entry.target = target; entry.tag = [item[@"Index"] integerValue];
                entry.enabled = ![item[@"Disabled"] boolValue]; [submenu addItem:entry];
            }
            heading.submenu = submenu; [menu addItem:heading];
        }
        // Keep the previous target and menu untouched until ownership is accepted.
        // Cancellation during construction cannot install a menu without callbacks.
        if (!vellumMenuCommit(token)) return;
        NSApp.mainMenu = menu; vellumTarget = target;
        vellumServiceResult(token, "", 0);
    });
}
void vellumShutdown(void) {
    dispatch_async(dispatch_get_main_queue(), ^{
        for (NSSavePanel *panel in vellumPanels.allValues.copy) { [panel cancel:nil]; [panel orderOut:nil]; }
        [vellumPanels removeAllObjects];
        if (vellumOriginalMenu) NSApp.mainMenu = vellumOriginalMenu;
        vellumTarget = nil;
    });
}
