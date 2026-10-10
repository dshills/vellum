#include <stdint.h>
#include "../native/accesskit-c/include/accesskit.h"
accesskit_macos_subclassing_adapter *vellum_attach(void *view, void *window, uintptr_t handle);
void vellum_update(accesskit_macos_subclassing_adapter *adapter, uintptr_t handle);
const char *vellum_action_value(accesskit_action_request *request);

uint64_t vellum_resident_bytes(void);
