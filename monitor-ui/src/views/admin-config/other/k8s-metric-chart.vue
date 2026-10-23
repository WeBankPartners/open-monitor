<template>
  <Card class="metric-card">
    <p slot="title" :title="title">{{title}}</p>
    <div slot="extra" class="metric-actions">
      <span>{{metric.unit}}</span>
      <Button size="small" icon="ios-expand" :disabled="!series.length" @click="expanded = true">{{$t('m_k8s_expand_chart')}}</Button>
    </div>
    <div class="chart-body">
      <Spin v-if="loading" fix />
      <Alert v-if="!loading && failed" type="error">{{$t('m_k8s_history_failed')}}</Alert>
      <Alert v-else-if="!loading && !series.length">{{$t('m_k8s_no_metric_data')}}</Alert>
      <div v-show="series.length" ref="chart" style="height:280px" ></div>
    </div>
    <Modal v-model="expanded" fullscreen footer-hide :mask-closable="false" @on-visible-change="expandedChanged">
      <p slot="header" class="expanded-title">{{title}}</p>
      <div v-if="expanded">
        <p class="expanded-summary">{{object.clusterId}} / {{object.namespace}} / {{object.podName}} / {{object.container}} · {{metric.unit}}</p>
        <div class="chart-body">
          <Spin v-if="loading" fix />
          <Alert v-if="!loading && failed" type="error">{{$t('m_k8s_history_failed')}}</Alert>
          <Alert v-else-if="!loading && !series.length">{{$t('m_k8s_no_metric_data')}}</Alert>
          <div v-show="series.length" ref="expandedChart" class="expanded-chart"></div>
        </div>
        <Collapse v-if="series.length" style="margin-top:12px">
          <Panel name="labels">
            {{$t('m_k8s_series_details', {count: series.length})}}
            <div slot="content" class="series-details">
              <div v-for="item in series" :key="item.name" class="series-detail">
                <strong>{{item.displayName}}</strong>
                <dl>
                  <div v-for="(value, key) in item.labels" :key="key" class="label-row"><dt>{{key}}</dt><dd>{{value}}</dd></div>
                </dl>
              </div>
            </div>
          </Panel>
        </Collapse>
      </div>
    </Modal>
  </Card>
</template>
<script>
import * as echarts from 'echarts'
import {chartSeries, formatMetricValue} from './k8s-metrics'

export default {
  props: {
    object: Object,
    metric: Object,
    range: Object,
    refreshToken: Number,
    title: String
  },
  data() {
    return {
      request: this.$root.$httpRequestEntrance.httpRequestEntrance,
      apiCenter: this.$root.apiCenter,
      series: [],
      expanded: false,
      loading: false,
      failed: false
    }
  },
  computed: {
    queryKey() {
      const object = this.object
      return JSON.stringify([object.clusterId, object.namespace, object.workloadType, object.workloadName, object.podName, object.container, this.metric.id, this.range, this.refreshToken])
    }
  },
  watch: {queryKey() { this.load() }},
  mounted() {
    this._chart = echarts.init(this.$refs.chart)
    this._resizeChart = () => this.renderChart()
    window.addEventListener('resize', this._resizeChart)
    this.load()
  },
  beforeDestroy() {
    this._destroyed = true
    this._generation = (this._generation || 0) + 1
    clearTimeout(this._deadline)
    window.removeEventListener('resize', this._resizeChart)
    if (this._chart) {this._chart.dispose()}
    if (this._expandedChart) {this._expandedChart.dispose()}
    this._chart = null
    this._expandedChart = null
  },
  methods: {
    expandedChanged(visible) {
      if (!visible) {
        if (this._expandedChart) {this._expandedChart.dispose()}
        this._expandedChart = null
        return
      }
      this.$nextTick(() => {
        if (this._destroyed || !this.expanded || !this.$refs.expandedChart) {return}
        this._expandedChart = echarts.init(this.$refs.expandedChart)
        this.renderChart()
      })
    },
    load() {
      const generation = this._generation = (this._generation || 0) + 1
      clearTimeout(this._deadline)
      this.series = []
      this.loading = true
      this.failed = false
      const range = {...this.range}
      let finished = false
      const finish = (data, failed) => {
        if (this._generation !== generation || finished) {return}
        finished = true
        clearTimeout(this._deadline)
        this.loading = false
        this.failed = failed
        this.series = failed ? [] : chartSeries(data, range)
        this.renderChart()
      }
      this._deadline = setTimeout(() => finish(null, true), 30000)
      const {
        clusterId, namespace, workloadType, workloadName, podName, container
      } = this.object
      this.request(
        'POST',
        this.apiCenter.k8sMetricHistory,
        {
          clusterId,
          namespace,
          workloadType,
          workloadName,
          podName,
          container,
          metric: this.metric.id,
          ...range
        },
        data => finish(data, false),
        {isNeedloading: false},
        () => finish(null, true)
      )
    },
    renderChart() {
      this.$nextTick(() => {
        if (!this._chart) {return}
        const unit = this.metric.unit
        const escape = text => String(text).replace(/[&<>"']/g, c => ({
          '&': '&amp;',
          '<': '&lt;',
          '>': '&gt;',
          '"': '&quot;',
          '\'': '&#39;'
        }[c]))
        const displayNames = new Map(this.series.map(item => [item.name, item.displayName]))
        const defaultLegend = ((this._chart.getOption() || {}).legend || [])[0] || {}
        for (const chart of [this._chart, this._expandedChart]) {
          if (!chart) {continue}
          chart.resize()
          const legend = ((chart.getOption() || {}).legend || [])[0] || defaultLegend
          chart.setOption({
            animation: false,
            tooltip: {
              trigger: 'axis',
              confine: true,
              enterable: true,
              extraCssText: 'box-sizing:border-box;white-space:normal;overflow-wrap:break-word;word-break:break-word;overflow-y:auto;max-width:' + Math.max(1, Math.min(460, chart.getWidth() - 24)) + 'px;max-height:' + Math.max(1, Math.min(300, chart.getHeight() - 24)) + 'px;',
              formatter: points => points.length ? escape(new Date(points[0].value[0]).toLocaleString()) + '<br/>' + points.map(point => {
                const name = displayNames.get(point.seriesName) || point.seriesName
                const color = /^#[0-9a-f]{3,8}$/i.test(point.color) ? point.color : '#888'
                return '<div style="display:flex;align-items:baseline;gap:12px;margin-top:4px"><span style="flex:1;min-width:0"><span style="color:' + color + '">&#9679;</span> ' + escape(name) + '</span><strong style="white-space:nowrap">' + escape(formatMetricValue(point.value[1], unit)) + '</strong></div>'
              }).join('') : ''
            },
            legend: {
              type: 'scroll',
              bottom: 0,
              selected: legend.selected || {},
              formatter: name => {
                const label = displayNames.get(name) || name
                const suffix = '… #' + (this.series.findIndex(item => item.name === name) + 1)
                return label.length > 60 ? label.slice(0, 60 - suffix.length) + suffix : label
              }
            },
            grid: {
              top: 15,
              left: 80,
              right: 20,
              bottom: 65
            },
            xAxis: {
              type: 'time',
              min: this.range.start * 1000,
              max: this.range.end * 1000
            },
            yAxis: {
              type: 'value',
              axisLabel: {formatter: value => formatMetricValue(value, unit)}
            },
            series: this.series
          }, true)
        }
      })
    }
  }
}
</script>
<style scoped>
.metric-card { margin-bottom:16px; }
.metric-card >>> .ivu-card-head { padding-right:180px; }
.metric-actions { display:flex; align-items:center; gap:8px; }
.chart-body { min-height:280px; position:relative; }
.expanded-title { margin-right:28px; overflow-wrap:anywhere; }
.expanded-summary { margin-bottom:12px; overflow-wrap:anywhere; }
.expanded-chart { height:calc(100vh - 220px); min-height:280px; }
.series-details { max-height:32vh; overflow:auto; }
.series-detail { padding:12px 0; border-bottom:1px solid #e8eaec; overflow-wrap:anywhere; }
.label-row { display:flex; gap:12px; padding-top:4px; }
.label-row dt { flex:0 0 140px; font-weight:normal; }
.label-row dd { margin:0; min-width:0; }
@media (max-width: 600px) {
  .label-row { display:block; }
}
</style>
