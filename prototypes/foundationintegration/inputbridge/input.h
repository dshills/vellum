#include <stdint.h>
#include <stdbool.h>
#include <stdlib.h>
void *vellum_input_attach(void *view, uintptr_t handle);
void vellum_input_update(void *input, const char *text, int anchor, int focus, int marked_start, int marked_end, bool active, uint64_t acknowledged, double x, double y, double width, double height);
void vellum_input_close(void *input);
// Deterministic protocol calls for tests, never a substitute for a live IME.
char *vellum_input_probe(void *input, int step);
char *vellum_input_source(void);

void vellum_input_caret(void *input, double x, double y, double width, double height);
char *vellum_input_environment(void);

char *vellum_input_inactive_probe(void *input);
