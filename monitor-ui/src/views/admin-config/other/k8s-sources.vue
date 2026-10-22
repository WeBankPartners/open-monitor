<template>
  <div class="k8s-sources">
    <Card>
      <p slot="title">{{$t('m_k8s_sources')}}</p>
      <Button @click="refreshLists">{{$t('m_refresh')}}</Button>
      <Alert>{{$t('m_k8s_token_help')}}</Alert>
      <Form inline>
        <FormItem :label="$t('m_k8s_cluster_id')"><Input v-model.trim="cluster.clusterId" :disabled="cluster.editing" /></FormItem>
        <FormItem :label="$t('m_k8s_api_server')"><Input v-model.trim="cluster.apiServer" style="width:260px" /></FormItem>
        <FormItem :label="$t('m_k8s_ca')"><Input v-model="cluster.ca" type="textarea" style="width:260px" /></FormItem>
        <FormItem :label="$t('m_k8s_token')"><Input v-model="cluster.token" type="password" /></FormItem>
        <FormItem :label="$t('m_enabled')"><Checkbox v-model="cluster.enabled" /></FormItem>
        <FormItem><Button @click="cluster = {clusterId: '', apiServer: '', ca: '', token: '', enabled: true, editing: false}">{{$t('m_button_add')}}</Button></FormItem><FormItem><Button type="primary" @click="saveCluster">{{$t('m_save')}}</Button></FormItem>
      </Form>
      <Table border :columns="clusterColumns" :data="clusters" style="margin-top:16px" />
      <Divider />
      <Form inline>
        <FormItem :label="$t('m_k8s_cluster_id')"><Select v-model="namespace.clusterId" :disabled="namespace.editing" style="width:180px"><Option v-for="c in clusters" :key="c.clusterId" :value="c.clusterId">{{c.clusterId}}</Option></Select></FormItem>
        <FormItem :label="$t('m_k8s_namespace')"><Input v-model.trim="namespace.namespace" :disabled="namespace.editing" /></FormItem>
        <FormItem :label="$t('m_k8s_token_optional')"><Input v-model="namespace.token" type="password" /></FormItem>
        <FormItem :label="$t('m_enabled')"><Checkbox v-model="namespace.enabled" /></FormItem>
        <FormItem><Button @click="namespace = {clusterId: '', namespace: '', token: '', enabled: true, editing: false}">{{$t('m_button_add')}}</Button></FormItem><FormItem><Button type="primary" @click="saveNamespace">{{$t('m_save')}}</Button></FormItem>
        <FormItem><Button type="error" @click="deleteNamespace">{{$t('m_button_remove')}}</Button></FormItem>
      </Form>
      <Table border :columns="namespaceColumns" :data="namespaces" style="margin-top:16px" />
      <Divider />
      <Form inline>
        <FormItem :label="$t('m_k8s_cluster_id')"><Select v-model="agent.clusterId" style="width:180px" @on-change="loadAgents"><Option v-for="c in clusters" :key="c.clusterId" :value="c.clusterId">{{c.clusterId}}</Option></Select></FormItem>
        <FormItem :label="$t('m_k8s_node')"><Input v-model.trim="agent.node" /></FormItem>
        <FormItem><Button @click="loadAgents">{{$t('m_refresh')}}</Button><Button @click="overrideAgent({clusterId:agent.clusterId,node:'',address:'',port:9105})">{{$t('m_button_add')}}</Button></FormItem><FormItem><Button @click="loadAgent">{{$t('m_search')}}</Button></FormItem>
      </Form>
      <Form v-if="agent.loaded && agent.loadedScope && agent.loadedScope.clusterId === agent.clusterId && agent.loadedScope.node === agent.node">
        <FormItem :label="$t('m_k8s_log_targets')"><Input v-model="agent.targetsJson" type="textarea" :rows="8" /></FormItem>
        <Button type="primary" @click="saveAgent">{{$t('m_save')}}</Button>
        <Button style="margin-left:8px" @click="editLogRules">{{$t('m_k8s_log_matching_rules')}}</Button>
        <Button type="error" style="margin-left:8px" @click="deleteAgent">{{$t('m_button_remove')}}</Button>
      </Form>
      <Table border :columns="agentColumns" :data="agents" style="margin-top:16px" />
      <Alert>{{$t('m_k8s_log_agent_auto_register')}}</Alert>
    </Card>
    <Modal v-model="rulesEditor.visible" :title="$t('m_k8s_log_matching_rules')" width="850" :mask-closable="false">
      <Alert>{{$t('m_k8s_log_exclude_help')}}</Alert>
      <Select v-model="rulesEditor.targetIndex" style="width:100%">
        <Option v-for="item in editableLogTargets" :key="item.index" :value="item.index">{{item.target.namespace}} / {{item.target.podName || item.target.id}} / {{item.target.container}}</Option>
      </Select>
      <div v-if="rulesEditor.targets[rulesEditor.targetIndex]" style="max-height:55vh;overflow-y:auto;margin-top:12px">
        <Card v-for="(rule, ruleIndex) in rulesEditor.targets[rulesEditor.targetIndex].rules" :key="ruleIndex" style="margin-bottom:12px">
          <Form :label-width="120">
            <FormItem :label="$t('m_k8s_rule_id')"><Input v-model.trim="rule.id" /></FormItem>
            <FormItem :label="$t('m_k8s_match_mode')"><Select v-model="rule.mode"><Option value="literal">{{$t('m_k8s_literal')}}</Option><Option value="regex">{{$t('m_k8s_regex')}}</Option></Select></FormItem>
            <FormItem :label="$t('m_k8s_match_pattern')"><Input v-model="rule.pattern" /></FormItem>
            <FormItem :label="$t('m_k8s_exclusions')">
              <div v-for="(exclude, index) in rule.exclude" :key="index" style="display:flex;gap:8px;margin-bottom:8px">
                <Select v-model="exclude.mode" style="width:130px"><Option value="literal">{{$t('m_k8s_literal')}}</Option><Option value="regex">{{$t('m_k8s_regex')}}</Option></Select>
                <Input v-model="exclude.pattern" />
                <Button @click="rule.exclude.splice(index, 1)">{{$t('m_button_remove')}}</Button>
              </div>
              <Button @click="addLogExclusion(rule)">{{$t('m_k8s_add_exclusion')}}</Button>
            </FormItem>
          </Form>
          <Button type="error" @click="rulesEditor.targets[rulesEditor.targetIndex].rules.splice(ruleIndex, 1)">{{$t('m_button_remove')}}</Button>
        </Card>
        <Button @click="addLogRule">{{$t('m_button_add')}}</Button>
      </div>
      <div slot="footer"><Button @click="rulesEditor.visible=false">{{$t('m_button_cancel')}}</Button><Button type="primary" @click="saveLogRules">{{$t('m_save')}}</Button></div>
    </Modal>
    <Modal v-model="endpoint.visible" :title="$t('m_k8s_agent_address')" :footer-hide="true">
      <Form :label-width="100">
        <FormItem :label="$t('m_k8s_cluster_id')"><Input v-model="endpoint.clusterId" :disabled="endpoint.editing" /></FormItem>
        <FormItem :label="$t('m_k8s_node')"><Input v-model="endpoint.node" :disabled="endpoint.editing" /></FormItem>
        <FormItem :label="$t('m_k8s_agent_address')"><Input v-model.trim="endpoint.address" /></FormItem>
        <FormItem :label="$t('m_k8s_agent_port')"><InputNumber v-model="endpoint.port" :min="1" :max="65535" /></FormItem>
        <Button type="primary" @click="saveEndpoint">{{$t('m_save')}}</Button>
      </Form>
    </Modal>
  </div>
</template>
<script>
export default {
  data() {
    return { request: this.$root.$httpRequestEntrance.httpRequestEntrance, apiCenter: this.$root.apiCenter,
      cluster: { guid: '', clusterId: '', apiServer: '', ca: '', token: '', enabled: true }, namespace: { guid: '', clusterId: '', namespace: '', token: '', enabled: true },
      agent: { clusterId: '', node: '', loaded: false, loadedScope: null, targetsJson: '[]' }, rulesEditor: {visible:false,targetIndex:0,targets:[]}, clusters: [], namespaces: [], agents: [], endpoint: {visible:false} }
  },
  computed: {
    editableLogTargets() { return this.rulesEditor.targets.map((target,index)=>({target,index})).filter(item=>!item.target.format) },
    clusterColumns() { return [
      { title: this.$t('m_k8s_cluster_id'), key: 'clusterId' }, { title: this.$t('m_k8s_api_server'), key: 'apiServer' }, { title: this.$t('m_enabled'), key: 'enabled', render: (h,p) => h('span',this.$t(p.row.enabled ? 'm_yes' : 'm_no')) },
      { title: this.$t('m_actions'), key: 'actions', render: (h, p) => h('div', [h('Button', { props: { size: 'small' }, on: { click: () => this.editCluster(p.row) } }, this.$t('m_button_edit')), h('Button', { props: { size: 'small', type: 'error' }, style: { marginLeft: '4px' }, on: { click: () => this.removeCluster(p.row) } }, this.$t('m_button_remove'))]) }
    ] },
    namespaceColumns() { return [
      { title: this.$t('m_k8s_cluster_id'), key: 'clusterId' }, { title: this.$t('m_k8s_namespace'), key: 'namespace' }, { title: this.$t('m_k8s_state'), key: 'state' }, { title: this.$t('m_k8s_container_count'), key: 'containerCount' }, { title: this.$t('m_k8s_observed_at'), key: 'observedAt', render: (h,p) => h('span', p.row.observedAt ? new Date(p.row.observedAt*1000).toLocaleString() : '-') },
      { title: this.$t('m_actions'), key: 'actions', render: (h, p) => h('div', [h('Button',{props:{size:'small'},on:{click:()=>this.openObjects(p.row)}},this.$t('m_k8s_objects')), h('Button', { props: { size: 'small' }, on: { click: () => this.editNamespace(p.row) } }, this.$t('m_button_edit')), h('Button', { props: { size: 'small', type: 'error' }, style: { marginLeft: '4px' }, on: { click: () => this.removeNamespace(p.row) } }, this.$t('m_button_remove'))]) }
    ] },
    agentColumns() { return [
      { title: this.$t('m_k8s_cluster_id'), key: 'clusterId' }, { title: this.$t('m_k8s_node'), key: 'node' }, { title: this.$t('m_k8s_agent_address'), key: 'address' }, { title: this.$t('m_k8s_agent_port'), key: 'port' }, { title: this.$t('m_k8s_last_seen'), key: 'lastSeen', render: (h,p) => h('span', p.row.lastSeen ? new Date(p.row.lastSeen*1000).toLocaleString() : '-') }, { title: this.$t('m_k8s_reported'), key: 'reportedAddress', render: (h,p) => h('span', p.row.reportedAddress ? p.row.reportedAddress+':'+p.row.reportedPort : '-') }, { title: this.$t('m_k8s_online'), key: 'online', render: (h,p) => h('span',this.$t(p.row.online ? 'm_yes' : 'm_no')) }, { title: this.$t('m_k8s_target_count'), key: 'targetCount' }, { title: this.$t('m_k8s_manual'), key: 'manualOverride' },
      { title: this.$t('m_actions'), key: 'actions', render: (h, p) => h('div', [h('Button', { props: { size: 'small' }, on: { click: () => this.overrideAgent(p.row) } }, this.$t('m_button_edit')), h('Button', { props: { size: 'small' }, style: { marginLeft: '4px' }, on: { click: () => this.releaseAgent(p.row) } }, this.$t('m_k8s_release_override')), h('Button', {props:{size:'small',type:'error'},on:{click:()=>this.removeEndpoint(p.row)}},this.$t('m_k8s_disable')), h('Button',{props:{size:'small'},on:{click:()=>this.selectAgent(p.row)}},this.$t('m_k8s_log_targets'))]) }
    ] }
  },
  mounted() { this.refreshLists() },
  methods: {
    editLogRules() {
      let targets
      try {
        targets = JSON.parse(this.agent.targetsJson)
        if (!Array.isArray(targets)) throw new Error('invalid targets')
        targets.forEach(target=>{
          if (!target || typeof target !== 'object' || (target.rules != null && !Array.isArray(target.rules))) throw new Error('invalid target')
          target.rules=target.rules || []
          target.rules.forEach(rule=>{
            if (!rule || typeof rule !== 'object' || (rule.exclude != null && !Array.isArray(rule.exclude))) throw new Error('invalid rule')
            rule.exclude=rule.exclude || []
            if (rule.exclude.some(exclude=>!exclude || typeof exclude !== 'object')) throw new Error('invalid exclusion')
          })
        })
      } catch (e) { this.$Message.error(this.$t('m_k8s_invalid_json')); return }
      const first = targets.findIndex(target=>!target.format)
      if (first < 0) { this.$Message.error(this.$t('m_k8s_no_log_targets')); return }
      this.rulesEditor={visible:true,targetIndex:first,targets}
    },
    addLogRule() { this.rulesEditor.targets[this.rulesEditor.targetIndex].rules.push({id:'',version:1,mode:'literal',pattern:'',exclude:[]}) },
    addLogExclusion(rule) { rule.exclude.push({mode:'literal',pattern:''}) },
    saveLogRules() {
      const valid = this.rulesEditor.targets.every(target=>{
        const ids = new Set()
        return (target.rules || []).every(rule=>{
          if (!rule.id || ids.has(rule.id) || !rule.pattern || !['literal','regex'].includes(rule.mode)) return false
          ids.add(rule.id)
          return (rule.exclude || []).every(exclude=>exclude.pattern && ['literal','regex'].includes(exclude.mode))
        })
      })
      if (!valid) { this.$Message.error(this.$t('m_k8s_invalid_log_rule')); return }
      this.agent.targetsJson=JSON.stringify(this.rulesEditor.targets,null,2)
      this.saveAgent(()=>{this.rulesEditor.visible=false})
    },
    openObjects(row) { this.$router.push({path:'/adminConfig/k8sObjects',query:{clusterId:row.clusterId,namespace:row.namespace}}) },
    refreshLists() { this.request('POST', this.apiCenter.k8sClusterList, {}, data => { this.clusters = data || []; this.loadAgents() }); this.request('POST', this.apiCenter.k8sNamespaceList, {}, data => { this.namespaces = data || [] }) },
    loadAgents() {
      if (!this.agent.clusterId && this.clusters.length) this.agent.clusterId = this.clusters[0].clusterId
      const clusterId = this.agent.clusterId
      const generation = this._agentsRequestGeneration = (this._agentsRequestGeneration || 0) + 1
      this.agents = []
      if (clusterId) this.request('GET', this.apiCenter.k8sLogAgentList, { clusterId }, data => {
        if (this._agentsRequestGeneration === generation && this.agent.clusterId === clusterId) this.agents = data.agents || []
      })
    },
    saveCluster() { this.request('PUT', this.apiCenter.k8sClusterSource, this.cluster, () => { this.$Message.success(this.$t('m_save_success')); this.refreshLists() }) },
    saveNamespace() { this.request('PUT', this.apiCenter.k8sNamespaceSource, this.namespace, () => { this.$Message.success(this.$t('m_save_success')); this.refreshLists() }) },
    deleteNamespace() { this.removeNamespace(this.namespace) },
    editCluster(row) { this.cluster = Object.assign({}, row, { token: '', editing:true }) },
    editNamespace(row) { this.namespace = Object.assign({}, row, { token: '', editing:true }) },
    removeCluster(row) { this.request('DELETE', this.apiCenter.k8sClusterSource, { clusterId: row.clusterId }, () => { this.$Message.success(this.$t('m_delete_success')); this.refreshLists() }) },
    removeNamespace(row) { this.request('DELETE', this.apiCenter.k8sNamespaceSource, { clusterId: row.clusterId, namespace: row.namespace }, () => { this.$Message.success(this.$t('m_delete_success')); this.refreshLists() }) },
    selectAgent(row) { this.agent.clusterId = row.clusterId; this.agent.node = row.node; this.loadAgent() },
    loadAgent() {
      const scope = { clusterId: this.agent.clusterId, node: this.agent.node }
      const generation = this._agentRequestGeneration = (this._agentRequestGeneration || 0) + 1
      this.agent.loaded = false
      if (!scope.clusterId || !scope.node) return
      this.request('GET', this.apiCenter.k8sLogAgentSource, scope, data => {
        if (this._agentRequestGeneration !== generation || this.agent.clusterId !== scope.clusterId || this.agent.node !== scope.node) return
        this.agent.targetsJson = JSON.stringify(data.targets || [], null, 2)
        this.agent.loadedScope = scope
        this.agent.loaded = true
      })
    },
    saveAgent(done) {
      const scope = this.agent.loadedScope
      if (!this.agent.loaded || !scope || scope.clusterId !== this.agent.clusterId || scope.node !== this.agent.node) return
      let targets
      try { targets = JSON.parse(this.agent.targetsJson); if (!Array.isArray(targets)) throw new Error('invalid targets') } catch (e) { this.$Message.error(this.$t('m_k8s_invalid_json')); return }
      this.request('PUT', this.apiCenter.k8sLogAgentSource, { ...scope, targets }, () => {
        this.$Message.success(this.$t('m_save_success'))
        if (this.agent.clusterId === scope.clusterId && this.agent.node === scope.node) { this.loadAgent(); if (typeof done === 'function') done() }
      })
    },
    deleteAgent() {
      const scope = this.agent.loadedScope
      if (!this.agent.loaded || !scope || scope.clusterId !== this.agent.clusterId || scope.node !== this.agent.node) return
      this.request('DELETE', this.apiCenter.k8sLogAgentSource, scope, () => {
        if (this.agent.clusterId === scope.clusterId && this.agent.node === scope.node) this.loadAgent()
        this.$Message.success(this.$t('m_delete_success'))
      })
    },
    overrideAgent(row) { this.endpoint = Object.assign({}, row, { visible:true, editing:!!row.node }) },
    saveEndpoint() {
      if (!this.endpoint.clusterId || !this.endpoint.node || !this.endpoint.address || !Number.isInteger(this.endpoint.port)) { this.$Message.error(this.$t('m_k8s_scope_required')); return }
      const payload = Object.assign({}, this.endpoint, { manualOverride: true })
      this.request('PUT', this.apiCenter.k8sLogAgentEndpoint, payload, () => { this.endpoint.visible=false; this.loadAgents() })
    },
    removeEndpoint(row) { this.$Modal.confirm({title:this.$t('m_k8s_disable'),content:this.$t('m_k8s_disable_help'),onOk:()=>this.request('DELETE',this.apiCenter.k8sLogAgentEndpoint,{clusterId:row.clusterId,node:row.node},()=>this.loadAgents())}) },
    releaseAgent(row) { this.request('DELETE', this.apiCenter.k8sLogAgentReleaseOverride, { clusterId: row.clusterId, node: row.node }, () => { this.$Message.success(this.$t('m_save_success')); this.loadAgents() }) }
  }
}
</script>
