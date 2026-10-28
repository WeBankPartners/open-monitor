<template>
  <div class="k8s-objects">
    <Card>
      <p slot="title">{{$t('m_k8s_objects')}}</p>
      <Button v-if="selected" slot="extra" size="small" @click="inventoryCollapsed = !inventoryCollapsed">{{$t(inventoryCollapsed ? 'm_k8s_switch_object' : 'm_k8s_collapse_inventory')}}</Button>
      <Alert v-if="localCollectionEnabled === false" type="warning">{{$t('m_k8s_local_collection_disabled')}}</Alert>
      <div v-show="!inventoryCollapsed || !selected" class="object-filters">
        <div class="object-filter filter-scope">
          <label>{{$t('m_k8s_cluster_id')}}</label>
          <Select v-model="scope.clusterId" filterable :loading="sourcesLoading" :disabled="sourcesLoading" @on-change="clusterChanged()"><Option v-for="cluster in clusters" :key="cluster.clusterId" :value="cluster.clusterId">{{cluster.clusterId}}</Option></Select>
        </div>
        <div class="object-filter filter-scope">
          <label>{{$t('m_k8s_namespace')}}</label>
          <Select v-model="scope.namespace" filterable :disabled="sourcesLoading || !scope.clusterId" @on-change="namespaceChanged"><Option v-for="namespace in namespaceOptions" :key="namespace.namespace" :value="namespace.namespace">{{namespace.namespace}}</Option></Select>
        </div>
        <div class="object-filter filter-workload">
          <label>{{$t('m_k8s_workload_name')}}</label>
          <Select v-model="filters.workload" clearable filterable @on-change="filterChanged"><Option v-for="name in workloads" :key="name" :value="name">{{name}}</Option></Select>
        </div>
        <div class="object-filter filter-text">
          <label>{{$t('m_k8s_pod_name')}}</label>
          <Input v-model="filters.pod" clearable @on-change="filterChanged" />
        </div>
        <div class="object-filter filter-text">
          <label>{{$t('m_k8s_instance_id')}}</label>
          <Input v-model="filters.instanceId" clearable @on-change="filterChanged" />
        </div>
        <div class="object-filter filter-actions">
          <Button @click="refresh">{{$t('m_refresh')}}</Button>
        </div>
        <div class="object-filter filter-auto">
          <Checkbox v-model="filters.onlyAbnormal" @on-change="filterChanged">{{$t('m_k8s_only_abnormal')}}</Checkbox>
          <Checkbox v-if="!selected" v-model="autoRefresh" @on-change="startRefresh">{{$t('m_k8s_auto_refresh', {seconds: scrapeInterval})}}</Checkbox>
        </div>
      </div>
      <Alert v-if="sourcesFailed" type="error">{{$t('m_k8s_sources_failed')}} <Button size="small" @click="loadSources">{{$t('m_refresh')}}</Button></Alert>
      <Alert v-else-if="unavailableScope" type="warning">{{$t('m_k8s_scope_unavailable', {scope: unavailableScope})}}</Alert>
      <Alert v-else-if="!sourcesLoading && !scope.namespace">{{$t('m_k8s_no_namespace')}}</Alert>
      <p>{{scope.clusterId}} / {{scope.namespace}} — {{state || '-'}} · {{$t('m_k8s_observed_at')}}: {{observedAt ? new Date(observedAt).toLocaleString() : '-'}}</p>
      <Alert v-if="inventoryFailed" type="error">{{$t('m_k8s_inventory_failed')}}</Alert>
      <div v-show="!inventoryCollapsed || !selected">
        <Table border highlight-row :max-height="420" :loading="loading || sourcesLoading" :columns="columns" :data="pageRows" @on-row-click="selectObject" style="margin-top:12px" />
        <Page :total="filteredObjects.length" :current="page" :page-size="20" show-total @on-change="page = $event" style="margin-top:12px" />
      </div>
      <Alert v-if="!selected" style="margin-top:12px">{{$t('m_k8s_select_object')}}</Alert>
    </Card>
    <Card v-if="selected" style="margin-top:16px">
      <p slot="title" class="object-title">{{selected.workloadType}} / {{selected.workloadName}} / {{selected.podName}} / {{selected.container}}</p>
      <p class="object-summary">{{selected.clusterId}} / {{selected.namespace}} · {{$t('m_k8s_node')}}: {{selected.node}} · instanceId: {{selected.instanceId || '-'}} · Pod UID: {{selected.podUid}}</p>
      <p class="object-summary">Phase: {{selected.phase || '-'}} · Ready: {{selected.statusKnown ? $t(selected.ready ? 'm_yes' : 'm_no') : '-'}} · {{$t('m_k8s_metric_restart_count')}}: {{selected.statusKnown ? selected.restartCount : '-'}}</p>
      <Alert v-if="localCollectionEnabled === false" type="warning">{{$t('m_k8s_local_collection_history')}}</Alert>
      <div class="history-toolbar">
        <div class="range-preset">
          <span>{{$t('m_k8s_time_range')}}</span>
          <Select v-model="timePreset" class="range-select" @on-change="applyRange">
            <Option v-for="seconds in [900,3600,21600,86400]" :key="seconds" :value="String(seconds)">{{$t('m_k8s_range_' + seconds)}}</Option>
            <Option value="custom">{{$t('m_k8s_custom_range')}}</Option>
          </Select>
        </div>
        <DatePicker v-if="timePreset === 'custom'" v-model="customRange" type="datetimerange" class="range-custom" @on-change="applyRange" />
        <Button @click="refresh">{{$t('m_refresh')}}</Button>
        <Checkbox v-model="autoRefresh" @on-change="startRefresh">{{$t('m_k8s_auto_refresh', {seconds: scrapeInterval})}}</Checkbox>
        <Poptip trigger="hover" placement="bottom-start" word-wrap width="360" :content="$t('m_k8s_logical_history_help')"><Button type="text" icon="ios-help-circle-outline">{{$t('m_k8s_metric_help')}}</Button></Poptip>
      </div>
      <Tabs v-model="category">
        <TabPane v-for="item in ['resource','log','application']" :key="item" :name="item" :label="$t('m_k8s_category_' + item)" />
      </Tabs>
      <Alert v-if="catalogFailed" type="error">{{$t('m_k8s_catalog_failed')}} <Button size="small" @click="loadCatalog">{{$t('m_refresh')}}</Button></Alert>
      <p v-if="category === 'log'" class="category-help">{{$t('m_k8s_count_window_help',{seconds: countWindow})}}</p>
      <K8sLogSamples v-if="category === 'log'" :object="selected" :range="range" :refresh-token="refreshToken" :periodic-refresh="periodicRefresh" />
      <Poptip v-if="category === 'application'" trigger="hover" placement="bottom-start" word-wrap width="360" :content="$t('m_k8s_app_chart_help')"><Button type="text" icon="ios-help-circle-outline">{{$t('m_k8s_app_metric_help')}}</Button></Poptip>
      <Select v-model="metricSelections[category]" multiple filterable style="width:100%;margin-bottom:16px" @on-change="limitMetrics">
        <Option v-for="metric in categoryMetrics" :key="metric.id" :value="metric.id">{{metricTitle(metric)}} ({{metric.unit}})</Option>
      </Select>
      <Row :gutter="16">
        <Col v-for="metric in shownMetrics" :key="category + metric.id" :xs="24" :xl="12">
        <K8sMetricChart :object="selected" :metric="metric" :range="range" :refresh-token="refreshToken" :periodic-refresh="periodicRefresh" :title="metricTitle(metric)" />
        </Col>
      </Row>
      <Divider>{{$t(alertEvaluationEnabled === true ? 'm_k8s_related_alerts' : 'm_k8s_alerts_last_evaluation')}}</Divider>
      <Table border :columns="alertColumns" :data="selected.alerts || []" />
    </Card>
  </div>
</template>
<script>
import K8sMetricChart from './k8s-metric-chart'
import K8sLogSamples from './k8s-log-samples'
import {filterObjects, metricRange} from './k8s-metrics'
import {readViewState, writeViewState, resolveScope} from './k8s-view-state'

export default {
  components: {
    K8sMetricChart,
    K8sLogSamples
  },
  data() {
    return {
      request: this.$root.$httpRequestEntrance.httpRequestEntrance,
      apiCenter: this.$root.apiCenter,
      clusters: [],
      namespaces: [],
      scope: {
        clusterId: '',
        namespace: ''
      },
      rows: [],
      selected: null,
      state: '',
      observedAt: '',
      loading: false,
      inventoryFailed: false,
      catalogFailed: false,
      localCollectionEnabled: null,
      alertEvaluationEnabled: null,
      sourcesFailed: false,
      sourcesLoading: false,
      unavailableScope: '',
      inventoryCollapsed: false,
      filters: {
        workload: '',
        pod: '',
        instanceId: '',
        onlyAbnormal: false
      },
      page: 1,
      catalog: [],
      category: 'resource',
      scrapeInterval: 30,
      countWindow: 300,
      metricSelections: {
        resource: ['cpu','memory'],
        log: ['log_matches','log_health'],
        application: ['app_jvm_memory_used_bytes','app_http_server_requests_seconds_rate','app_http_server_requests_seconds_mean','app_metric_available']
      },
      timePreset: '3600',
      customRange: [],
      range: metricRange(3600, 30),
      refreshToken: 0,
      periodicRefresh: false,
      autoRefresh: true
    }
  },
  computed: {
    namespaceOptions() { return this.namespaces.filter(row => row.clusterId === this.scope.clusterId && row.enabled) },
    workloads() { return [...new Set(this.rows.map(row => row.workloadName))].sort() },
    filteredObjects() { return filterObjects(this.rows, this.filters, this.alertEvaluationEnabled === true) },
    pageRows() { return this.filteredObjects.slice((this.page-1)*20, this.page*20) },
    categoryMetrics() { return this.catalog.filter(metric => metric.category === this.category) },
    shownMetrics() { return this.categoryMetrics.filter(metric => this.metricSelections[this.category].includes(metric.id)) },
    columns() {
      return [['workloadName','m_k8s_workload_name'],['podName','m_k8s_pod_name'],['container','m_k8s_container'],['node','m_k8s_node'],['instanceId','m_k8s_instance_id']]
        .map(([key,label]) => ({
          title: this.$t(label),
          key,
          minWidth: 140,
          tooltip: true,
          ellipsis: true
        }))
        .concat([{
          title: this.$t('m_k8s_object_health'),
          width: 230,
          render: (h,p) => {
            const alerts = this.alertEvaluationEnabled === true ? (p.row.alerts || []).filter(alert=>['pending','firing'].includes(alert.Status)) : []
            const status = alerts.some(alert=>alert.Status === 'firing') ? 'firing' : 'pending'
            return h('div', [
              h('Tag',{props:{color:p.row.statusKnown ? (p.row.ready ? 'success' : 'error') : 'default'}},this.$t(p.row.statusKnown ? (p.row.ready ? 'm_k8s_ready_yes' : 'm_k8s_ready_no') : 'm_k8s_state_unknown')),
              alerts.length ? h('Tag',{props:{color:status === 'firing' ? 'error' : 'warning'}},this.$t('m_k8s_alert_'+status)+' ('+alerts.length+')') : null
            ])
          }
        }, {
          title: this.$t('m_k8s_metric_restart_count'),
          width: 100,
          render: (h,p) => h('span',p.row.statusKnown ? String(p.row.restartCount) : '-')
        }, {
          title: this.$t('m_actions'),
          width: 100,
          render: (h,p) => h('Button',{
            props: {size: 'small'},
            on: {click: () => this.selectObject(p.row)}
          },this.$t('m_search'))
        }])
    },
    alertColumns() {
      return [{
        title: this.$t('m_k8s_rule_id'),
        render: (h,p) => h('span',p.row.Key.RuleID)
      }, {
        title: this.$t(this.alertEvaluationEnabled === true ? 'm_k8s_state' : 'm_k8s_last_evaluation_state'),
        render: (h,p) => h('Tag',{props:{color:({firing:'error',pending:'warning',normal:'success',recovered:'success'})[p.row.Status] || 'default'}},this.$t('m_k8s_alert_'+p.row.Status))
      }, {
        title: this.$t('m_k8s_metric'),
        key: 'Value'
      }]
    }
  },
  mounted() {
    const saved = this._restoredView = readViewState('objects')
    if (['resource','log','application'].includes(saved.category)) this.category = saved.category
    for (const category of ['resource','log','application']) {
      if (saved.metricSelections && Array.isArray(saved.metricSelections[category])) this.metricSelections[category] = saved.metricSelections[category].filter(id=>typeof id === 'string').slice(0,6)
    }
    if (['900','3600','21600','86400','custom'].includes(saved.timePreset)) this.timePreset = saved.timePreset
    if (Array.isArray(saved.customRange) && saved.customRange.length === 2) this.customRange = saved.customRange.map(value=>new Date(value))
    if (typeof saved.autoRefresh === 'boolean') this.autoRefresh = saved.autoRefresh
    this.inventoryCollapsed = saved.inventoryCollapsed === true
    this.loadSources()
    this.loadCatalog()
    this.startRefresh()
  },
  beforeDestroy() {
    writeViewState('objects', {
      scope: this.scope, filters: this.filters, category: this.category, metricSelections: this.metricSelections,
      timePreset: this.timePreset, customRange: this.customRange.map(value=>new Date(value).getTime()), autoRefresh: this.autoRefresh,
      inventoryCollapsed: this.inventoryCollapsed,
      selected: this.selected ? {workloadType:this.selected.workloadType,workloadName:this.selected.workloadName,podName:this.selected.podName,container:this.selected.container} : null
    })
    this._destroyed = true
    this._inventoryGeneration = (this._inventoryGeneration || 0) + 1
    clearTimeout(this._inventoryDeadline)
    clearTimeout(this._catalogDeadline)
    clearTimeout(this._sourcesDeadline)
    clearInterval(this._refreshTimer)
  },
  methods: {
    loadSources() {
      const generation = this._sourcesGeneration = (this._sourcesGeneration || 0) + 1
      clearTimeout(this._sourcesDeadline)
      this.sourcesFailed = false
      this.sourcesLoading = true
      const remembered = readViewState('scope'), routeQuery = this.$route.query
      const query = {...remembered, ...routeQuery}
      if (routeQuery.clusterId && routeQuery.clusterId !== remembered.clusterId && !routeQuery.namespace) query.namespace = ''
      let finished = false
      const current = () => !this._destroyed && this._sourcesGeneration === generation && !finished
      const fail = () => {
        if (!current()) {return}
        finished = true
        clearTimeout(this._sourcesDeadline)
        this.sourcesFailed = true
        this.sourcesLoading = false
      }
      this._sourcesDeadline = setTimeout(fail,30000)
      this.request('POST', this.apiCenter.k8sClusterList, {}, data => {
        if (!current()) {return}
        this.clusters = (data || []).filter(row => row.enabled)
        this.request('POST', this.apiCenter.k8sNamespaceList, {}, namespaces => {
          if (!current()) {return}
          finished = true
          clearTimeout(this._sourcesDeadline)
          this.namespaces = namespaces || []
          this.sourcesLoading = false
          const resolved = resolveScope(this.clusters,this.namespaces,query)
          this.unavailableScope = resolved.unavailable ? [query.clusterId,query.namespace].filter(Boolean).join(' / ') : ''
          this.scope = {clusterId:resolved.clusterId,namespace:resolved.namespace}
          if (this.scope.namespace) writeViewState('scope',this.scope)
          this.loadObjects(true)
        }, {isNeedloading: false}, fail)
      }, {isNeedloading: false}, fail)
    },
    metricTitle(metric) {
      return metric.category === 'application' ? metric.name + (metric.id.endsWith('_mean') ? ' ('+this.$t('m_k8s_metric_mean')+')' : metric.id.endsWith('_rate') ? ' ('+this.$t('m_k8s_metric_rate')+')' : '') : this.$t('m_k8s_metric_'+metric.id)
    },
    clusterChanged(namespace) {
      this.unavailableScope = ''
      this.scope.namespace = this.namespaceOptions.some(row => row.namespace === namespace) ? namespace : ((this.namespaceOptions[0] || {}).namespace || '')
      if (this.scope.namespace) writeViewState('scope',this.scope)
      this.loadObjects(true)
    },
    namespaceChanged() { this.unavailableScope = ''; writeViewState('scope',this.scope); this.loadObjects(true) },
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
        this.localCollectionEnabled = !failed && typeof data.localCollectionEnabled === 'boolean' ? data.localCollectionEnabled : null
        this.alertEvaluationEnabled = !failed && typeof data.alertEvaluationEnabled === 'boolean' ? data.alertEvaluationEnabled : null
        if (failed) {return}
        this.catalog = data.metrics || []
        this.scrapeInterval = data.scrapeIntervalSeconds
        this.countWindow = data.countWindowSeconds
        this.applyRange()
        this.startRefresh()
      }
      this._catalogDeadline = setTimeout(() => finish(null,true),30000)
      this.request('GET', this.apiCenter.k8sMetricCatalog, {}, data => finish(data,false), {isNeedloading: false}, () => finish(null,true))
    },
    loadObjects(clear = false, periodic = false) {
      const scopeKey = JSON.stringify([this.scope.clusterId,this.scope.namespace])
      if (periodic === true && this.loading && this._inventoryScopeKey === scopeKey) {return}
      this._inventoryScopeKey = scopeKey
      const generation = this._inventoryGeneration = (this._inventoryGeneration || 0) + 1
      clearTimeout(this._inventoryDeadline)
      if (clear) {
        this.rows=[]; this.selected=null; this.state=''; this.observedAt=''; this.page=1
        this.filters = {workload:'',pod:'',instanceId:'',onlyAbnormal:false}
        const saved = this._restoredView
        if (saved && saved.scope && saved.scope.clusterId === this.scope.clusterId && saved.scope.namespace === this.scope.namespace) {
          for (const field of ['workload','pod','instanceId']) if (typeof (saved.filters || {})[field] === 'string') this.filters[field] = saved.filters[field]
          this.filters.onlyAbnormal = (saved.filters || {}).onlyAbnormal === true
          this._restoreSelected = saved.selected
        }
        this._restoredView = null
        const query = (this.$route || {}).query || {}
        if (query.clusterId === this.scope.clusterId && query.namespace === this.scope.namespace) {
          if (typeof query.podName === 'string') this.filters.pod = query.podName
          if (typeof query.instanceId === 'string') this.filters.instanceId = query.instanceId
        }
      }
      this.loading = false
      if (!this.scope.clusterId || !this.scope.namespace) {return}
      this.loading = true
      this.inventoryFailed = false
      let finished = false
      const finish = (data, failed) => {
        if (this._inventoryGeneration !== generation || finished) {return}
        finished = true
        clearTimeout(this._inventoryDeadline)
        this.loading = false
        this.inventoryFailed = failed
        if (failed) {return}
        this.rows = data.objects || []
        this.state = data.state
        this.observedAt = data.observedAt
        if (this.selected) {
          const selected = this.selected
          this.selected = this.rows.find(row => row.workloadType===selected.workloadType && row.workloadName===selected.workloadName && row.podName===selected.podName && row.container===selected.container) || null
        }
        if (this._restoreSelected) {
          const saved = this._restoreSelected
          this.selected = filterObjects(this.rows,this.filters,this.alertEvaluationEnabled === true).find(row=>row.workloadType===saved.workloadType && row.workloadName===saved.workloadName && row.podName===saved.podName && row.container===saved.container) || null
          this._restoreSelected = null
        }
        if ((this.page-1)*20 >= filterObjects(this.rows,this.filters,this.alertEvaluationEnabled === true).length) {this.page=1}
      }
      this._inventoryDeadline = setTimeout(() => finish(null,true),30000)
      this.request('GET', this.apiCenter.k8sObjects, {...this.scope}, data => finish(data,false), {isNeedloading: false}, () => finish(null,true))
    },
    filterChanged() {
      this.page = 1
      if (this.selected && !this.filteredObjects.some(row => row.workloadType===this.selected.workloadType && row.workloadName===this.selected.workloadName && row.podName===this.selected.podName && row.container===this.selected.container)) {this.selected=null}
    },
    selectObject(row) { this.selected=row; this.inventoryCollapsed=true; this.applyRange() },
    applyRange() {
      this.periodicRefresh = false
      if (this.timePreset !== 'custom') {
        this.range = metricRange(Number(this.timePreset),this.scrapeInterval)
        return
      }
      const start = Math.floor(new Date(this.customRange[0]).getTime()/1000)
      const end = Math.floor(new Date(this.customRange[1]).getTime()/1000)
      if (!Number.isFinite(start) || !Number.isFinite(end)) {return}
      if (start < 0 || end <= start || end > Math.floor(Date.now()/1000)+60 || end-start > 30*86400) { this.$Message.error(this.$t('m_k8s_invalid_time_range')); return }
      this.range = {
        ...metricRange(end-start,this.scrapeInterval,end),
        start,
        end
      }
    },
    limitMetrics() {
      if (this.metricSelections[this.category].length > 6) {
        this.metricSelections[this.category] = this.metricSelections[this.category].slice(0,6)
        this.$Message.info(this.$t('m_k8s_chart_limit'))
      }
    },
    refresh(periodic = false) { this.loadObjects(false,periodic === true); this.applyRange(); this.periodicRefresh = periodic === true; this.refreshToken++ },
    startRefresh() {
      clearInterval(this._refreshTimer)
      if (this.autoRefresh) {this._refreshTimer=setInterval(() => { if (!document.hidden) {this.refresh(true)} },this.scrapeInterval*1000)}
    }
  }
}
</script>
<style scoped>
.object-filters {
  display: flex;
  flex-wrap: wrap;
  align-items: flex-end;
  gap: 12px 16px;
  margin: 0 0 18px;
  padding: 14px 16px;
  background: #f7f9fc;
  border: 1px solid #e8edf3;
  border-radius: 4px;
}
.object-filter { flex-shrink: 0; min-width: 0; max-width: 100%; }
.object-filter > label {
  display: block;
  margin-bottom: 6px;
  color: #515a6e;
  font-size: 12px;
  line-height: 16px;
}
.filter-scope { width: 150px; }
.filter-workload { width: 230px; }
.filter-text { width: 190px; }
.filter-auto >>> .ivu-checkbox-wrapper { margin-right: 0; line-height: 32px; white-space: nowrap; }
.object-title { height: auto; white-space: normal; overflow-wrap: anywhere; }
.object-summary { line-height: 1.7; overflow-wrap: anywhere; }
.category-help { margin:0 0 10px; color:#808695; line-height:1.6; }
.history-toolbar, .range-preset { display: flex; flex-wrap: wrap; align-items: center; gap: 8px; }
.history-toolbar { gap: 12px; margin: 12px 0; }
.range-select { width: 150px; }
.range-custom { width: 340px; max-width: 100%; }
@media (max-width: 1200px) {
  .filter-workload { width: 210px; }
  .filter-text { width: 170px; }
}
@media (max-width: 768px) {
  .object-filters { align-items: stretch; }
  .filter-scope, .filter-workload, .filter-text { width: 100%; }
  .range-custom { width: 100%; }
}
</style>
