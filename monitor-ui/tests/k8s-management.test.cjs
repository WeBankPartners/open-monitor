const { test } = require('node:test')
const assert = require('node:assert/strict')
const fs = require('node:fs')
const vm = require('node:vm')
const path = require('node:path')
const root = path.resolve(__dirname, '..')

test('log rule editor saves optional exclusions and reloads server-managed versions', () => {
  const source = fs.readFileSync(path.join(root, 'src/views/admin-config/other/k8s-sources.vue'), 'utf8')
  const component = vm.runInNewContext(source.match(/<script>([\s\S]*?)<\/script>/)[1].replace('export default', 'module.exports ='), { module: { exports: {} } })
  let saved, loaded = false
  const context = { agent: {clusterId:'cluster-a',node:'node-a',loaded:true,loadedScope:{clusterId:'cluster-a',node:'node-a'},targetsJson:JSON.stringify([{id:'app',rules:[{id:'error',version:3,mode:'literal',pattern:'ERROR'}]}])}, rulesEditor: {visible:false}, apiCenter:{k8sLogAgentSource:'/source'}, $t:k=>k, $Message:{error:message=>assert.fail(message),success(){}}, loadAgent(){loaded=true}, request(method,url,body,done){saved=body;done()} }
  Object.assign(context, {saveAgent:component.methods.saveAgent, saveLogRules:component.methods.saveLogRules})
  component.methods.editLogRules.call(context)
  const rule = context.rulesEditor.targets[0].rules[0]
  component.methods.addLogExclusion.call(context, rule)
  rule.exclude[0].pattern = 'EXPECTED_ERROR'
  context.saveLogRules()
  assert.equal(saved.targets[0].rules[0].exclude[0].mode, 'literal')
  assert.equal(saved.targets[0].rules[0].exclude[0].pattern, 'EXPECTED_ERROR')
  assert.equal(JSON.parse(context.agent.targetsJson)[0].rules[0].exclude.length, 1)
  assert.equal(loaded, true)
  assert.equal(context.rulesEditor.visible, false)
})

function loadComponent(name) {
  const source = fs.readFileSync(path.join(root, 'src/views/admin-config/other/' + name + '.vue'), 'utf8')
  return vm.runInNewContext(source.match(/<script>([\s\S]*?)<\/script>/)[1].replace('export const', 'const').replace('export default', 'module.exports ='), { module: { exports: {} } })
}

test('agent editor ignores obsolete responses and cannot save another node configuration', () => {
  const component = loadComponent('k8s-sources')
  const pending = []
  const context = {agent:{clusterId:'cluster-a',node:'node-a',targetsJson:'[]'},apiCenter:{k8sLogAgentSource:'/source'},$t:k=>k,$Message:{error(){},success(){}},request(method,url,body,done){assert.equal(method,'GET');pending.push(done)}}
  for (const node of ['node-a','node-b','node-a']) {
    context.agent.node = node
    component.methods.loadAgent.call(context)
  }
  pending[2]({targets:[{id:'current'}]})
  pending[0]({targets:[{id:'stale'}]})
  pending[1]({targets:[{id:'wrong-node'}]})
  assert.equal(JSON.parse(context.agent.targetsJson)[0].id,'current')
  context.agent.node = 'node-b'
  component.methods.saveAgent.call(context)
  component.methods.deleteAgent.call(context)
})

test('agent list ignores responses from an older cluster selection', () => {
  const component = loadComponent('k8s-sources')
  const pending = []
  const context = {agent:{clusterId:'cluster-a'},clusters:[],agents:[],apiCenter:{k8sLogAgentList:'/list'},request(method,url,body,done){pending.push(done)}}
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
  const source = fs.readFileSync(path.join(root, 'src/views/admin-config/other/k8s-sources.vue'), 'utf8')
  const component = vm.runInNewContext(source.match(/<script>([\s\S]*?)<\/script>/)[1].replace('export default', 'module.exports ='), { module: { exports: {} } })
  let message
  const context = {rulesEditor:{visible:true,targets:[{rules:[{id:'error',mode:'literal',pattern:'ERROR',exclude:[{mode:'literal',pattern:''}]}]}]}, $t:k=>k, $Message:{error:value=>{message=value}}, saveAgent(){assert.fail('invalid configuration submitted')} }
  component.methods.saveLogRules.call(context)
  assert.equal(message, 'm_k8s_invalid_log_rule')
  assert.equal(context.rulesEditor.visible, true)
})

test('log rule editor rejects malformed target and exclusion JSON without crashing', () => {
  const source = fs.readFileSync(path.join(root, 'src/views/admin-config/other/k8s-sources.vue'), 'utf8')
  const component = vm.runInNewContext(source.match(/<script>([\s\S]*?)<\/script>/)[1].replace('export default', 'module.exports ='), { module: { exports: {} } })
  for (const targetsJson of ['[null]', '[{"rules":"invalid"}]', '[{"rules":[{"exclude":"invalid"}]}]']) {
    let message
    const context = {agent:{targetsJson},rulesEditor:{visible:false},$t:k=>k,$Message:{error:value=>{message=value}}}
    component.methods.editLogRules.call(context)
    assert.equal(message,'m_k8s_invalid_json')
    assert.equal(context.rulesEditor.visible,false)
  }
})

test('editing an automatically registered agent explicitly overrides its address', () => {
  const source = fs.readFileSync(path.join(root, 'src/views/admin-config/other/k8s-sources.vue'), 'utf8')
  const script = source.match(/<script>([\s\S]*?)<\/script>/)[1]
  const component = vm.runInNewContext(script.replace('export default', 'module.exports ='), { module: { exports: {} } })
  const context = { endpoint: { clusterId: 'cluster-a', node: 'node-a', address: 'agent.example.test', port: 9105, manualOverride: false }, apiCenter: { k8sLogAgentEndpoint: '/endpoint' }, request(method, url, body) { assert.equal(method, 'PUT'); assert.equal(body.manualOverride, true) } }
  component.methods.saveEndpoint.call(context)
})

test('generated administrator permissions contain every Kubernetes management operation', () => {
  let output
  const script = fs.readFileSync(path.join(root, 'node-api-config.js'), 'utf8')
  const proxyFs = { ...fs, writeFileSync(file, data) { output = JSON.parse(data) } }
  vm.runInNewContext(script, { require(name) { return name === 'fs' ? proxyFs : require(require.resolve(name, { paths: [root] })) }, __dirname: root, console }, { timeout: 30000 })
  const urls = output.find(row => row.menu === 'MONITOR_ADMINISTRATOR_CONFIG').urls
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
  assert.equal(route.path,'/adminConfig/k8sObjects')
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
