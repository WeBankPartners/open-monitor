const {test} = require('node:test')
const assert = require('node:assert/strict')
const fs = require('node:fs')
const path = require('node:path')
const vm = require('node:vm')
const root = path.resolve(__dirname, '../src/views/admin-config/other')

function helpers(file, names, globals = {}) {
  const source = fs.readFileSync(path.join(root, file), 'utf8').replace(/export /g, '')
  return vm.runInNewContext(source + '\nmodule.exports = {' + names.join(',') + '}', {module:{exports:{}},...globals})
}

function setup(name, props = {}, globals = {}) {
  const requests = [], confirmations = [], deadlines = []
  const storage = new Map()
  const sessionStorage = {getItem:key=>storage.get(key) || null,setItem:(key,value)=>storage.set(key,value)}
  const statePath = path.join(root, 'k8s-view-state.js')
  const state = fs.existsSync(statePath) ? helpers('k8s-view-state.js', ['readViewState','writeViewState','resolveScope'], {sessionStorage}) : {}
  const metric = helpers('k8s-metrics.js', ['filterObjects','metricRange','chartSeries','formatMetricValue'])
  const source = fs.readFileSync(path.join(root, name + '.vue'), 'utf8')
  const script = source.match(/<script>([\s\S]*?)<\/script>/)[1].replace(/^import .*$/gm, '').replace('export const', 'const').replace('export default', 'module.exports =')
  const view = vm.runInNewContext(script, {module:{exports:{}},...metric,...state,
    K8sMetricChart:{},K8sLogSamples:{},setTimeout:fn=>{deadlines.push(fn);return deadlines.length},clearTimeout(){},clearInterval(){},
    sessionStorage,setInterval:()=>0,document:{hidden:false},...globals})
  const ctx = {$root:{apiCenter:{k8sMetricHistory:'/history',k8sLogSamples:'/samples',k8sLogAgentSource:'/agent',k8sClusterList:'/clusters',k8sNamespaceList:'/namespaces',k8sLogAgentList:'/agents',k8sAlertRules:'/rules',k8sAlertRule:'/rules/:ruleId',k8sMetricCatalog:'/catalog',k8sObjects:'/objects',k8sClusterSource:'/cluster',k8sNamespaceSource:'/namespace'},
    $httpRequestEntrance:{httpRequestEntrance(method,url,body,done,options,fail){requests.push({method,url,body,done,options,fail})}}},
    $route:{query:{}},$t:(key,args)=>key+(args ? JSON.stringify(args) : ''),$Message:{success(){},error(){}},
    $Modal:{confirm:options=>confirmations.push(options)},$nextTick:fn=>fn(),...props}
  Object.assign(ctx, view.data.call(ctx), props)
  for (const [name, method] of Object.entries(view.methods)) ctx[name] = method.bind(ctx)
  for (const [name, get] of Object.entries(view.computed || {})) Object.defineProperty(ctx, name, {get:()=>get.call(ctx),configurable:true})
  return {view,ctx,requests,confirmations,deadlines,storage,state}
}

const object = {clusterId:'cluster-a',namespace:'ns-a',workloadType:'StatefulSet',workloadName:'app',podName:'app-0',container:'main'}
const range = {start:1000,end:1900,step:30}

test('reading page two of log summaries is not interrupted by the parent periodic refresh', () => {
  const {view,ctx,requests} = setup('k8s-log-samples', {object,range})
  ctx.open()
  requests[0].done({samples:Array.from({length:12},(_,i)=>({observedAt:1800-i,labels:{preview:'sample '+i}})),limit:200,limited:false})
  ctx.page = 2
  ctx.periodicRefresh = true
  ctx.range = {start:1030,end:1930,step:30}
  view.watch.queryKey.call(ctx)
  assert.equal(ctx.page,2)
  assert.equal(ctx.pageRows[0].labels.preview,'sample 5')
  assert.equal(requests.length,1)
  assert.equal(ctx.displayRange.end,1900,'the displayed range must describe the visible summaries')
  ctx.load()
  assert.equal(requests.length,2,'manual refresh must still fetch the current range')
  ctx.periodicRefresh = false
  ctx.object = {...object,podName:'app-1'}
  view.watch.queryKey.call(ctx)
  assert.equal(requests[2].body.podName,'app-1')
})

test('background chart refresh keeps successful data through a failed update but clears a different object', () => {
  const {ctx,requests} = setup('k8s-metric-chart', {object,range,metric:{id:'cpu',unit:'cores'}})
  ctx.renderChart = () => {}
  ctx.load()
  requests[0].done({result:[{metric:{pod_uid:'pod-a'},values:[[1000,'0.5']]}]})
  const series = ctx.series
  const updatedAt = ctx.updatedAt
  ctx.range = {start:1030,end:1930,step:30}
  ctx.load(true)
  assert.equal(ctx.series,series)
  requests[1].fail()
  assert.equal(ctx.series,series)
  assert.equal(ctx.failed,true)
  assert.equal(ctx.updatedAt,updatedAt)
  assert.equal(ctx.loadedRange.end,1900)
  ctx.object = {...object,podName:'app-1'}
  ctx.load(true)
  assert.equal(ctx.series.length,0,'data from another Pod must never be shown as current')
})

test('source disable and node target clearing wait for confirmation and retain the original scope', () => {
  const {ctx,requests,confirmations} = setup('k8s-sources')
  ctx.removeCluster({clusterId:'cluster-a'})
  ctx.removeNamespace({clusterId:'cluster-a',namespace:'ns-a'})
  ctx.agent = {...ctx.agent,loaded:true,clusterId:'cluster-a',node:'node-a',loadedScope:{clusterId:'cluster-a',node:'node-a'}}
  ctx.deleteAgent()
  assert.equal(requests.length,0)
  assert.equal(confirmations.length,3)
  ctx.agent.node = 'node-b'
  confirmations.forEach(value=>value.onOk())
  assert.equal(requests[0].method,'DELETE')
  assert.equal(requests[1].body.namespace,'ns-a')
  assert.equal(requests[2].body.node,'node-a')
})

test('applying log matching rules only changes the outer draft and cancelling retains that draft', () => {
  const {ctx,requests,confirmations} = setup('k8s-sources')
  ctx.agent = {...ctx.agent,loaded:true,clusterId:'cluster-a',node:'node-a',loadedScope:{clusterId:'cluster-a',node:'node-a'},editorVisible:true,targetsJson:JSON.stringify([{id:'target-a',rules:[{id:'error',mode:'literal',pattern:'ERROR'}]}])}
  ctx.agent.originalJson = ctx.agent.targetsJson
  ctx.editLogRules()
  ctx.rulesEditor.targets[0].rules[0].pattern = 'NEW_ERROR'
  ctx.saveLogRules()
  assert.equal(requests.length,0,'inner Apply must not publish')
  assert.equal(JSON.parse(ctx.agent.targetsJson)[0].rules[0].pattern,'NEW_ERROR')
  assert.equal(ctx.rulesEditor.visible,false)
  ctx.closeAgentEditor()
  assert.equal(ctx.agent.editorVisible,true)
  assert.equal(confirmations.length,1)
  confirmations[0].onCancel()
  assert.equal(ctx.agent.editorVisible,true)
  ctx.saveAgent()
  assert.equal(requests.length,1)
  assert.equal(requests[0].body.targets[0].rules[0].pattern,'NEW_ERROR')
})

test('alert rule threshold units are descriptive and saving preserves the original numeric value', () => {
  const {ctx,requests} = setup('k8s-alert-rules')
  ctx.catalog = [{id:'cpu_throttled',unit:'s/s'},{id:'memory',unit:'bytes'}]
  ctx.openEdit({...object,ruleId:'rule-a',metric:'cpu_throttled',threshold:0.25,version:2})
  assert.match(ctx.metricLabel('cpu_throttled'),/m_k8s_metric_cpu_throttled/)
  assert.match(ctx.thresholdHelp,/m_k8s_threshold_throttled_help/)
  ctx.saveRule()
  assert.equal(requests[0].body.threshold,0.25)
  assert.equal(requests[0].body.version,2)
})

test('abnormal filtering distinguishes unknown state and paused alert evaluation', () => {
  const {ctx} = setup('k8s-objects')
  ctx.rows = [
    {...object,statusKnown:false,ready:false,restartCount:0},
    {...object,podName:'not-ready',statusKnown:true,ready:false,restartCount:0},
    {...object,podName:'restarted',statusKnown:true,ready:true,restartCount:1},
    {...object,podName:'firing',statusKnown:true,ready:true,restartCount:0,alerts:[{Status:'firing'}]}
  ]
  ctx.filters.onlyAbnormal = true
  ctx.alertEvaluationEnabled = true
  assert.deepEqual(Array.from(ctx.filteredObjects,row=>row.podName),['not-ready','restarted','firing'])
  ctx.alertEvaluationEnabled = false
  assert.deepEqual(Array.from(ctx.filteredObjects,row=>row.podName),['not-ready','restarted'])
})

test('selecting an object collapses inventory without changing the chosen metrics or custom time range', () => {
  const {ctx} = setup('k8s-objects')
  ctx.timePreset = 'custom'
  ctx.customRange = [new Date(1000000),new Date(1900000)]
  const selections = ctx.metricSelections
  ctx.selectObject(object)
  assert.equal(ctx.inventoryCollapsed,true)
  assert.equal(ctx.metricSelections,selections)
  assert.equal(ctx.range.start,1000)
  assert.equal(ctx.range.end,1900)
})

test('the mounted Card retains a working inventory toggle after the first object selection', async () => {
  const Vue=require('vue')
  const compiler=require('vue-template-compiler')
  const cardSource=fs.readFileSync(path.join(__dirname,'../node_modules/view-design/src/components/card/card.vue'),'utf8')
  const cardScript=cardSource.match(/<script>([\s\S]*?)<\/script>/)[1].replace(/^\s*import .*$/gm,'').replace('export default','module.exports =')
  const cardOptions=vm.runInNewContext(cardScript,{module:{exports:{}},Icon:{},mixinsLink:{}})
  const compile=source=>{
    const result=compiler.compile(source.match(/<template>([\s\S]*?)<\/template>/)[1])
    assert.deepEqual(result.errors,[])
    return {render:new Function(result.render),staticRenderFns:result.staticRenderFns.map(code=>new Function(code))}
  }
  const {view,ctx}=setup('k8s-objects')
  const parent=new Vue({data:()=>({apiCenter:ctx.apiCenter})})
  parent.$httpRequestEntrance={httpRequestEntrance:ctx.request}
  const page=new Vue({...view,parent,components:{...view.components,Card:{...cardOptions,...compile(cardSource)}},...compile(fs.readFileSync(path.join(root,'k8s-objects.vue'),'utf8'))})
  page.$route={query:{}}
  page.$t=key=>key
  const inventory=()=>page._render().children[0]
  const createCard=vnode=>new vnode.componentOptions.Ctor({parent:page,_parentVnode:vnode,_renderChildren:vnode.componentOptions.children})
  const card=createCard(inventory())
  card.$mount()
  // Without a browser, Vue does not insert child components into a DOM tree.
  card._isMounted=true
  card.$options.mounted.forEach(hook=>hook.call(card))
  const find=(node,predicate)=>{
    if (!node) return null
    if (predicate(node)) return node
    for (const child of node.children || (node.componentOptions || {}).children || []) {
      const match=find(child,predicate)
      if (match) return match
    }
    return null
  }
  const renderUpdatedCard=()=>{
    const vnode=inventory()
    const updated=createCard(vnode)
    card.$slots=updated.$slots
    card.$options._parentVnode=vnode
    updated.$destroy()
    return card._render()
  }
  try {
    page.selectObject(object)
    await Vue.nextTick()
    const selected=page.selected, range=page.range, metrics=page.metricSelections
    assert.equal(page.inventoryCollapsed,true)
    let toggle=find(renderUpdatedCard(),node=>node.data && node.data.on && node.data.on.click && (node.children || []).some(child=>child.text==='m_k8s_switch_object'))
    assert.ok(toggle,'the switch-object button must appear after selecting the first object')
    toggle.data.on.click()
    assert.equal(page.inventoryCollapsed,false)
    assert.equal(page.selected,selected)
    assert.equal(page.range,range)
    assert.equal(page.metricSelections,metrics)
    toggle=find(renderUpdatedCard(),node=>node.data && node.data.on && node.data.on.click && (node.children || []).some(child=>child.text==='m_k8s_collapse_inventory'))
    assert.ok(toggle,'the expanded list must still offer its collapse action')
    toggle.data.on.click()
    assert.equal(page.inventoryCollapsed,true)
  } finally {
    card.$destroy()
    page.$destroy()
    parent.$destroy()
  }
})

test('an explicitly requested disabled namespace never falls back to a different namespace', () => {
  for (const name of ['k8s-objects','k8s-alert-rules']) {
    const {ctx,requests} = setup(name, {$route:{query:{clusterId:'cluster-a',namespace:'off'}}})
    ctx.loadSources()
    requests[0].done([{clusterId:'cluster-a',enabled:true}])
    requests[1].done([{clusterId:'cluster-a',namespace:'off',enabled:false},{clusterId:'cluster-a',namespace:'ns-a',enabled:true}])
    assert.equal(ctx.scope.namespace,'',name)
    assert.ok(ctx.unavailableScope,name)
    assert.equal(requests.length,2,name+' must not query another namespace')
  }
})

test('agent list failures remain distinguishable from a successful empty result and can be retried', () => {
  const {ctx,requests,deadlines} = setup('k8s-sources')
  ctx.agent.clusterId = 'cluster-a'
  ctx.loadAgents()
  assert.equal(ctx.agentsLoading,true)
  requests[0].fail()
  assert.equal(ctx.agentsLoading,false)
  assert.equal(ctx.agentsFailed,true)
  ctx.loadAgents()
  requests[1].done({agents:[]})
  assert.equal(ctx.agentsFailed,false)
  assert.equal(ctx.agentsLoading,false)
  assert.equal(ctx.agents.length,0)
  ctx.loadAgents()
  deadlines.at(-1)()
  assert.equal(ctx.agentsFailed,true,'network failures without an HTTP response must not leave a permanent spinner')
})

test('returning to objects restores the selected object, filters, metrics and absolute time range', () => {
  const {view,ctx,requests,state} = setup('k8s-objects')
  state.writeViewState('scope',{clusterId:'cluster-a',namespace:'ns-a'})
  state.writeViewState('objects',{
    scope:{clusterId:'cluster-a',namespace:'ns-a'},filters:{workload:'app',pod:'app-',instanceId:'instance-a',onlyAbnormal:false},
    category:'log',metricSelections:{resource:['memory'],log:['log_matches']},timePreset:'custom',customRange:[1000000,1900000],
    selected:object,autoRefresh:false,inventoryCollapsed:true
  })
  view.mounted.call(ctx)
  requests.find(row=>row.url==='/clusters').done([{clusterId:'cluster-a',enabled:true}])
  requests.find(row=>row.url==='/namespaces').done([{clusterId:'cluster-a',namespace:'ns-a',enabled:true}])
  requests.find(row=>row.url==='/catalog').done({metrics:[],scrapeIntervalSeconds:30,countWindowSeconds:300})
  requests.find(row=>row.url==='/objects').done({objects:[{...object,instanceId:'instance-a'}],state:'ACTIVE'})
  assert.equal(ctx.selected.podName,'app-0')
  assert.equal(ctx.filters.instanceId,'instance-a')
  assert.equal(ctx.category,'log')
  assert.deepEqual(Array.from(ctx.metricSelections.log),['log_matches'])
  assert.equal(ctx.range.start,1000)
  assert.equal(ctx.range.end,1900)
  assert.equal(ctx.autoRefresh,false)
  assert.equal(ctx.inventoryCollapsed,true)
  view.beforeDestroy.call(ctx)
  assert.equal(state.readViewState('objects').selected.podName,'app-0')
})

test('an explicit cluster link does not inherit another cluster remembered namespace', () => {
  for (const name of ['k8s-objects','k8s-alert-rules']) {
    const {ctx,requests,state} = setup(name, {$route:{query:{clusterId:'cluster-b'}}})
    state.writeViewState('scope',{clusterId:'cluster-a',namespace:'ns-a'})
    ctx.loadSources()
    requests[0].done([{clusterId:'cluster-a',enabled:true},{clusterId:'cluster-b',enabled:true}])
    requests[1].done([{clusterId:'cluster-a',namespace:'ns-a',enabled:true},{clusterId:'cluster-b',namespace:'ns-b',enabled:true}])
    assert.equal(ctx.scope.clusterId,'cluster-b',name)
    assert.equal(ctx.scope.namespace,'ns-b',name)
    assert.equal(requests[2].body.namespace,'ns-b',name)
  }
})

test('an object deep link cannot restore a previously selected object outside its filters', () => {
  const {view,ctx,requests,state} = setup('k8s-objects', {$route:{query:{clusterId:'cluster-a',namespace:'ns-a',podName:'app-1'}}})
  state.writeViewState('objects',{scope:{clusterId:'cluster-a',namespace:'ns-a'},selected:object,filters:{}})
  view.mounted.call(ctx)
  requests.find(row=>row.url==='/clusters').done([{clusterId:'cluster-a',enabled:true}])
  requests.find(row=>row.url==='/namespaces').done([{clusterId:'cluster-a',namespace:'ns-a',enabled:true}])
  requests.find(row=>row.url==='/objects').done({objects:[object,{...object,podName:'app-1'}],state:'ACTIVE'})
  assert.equal(ctx.filters.pod,'app-1')
  assert.equal(ctx.selected,null)
})

test('session state recovers from malformed JSON and never stores credentials or target configuration', () => {
  const {view,ctx,state,storage} = setup('k8s-sources')
  storage.set('monitor.k8s.view','invalid-json')
  state.writeViewState('scope',{clusterId:'cluster-a',namespace:'ns-a'})
  assert.equal(state.readViewState('scope').namespace,'ns-a')
  ctx.cluster.token='secret-placeholder'
  ctx.agent.targetsJson='[{"id":"private-config"}]'
  ctx.activeTab='agents'
  view.beforeDestroy.call(ctx)
  const saved=storage.get('monitor.k8s.view')
  assert.ok(!saved.includes('secret-placeholder'))
  assert.ok(!saved.includes('private-config'))
  assert.equal(state.readViewState('sources').activeTab,'agents')
})

test('unsaved inner edits are discarded independently and navigation cancellation preserves the draft', () => {
  const {view,ctx,confirmations,requests} = setup('k8s-sources')
  ctx.agent.loaded=true
  ctx.agent.editorVisible=true
  ctx.agent.targetsJson='[{"id":"target-a","rules":[{"id":"error","mode":"literal","pattern":"ERROR"}]}]'
  ctx.agent.originalJson=ctx.agent.targetsJson
  ctx.editLogRules()
  ctx.rulesEditor.targets[0].rules[0].pattern='NEW'
  ctx.closeRulesEditor()
  assert.equal(ctx.rulesEditor.visible,true)
  confirmations[0].onOk()
  assert.equal(ctx.rulesEditor.visible,false)
  assert.equal(JSON.parse(ctx.agent.targetsJson)[0].rules[0].pattern,'ERROR')
  ctx.agent.targetsJson='[]'
  let navigation
  view.beforeRouteLeave.call(ctx,{}, {},value=>{navigation=value})
  confirmations[1].onCancel()
  assert.equal(navigation,false)
  assert.equal(ctx.agent.targetsJson,'[]')
  assert.equal(requests.length,0)
})

test('failed saves keep the draft and a repeated save click cannot publish twice', () => {
  const {ctx,requests,deadlines} = setup('k8s-sources')
  ctx.agent={...ctx.agent,loaded:true,editorVisible:true,clusterId:'cluster-a',node:'node-a',loadedScope:{clusterId:'cluster-a',node:'node-a'},targetsJson:'[{"id":"target-a"}]'}
  ctx.saveAgent()
  ctx.saveAgent()
  assert.equal(requests.length,1)
  deadlines.at(-1)()
  assert.equal(ctx.agent.saving,false)
  assert.equal(ctx.agent.saveFailed,true)
  assert.equal(ctx.agent.editorVisible,true)
  assert.equal(ctx.agent.targetsJson,'[{"id":"target-a"}]')
  ctx.saveAgent()
  requests[0].done({version:1})
  assert.equal(ctx.agent.saving,true,'a late response from a timed out save must not complete a newer save')
  requests[1].done({version:1})
  assert.equal(ctx.agent.saving,false)
  assert.equal(ctx.agent.editorVisible,false)
})

test('source lists reject obsolete and unmounted responses without reporting empty success', () => {
  const {view,ctx,requests} = setup('k8s-sources')
  ctx.refreshLists()
  assert.equal(ctx.clustersLoading,true)
  ctx.refreshLists()
  requests[0].done([{clusterId:'old'}])
  assert.equal(ctx.clusters.length,0)
  requests[2].done([{clusterId:'cluster-a'}])
  assert.equal(ctx.clusters[0].clusterId,'cluster-a')
  const agentRequest=requests.find(row=>row.url==='/agents')
  view.beforeDestroy.call(ctx)
  agentRequest.done({agents:[{node:'late'}]})
  assert.equal(ctx.agents.length,0)
})

test('object list alert summary prioritizes firing over pending and counts both', () => {
  const {ctx} = setup('k8s-objects')
  ctx.alertEvaluationEnabled=true
  const column=ctx.columns.find(row=>row.title==='m_k8s_object_health')
  const h=(tag,data,children)=>({tag,data,children:children === undefined ? data : children})
  const cell=column.render(h,{row:{...object,statusKnown:true,ready:true,alerts:[{Status:'pending'},{Status:'firing'}]}})
  assert.ok(cell.children.some(tag=>tag.data.props.color==='error' && tag.children.includes('m_k8s_alert_firing') && tag.children.includes('2')))
})

test('compiled K8s templates render localized controls and state without losing threshold units', () => {
  const Vue=require('vue')
  const compiler=require('vue-template-compiler')
  for (const language of ['zh-CN','en']) {
    const messages=JSON.parse(fs.readFileSync(path.join(root,'../../../assets/locale/lang',language+'.json'),'utf8'))
    const missing=[]
    for (const name of ['k8s-sources','k8s-objects','k8s-alert-rules','k8s-metric-chart','k8s-log-samples']) {
      const {view,ctx}=setup(name)
      const source=fs.readFileSync(path.join(root,name+'.vue'),'utf8')
      const template=compiler.compile(source.match(/<template>([\s\S]*?)<\/template>/)[1])
      assert.deepEqual(template.errors,[])
      const parent=new Vue({data:()=>({apiCenter:ctx.apiCenter})})
      parent.$httpRequestEntrance={$httpRequestEntrance:ctx.request,httpRequestEntrance:ctx.request}
      const instance=new Vue({...view,parent,propsData:{object,range,metric:{id:'cpu',unit:'cores'},title:'CPU'},render:new Function(template.render),staticRenderFns:template.staticRenderFns.map(code=>new Function(code))})
      instance.$route={query:{}}
      instance.$t=(key,args={})=>{
        if (!(key in messages)) missing.push(key)
        return (messages[key] || key).replace(/\{(\w+)\}/g,(_,field)=>String(args[field]))
      }
      if (name==='k8s-objects') instance.selected={...object,statusKnown:true,ready:true,restartCount:0,alerts:[{Key:{RuleID:'rule-a'},Status:'firing',Value:0.5}]}
      if (name==='k8s-alert-rules') {
        instance.catalog=[{id:'cpu_throttled',unit:'s/s'}]
        instance.openEdit({...object,metric:'cpu_throttled',ruleId:'rule-a',threshold:0.25,version:2})
      }
      const rendered=instance._render()
      assert.ok(rendered)
      const texts=[]
      const visit=node=>{if (!node) return;if (node.text) texts.push(node.text);for (const child of node.children || []) visit(child);for (const child of (node.componentOptions || {}).children || []) visit(child)}
      visit(rendered)
      if (name==='k8s-alert-rules') assert.ok(texts.join(' ').includes('s/s'))
      if (name==='k8s-sources') assert.ok(texts.join(' ').includes(messages.m_k8s_save_publish))
    }
    assert.deepEqual(missing,[])
  }
})
