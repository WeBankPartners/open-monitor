<template>
  <div class="k8s-objects">
    <Card>
      <p slot="title">{{$t('m_k8s_objects')}}</p>
      <div class="object-filters">
        <div class="object-filter filter-scope">
          <label>{{$t('m_k8s_cluster_id')}}</label>
          <Select v-model="scope.clusterId" @on-change="clusterChanged()"><Option v-for="cluster in clusters" :key="cluster.clusterId" :value="cluster.clusterId">{{cluster.clusterId}}</Option></Select>
        </div>
        <div class="object-filter filter-scope">
          <label>{{$t('m_k8s_namespace')}}</label>
          <Select v-model="scope.namespace" @on-change="loadObjects(true)"><Option v-for="namespace in namespaceOptions" :key="namespace.namespace" :value="namespace.namespace">{{namespace.namespace}}</Option></Select>
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
          <Checkbox v-model="autoRefresh" @on-change="startRefresh">{{$t('m_k8s_auto_refresh', {seconds: scrapeInterval})}}</Checkbox>
        </div>
      </div>
      <Alert v-if="sourcesFailed" type="error">{{$t('m_k8s_sources_failed')}} <Button size="small" @click="loadSources">{{$t('m_refresh')}}</Button></Alert>
      <Alert v-else-if="!scope.namespace">{{$t('m_k8s_no_namespace')}}</Alert>
      <p>{{scope.clusterId}} / {{scope.namespace}} — {{state || '-'}} · {{$t('m_k8s_observed_at')}}: {{observedAt ? new Date(observedAt).toLocaleString() : '-'}}</p>
      <Alert v-if="inventoryFailed" type="error">{{$t('m_k8s_inventory_failed')}}</Alert>
      <Table border highlight-row :max-height="420" :loading="loading" :columns="columns" :data="pageRows" @on-row-click="selectObject" style="margin-top:12px" />
      <Page :total="filteredObjects.length" :current="page" :page-size="20" show-total @on-change="page = $event" style="margin-top:12px" />
      <Alert v-if="!selected" style="margin-top:12px">{{$t('m_k8s_select_object')}}</Alert>
    </Card>
    <Card v-if="selected" style="margin-top:16px">
      <p slot="title" class="object-title">{{selected.workloadType}} / {{selected.workloadName}} / {{selected.podName}} / {{selected.container}}</p>
      <p class="object-summary">{{selected.clusterId}} / {{selected.namespace}} · {{$t('m_k8s_node')}}: {{selected.node}} · instanceId: {{selected.instanceId || '-'}} · Pod UID: {{selected.podUid}}</p>
      <p class="object-summary">Phase: {{selected.phase || '-'}} · Ready: {{selected.statusKnown ? $t(selected.ready ? 'm_yes' : 'm_no') : '-'}} · {{$t('m_k8s_metric_restart_count')}}: {{selected.statusKnown ? selected.restartCount : '-'}}</p>
      <Alert style="margin-top:12px">{{$t('m_k8s_logical_history_help')}}</Alert>
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
      </div>
      <Tabs v-model="category">
        <TabPane v-for="item in ['resource','log','application']" :key="item" :name="item" :label="$t('m_k8s_category_' + item)" />
      </Tabs>
      <Alert v-if="catalogFailed" type="error">{{$t('m_k8s_catalog_failed')}} <Button size="small" @click="loadCatalog">{{$t('m_refresh')}}</Button></Alert>
      <Alert v-if="category === 'log'">{{$t('m_k8s_count_window_help',{seconds: countWindow})}}</Alert>
      <Alert v-if="category === 'application'">{{$t('m_k8s_app_chart_help')}}</Alert>
      <Select v-model="metricSelections[category]" multiple filterable style="width:100%;margin-bottom:16px" @on-change="limitMetrics">
        <Option v-for="metric in categoryMetrics" :key="metric.id" :value="metric.id">{{metricTitle(metric)}} ({{metric.unit}})</Option>
      </Select>
      <Row :gutter="16">
        <Col v-for="metric in shownMetrics" :key="category + metric.id" :xs="24" :xl="12">
        <K8sMetricChart :object="selected" :metric="metric" :range="range" :refresh-token="refreshToken" :title="metricTitle(metric)" />
        </Col>
      </Row>
      <Divider>{{$t('m_k8s_related_alerts')}}</Divider>
      <Table border :columns="alertColumns" :data="selected.alerts || []" />
    </Card>
  </div>
</template>
<script>
import K8sMetricChart from './k8s-metric-chart'
import {filterObjects, metricRange} from './k8s-metrics'

export default {
  components: {K8sMetricChart},
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
      sourcesFailed: false,
      filters: {
        workload: '',
        pod: '',
        instanceId: ''
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
      autoRefresh: true
    }
  },
  computed: {
    namespaceOptions() { return this.namespaces.filter(row => row.clusterId === this.scope.clusterId && row.enabled) },
    workloads() { return [...new Set(this.rows.map(row => row.workloadName))].sort() },
    filteredObjects() { return filterObjects(this.rows, this.filters) },
    pageRows() { return this.filteredObjects.slice((this.page-1)*20, this.page*20) },
    categoryMetrics() { return this.catalog.filter(metric => metric.category === this.category) },
    shownMetrics() { return this.categoryMetrics.filter(metric => this.metricSelections[this.category].includes(metric.id)) },
    columns() {
      return [['workloadName','m_k8s_workload_name'],['podName','m_k8s_pod_name'],['container','m_k8s_container'],['node','m_k8s_node'],['instanceId','m_k8s_instance_id']]
        .map(([key,label]) => ({
          title: this.$t(label),
          key,
          minWidth: 140
        }))
        .concat([{
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
        title: this.$t('m_k8s_state'),
        key: 'Status'
      }, {
        title: this.$t('m_k8s_metric'),
        key: 'Value'
      }]
    }
  },
  mounted() {
    const query = this.$route.query
    this.filters.pod = query.podName || ''
    this.filters.instanceId = query.instanceId || ''
    this.loadSources()
    this.loadCatalog()
    this.startRefresh()
  },
  beforeDestroy() {
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
      const query = {...this.$route.query}
      let finished = false
      const current = () => !this._destroyed && this._sourcesGeneration === generation && !finished
      const fail = () => {
        if (!current()) {return}
        finished = true
        clearTimeout(this._sourcesDeadline)
        this.sourcesFailed = true
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
          this.scope.clusterId = this.clusters.some(row => row.clusterId === query.clusterId) ? query.clusterId : ((this.clusters[0] || {}).clusterId || '')
          this.clusterChanged(query.namespace)
        }, {isNeedloading: false}, fail)
      }, {isNeedloading: false}, fail)
    },
    metricTitle(metric) {
      return metric.category === 'application' ? metric.name + (metric.id.endsWith('_mean') ? ' ('+this.$t('m_k8s_metric_mean')+')' : metric.id.endsWith('_rate') ? ' ('+this.$t('m_k8s_metric_rate')+')' : '') : this.$t('m_k8s_metric_'+metric.id)
    },
    clusterChanged(namespace) {
      this.scope.namespace = this.namespaceOptions.some(row => row.namespace === namespace) ? namespace : ((this.namespaceOptions[0] || {}).namespace || '')
      this.loadObjects(true)
    },
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
    loadObjects(clear = false) {
      const generation = this._inventoryGeneration = (this._inventoryGeneration || 0) + 1
      clearTimeout(this._inventoryDeadline)
      if (clear) { this.rows=[]; this.selected=null; this.state=''; this.observedAt=''; this.filters.workload=''; this.page=1 }
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
        if ((this.page-1)*20 >= filterObjects(this.rows,this.filters).length) {this.page=1}
      }
      this._inventoryDeadline = setTimeout(() => finish(null,true),30000)
      this.request('GET', this.apiCenter.k8sObjects, {...this.scope}, data => finish(data,false), {isNeedloading: false}, () => finish(null,true))
    },
    filterChanged() {
      this.page = 1
      if (this.selected && !this.filteredObjects.some(row => row.workloadType===this.selected.workloadType && row.workloadName===this.selected.workloadName && row.podName===this.selected.podName && row.container===this.selected.container)) {this.selected=null}
    },
    selectObject(row) { this.selected=row; this.applyRange() },
    applyRange() {
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
    refresh() { this.loadObjects(); this.applyRange(); this.refreshToken++ },
    startRefresh() {
      clearInterval(this._refreshTimer)
      if (this.autoRefresh) {this._refreshTimer=setInterval(() => { if (!document.hidden) {this.refresh()} },this.scrapeInterval*1000)}
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
