// Copyright (c) 2026 Davin Hills. SPDX-License-Identifier: MIT.
#pragma once
int dc_configure(int enabled);
int dc_start(void);
int dc_stop(void);
char *dc_snapshot(void); // malloc-owned; caller frees; at most 2 MiB
