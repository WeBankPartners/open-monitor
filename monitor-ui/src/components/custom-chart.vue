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
import {readyToDraw} from '@/assets/config/chart-rely'

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
      chartInstance: null
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
        this.getchartdata()
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
    this.getchartdata()
    this.isAutoRefresh()
    window.addEventListener('scroll', this.scrollHandle, true)
    window.addEventListener('visibilitychange', this.isTabActive, true)
  },
  destroyed() {
    this.clearInterval()
    window.removeEventListener('scroll', this.scrollHandle, true)
    window.removeEventListener('visibilitychange', this.isTabActive, true)
    if (this.chartInstance) {
      setTimeout(() => {
        this.chartInstance.dispatchAction({
          type: 'hideTip'
        })
      }, 100)
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
    getchartdata() {
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
      window.intervalFrom = 'custom-chart'
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST',this.$root.apiCenter.metricConfigView.api, params, responseData => {
        if (responseData.legend.length === 0) {
          this.noDataType = 'noData'
        } else {
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
