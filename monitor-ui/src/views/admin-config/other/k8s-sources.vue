<template>
  <div class="k8s-sources">
    <Card class="management-card">
      <p slot="title">{{$t('m_k8s_sources')}}</p>
      <p class="page-description">{{$t('m_k8s_sources_intro')}}</p>
      <Tabs v-model="activeTab" class="management-tabs">
        <TabPane :label="$t('m_k8s_clusters')" name="clusters">
          <div class="section-toolbar">
            <div>
              <h3>{{$t('m_k8s_clusters')}}</h3>
              <p class="section-description">{{$t('m_k8s_clusters_help')}}</p>
            </div>
            <div class="toolbar-actions">
              <Button @click="refreshLists">{{$t('m_refresh')}}</Button>
              <Button type="primary" @click="openCreateCluster">{{$t('m_k8s_add_cluster')}}</Button>
            </div>
          </div>
          <Alert v-if="!clusters.length" type="info">{{$t('m_k8s_no_cluster')}}</Alert>
          <div class="table-wrap"><Table border :columns="clusterColumns" :data="clusters" /></div>
        </TabPane>

        <TabPane :label="$t('m_k8s_namespaces')" name="namespaces">
          <div class="section-toolbar">
            <div>
              <h3>{{$t('m_k8s_namespaces')}}</h3>
              <p class="section-description">{{$t('m_k8s_namespaces_help')}}</p>
            </div>
            <div class="toolbar-actions">
              <Select v-model="namespaceClusterFilter" clearable :placeholder="$t('m_k8s_cluster_id')" class="namespace-filter">
                <Option v-for="c in clusters" :key="c.clusterId" :value="c.clusterId">{{c.clusterId}}</Option>
              </Select>
              <Button type="primary" :disabled="!clusters.length" @click="openCreateNamespace">{{$t('m_k8s_add_namespace')}}</Button>
            </div>
          </div>
          <Alert v-if="!clusters.length" type="warning">{{$t('m_k8s_add_cluster_first')}}</Alert>
          <div class="table-wrap"><Table border :columns="namespaceColumns" :data="filteredNamespaces" /></div>
        </TabPane>

        <TabPane :label="$t('m_k8s_log_agent_management')" name="agents">
          <p class="section-description agent-description">{{$t('m_k8s_log_agent_auto_register')}}</p>
          <div class="agent-filters">
            <div class="filter-field">
              <label>{{$t('m_k8s_cluster_id')}}</label>
              <Select v-model="agent.clusterId" :placeholder="$t('m_k8s_cluster_id')" @on-change="loadAgents">
                <Option v-for="c in clusters" :key="c.clusterId" :value="c.clusterId">{{c.clusterId}}</Option>
              </Select>
            </div>
            <div class="filter-field node-filter">
              <label>{{$t('m_k8s_node')}}</label>
              <Input v-model.trim="agent.searchNode" clearable :placeholder="$t('m_k8s_node')" />
            </div>
            <div class="filter-actions">
              <Button @click="loadAgents">{{$t('m_refresh')}}</Button>
              <Button type="primary" :disabled="!agent.clusterId" @click="overrideAgent({clusterId: agent.clusterId, node: '', address: '', port: 9105})">{{$t('m_k8s_add_agent_address')}}</Button>
            </div>
          </div>
          <Alert v-if="!clusters.length" type="warning">{{$t('m_k8s_add_cluster_first')}}</Alert>
          <Alert v-else-if="!agents.length" type="info">{{$t('m_k8s_agent_empty')}}</Alert>
          <div class="table-wrap"><Table border size="small" :columns="agentColumns" :data="visibleAgents" /></div>
        </TabPane>
      </Tabs>
    </Card>

    <Modal v-model="cluster.visible" :title="cluster.editing ? $t('m_k8s_edit_cluster') : $t('m_k8s_add_cluster')" :mask-closable="false">
      <Form :label-width="135">
        <FormItem :label="$t('m_k8s_cluster_id')"><Input v-model.trim="cluster.clusterId" :disabled="cluster.editing" /></FormItem>
        <FormItem :label="$t('m_k8s_api_server')"><Input v-model.trim="cluster.apiServer" :placeholder="$t('m_k8s_api_server_placeholder')" /></FormItem>
        <FormItem :label="$t('m_k8s_ca')"><Input v-model="cluster.ca" type="textarea" :autosize="{minRows: 3, maxRows: 6}" /></FormItem>
        <FormItem :label="$t('m_k8s_token')"><Input v-model="cluster.token" type="password" autocomplete="new-password" /></FormItem>
        <FormItem v-if="cluster.editing" class="form-help"><Alert type="info">{{$t('m_k8s_cluster_token_edit_help')}}</Alert></FormItem>
        <FormItem v-else class="form-help"><Alert type="warning">{{$t('m_k8s_cluster_token_required')}}</Alert></FormItem>
        <FormItem :label="$t('m_enabled')"><Checkbox v-model="cluster.enabled" /></FormItem>
      </Form>
      <div slot="footer"><Button @click="cluster.visible = false">{{$t('m_button_cancel')}}</Button><Button type="primary" @click="saveCluster">{{$t('m_save')}}</Button></div>
    </Modal>

    <Modal v-model="namespace.visible" :title="namespace.editing ? $t('m_k8s_edit_namespace') : $t('m_k8s_add_namespace')" :mask-closable="false">
      <Form :label-width="135">
        <FormItem :label="$t('m_k8s_cluster_id')"><Select v-model="namespace.clusterId" :disabled="namespace.editing"><Option v-for="c in clusters" :key="c.clusterId" :value="c.clusterId">{{c.clusterId}}</Option></Select></FormItem>
        <FormItem :label="$t('m_k8s_namespace')"><Input v-model.trim="namespace.namespace" :disabled="namespace.editing" /></FormItem>
        <FormItem :label="$t('m_k8s_token_optional')"><Input v-model="namespace.token" type="password" autocomplete="new-password" /></FormItem>
        <FormItem class="form-help"><Alert type="info">{{$t('m_k8s_namespace_token_help')}}</Alert></FormItem>
        <FormItem :label="$t('m_enabled')"><Checkbox v-model="namespace.enabled" /></FormItem>
      </Form>
      <div slot="footer"><Button @click="namespace.visible = false">{{$t('m_button_cancel')}}</Button><Button type="primary" @click="saveNamespace">{{$t('m_save')}}</Button></div>
    </Modal>
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
      <div slot="footer"><Button @click="rulesEditor.visible = false">{{$t('m_button_cancel')}}</Button><Button type="primary" @click="saveLogRules">{{$t('m_save')}}</Button></div>
    </Modal>
    <Modal v-model="agent.editorVisible" :title="$t('m_k8s_log_targets')" width="850" :mask-closable="false">
      <div v-if="agent.loadedScope" class="editor-scope">{{agent.loadedScope.clusterId}} / {{agent.loadedScope.node}}</div>
      <Alert v-if="agent.defaultRules.length" type="info">{{$t('m_k8s_log_defaults_help')}}</Alert>
      <Form v-if="agent.loaded && agent.loadedScope && agent.loadedScope.clusterId === agent.clusterId && agent.loadedScope.node === agent.node">
        <FormItem><Input v-model="agent.targetsJson" type="textarea" :rows="14" /></FormItem>
      </Form>
      <div slot="footer" class="editor-actions">
        <Button type="error" @click="deleteAgent">{{$t('m_button_remove')}}</Button>
        <Button @click="editLogRules">{{$t('m_k8s_log_matching_rules')}}</Button>
        <Button :disabled="!agent.loaded || !agent.defaultRules.length" @click="resetDefaultLogRules">{{$t('m_k8s_reset_log_defaults')}}</Button>
        <Button @click="agent.editorVisible = false">{{$t('m_button_cancel')}}</Button>
        <Button type="primary" @click="saveAgent">{{$t('m_save')}}</Button>
      </div>
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
    return { request: this.$root.$httpRequestEntrance.httpRequestEntrance, apiCenter: this.$root.apiCenter, activeTab: 'clusters', namespaceClusterFilter: '',
      cluster: { clusterId: '', apiServer: '', ca: '', token: '', enabled: true, editing: false, visible: false }, namespace: { clusterId: '', namespace: '', token: '', enabled: true, editing: false, visible: false },
      agent: { clusterId: '', node: '', searchNode: '', loaded: false, loadedScope: null, targetsJson: '[]', defaultRules: [], editorVisible: false }, rulesEditor: {visible:false,targetIndex:0,targets:[]}, clusters: [], namespaces: [], agents: [], endpoint: {visible:false} }
  },
  computed: {
    editableLogTargets() { return this.rulesEditor.targets.map((target,index)=>({target,index})).filter(item=>!item.target.format) },
    filteredNamespaces() { return this.namespaceClusterFilter ? this.namespaces.filter(row => row.clusterId === this.namespaceClusterFilter) : this.namespaces },
    visibleAgents() {
      const node = (this.agent.searchNode || '').toLowerCase()
      return node ? this.agents.filter(row => String(row.node || '').toLowerCase()
        .includes(node)) : this.agents
    },
    clusterColumns() { return [
      { title: this.$t('m_k8s_cluster_id'), key: 'clusterId', minWidth: 140 }, { title: this.$t('m_k8s_api_server'), key: 'apiServer', minWidth: 280 }, { title: this.$t('m_k8s_token_configured'), key: 'hasToken', width: 150, render: (h,p) => h('Tag', { props: { color: p.row.hasToken ? 'success' : 'warning' } }, this.$t(p.row.hasToken ? 'm_yes' : 'm_no')) }, { title: this.$t('m_enabled'), key: 'enabled', width: 110, render: (h,p) => h('Tag', { props: { color: p.row.enabled ? 'success' : 'default' } }, this.$t(p.row.enabled ? 'm_yes' : 'm_no')) },
      { title: this.$t('m_actions'), key: 'actions', width: 160, render: (h, p) => h('div', { class: 'row-actions' }, [h('Button', { props: { size: 'small' }, on: { click: () => this.editCluster(p.row) } }, this.$t('m_button_edit')), h('Button', { props: { size: 'small', type: 'error' }, on: { click: () => this.removeCluster(p.row) } }, this.$t('m_button_remove'))]) }
    ] },
    namespaceColumns() { return [
      { title: this.$t('m_k8s_cluster_id'), key: 'clusterId', minWidth: 130 }, { title: this.$t('m_k8s_namespace'), key: 'namespace', minWidth: 160 }, { title: this.$t('m_enabled'), key: 'enabled', width: 100, render: (h,p) => h('Tag', { props: { color: p.row.enabled ? 'success' : 'default' } }, this.$t(p.row.enabled ? 'm_yes' : 'm_no')) }, { title: this.$t('m_k8s_state'), key: 'state', minWidth: 120 }, { title: this.$t('m_k8s_container_count'), key: 'containerCount', width: 150 }, { title: this.$t('m_k8s_observed_at'), key: 'observedAt', minWidth: 180, render: (h,p) => h('span', p.row.observedAt ? new Date(p.row.observedAt*1000).toLocaleString() : '-') },
      { title: this.$t('m_actions'), key: 'actions', width: 240, render: (h, p) => h('div', { class: 'row-actions' }, [h('Button',{props:{size:'small'},on:{click:()=>this.openObjects(p.row)}},this.$t('m_k8s_objects')), h('Button', { props: { size: 'small' }, on: { click: () => this.editNamespace(p.row) } }, this.$t('m_button_edit')), h('Button', { props: { size: 'small', type: 'error' }, on: { click: () => this.removeNamespace(p.row) } }, this.$t('m_button_remove'))]) }
    ] },
    agentColumns() { return [
      { title: this.$t('m_k8s_node'), key: 'node', minWidth: 140 },
      { title: this.$t('m_k8s_agent_address'), minWidth: 210, render: (h,p) => h('div', { class: 'endpoint-cell' }, [h('div', (p.row.address || '-') + (p.row.port ? ':' + p.row.port : '')), p.row.reportedAddress ? h('small', this.$t('m_k8s_reported') + ': ' + p.row.reportedAddress + ':' + p.row.reportedPort) : null]) },
      { title: this.$t('m_k8s_online'), minWidth: 140, render: (h,p) => h('div', { class: 'status-cell' }, [h('Tag', { props: { color: p.row.online ? 'success' : 'default' } }, this.$t(p.row.online ? 'm_yes' : 'm_no')), p.row.manualOverride ? h('Tag', { props: { color: 'warning' } }, this.$t('m_k8s_manual')) : null]) },
      { title: this.$t('m_k8s_last_seen'), key: 'lastSeen', minWidth: 175, render: (h,p) => h('span', p.row.lastSeen ? new Date(p.row.lastSeen*1000).toLocaleString() : '-') },
      { title: this.$t('m_k8s_target_count'), key: 'targetCount', width: 130 },
      { title: this.$t('m_actions'), key: 'actions', minWidth: 300, render: (h, p) => {
        const actions = [
          h('Button', { props: { size: 'small', type: 'primary' }, on: { click: () => this.selectAgent(p.row) } }, this.$t('m_k8s_log_targets')),
          h('Button', { props: { size: 'small' }, on: { click: () => this.overrideAgent(p.row) } }, this.$t('m_button_edit'))
        ]
        if (p.row.manualOverride) actions.push(h('Button', { props: { size: 'small' }, on: { click: () => this.releaseAgent(p.row) } }, this.$t('m_k8s_release_override')))
        actions.push(h('Button', { props: { size: 'small', type: 'error' }, on: { click: () => this.removeEndpoint(p.row) } }, this.$t('m_k8s_disable')))
        return h('div', { class: 'row-actions' }, actions)
      } }
    ] }
  },
  mounted() { this.refreshLists() },
  methods: {
    resetDefaultLogRules() {
      const scope = this.agent.loadedScope
      if (!this.agent.loaded || !scope || scope.clusterId !== this.agent.clusterId || scope.node !== this.agent.node || !this.agent.defaultRules.length) return
      let targets
      try {
        targets = JSON.parse(this.agent.targetsJson)
        if (!Array.isArray(targets)) throw new Error('invalid targets')
        targets.forEach(target=>{
          if (!target || typeof target !== 'object' || Array.isArray(target) || (target.rules != null && !Array.isArray(target.rules))) throw new Error('invalid target')
          if ((target.rules || []).some(rule=>!rule || typeof rule !== 'object' || Array.isArray(rule))) throw new Error('invalid rule')
        })
      } catch (e) { this.$Message.error(this.$t('m_k8s_invalid_json')); return }
      const logTargets = targets.filter(target=>!target.format)
      if (!logTargets.length) { this.$Message.error(this.$t('m_k8s_no_log_targets')); return }
      logTargets.forEach(target=>{
        target.rules = target.rules || []
        this.agent.defaultRules.forEach(rule=>{
          const index = target.rules.findIndex(current=>current.id === rule.id)
          const replacement = JSON.parse(JSON.stringify(rule))
          if (index < 0) target.rules.push(replacement)
          else target.rules.splice(index, 1, replacement)
        })
      })
      this.agent.targetsJson = JSON.stringify(targets,null,2)
      this.$Message.success(this.$t('m_k8s_log_defaults_reset'))
    },
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
    openObjects(row) { this.$router.push({path:'/k8sConfig/k8sObjects',query:{clusterId:row.clusterId,namespace:row.namespace}}) },
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
    openCreateCluster() { this.cluster = { clusterId: '', apiServer: '', ca: '', token: '', enabled: true, editing: false, visible: true } },
    openCreateNamespace() { this.namespace = { clusterId: this.clusters.length ? this.clusters[0].clusterId : '', namespace: '', token: '', enabled: true, editing: false, visible: true } },
    saveCluster() {
      if (!this.cluster.clusterId || !this.cluster.apiServer) { this.$Message.error(this.$t('m_k8s_cluster_fields_required')); return }
      if (!this.cluster.editing && !String(this.cluster.token || '').trim()) { this.$Message.error(this.$t('m_k8s_cluster_token_required')); return }
      const payload = { clusterId: this.cluster.clusterId, apiServer: this.cluster.apiServer, ca: this.cluster.ca, token: this.cluster.token, enabled: this.cluster.enabled }
      this.request('PUT', this.apiCenter.k8sClusterSource, payload, () => { this.cluster.visible = false; this.$Message.success(this.$t('m_save_success')); this.refreshLists() })
    },
    saveNamespace() {
      if (!this.namespace.clusterId || !this.namespace.namespace) { this.$Message.error(this.$t('m_k8s_scope_required')); return }
      const payload = { clusterId: this.namespace.clusterId, namespace: this.namespace.namespace, token: this.namespace.token, enabled: this.namespace.enabled }
      this.request('PUT', this.apiCenter.k8sNamespaceSource, payload, () => { this.namespace.visible = false; this.$Message.success(this.$t('m_save_success')); this.refreshLists() })
    },
    editCluster(row) { this.cluster = Object.assign({}, row, { token: '', editing:true, visible:true }) },
    editNamespace(row) { this.namespace = Object.assign({}, row, { token: '', editing:true, visible:true }) },
    removeCluster(row) { this.request('DELETE', this.apiCenter.k8sClusterSource, { clusterId: row.clusterId }, () => { this.$Message.success(this.$t('m_delete_success')); this.refreshLists() }) },
    removeNamespace(row) { this.request('DELETE', this.apiCenter.k8sNamespaceSource, { clusterId: row.clusterId, namespace: row.namespace }, () => { this.$Message.success(this.$t('m_delete_success')); this.refreshLists() }) },
    selectAgent(row) { this.agent.clusterId = row.clusterId; this.agent.node = row.node; this.loadAgent(true) },
    loadAgent(openEditor = false) {
      const scope = { clusterId: this.agent.clusterId, node: this.agent.node }
      const generation = this._agentRequestGeneration = (this._agentRequestGeneration || 0) + 1
      this.agent.loaded = false
      this.agent.defaultRules = []
      if (openEditor) this.agent.editorVisible = false
      if (!scope.clusterId || !scope.node) return
      this.request('GET', this.apiCenter.k8sLogAgentSource, scope, data => {
        if (this._agentRequestGeneration !== generation || this.agent.clusterId !== scope.clusterId || this.agent.node !== scope.node) return
        this.agent.targetsJson = JSON.stringify(data.targets || [], null, 2)
        this.agent.defaultRules = data.defaultRules || []
        this.agent.loadedScope = scope
        this.agent.loaded = true
        if (openEditor) this.agent.editorVisible = true
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
<style scoped>
.management-card { min-height: 420px; }
.page-description { margin: -4px 0 20px; color: #657180; }
.management-tabs >>> .ivu-tabs-bar { margin-bottom: 22px; }
.section-toolbar { display: flex; align-items: center; justify-content: space-between; gap: 16px; margin: 4px 0 18px; }
.section-toolbar h3 { margin: 0; color: #17233d; font-size: 16px; font-weight: 600; }
.section-description { margin: 5px 0 0; color: #808695; line-height: 1.5; }
.toolbar-actions, .filter-actions, .row-actions, .editor-actions { display: flex; flex-wrap: wrap; align-items: center; gap: 8px; }
.namespace-filter { width: 220px; }
.agent-description { margin: 0 0 16px; }
.agent-filters { display: flex; align-items: flex-end; gap: 16px; margin: 0 0 18px; padding: 16px; background: #f7f9fc; border: 1px solid #e8edf3; border-radius: 4px; }
.filter-field { width: 240px; }
.filter-field label { display: block; margin-bottom: 6px; color: #515a6e; font-size: 12px; }
.filter-actions { margin-left: auto; }
.table-wrap { width: 100%; overflow-x: auto; }
.endpoint-cell { line-height: 1.5; overflow-wrap: anywhere; }
.endpoint-cell small { color: #808695; }
.status-cell { white-space: nowrap; }
.editor-scope { margin-bottom: 12px; color: #657180; font-weight: 600; }
.form-help { margin-bottom: 12px; }
.form-help >>> .ivu-form-item-content { margin-left: 0 !important; }
@media (max-width: 768px) {
  .section-toolbar { align-items: stretch; flex-direction: column; }
  .toolbar-actions { width: 100%; align-items: stretch; flex-direction: column; }
  .toolbar-actions > *, .namespace-filter, .filter-field { width: 100%; }
  .agent-filters { align-items: stretch; flex-direction: column; }
  .filter-actions { margin-left: 0; }
  .filter-actions > * { flex: 1; }
}
</style>
