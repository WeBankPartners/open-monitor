<template>
  <div class="single-chart">
    <div v-if="!noDataTip" :id="elId" class="echart" :style="chartInfo.style">
    </div>
    <div v-if="noDataTip" class="echart echart-no-data-tip">
      {{chartTitle}}:
      <span>~~~No Data!~~~</span>
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
    chartIndex: Number
  },
  watch: {
    params: {
      handler () {
        this.getchartdata()
      },
      deep: true
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
      this.isAutoRefresh()
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST',this.$root.apiCenter.metricConfigView.api, this.chartInfo.chartParams, responseData => {
        responseData.yaxis.unit =  this.chartInfo.panalUnit  
        this.elId = this.chartInfo.elId
        const chartConfig = {eye: false,dataZoom:false, lineBarSwitch: true, chartType: this.chartInfo.chartType, params: this.chartInfo.chartParams}
        this.$nextTick( () => {
          readyToDraw(this,responseData, this.chartIndex, chartConfig)
        })
      }, { isNeedloading: false })
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
    }
  }
</style>
