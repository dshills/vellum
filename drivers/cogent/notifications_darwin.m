//go:build darwin && cgo

#import <Foundation/Foundation.h>
#import <UserNotifications/UserNotifications.h>
#import "services_native.h"
#include <CoreFoundation/CoreFoundation.h>
#include <stdlib.h>
#include <string.h>

extern int vellumServiceActive(uint64_t token);
extern void vellumServiceResult(uint64_t token, char *text, int status);
static const size_t notificationWireMax=32u<<10;
static NSString *const notificationPrefix=@"vellum.notification.";

static BOOL notificationText(id value, NSUInteger limit) {
    if (![value isKindOfClass:NSString.class]) return NO;
    NSData *utf8=[value dataUsingEncoding:NSUTF8StringEncoding allowLossyConversion:NO];
    return utf8 && utf8.length<=limit &&
        [value rangeOfString:[NSString stringWithFormat:@"%C",(unichar)0]].location==NSNotFound;
}
static BOOL notificationID(id value) {
    if (!notificationText(value,128) || [value length]==0) return NO;
    const char *p=[value UTF8String]; if (!p) return NO;
    for (;*p;p++) if (!(*p>='a'&&*p<='z'||*p>='A'&&*p<='Z'||*p>='0'&&*p<='9'||*p=='.'||*p=='_'||*p=='-')) return NO;
    return YES;
}
static NSDictionary *notificationParse(NSData *data) {
    id root=[NSJSONSerialization JSONObjectWithData:data options:0 error:nil];
    if (![root isKindOfClass:NSDictionary.class] || [root count]!=3) return nil;
    id action=root[@"Action"],n=root[@"Notification"],ids=root[@"IDs"];
    if (![action isKindOfClass:NSNumber.class] || CFGetTypeID((__bridge CFTypeRef)action)==CFBooleanGetTypeID()) return nil;
    double number=[action doubleValue]; if (!(number>=1 && number<=4) || number!=(int)number) return nil;
    if (![n isKindOfClass:NSDictionary.class] || [n count]!=5) return nil;
    if (!notificationText(n[@"ID"],128) || !notificationText(n[@"Title"],256) || !notificationText(n[@"Subtitle"],256) || !notificationText(n[@"Body"],4096)) return nil;
    id sound=n[@"Sound"]; if (![sound isKindOfClass:NSNumber.class] || CFGetTypeID((__bridge CFTypeRef)sound)!=CFBooleanGetTypeID()) return nil;
    if (ids==NSNull.null) ids=@[];
    if (![ids isKindOfClass:NSArray.class] || [ids count]>32) return nil;
    NSMutableSet *seen=[NSMutableSet new];
    for (id identifier in ids) { if (!notificationID(identifier) || [seen containsObject:identifier]) return nil; [seen addObject:identifier]; }
    if (number==3) {
        if ([ids count] || !notificationID(n[@"ID"])) return nil;
        NSCharacterSet *space=NSCharacterSet.whitespaceAndNewlineCharacterSet;
        if (![n[@"Title"] stringByTrimmingCharactersInSet:space].length && ![n[@"Body"] stringByTrimmingCharactersInSet:space].length) return nil;
    } else {
        if ([n[@"ID"] length] || [n[@"Title"] length] || [n[@"Subtitle"] length] || [n[@"Body"] length] || [sound boolValue]) return nil;
        if (number==4 ? [ids count]==0 : [ids count]!=0) return nil;
    }
    return root;
}
static NSData *notificationData(const char *json) {
    if (!json) return nil;
    size_t length=strnlen(json,notificationWireMax+1);
    if (length>notificationWireMax) return nil;
    return [NSData dataWithBytes:json length:length];
}
static UNNotificationRequest *notificationRequest(NSDictionary *n) {
    UNMutableNotificationContent *content=[UNMutableNotificationContent new];
    content.title=n[@"Title"]; content.subtitle=n[@"Subtitle"]; content.body=n[@"Body"];
    if ([n[@"Sound"] boolValue]) content.sound=UNNotificationSound.defaultSound;
    return [UNNotificationRequest requestWithIdentifier:[notificationPrefix stringByAppendingString:n[@"ID"]] content:content trigger:nil];
}
static NSString *notificationSettingsWire(NSInteger auth, NSInteger alert, NSInteger sound) {
    int permission=0;
    switch(auth) {
        case UNAuthorizationStatusNotDetermined:permission=1;break;
        case UNAuthorizationStatusDenied:permission=2;break;
        case UNAuthorizationStatusAuthorized:permission=3;break;
        case UNAuthorizationStatusProvisional:permission=4;break;
        default:break;
    }
    return [NSString stringWithFormat:@"{\"Permission\":%d,\"AlertEnabled\":%s,\"SoundEnabled\":%s}",permission,alert==UNNotificationSettingEnabled?"true":"false",sound==UNNotificationSettingEnabled?"true":"false"];
}
static void notificationReadSettings(UNUserNotificationCenter *center,uint64_t token) {
    [center getNotificationSettingsWithCompletionHandler:^(UNNotificationSettings *settings) {
        @autoreleasepool { @try {
            if (!vellumServiceActive(token) || !settings) { vellumServiceResult(token,"",2);return; }
            NSString *wire=notificationSettingsWire(settings.authorizationStatus,settings.alertSetting,settings.soundSetting);
            vellumServiceResult(token,(char *)wire.UTF8String,0);
        } @catch(NSException *exception) { vellumServiceResult(token,"",2); } }
    }];
}

void vellumNotifications(uint64_t token,const char *json) {
    if (!json || strnlen(json,notificationWireMax+1)>notificationWireMax) { vellumServiceResult(token,"",json?3:2);return; }
    vellumWithServiceJSON(token,json,^(NSData *data) {
        dispatch_async(dispatch_get_main_queue(),^{
            @autoreleasepool { @try {
                if (!vellumServiceActive(token)) { vellumServiceResult(token,"",2);return; }
                NSDictionary *root=notificationParse(data);
                if (!root) { vellumServiceResult(token,"",2);return; }
                NSBundle *bundle=NSBundle.mainBundle;
                if (!bundle.bundleIdentifier.length || ![bundle.bundleURL.pathExtension.lowercaseString isEqualToString:@"app"] || ![[bundle objectForInfoDictionaryKey:@"CFBundlePackageType"] isEqual:@"APPL"]) { vellumServiceResult(token,"",2);return; }
                UNUserNotificationCenter *center=UNUserNotificationCenter.currentNotificationCenter;
                if (!vellumServiceActive(token)) { vellumServiceResult(token,"",2);return; }
                switch([root[@"Action"] intValue]) {
                    case 1:{notificationReadSettings(center,token);break;}
                    case 2: {
                        [center requestAuthorizationWithOptions:UNAuthorizationOptionAlert|UNAuthorizationOptionSound completionHandler:^(BOOL granted,NSError *error) {
                            @autoreleasepool { @try {
                                if (error || !vellumServiceActive(token)) { vellumServiceResult(token,"",2);return; }
                                notificationReadSettings(center,token);
                            } @catch(NSException *exception) { vellumServiceResult(token,"",2); } }
                        }];break;
                    }
                    case 3: {
                        [center getNotificationSettingsWithCompletionHandler:^(UNNotificationSettings *settings) {
                            @autoreleasepool { @try {
                                if (!vellumServiceActive(token) || !settings) { vellumServiceResult(token,"",2);return; }
                                NSInteger permission=settings.authorizationStatus;
                                if (permission==UNAuthorizationStatusDenied || permission==UNAuthorizationStatusNotDetermined) { vellumServiceResult(token,"",4);return; }
                                if (permission!=UNAuthorizationStatusAuthorized && permission!=UNAuthorizationStatusProvisional) { vellumServiceResult(token,"",2);return; }
                                dispatch_async(dispatch_get_main_queue(),^{
                                    @autoreleasepool { @try {
                                        if (!vellumServiceActive(token)) { vellumServiceResult(token,"",2);return; }
                                        UNNotificationRequest *request=notificationRequest(root[@"Notification"]);
                                        [center addNotificationRequest:request withCompletionHandler:^(NSError *error) {
                                            @autoreleasepool { vellumServiceResult(token,"",error?2:0); }
                                        }];
                                    } @catch(NSException *exception) { vellumServiceResult(token,"",2); } }
                                });
                            } @catch(NSException *exception) { vellumServiceResult(token,"",2); } }
                        }];break;
                    }
                    case 4: {
                        NSMutableArray *ids=[NSMutableArray new];
                        for (NSString *identifier in root[@"IDs"]) [ids addObject:[notificationPrefix stringByAppendingString:identifier]];
                        [center removePendingNotificationRequestsWithIdentifiers:ids];
                        [center removeDeliveredNotificationsWithIdentifiers:ids];
                        vellumServiceResult(token,"",0);break;
                    }
                }
            } @catch(NSException *exception) { vellumServiceResult(token,"",2); } }
        });
    });
}

// Pure data-object fixtures; never access the notification center or preferences.
int vellumNotificationValidate(const char *json) {
    @autoreleasepool { @try { NSData *data=notificationData(json);return data && notificationParse(data)!=nil; } @catch(NSException *exception) { return 0; } }
}
int vellumNotificationContentFixture(const char *json) {
    @autoreleasepool { @try {
        NSDictionary *root=notificationParse(notificationData(json));if (!root || [root[@"Action"] intValue]!=3) return 0;
        NSDictionary *n=root[@"Notification"];UNNotificationRequest *r=notificationRequest(n);
        return [r.identifier isEqual:[notificationPrefix stringByAppendingString:n[@"ID"]]] && [r.content.title isEqual:n[@"Title"]] && [r.content.subtitle isEqual:n[@"Subtitle"]] && [r.content.body isEqual:n[@"Body"]] && (r.content.sound!=nil)==[n[@"Sound"] boolValue] && r.trigger==nil;
    } @catch(NSException *exception) { return 0; } }
}
const char *vellumNotificationSettingsFixture(int auth,int alert,int sound) {
    @autoreleasepool { return strdup(notificationSettingsWire(auth,alert,sound).UTF8String); }
}
