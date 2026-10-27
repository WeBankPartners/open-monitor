<template>
  <div class="k8s-alert-rules">
    <Card>
      <p slot="title">{{$t('m_k8s_alert_rules')}}</p>
      <Alert v-if="alertEvaluationEnabled === false" type="warning">{{$t('m_k8s_alert_rules_paused')}}</Alert>
      <Alert v-if="catalogFailed" type="error">{{$t('m_k8s_catalog_failed')}} <Button size="small" @click="loadCatalog">{{$t('m_refresh')}}</Button></Alert>
      <div class="scope-form">
        <div class="scope-field">
          <label>{{$t('m_k8s_cluster_id')}}</label>
          <Select v-model="scope.clusterId" filterable :loading="sourcesLoading" :disabled="sourcesLoading" @on-change="clusterChanged()">
            <Option v-for="cluster in clusters" :key="cluster.clusterId" :value="cluster.clusterId">{{cluster.clusterId}}</Option>
          </Select>
        </div>
        <div class="scope-field">
          <label>{{$t('m_k8s_namespace')}}</label>
          <Select v-model="scope.namespace" filterable :disabled="sourcesLoading || !scope.clusterId" @on-change="loadRules">
            <Option v-for="namespace in namespaceOptions" :key="namespace.namespace" :value="namespace.namespace">{{namespace.namespace}}</Option>
          </Select>
        </div>
        <Button type="primary" :disabled="sourcesLoading || !scope.namespace" @click="loadRules">{{$t('m_search')}}</Button>
        <Button type="success" :disabled="sourcesLoading || !scope.namespace" @click="openCreate">{{$t('m_button_add')}}</Button>
      </div>
      <Alert v-if="sourcesFailed" type="error" style="margin-top:12px">{{$t('m_k8s_sources_failed')}} <Button size="small" @click="loadSources">{{$t('m_refresh')}}</Button></Alert>
      <Alert v-else-if="!sourcesLoading && !scope.namespace" style="margin-top:12px">{{$t('m_k8s_no_namespace')}}</Alert>
      <Table border :columns="columns" :data="rules" style="margin-top:16px" />
    </Card>
    <Modal v-model="editor.visible" :title="editor.editing ? $t('m_k8s_edit_alert_rule') : $t('m_k8s_add_alert_rule')" :mask-closable="false">
      <Form :label-width="130">
        <FormItem :label="$t('m_k8s_rule_id')"><Input v-model.trim="editor.rule.ruleId" :disabled="editor.editing" /></FormItem>
        <FormItem :label="$t('m_k8s_metric')"><Select v-model="editor.rule.metric"><Option v-for="metric in metrics" :key="metric" :value="metric">{{metric}}</Option></Select></FormItem>
        <FormItem :label="$t('m_k8s_operator')"><Select v-model="editor.rule.operator"><Option v-for="operator in operators" :key="operator" :value="operator">{{operator}}</Option></Select></FormItem>
        <FormItem :label="$t('m_k8s_threshold')"><InputNumber v-model="editor.rule.threshold" :step="0.1" style="width:100%" /></FormItem>
        <FormItem :label="$t('m_k8s_workload_type')"><Input v-model.trim="editor.rule.workloadType" /></FormItem>
        <FormItem :label="$t('m_k8s_workload_name')"><Input v-model.trim="editor.rule.workloadName" /></FormItem>
        <FormItem :label="$t('m_k8s_pod_name')"><Input v-model.trim="editor.rule.podName" /></FormItem>
        <FormItem :label="$t('m_k8s_container')"><Input v-model.trim="editor.rule.container" /></FormItem>
        <FormItem :label="$t('m_k8s_for_seconds')"><InputNumber v-model="editor.rule.forSeconds" :min="0" style="width:100%" /></FormItem>
        <FormItem :label="$t('m_k8s_recover_after')"><InputNumber v-model="editor.rule.recoverAfter" :min="0" style="width:100%" /></FormItem>
        <FormItem :label="$t('m_k8s_notify_strategy')"><InputNumber v-model="editor.rule.notifyStrategyId" :min="0" style="width:100%" /></FormItem>
        <FormItem :label="$t('m_enabled')"><Checkbox v-model="editor.rule.enabled" /></FormItem>
      </Form>
      <div slot="footer"><Button @click="editor.visible=false">{{$t('m_button_cancel')}}</Button><Button type="primary" @click="saveRule">{{$t('m_save')}}</Button></div>
    </Modal>
  </div>
</template>

<script>
export const custom_api_enum = [
  { k8sAlertStates: 'get' },
  { k8sAlertRules: 'post' },
  { k8sAlertRule: 'put' },
  { k8sAlertRule: 'delete' }
]

export default {
  data() {
    return {
      request: this.$root.$httpRequestEntrance.httpRequestEntrance,
      apiCenter: this.$root.apiCenter,
      clusters: [],
      namespaces: [],
      sourcesLoading: false,
      sourcesFailed: false,
      catalogFailed: false,
      alertEvaluationEnabled: null,
      scope: { clusterId: '', namespace: '' },
      rules: [],
      metrics: ['log_matches', 'log_health', 'cpu', 'memory', 'network_receive', 'network_transmit', 'block_read', 'block_write', 'cpu_throttled', 'cpu_throttled_periods', 'cpu_request', 'cpu_limit', 'memory_request', 'memory_limit', 'restart_count', 'ready', 'oom_events', 'memory_failcnt', 'container_ready', 'last_termination_oom'],
      operators: ['>', '>=', '<', '<=', '==', '!='],
      editor: { visible: false, editing: false, rule: {}, scope: {} }
    }
  },
  computed: {
    namespaceOptions() { return this.namespaces.filter(row => row.clusterId === this.scope.clusterId && row.enabled) },
    columns() {
      return [
        { title: this.$t('m_k8s_rule_id'), key: 'ruleId' },
        { title: this.$t('m_k8s_metric'), key: 'metric' },
        { title: this.$t('m_k8s_operator'), key: 'operator' },
        { title: this.$t('m_k8s_threshold'), key: 'threshold' },
        { title: this.$t('m_k8s_version'), key: 'version' },
        { title: this.$t('m_enabled'), key: 'enabled', render: (h, params) => h('span', {}, params.row.enabled ? this.$t('m_yes') : this.$t('m_no')) },
        { title: this.$t('m_actions'), key: 'actions', width: 150, render: (h, params) => h('div', [
          h('Button', { props: { size: 'small' }, on: { click: () => this.openEdit(params.row) } }, this.$t('m_button_edit')),
          h('Button', { props: { size: 'small', type: 'error' }, style: { marginLeft: '4px' }, on: { click: () => this.deleteRule(params.row) } }, this.$t('m_button_remove'))
        ]) }
      ]
    }
  },
  mounted() { this.loadSources(); this.loadCatalog() },
  beforeDestroy() {
    this._destroyed = true
    clearTimeout(this._sourcesDeadline)
    clearTimeout(this._catalogDeadline)
  },
  methods: {
    loadCatalog() {
      const generation = this._catalogGeneration = (this._catalogGeneration || 0) + 1
      clearTimeout(this._catalogDeadline)
      this.catalogFailed = false
      let finished = false
      const finish = (data, failed) => {
        if (this._destroyed || this._catalogGeneration !== generation || finished) {return}
        finished = true
        clearTimeout(this._catalogDeadline)
        this.catalogFailed = failed
        this.alertEvaluationEnabled = !failed && typeof data.alertEvaluationEnabled === 'boolean' ? data.alertEvaluationEnabled : null
      }
      this._catalogDeadline = setTimeout(() => finish(null,true),30000)
      this.request('GET', this.apiCenter.k8sMetricCatalog, {}, data => finish(data,false), {isNeedloading: false}, () => finish(null,true))
    },
    loadSources() {
      const generation = this._sourcesGeneration = (this._sourcesGeneration || 0) + 1
      clearTimeout(this._sourcesDeadline)
      this.sourcesLoading = true
      this.sourcesFailed = false
      this.rules = []
      this._rulesRequestGeneration = (this._rulesRequestGeneration || 0) + 1
      let finished = false
      const current = () => !this._destroyed && this._sourcesGeneration === generation && !finished
      const fail = () => {
        if (!current()) {return}
        finished = true
        clearTimeout(this._sourcesDeadline)
        this.sourcesLoading = false
        this.sourcesFailed = true
        this.clusters = []
        this.namespaces = []
        this.scope.clusterId = ''
        this.scope.namespace = ''
      }
      this._sourcesDeadline = setTimeout(fail, 30000)
      this.request('POST', this.apiCenter.k8sClusterList, {}, clusters => {
        if (!current()) {return}
        this.request('POST', this.apiCenter.k8sNamespaceList, {}, namespaces => {
          if (!current()) {return}
          finished = true
          clearTimeout(this._sourcesDeadline)
          this.sourcesLoading = false
          this.clusters = (clusters || []).filter(row => row.enabled)
          this.namespaces = namespaces || []
          this.scope.clusterId = this.clusters.some(row => row.clusterId === this.scope.clusterId) ? this.scope.clusterId : ((this.clusters[0] || {}).clusterId || '')
          this.clusterChanged(this.scope.namespace)
        }, {isNeedloading: false}, fail)
      }, {isNeedloading: false}, fail)
    },
    clusterChanged(namespace) {
      this.scope.namespace = this.namespaceOptions.some(row => row.namespace === namespace) ? namespace : ((this.namespaceOptions[0] || {}).namespace || '')
      this.loadRules()
    },
    loadRules() {
      const scope = Object.assign({}, this.scope)
      const generation = this._rulesRequestGeneration = (this._rulesRequestGeneration || 0) + 1
      this.rules = []
      if (this.sourcesLoading || !scope.clusterId || !scope.namespace) {return}
      this.request('GET', this.apiCenter.k8sAlertRules, scope, data => {
        if (!this._destroyed && this._rulesRequestGeneration === generation && this.scope.clusterId === scope.clusterId && this.scope.namespace === scope.namespace) {
          this.rules = data || []
        }
      })
    },
    emptyRule() {
      return { ruleId: '', metric: 'cpu', operator: '>', threshold: 0, workloadType: '', workloadName: '', podName: '', container: '', forSeconds: 0, recoverAfter: 3, notifyStrategyId: 0, enabled: true, version: 0 }
    },
    openCreate() {
      if (!this.scope.clusterId || !this.scope.namespace) { this.$Message.error(this.$t('m_k8s_scope_required')); return }
      this.editor = { visible: true, editing: false, rule: this.emptyRule(), scope: Object.assign({}, this.scope) }
    },
    openEdit(row) { this.editor = { visible: true, editing: true, rule: Object.assign({}, row), scope: {clusterId:row.clusterId, namespace:row.namespace} } },
    saveRule() {
      const body = Object.assign({}, this.editor.rule, this.editor.scope)
      const method = this.editor.editing ? 'PUT' : 'POST'
      const url = this.editor.editing ? this.apiCenter.k8sAlertRule.replace(':ruleId', encodeURIComponent(body.ruleId)) : this.apiCenter.k8sAlertRules
      this.request(method, url, body, () => { this.editor.visible = false; this.loadRules() })
    },
    deleteRule(row) {
      const scope = { clusterId: row.clusterId, namespace: row.namespace, version: row.version }
      this.$Modal.confirm({ title: this.$t('m_delConfirm_title'), content: this.$t('m_k8s_delete_alert_rule'), onOk: () => {
        const url = this.apiCenter.k8sAlertRule.replace(':ruleId', encodeURIComponent(row.ruleId))
        this.request('DELETE', url, scope, () => this.loadRules())
      } })
    }
  }
}
</script>
<style scoped>
.scope-form {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 12px;
}
.scope-field {
  display: flex;
  align-items: center;
  gap: 8px;
}
.scope-field label { white-space: nowrap; }
.scope-field .ivu-select { width: 220px; }
@media (max-width: 600px) {
  .scope-field { width: 100%; }
  .scope-field .ivu-select { flex: 1; min-width: 0; }
}
</style>
