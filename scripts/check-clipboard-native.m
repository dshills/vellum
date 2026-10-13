#import "check-clipboard-native.h"
NSPasteboard *VellumGuardGeneral(void) { @try { return NSPasteboard.generalPasteboard; } @catch(NSException *e) { return nil; } }
NSPasteboard *VellumGuardPrivate(void) { @try { return NSPasteboard.pasteboardWithUniqueName; } @catch(NSException *e) { return nil; } }
BOOL VellumGuardRelease(NSPasteboard *board) { @try { [board releaseGlobally];return YES; } @catch(NSException *e) { return NO; } }
NSPasteboardItem *VellumGuardItem(void) { @try { return [NSPasteboardItem new]; } @catch(NSException *e) { return nil; } }
NSArray<NSPasteboardItem *> *VellumGuardItems(NSPasteboard *board) { @try { return board.pasteboardItems.copy ?: @[]; } @catch(NSException *e) { return nil; } }
NSArray<NSPasteboardType> *VellumGuardBoardTypes(NSPasteboard *board) { @try { return board.types.copy ?: @[]; } @catch(NSException *e) { return nil; } }
NSArray<NSPasteboardType> *VellumGuardItemTypes(NSPasteboardItem *item) { @try { return item.types.copy ?: @[]; } @catch(NSException *e) { return nil; } }
NSInteger VellumGuardRevision(NSPasteboard *board) { @try { return board.changeCount; } @catch(NSException *e) { return -1; } }
NSData *VellumGuardData(NSPasteboardItem *item,NSPasteboardType type) { @try { return [item dataForType:type].copy; } @catch(NSException *e) { return nil; } }
NSData *VellumGuardBoardData(NSPasteboard *board,NSPasteboardType type) { @try { return [board dataForType:type].copy; } @catch(NSException *e) { return nil; } }
BOOL VellumGuardSetData(NSPasteboardItem *item,NSData *data,NSPasteboardType type) { @try { return [item setData:data forType:type]; } @catch(NSException *e) { return NO; } }
BOOL VellumGuardWrite(NSPasteboard *board,NSArray<NSPasteboardItem *> *items) { @try { return [board writeObjects:items]; } @catch(NSException *e) { return NO; } }
NSInteger VellumGuardClear(NSPasteboard *board) { @try { return [board clearContents]; } @catch(NSException *e) { return -1; } }
