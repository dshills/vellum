#ifdef VELLUM_NATIVE_CHECKS
#import <Cocoa/Cocoa.h>
// Main-thread, opt-in acceptance hooks. Ordinary builds contain none of them.
int vellumClipboardCheckBeforeWrite(NSPasteboard *board, NSArray *entries);
int vellumClipboardCheckBeforeClear(NSPasteboard *board);
int vellumClipboardCheckAfterClear(NSPasteboard *board, NSInteger clearedRevision);
int vellumClipboardCheckBeforeItems(NSPasteboard *board);
int vellumClipboardCheckAfterWrite(NSPasteboard *board);
#endif
