const { test } = require('node:test')
const assert = require('node:assert/strict')
const fs = require('node:fs')
const vm = require('node:vm')
const path = require('node:path')
const root = path.resolve(__dirname, '..')

test('log rule editor applies exclusions to the draft before publishing and reloading versions', () => {
  const component = loadComponent('k8s-sources')
  let saved, loaded = false
  const context = { agent: {clusterId:'cluster-a',node:'node-a',loaded:true,loadedScope:{clusterId:'cluster-a',node:'node-a'},targetsJson:JSON.stringify([{id:'app',rules:[{id:'error',version:3,mode:'literal',pattern:'ERROR'}]}])}, rulesEditor: {visible:false}, apiCenter:{k8sLogAgentSource:'/source'}, $t:k=>k, $Message:{error:message=>assert.fail(message),success(){}}, loadAgent(){loaded=true},loadAgents(){}, request(method,url,body,done){saved=body;done()} }
  Object.assign(context, {saveAgent:component.methods.saveAgent, saveLogRules:component.methods.saveLogRules})
  component.methods.editLogRules.call(context)
  const rule = context.rulesEditor.targets[0].rules[0]
  component.methods.addLogExclusion.call(context, rule)
  rule.exclude[0].pattern = 'EXPECTED_ERROR'
  context.saveLogRules()
  assert.equal(saved,undefined,'applying rules must only update the draft')
  context.saveAgent()
  assert.equal(saved.targets[0].rules[0].exclude[0].mode, 'literal')
  assert.equal(saved.targets[0].rules[0].exclude[0].pattern, 'EXPECTED_ERROR')
  assert.equal(JSON.parse(context.agent.targetsJson)[0].rules[0].exclude.length, 1)
  assert.equal(loaded, true)
  assert.equal(context.rulesEditor.visible, false)
})

function loadComponent(name, globals = {}) {
  const source = fs.readFileSync(path.join(root, 'src/views/admin-config/other/' + name + '.vue'), 'utf8')
  const stateSource = fs.readFileSync(path.join(root, 'src/views/admin-config/other/k8s-view-state.js'), 'utf8')
  const state = vm.runInNewContext(stateSource.replace(/export /g,'')+'\nmodule.exports={readViewState,writeViewState,resolveScope}',{module:{exports:{}}})
  return vm.runInNewContext(source.match(/<script>([\s\S]*?)<\/script>/)[1].replace(/^import .*$/gm,'').replace('export const', 'const').replace('export default', 'module.exports ='), { module: { exports: {} }, setTimeout: () => 0, clearTimeout() {}, ...state,...globals })
}

test('reset defaults changes all ordinary log targets in the draft and saves only on request', () => {
  const component = loadComponent('k8s-sources')
  const pending = []
  const context = {
    agent: {clusterId:'cluster-a',node:'node-a'},
    apiCenter: {k8sLogAgentSource:'/source'},
    $t:key=>key,
    $Message: {error:message=>assert.fail(message),success(){}},
    request(method,url,body,done){pending.push({method,url,body,done})}
  }
  const defaults = [{id:'default-error',version:1,mode:'regex',pattern:'server-default'}]
  const fileExclusions = ['*.log*2026-*.log','*.log*2026*.log']
  const targets = [
    {id:'target-a',namespace:'ns-a',podUid:'pod-a',directory:'/var/log/app',glob:'*.log',rules:[{id:'default-error',version:3,mode:'literal',pattern:'old',exclude:[{mode:'literal',pattern:'custom-exclusion'}]},{id:'custom',version:8,mode:'literal',pattern:'CUSTOM'}]},
    {id:'target-b',namespace:'ns-b',rules:[]},
    {id:'metrics',format:'otlp_metric',rules:null}
  ]
  component.methods.loadAgent.call(context, true)
  pending[0].done({targets,defaultRules:defaults,defaultExcludeGlobs:fileExclusions})
  component.methods.resetDefaultLogRules.call(context)
  const reset = JSON.parse(context.agent.targetsJson)
  assert.equal(pending.length,1,'reset must not save configuration')
  assert.deepEqual(reset[0], {...targets[0],excludeGlobs:fileExclusions,rules:[defaults[0],targets[0].rules[1]]})
  assert.deepEqual(reset[1].rules,defaults)
  assert.deepEqual(reset[1].excludeGlobs,fileExclusions)
  assert.deepEqual(reset[2],targets[2])
  const draft = context.agent.targetsJson
  component.methods.resetDefaultLogRules.call(context)
  assert.equal(context.agent.targetsJson,draft,'repeated reset must be idempotent')
  component.methods.editLogRules.call(context)
  context.rulesEditor.targets[0].rules[0].pattern='edited-draft'
  assert.equal(context.rulesEditor.targets[1].rules[0].pattern,'server-default')
  assert.equal(context.agent.defaultRules[0].pattern,'server-default')
  component.methods.saveAgent.call(context)
  assert.equal(pending[1].method,'PUT')
  assert.deepEqual(JSON.parse(JSON.stringify(pending[1].body.targets)),reset)
})

test('reset defaults rejects invalid JSON and leaves OTLP-only targets unchanged', () => {
  const component = loadComponent('k8s-sources')
  for (const targetsJson of ['{', '[null]', '[{"rules":"invalid"}]', '[{"rules":[null]}]', '[{"format":"otlp_metric","rules":null}]']) {
    let message
    const context = {
      agent: {clusterId:'cluster-a',node:'node-a',loaded:true,loadedScope:{clusterId:'cluster-a',node:'node-a'},targetsJson,defaultRules:[{id:'default-error',version:1,mode:'regex',pattern:'server-default'}]},
      $t:key=>key,$Message:{error:value=>{message=value},success:()=>assert.fail('invalid reset succeeded')},
      request:()=>assert.fail('reset submitted a request')
    }
    component.methods.resetDefaultLogRules.call(context)
    assert.equal(context.agent.targetsJson,targetsJson)
    assert.equal(message,targetsJson.includes('otlp_metric')?'m_k8s_no_log_targets':'m_k8s_invalid_json')
  }
})

function alertRuleContext(component) {
  const pending = []
  const catalogPending = []
  const context = {
    $root: {$httpRequestEntrance: {httpRequestEntrance(method, url, body, done, config, fail) { (url === '/catalog' ? catalogPending : pending).push({method, url, body, done, fail}) }}, apiCenter: {k8sClusterList: '/clusters', k8sNamespaceList: '/namespaces', k8sAlertRules: '/rules', k8sMetricCatalog:'/catalog'}},
    $t: key => key,
    $Message: {error: message => assert.fail(message)}
  }
  Object.assign(context, component.data.call(context))
  for (const [name, method] of Object.entries(component.methods)) context[name] = method.bind(context)
  Object.defineProperty(context, 'namespaceOptions', {get: () => component.computed.namespaceOptions.call(context)})
  return {context, pending, catalogPending}
}

test('alert rule mode is read from catalog without preventing rule management', () => {
  const component = loadComponent('k8s-alert-rules')
  const {context,pending,catalogPending} = alertRuleContext(component)
  component.mounted.call(context)
  assert.equal(catalogPending.length,1)
  catalogPending[0].done({localCollectionEnabled:false,alertEvaluationEnabled:false})
  assert.equal(context.alertEvaluationEnabled,false)
  context.scope={clusterId:'cluster-a',namespace:'ns-a'}
  context.openCreate()
  assert.equal(context.editor.visible,true)
  context.loadCatalog()
  catalogPending[1].fail()
  assert.equal(context.alertEvaluationEnabled,null)
  context.loadCatalog()
  component.beforeDestroy.call(context)
  catalogPending[2].done({localCollectionEnabled:true,alertEvaluationEnabled:true})
  assert.equal(context.alertEvaluationEnabled,null,'late response must not update an unmounted page')
  assert.equal(pending[0].url,'/clusters')
})

test('alert rule scope selects enabled sources and keeps rules within the selected cluster', () => {
  const component = loadComponent('k8s-alert-rules')
  const {context, pending} = alertRuleContext(component)
  component.mounted.call(context)
  assert.equal(pending[0].url, '/clusters')
  assert.equal(pending[0].method, 'POST')
  pending[0].done([{clusterId: 'disabled', enabled: false}, {clusterId: 'cluster-a', enabled: true}, {clusterId: 'cluster-b', enabled: true}])
  assert.equal(pending[1].url, '/namespaces')
  pending[1].done([{clusterId: 'cluster-a', namespace: 'off', enabled: false}, {clusterId: 'cluster-a', namespace: 'ns-a', enabled: true}, {clusterId: 'cluster-b', namespace: 'ns-b', enabled: true}])
  assert.equal(context.clusters.length, 2)
  assert.equal(context.namespaceOptions.length, 1)
  assert.equal(context.namespaceOptions[0].namespace, 'ns-a')
  assert.equal(context.scope.clusterId, 'cluster-a')
  assert.equal(context.scope.namespace, 'ns-a')
  assert.equal(pending[2].url, '/rules')
  assert.equal(pending[2].body.namespace, 'ns-a')
  context.rules = [{ruleId: 'previous'}]
  context.scope.clusterId = 'cluster-b'
  context.clusterChanged()
  assert.equal(context.rules.length, 0)
  assert.equal(context.namespaceOptions[0].namespace, 'ns-b')
  assert.equal(context.scope.namespace, 'ns-b')
  assert.equal(pending[3].body.clusterId, 'cluster-b')
  pending[3].done([{ruleId: 'current'}])
  pending[2].done([{ruleId: 'stale'}])
  assert.equal(context.rules[0].ruleId, 'current')
})

test('alert rule scope with no enabled namespace does not query rules', () => {
  const component = loadComponent('k8s-alert-rules')
  const {context, pending} = alertRuleContext(component)
  context.loadSources()
  pending[0].done([{clusterId: 'cluster-a', enabled: true}])
  pending[1].done([{clusterId: 'cluster-a', namespace: 'off', enabled: false}])
  assert.equal(context.sourcesLoading, false)
  assert.equal(context.scope.namespace, '')
  assert.equal(context.namespaceOptions.length, 0)
  assert.equal(context.rules.length, 0)
  assert.equal(pending.length, 2)
})

test('alert rule source failure is recoverable and late responses cannot replace the retried scope', () => {
  let deadline
  const component = loadComponent('k8s-alert-rules', {setTimeout: fn => { deadline = fn; return 1 }})
  const {context, pending} = alertRuleContext(component)
  context.loadSources()
  deadline()
  assert.equal(context.sourcesLoading, false)
  assert.equal(context.sourcesFailed, true)
  pending[0].done([{clusterId: 'stale', enabled: true}])
  assert.equal(pending.length, 1)
  context.loadSources()
  pending[1].done([{clusterId: 'cluster-b', enabled: true}])
  pending[2].done([{clusterId: 'cluster-b', namespace: 'ns-b', enabled: true}])
  assert.equal(context.sourcesFailed, false)
  assert.equal(context.scope.clusterId, 'cluster-b')
  component.beforeDestroy.call(context)
  pending[3].done([{ruleId: 'stale-after-destroy'}])
  assert.equal(context.rules.length, 0)
})

test('agent editor ignores obsolete responses and cannot save another node configuration', () => {
  const component = loadComponent('k8s-sources')
  const pending = []
  const context = {agent:{clusterId:'cluster-a',node:'node-a',targetsJson:'[]'},apiCenter:{k8sLogAgentSource:'/source'},$t:k=>k,$Message:{error(){},success(){}},request(method,url,body,done){assert.equal(method,'GET');pending.push(done)}}
  for (const node of ['node-a','node-b','node-a']) {
    context.agent.node = node
    component.methods.loadAgent.call(context)
  }
  pending[2]({targets:[{id:'current'}],defaultRules:[{id:'default-error',pattern:'current-default'}]})
  pending[0]({targets:[{id:'stale'}],defaultRules:[{id:'default-error',pattern:'stale-default'}]})
  pending[1]({targets:[{id:'wrong-node'}],defaultRules:[{id:'default-error',pattern:'wrong-default'}]})
  assert.equal(JSON.parse(context.agent.targetsJson)[0].id,'current')
  assert.equal(context.agent.defaultRules[0].pattern,'current-default')
  context.agent.node = 'node-b'
  component.methods.resetDefaultLogRules.call(context)
  assert.equal(JSON.parse(context.agent.targetsJson)[0].id,'current')
  component.methods.saveAgent.call(context)
  component.methods.deleteAgent.call(context)
})

test('agent list ignores responses from an older cluster selection', () => {
  const component = loadComponent('k8s-sources')
  const pending = []
  const context = {agent:{clusterId:'cluster-a'},clusters:[],agents:[],apiCenter:{k8sLogAgentList:'/list'},request(method,url,body,done){pending.push(done)}}
  context.loadManagementList=component.methods.loadManagementList.bind(context)
  component.methods.loadAgents.call(context)
  context.agent.clusterId = 'cluster-b'
  component.methods.loadAgents.call(context)
  pending[1]({agents:[{node:'current'}]}); pending[0]({agents:[{node:'stale'}]})
  assert.equal(context.agents[0].node,'current')
})

test('alert rule updates and delayed deletion keep the original namespace', () => {
  const component = loadComponent('k8s-alert-rules')
  const requests = []
  let confirmation
  const row = {clusterId:'cluster-a',namespace:'sit-a',ruleId:'cpu',version:2}
  const context = {scope:{clusterId:'cluster-b',namespace:'uat-a'},editor:{},apiCenter:{k8sAlertRule:'/rules/:ruleId'},$t:k=>k,$Modal:{confirm(value){confirmation=value}},request(method,url,body){requests.push(body)}}
  component.methods.openEdit.call(context,row)
  component.methods.saveRule.call(context)
  component.methods.deleteRule.call(context,row)
  context.scope.namespace = 'uat-b'
  confirmation.onOk()
  for (const body of requests) {
    assert.equal(body.clusterId,'cluster-a')
    assert.equal(body.namespace,'sit-a')
  }
  assert.equal(context.editor.visible,true, 'failed save must keep the editor open')
})

test('alert rule list ignores obsolete namespace requests', () => {
  const component = loadComponent('k8s-alert-rules')
  const pending = []
  const context = {scope:{clusterId:'cluster-a',namespace:'sit-a'},apiCenter:{k8sAlertRules:'/rules'},request(method,url,body,done){pending.push(done)}}
  for (const namespace of ['sit-a','sit-b','sit-a']) {
    context.scope.namespace=namespace
    component.methods.loadRules.call(context)
  }
  pending[2]([{ruleId:'current'}]);pending[0]([{ruleId:'stale'}]);pending[1]([{ruleId:'wrong-namespace'}])
  assert.equal(context.rules[0].ruleId,'current')
})

test('log rule editor rejects empty exclusions without submitting configuration', () => {
  const component = loadComponent('k8s-sources')
  let message
  const context = {rulesEditor:{visible:true,targets:[{rules:[{id:'error',mode:'literal',pattern:'ERROR',exclude:[{mode:'literal',pattern:''}]}]}]}, $t:k=>k, $Message:{error:value=>{message=value}}, saveAgent(){assert.fail('invalid configuration submitted')} }
  component.methods.saveLogRules.call(context)
  assert.equal(message, 'm_k8s_invalid_log_rule')
  assert.equal(context.rulesEditor.visible, true)
})

test('log rule editor rejects malformed target and exclusion JSON without crashing', () => {
  const component = loadComponent('k8s-sources')
  for (const targetsJson of ['[null]', '[{"rules":"invalid"}]', '[{"rules":[{"exclude":"invalid"}]}]']) {
    let message
    const context = {agent:{targetsJson},rulesEditor:{visible:false},$t:k=>k,$Message:{error:value=>{message=value}}}
    component.methods.editLogRules.call(context)
    assert.equal(message,'m_k8s_invalid_json')
    assert.equal(context.rulesEditor.visible,false)
  }
})

test('editing an automatically registered agent explicitly overrides its address', () => {
  const component = loadComponent('k8s-sources')
  const context = { endpoint: { clusterId: 'cluster-a', node: 'node-a', address: 'agent.example.test', port: 9105, manualOverride: false }, apiCenter: { k8sLogAgentEndpoint: '/endpoint' }, request(method, url, body) { assert.equal(method, 'PUT'); assert.equal(body.manualOverride, true) } }
  component.methods.saveEndpoint.call(context)
})

test('generated Kubernetes menu permissions contain every Kubernetes management operation', () => {
  let output
  const script = fs.readFileSync(path.join(root, 'node-api-config.js'), 'utf8')
  const proxyFs = { ...fs, writeFileSync(file, data) { output = JSON.parse(data) } }
  vm.runInNewContext(script, { require(name) { return name === 'fs' ? proxyFs : require(require.resolve(name, { paths: [root] })) }, __dirname: root, console }, { timeout: 30000 })
  const urls = output.find(row => row.menu === 'MONITOR_K8S_CONFIG').urls
  assert.ok(!output.find(row => row.menu === 'MONITOR_ADMINISTRATOR_CONFIG').urls.some(row => row.url.startsWith('/monitor/api/v1/k8s/')))
  assert.ok(urls.some(row => row.url === '/monitor/api/v1/k8s/alerts/states' && row.method === 'get'))
  for (const [suffix, methods] of Object.entries({ 'objects': ['get'], 'metrics/history': ['post'], 'metrics/catalog': ['get'], 'cluster/source': ['put', 'delete'], 'namespace/source': ['put', 'delete'], 'cluster/list': ['post'], 'namespace/list': ['post'], 'log-agent/list': ['get'], 'log-agent/source': ['get', 'put', 'delete'], 'log-agent/endpoint': ['put', 'delete'], 'log-agent/endpoint/override': ['delete'], 'alert-rules': ['get', 'post'], 'alert-rules/:ruleId': ['put', 'delete'] })) {
    for (const method of methods) assert.ok(urls.some(row => row.url === '/monitor/api/v1/k8s/' + suffix && row.method === method), method + ' ' + suffix)
  }
})

test('agent inventory GET route is assigned an API permission code', () => {
  const source = fs.readFileSync(path.join(root, '../monitor-server/api/api.go'), 'utf8')
  assert.match(source, /apiCodeMap\[http\.MethodGet\+"_"\+urlPrefix\+"\/api\/v1\/k8s\/log-agent\/list"\]\s*=\s*"k8s_log_agent_list"/)
})

test('namespace object action opens the complete browser with its original scope', () => {
  const component = loadComponent('k8s-sources')
  let route
  component.methods.openObjects.call({$router:{push(value){route=value}}},{clusterId:'cluster-a',namespace:'sit-a'})
  assert.equal(route.path,'/k8sConfig/k8sObjects')
  assert.equal(route.query.clusterId,'cluster-a')
  assert.equal(route.query.namespace,'sit-a')
})

test('cluster creation requires a token while cluster edits may keep the current token', () => {
  const component = loadComponent('k8s-sources')
  const requests = []
  const errors = []
  const context = {
    cluster: { clusterId: 'cluster-a', apiServer: 'https://api.example.invalid', ca: '', token: '', enabled: true, editing: false },
    apiCenter: { k8sClusterSource: '/cluster/source' },
    $t: key => key,
    $Message: { error: value => errors.push(value), success() {} },
    request(method, url, body) { requests.push({ method, url, body }) },
    refreshLists() {}
  }
  component.methods.saveCluster.call(context)
  assert.deepEqual(errors, ['m_k8s_cluster_token_required'])
  assert.equal(requests.length, 0)

  context.cluster.editing = true
  component.methods.saveCluster.call(context)
  assert.equal(requests.length, 1)
  assert.equal(requests[0].body.token, '')
})

test('cluster and namespace forms are opened in dedicated dialogs', () => {
  const source = fs.readFileSync(path.join(root, 'src/views/admin-config/other/k8s-sources.vue'), 'utf8')
  assert.match(source, /<Modal[^>]*v-model="cluster\.visible"/)
  assert.match(source, /<Modal[^>]*v-model="namespace\.visible"/)
  assert.doesNotMatch(source, /<Form inline>/)
})

function loadRoutes(file) {
  const source = fs.readFileSync(path.join(root, 'project-config', file), 'utf8')
    .replace(/^import (\w+) from [^\n]+;?$/gm, (_, name) => {
      if (name === 'Vue') return 'const Vue = {use() {}};'
      if (name === 'Router') return 'const Router = routerConstructor;'
      return 'const ' + name + ' = {};'
    })
    .replace('export default router;', 'module.exports = router;')
  const context = {module: {exports: {}}, routerConstructor: require('vue-router')}
  vm.runInNewContext(source, context)
  return context.module.exports
}

test('plugin and standalone routes group all Kubernetes pages and preserve old links', () => {
  for (const file of ['router-plugin.js', 'router.js']) {
    const loaded = loadRoutes(file)
    const routes = Array.isArray(loaded) ? loaded : loaded.options.routes[0].children
    const group = routes.find(row => row.name === 'k8sConfig')
    assert.ok(group, file + ' is missing the Kubernetes menu')
    assert.equal(group.redirect, '/k8sConfig/k8sSources')
    const admin = routes.find(row => row.name === 'adminConfig')
    const router = Array.isArray(loaded) ? new (require('vue-router'))({routes: loaded}) : loaded
    for (const name of ['k8sSources', 'k8sObjects', 'k8sAlertRules']) {
      assert.ok(group.children.some(row => row.name === name), file + ': ' + name)
      const old = admin.children.find(row => row.path === name)
      const query = {clusterId: 'cluster-a', namespace: 'sit-a'}
      const target = old.redirect({query, hash: '#detail'})
      assert.equal(target.path, '/k8sConfig/' + name)
      assert.equal(target.query, query)
      assert.equal(target.hash, '#detail')
      const resolved = router.resolve('/adminConfig/' + name + '?clusterId=cluster-a&namespace=sit-a#detail').route
      assert.equal(resolved.path, '/k8sConfig/' + name)
      assert.equal(resolved.query.namespace, 'sit-a')
      assert.equal(resolved.hash, '#detail')
      assert.ok(resolved.matched.some(row => row.name === 'k8sConfig'))
    }
  }
})
