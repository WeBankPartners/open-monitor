<template>
  <div class="single-chart">
    <div v-show="!noDataTip" :id="elId" class="echart" :style="chartInfo.style">
    </div>
    <div v-show="noDataTip" class="echart echart-no-data-tip">
      <span>{{this.$t('m_nodata_tips')}}</span>
    </div>
  </div>
</template>

<script>
// 引入 ECharts 主模块
import {drawPieChart} from  '@/assets/config/chart-rely'

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
        this.isAutoRefresh()
      },
      deep: true
    },
    refreshNow: {
      handler () {
        this.getchartdata()
      }
    }
  },
  mounted() {
    this.getchartdata()
  },
  destroyed() {
    clearInterval(this.interval)
  },
  methods: {
    isAutoRefresh () {
      clearInterval(this.interval)
      if (this.params.autoRefresh > 0 && this.params.dateRange[0] === '') {
        this.interval = setInterval(()=>{
          this.getchartdata()
        },this.params.autoRefresh*1000)
      }
    },
    getchartdata () {
      this.noDataTip = false
      if (this.chartInfo.chartParams.data.length === 0) {
        this.noDataTip = true
        return
      }
      this.isAutoRefresh()
      let params = this.chartInfo.chartParams.data
      params.forEach(p => {
        p.start = this.chartInfo.start
        p.end = this.chartInfo.end
        p.time_second = this.chartInfo.time_second,
        p.custom_chart_guid = this.chartInfo.elId
      });
      this.elId = this.chartInfo.elId
      this.$root.$httpRequestEntrance.httpRequestEntrance(
        'POST',this.$root.apiCenter.metricConfigPieView.api, params,
        responseData => {
          if (responseData.legend && responseData.legend.length === 0) {
            this.noDataTip = true
          } else {
            this.noDataTip = false
            drawPieChart(this, responseData)
          }
        },
        { isNeedloading: false }
      )
    }
  },
  components: {},
}
</script>

<style scoped lang="less">
  .single-chart {
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
