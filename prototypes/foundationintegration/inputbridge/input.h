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
// Read-only deterministic NSTextInputClient query observations (UTF-16 ranges).
char *vellum_input_query_probe(void *input, int start, int length, double x, double y);

typedef struct { int start,end,line; double x,y,width,height; } VellumTextCell;
typedef struct { int offset,line; double x,y,height; } VellumTextCaret;
void vellum_input_geometry_update(void *input, const char *text, int anchor, int focus, int marked_start, int marked_end, bool active, uint64_t acknowledged, double x, double y, double width, double height, const VellumTextCell *cells, int cell_count, const VellumTextCaret *carets, int caret_count, double vx, double vy, double vw, double vh, int focus_line);

typedef struct {double x,y,width,height;} VellumScreenRect;
// Test-only coordinate oracle: framebuffer pixels -> actual native-view points.
VellumScreenRect vellum_input_rendered_rect(void *input,double x,double y,double width,double height,int pixel_width,int pixel_height);

bool vellum_input_client_geometry_update(void *input,uint64_t clientID,const char *text,int anchor,int focus,int markedStart,int markedEnd,bool active,uint64_t acknowledged,double x,double y,double width,double height,const VellumTextCell *cells,int nc,const VellumTextCaret *carets,int nk,double vx,double vy,double vw,double vh,int focusLine);
bool vellum_input_client_update(void *input,uint64_t clientID,const char *text,int anchor,int focus,int markedStart,int markedEnd,bool active,uint64_t acknowledged,double x,double y,double width,double height);

void vellum_input_enable_deltas(void *input);
