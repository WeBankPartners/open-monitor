<template>
  <Card class="metric-card">
    <p slot="title">{{title}}</p>
    <span slot="extra">{{metric.unit}}</span>
    <div class="chart-body">
      <Spin v-if="loading" fix />
      <Alert v-if="!loading && failed" type="error">{{$t('m_k8s_history_failed')}}</Alert>
      <Alert v-else-if="!loading && !series.length">{{$t('m_k8s_no_metric_data')}}</Alert>
      <div v-show="series.length" ref="chart" style="height:280px" ></div>
    </div>
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
    this._resizeChart = () => this._chart && this._chart.resize()
    window.addEventListener('resize', this._resizeChart)
    this.load()
  },
  beforeDestroy() {
    this._generation = (this._generation || 0) + 1
    clearTimeout(this._deadline)
    window.removeEventListener('resize', this._resizeChart)
    if (this._chart) {this._chart.dispose()}
    this._chart = null
  },
  methods: {
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
        this._chart.setOption({
          animation: false,
          tooltip: {
            trigger: 'axis',
            confine: true,
            formatter: points => points.length ? escape(new Date(points[0].value[0]).toLocaleString()) + '<br/>' + points.map(point => escape(point.seriesName) + ': ' + escape(formatMetricValue(point.value[1], unit))).join('<br/>') : ''
          },
          legend: {
            type: 'scroll',
            bottom: 0
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
        this._chart.resize()
      })
    }
  }
}
</script>
<style scoped>
.metric-card { margin-bottom:16px; }
.chart-body { min-height:280px; position:relative; }
</style>
