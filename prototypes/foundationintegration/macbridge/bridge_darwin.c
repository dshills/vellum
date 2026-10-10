#include "bridge.h"
#include <objc/runtime.h>
extern accesskit_tree_update *goVellumSnapshot(uintptr_t handle);
extern void goVellumAction(uintptr_t handle, accesskit_action_request *request);
static accesskit_tree_update *snapshot(void *userdata) { return goVellumSnapshot((uintptr_t)userdata); }
static void action(accesskit_action_request *request, void *userdata) {
 goVellumAction((uintptr_t)userdata, request);
 accesskit_action_request_free(request);
}
accesskit_macos_subclassing_adapter *vellum_attach(void *view, void *window, uintptr_t handle) {
 // Installation is per class; different native window classes need their own forwarding.
 static Class installed[64];static size_t count=0;
 Class cls=object_getClass((id)window);bool found=false;
 for(size_t i=0;i<count;i++)if(installed[i]==cls){found=true;break;}
 if(!found){if(count==64)return NULL;accesskit_macos_add_focus_forwarder_to_window_class(class_getName(cls));installed[count++]=cls;}
 return accesskit_macos_subclassing_adapter_new(view,snapshot,(void *)handle,action,(void *)handle);
}
void vellum_update(accesskit_macos_subclassing_adapter *adapter, uintptr_t handle) {
 accesskit_macos_queued_events *events=accesskit_macos_subclassing_adapter_update_if_active(adapter,snapshot,(void *)handle);
 if(events)accesskit_macos_queued_events_raise(events);
}
const char *vellum_action_value(accesskit_action_request *r) {
 return r->data.has_value && r->data.value.tag==ACCESSKIT_ACTION_DATA_VALUE ? r->data.value.value : NULL;
}

#include <mach/mach.h>
uint64_t vellum_resident_bytes(void) {
 mach_task_basic_info_data_t info;mach_msg_type_number_t count=MACH_TASK_BASIC_INFO_COUNT;
 kern_return_t err=task_info(mach_task_self(),MACH_TASK_BASIC_INFO,(task_info_t)&info,&count);
 return err==KERN_SUCCESS ? info.resident_size : 0;
}

const accesskit_text_selection *vellum_action_selection(accesskit_action_request *request) {
 return request->data.has_value && request->data.value.tag == ACCESSKIT_ACTION_DATA_SET_TEXT_SELECTION ? &request->data.value.set_text_selection : NULL;
}
