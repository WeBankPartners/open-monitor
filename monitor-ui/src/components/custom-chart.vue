<template>
  <div class="single-chart">
    <div v-if="!noDataTip">
      <div :id="elId" class="echart" :style="chartInfo.style">
    </div>
    </div>
    <div v-if="noDataTip" class="echart echart-no-data-tip">
      <span>{{this.$t('m_nodata_tips')}}</span>
    </div>
  </div>
</template>

<script>
// 引入 ECharts 主模块
import {readyToDraw} from  '@/assets/config/chart-rely'

export default {
  name: '',
  data() {
    return {
      elId: null,
      chartTitle: null,
      noDataTip: false,
      config: '',
      myChart: '',
      interval: ''
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
      handler () {
        this.getchartdata()
      },
      deep: true
    },
    refreshNow: {
      handler () {
        this.getchartdata();
      }
    }
  },
  mounted() {
    this.getchartdata();
    this.isAutoRefresh();
    window.addEventListener("scroll", this.scrollHandle, true)
    window.addEventListener("visibilitychange", this.isTabActive, true)
  },
  destroyed() {
    this.clearInterval()
    window.removeEventListener('scroll', this.scrollHandle, true)
    window.removeEventListener("visibilitychange", this.isTabActive, true)
  },
  methods: {
    isTabActive () {
       if (document.hidden) {
        this.clearInterval()
      } else {
        this.isAutoRefresh()
      }
    },
    clearInterval () {
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
    isAutoRefresh () {
      clearInterval(this.interval)
      if (this.params.autoRefresh > 0 && this.params.dateRange[0] === '') {
        this.interval = setInterval(()=>{
          this.getchartdata();
          this.isAutoRefresh()
        },this.params.autoRefresh * 1000)
      }
    },
    getchartdata () {
      if (this.chartInfo.chartParams.data.length === 0) {
        return
      }
      const params = {
        ...this.chartInfo.chartParams,
        custom_chart_guid: this.chartInfo.elId
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST',this.$root.apiCenter.metricConfigView.api, params, responseData => {
        if (responseData.legend.length === 0) {
          this.noDataTip = true
        } else {
          responseData.yaxis.unit =  this.chartInfo.panalUnit  
          this.elId = this.chartInfo.elId
          this.noDataTip = false
          const chartConfig = {eye: false,clear: true,dataZoom:false, lineBarSwitch: true, chartType: this.chartInfo.chartType, params: this.chartInfo.chartParams};
          this.$nextTick( () => {
            readyToDraw(this, responseData, this.chartIndex, chartConfig)
            this.scrollHandle()
          })
        }
        
      }, { isNeedloading: false })
    }
  },
  components: {},
}
</script>

<style scoped lang="less">
  .single-chart {
    font-size: 14px;
    padding: 5px;
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
