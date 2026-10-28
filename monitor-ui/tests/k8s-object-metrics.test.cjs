const {test} = require('node:test')
const assert = require('node:assert/strict')
const fs = require('node:fs')
const path = require('node:path')
const vm = require('node:vm')
const root = path.resolve(__dirname, '..')

function loadHelpers() {
  const source = fs.readFileSync(path.join(root, 'src/views/admin-config/other/k8s-metrics.js'), 'utf8')
  const module = {exports:{}}
  vm.runInNewContext(source.replace(/export /g, '') + '\nmodule.exports={filterObjects,metricRange,chartSeries,formatMetricValue};', {module})
  return module.exports
}

test('object filters combine workload, pod and instanceId without mixing namespaces', () => {
  const {filterObjects} = loadHelpers()
  const rows = [{workloadName:'app',podName:'app-0',instanceId:'guid-a'}, {workloadName:'app',podName:'app-1',instanceId:'guid-b'}, {workloadName:'db',podName:'db-0',instanceId:'guid-a'}]
  assert.equal(filterObjects(rows, {workload:'app',pod:'app-',instanceId:'guid-a'}).length, 1)
  assert.equal(filterObjects(rows, {workload:'',pod:'',instanceId:''}).length, 3)
})

test('range resolution respects the configured scrape interval and caps chart points', () => {
  const {metricRange} = loadHelpers()
  const range = metricRange(86400, 60, 200000)
  assert.equal(range.end, 200000)
  assert.equal(range.start, 113600)
  assert.ok(range.step >= 60)
  assert.ok((range.end-range.start)/range.step <= 600)
  assert.equal(range.step % 60, 0)
})

test('chart data retains incarnation and label series and leaves missing samples as gaps', () => {
  const {chartSeries} = loadHelpers()
  const series = chartSeries({result:[
    {metric:{pod_uid:'old',pool:'heap'},values:[[0,'1'],[60,'NaN'],[120,'3']]},
    {metric:{pod_uid:'new',pool:'heap'},values:[[120,'4']]}
  ]}, {start:0,end:180,step:60})
  assert.equal(series.length,2)
  assert.notEqual(series[0].name,series[1].name)
  assert.deepEqual(JSON.parse(JSON.stringify(series[0].data)),[[0,1],[60000,null],[120000,3],[180000,null]])
  assert.equal(series[0].connectNulls,false)
  assert.equal(chartSeries({result:[]},{start:0,end:60,step:60}).length,0)
})

test('chart labels omit infrastructure fields without merging incarnations or losing full labels', () => {
  const {chartSeries} = loadHelpers()
  const labels = {cluster_id:'cluster-a',namespace:'ns-a',node:'node-a',instance:'agent.example.test:9105',instance_id:'instance-a',container:'main',source_id:'source-a',series_id:'series-a',pod_uid:'old',area:'heap',pool:'Young'}
  const series = chartSeries({result:[
    {metric:labels,values:[[0,'1']]},
    {metric:{...labels,pod_uid:'new'},values:[[0,'2']]},
    {metric:{...labels,pool:'Old'},values:[[0,'3']]}
  ]}, {start:0,end:0,step:60})
  assert.equal(series.length,3)
  assert.match(series[0].displayName,/area=heap/)
  assert.match(series[0].displayName,/pool=Young/)
  assert.doesNotMatch(series[0].displayName,/node|instance|source_id|series_id|pod_uid|container/)
  assert.notEqual(series[0].displayName,series[1].displayName)
  assert.notEqual(series[0].name,series[1].name)
  assert.equal(series[0].labels.pod_uid,'old')
  assert.equal(series[1].labels.pod_uid,'new')
  assert.equal(series[0].labels.source_id,'source-a')
  assert.equal(series[0].data[0][1],1)
  assert.equal(series[1].data[0][1],2)
})

test('metric formatting distinguishes cores, memory, throughput and durations', () => {
  const {formatMetricValue} = loadHelpers()
  assert.match(formatMetricValue(1048576,'bytes'), /1.*MiB/)
  assert.match(formatMetricValue(1024,'bytes/s'), /1.*KiB\/s/)
  assert.match(formatMetricValue(0.5,'cores'), /0.5.*cores/)
  assert.match(formatMetricValue(0.001,'seconds'), /1.*ms/)
  assert.match(formatMetricValue(0.5,'ratio'), /50.*%/)
  assert.equal(formatMetricValue(null,'count'), '-')
})

function loadComponent(file, globals = {}) {
  const source = fs.readFileSync(path.join(root, 'src/views/admin-config/other/' + file + '.vue'), 'utf8')
  const script = source.match(/<script>([\s\S]*?)<\/script>/)[1].replace(/^import .*$/gm,'').replace('export default','module.exports =')
  const stateSource = fs.readFileSync(path.join(root, 'src/views/admin-config/other/k8s-view-state.js'), 'utf8')
  const state = vm.runInNewContext(stateSource.replace(/export /g,'')+'\nmodule.exports={readViewState,writeViewState,resolveScope}',{module:{exports:{}}})
  return vm.runInNewContext(script,{module:{exports:{}},setTimeout:()=>0,clearTimeout(){},K8sLogSamples:{},...state,...globals})
}

test('object mode comes from catalog and preserves history controls and raw alert states', () => {
  const component = loadComponent('k8s-objects', {...loadHelpers(),K8sMetricChart:{}})
  let done, fail
  const selections = {resource:['cpu','memory']}
  const range = {start:0,end:60,step:30}
  const ctx = {catalog:[],metricSelections:selections,range,request(method,url,body,success,config,error){done=success;fail=error},apiCenter:{k8sMetricCatalog:'/catalog'},applyRange(){},startRefresh(){},$t:k=>k}
  component.methods.loadCatalog.call(ctx)
  done({metrics:[],scrapeIntervalSeconds:30,countWindowSeconds:300,localCollectionEnabled:false,alertEvaluationEnabled:false})
  assert.equal(ctx.localCollectionEnabled,false)
  assert.equal(ctx.alertEvaluationEnabled,false)
  assert.equal(ctx.metricSelections,selections)
  assert.equal(ctx.range,range)
  assert.equal(component.computed.alertColumns.call(ctx)[1].title,'m_k8s_last_evaluation_state')
  component.methods.loadCatalog.call(ctx)
  fail()
  assert.equal(ctx.localCollectionEnabled,null,'unavailable mode must not be inferred from lack of samples')
  assert.equal(ctx.alertEvaluationEnabled,null)
})

test('chart queries the selected logical object and ignores obsolete responses', () => {
  const helpers = loadHelpers()
  const component = loadComponent('k8s-metric-chart', helpers)
  const pending = []
  const ctx = {object:{clusterId:'cluster-a',namespace:'sit-a',workloadType:'StatefulSet',workloadName:'app',podName:'app-0',container:'main',podUid:'current'},metric:{id:'cpu',unit:'cores'},range:{start:0,end:120,step:60},series:[],request(method,url,body,done){assert.equal(body.podName,'app-0');assert.equal(body.container,'main');assert.equal(body.podUid,undefined);pending.push(done)},apiCenter:{k8sMetricHistory:'/history'},renderChart(){},$set(obj,key,value){obj[key]=value}}
  component.methods.load.call(ctx)
  component.methods.load.call(ctx)
  pending[1]({result:[{metric:{pod_uid:'new'},values:[[120,'2']]}]})
  pending[0]({result:[{metric:{pod_uid:'old'},values:[[120,'999']]}]})
  assert.equal(ctx.series[0].data[2][1],2)
  assert.equal(ctx.loading,false)
})

test('chart failure is separate from an empty successful result', () => {
  const component = loadComponent('k8s-metric-chart', loadHelpers())
  let fail
  const ctx = {object:{podName:'app-0'},metric:{id:'cpu'},range:{start:0,end:60,step:60},request(method,url,body,done,config,error){fail=error},apiCenter:{k8sMetricHistory:'/history'},renderChart(){}}
  component.methods.load.call(ctx)
  fail()
  assert.equal(ctx.failed,true)
  assert.equal(ctx.loading,false)
  assert.equal(ctx.series.length,0)
})

test('object browser ignores old namespace inventory and preserves logical selection after replacement', () => {
  const component = loadComponent('k8s-objects', {...loadHelpers(), K8sMetricChart:{}})
  const pending = []
  const ctx = {scope:{clusterId:'cluster-a',namespace:'sit-a'},rows:[],selected:null,filters:{workload:'',pod:'',instanceId:''},apiCenter:{k8sObjects:'/objects'},request(method,url,body,done){pending.push(done)}}
  component.methods.loadObjects.call(ctx,true)
  ctx.scope.namespace='sit-b'
  component.methods.loadObjects.call(ctx,true)
  pending[1]({objects:[{podName:'app-0',container:'main',podUid:'new'}],state:'ACTIVE'})
  pending[0]({objects:[{podName:'other',container:'main'}],state:'DEGRADED'})
  assert.equal(ctx.rows[0].podName,'app-0')
  assert.equal(ctx.state,'ACTIVE')
  ctx.selected=ctx.rows[0]
  component.methods.loadObjects.call(ctx)
  pending[2]({objects:[{podName:'app-0',container:'main',podUid:'replacement'}],state:'ACTIVE'})
  assert.equal(ctx.selected.podUid,'replacement')
  component.methods.loadObjects.call(ctx)
  pending[3]({objects:[],state:'ACTIVE'})
  assert.equal(ctx.selected,null)
})

test('automatic refresh stops when disabled or when the page is destroyed', () => {
  const callbacks=[],cleared=[]
  const component = loadComponent('k8s-objects', {...loadHelpers(),K8sMetricChart:{},document:{hidden:false},setInterval(callback,delay){callbacks.push({callback,delay});return callbacks.length},clearInterval(id){cleared.push(id)}})
  let refreshes=0
  const ctx = {autoRefresh:true,scrapeInterval:60,scope:{},filters:{},metricSelections:{},customRange:[],selected:null,refresh(){refreshes++}}
  component.methods.startRefresh.call(ctx)
  assert.equal(callbacks[0].delay,60000)
  callbacks[0].callback()
  assert.equal(refreshes,1)
  ctx.autoRefresh=false
  component.methods.startRefresh.call(ctx)
  assert.ok(cleared.includes(1))
  component.beforeDestroy.call(ctx)
})

function clock() {
  let now=0, next=1
  const tasks=new Map()
  const schedule=(callback,delay,interval=0)=>{const id=next++;tasks.set(id,{callback,time:now+delay,interval});return id}
  return {
    get now() {return now},
    setTimeout:(callback,delay)=>schedule(callback,delay),
    clearTimeout:id=>tasks.delete(id),
    setInterval:(callback,delay)=>schedule(callback,delay,delay),
    clearInterval:id=>tasks.delete(id),
    advance(end) {
      while (tasks.size) {
        const entry=[...tasks].filter(([,task])=>task.time<=end).sort((a,b)=>a[1].time-b[1].time || a[0]-b[0])[0]
        if (!entry) {break}
        const [id,task]=entry
        now=task.time
        if (task.interval) {task.time+=task.interval} else {tasks.delete(id)}
        task.callback()
      }
      now=end
    }
  }
}

test('periodic refresh lets slow inventory and moving-range chart requests finish', () => {
  const timer=clock(), helpers=loadHelpers(), requests={objects:0,history:0}
  const page=loadComponent('k8s-objects',{...helpers,...timer,K8sMetricChart:{},document:{hidden:false},metricRange:(duration,interval)=>helpers.metricRange(duration,interval,10000+timer.now/1000)})
  const chart=loadComponent('k8s-metric-chart',{...helpers,...timer})
  const object={clusterId:'cluster-a',namespace:'ns-a',workloadType:'StatefulSet',workloadName:'app',podName:'app-0',container:'main'}
  const parent={scope:{clusterId:'cluster-a',namespace:'ns-a'},rows:[],selected:object,filters:{workload:'',pod:'',instanceId:''},page:1,timePreset:'3600',scrapeInterval:15,autoRefresh:true,refreshToken:0,range:{start:6400,end:10000,step:15},apiCenter:{k8sObjects:'objects'},request(method,url,body,done){requests.objects++;timer.setTimeout(()=>done({objects:[object],state:'ACTIVE'}),20000)}}
  let renders=0
  const child={object,metric:{id:'cpu'},range:parent.range,refreshToken:0,apiCenter:{k8sMetricHistory:'history'},renderChart(){renders++},request(method,url,body,done){requests.history++;timer.setTimeout(()=>done({result:[{metric:{pod_uid:'current'},values:[[body.end,'2']]}]}),20000)}}
  child.load=chart.methods.load.bind(child)
  for (const [name,method] of Object.entries(page.methods)) {parent[name]=method.bind(parent)}
  const refresh=parent.refresh
  parent.refresh=automatic=>{
    refresh(automatic)
    Object.assign(child,{range:parent.range,refreshToken:parent.refreshToken,periodicRefresh:parent.periodicRefresh})
    chart.watch.queryKey.call(child)
  }
  parent.loadObjects()
  child.load()
  parent.startRefresh()
  timer.advance(20000)
  assert.equal(parent.rows.length,1,'the first inventory response must remain current after a periodic tick')
  assert.equal(parent.rows[0].podName,'app-0','the first inventory response must remain current after a periodic tick')
  assert.equal(parent.loading,false)
  assert.equal(renders,1,'a successful history response must not be discarded by periodic refresh')
  assert.equal(child.series[0].data.some(point=>point[1]===2),true)
  assert.equal(child.range.end,10015,'relative ranges must keep advancing during automatic refresh')
  assert.deepEqual(requests,{objects:1,history:1})
  timer.advance(50000)
  assert.equal(renders,2,'polling must resume after the previous request finishes')
  assert.deepEqual(requests,{objects:2,history:2})
  assert.equal(parent.loading,false)
})

test('periodic inventory refresh preserves its timeout and changing scope still replaces the request', () => {
  const timer=clock(), pending=[]
  const page=loadComponent('k8s-objects',{...loadHelpers(),...timer,K8sMetricChart:{}})
  const ctx={scope:{clusterId:'cluster-a',namespace:'ns-a'},rows:[],selected:null,filters:{workload:'',pod:'',instanceId:''},page:1,apiCenter:{k8sObjects:'/objects'},request(method,url,body,done){pending.push(done)}}
  page.methods.loadObjects.call(ctx)
  timer.advance(15000)
  page.methods.loadObjects.call(ctx,false,true)
  timer.advance(30000)
  assert.equal(ctx.inventoryFailed,true,'periodic ticks must not postpone the original deadline')
  assert.equal(ctx.loading,false)
  page.methods.loadObjects.call(ctx,false,true)
  ctx.scope.namespace='ns-b'
  page.methods.loadObjects.call(ctx,false,true)
  pending[2]({objects:[{podName:'current'}],state:'ACTIVE'})
  pending[1]({objects:[{podName:'obsolete'}],state:'DEGRADED'})
  assert.equal(ctx.rows[0].podName,'current')
  assert.equal(ctx.state,'ACTIVE')
})

test('chart time changes and object changes replace pending periodic requests', () => {
  const chart=loadComponent('k8s-metric-chart',loadHelpers()), pending=[]
  const page=loadComponent('k8s-objects',{...loadHelpers(),K8sMetricChart:{}})
  const ctx={object:{clusterId:'cluster-a',namespace:'ns-a',podName:'app-0',container:'main'},metric:{id:'cpu'},range:{start:0,end:60,step:30},apiCenter:{k8sMetricHistory:'/history'},renderChart(){},request(method,url,body,done){pending.push(done)}}
  ctx.load=chart.methods.load.bind(ctx)
  ctx.load()
  ctx.periodicRefresh=true
  ctx.range={start:15,end:75,step:30}
  chart.watch.queryKey.call(ctx)
  ctx.timePreset='custom'
  ctx.customRange=[new Date(60000),new Date(120000)]
  ctx.scrapeInterval=30
  page.methods.applyRange.call(ctx)
  chart.watch.queryKey.call(ctx)
  ctx.periodicRefresh=true
  ctx.object={...ctx.object,podName:'app-1'}
  chart.watch.queryKey.call(ctx)
  assert.equal(pending.length,3,'periodic ticks must not replace a pending request, but time and object changes must')
  pending[2]({result:[{metric:{pod:'app-1'},values:[[120,'3']]}]})
  pending[1]({result:[{metric:{pod:'app-0'},values:[[120,'99']]}]})
  pending[0]({result:[{metric:{pod:'app-0'},values:[[60,'99']]}]})
  assert.equal(ctx.series[0].labels.pod,'app-1')
  assert.equal(ctx.series[0].data[2][1],3)
})

test('manual refresh reissues a chart query even for a fixed custom time range', () => {
  const globals = {...loadHelpers(),K8sMetricChart:{}}
  const page = loadComponent('k8s-objects',globals)
  const chart = loadComponent('k8s-metric-chart',globals)
  const ctx = {object:{clusterId:'cluster-a',namespace:'sit-a',podName:'app-0',container:'main'},metric:{id:'cpu'},range:{start:0,end:60,step:30},refreshToken:0,loadObjects(){},applyRange(){}}
  const before = chart.computed.queryKey.call(ctx)
  page.methods.refresh.call(ctx)
  assert.notEqual(chart.computed.queryKey.call(ctx),before)
})

test('chart disposal releases resize handling and rejects responses after unmount', () => {
  let removed,disposed=false,done
  const component = loadComponent('k8s-metric-chart',{...loadHelpers(),window:{removeEventListener(name,callback){removed=callback}}})
  const resize=()=>{}
  const ctx = {object:{podName:'app-0'},metric:{id:'cpu'},range:{start:0,end:60,step:30},request(method,url,body,callback){done=callback},apiCenter:{k8sMetricHistory:'/history'},renderChart(){assert.fail('render after disposal')},_chart:{dispose(){disposed=true}},_resizeChart:resize}
  component.methods.load.call(ctx)
  component.beforeDestroy.call(ctx)
  done({result:[{metric:{},values:[[0,'1']]}]})
  assert.equal(removed,resize)
  assert.equal(disposed,true)
  assert.equal(ctx._chart,null)
  assert.equal(ctx.series.length,0)
})

test('chart options retain gaps and escape application labels in the HTML tooltip', () => {
  const component=loadComponent('k8s-metric-chart',loadHelpers())
  let option
  let width = 1000
  const ctx={range:{start:0,end:60},metric:{unit:'bytes'},series:[{name:'internal',displayName:'pool=<script>bad</script>',data:[[0,null]]},{name:'long',displayName:'x'.repeat(200),data:[[0,1]]},{name:'longer',displayName:'x'.repeat(201),data:[[0,2]]}],$nextTick(callback){callback()},_chart:{getOption(){return undefined},getWidth(){return width},getHeight(){return 280},setOption(value){option=value},resize(){width=320}}}
  component.methods.renderChart.call(ctx)
  assert.equal(option.xAxis.min,0)
  assert.equal(option.xAxis.max,60000)
  assert.equal(option.series[0].data[0][1],null)
  const tooltip=option.tooltip.formatter([{value:[0,1024],seriesIndex:0,seriesName:'internal'}])
  assert.ok(!tooltip.includes('<script>'))
  assert.ok(tooltip.includes('&lt;script&gt;'))
  assert.equal(option.tooltip.enterable,true)
  assert.equal(option.tooltip.confine,true)
  assert.match(option.tooltip.extraCssText,/max-width:296px/)
  assert.match(option.tooltip.extraCssText,/max-height:256px/)
  assert.match(option.tooltip.extraCssText,/overflow-y:auto/)
  assert.equal(option.legend.formatter('internal'),'pool=<script>bad</script>')
  assert.ok(option.legend.formatter('long').length < 100)
  assert.notEqual(option.legend.formatter('long'),option.legend.formatter('longer'))
  const manyPoints = Array.from({length:30},(_,i) => ({value:[0,i],seriesName:'line-'+i}))
  assert.match(option.tooltip.formatter(manyPoints),/line-29/)
  assert.equal(option.tooltip.formatter([]),'')
})

test('expanded chart shares the current data and range without requesting history again', () => {
  const options = []
  let disposed = false
  const expandedChart = {getOption(){return undefined},getWidth(){return 1000},getHeight(){return 600},setOption(option){options.push(option)},resize(){},dispose(){disposed=true}}
  const component = loadComponent('k8s-metric-chart',{...loadHelpers(),echarts:{init(){return expandedChart}}})
  const ctx = {expanded:true,$refs:{expandedChart:{}},range:{start:0,end:60},metric:{unit:'bytes'},series:[{name:'a',displayName:'pool=Young',labels:{pod_uid:'old'},data:[[0,1024]]}],$nextTick(fn){fn()},request(){assert.fail('expanding a chart must not request history')},_chart:{getOption(){return {legend:[{selected:{a:false}}]}},getWidth(){return 320},getHeight(){return 280},setOption(option){options.push(option)},resize(){}}}
  ctx.renderChart = component.methods.renderChart.bind(ctx)
  component.methods.expandedChanged.call(ctx,true)
  assert.equal(options.length,2)
  for (const option of options) {
    assert.equal(option.series,ctx.series)
    assert.equal(option.xAxis.max,60000)
    assert.equal(option.series[0].labels.pod_uid,'old')
    assert.equal(option.legend.selected.a,false)
  }
  assert.match(options[1].tooltip.extraCssText,/max-width:460px/)
  component.methods.expandedChanged.call(ctx,false)
  assert.equal(disposed,true)
  assert.equal(ctx._expandedChart,null)
})

test('closing the expanded chart before the next tick does not create a detached chart', () => {
  const callbacks = []
  const component = loadComponent('k8s-metric-chart',{...loadHelpers(),echarts:{init(){assert.fail('chart initialized after close')}}})
  const ctx = {expanded:true,$refs:{expandedChart:{}},$nextTick(fn){callbacks.push(fn)}}
  component.methods.expandedChanged.call(ctx,true)
  ctx.expanded = false
  component.methods.expandedChanged.call(ctx,false)
  callbacks[0]()
  assert.equal(ctx._expandedChart,null)
})

test('source list failures can be retried and obsolete source responses cannot reset scope', () => {
  const component=loadComponent('k8s-objects',{...loadHelpers(),K8sMetricChart:{}})
  const requests=[]
  const ctx={$route:{query:{clusterId:'cluster-a',namespace:'sit-a'}},scope:{},request(method,url,body,done,config,fail){requests.push({done,fail})},apiCenter:{k8sClusterList:'/clusters',k8sNamespaceList:'/namespaces'},loadObjects(){}}
  component.methods.loadSources.call(ctx)
  requests[0].fail()
  assert.equal(ctx.sourcesFailed,true)
  component.methods.loadSources.call(ctx)
  requests[0].done([{clusterId:'wrong-cluster',enabled:true}])
  assert.equal(requests.length,2)
  requests[1].done([{clusterId:'cluster-a',enabled:true}])
  requests[2].done([{clusterId:'cluster-a',namespace:'sit-a',enabled:true}])
  assert.equal(ctx.scope.clusterId,'cluster-a')
  assert.equal(ctx.scope.namespace,'sit-a')
  assert.equal(ctx.sourcesFailed,false)
})
