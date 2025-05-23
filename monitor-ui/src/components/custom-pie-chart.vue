<template>
  <div class="single-chart">
    <div v-show="noDataType === 'normal'" :id="elId" class="echart" :style="chartInfo.style" ></div>
    <div v-show="noDataType !== 'normal'" class="echart echart-no-data-tip">
      <span v-if="noDataType === 'noConfig'">{{this.$t('m_noConfig')}}</span>
      <span v-else>{{this.$t('m_noData')}}</span>
    </div>
  </div>
</template>

<script>
// 引入 ECharts 主模块
import {cloneDeep, isEmpty, hasIn} from 'lodash'
import {drawPieChart} from '@/assets/config/chart-rely'
import {dateToTimestamp} from '@/assets/js/utils'

export default {
  name: '',
  data() {
    return {
      elId: null,
      chartTitle: null,
      config: '',
      myChart: '',
      interval: '',
      noDataType: 'normal', // 该字段为枚举，noConfig (没有配置信息)， noData(没有请求到数据)， normal(有数据正常)
      apiCenter: this.$root.apiCenter,
      request: this.$root.$httpRequestEntrance.httpRequestEntrance,
      isChartInWindow: null
    }
  },
  props: {
    chartInfo: Object,
    params: Object,
    chartIndex: Number,
    refreshNow: Boolean
  },
  watch: {
    params: {
      handler() {
        this.isChartInWindow = this.calcIsChartInWindow()
        if (this.isChartInWindow) {
          this.isAutoRefresh()
          this.getchartdata()
        } else {
          this.clearIntervalInfo()
        }
      },
      deep: true
    },
    refreshNow: {
      handler() {
        this.getchartdata()
      }
    }
  },
  mounted() {
    this.getchartdata('mounted')
  },
  beforeDestroy() {
    clearInterval(this.interval)
  },
  methods: {
    isAutoRefresh() {
      clearInterval(this.interval)
      if (this.params.autoRefresh > 0 && this.params.dateRange[0] === '') {
        this.interval = setInterval(() => {
          this.getchartdata()
        },this.params.autoRefresh*1000)
      }
    },
    getchartdata(type = '') {
      if (hasIn(this, '$parent.$parent.$parent.isNeedRefresh') && !this.$parent.$parent.$parent.isNeedRefresh) {
        this.clearIntervalInfo()
        return
      }
      this.noDataType = 'normal'
      if (this.chartInfo.chartParams.data.length === 0) {
        this.noDataType = 'noConfig'
        return
      }
      this.isAutoRefresh()
      const params = this.chartInfo.chartParams.data
      params.forEach(p => {
        p.start = dateToTimestamp(this.params.dateRange[0])
        p.end = dateToTimestamp(this.params.dateRange[1])
        p.time_second = this.params.timeTnterval,
        p.custom_chart_guid = this.chartInfo.elId
      })
      this.elId = this.chartInfo.elId

      const modalElement = document.querySelector('#edit-view')
      const maxViewElement = document.querySelector('#max-view-chart')
      const offset = this.$el.getBoundingClientRect()
      const offsetTop = offset.top
      const offsetBottom = offset.bottom
      if ((offsetTop <= window.innerHeight && offsetBottom >= 0 && !modalElement && !maxViewElement) || type === 'mounted') {
        this.request(
          'POST',
          this.apiCenter.metricConfigPieView.api,
          params,
          responseData => {
            if (responseData.legend && responseData.legend.length === 0) {
              this.noDataType = 'noData'
            } else {
              this.noDataType = 'normal'
              window['view-config-selected-line-data'] = window['view-config-selected-line-data'] || {}
              window['view-config-selected-line-data'][this.chartInfo.elId] = window['view-config-selected-line-data'][this.chartInfo.elId] || {}
              if (isEmpty(window['view-config-selected-line-data'][this.chartInfo.elId])
                || (!isEmpty(window['view-config-selected-line-data'][this.chartInfo.elId]) && Object.keys(window['view-config-selected-line-data'][this.chartInfo.elId]).length !== responseData.legend.length)) {
                window['view-config-selected-line-data'][this.chartInfo.elId] = {}
                responseData.legend.forEach(name => {
                  window['view-config-selected-line-data'][this.chartInfo.elId][name] = true
                })
              }
              responseData.chartId = this.chartInfo.elId
              const chartInstance = drawPieChart(this, responseData)
              if (chartInstance) {
                chartInstance.on('legendselectchanged', params => {
                  window['view-config-selected-line-data'][this.chartInfo.elId] = cloneDeep(params.selected)
                })
              }
            }
          },
          { isNeedloading: false }
        )
      }
    },
    clearIntervalInfo() {
      clearInterval(this.interval)
      this.interval = null
    },
    calcIsChartInWindow() {
      const offset = this.$el.getBoundingClientRect()
      const offsetTop = offset.top
      const offsetBottom = offset.bottom
      return offsetTop <= window.innerHeight && offsetBottom >= 0
    }
  },
  components: {},
}
</script>

<style scoped lang="less">
  .single-chart {
    padding: 5px;
    padding-top: 0;
    .echart {
       border-radius: 4px;
    }
    .echart-no-data-tip {
      text-align: center;
      vertical-align: middle;
      display: table-cell;
      position: absolute;
      top: 50%;
      left: 50%;
      transform: translate(-50%, -50%);
    }
  }
</style>
