// Copyright (c) 2026 Davin Hills. SPDX-License-Identifier: MIT.
#import <Foundation/Foundation.h>
static inline BOOL cal_ready(NSDictionary *state) {
 return [state[@"active"] boolValue] && [state[@"key"] boolValue] &&
        [state[@"visible"] boolValue] && [state[@"pid"] intValue]>0 &&
        [state[@"pid"] isEqual:state[@"front_pid"]] && [state[@"losses"] intValue]==0;
}
