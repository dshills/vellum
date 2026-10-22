import copy,unittest
import native_check as check

def data():
    match=dict(width=100,height=100,scale=1,pixel_format=81,drawables=2,vsync=True,transactions=False,timeout=False,rect=[5,5,95,95])
    def state(clock):return {**match,'clock':clock,'pid':42,'front_pid':42,'active':True,'key':True,'visible':True,'losses':0}
    setup=dict(Sequence=1,Scope=1,Tag=0,Generation=0,Armed=False,Acquired=.5,Presented=.6,State='presented')
    samples=[setup];inputs=[];measurements=[]
    for i,tag in enumerate([9,31,72],1):
        inputs.append(dict(Scope=1,Tag=tag,Received=i+1.0));samples.append(dict(Sequence=i+1,Scope=1,Tag=tag,Generation=i+2,Armed=True,Acquired=i+1.01,Presented=i+1.03,State='presented'));measurements.append(dict(Scope=1,AcceptedRevision=i,ViewRevisionBefore=i+1,Frame=dict(Tag=tag,Generation=i+2,Completed=True,Failure='')))
    before=dict(Failed=False,Overflow=False,Pending=0,Scopes=[dict(ID=1,Closed=False)],Samples=samples,Inputs=inputs);after=copy.deepcopy(before);after['Scopes'][0]['Closed']=True
    red=[65535,0,0,65535];green=[0,65535,0,65535];frames=[]
    for admitted,color in [(False,green),(True,red),(True,green),(True,red)]:
        row=dict(Index=0,Type='*pimage.Params',Decision='render-image',Scroll=False,Command=0,Mask=False,HasUniformRGBA=True,UniformRGBA=color,Rect=dict(Min=dict(X=5,Y=5),Max=dict(X=95,Y=95)))
        frames.append(dict(Complete=True,Cohort=admitted,OwnedScanner=True,Target=dict(Min=dict(X=0,Y=0),Max=dict(X=100,Y=100)),Items=[row]))
    return dict(kind='framework',diagnostic_mode='three-transition-raster-commands',runtime_returned=True,error='<nil>',release_qualified=False,match=match,begin=state(1),end=state(5),before_close=before,after_close=after,measurements=dict(Pending=False,Samples=measurements),lateness_ns=[20]*3,receipt_lateness_ns=[30]*3,request_schedule_phases=[dict(Sequence=i,WakeLatenessNS=1,ForegroundCheckNS=15,PostCheckLatenessNS=20) for i in range(1,4)],raster_commands=dict(Overflow=False,Invalid=False,CohortStarted=True,Frames=frames)),match
class NativeCommands(unittest.TestCase):
    def test_three_routes_sparse_tags_and_startup_extent(self):
        d,m=data();r=check.validate(d,m);self.assertFalse(r['latency_qualified']);self.assertEqual([r['opaque_tag'] for r in r['responses']],[9,31,72])
        d['raster_commands']['Frames'][0]['Target']['Max']=dict(X=320,Y=200);check.validate(d,m)
    def test_route_feedback_and_schedule_refusals(self):
        for change in [lambda d:d.update(error='failure'),lambda d:d['end'].update(losses=1),lambda d:d['measurements']['Samples'][0].update(AcceptedRevision=2),lambda d:d['before_close']['Inputs'].pop(),lambda d:d['after_close']['Scopes'][0].update(Closed=False),lambda d:d['before_close']['Samples'][-1].update(State='skipped'),lambda d:d['lateness_ns'].__setitem__(0,20_000_001),lambda d:d['request_schedule_phases'][0].update(ForegroundCheckNS=-1)]:
            d,m=data();change(d)
            with self.assertRaises(ValueError):check.validate(d,m)
    def test_capture_boundary_and_decision_refusals(self):
        for change in [lambda d:d['raster_commands'].update(Overflow=True),lambda d:d['raster_commands'].update(Invalid=True),lambda d:d['raster_commands'].update(CohortStarted=False),lambda d:d['raster_commands']['Frames'][1].update(Complete=False),lambda d:d['raster_commands']['Frames'][1]['Target']['Max'].update(X=101),lambda d:d['raster_commands']['Frames'][2].update(Cohort=False),lambda d:d['raster_commands']['Frames'][1]['Items'][0].update(Decision=''),lambda d:d['raster_commands']['Frames'][1]['Items'][0].update(Scroll=True),lambda d:d['raster_commands']['Frames'][1]['Items'][0].update(Index=1),lambda d:d['raster_commands']['Frames'][3]['Items'][0].update(UniformRGBA=[0,65535,0,65535])]:
            d,m=data();change(d)
            with self.assertRaises(ValueError):check.validate(d,m)
if __name__=='__main__':unittest.main()
