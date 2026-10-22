<template>
  <div class="k8s-alert-rules">
    <Card>
      <p slot="title">{{$t('m_k8s_alert_rules')}}</p>
      <div class="scope-form">
        <Input v-model.trim="scope.clusterId" :placeholder="$t('m_k8s_cluster_id')" style="width:180px" />
        <Input v-model.trim="scope.namespace" :placeholder="$t('m_k8s_namespace')" style="width:180px;margin-left:8px" />
        <Button type="primary" style="margin-left:8px" @click="loadRules">{{$t('m_search')}}</Button>
        <Button type="success" style="margin-left:8px" @click="openCreate">{{$t('m_button_add')}}</Button>
      </div>
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
  { k8sAlertRules: 'post' },
  { k8sAlertRule: 'put' },
  { k8sAlertRule: 'delete' }
]

export default {
  data() {
    return {
      request: this.$root.$httpRequestEntrance.httpRequestEntrance,
      apiCenter: this.$root.apiCenter,
      scope: { clusterId: '', namespace: '' },
      rules: [],
      metrics: ['log_matches', 'log_health', 'cpu', 'memory', 'network_receive', 'network_transmit', 'block_read', 'block_write', 'cpu_throttled', 'cpu_throttled_periods', 'cpu_request', 'cpu_limit', 'memory_request', 'memory_limit', 'restart_count', 'ready', 'oom_events', 'memory_failcnt', 'container_ready', 'last_termination_oom'],
      operators: ['>', '>=', '<', '<=', '==', '!='],
      editor: { visible: false, editing: false, rule: {}, scope: {} }
    }
  },
  computed: {
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
  methods: {
    loadRules() {
      const scope = Object.assign({}, this.scope)
      const generation = this._rulesRequestGeneration = (this._rulesRequestGeneration || 0) + 1
      this.rules = []
      if (!this.scope.clusterId || !this.scope.namespace) {
        this.$Message.error(this.$t('m_k8s_scope_required'))
        return
      }
      this.request('GET', this.apiCenter.k8sAlertRules, scope, data => {
        if (this._rulesRequestGeneration === generation && this.scope.clusterId === scope.clusterId && this.scope.namespace === scope.namespace) this.rules = data || []
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
