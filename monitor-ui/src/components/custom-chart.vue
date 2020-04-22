<template>
  <div class="single-chart">
    {{interval}}
    <div v-if="!noDataTip" :id="elId" class="echart" :style="chartItemx.style">
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
    chartItemx: Object,
    params: Object,
    chartIndex: Number
  },
  // watch: {
  //   params: function () {
  //     this.getchartdata()
  //     if (this.params.autoRefresh > 0) {
  //       this.interval = setInterval(()=>{
  //         this.getchartdata()
  //       },this.params.autoRefresh*1000)
  //     }
  //   }
  // },
  mounted() {
    this.elId = this.chartItemx.elId
    this.getchartdata()
    if (this.params.autoRefresh > 0) {
      this.interval = setInterval(()=>{
        this.getchartdata()
      },this.params.autoRefresh*1000)
    }
  },
  destroyed() {
    clearInterval(this.interval)
  },
  methods: {
    getchartdata () {
      let params = []
      this.chartItemx.chartParams.forEach((item) => {
        params.push({
          ...item,
          time: this.params.timeTnterval + '',
          start: this.params.dateRange[0] ===''? '':Date.parse(this.params.dateRange[0])/1000+'',
          end: this.params.dateRange[1] ===''? '':Date.parse(this.params.dateRange[1])/1000+'',
        })
      })

      this.$root.$httpRequestEntrance.httpRequestEntrance('POST',this.$root.apiCenter.metricConfigView.api, params, responseData => {
        responseData.yaxis.unit =  this.chartItemx.panalUnit  
        this.elId = this.chartItemx.elId + ((new Date()).valueOf()).toString().substring(10) 
        const chartConfig = {eye: false,dataZoom:false, lineBarSwitch: true}
        this.$nextTick( () => {
          readyToDraw(this,responseData, this.chartIndex, chartConfig)
        })
      })
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
