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
import {cloneDeep, isEmpty} from 'lodash'
import {drawPieChart} from '@/assets/config/chart-rely'

export default {
  name: '',
  data() {
    return {
      elId: null,
      chartTitle: null,
      config: '',
      myChart: '',
      interval: '',
      noDataType: 'normal' // 该字段为枚举，noConfig (没有配置信息)， noData(没有请求到数据)， normal(有数据正常)
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
        this.isAutoRefresh()
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
  destroyed() {
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
      this.noDataType = 'normal'
      if (this.chartInfo.chartParams.data.length === 0) {
        this.noDataType = 'noConfig'
        return
      }
      this.isAutoRefresh()
      const params = this.chartInfo.chartParams.data
      params.forEach(p => {
        p.start = this.chartInfo.start
        p.end = this.chartInfo.end
        p.time_second = this.chartInfo.time_second,
        p.custom_chart_guid = this.chartInfo.elId
      })
      this.elId = this.chartInfo.elId

      const modalElement = document.querySelector('#edit-view')
      const offset = this.$el.getBoundingClientRect()
      const offsetTop = offset.top
      const offsetBottom = offset.bottom
      if ((offsetTop <= window.innerHeight && offsetBottom >= 0 && !modalElement) || type === 'mounted') {
        this.$root.$httpRequestEntrance.httpRequestEntrance(
          'POST',
          this.$root.apiCenter.metricConfigPieView.api,
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
