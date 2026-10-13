//go:build darwin && cgo && !offscreen && vellum_native_checks

#import <Foundation/Foundation.h>
#import <UserNotifications/UserNotifications.h>
#import "services_native.h"
#include <stdint.h>
#include <stdlib.h>
#include <string.h>

extern int vellumServiceActive(uint64_t token);
extern void vellumServiceResult(uint64_t token, char *text, int status);
static NSString *const checkBundlePrefix=@"com.vellum.native-checks.notifications.";
static NSString *const checkIDPrefix=@"vellum.notification.";
static const NSUInteger checkReplyMax=64u<<10;

static NSDictionary *checkBundle(void) {
    NSBundle *bundle=NSBundle.mainBundle;
    BOOL app=[bundle.bundleURL.pathExtension.lowercaseString isEqual:@"app"] && [[bundle objectForInfoDictionaryKey:@"CFBundlePackageType"] isEqual:@"APPL"];
    return @{@"AppBundle":@(app),@"BundleIdentifier":bundle.bundleIdentifier ?: @""};
}
static char *checkWire(id value) {
    NSData *data=value ? [NSJSONSerialization dataWithJSONObject:value options:0 error:nil] : nil;
    if (!data || data.length>checkReplyMax) return NULL;
    NSString *text=[[NSString alloc] initWithData:data encoding:NSUTF8StringEncoding];
    const char *utf8=text.UTF8String;
    return utf8 ? strdup(utf8) : NULL;
}
char *vellumNotificationCheckBundle(void) {
    @autoreleasepool { @try { return checkWire(checkBundle()); } @catch(NSException *exception) { return NULL; } }
}
static NSArray *checkIDs(const char *json) {
    if (!json || strnlen(json,8193)>8192) return nil;
    id value=[NSJSONSerialization JSONObjectWithData:[NSData dataWithBytes:json length:strlen(json)] options:0 error:nil];
    if (![value isKindOfClass:NSDictionary.class] || [value count]!=1) return nil;
    id ids=value[@"IDs"];
    if (![ids isKindOfClass:NSArray.class] || ![ids count] || [ids count]>32) return nil;
    NSMutableSet *seen=[NSMutableSet new];
    for (id identifier in ids) {
        if (![identifier isKindOfClass:NSString.class] || ![identifier length] || [identifier length]>128 || [seen containsObject:identifier]) return nil;
        const char *bytes=[identifier UTF8String];
        if (!bytes || strlen(bytes)!=[identifier length] || strlen(bytes)>128) return nil;
        for (const char *c=bytes;*c;c++) if (!(*c>='a'&&*c<='z'||*c>='A'&&*c<='Z'||*c>='0'&&*c<='9'||*c=='.'||*c=='_'||*c=='-')) return nil;
        [seen addObject:identifier];
    }
    return ids;
}
int vellumNotificationCheckValidate(const char *json) {
    @autoreleasepool { @try { return checkIDs(json)!=nil; } @catch(NSException *exception) { return 0; } }
}
static NSData *checkText(NSString *value,NSUInteger limit) {
    // Refuse an oversized NSString before allocating its UTF-8 representation.
    if (![value isKindOfClass:NSString.class] || value.length>limit || [value rangeOfString:[NSString stringWithFormat:@"%C",(unichar)0]].location!=NSNotFound) return nil;
    NSData *bytes=[value dataUsingEncoding:NSUTF8StringEncoding allowLossyConversion:NO];
    return bytes && bytes.length<=limit ? bytes : nil;
}
static NSArray *checkItems(NSArray *requests,NSSet *wanted,NSUInteger *budget) {
    if (![requests isKindOfClass:NSArray.class] || requests.count>4096) return nil;
    NSMutableArray *result=[NSMutableArray new];NSMutableSet *seen=[NSMutableSet new];
    for (UNNotificationRequest *request in requests) {
        if (![request isKindOfClass:UNNotificationRequest.class]) return nil;
        NSString *identifier=request.identifier;
        if (![wanted containsObject:identifier]) continue;
        if ([seen containsObject:identifier] || result.count>=wanted.count) return nil;
        UNNotificationContent *content=request.content;
        NSData *title=checkText(content.title,256),*subtitle=checkText(content.subtitle,256),*body=checkText(content.body,4096);
        if (!title || !subtitle || !body) return nil;
        // Six bytes per input UTF-8 byte covers JSON escaping. Fixed row/ID
        // overhead is reserved before adding a dictionary; never truncate.
        NSUInteger cost=512+6*(title.length+subtitle.length+body.length);
        if (cost>*budget) return nil;
        *budget-=cost;
        [result addObject:@{@"ID":[identifier substringFromIndex:checkIDPrefix.length],@"Title":content.title,@"Subtitle":content.subtitle,@"Body":content.body,@"Sound":@((BOOL)(content.sound!=nil))}];
        [seen addObject:identifier];
    }
    return [result sortedArrayUsingComparator:^NSComparisonResult(NSDictionary *a,NSDictionary *b) { return [a[@"ID"] compare:b[@"ID"]]; }];
}
static NSSet *checkWanted(NSArray *ids) {
    NSMutableSet *wanted=[NSMutableSet new];
    for (NSString *identifier in ids) [wanted addObject:[checkIDPrefix stringByAppendingString:identifier]];
    return wanted;
}
void vellumNotificationCheck(uint64_t token,const char *json) {
    if (!json || strnlen(json,8193)>8192) { vellumServiceResult(token,"",json?3:2);return; }
    vellumWithServiceJSON(token,json,^(NSData *data) {
        dispatch_async(dispatch_get_main_queue(),^{ @autoreleasepool { @try {
            if (!vellumServiceActive(token)) { vellumServiceResult(token,"",2);return; }
            NSString *wire=[[NSString alloc] initWithData:data encoding:NSUTF8StringEncoding];
            NSArray *ids=wire ? checkIDs(wire.UTF8String) : nil;
            NSDictionary *bundle=checkBundle();NSString *identifier=bundle[@"BundleIdentifier"];
            if (!ids || ![bundle[@"AppBundle"] boolValue] || ![identifier hasPrefix:checkBundlePrefix] || !checkText(identifier,256)) { vellumServiceResult(token,"",2);return; }
            BOOL mainThread=NSThread.isMainThread;
            NSSet *wanted=checkWanted(ids);
            UNUserNotificationCenter *center=UNUserNotificationCenter.currentNotificationCenter;
            [center getPendingNotificationRequestsWithCompletionHandler:^(NSArray<UNNotificationRequest *> *pending) { @autoreleasepool { @try {
                if (!vellumServiceActive(token)) { vellumServiceResult(token,"",2);return; }
                __block NSUInteger budget=checkReplyMax-2048;
                NSArray *pendingRows=checkItems(pending,wanted,&budget);
                if (!pendingRows) { vellumServiceResult(token,"",3);return; }
                [center getDeliveredNotificationsWithCompletionHandler:^(NSArray<UNNotification *> *delivered) { @autoreleasepool { @try {
                    if (!vellumServiceActive(token)) { vellumServiceResult(token,"",2);return; }
                    if (![delivered isKindOfClass:NSArray.class] || delivered.count>4096) { vellumServiceResult(token,"",3);return; }
                    NSMutableArray *requests=[NSMutableArray new];
                    for (UNNotification *notification in delivered) {
                        if (![notification isKindOfClass:UNNotification.class]) { vellumServiceResult(token,"",2);return; }
                        if ([wanted containsObject:notification.request.identifier]) [requests addObject:notification.request];
                    }
                    NSArray *deliveredRows=checkItems(requests,wanted,&budget);
                    char *reply=deliveredRows ? checkWire(@{@"BundleIdentifier":identifier,@"MainThreadDispatch":@(mainThread),@"Pending":pendingRows,@"Delivered":deliveredRows}) : NULL;
                    vellumServiceResult(token,reply,reply?0:3);free(reply);
                } @catch(NSException *exception) { vellumServiceResult(token,"",2); } }}];
            } @catch(NSException *exception) { vellumServiceResult(token,"",2); } }}];
        } @catch(NSException *exception) { vellumServiceResult(token,"",2); } }});
    });
}

// Pure native data-object fixtures. No center, permission, app or OS state access.
char *vellumNotificationCheckFixture(int mode) {
    @autoreleasepool { @try {
        if (mode<0 || mode>2) return NULL;
        NSMutableArray *ids=[NSMutableArray new],*requests=[NSMutableArray new];
        int count=mode==1?32:2;
        for (int i=0;i<count;i++) {
            NSString *identifier=[NSString stringWithFormat:@"fixture-%d",i];[ids addObject:identifier];
            UNMutableNotificationContent *content=[UNMutableNotificationContent new];
            content.title=@"Vellum 🙂";content.subtitle=@"\"quoted\"";
            content.body=mode==1 ? [@"x" stringByPaddingToLength:4096 withString:@"x" startingAtIndex:0] : @"detached\nbody";
            [requests addObject:[UNNotificationRequest requestWithIdentifier:[checkIDPrefix stringByAppendingString:identifier] content:content trigger:nil]];
        }
        if (mode==2) [requests addObject:requests[0]];
        NSUInteger budget=checkReplyMax-2048;NSArray *rows=checkItems(requests,checkWanted(ids),&budget);
        return rows ? checkWire(@{@"BundleIdentifier":[checkBundlePrefix stringByAppendingString:@"fixture"],@"MainThreadDispatch":@YES,@"Pending":@[],@"Delivered":rows}) : NULL;
    } @catch(NSException *exception) { return NULL; } }
}
