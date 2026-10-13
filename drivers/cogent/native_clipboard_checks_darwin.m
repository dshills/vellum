//go:build darwin && cgo && !offscreen && vellum_native_checks

#import "native_clipboard_checks.h"
#import "services_native.h"
#import <UniformTypeIdentifiers/UniformTypeIdentifiers.h>
#include <stdint.h>
#include <stdlib.h>
#include <string.h>
#include <fcntl.h>
#include <unistd.h>
#include <sys/stat.h>
#include <errno.h>
extern int vellumServiceActive(uint64_t token);
extern void vellumServiceResult(uint64_t token,char *text,int status);
static NSString *guardNonce,*guardPath,*guardType,*guardPhase,*guardKind;
static NSData *guardMarker,*guardLegacy;
static NSInteger guardRevision,guardItemCount;
static NSString *guardPriorPhase,*guardPriorKind;
static NSInteger guardPriorRevision,guardPriorCount;

// Single native main-thread writer while the guardian waits for this process.
// The parent/recovery writer starts only after the fixture is fully terminated.
static BOOL guardRecord(NSString *phase,NSString *kind,NSInteger revision,NSInteger count) {
    if (!guardPath || !NSThread.isMainThread) return NO;
    struct stat info;
    NSString *directory=guardPath.stringByDeletingLastPathComponent;
    if (lstat(directory.fileSystemRepresentation,&info) || !S_ISDIR(info.st_mode) || (info.st_mode&0777)!=0700 || info.st_uid!=geteuid()) return NO;
    if (lstat(guardPath.fileSystemRepresentation,&info) || !S_ISREG(info.st_mode) || (info.st_mode&0777)!=0600 || info.st_uid!=geteuid()) return NO;
    NSData *wire=[NSJSONSerialization dataWithJSONObject:@{@"version":@1,@"nonce":guardNonce,@"phase":phase,@"kind":kind,@"revision":@(revision),@"itemCount":@(count)} options:0 error:nil];
    if (!wire || wire.length>2048) return NO;
    // A crashed writer may leave its temporary behind. Allocate a fresh name;
    // never unlink another writer's leftover before admitting this operation.
    char *temporary=strdup([[guardPath stringByAppendingString:@".new.XXXXXX"] fileSystemRepresentation]);
    if (!temporary) return NO;
    int fd=mkstemp(temporary);
    if (fd<0) { free(temporary);return NO; }
    const uint8_t *bytes=wire.bytes;size_t remaining=wire.length;BOOL good=YES;
    while (remaining) {
        ssize_t written=write(fd,bytes,remaining);
        if (written<0 && errno==EINTR) continue;
        if (written<=0) { good=NO;break; }
        bytes+=written;remaining-=written;
    }
    if (good && fsync(fd)) good=NO;
    if (close(fd)) good=NO;
    if (good && rename(temporary,guardPath.fileSystemRepresentation)) good=NO;
    if (good) {
        int directoryFD=open(directory.fileSystemRepresentation,O_RDONLY|O_DIRECTORY|O_NOFOLLOW);
        if (directoryFD<0) good=NO;
        else { if (fsync(directoryFD)) good=NO; if (close(directoryFD)) good=NO; }
    }
    if (!good) unlink(temporary);
    free(temporary);
    if (!good) return NO;
    guardPhase=phase;guardKind=kind;guardRevision=revision;guardItemCount=count;return YES;
}
static NSString *guardOwnedKind(NSPasteboard *board) {
    NSPasteboardItem *first=board.pasteboardItems.firstObject;
    if (!first) return nil;
    if ([first.types containsObject:guardType] && [[first dataForType:guardType] isEqualToData:guardMarker]) return @"rich";
    if (board.pasteboardItems.count==1 && [first.types containsObject:NSPasteboardTypeString] && [[first dataForType:NSPasteboardTypeString] isEqualToData:guardLegacy]) return @"legacy";
    return nil;
}
static BOOL guardMatchesAt(NSPasteboard *board,NSString *phase,NSString *kind,NSInteger revision,NSInteger count) {
    if (!guardNonce || board.changeCount!=revision) return NO;
    if ([phase isEqualToString:@"original"]) return YES; // Guardian captured this exact revision.
    if ([phase isEqualToString:@"empty"]) return !board.pasteboardItems.count && !board.types.count && board.changeCount==revision;
    if (![phase isEqualToString:@"owned"]) return NO;
    NSString *observed=guardOwnedKind(board);
    return [observed isEqualToString:kind] && board.pasteboardItems.count==count && board.changeCount==revision;
}
static BOOL guardMatches(NSPasteboard *board) {
    return guardMatchesAt(board,guardPhase,guardKind,guardRevision,guardItemCount);
}
int vellumClipboardCheckBeforeWrite(NSPasteboard *board,NSArray *entries) {
    if (!NSThread.isMainThread || !guardMatches(board)) return 3;
    BOOL marked=NO;
    for (NSDictionary *entry in entries) if ([entry[@"type"] isEqualToString:guardType] && [entry[@"data"] isEqualToData:guardMarker]) marked=YES;
    BOOL legacy=entries.count==1 && [entries[0][@"type"] isEqualToString:NSPasteboardTypeString] && [entries[0][@"data"] isEqualToData:guardLegacy];
    if (!marked && !legacy) return 3;
    guardPriorPhase=guardPhase;guardPriorKind=guardKind;guardPriorRevision=guardRevision;guardPriorCount=guardItemCount;
    // Unknown interrupted admission is retained by recovery, never guessed.
    return guardRecord(@"admitted",legacy?@"legacy":@"rich",board.changeCount,0) ? 0 : 2;
}
int vellumClipboardCheckBeforeClear(NSPasteboard *board) {
    return NSThread.isMainThread && [guardPhase isEqualToString:@"admitted"] && guardMatchesAt(board,guardPriorPhase,guardPriorKind,guardPriorRevision,guardPriorCount) ? 0 : 3;
}
int vellumClipboardCheckAfterClear(NSPasteboard *board,NSInteger clearedRevision) {
    if (!NSThread.isMainThread || clearedRevision<0 || !guardMatchesAt(board,@"empty",@"empty",clearedRevision,0)) return 3;
    return guardRecord(@"empty",@"empty",clearedRevision,0) ? 0 : 2;
}
int vellumClipboardCheckBeforeItems(NSPasteboard *board) {
    return NSThread.isMainThread && [guardPhase isEqualToString:@"empty"] && guardMatches(board) ? 0 : 3;
}
int vellumClipboardCheckAfterWrite(NSPasteboard *board) {
    NSInteger revision=board.changeCount;NSString *kind=guardOwnedKind(board);
    if (!kind || board.changeCount!=revision) return 3;
    return guardRecord(@"owned",kind,revision,board.pasteboardItems.count) ? 0 : 2;
}
static BOOL guardConfigure(NSDictionary *d) {
    if (guardNonce || ![d isKindOfClass:NSDictionary.class] || d.count!=3) return NO;
    NSString *nonce=d[@"Nonce"],*path=d[@"StatePath"];NSNumber *revision=d[@"Revision"];
    if (![nonce isKindOfClass:NSString.class] || nonce.length!=64 || [nonce rangeOfCharacterFromSet:[NSCharacterSet characterSetWithCharactersInString:@"0123456789abcdef"].invertedSet].location!=NSNotFound) return NO;
    if (![path isKindOfClass:NSString.class] || !path.isAbsolutePath || ![path.lastPathComponent isEqualToString:@"state.json"]) return NO;
    NSData *pathBytes=[path dataUsingEncoding:NSUTF8StringEncoding];
    if (!pathBytes.length || pathBytes.length>4096 || memchr(pathBytes.bytes,0,pathBytes.length)) return NO;
    if (![revision isKindOfClass:NSNumber.class] || CFGetTypeID((__bridge CFTypeRef)revision)==CFBooleanGetTypeID() || revision.longLongValue<0 || revision.longLongValue>9007199254740991LL || revision.doubleValue!=revision.longLongValue || NSPasteboard.generalPasteboard.changeCount!=revision.longLongValue) return NO;
    NSString *type=[UTType typeWithMIMEType:@"application/x-vellum-clipboard-test-owner"].identifier;if (!type) return NO;
    guardNonce=nonce;guardPath=path;guardType=type;guardMarker=[nonce dataUsingEncoding:NSASCIIStringEncoding];
    guardLegacy=[[@"Vellum legacy clipboard check " stringByAppendingString:nonce] dataUsingEncoding:NSUTF8StringEncoding];
    if (!guardRecord(@"original",@"",revision.longLongValue,0)) { guardNonce=nil;return NO; }
    return YES;
}
static BOOL guardSeed(NSData *png) {
    if (!png.length || png.length>8192) return NO;
    NSPasteboardItem *first=[NSPasteboardItem new],*second=[NSPasteboardItem new];
    NSData *text=[[@"Vellum first item " stringByAppendingString:guardNonce] dataUsingEncoding:NSUTF8StringEncoding];
    if (![first setData:text forType:NSPasteboardTypeString] || ![first setData:guardMarker forType:guardType] || ![second setData:png forType:NSPasteboardTypePNG]) return NO;
    NSArray *entries=@[@{@"type":NSPasteboardTypeString,@"data":text},@{@"type":guardType,@"data":guardMarker}];
    NSPasteboard *board=NSPasteboard.generalPasteboard;
    if (vellumClipboardCheckBeforeWrite(board,entries)) return NO;
    if (vellumClipboardCheckBeforeClear(board)) return NO;
    NSInteger cleared=[board clearContents];if (vellumClipboardCheckAfterClear(board,cleared)) return NO;
    if (vellumClipboardCheckBeforeItems(board)) return NO;
    BOOL written=[board writeObjects:@[first,second]];
    return !vellumClipboardCheckAfterWrite(board) && written;
}

// Actual guard functions, private pasteboard and owned bytes only. The caller
// supplies an existing0600 state file in its0700 private test directory.
int vellumClipboardGuardPrivateFixture(const char *path) {
    if (!path || !NSThread.isMainThread || guardNonce) return 0;
    @autoreleasepool {
        NSPasteboard *board=nil;int result=0;NSString *stale=nil;
        @try {
            guardNonce=[@"" stringByPaddingToLength:64 withString:@"a" startingAtIndex:0];
            guardPath=[NSString stringWithUTF8String:path];
            guardType=[UTType typeWithMIMEType:@"application/x-vellum-clipboard-test-owner"].identifier;
            guardMarker=[guardNonce dataUsingEncoding:NSASCIIStringEncoding];
            guardLegacy=[[@"Vellum legacy clipboard check " stringByAppendingString:guardNonce] dataUsingEncoding:NSUTF8StringEncoding];
            if (!guardPath || !guardType) return 0;
            stale=[guardPath stringByAppendingString:@".new"];
            int staleFD=open(stale.fileSystemRepresentation,O_WRONLY|O_CREAT|O_EXCL|O_NOFOLLOW,0600);
            if (staleFD<0) { stale=nil;return 0; }
            const char *dummy="private stale journal";
            BOOL staleCreated=write(staleFD,dummy,strlen(dummy))==(ssize_t)strlen(dummy);
            if (close(staleFD)) staleCreated=NO;
            if (!staleCreated) return 0;
            board=NSPasteboard.pasteboardWithUniqueName;
            NSArray *entries=@[@{@"type":guardType,@"data":guardMarker}];
            NSData *foreign=[@"private native foreign writer" dataUsingEncoding:NSUTF8StringEncoding];
            // Interpose a foreign write after admitted fsync, before clear.
            [board clearContents];
            if (!guardRecord(@"original",@"",board.changeCount,0) || vellumClipboardCheckBeforeWrite(board,entries)) return 0;
            NSPasteboardItem *item=[NSPasteboardItem new];[item setData:foreign forType:NSPasteboardTypeString];
            [board clearContents];if (![board writeObjects:@[item]]) return 0;
            NSInteger revision=board.changeCount;
            if (vellumClipboardCheckBeforeClear(board)!=3 || board.changeCount!=revision || ![[board.pasteboardItems.firstObject dataForType:NSPasteboardTypeString] isEqualToData:foreign]) return 0;
            result|=1;
            // Interpose after known-empty fsync, before writeObjects.
            [board clearContents];
            if (!guardRecord(@"original",@"",board.changeCount,0) || vellumClipboardCheckBeforeWrite(board,entries) || vellumClipboardCheckBeforeClear(board)) return 0;
            NSInteger cleared=[board clearContents];
            if (vellumClipboardCheckAfterClear(board,cleared)) return 0;
            item=[NSPasteboardItem new];[item setData:foreign forType:NSPasteboardTypeString];
            [board clearContents];if (![board writeObjects:@[item]]) return 0;
            revision=board.changeCount;
            if (vellumClipboardCheckBeforeItems(board)!=3 || board.changeCount!=revision || ![[board.pasteboardItems.firstObject dataForType:NSPasteboardTypeString] isEqualToData:foreign]) return 0;
            result|=2;
            // A guessed/different clear revision is not accepted as owned empty.
            cleared=[board clearContents];
            if (vellumClipboardCheckAfterClear(board,cleared-1)!=3) return 0;
            result|=4;
            // Uninterposed path still accepts the owned marker and records it.
            if (!guardRecord(@"original",@"",board.changeCount,0) || vellumClipboardCheckBeforeWrite(board,entries) || vellumClipboardCheckBeforeClear(board)) return 0;
            cleared=[board clearContents];
            if (vellumClipboardCheckAfterClear(board,cleared) || vellumClipboardCheckBeforeItems(board)) return 0;
            item=[NSPasteboardItem new];if (![item setData:guardMarker forType:guardType] || ![board writeObjects:@[item]] || vellumClipboardCheckAfterWrite(board) || !guardMatches(board)) return 0;
            result|=8;
            if (![[NSData dataWithContentsOfFile:stale] isEqualToData:[@"private stale journal" dataUsingEncoding:NSUTF8StringEncoding]]) return 0;
            result|=16;
        } @catch(NSException *exception) { result=0; }
        @finally {
            if (stale) unlink(stale.fileSystemRepresentation);
            if (board) { @try { [board releaseGlobally]; } @catch(NSException *exception) { result=0; } }
            guardNonce=nil;guardPath=nil;guardType=nil;guardPhase=nil;guardKind=nil;guardMarker=nil;guardLegacy=nil;guardPriorPhase=nil;guardPriorKind=nil;
            guardRevision=0;guardItemCount=0;guardPriorRevision=0;guardPriorCount=0;
        }
        return result;
    }
}
void vellumNativeClipboardCheck(uint64_t token,const char *json) {
    vellumWithServiceJSON(token,json,^(NSData *wire) {
        dispatch_async(dispatch_get_main_queue(),^{ @autoreleasepool { @try {
            if (!vellumServiceActive(token)) { vellumServiceResult(token,"",2);return; }
            NSDictionary *request=[NSJSONSerialization JSONObjectWithData:wire options:0 error:nil];
            if (![request isKindOfClass:NSDictionary.class]) { vellumServiceResult(token,"",2);return; }
            NSString *op=request[@"Operation"];BOOL good=NO;
            if ([op isEqualToString:@"configure"] && request.count==2) good=guardConfigure(request[@"Config"]);
            else if ([op isEqualToString:@"inspect"] && request.count==1) good=guardNonce!=nil;
            else if ([op isEqualToString:@"seed"] && request.count==2 && [request[@"PNG"] isKindOfClass:NSString.class]) good=guardNonce && guardSeed([[NSData alloc] initWithBase64EncodedString:request[@"PNG"] options:0]);
            if (!good) { vellumServiceResult(token,"",2);return; }
            NSDictionary *state=@{@"MainThread":@(NSThread.isMainThread),@"Configured":@YES,@"Revision":@(guardRevision),@"Phase":guardPhase?:@"",@"Kind":guardKind?:@"",@"ItemCount":@(guardItemCount)};
            NSData *data=[NSJSONSerialization dataWithJSONObject:state options:0 error:nil];NSString *text=data?[[NSString alloc] initWithData:data encoding:NSUTF8StringEncoding]:nil;
            vellumServiceResult(token,(char *)(text.UTF8String?:""),text?0:2);
        } @catch(NSException *exception) { vellumServiceResult(token,"",2); } }});
    });
}
