// Copyright (c) 2026 Davin Hills. SPDX-License-Identifier: MIT.
#import "ready.h"
#include <stdio.h>
int main(void){@autoreleasepool{
 NSMutableDictionary *d=[@{@"active":@YES,@"key":@YES,@"visible":@NO,@"pid":@123,@"front_pid":@123,@"losses":@0} mutableCopy];
 // The original three-field predicate passes while occlusion visibility is false.
 if(!([d[@"active"] boolValue] && [d[@"key"] boolValue] && [d[@"pid"] isEqual:d[@"front_pid"]]) || cal_ready(d))return 1;
 d[@"visible"]=@YES;if(!cal_ready(d))return 2;
 for(NSString *flag in @[@"active",@"key",@"visible"]){d[flag]=@NO;if(cal_ready(d))return 3;d[flag]=@YES;}
 d[@"front_pid"]=@456;if(cal_ready(d))return 4;d[@"front_pid"]=@123;
 d[@"losses"]=@1;if(cal_ready(d))return 5;
 puts("PASS: complete foreground readiness; original predicate accepts invisible state; loss/key/active/visibility/PID negatives refused");
}return 0;}
