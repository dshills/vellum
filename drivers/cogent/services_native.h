#ifndef VELLUM_SERVICES_NATIVE_H
#define VELLUM_SERVICES_NATIVE_H
#define VELLUM_SERVICE_REQUEST_MAX (1u << 20)
#ifdef __OBJC__
#import <Foundation/Foundation.h>
#include <stdint.h>
BOOL vellumWithServiceJSON(uint64_t token, const char *json, void (^consume)(NSData *));
void vellumTrayShutdown(void);
#endif
#endif
