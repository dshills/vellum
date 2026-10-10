#import <Foundation/Foundation.h>
void *vellum_check_pool_begin(void) { return [[NSAutoreleasePool alloc] init]; }
void vellum_check_pool_end(void *pool) { [(NSAutoreleasePool *)pool drain]; }
