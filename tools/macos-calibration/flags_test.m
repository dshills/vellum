// Copyright (c) 2026 Davin Hills. SPDX-License-Identifier: MIT.
#import "flags.h"
#include <stdio.h>
int main(void) {@autoreleasepool {
 unsigned flags=4;
 NSDictionary *values=@{@"visible":cal_flag((flags&4)!=0),@"hidden":cal_flag((flags&2)!=0)};
 NSData *data=[NSJSONSerialization dataWithJSONObject:values options:NSJSONWritingSortedKeys error:nil];
 NSString *actual=[[NSString alloc]initWithData:data encoding:NSUTF8StringEncoding];
 if(![actual isEqualToString:@"{\"hidden\":false,\"visible\":true}"]){fprintf(stderr,"flags are not JSON booleans: %s\n",actual.UTF8String);return 1;}
 // Demonstrate that the original expression produces a numeric JSON value.
 NSData *old=[NSJSONSerialization dataWithJSONObject:@{@"visible":@((flags&4)!=0)} options:0 error:nil];
 NSString *prior=[[NSString alloc]initWithData:old encoding:NSUTF8StringEncoding];
 if(![prior isEqualToString:@"{\"visible\":1}"]){fprintf(stderr,"original-source control changed\n");return 1;}
 puts("PASS: actual Foundation serializer emits true/false; original comparison boxing emits1");
}return 0;}
