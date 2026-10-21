// Copyright (c) 2026 Davin Hills. SPDX-License-Identifier: MIT.
#import <Foundation/Foundation.h>
// Comparisons have C int type; direct @() boxing can encode 1 instead of true.
static inline NSNumber *cal_flag(BOOL value) { return value ? @YES : @NO; }
