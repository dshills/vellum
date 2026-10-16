//go:build darwin && cgo

#import "native_window_snapshot.h"
#include <stdlib.h>
#include <string.h>

char *vellumNativeWindowSnapshotFixture(void) {
    // Foundation values only: no NSApplication, NSWindow, native dispatch or
    // window-server queries. This calls the actual production serializer.
    @autoreleasepool {
        NSMutableArray *windows=[NSMutableArray array];
        for (NSNumber *mask in @[@0,@1,@(NSWindowOcclusionStateVisible),@(NSWindowOcclusionStateVisible|1)]) {
            [windows addObject:vellumWindowSnapshot(@"fixture",YES,NO,YES,[mask unsignedLongLongValue],NSMakeRect(10,20,30,40),@"FixtureWindow",@"FixtureView",41+windows.count)];
        }
        NSRect frame=NSMakeRect(1,2,3,4);
        NSString *ascii256=[@"" stringByPaddingToLength:256 withString:@"a" startingAtIndex:0];
        NSString *ascii257=[ascii256 stringByAppendingString:@"a"];
        NSString *utf8_256=[@"" stringByPaddingToLength:128 withString:@"é" startingAtIndex:0];
        NSString *utf8_258=[utf8_256 stringByAppendingString:@"é"];
        NSArray *identities=@[
            vellumWindowSnapshot(@"",NO,NO,NO,0,frame,@"FixtureWindow",nil,-1),
            vellumWindowSnapshot(@"",NO,NO,NO,0,frame,ascii256,utf8_256,NSIntegerMax),
            vellumWindowSnapshot(@"",NO,NO,NO,0,frame,utf8_256,ascii256,NSIntegerMin),
            vellumWindowSnapshot(@"",NO,NO,NO,0,frame,@"FixtureWindow",@"",0)
        ];
        NSArray *refused=@[
            vellumWindowSnapshot(@"",NO,NO,NO,0,frame,nil,@"View",1)==nil?@YES:@NO,
            vellumWindowSnapshot(@"",NO,NO,NO,0,frame,@"",@"View",1)==nil?@YES:@NO,
            vellumWindowSnapshot(@"",NO,NO,NO,0,frame,ascii257,@"View",1)==nil?@YES:@NO,
            vellumWindowSnapshot(@"",NO,NO,NO,0,frame,utf8_258,@"View",1)==nil?@YES:@NO,
            vellumWindowSnapshot(@"",NO,NO,NO,0,frame,@"Window",ascii257,1)==nil?@YES:@NO,
            vellumWindowSnapshot(@"",NO,NO,NO,0,frame,@"Window",utf8_258,1)==nil?@YES:@NO
        ];
        NSData *data=[NSJSONSerialization dataWithJSONObject:@{@"MainThread":@YES,@"Windows":windows,@"IdentityCases":identities,@"RefusedMetadata":refused} options:0 error:nil];
        NSString *text=data?[[NSString alloc]initWithData:data encoding:NSUTF8StringEncoding]:nil;
        return text.UTF8String?strdup(text.UTF8String):NULL;
    }
}
