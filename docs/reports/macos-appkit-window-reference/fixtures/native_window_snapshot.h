#ifndef VELLUM_NATIVE_WINDOW_SNAPSHOT_H
#define VELLUM_NATIVE_WINDOW_SNAPSHOT_H

#import <Cocoa/Cocoa.h>
#define VELLUM_WINDOW_CLASS_BYTES_MAX 256u

// Keep the sampled DTO independent of NSApplication and NSWindow access so its
// actual production serializer can be exercised with inert Foundation values.
// Class names are descriptive metadata, never a process-wide lifetime identity.
static NSDictionary *vellumWindowSnapshot(NSString *title,BOOL visible,BOOL miniaturized,BOOL key,NSWindowOcclusionState occlusion,NSRect frame,NSString *windowClass,NSString *contentViewClass,NSInteger windowNumber) {
    if (!windowClass.length || [windowClass lengthOfBytesUsingEncoding:NSUTF8StringEncoding]>VELLUM_WINDOW_CLASS_BYTES_MAX || [contentViewClass lengthOfBytesUsingEncoding:NSUTF8StringEncoding]>VELLUM_WINDOW_CLASS_BYTES_MAX) return nil;
    // A boxed C comparison is an integer NSNumber, not a JSON boolean.
    return @{@"Title":title?:@"",@"Visible":@(visible),@"Miniaturized":@(miniaturized),@"Key":@(key),@"OcclusionVisible":(occlusion&NSWindowOcclusionStateVisible)?@YES:@NO,@"Frame":@[@(frame.origin.x),@(frame.origin.y),@(frame.size.width),@(frame.size.height)],@"IdentityAvailable":@YES,@"WindowClass":windowClass,@"ContentViewClass":contentViewClass?:@"",@"WindowNumber":@(windowNumber)};
}

#endif
