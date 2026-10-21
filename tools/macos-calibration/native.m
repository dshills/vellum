// Copyright (c) 2026 Davin Hills. SPDX-License-Identifier: MIT.
#import <Cocoa/Cocoa.h>
#import <Metal/Metal.h>
#import <QuartzCore/CAMetalLayer.h>
#include <stdatomic.h>
#include <stdio.h>
#include <unistd.h>
#include "observe.h"
typedef struct{_Atomic(double) receipt,acquired;_Atomic(int) accepted;} Entry;
static Entry records[512];static _Atomic(double) shown[512];static _Atomic(int) pending;
static void refuse(const char *why){
 for(int i=0;i<512;i++)if(atomic_load(&records[i].receipt)>0)printf("{\"failure_prefix_index\":%d,\"receipt\":%.9f,\"acquired\":%.9f,\"shown\":%.9f,\"accepted\":%d}\n",i,atomic_load(&records[i].receipt),atomic_load(&records[i].acquired),atomic_load(&shown[i]),atomic_load(&records[i].accepted));
 fflush(stdout);fprintf(stderr,"%s\n",why);fflush(stderr);_exit(2);
}
@interface Reference : NSView
@property(strong) CAMetalLayer *metal;
@property(strong) id<MTLCommandQueue> queue;
@property(strong) NSArray<id<MTLTexture>> *textures;
@property(strong) dispatch_group_t callbacks;
@property int sequence;
- (void)render:(int)index;
@end
@implementation Reference
- (BOOL)acceptsFirstResponder{return YES;}
- (void)keyDown:(NSEvent*)event {
 if(event.keyCode!=49 || ![event.characters isEqualToString:@" "]){refuse("unexpected key");}
 int i=atomic_load(&pending);if(i<=0 || atomic_load(&records[i].accepted)!=0){refuse("duplicate or unrelated event");}
 atomic_store(&records[i].accepted,++_sequence);[self render:i];
}
- (void)render:(int)i {
 id<CAMetalDrawable>d=[_metal nextDrawable];atomic_store(&records[i].acquired,cal_clock());if(!d){refuse("no drawable");}
 dispatch_group_enter(_callbacks);[d addPresentedHandler:^(id<MTLDrawable>x){atomic_store(&shown[i],x.presentedTime);dispatch_group_leave(self.callbacks);}];
 id<MTLCommandBuffer>cmd=[_queue commandBuffer];id<MTLBlitCommandEncoder>b=[cmd blitCommandEncoder];
 [b copyFromTexture:_textures[_sequence%2] sourceSlice:0 sourceLevel:0 sourceOrigin:MTLOriginMake(0,0,0) sourceSize:MTLSizeMake(d.texture.width,d.texture.height,1) toTexture:d.texture destinationSlice:0 destinationLevel:0 destinationOrigin:MTLOriginMake(0,0,0)];[b endEncoding];[cmd presentDrawable:d];[cmd commit];
}
@end
static void printJSON(NSDictionary *d){NSData *b=[NSJSONSerialization dataWithJSONObject:d options:NSJSONWritingSortedKeys error:nil];fwrite(b.bytes,1,b.length,stdout);puts("");fflush(stdout);}
int main(int argc,const char **argv){@autoreleasepool{
 if(argc!=2){fprintf(stderr,"MATCH.json\n");return 2;}
 NSData *data=[NSData dataWithContentsOfFile:@(argv[1])];if(data.length>16384 || !data){fprintf(stderr,"match file\n");return 2;}
 NSDictionary *match=[NSJSONSerialization JSONObjectWithData:data options:0 error:nil];int width=[match[@"width"] intValue],height=[match[@"height"] intValue];double scale=[match[@"scale"] doubleValue];NSArray *r=match[@"rect"];
 if(width<1 || height<1 || width>4096 || height>4096 || scale!=1 || r.count!=4){fprintf(stderr,"bounded match geometry\n");return 2;}
 int x=[r[0] intValue],y=[r[1] intValue],right=[r[2] intValue],bottom=[r[3] intValue];int format=[match[@"pixel_format"] intValue];
 if(x<0 || y<0 || right>width || bottom>height || right<=x || bottom<=y || (format!=MTLPixelFormatBGRA8Unorm && format!=MTLPixelFormatBGRA8Unorm_sRGB)){fprintf(stderr,"bounded format/rect\n");return 2;}
 NSApplication *app=NSApplication.sharedApplication;[app setActivationPolicy:NSApplicationActivationPolicyRegular];[app finishLaunching];
 NSWindow *window=[[NSWindow alloc]initWithContentRect:NSMakeRect(50,50,width,height) styleMask:NSWindowStyleMaskTitled|NSWindowStyleMaskClosable backing:NSBackingStoreBuffered defer:NO];window.releasedWhenClosed=NO;window.title=@"Vellum matched native calibration";
 Reference *view=[[Reference alloc]initWithFrame:NSMakeRect(0,0,width,height)];window.contentView=view;view.callbacks=dispatch_group_create();
 id<MTLDevice>device=MTLCreateSystemDefaultDevice();view.queue=[device newCommandQueue];CAMetalLayer *layer=CAMetalLayer.layer;view.metal=layer;layer.device=device;layer.pixelFormat=format;layer.drawableSize=CGSizeMake(width,height);layer.maximumDrawableCount=2;layer.displaySyncEnabled=YES;layer.presentsWithTransaction=NO;layer.allowsNextDrawableTimeout=NO;layer.framebufferOnly=NO;view.wantsLayer=YES;view.layer=layer;
 NSMutableArray *textures=[NSMutableArray new];MTLTextureDescriptor *desc=[MTLTextureDescriptor texture2DDescriptorWithPixelFormat:format width:width height:height mipmapped:NO];desc.storageMode=MTLStorageModeShared;
 for(int parity=0;parity<2;parity++){
  size_t length=(size_t)width*height*4;unsigned char *bytes=malloc(length);if(!bytes)return 2;memset(bytes,255,length);
  for(int row=0;row<height;row++)for(int col=0;col<width;col++){size_t at=((size_t)row*width+col)*4;bytes[at+3]=255;if(row>=y && row<bottom && col>=x && col<right){bytes[at]=0;bytes[at+1]=parity?0:255;bytes[at+2]=parity?255:0;}}
  id<MTLTexture>texture=[device newTextureWithDescriptor:desc];if(!texture){free(bytes);return 2;}[texture replaceRegion:MTLRegionMake2D(0,0,width,height) mipmapLevel:0 withBytes:bytes bytesPerRow:width*4];free(bytes);[textures addObject:texture];
 }view.textures=textures;
 [window makeKeyAndOrderFront:nil];[window makeFirstResponder:view];[app activateIgnoringOtherApps:YES];
 dispatch_async(dispatch_get_global_queue(QOS_CLASS_USER_INTERACTIVE,0),^{@autoreleasepool{
  NSDictionary *begin=nil;double readyDeadline=cal_clock()+2;
  do{char *raw=cal_observe(0);begin=[NSJSONSerialization JSONObjectWithData:[NSData dataWithBytes:raw length:strlen(raw)] options:0 error:nil];free(raw);if([begin[@"active"] boolValue] && [begin[@"key"] boolValue] && [begin[@"pid"] isEqual:begin[@"front_pid"]])break;usleep(1000);}while(cal_clock()<readyDeadline);
  char *raw=cal_observe(1);begin=[NSJSONSerialization JSONObjectWithData:[NSData dataWithBytes:raw length:strlen(raw)] options:0 error:nil];free(raw);
  if(![begin[@"active"] boolValue] || ![begin[@"key"] boolValue] || ![begin[@"visible"] boolValue] || ![begin[@"pid"] isEqual:begin[@"front_pid"]] || [begin[@"losses"] intValue])refuse("foreground readiness refused before startup");
  int setup=0;BOOL ready=NO;
  for(;setup<64;setup++){atomic_store(&records[setup].receipt,cal_clock());dispatch_sync(dispatch_get_main_queue(),^{[view render:setup];});if(dispatch_group_wait(view.callbacks,dispatch_time(DISPATCH_TIME_NOW,NSEC_PER_SEC))){refuse("setup callback timeout");}if(atomic_load(&shown[setup])>0){setup++;ready=YES;break;}}
  if(!ready){refuse("no positive ready frame");}
  double base=cal_clock()+.12;NSMutableArray *lateness=[NSMutableArray new];
  for(int n=1;n<=248;n++){@autoreleasepool{
   double deadline=base+(n-1)*.12;while(cal_clock()<deadline){double left=deadline-cal_clock();if(left>0)usleep((useconds_t)fmin(left*1e6,1000));}
   char *stateRaw=cal_observe(0);NSDictionary *state=[NSJSONSerialization JSONObjectWithData:[NSData dataWithBytes:stateRaw length:strlen(stateRaw)] options:0 error:nil];free(stateRaw);
   if(![state[@"active"] boolValue] || ![state[@"key"] boolValue] || ![state[@"visible"] boolValue] || ![state[@"pid"] isEqual:state[@"front_pid"]] || [state[@"losses"] intValue])refuse("foreground changed during cohort");
   double late=cal_clock()-deadline;[lateness addObject:@(late*1e9)];if(late>.020){refuse("request deadline lateness");}
   int i=setup+n-1;atomic_store(&pending,i);NSEvent *event=[NSEvent keyEventWithType:NSEventTypeKeyDown location:NSZeroPoint modifierFlags:0 timestamp:cal_clock() windowNumber:window.windowNumber context:nil characters:@" " charactersIgnoringModifiers:@" " isARepeat:NO keyCode:49];atomic_store(&records[i].receipt,cal_clock());[NSApp postEvent:event atStart:NO];
   while(atomic_load(&shown[i])<=0 && cal_clock()<deadline+.12)usleep(1000);
   if(atomic_load(&shown[i])<=0 || cal_clock()>=deadline+.12){refuse("response missed next deadline");}
  }}
  if(dispatch_group_wait(view.callbacks,dispatch_time(DISPATCH_TIME_NOW,NSEC_PER_SEC))){refuse("callback drain failed");}
  raw=cal_observe(2);NSDictionary *end=[NSJSONSerialization JSONObjectWithData:[NSData dataWithBytes:raw length:strlen(raw)] options:0 error:nil];free(raw);NSMutableArray *rows=[NSMutableArray new];
  for(int i=0;i<setup+248;i++){double receipt=atomic_load(&records[i].receipt),acquired=atomic_load(&records[i].acquired);int accepted=atomic_load(&records[i].accepted);[rows addObject:@{@"sequence":@(i<setup?0:i-setup+1),@"phase":@(i<setup?0:(i-setup<128?1:2)),@"receipt":@(receipt),@"acquired":@(acquired),@"shown":@(atomic_load(&shown[i])),@"accepted":@(accepted)}];}
  printJSON(@{@"kind":@"native",@"match":match,@"begin":begin,@"end":end,@"records":rows,@"lateness_ns":lateness,@"release_qualified":@NO});
  dispatch_async(dispatch_get_main_queue(),^{[window close];[NSApp stop:nil];[NSApp postEvent:[NSEvent otherEventWithType:NSEventTypeApplicationDefined location:NSZeroPoint modifierFlags:0 timestamp:0 windowNumber:0 context:nil subtype:0 data1:0 data2:0] atStart:NO];});
 }});[app run];puts("{\"runtime_returned\":true}");
}return 0;}
