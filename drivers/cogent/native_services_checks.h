#ifndef VELLUM_NATIVE_SERVICES_CHECKS_H
#define VELLUM_NATIVE_SERVICES_CHECKS_H
// Only the explicitly opted-in native acceptance build defines these probes.
#ifdef VELLUM_NATIVE_CHECKS
#import <Cocoa/Cocoa.h>
NSDictionary *vellumCheckMenu(int operation, int index);
NSDictionary *vellumCheckTray(int operation, int index);
#endif
#endif
