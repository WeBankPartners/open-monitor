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
  return vm.runInNewContext(script,{module:{exports:{}},setTimeout:()=>0,clearTimeout(){},...globals})
}

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
  const ctx = {autoRefresh:true,scrapeInterval:60,refresh(){refreshes++}}
  component.methods.startRefresh.call(ctx)
  assert.equal(callbacks[0].delay,60000)
  callbacks[0].callback()
  assert.equal(refreshes,1)
  ctx.autoRefresh=false
  component.methods.startRefresh.call(ctx)
  assert.ok(cleared.includes(1))
  component.beforeDestroy.call(ctx)
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
  const ctx={range:{start:0,end:60},metric:{unit:'bytes'},series:[{data:[[0,null]]}],$nextTick(callback){callback()},_chart:{setOption(value){option=value},resize(){}}}
  component.methods.renderChart.call(ctx)
  assert.equal(option.xAxis.min,0)
  assert.equal(option.xAxis.max,60000)
  assert.equal(option.series[0].data[0][1],null)
  const tooltip=option.tooltip.formatter([{value:[0,1024],seriesName:'pool=<script>bad</script>'}])
  assert.ok(!tooltip.includes('<script>'))
  assert.ok(tooltip.includes('&lt;script&gt;'))
})

test('source list failures can be retried and obsolete source responses cannot reset scope', () => {
  const component=loadComponent('k8s-objects',{...loadHelpers(),K8sMetricChart:{}})
  const requests=[]
  const ctx={$route:{query:{clusterId:'cluster-a',namespace:'sit-a'}},scope:{},request(method,url,body,done,config,fail){requests.push({done,fail})},apiCenter:{k8sClusterList:'/clusters',k8sNamespaceList:'/namespaces'},clusterChanged(namespace){this.scope.namespace=namespace}}
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
