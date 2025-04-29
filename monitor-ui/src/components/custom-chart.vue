<template>
  <div id='custome-chart-view' class="single-chart" @mouseleave="onMouseLeaveContent">
    <div v-show="noDataType === 'normal'">
      <div :id="elId" class="echart" :style="chartInfo.style">
      </div>
    </div>
    <div v-show="noDataType !== 'normal'" class="echart echart-no-data-tip">
      <span v-if="noDataType === 'noConfig'">{{this.$t('m_noConfig')}}</span>
      <span v-else>{{this.$t('m_noData')}}</span>
    </div>
  </div>
</template>

<script>
// 引入 ECharts 主模块
import {cloneDeep, isEmpty, hasIn} from 'lodash'
import {readyToDraw} from '@/assets/config/chart-rely'
import { changeSeriesColor } from '@/assets/config/random-color'
import {chartTooltipContain, dateToTimestamp} from '@/assets/js/utils'

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
      chartInstance: null,
      isFirstRefresh: true,
      hasNotRequest: true,
      isToolTipShow: false,
      apiCenter: this.$root.apiCenter,
      request: this.$root.$httpRequestEntrance.httpRequestEntrance,
      isChartInWindow: null
    }
  },
  props: {
    chartInfo: Object,
    params: Object,
    chartIndex: Number,
    refreshNow: Boolean,
    scrollRefresh: Boolean, // 外层组件滚动时刷新
    hasNotRequestStatus: Boolean // 基于该值的变化，来改变hasNotRequest，用于重置hasNotRequest
  },
  computed: {
    isInViewConfig() { // 是否是在自定义看板中
      return !isEmpty(this.chartInfo.parsedDisplayConfig)
    }
  },
  watch: {
    params: {
      handler() {
        this.isChartInWindow = this.calcIsChartInWindow()
        if (this.isChartInWindow) {
          this.isAutoRefresh()
          this.getchartdata()
        }
      },
      deep: true
    },
    refreshNow: {
      handler() {
        this.getchartdata()
        // console.error(9998, this.isFirstRefresh)
        // if (this.isFirstRefresh) {
        //   this.getchartdata('mounted')
        //   setTimeout(() => {
        //     this.isFirstRefresh = false
        //   }, 5000)
        // } else {
        //   this.getchartdata()
        // }
      }
    },
    scrollRefresh: { // 当外层组件滚动时刷新，假如有数据时不加载
      handler() {
        if (this.hasNotRequest) {
          this.getchartdata()
        }
      }
    },
    hasNotRequestStatus: {
      handler() {
        this.hasNotRequest = true
      }
    }
  },
  mounted() {
    if (!this.isInViewConfig) {
      this.getchartdata('mounted')
    } else {
      if (this.isFirstRefresh) {
        this.getchartdata('mounted')
        setTimeout(() => {
          this.isFirstRefresh = false
        }, 5000)
      }
    }
    window.addEventListener('scroll', this.scrollHandle, true)
    window.addEventListener('visibilitychange', this.isTabActive, true)
  },
  beforeDestroy() {
    this.clearIntervalInfo()
    window.removeEventListener('scroll', this.scrollHandle, true)
    window.removeEventListener('visibilitychange', this.isTabActive, true)
    if (this.chartInstance) {
      setTimeout(() => {
        this.chartInstance.dispatchAction({
          type: 'hideTip'
        })
        this.$nextTick(() => {
          this.chartInstance = null
        })
      }, 500)
    }
  },
  methods: {
    isTabActive() {
      if (document.hidden) {
        this.clearIntervalInfo()
      } else {
        this.isAutoRefresh()
      }
    },
    clearIntervalInfo() {
      clearInterval(this.interval)
      this.interval = null
    },
    scrollHandle() {
      this.isChartInWindow = this.calcIsChartInWindow()
      if (this.isChartInWindow) {
        this.isAutoRefresh()
      } else {
        this.clearIntervalInfo()
      }
    },
    isAutoRefresh() {
      this.clearIntervalInfo()
      if (!this.isChartInWindow) {
        return
      }
      const element = document.querySelector('#custome-chart-view')
      if (!element) {
        return
      }
      if (this.params.autoRefresh > 0 && this.params.dateRange[0] === '') {
        this.interval = setInterval(() => {
          this.getchartdata()
        },this.params.autoRefresh * 1000)
      }
    },
    async getchartdata(type = '') {
      if (hasIn(this, '$parent.$parent.$parent.isNeedRefresh') && !this.$parent.$parent.$parent.isNeedRefresh) {
        return
      }
      window.intervalFrom = 'custom-chart'
      const modalElement = document.querySelector('#edit-view')
      const maxViewElement = document.querySelector('#max-view-chart')
      const customChartEle = document.querySelector('#custome-chart-view')
      this.isChartInWindow = this.calcIsChartInWindow()
      if (type === 'mounted') {
        if (this.isInViewConfig) {
          if (this.chartInfo.parsedDisplayConfig.y < 14) {
            // 这里用在自定义视图中首屏渲染
            this.$emit('isChartInWindow')
            this.requestChartData()
            this.isAutoRefresh()
          }
        } else {
          // 这里用于其余的场景，首屏渲染全量
          this.requestChartData()
        }
      } else if (this.isChartInWindow && customChartEle && !modalElement && !maxViewElement && !this.isToolTipShow) {
        this.$emit('isChartInWindow')
        this.requestChartData()
      }
    },
    requestChartData() {
      this.noDataType = 'normal'
      if (this.chartInfo.chartParams.data.length === 0) {
        this.noDataType = 'noConfig'
        return
      }
      const params = {
        ...this.chartInfo.chartParams,
        custom_chart_guid: this.chartInfo.elId,
        start: dateToTimestamp(this.params.dateRange[0]),
        end: dateToTimestamp(this.params.dateRange[1]),
      }
      this.elId = this.chartInfo.elId
      window.viewTimeStepArr.push(+new Date() - window.startTimeStep + '$1015')
      this.request('POST',this.apiCenter.metricConfigView.api, params, responseData => {
        this.hasNotRequest = false
        window.viewTimeStepArr.push(+new Date() - window.startTimeStep + '$1016')
        if (responseData.legend.length === 0) {
          this.noDataType = 'noData'
        } else {
          this.chartInstance = null
          responseData.yaxis.unit = this.chartInfo.panalUnit
          this.noDataType = 'normal'
          const chartConfig = {
            title: false,
            eye: false,
            clear: true,
            dataZoom: false,
            lineBarSwitch: true,
            chartType: this.chartInfo.chartType,
            params: this.chartInfo.chartParams,
            chartId: this.chartInfo.elId
          }
          // this.$nextTick(() => {
          window.viewTimeStepArr.push(+new Date() - window.startTimeStep + '$1017')
          window['view-config-selected-line-data'][chartConfig.chartId] = window['view-config-selected-line-data'][chartConfig.chartId] || {}
          const metricList = chartConfig.params.data.map(one => one.metric)
          // 该逻辑是先筛选掉此时window中存在的需要删除的数据
          for (const key in window['view-config-selected-line-data'][chartConfig.chartId]) {
            if (metricList.length && !metricList.some(one => key.startsWith(`${one}:`))) {
              delete window['view-config-selected-line-data'][chartConfig.chartId][key]
            }
          }

          !isEmpty(chartConfig.params.data) && chartConfig.params.data.forEach(item => {
            if (isEmpty(item.series)) {
              item.series = []
              item.metricToColor = []
              const metric = item.metric
              responseData.legend.forEach(one => {
                if (one.startsWith(`${metric}:`)){
                  item.series.push({
                    seriesName: one,
                    new: true,
                    color: ''
                  })
                }
              })
              changeSeriesColor(item.series, item.colorGroup)
            }
            if (item.series && !isEmpty(item.series)) {
              if (isEmpty(window['view-config-selected-line-data'][chartConfig.chartId])
                  || (!isEmpty(window['view-config-selected-line-data'][chartConfig.chartId]) && Object.keys(window['view-config-selected-line-data'][chartConfig.chartId]).every(one => !one.startsWith(`${item.metric}:`)))
                  || (Object.keys(window['view-config-selected-line-data'][chartConfig.chartId]).filter(single => single.startsWith(`${item.metric}:`))).length !== item.series.length) {
                // 当widow中当前线条为空，或者不为空但是window中线条每个线条都不是以当前item.metric开头,或者两者length不一样，则进入该逻辑
                for (const key in window['view-config-selected-line-data'][chartConfig.chartId]) {
                  if (key.startsWith(`${item.metric}:`)) {
                    delete window['view-config-selected-line-data'][chartConfig.chartId][key]
                  }
                }
                item.series.forEach(one => {
                  window['view-config-selected-line-data'][chartConfig.chartId][one.seriesName] = true
                })
              }
              if (isEmpty(item.metricToColor)) {
                item.metricToColor = item.series.map(one => {
                  one.metric = one.seriesName
                  return one
                })
              }
            } else {
              item.metricToColor = []
            }
          })
          this.chartInstance = readyToDraw(this, responseData, this.chartIndex, chartConfig)
          window.viewTimeStepArr.push(+new Date() - window.startTimeStep + '$1018')
          if (window.viewTimeStepArr.length > 30) {
            window.viewTimeStepArr.length = 30
          }
          this.scrollHandle()
          if (this.chartInstance) {
            this.chartInstance.on('legendselectchanged', params => {
              window['view-config-selected-line-data'][chartConfig.chartId] = cloneDeep(params.selected)
            })
            this.chartInstance.on('showTip', () => {
              this.isToolTipShow = true
              const className = `.echarts-custom-tooltip-${chartConfig.chartId}`
              chartTooltipContain(className)
            })
            this.chartInstance.on('hideTip', () => {
              this.isToolTipShow = false
            })
          }
          // })
        }
      }, { isNeedloading: false })
    },
    onMouseLeaveContent() {
      if (this.chartInstance) {
        this.chartInstance.dispatchAction({
          type: 'hideTip'
        })
      }
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
    font-size: 14px;
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

<style lang="less">
// .echart> div[style]:nth-child(2){
//   pointer-events: all !important; /*强制tooltip响应事件*/
// }
// .echart > div[style]:nth-child(2):hover {
//   display: block !important; /*强制鼠标在时tooltip不消失*/
// }

</style>
