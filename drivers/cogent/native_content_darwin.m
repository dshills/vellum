//go:build darwin && cgo

#import <Cocoa/Cocoa.h>
#import <PDFKit/PDFKit.h>
#import <AVKit/AVKit.h>
#import <AVFoundation/AVFoundation.h>
#include <fcntl.h>
#include <sys/stat.h>
#include <unistd.h>
#include <stdatomic.h>

extern NSView *vellumAXNativeParent(uint64_t);
extern NSDictionary *vellumAXNativeDiagnostics(uint64_t);
extern NSArray *vellumAXNativeItems(uint64_t);
extern void vellumAXNativeExit(uint64_t,NSString *,BOOL);
extern void vellumServiceResult(uint64_t,const char *,int);

// Native objects are main-owned. A worker retains only an immutable load job,
// never a view, host, window or callback into a retired managed tree.
@interface VellumContentJob : NSObject { @public atomic_bool canceled; }
@property(copy) NSString *key;
@property(copy) NSDictionary *config;
@end
@implementation VellumContentJob
- (instancetype)init {self=[super init];if(self)atomic_init(&canceled,false);return self;}
@end

@protocol VellumContentController <NSObject>
@property(readonly) NSView *view;
@property(readonly) NSView *focusView;
@property(readonly) NSDictionary *inspection;
- (void)install:(id)document error:(NSString *)error;
- (void)stop;
- (BOOL)command:(NSString *)key;
- (void)update:(NSDictionary *)config;
@end

// Stack layout adapts without replacing the document, native controls or focus.
@interface VellumPDFBar : NSStackView
@end
@implementation VellumPDFBar
- (void)layout {
 BOOL compact=self.superview.bounds.size.width<480;
 NSUserInterfaceLayoutOrientation direction=compact?NSUserInterfaceLayoutOrientationVertical:NSUserInterfaceLayoutOrientationHorizontal;
 if(self.orientation!=direction){self.orientation=direction;self.alignment=compact?NSLayoutAttributeLeading:NSLayoutAttributeCenterY;}
 [super layout];
}
@end

@class VellumPDFController;
@interface VellumPDFFindDelegate : NSObject <PDFDocumentDelegate>
@property(weak) VellumPDFController *owner;
@property(strong) NSObject *token;
@end

@interface VellumPDFController : NSObject <VellumContentController,PDFViewDelegate,PDFDocumentDelegate,NSSearchFieldDelegate>
@property(strong) NSView *view;
@property(strong) PDFView *pdf;
@property(strong) PDFThumbnailView *thumbnails;
@property(strong) NSSearchField *search;
@property(strong) NSTextField *pageLabel;
@property(strong) NSTextField *status;
@property(strong) id observer;
@property BOOL stopped;
@property NSUInteger matches;
@property BOOL showThumbnails;
@property(strong) NSLayoutConstraint *pdfLeading;
@property(strong) VellumPDFBar *toolbar;
@property(strong) VellumPDFFindDelegate *findDelegate;
@property(strong) NSObject *findToken;
- (void)found:(PDFSelection *)selection token:(NSObject *)token;
@end

static NSButton *vellumContentButton(NSString *label,NSString *symbol,id target,SEL action){
 NSButton *b=[NSButton buttonWithTitle:label target:target action:action];b.bezelStyle=NSBezelStyleTexturedRounded;
 if(symbol){b.image=[NSImage imageWithSystemSymbolName:symbol accessibilityDescription:label];b.imagePosition=NSImageOnly;b.toolTip=label;}
 b.accessibilityLabel=label;return b;
}
@implementation VellumPDFController
- (instancetype)initWithConfig:(NSDictionary *)config {
 self=[super init];if(!self)return nil;
 _view=[[NSView alloc]initWithFrame:NSMakeRect(0,0,600,420)];
 _pdf=[[PDFView alloc]initWithFrame:NSZeroRect];_pdf.translatesAutoresizingMaskIntoConstraints=NO;_pdf.delegate=self;
 _pdf.autoScales=YES;_pdf.displayMode=kPDFDisplaySinglePageContinuous;_pdf.displayDirection=kPDFDisplayDirectionVertical;
 _thumbnails=[[PDFThumbnailView alloc]initWithFrame:NSZeroRect];_thumbnails.translatesAutoresizingMaskIntoConstraints=NO;_thumbnails.PDFView=_pdf;_thumbnails.thumbnailSize=NSMakeSize(72,96);
 _showThumbnails=[config[@"thumbnails"] boolValue];_thumbnails.hidden=!_showThumbnails;
 _search=[[NSSearchField alloc]initWithFrame:NSMakeRect(0,0,120,24)];_search.translatesAutoresizingMaskIntoConstraints=NO;_search.placeholderString=@"Find in PDF";_search.accessibilityLabel=@"Find in PDF";_search.delegate=self;_search.target=self;_search.action=@selector(find:);
 _pageLabel=[NSTextField labelWithString:@"—"];_pageLabel.accessibilityLabel=@"PDF page";
 NSStackView *actions=[NSStackView stackViewWithViews:@[
  vellumContentButton(@"Thumbnails",@"sidebar.left",self,@selector(toggleThumbnails:)),
  vellumContentButton(@"Previous page",@"chevron.up",self,@selector(previous:)),
  vellumContentButton(@"Next page",@"chevron.down",self,@selector(next:)),_pageLabel,
  vellumContentButton(@"Zoom out",@"minus.magnifyingglass",self,@selector(zoomOut:)),
  vellumContentButton(@"Zoom in",@"plus.magnifyingglass",self,@selector(zoomIn:)),
  vellumContentButton(@"Fit",nil,self,@selector(fit:))]];
 actions.orientation=NSUserInterfaceLayoutOrientationHorizontal;actions.spacing=4;
 [actions setContentHuggingPriority:NSLayoutPriorityDefaultHigh forOrientation:NSLayoutConstraintOrientationHorizontal];
 [_search.widthAnchor constraintGreaterThanOrEqualToConstant:120].active=YES;[_search.heightAnchor constraintEqualToConstant:24].active=YES;
 [_search setContentHuggingPriority:NSLayoutPriorityDefaultLow forOrientation:NSLayoutConstraintOrientationHorizontal];
 VellumPDFBar *toolbar=[VellumPDFBar new];self.toolbar=toolbar;[toolbar addArrangedSubview:actions];[toolbar addArrangedSubview:_search];
 toolbar.translatesAutoresizingMaskIntoConstraints=NO;toolbar.spacing=5;toolbar.orientation=NSUserInterfaceLayoutOrientationHorizontal;
 [_view addSubview:toolbar];[_view addSubview:_pdf];[_view addSubview:_thumbnails];
 _status=[NSTextField wrappingLabelWithString:@"Loading local PDF…"];_status.translatesAutoresizingMaskIntoConstraints=NO;[_view addSubview:_status];
 [NSLayoutConstraint activateConstraints:@[
  [toolbar.topAnchor constraintEqualToAnchor:_view.topAnchor constant:6],[toolbar.leadingAnchor constraintEqualToAnchor:_view.leadingAnchor constant:6],[toolbar.trailingAnchor constraintEqualToAnchor:_view.trailingAnchor constant:-6],
  [_pdf.topAnchor constraintEqualToAnchor:toolbar.bottomAnchor constant:6],[_pdf.bottomAnchor constraintEqualToAnchor:_view.bottomAnchor],[_pdf.trailingAnchor constraintEqualToAnchor:_view.trailingAnchor],
  [_thumbnails.leadingAnchor constraintEqualToAnchor:_view.leadingAnchor],[_thumbnails.topAnchor constraintEqualToAnchor:_pdf.topAnchor],[_thumbnails.bottomAnchor constraintEqualToAnchor:_view.bottomAnchor],[_thumbnails.widthAnchor constraintEqualToConstant:96],
  [_status.centerXAnchor constraintEqualToAnchor:_view.centerXAnchor],[_status.centerYAnchor constraintEqualToAnchor:_view.centerYAnchor],[_status.widthAnchor constraintLessThanOrEqualToAnchor:_view.widthAnchor constant:-24]]];
 [self updateThumbnailConstraint];
 __weak VellumPDFController *weak=self;
 _observer=[[NSNotificationCenter defaultCenter]addObserverForName:PDFViewPageChangedNotification object:_pdf queue:NSOperationQueue.mainQueue usingBlock:^(NSNotification *n){[weak updatePage];}];
 return self;
}
- (void)updateThumbnailConstraint {
 if(!self.pdfLeading){self.pdfLeading=[self.pdf.leadingAnchor constraintEqualToAnchor:self.view.leadingAnchor];self.pdfLeading.active=YES;}
 self.pdfLeading.constant=self.showThumbnails?96:0;
}
- (void)update:(NSDictionary *)config {}
- (NSView *)focusView{return self.pdf;}
- (NSDictionary *)inspection{return @{@"search_width":@(self.search.frame.size.width),@"compact_toolbar":(self.toolbar.orientation==NSUserInterfaceLayoutOrientationVertical?@YES:@NO),@"page":self.pageLabel.stringValue,@"scale":@(self.pdf.scaleFactor),@"class":@"PDFView",@"pages":@(self.pdf.document.pageCount),@"loaded":(self.pdf.document?@YES:@NO),@"status":self.status.stringValue?:@""};}
- (void)install:(id)document error:(NSString *)error {
 if(self.stopped)return;[self.pdf.document cancelFindString];self.pdf.document.delegate=nil;self.matches=0;self.findDelegate=nil;self.findToken=nil;self.pdf.document=document;
 self.status.stringValue=error?:@"";self.status.hidden=document!=nil;[self updatePage];
}
- (void)updatePage {PDFDocument *d=self.pdf.document;NSUInteger i=d?[d indexForPage:self.pdf.currentPage]:NSNotFound;self.pageLabel.stringValue=i!=NSNotFound?[NSString stringWithFormat:@"%lu / %lu",(unsigned long)i+1,(unsigned long)d.pageCount]:@"—";}
- (void)previous:(id)sender{[self.pdf goToPreviousPage:sender];}
- (void)next:(id)sender{[self.pdf goToNextPage:sender];}
- (void)zoomOut:(id)sender{[self.pdf zoomOut:sender];[self revealSelection];}
- (void)zoomIn:(id)sender{[self.pdf zoomIn:sender];[self revealSelection];}
- (void)fit:(id)sender{self.pdf.autoScales=YES;[self revealSelection];}
- (void)toggleThumbnails:(id)sender{self.showThumbnails=!self.showThumbnails;self.thumbnails.hidden=!self.showThumbnails;[self updateThumbnailConstraint];}
- (void)find:(id)sender {
 PDFDocument *d=self.pdf.document;[d cancelFindString];d.delegate=nil;self.matches=0;self.findToken=[NSObject new];self.findDelegate=nil;self.pdf.currentSelection=nil;
 NSString *text=self.search.stringValue;if(self.stopped||text.length==0||text.length>256)return;
 VellumPDFFindDelegate *delegate=[VellumPDFFindDelegate new];delegate.owner=self;delegate.token=self.findToken;self.findDelegate=delegate;d.delegate=delegate;
 [d beginFindString:text withOptions:NSCaseInsensitiveSearch];
}
- (void)revealSelection {
 PDFSelection *selection=self.pdf.currentSelection;PDFPage *page=selection.pages.firstObject;
 if(!page||page.document!=self.pdf.document)return;
 [self.pdf layoutDocumentView];[self.pdf goToRect:NSInsetRect([selection boundsForPage:page],-12,-18) onPage:page];
}
- (void)found:(PDFSelection *)selection token:(NSObject *)token {
 if(!NSThread.isMainThread||self.stopped||token!=self.findToken||selection.pages.firstObject.document!=self.pdf.document)return;
 if(++self.matches==1){[self.pdf setCurrentSelection:selection animate:NO];[self revealSelection];}
 if(self.matches>=200)[self.pdf.document cancelFindString];
}
- (void)controlTextDidChange:(NSNotification *)n {if(n.object==self.search)[self find:self.search];}
- (BOOL)command:(NSString *)key {
 if([key isEqual:@"f"]){[self.view.window makeFirstResponder:self.search];return YES;}
 if([key isEqual:@"c"]){NSResponder *r=self.view.window.firstResponder;if([r isKindOfClass:NSTextView.class])[(NSTextView *)r copy:self];else [self.pdf copy:self];return YES;}
 if([key isEqual:@"a"]){NSResponder *r=self.view.window.firstResponder;if([r isKindOfClass:NSTextView.class])[(NSTextView *)r selectAll:self];else [self.pdf selectAll:self];return YES;}return NO;
}
// Embedded documents cannot initiate external URL/file navigation or printing.
- (void)PDFViewWillClickOnLink:(PDFView *)sender withURL:(NSURL *)url {}
- (void)PDFViewPerformGoToPage:(PDFView *)sender { [self.view.window makeFirstResponder:self.pdf]; }
- (void)PDFViewPerformFind:(PDFView *)sender {[self.view.window makeFirstResponder:self.search];}
- (void)PDFViewPerformPrint:(PDFView *)sender {}
- (void)PDFViewOpenPDF:(PDFView *)sender forRemoteGoToAction:(PDFActionRemoteGoTo *)action {}
- (void)stop {
 if(self.stopped)return;self.stopped=YES;[self.pdf.document cancelFindString];self.pdf.document.delegate=nil;self.pdf.delegate=nil;
 self.findDelegate=nil;self.findToken=nil;self.pdf.document=nil;self.thumbnails.PDFView=nil;self.search.delegate=nil;self.search.target=nil;
 if(self.observer)[[NSNotificationCenter defaultCenter]removeObserver:self.observer];self.observer=nil;
}
- (void)dealloc {if(_observer)[[NSNotificationCenter defaultCenter]removeObserver:_observer];}
@end

@implementation VellumPDFFindDelegate
- (void)didMatchString:(PDFSelection *)selection {
 __weak VellumPDFController *owner=self.owner;NSObject *token=self.token;
 void (^deliver)(void)=^{[owner found:selection token:token];};
 if(NSThread.isMainThread)deliver();else dispatch_async(dispatch_get_main_queue(),deliver);
}
@end

@interface VellumMediaController : NSObject <VellumContentController>
@property(strong) NSView *view;
@property(strong) AVPlayerView *playerView;
@property(strong) AVPlayer *player;
@property(strong) NSTextField *status;
@property(strong) NSImageView *poster;
@property(strong) id timeObserver;
@property BOOL stopped;
@property BOOL muted;
@property BOOL started;
@property(copy) NSString *encodedPoster;
@end
@implementation VellumMediaController
- (instancetype)initWithConfig:(NSDictionary *)config {
 self=[super init];if(!self)return nil;_muted=[config[@"initial_muted"] boolValue];
 _view=[[NSView alloc]initWithFrame:NSMakeRect(0,0,600,400)];
 _playerView=[[AVPlayerView alloc]initWithFrame:_view.bounds];_playerView.autoresizingMask=NSViewWidthSizable|NSViewHeightSizable;
 _playerView.controlsStyle=AVPlayerViewControlsStyleInline;_playerView.videoGravity=AVLayerVideoGravityResizeAspect;
 _playerView.showsSharingServiceButton=NO;_playerView.showsFullScreenToggleButton=NO;_playerView.allowsPictureInPicturePlayback=NO;
 if(@available(macOS 13.0,*)){_playerView.allowsVideoFrameAnalysis=NO;_playerView.allowsMagnification=NO;}
 [_view addSubview:_playerView];
 [self update:config];
 _status=[NSTextField wrappingLabelWithString:@"Loading local media…"];_status.translatesAutoresizingMaskIntoConstraints=NO;[_view addSubview:_status];
 [NSLayoutConstraint activateConstraints:@[[_status.centerXAnchor constraintEqualToAnchor:_view.centerXAnchor],[_status.centerYAnchor constraintEqualToAnchor:_view.centerYAnchor],[_status.widthAnchor constraintLessThanOrEqualToAnchor:_view.widthAnchor constant:-24]]];return self;
}
- (void)update:(NSDictionary *)config {
 NSString *encoded=config[@"poster_png"]?:@"";if([self.encodedPoster isEqual:encoded])return;self.encodedPoster=encoded;
 [self.poster removeFromSuperview];self.poster=nil;
 NSData *data=encoded.length<=350000?[[NSData alloc]initWithBase64EncodedString:encoded options:0]:nil;
 if(data.length&&data.length<=262144){self.poster=[[NSImageView alloc]initWithFrame:self.playerView.contentOverlayView.bounds];self.poster.image=[[NSImage alloc]initWithData:data];self.poster.imageScaling=NSImageScaleProportionallyUpOrDown;self.poster.autoresizingMask=NSViewWidthSizable|NSViewHeightSizable;self.poster.hidden=self.started;[self.playerView.contentOverlayView addSubview:self.poster];}
}
- (NSView *)focusView{return self.playerView;}
- (NSDictionary *)inspection{return @{@"class":@"AVPlayerView",@"loaded":(self.player.currentItem&&self.player.currentItem.status==AVPlayerItemStatusReadyToPlay?@YES:@NO),@"rate":@(self.player.rate),@"muted":(self.player.muted?@YES:@NO),@"poster":(self.poster.image?@YES:@NO),@"status":self.status.stringValue?:@""};}
- (void)install:(id)document error:(NSString *)error {
 if(self.stopped)return;
 if(error||!document){self.status.stringValue=error?:@"Cannot load local media.";return;}
 AVPlayerItem *item=[AVPlayerItem playerItemWithAsset:document];self.player=[AVPlayer playerWithPlayerItem:item];self.player.muted=self.muted;self.playerView.player=self.player;
 [self.player pause];self.status.stringValue=@"";self.status.hidden=YES;
 __weak VellumMediaController *weak=self;
 self.timeObserver=[self.player addPeriodicTimeObserverForInterval:CMTimeMake(1,4) queue:dispatch_get_main_queue() usingBlock:^(CMTime time){VellumMediaController *live=weak;if(live&&!live.stopped&&live.player.rate>0){live.started=YES;live.poster.hidden=YES;}}];
 [item addObserver:self forKeyPath:@"status" options:NSKeyValueObservingOptionInitial|NSKeyValueObservingOptionNew context:NULL];
}
- (void)observeValueForKeyPath:(NSString *)key ofObject:(id)object change:(NSDictionary *)change context:(void *)context {
 if(![key isEqual:@"status"]){[super observeValueForKeyPath:key ofObject:object change:change context:context];return;}
 __weak VellumMediaController *weak=self;dispatch_async(dispatch_get_main_queue(),^{VellumMediaController *live=weak;if(!live||live.stopped||object!=live.player.currentItem)return;
  if(live.player.currentItem.status==AVPlayerItemStatusFailed){live.status.stringValue=@"This local media file cannot be played.";live.status.hidden=NO;}
 });
}
- (BOOL)command:(NSString *)key{return NO;}
- (void)stop {
 if(self.stopped)return;self.stopped=YES;[self.player pause];
 AVPlayerItem *item=self.player.currentItem;if(item)[item removeObserver:self forKeyPath:@"status"];
 [item cancelPendingSeeks];[item.asset cancelLoading];
 if(self.timeObserver)[self.player removeTimeObserver:self.timeObserver];self.timeObserver=nil;
 self.playerView.player=nil;[self.player replaceCurrentItemWithPlayerItem:nil];self.player=nil;self.poster.image=nil;
}
@end

static AVURLAsset *vellumContentReadMedia(VellumContentJob *job,NSString **error){
 const char *path=[job.config[@"path"] fileSystemRepresentation];int fd=open(path,O_RDONLY|O_NONBLOCK|O_CLOEXEC);struct stat st;
 if(fd<0){*error=@"Cannot open local media.";return nil;}
 BOOL admitted=fstat(fd,&st)==0&&S_ISREG(st.st_mode)&&st.st_size>0&&st.st_size<=1024ll*1024ll*1024ll;close(fd);
 if(!admitted){*error=@"Media must be a regular file of at most 1 GiB.";return nil;}
 if(atomic_load(&job->canceled))return nil;
 AVURLAsset *asset=[AVURLAsset URLAssetWithURL:[NSURL fileURLWithPath:job.config[@"path"]] options:@{AVURLAssetPreferPreciseDurationAndTimingKey:@NO}];
 dispatch_semaphore_t ready=dispatch_semaphore_create(0);
 [asset loadValuesAsynchronouslyForKeys:@[@"playable"] completionHandler:^{dispatch_semaphore_signal(ready);}];
 // Bound our own metadata wait; AVFoundation owns decoder/resource internals.
 long wait=dispatch_semaphore_wait(ready,dispatch_time(DISPATCH_TIME_NOW,5ll*NSEC_PER_SEC));
 if(wait||atomic_load(&job->canceled)){[asset cancelLoading];*error=@"Local media loading was canceled or timed out.";return nil;}
 NSError *reason=nil;if([asset statusOfValueForKey:@"playable" error:&reason]!=AVKeyValueStatusLoaded||!asset.playable){[asset cancelLoading];*error=@"This local media file cannot be played.";return nil;}
 return asset;
}

@interface VellumContentClip : NSView
@property BOOL coordinateFlipped;
@end
@implementation VellumContentClip
- (BOOL)isFlipped{return self.coordinateFlipped;}
@end
@interface VellumContentHost : NSObject
@property uint64_t token;
@property(copy) NSString *identifier;
@property(copy) NSDictionary *config;
@property(strong) VellumContentClip *clip;
@property(strong) id<VellumContentController> controller;
@property(strong) VellumContentJob *job;
@property BOOL pending;
@property BOOL focused;
@end
@implementation VellumContentHost @end
static NSMutableDictionary<NSString *,VellumContentHost *> *vellumContentHosts;
static NSUInteger vellumContentWorkers;
static id vellumContentKeyMonitor;
static BOOL vellumContentStopping;
static NSString *vellumContentKey(uint64_t token,NSString *identifier){return [NSString stringWithFormat:@"%llu:%@",(unsigned long long)token,identifier];}
static BOOL vellumContentContains(NSView *root,NSResponder *responder){return [responder isKindOfClass:NSView.class]&&[(NSView *)responder isDescendantOf:root];}
static BOOL vellumContentRecipe(NSDictionary *config){
 if(![config isKindOfClass:NSDictionary.class]||(![config[@"type"] isEqual:@"pdf"]&&![config[@"type"] isEqual:@"media"]))return NO;
 NSString *path=config[@"path"];return [path isKindOfClass:NSString.class]&&path.length>0&&[path lengthOfBytesUsingEncoding:NSUTF8StringEncoding]<=4096&&path.isAbsolutePath&&[path rangeOfString:[NSString stringWithCharacters:(unichar[]){0} length:1]].location==NSNotFound;
}
static PDFDocument *vellumContentReadPDF(VellumContentJob *job,NSString **error){
 const NSUInteger limit=64u*1024u*1024u;const char *path=[job.config[@"path"] fileSystemRepresentation];
 int fd=open(path,O_RDONLY|O_NONBLOCK|O_CLOEXEC);struct stat st;
 if(fd<0){*error=@"Cannot open this local PDF.";return nil;}
 if(fstat(fd,&st)||!S_ISREG(st.st_mode)||st.st_size<1||st.st_size>limit){close(fd);*error=@"PDF must be a regular file of at most 64 MiB.";return nil;}
 NSMutableData *data=[NSMutableData dataWithCapacity:(NSUInteger)st.st_size];char buffer[65536];
 while(!atomic_load(&job->canceled)){
  ssize_t n=read(fd,buffer,sizeof(buffer));if(n==0)break;
  if(n<0){close(fd);*error=@"Cannot read this PDF.";return nil;}
  if(data.length+(NSUInteger)n>limit){close(fd);*error=@"PDF exceeds 64 MiB.";return nil;}[data appendBytes:buffer length:(NSUInteger)n];
 }
 close(fd);if(atomic_load(&job->canceled))return nil;
 PDFDocument *d=[[PDFDocument alloc]initWithData:data];
 if(!d||d.isLocked||d.pageCount==0||d.pageCount>1024){*error=@"PDF is unreadable, locked, empty, or exceeds 1024 pages.";return nil;}return d;
}
static void vellumContentPump(void){
 if(vellumContentStopping||vellumContentWorkers>=2)return;
 for(NSString *key in vellumContentHosts){VellumContentHost *h=vellumContentHosts[key];if(!h.pending||h.job)continue;
  h.pending=NO;VellumContentJob *job=[VellumContentJob new];job.key=key;job.config=h.config;h.job=job;vellumContentWorkers++;
  dispatch_async(dispatch_get_global_queue(QOS_CLASS_USER_INITIATED,0),^{@autoreleasepool{
   NSString *error=nil;id document=[job.config[@"type"] isEqual:@"media"]?vellumContentReadMedia(job,&error):vellumContentReadPDF(job,&error);
   dispatch_async(dispatch_get_main_queue(),^{@autoreleasepool{
    vellumContentWorkers--;VellumContentHost *live=vellumContentHosts[job.key];
    if(live.job==job&&!atomic_load(&job->canceled)){live.job=nil;[live.controller install:document error:error];}
    vellumContentPump();
   }});
  }});vellumContentPump();break;
 }
}
static void vellumContentRemove(NSString *key){
 VellumContentHost *h=vellumContentHosts[key];if(!h)return;
 if(h.job)atomic_store(&h.job->canceled,true);h.job=nil;h.pending=NO;
 NSWindow *window=h.clip.window;if(vellumContentContains(h.clip,window.firstResponder))[window makeFirstResponder:h.clip.superview];
 [h.controller stop];[h.clip removeFromSuperview];h.controller=nil;h.clip=nil;[vellumContentHosts removeObjectForKey:key];
 if(vellumContentHosts.count==0&&vellumContentKeyMonitor){[NSEvent removeMonitor:vellumContentKeyMonitor];vellumContentKeyMonitor=nil;}
}
static void vellumContentMonitor(void){
 if(vellumContentKeyMonitor)return;
 vellumContentKeyMonitor=[NSEvent addLocalMonitorForEventsMatchingMask:NSEventMaskKeyDown handler:^NSEvent *(NSEvent *event){
  for(VellumContentHost *h in vellumContentHosts.allValues){NSWindow *w=h.clip.window;if(event.window!=w||!vellumContentContains(h.clip,w.firstResponder))continue;
   if(event.modifierFlags&NSEventModifierFlagCommand){NSString *key=event.charactersIgnoringModifiers.lowercaseString;if([h.controller command:key])return nil;
    NSDictionary *edits=@{@"c":@"copy:",@"x":@"cut:",@"v":@"paste:",@"a":@"selectAll:",@"z":@"undo:",@"y":@"redo:"};NSString *selector=edits[key];
    if(selector){NSResponder *r=w.firstResponder;if([r isKindOfClass:NSTextView.class]){
      if([key isEqual:@"z"]){if(event.modifierFlags&NSEventModifierFlagShift)[r.undoManager redo];else [r.undoManager undo];}
      else if([key isEqual:@"y"])[r.undoManager redo];else [r tryToPerform:NSSelectorFromString(selector) with:h.controller];
    }return nil;}}
   if(event.keyCode==48){NSView *current=[w.firstResponder isKindOfClass:NSView.class]?(NSView *)w.firstResponder:nil;
    if([current isKindOfClass:NSTextView.class]&&[(NSTextView *)current isFieldEditor]&&[[(NSTextView *)current delegate] isKindOfClass:NSView.class])current=(NSView *)[(NSTextView *)current delegate];
    BOOL reverse=(event.modifierFlags&NSEventModifierFlagShift)!=0;NSView *next=reverse?current.previousValidKeyView:current.nextValidKeyView;
    if(!next||![next isDescendantOf:h.clip]){vellumAXNativeExit(h.token,h.identifier,reverse);return nil;}
   }break;
  }return event;
 }];
}
static NSRect vellumContentRectValues(NSArray *r,CGFloat scale,CGFloat height,BOOL flipped){
 if(scale<=0)scale=1;
 NSRect result=NSMakeRect([r[0] doubleValue]/scale,[r[1] doubleValue]/scale,[r[2] doubleValue]/scale,[r[3] doubleValue]/scale);
 if(!flipped)result.origin.y=height-NSMaxY(result);return result;
}
static NSRect vellumContentRect(NSArray *r,NSView *parent){return vellumContentRectValues(r,parent.window.backingScaleFactor,NSHeight(parent.bounds),parent.isFlipped);}
static NSRect vellumContentLocalFrame(NSRect raw,NSRect clip){raw.origin.x-=clip.origin.x;raw.origin.y-=clip.origin.y;return raw;
}
void vellumNativeContentSync(uint64_t token){
 if(![NSThread isMainThread]||vellumContentStopping)return;
 NSView *parent=vellumAXNativeParent(token);if(!parent)return;
 if(!vellumContentHosts)vellumContentHosts=[NSMutableDictionary new];NSMutableSet *wanted=[NSMutableSet new];
 NSArray *items=vellumAXNativeItems(token);NSMutableSet *eligible=[NSMutableSet new];
 for(NSDictionary *item in items){if(!vellumContentRecipe(item[@"config"]))continue;NSRect raw=vellumContentRect(item[@"raw"],parent),clip=NSIntersectionRect(vellumContentRect(item[@"clip"],parent),parent.visibleRect);if(!NSIsEmptyRect(raw)&&!NSIsEmptyRect(clip))[eligible addObject:vellumContentKey(token,item[@"id"])];}
 for(NSString *key in vellumContentHosts.allKeys.copy)if(vellumContentHosts[key].token==token&&![eligible containsObject:key])vellumContentRemove(key);
 for(NSDictionary *item in items){
  NSDictionary *config=item[@"config"];if(!vellumContentRecipe(config))continue;
  NSRect raw=vellumContentRect(item[@"raw"],parent),clip=NSIntersectionRect(vellumContentRect(item[@"clip"],parent),parent.visibleRect);
  if(NSIsEmptyRect(clip)||NSIsEmptyRect(raw))continue;
  NSString *identifier=item[@"id"],*key=vellumContentKey(token,identifier);[wanted addObject:key];VellumContentHost *h=vellumContentHosts[key];
  if(h&&(![h.config[@"type"] isEqual:config[@"type"]]||![h.config[@"path"] isEqual:config[@"path"]]||![h.config[@"revision"] isEqual:config[@"revision"]])){vellumContentRemove(key);h=nil;}
  if(!h){if(vellumContentHosts.count>=8)continue;
   h=[VellumContentHost new];h.token=token;h.identifier=identifier;h.config=config;h.pending=YES;
   h.clip=[[VellumContentClip alloc]initWithFrame:clip];h.clip.coordinateFlipped=parent.isFlipped;h.clip.wantsLayer=YES;h.clip.layer.masksToBounds=YES;
   h.controller=[config[@"type"] isEqual:@"media"]?[[VellumMediaController alloc]initWithConfig:config]:[[VellumPDFController alloc]initWithConfig:config];[h.clip addSubview:h.controller.view];[parent addSubview:h.clip positioned:NSWindowAbove relativeTo:nil];vellumContentHosts[key]=h;vellumContentMonitor();
  }
  BOOL appearanceChanged=!h.clip.appearance||[h.config[@"dark"] boolValue]!=[config[@"dark"] boolValue];
  NSRect local=vellumContentLocalFrame(raw,clip);BOOL geometryChanged=!NSEqualRects(h.clip.frame,clip)||!NSEqualRects(h.controller.view.frame,local);
  [h.controller update:config];h.config=config;h.clip.frame=clip;h.controller.view.frame=local;
  if(appearanceChanged){h.clip.appearance=[NSAppearance appearanceNamed:[config[@"dark"] boolValue]?NSAppearanceNameDarkAqua:NSAppearanceNameAqua];h.controller.view.appearance=h.clip.appearance;
   [h.clip.appearance performAsCurrentDrawingAppearance:^{h.clip.layer.backgroundColor=NSColor.windowBackgroundColor.CGColor;}];}
  if(geometryChanged||appearanceChanged)[h.controller.view layoutSubtreeIfNeeded];
  BOOL focus=[item[@"focused"] boolValue];if(focus&&!h.focused)[parent.window makeFirstResponder:h.controller.focusView];h.focused=focus;
 }
 for(NSString *key in vellumContentHosts.allKeys.copy)if(vellumContentHosts[key].token==token&&![wanted containsObject:key])vellumContentRemove(key);
 vellumContentPump();
}
void vellumNativeContentCloseWindow(uint64_t token){for(NSString *key in vellumContentHosts.allKeys.copy)if(vellumContentHosts[key].token==token)vellumContentRemove(key);}
void vellumNativeContentShutdown(void){vellumContentStopping=YES;for(NSString *key in vellumContentHosts.allKeys.copy)vellumContentRemove(key);}
NSArray *vellumNativeContentChildren(uint64_t token,NSString *identifier){VellumContentHost *h=vellumContentHosts[vellumContentKey(token,identifier)];return h?@[h.controller.view]:nil;}
id vellumNativeContentFocused(uint64_t token){for(VellumContentHost *h in vellumContentHosts.allValues)if(h.token==token&&vellumContentContains(h.clip,h.clip.window.firstResponder))return h.clip.window.firstResponder;return nil;}
id vellumNativeContentHit(uint64_t token,NSPoint point){for(VellumContentHost *h in vellumContentHosts.allValues)if(h.token==token){NSPoint p=[h.clip convertPoint:[h.clip.window convertPointFromScreen:point] fromView:nil];if(NSPointInRect(p,h.clip.bounds))return [h.controller.view accessibilityHitTest:point];}return nil;}
void vellumNativeContentProbe(uint64_t parent,uint64_t request){dispatch_async(dispatch_get_main_queue(),^{@autoreleasepool{
 NSMutableArray *hosts=[NSMutableArray new];for(VellumContentHost *h in vellumContentHosts.allValues)if(h.token==parent){NSMutableDictionary *v=[h.controller.inspection mutableCopy];v[@"id"]=h.identifier;v[@"frame"]=NSStringFromRect(h.clip.frame);v[@"attached"]=(h.clip.window?@YES:@NO);v[@"opaque_background"]=(h.clip.layer.backgroundColor&&CGColorGetAlpha(h.clip.layer.backgroundColor)>=0.99?@YES:@NO);[hosts addObject:v];}
 NSData *data=[NSJSONSerialization dataWithJSONObject:@{@"hosts":hosts,@"workers":@(vellumContentWorkers),@"geometry":vellumAXNativeDiagnostics(parent)} options:0 error:nil];vellumServiceResult(request,[[[NSString alloc]initWithData:data encoding:NSUTF8StringEncoding] UTF8String],0);
 }});}
int vellumPDFDecodeFixture(const char *path,int want){@autoreleasepool{VellumContentJob *j=[VellumContentJob new];j.config=@{@"path":[NSString stringWithUTF8String:path]};NSString *error=nil;PDFDocument *d=vellumContentReadPDF(j,&error);return want>0?d.pageCount==want:(!d&&error!=nil);}}
int vellumNativeFrameFixture(void){@autoreleasepool{
 NSArray *raw=@[@20,@(-80),@600,@800],*clip=@[@20,@0,@600,@400];
 NSRect flipped=vellumContentLocalFrame(vellumContentRectValues(raw,2,500,YES),vellumContentRectValues(clip,2,500,YES));
 NSRect unflipped=vellumContentLocalFrame(vellumContentRectValues(raw,2,500,NO),vellumContentRectValues(clip,2,500,NO));
 NSRect one=vellumContentLocalFrame(vellumContentRectValues(raw,1,1000,YES),vellumContentRectValues(clip,1,1000,YES));
 return NSEqualRects(flipped,NSMakeRect(0,-40,300,400))&&NSEqualRects(unflipped,NSMakeRect(0,-160,300,400))&&NSEqualRects(one,NSMakeRect(0,-80,600,800));
}}

int vellumMediaDecodeFixture(const char *path,int want){@autoreleasepool{VellumContentJob *j=[VellumContentJob new];j.config=@{@"path":[NSString stringWithUTF8String:path]};NSString *error=nil;id asset=vellumContentReadMedia(j,&error);return want?(asset!=nil):(!asset&&error!=nil);}}
