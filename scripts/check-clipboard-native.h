#import <Cocoa/Cocoa.h>
// Fence Objective-C exceptions before returning to Swift. No exception payload
// or original clipboard metadata is logged. nil/false signals a refused call.
NSPasteboard *VellumGuardGeneral(void);
NSPasteboard *VellumGuardPrivate(void);
BOOL VellumGuardRelease(NSPasteboard *board);
NSPasteboardItem *VellumGuardItem(void);
NSArray<NSPasteboardItem *> *VellumGuardItems(NSPasteboard *board);
NSArray<NSPasteboardType> *VellumGuardBoardTypes(NSPasteboard *board);
NSArray<NSPasteboardType> *VellumGuardItemTypes(NSPasteboardItem *item);
NSInteger VellumGuardRevision(NSPasteboard *board);
NSData *VellumGuardData(NSPasteboardItem *item,NSPasteboardType type);
NSData *VellumGuardBoardData(NSPasteboard *board,NSPasteboardType type);
BOOL VellumGuardSetData(NSPasteboardItem *item,NSData *data,NSPasteboardType type);
BOOL VellumGuardWrite(NSPasteboard *board,NSArray<NSPasteboardItem *> *items);
NSInteger VellumGuardClear(NSPasteboard *board);
