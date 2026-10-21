#import <Foundation/Foundation.h>
#import <Metal/Metal.h>
#import <QuartzCore/QuartzCore.h>
#include <math.h>
#include <string.h>

// Standalone offscreen diagnostic. No framework, window, pool, or shipping policy.
static const NSUInteger Width = 1519, Height = 1266, Warmup = 128, Measured = 120;
static void Require(BOOL valid, NSString *why) {
    if (!valid) @throw [NSException exceptionWithName:@"CopyControl" reason:why userInfo:nil];
}
static double Now(void) { return CACurrentMediaTime(); }
static double NS(double begin, double end) {
    Require(isfinite(begin) && isfinite(end) && end >= begin, @"nonfinite/reversed clock");
    return (end - begin) * 1e9;
}

static NSDictionary *Transfer(id<MTLCommandQueue> queue, id<MTLBuffer> source,
                             id<MTLTexture> texture, id<MTLBuffer> readback,
                             const void *expected, NSUInteger bytes) {
    double begin = Now();
    id<MTLCommandBuffer> command = [queue commandBuffer];
    Require(command != nil, @"command buffer unavailable");
    id<MTLBlitCommandEncoder> blit = [command blitCommandEncoder];
    Require(blit != nil, @"blit encoder unavailable");
    [blit copyFromBuffer:source sourceOffset:0 sourceBytesPerRow:Width * 4
      sourceBytesPerImage:bytes sourceSize:MTLSizeMake(Width, Height, 1)
      toTexture:texture destinationSlice:0 destinationLevel:0 destinationOrigin:MTLOriginMake(0, 0, 0)];
    [blit copyFromTexture:texture sourceSlice:0 sourceLevel:0
      sourceOrigin:MTLOriginMake(0, 0, 0) sourceSize:MTLSizeMake(Width, Height, 1)
      toBuffer:readback destinationOffset:0 destinationBytesPerRow:Width * 4 destinationBytesPerImage:bytes];
    [blit endEncoding];
    [command commit];
    [command waitUntilCompleted];
    double completed = Now();
    Require(command.status == MTLCommandBufferStatusCompleted && command.error == nil,
            [NSString stringWithFormat:@"GPU command failed: %@", command.error]);
    Require(readback.contents != NULL, @"readback not mapped");
    BOOL verified = memcmp(readback.contents, expected, bytes) == 0;
    double checked = Now();
    Require(verified, @"full GPU readback mismatch");
    return @{@"completed": @YES, @"verified_bytes": @(bytes),
             @"submit_wait_ns": @(NS(begin, completed)), @"verify_ns": @(NS(completed, checked))};
}

int main(int argc, const char *argv[]) {
    @autoreleasepool {
        NSMutableDictionary *doc = [@{@"kind": @"offscreen-metal-copy-control", @"release_qualified": @NO,
            @"runtime_returned": @NO, @"records": [NSMutableArray array]} mutableCopy];
        int result = 2;
        @try {
            Require(argc == 2, @"expected writecombined or default");
            NSString *mode = [NSString stringWithUTF8String:argv[1]];
            Require([mode isEqual:@"writecombined"] || [mode isEqual:@"default"], @"unknown cache mode");
            doc[@"mode"] = mode;
            id<MTLDevice> device = MTLCreateSystemDefaultDevice();
            Require(device != nil && device.hasUnifiedMemory, @"unified-memory Metal device required");
            id<MTLCommandQueue> queue = [device newCommandQueue];
            Require(queue != nil, @"queue unavailable");
            NSUInteger bytes = Width * Height * 4;
            MTLTextureDescriptor *descriptor = [MTLTextureDescriptor texture2DDescriptorWithPixelFormat:MTLPixelFormatRGBA8Unorm_sRGB
                width:Width height:Height mipmapped:NO];
            descriptor.storageMode = MTLStorageModePrivate;
            descriptor.usage = MTLTextureUsageShaderRead | MTLTextureUsageRenderTarget;
            id<MTLTexture> texture = [device newTextureWithDescriptor:descriptor];
            id<MTLBuffer> readback = [device newBufferWithLength:bytes options:MTLResourceStorageModeShared | MTLResourceCPUCacheModeDefaultCache];
            Require(texture != nil && readback != nil && readback.contents != NULL, @"persistent transfer resources unavailable");
            NSMutableData *first = [NSMutableData dataWithLength:bytes], *second = [NSMutableData dataWithLength:bytes];
            uint8_t *a = first.mutableBytes, *b = second.mutableBytes;
            for (NSUInteger i = 0; i < bytes; i += 4) {
                a[i] = (i / 4 + 11) % 256; a[i+1] = (i / (Width * 4) + 71) % 256;
                a[i+2] = (i / 4 * 17 + 23) % 256; a[i+3] = 255;
                b[i] = 255 - a[i]; b[i+1] = 255 - a[i+1]; b[i+2] = 255 - a[i+2]; b[i+3] = 255;
            }
            NSMutableData *negative = [first mutableCopy];
            ((uint8_t *)negative.mutableBytes)[bytes-1] ^= 1;
            Require(memcmp(first.bytes, negative.bytes, bytes) != 0, @"readback comparator missed corrupt byte");
            ((uint8_t *)negative.mutableBytes)[bytes-1] ^= 1;
            Require(memcmp(first.bytes, negative.bytes, bytes) == 0, @"readback comparator rejected restoration");
            doc[@"comparator_negative_control"] = @YES;
            doc[@"device"] = device.name; doc[@"unified_memory"] = @(device.hasUnifiedMemory);
            doc[@"width"] = @(Width); doc[@"height"] = @(Height); doc[@"pitch"] = @(Width*4);
            doc[@"bytes"] = @(bytes); doc[@"pixel_format"] = @(texture.pixelFormat);
            doc[@"warmup"] = @(Warmup); doc[@"measured"] = @(Measured); doc[@"begin"] = @(Now());
            MTLCPUCacheMode cache = [mode isEqual:@"writecombined"] ? MTLCPUCacheModeWriteCombined : MTLCPUCacheModeDefaultCache;
            MTLResourceOptions options = MTLResourceStorageModeShared |
                (cache == MTLCPUCacheModeWriteCombined ? MTLResourceCPUCacheModeWriteCombined : MTLResourceCPUCacheModeDefaultCache);
            for (NSUInteger i = 0; i < Warmup + Measured; ++i) {
                @autoreleasepool {
                    double begin = Now();
                    id<MTLBuffer> source = [device newBufferWithLength:bytes options:options];
                    double created = Now();
                    Require(source != nil && source.length == bytes && source.storageMode == MTLStorageModeShared && source.cpuCacheMode == cache,
                            @"source buffer options/length drift");
                    void *mapped = source.contents;
                    double mappedAt = Now();
                    Require(mapped != NULL, @"source map unavailable");
                    double firstBegin = Now();
                    memcpy(mapped, first.bytes, bytes);
                    double firstEnd = Now();
                    NSDictionary *firstGPU = Transfer(queue, source, texture, readback, first.bytes, bytes);
                    // First transfer has completed: no in-flight CPU overwrite or resource reuse.
                    double secondBegin = Now();
                    memcpy(mapped, second.bytes, bytes);
                    double secondEnd = Now();
                    NSDictionary *secondGPU = Transfer(queue, source, texture, readback, second.bytes, bytes);
                    [doc[@"records"] addObject:@{@"sequence": @(i+1), @"phase": @(i < Warmup ? 1 : 2),
                        @"storage_mode": @(source.storageMode), @"cpu_cache_mode": @(source.cpuCacheMode),
                        @"begin": @(begin), @"end": @(Now()), @"create_ns": @(NS(begin, created)),
                        @"map_ns": @(NS(created, mappedAt)), @"first_copy_ns": @(NS(firstBegin, firstEnd)),
                        @"subsequent_copy_ns": @(NS(secondBegin, secondEnd)),
                        @"first_gpu": firstGPU, @"subsequent_gpu": secondGPU}];
                }
            }
            doc[@"end"] = @(Now()); doc[@"runtime_returned"] = @YES; result = 0;
        } @catch (NSException *failure) {
            doc[@"error"] = failure.reason ?: failure.name;
            doc[@"end"] = @(Now());
        }
        NSError *error = nil;
        NSData *data = [NSJSONSerialization dataWithJSONObject:doc options:0 error:&error];
        if (data == nil || error != nil) return 3;
        fwrite(data.bytes, 1, data.length, stdout); fputc('\n', stdout); fflush(stdout);
        return result;
    }
}
