//go:build darwin && !offscreen
// Diagnostic-only bounded scalar observations. No drawable/window references.
#import <Foundation/Foundation.h>
#import <QuartzCore/CAMediaTiming.h>
#include <pthread.h>
#include <math.h>
#include <unistd.h>
#include "control.h"
#define PC_EVENTS 4096
#define PC_TICKS 8192
#define PC_ATTEMPTS 512
#define PC_LIVE 8
struct PCTick {double observed,timestamp,target,duration;};
struct PCAttempt {uint64_t scope,tag,generation,drawable;int state;};
struct PCEvent {int kind,request,pending_before,pending_after,acquiring_before,acquiring_after;uint64_t attempt,tag;double at,observed,shown;uint64_t before[PC_LIVE],after[PC_LIVE];};
struct PCRequest {int sequence,anchor,phase;double target;};
static struct {
 int configured,invalid,overflow,count,ticks,attempts,requests,current,base;
 struct PCTick tick[PC_TICKS];struct PCAttempt attempt[PC_ATTEMPTS];
 struct PCEvent event[PC_EVENTS];struct PCRequest request[248];
} pc;
static pthread_mutex_t pc_lock=PTHREAD_MUTEX_INITIALIZER;
static int pc_census(uint64_t ids[PC_LIVE],int *acquiring){
 int pending=0;*acquiring=0;
 for(int i=0;i<pc.attempts;i++){
  int state=pc.attempt[i].state;
  if(state==1)(*acquiring)++;
  if(state==2){if(pending<PC_LIVE)ids[pending]=i+1;pending++;}
 }
 if(pending>PC_LIVE || *acquiring>PC_LIVE){pc.overflow=1;pc.invalid=1;}
 return pending;
}
static struct PCEvent *pc_begin(int kind,uint64_t attempt,uint64_t tag,double at){
 if(!pc.configured || pc.overflow || pc.count==PC_EVENTS){pc.overflow=1;return NULL;}
 if(!isfinite(at)||at<=0){pc.invalid=1;return NULL;}
 struct PCEvent *e=&pc.event[pc.count++];e->kind=kind;e->attempt=attempt;e->tag=tag;e->at=at;e->request=pc.current;
 e->pending_before=pc_census(e->before,&e->acquiring_before);return e;
}
static void pc_end(struct PCEvent *e){if(e){e->pending_after=pc_census(e->after,&e->acquiring_after);e->observed=CACurrentMediaTime();}}
int pc_configure(void){pthread_mutex_lock(&pc_lock);int ok=!pc.configured;if(ok)pc.configured=1;pthread_mutex_unlock(&pc_lock);return ok;}
void pc_tick(double observed,double timestamp,double target,double duration){
 pthread_mutex_lock(&pc_lock);
 if(!pc.configured || pc.ticks==PC_TICKS){pc.overflow=1;}
 else if(!isfinite(observed)||!isfinite(timestamp)||!isfinite(target)||!isfinite(duration)||duration<=0||target<=timestamp||(pc.ticks && (timestamp<=pc.tick[pc.ticks-1].timestamp || observed<pc.tick[pc.ticks-1].observed))){pc.invalid=1;}
 else pc.tick[pc.ticks++]=(struct PCTick){observed,timestamp,target,duration};
 pthread_mutex_unlock(&pc_lock);
}
double pc_target(int sequence){
 if(sequence<1||sequence>248)return 0;
 double deadline=CACurrentMediaTime()+1;
 while(CACurrentMediaTime()<deadline){
  pthread_mutex_lock(&pc_lock);
  if(pc.invalid||pc.overflow||!pc.configured||sequence!=pc.requests+1){pc.invalid=1;pthread_mutex_unlock(&pc_lock);return 0;}
  if(pc.ticks>=2){
   if(sequence==1)pc.base=pc.ticks-1;
   int anchor=pc.base+(sequence-1)*8;
   if(anchor<pc.ticks){
    struct PCTick tick=pc.tick[anchor];int phase=(sequence-1)%4;
    double target=tick.timestamp+(3+(phase+.5)/4)*tick.duration;
    pc.request[pc.requests++]=(struct PCRequest){sequence,anchor,phase,target};pc.current=sequence;
    pthread_mutex_unlock(&pc_lock);return target;
   }
  }
  pthread_mutex_unlock(&pc_lock);usleep(250);
 }
 return 0;
}
int pc_wait(double target){
 if(!isfinite(target)||target<=0)return 0;
 double now=CACurrentMediaTime();if(target-now>1||now-target>.002)return 0;
 while((now=CACurrentMediaTime())<target){double remaining=target-now;usleep((useconds_t)fmax(1,fmin(remaining*1e6,250)));}
 return CACurrentMediaTime()-target<=.002;
}
void pc_receipt(uint64_t tag,double at){pthread_mutex_lock(&pc_lock);struct PCEvent *e=pc_begin(1,0,tag,at);pc_end(e);pthread_mutex_unlock(&pc_lock);}
uint64_t pc_acquire_begin(uint64_t scope,uint64_t tag,uint64_t generation,double at){
 pthread_mutex_lock(&pc_lock);uint64_t id=0;
 if(pc.attempts==PC_ATTEMPTS)pc.overflow=1;
 else {id=++pc.attempts;struct PCEvent *e=pc_begin(2,id,tag,at);pc.attempt[id-1]=(struct PCAttempt){scope,tag,generation,0,1};pc_end(e);}
 pthread_mutex_unlock(&pc_lock);return id;
}
static void pc_transition(uint64_t id,int kind,int before,int after,uint64_t drawable,double at,double shown){
 pthread_mutex_lock(&pc_lock);
 if(!id||id>(uint64_t)pc.attempts||pc.attempt[id-1].state!=before){pc.invalid=1;}
 else {struct PCEvent *e=pc_begin(kind,id,pc.attempt[id-1].tag,at);pc.attempt[id-1].state=after;if(drawable)pc.attempt[id-1].drawable=drawable;if(e)e->shown=shown;pc_end(e);}
 pthread_mutex_unlock(&pc_lock);
}
void pc_acquired(uint64_t id,uint64_t drawable,double at){pc_transition(id,3,1,2,drawable,at,0);}
void pc_unavailable(uint64_t id,double at){pc_transition(id,4,1,4,0,at,0);}
void pc_presented(uint64_t id,double shown){if(!isfinite(shown)||shown<0){pthread_mutex_lock(&pc_lock);pc.invalid=1;pthread_mutex_unlock(&pc_lock);shown=0;}pc_transition(id,5,2,3,0,CACurrentMediaTime(),shown);}
char *pc_snapshot(void){@autoreleasepool {
 pthread_mutex_lock(&pc_lock);NSMutableArray *events=[NSMutableArray array],*attempts=[NSMutableArray array],*requests=[NSMutableArray array];
 for(int i=0;i<pc.count;i++){struct PCEvent e=pc.event[i];NSMutableArray *before=[NSMutableArray array],*after=[NSMutableArray array];for(int j=0;j<e.pending_before&&j<PC_LIVE;j++)[before addObject:@(e.before[j])];for(int j=0;j<e.pending_after&&j<PC_LIVE;j++)[after addObject:@(e.after[j])];
  [events addObject:@{@"ordinal":@(i),@"kind":@(e.kind),@"request":@(e.request),@"attempt":@(e.attempt),@"tag":@(e.tag),@"at":@(e.at),@"observed":@(e.observed),@"shown":@(e.shown),@"pending_before":before,@"pending_after":after,@"acquiring_before":@(e.acquiring_before),@"acquiring_after":@(e.acquiring_after)}];}
 for(int i=0;i<pc.attempts;i++){struct PCAttempt a=pc.attempt[i];[attempts addObject:@{@"attempt":@(i+1),@"scope":@(a.scope),@"tag":@(a.tag),@"generation":@(a.generation),@"drawable":@(a.drawable),@"state":@(a.state)}];}
 for(int i=0;i<pc.requests;i++){struct PCRequest a=pc.request[i];[requests addObject:@{@"sequence":@(a.sequence),@"anchor":@(a.anchor),@"phase":@(a.phase),@"target":@(a.target)}];}
 uint64_t ids[PC_LIVE]={0};int acquiring=0;int pending=pc_census(ids,&acquiring);
 NSDictionary *d=@{@"pending":@(pending),@"acquiring":@(acquiring),@"configured":pc.configured?@YES:@NO,@"invalid":pc.invalid?@YES:@NO,@"overflow":pc.overflow?@YES:@NO,@"events":events,@"attempts":attempts,@"requests":requests};NSData *data=[NSJSONSerialization dataWithJSONObject:d options:NSJSONWritingSortedKeys error:nil];char *out=(data&&data.length<(4u<<20))?strndup(data.bytes,data.length):NULL;pthread_mutex_unlock(&pc_lock);return out;
}}
