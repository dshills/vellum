//go:build darwin && cgo

#import <Cocoa/Cocoa.h>
#import <UniformTypeIdentifiers/UniformTypeIdentifiers.h>
#include <stdint.h>
#include <string.h>

// Keep synchronized with app/clipboard.go and clipboard.go. The native parser
// repeats the caps before copying into Cocoa's asynchronously retained payload.
#define VELLUM_CLIP_FORMATS 16
#define VELLUM_CLIP_MIME 127
#define VELLUM_CLIP_REP (8u << 20)
#define VELLUM_CLIP_TOTAL (16u << 20)
#define VELLUM_CLIP_WIRE (4u + VELLUM_CLIP_FORMATS * (6u + VELLUM_CLIP_MIME) + VELLUM_CLIP_TOTAL)

extern int vellumClipboardActive(uint64_t token);
extern void vellumClipboardResult(uint64_t token, void *wire, size_t length, int status);

static uint32_t clipU32(const uint8_t *p) {
    return (uint32_t)p[0] | ((uint32_t)p[1] << 8) | ((uint32_t)p[2] << 16) | ((uint32_t)p[3] << 24);
}
static void clipAppendU32(NSMutableData *out, uint32_t n) {
    uint8_t bytes[4] = {n & 255, (n >> 8) & 255, (n >> 16) & 255, (n >> 24) & 255};
    [out appendBytes:bytes length:4];
}
static BOOL clipMIME(const uint8_t *p, size_t length) {
    size_t slash = 0;
    for (size_t i = 0; i < length; i++) {
        uint8_t c = p[i];
        if (c == '/') { if (!i || i == length - 1 || slash++) return NO; }
        else if (!((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') ||
                   (c && strchr("!#$%&'+-.^_`|~", c)))) return NO;
    }
    return length && slash == 1;
}
static NSString *clipType(NSString *mime) {
    if ([mime isEqualToString:@"text/plain"]) return NSPasteboardTypeString;
    if ([mime isEqualToString:@"image/png"]) return NSPasteboardTypePNG;
    if ([mime isEqualToString:@"text/html"]) return NSPasteboardTypeHTML;
    if (@available(macOS 11.0, *)) return [UTType typeWithMIMEType:mime].identifier;
    return nil;
}
int vellumClipboardCustomMIMEAvailable(void) {
    if (@available(macOS 11.0, *)) return 1;
    return 0;
}
static NSArray<NSDictionary *> *clipParseWithTypes(NSData *wire, NSString *(^resolve)(NSString *)) {
    if (wire.length < 4 || wire.length > VELLUM_CLIP_WIRE) return nil;
    const uint8_t *p = wire.bytes;
    uint32_t count = clipU32(p);
    if (!count || count > VELLUM_CLIP_FORMATS) return nil;
    size_t offset = 4, total = 0;
    NSMutableArray *items = [NSMutableArray new];
    NSMutableSet *types = [NSMutableSet new];
    for (uint32_t i = 0; i < count; i++) {
        if (wire.length - offset < 6) return nil;
        size_t m = p[offset] | ((size_t)p[offset + 1] << 8), n = clipU32(p + offset + 2);
        offset += 6;
        if (!m || m > VELLUM_CLIP_MIME || n > VELLUM_CLIP_REP || n > VELLUM_CLIP_TOTAL - total || m > wire.length - offset) return nil;
        if (!clipMIME(p + offset, m)) return nil;
        NSString *mime = [[NSString alloc] initWithBytes:p + offset length:m encoding:NSASCIIStringEncoding];
        offset += m;
        if (n > wire.length - offset) return nil;
        NSString *type = resolve(mime);
        // Aliases must not overwrite one another on a single pasteboard item.
        if (!type || [types containsObject:type]) return nil;
        [types addObject:type];
        NSData *data = [wire subdataWithRange:NSMakeRange(offset, n)];
        if ([mime isEqualToString:@"text/plain"] &&
            ((n && memchr(data.bytes, 0, n)) || ![[NSString alloc] initWithData:data encoding:NSUTF8StringEncoding])) return nil;
        [items addObject:@{@"mime":mime, @"type":type, @"data":data}];
        offset += n; total += n;
    }
    return offset == wire.length ? items : nil;
}
static NSArray<NSDictionary *> *clipParse(NSData *wire) {
    return clipParseWithTypes(wire, ^NSString *(NSString *mime) { return clipType(mime); });
}
int vellumClipboardValidate(const void *wire, size_t length) {
    if (!wire || length > VELLUM_CLIP_WIRE) return 0;
    @autoreleasepool {
        @try { return clipParse([NSData dataWithBytesNoCopy:(void *)wire length:length freeWhenDone:NO]) != nil; }
        @catch (NSException *exception) { return 0; }
    }
}
int vellumClipboardAliasValidate(const void *bytes, size_t length) {
    if (!bytes || length > VELLUM_CLIP_WIRE) return 0;
    @autoreleasepool {
        NSData *wire = [NSData dataWithBytesNoCopy:(void *)bytes length:length freeWhenDone:NO];
        return clipParseWithTypes(wire, ^NSString *(NSString *mime) { return NSPasteboardTypeString; }) != nil;
    }
}

// The producer pool drains on the originating Go worker. The consumer's strong
// block capture owns the detached bytes until it runs, even after cancellation.
static BOOL clipWithWire(const void *bytes, size_t length, void (^consume)(NSData *)) {
    if (!bytes || length > VELLUM_CLIP_WIRE) return NO;
    @autoreleasepool {
        @try {
            NSData *wire = [NSData dataWithBytes:bytes length:length];
            if (!wire) return NO;
            consume(wire);
            return YES;
        } @catch (NSException *exception) { return NO; }
    }
}

// Delay an injected consumer; exercise the production producer without using a
// pasteboard or dispatching onto the native main queue. Bit flags: alive/copy/free.
int vellumClipboardPayloadFixture(const void *bytes, size_t length) {
    if (!bytes || !length || length > VELLUM_CLIP_WIRE) return 0;
    @autoreleasepool {
        void *input = malloc(length); if (!input) return 0;
        memcpy(input, bytes, length);
        NSData *expected = [[NSData alloc] initWithBytes:bytes length:length];
        __block dispatch_block_t queued = nil;
        __block __weak NSData *weakWire = nil;
        __block BOOL copied = NO;
        BOOL accepted = clipWithWire(input, length, ^(NSData *wire) {
            weakWire = wire;
            queued = [^{ @autoreleasepool { copied = [wire isEqualToData:expected]; } } copy];
        });
        memset(input, 0, length); free(input);
        BOOL alive; @autoreleasepool { alive = weakWire != nil; }
        if (queued) queued(); queued = nil;
        BOOL released; @autoreleasepool { released = weakWire == nil; }
        return accepted ? ((alive ? 1 : 0) | (copied ? 2 : 0) | (released ? 4 : 0)) : 0;
    }
}
static NSData *clipRead(uint64_t token, NSArray<NSDictionary *> *request, int *status) {
    if (!vellumClipboardActive(token)) { *status = 2; return nil; }
    NSPasteboard *board = [NSPasteboard generalPasteboard];
    NSInteger revision = board.changeCount;
    NSPasteboardItem *item = board.pasteboardItems.firstObject;
    NSMutableArray *present = [NSMutableArray new];
    size_t total = 0;
    for (NSDictionary *entry in request) {
        if (![item.types containsObject:entry[@"type"]]) continue;
        // Cocoa may materialize a lazy provider here. Limits apply to accepted
        // data/bridge copies, not to memory or time spent inside the OS provider.
        NSData *data = [item dataForType:entry[@"type"]];
        if (!data) { *status = 2; return nil; }
        if (data.length > VELLUM_CLIP_REP || data.length > VELLUM_CLIP_TOTAL - total) { *status = 4; return nil; }
        [present addObject:@{@"mime":entry[@"mime"], @"data":data}]; total += data.length;
    }
    if (board.changeCount != revision) { *status = 3; return nil; }
    NSMutableData *result = [NSMutableData new];
    clipAppendU32(result, (uint32_t)present.count);
    for (NSDictionary *entry in present) {
        NSData *mime = [entry[@"mime"] dataUsingEncoding:NSASCIIStringEncoding];
        NSData *data = entry[@"data"];
        uint8_t m[2] = {mime.length & 255, (mime.length >> 8) & 255};
        [result appendBytes:m length:2]; clipAppendU32(result, (uint32_t)data.length);
        [result appendData:mime]; [result appendData:data];
    }
    return result;
}
static int clipWrite(uint64_t token, NSArray<NSDictionary *> *entries) {
    NSPasteboardItem *item = [NSPasteboardItem new];
    for (NSDictionary *entry in entries) {
        if (![item setData:entry[@"data"] forType:entry[@"type"]]) return 2;
    }
    if (!vellumClipboardActive(token)) return 2;
    NSPasteboard *board = [NSPasteboard generalPasteboard];
    // This final check bounds cancellation before destructive mutation. The OS
    // clear/write pair is not transactional; failures can leave an empty board.
    if (!vellumClipboardActive(token)) return 2;
    [board clearContents];
    return [board writeObjects:@[item]] ? 0 : 2;
}
void vellumClipboardData(uint64_t token, const void *bytes, size_t length, int write) {
    if (!bytes || length > VELLUM_CLIP_WIRE) { vellumClipboardResult(token, NULL, 0, 4); return; }
    BOOL accepted = clipWithWire(bytes, length, ^(NSData *wire) {
      dispatch_async(dispatch_get_main_queue(), ^{
        @autoreleasepool {
            NSData *result = nil;
            int status = 2;
            @try {
                if (vellumClipboardActive(token)) {
                    NSArray *entries = clipParse(wire);
                    if (entries) {
                        status = 0;
                        if (write) status = clipWrite(token, entries);
                        else result = clipRead(token, entries, &status);
                    }
                }
            } @catch (NSException *exception) { status = 2; result = nil; }
            // Every dispatch, including canceled ones, releases its Go credit.
            vellumClipboardResult(token, (void *)result.bytes, result.length, status);
        }
      });
    });
    if (!accepted) vellumClipboardResult(token, NULL, 0, 2);
}
