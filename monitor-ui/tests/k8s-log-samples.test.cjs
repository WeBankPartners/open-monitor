const {test} = require('node:test')
const assert = require('node:assert/strict')
const fs = require('node:fs')
const path = require('node:path')
const vm = require('node:vm')

function component(globals = {}) {
  const source = fs.readFileSync(path.join(__dirname, '../src/views/admin-config/other/k8s-log-samples.vue'), 'utf8')
  return vm.runInNewContext(source.match(/<script>([\s\S]*?)<\/script>/)[1].replace('export default', 'module.exports ='), {module:{exports:{}},setTimeout:()=>0,clearTimeout(){},...globals})
}

function setup() {
  const view = component()
  const requests = []
  const ctx = {
    object:{clusterId:'cluster-a',namespace:'ns-a',workloadType:'StatefulSet',workloadName:'app',podName:'app-0',container:'main'},
    range:{start:1000,end:1900,step:300},refreshToken:0,
    $root:{apiCenter:{k8sLogSamples:'/samples'},$httpRequestEntrance:{httpRequestEntrance(method,url,body,done,options,fail){requests.push({method,url,body,done,fail})}}}
  }
  Object.assign(ctx,view.data.call(ctx))
  for (const [name, method] of Object.entries(view.methods)) {ctx[name] = method.bind(ctx)}
  return {view,ctx,requests}
}

test('opening summaries sends the selected object and full time range independently of chart step', () => {
  const {view,ctx,requests} = setup()
  view.watch.queryKey.call(ctx)
  assert.equal(requests.length,0,'hidden summaries must not issue queries')
  ctx.open()
  assert.equal(ctx.visible,true)
  assert.equal(ctx.loading,true)
  assert.equal(requests[0].url,'/samples')
  assert.deepEqual(JSON.parse(JSON.stringify(requests[0].body)),{clusterId:'cluster-a',namespace:'ns-a',workloadType:'StatefulSet',workloadName:'app',podName:'app-0',container:'main',start:1000,end:1900})
  const rows=[{observedAt:1800,labels:{pod_uid:'old-pod',file:'app.log',byte_offset:'9007199254740993',preview:'ERROR <script>text</script>'}}]
  requests[0].done({samples:rows,limited:true,limit:200})
  assert.equal(ctx.rows,rows)
  assert.equal(ctx.rows[0].labels.byte_offset,'9007199254740993')
  assert.equal(ctx.loading,false)
  assert.equal(ctx.failed,false)
  assert.equal(ctx.limited,true)
})

test('old object responses and closed drawer responses cannot replace current summaries', () => {
  const {view,ctx,requests} = setup()
  ctx.open()
  ctx.object.podName='app-1'
  view.watch.queryKey.call(ctx)
  requests[1].done({samples:[{observedAt:1800,labels:{pod:'app-1'}}],limited:false,limit:200})
  requests[0].done({samples:[{observedAt:1700,labels:{pod:'app-0'}}],limited:false,limit:200})
  requests[0].fail()
  assert.equal(ctx.rows[0].labels.pod,'app-1')
  assert.equal(ctx.failed,false)
  ctx.load()
  ctx.visible=false
  ctx.visibleChanged(false)
  requests[2].done({samples:[{labels:{pod:'obsolete'}}]})
  assert.equal(ctx.rows.length,0)
})

test('empty summaries and query failures have different states and can be retried', () => {
  const {ctx,requests} = setup()
  ctx.open()
  requests[0].done({samples:[],limited:false,limit:200})
  assert.equal(ctx.rows.length,0)
  assert.equal(ctx.failed,false)
  ctx.load()
  requests[1].fail()
  assert.equal(ctx.rows.length,0)
  assert.equal(ctx.failed,true)
  assert.equal(ctx.loading,false)
  ctx.load()
  requests[2].done({samples:[{observedAt:1800,labels:{preview:'ERROR retry'}}],limited:false,limit:200})
  assert.equal(ctx.failed,false)
  assert.equal(ctx.rows.length,1)
})

test('summary pagination limits visible rows and disposal ignores in-flight queries', () => {
  const {view,ctx,requests} = setup()
  ctx.open()
  requests[0].done({samples:Array.from({length:12},(_,i)=>({observedAt:1800-i,labels:{preview:'ERROR '+i}})),limited:false,limit:200})
  ctx.page=2
  const rows=view.computed.pageRows.call(ctx)
  assert.equal(rows.length,5)
  assert.equal(rows[0].labels.preview,'ERROR 5')
  ctx.load()
  view.beforeDestroy.call(ctx)
  requests[1].done({samples:[{labels:{preview:'obsolete'}}]})
  assert.equal(ctx.rows.length,0)
})

test('the compiled Vue template renders log HTML as text and retains source metadata', () => {
  const Vue = require('vue')
  const compiler = require('vue-template-compiler')
  const source = fs.readFileSync(path.join(__dirname, '../src/views/admin-config/other/k8s-log-samples.vue'), 'utf8')
  const template = source.match(/<template>([\s\S]*?)<\/template>/)[1]
  const compiled = compiler.compileToFunctions(template)
  const instance = new Vue({
    ...component(),...compiled,
    propsData:{object:{clusterId:'cluster-a',namespace:'ns-a',podName:'app-0',container:'main'},range:{start:1000,end:1900}},
    beforeCreate() {this.$httpRequestEntrance={httpRequestEntrance(){}};this.apiCenter={};this.$t=key=>key}
  })
  instance.rows=[{observedAt:1800,labels:{file:'app.log',file_id:'device:inode',byte_offset:'42',pod_uid:'old-pod',rule:'error',rule_version:'1',preview:'ERROR <script>throw 1</script>'}}]
  const nodes=[]
  const visit = node => {if (!node) {return};nodes.push(node);for (const child of node.children || []) {visit(child)}}
  visit(instance._render())
  const preview=nodes.find(node=>node.tag==='pre')
  assert.equal(preview.children[0].text,'ERROR <script>throw 1</script>')
  assert.equal(preview.data && preview.data.domProps && preview.data.domProps.innerHTML,undefined)
  assert.ok(nodes.some(node=>node.text==='old-pod'))
  instance.$destroy()
})

test('copy includes the original byte offset and reports a rejected clipboard operation', () => {
  let input, copied, message, removed=0
  let allowed=true
  const view=component({document:{
    createElement(){input={style:{},select(){},remove(){removed++}};return input},
    body:{appendChild(){}},
    execCommand(){copied=input.value;return allowed}
  }})
  const ctx={$t:key=>key,$Message:{success:value=>{message=value},error:value=>{message=value}}}
  const row={observedAt:1800,labels:{file:'app.log',byte_offset:'9007199254740993',preview:'ERROR 中文'}}
  view.methods.copySample.call(ctx,row)
  assert.equal(JSON.parse(copied).labels.byte_offset,'9007199254740993')
  assert.equal(JSON.parse(copied).labels.preview,'ERROR 中文')
  assert.equal(message,'m_copied_to_clipboard')
  allowed=false
  view.methods.copySample.call(ctx,row)
  assert.equal(message,'m_k8s_log_sample_copy_failed')
  assert.equal(removed,2)
})
