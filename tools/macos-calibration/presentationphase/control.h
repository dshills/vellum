#pragma once
#include <stdint.h>
int pc_configure(void);
void pc_tick(double observed,double timestamp,double target,double duration);
double pc_target(int sequence);
int pc_wait(double target);
void pc_receipt(uint64_t tag,double at);
uint64_t pc_acquire_begin(uint64_t scope,uint64_t tag,uint64_t generation,double at);
void pc_acquired(uint64_t attempt,uint64_t drawable,double at);
void pc_unavailable(uint64_t attempt,double at);
void pc_presented(uint64_t attempt,double shown);
char *pc_snapshot(void); /* malloc-owned; <=4MiB; caller frees after shutdown */
