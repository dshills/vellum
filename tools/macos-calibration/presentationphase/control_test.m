#import "control.m"
#include <assert.h>
static void reset(void){memset(&pc,0,sizeof(pc));}
static void *thread(void *arg){uint64_t t=(uintptr_t)arg;for(int n=0;n<40;n++){uint64_t a=pc_acquire_begin(1,t,n,CACurrentMediaTime());pc_acquired(a,t,CACurrentMediaTime());pc_presented(a,CACurrentMediaTime());}return NULL;}
int main(void){@autoreleasepool {
 assert(pc_configure()==1&&pc_configure()==0);
 uint64_t a=pc_acquire_begin(1,7,9,CACurrentMediaTime());assert(a==1);pc_acquired(a,10,CACurrentMediaTime());pc_receipt(8,CACurrentMediaTime());assert(pc.event[2].pending_before==1&&pc.event[2].before[0]==1);pc_presented(a,CACurrentMediaTime());assert(pc.event[3].pending_before==1&&pc.event[3].pending_after==0);pc_presented(a,0);assert(pc.invalid&&pc.count==4);reset();
 assert(pc_configure());a=pc_acquire_begin(1,1,1,CACurrentMediaTime());pc_unavailable(a,CACurrentMediaTime());assert(pc.attempt[0].state==4&&pc.event[1].acquiring_after==0);reset();
 assert(pc_configure());double now=CACurrentMediaTime();pc_tick(now,now-.02,now,.02);pc_tick(now+.02,now,now+.02,.02);double target=pc_target(1);assert(fabs(target-(now+3.125*.02))<1e-8&&pc.request[0].anchor==1&&pc.request[0].phase==0);assert(pc_target(1)==0&&pc.invalid);assert(!pc_wait(NAN));reset();
 assert(pc_configure());for(int i=0;i<PC_TICKS;i++)pc_tick(10+i*.02,9+i*.02,9.02+i*.02,.02);pc_tick(1000,999,999.02,.02);assert(pc.overflow&&pc.ticks==PC_TICKS);reset();
 assert(pc_configure());for(int i=0;i<PC_EVENTS;i++)pc_receipt(i,CACurrentMediaTime());pc_receipt(1,CACurrentMediaTime());assert(pc.overflow&&pc.count==PC_EVENTS);reset();
 assert(pc_configure());for(int i=0;i<PC_LIVE+1;i++){uint64_t id=pc_acquire_begin(1,i,1,CACurrentMediaTime());pc_acquired(id,i,CACurrentMediaTime());}assert(pc.overflow&&pc.invalid);int prefix=pc.count;pc_receipt(99,CACurrentMediaTime());assert(pc.count==prefix);char *overflow=pc_snapshot();assert(overflow);NSData *overflowData=[NSData dataWithBytes:overflow length:strlen(overflow)];free(overflow);NSDictionary *failure=[NSJSONSerialization JSONObjectWithData:overflowData options:0 error:nil];assert([failure[@"overflow"] boolValue]);for(NSDictionary *event in failure[@"events"]){assert([event[@"pending_before"] count]<=PC_LIVE&&[event[@"pending_after"] count]<=PC_LIVE);}reset();
 assert(pc_configure());pthread_t ts[4];for(uintptr_t i=0;i<4;i++)assert(!pthread_create(&ts[i],NULL,thread,(void*)(i+1)));for(int i=0;i<4;i++)assert(!pthread_join(ts[i],NULL));assert(!pc.invalid&&!pc.overflow&&pc.attempts==160&&pc.count==480);
 char *raw=pc_snapshot();assert(raw&&strlen(raw)<(4u<<20));NSData *data=[NSData dataWithBytes:raw length:strlen(raw)];free(raw);NSDictionary *d=[NSJSONSerialization JSONObjectWithData:data options:0 error:nil];assert([d[@"pending"] intValue]==0&&[d[@"acquiring"] intValue]==0);for(NSString *k in @[@"configured",@"invalid",@"overflow"])assert(CFGetTypeID((CFTypeRef)d[k])==CFBooleanGetTypeID());
 puts("PASS scalar census/replay, unavailable/duplicate lifecycle, phase anchor/order, tick/event bounds/no eviction, concurrent transitions and native JSON booleans; no application/window");
}return 0;}
