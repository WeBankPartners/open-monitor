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
import {isEmpty} from 'lodash'
import {readyToDraw} from '@/assets/config/chart-rely'
import { changeSeriesColor } from '@/assets/config/random-color'

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
      hasNotRequest: true
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
        this.isAutoRefresh()
        this.getchartdata()
      },
      deep: true
    },
    refreshNow: {
      handler() {
        if (this.isFirstRefresh) {
          this.getchartdata('mounted')
          setTimeout(() => {
            this.isFirstRefresh = false
          }, 5000)
        } else {
          this.getchartdata()
        }
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
    }
    this.isAutoRefresh()
    window.addEventListener('scroll', this.scrollHandle, true)
    window.addEventListener('visibilitychange', this.isTabActive, true)
  },
  destroyed() {
    this.clearInterval()
    this.chartInstance = null
    window.removeEventListener('scroll', this.scrollHandle, true)
    window.removeEventListener('visibilitychange', this.isTabActive, true)
    if (this.chartInstance) {
      setTimeout(() => {
        this.chartInstance.dispatchAction({
          type: 'hideTip'
        })
      }, 500)
    }
  },
  methods: {
    isTabActive() {
      if (document.hidden) {
        this.clearInterval()
      } else {
        this.isAutoRefresh()
      }
    },
    clearInterval() {
      clearInterval(this.interval)
      this.interval = null
    },
    scrollHandle() {
      const offset = this.$el.getBoundingClientRect()
      const offsetTop = offset.top
      const offsetBottom = offset.bottom
      // const offsetHeight = offset.height;
      // 进入可视区域
      if (offsetTop <= window.innerHeight && offsetBottom >= 0) {
        if (this.interval === null) {
          this.isAutoRefresh()
        }
      } else {
        clearInterval(this.interval)
        this.interval = null
      }
    },
    isAutoRefresh() {
      clearInterval(this.interval)
      const element = document.querySelector('#custome-chart-view')
      if (!element) {
        return
      }
      if (this.params.autoRefresh > 0 && this.params.dateRange[0] === '') {
        this.interval = setInterval(() => {
          this.getchartdata()
          this.isAutoRefresh()
        },this.params.autoRefresh * 1000)
      }
    },
    async getchartdata(type = '') {
      window.intervalFrom = 'custom-chart'
      const modalElement = document.querySelector('#edit-view')
      const offset = this.$el.getBoundingClientRect()
      const offsetTop = offset.top
      const offsetBottom = offset.bottom
      if (type === 'mounted') {
        if (this.isInViewConfig) {
          if (this.chartInfo.parsedDisplayConfig.y < 35) {
            // 这里用在自定义视图中首屏渲染
            this.requestChartData()
          }
        } else {
          // 这里用于其余的场景，首屏渲染全量
          this.requestChartData()
        }
      } else if ((offsetTop <= window.innerHeight && offsetBottom >= 0 && !modalElement)) {
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
        custom_chart_guid: this.chartInfo.elId
      }
      this.elId = this.chartInfo.elId
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST',this.$root.apiCenter.metricConfigView.api, params, responseData => {
        this.hasNotRequest = false
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
            params: this.chartInfo.chartParams
          }
          this.$nextTick(() => {
            !isEmpty(chartConfig.params.data) && chartConfig.params.data.forEach(item => {
              if (isEmpty(item.series)) {
                item.series = []
                item.metricToColor = []
                const metric = item.metric
                responseData.legend.forEach(one => {
                  if (one.startsWith(metric)){
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
                if (isEmpty(item.metricToColor)) {
                  item.metricToColor = item.series.map(one => {
                    one.metric = one.seriesName
                    delete one.seriesName
                    return one
                  })
                }
              } else {
                item.metricToColor = []
              }
            })
            this.chartInstance = readyToDraw(this, responseData, this.chartIndex, chartConfig)
            this.scrollHandle()
          })
        }
      }, { isNeedloading: false })
    },
    onMouseLeaveContent() {
      if (this.chartInstance) {
        this.chartInstance.dispatchAction({
          type: 'hideTip'
        })
      }
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
