<template>
  <div class="single-chart">
    <div v-if="!noDataTip" :id="elId" class="echart">
    </div>
    <div v-if="noDataTip" class="echart echart-no-data-tip">
      {{chartTitle}}:
      <span>~~~No Data!~~~</span>
    </div>
  </div>
</template>

<script>
import {generateUuid} from '@/assets/js/utils'

// 引入 ECharts 主模块
import {readyToDraw} from  '@/assets/config/chart-rely'
// const echarts = require('echarts/lib/echarts');

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
  created (){
    // 外部触发清除刷新
    this.$root.$eventBus.$on('clearSingleChartInterval', () => {
      clearInterval(this.interval)
    })
    generateUuid().then((elId)=>{
      this.elId =  `id_${elId}`; 
    })
  },
  watch: {
    params: function () {
      this.getChartData()
      if (this.params.autoRefresh > 0) {
        this.interval = setInterval(()=>{
          this.getChartData()
        },this.params.autoRefresh*1000)
      }
    }
  },
  mounted() {
    this.getChartData()
    if (this.params.autoRefresh > 0) {
      this.interval = setInterval(()=>{
        this.getChartData()
      },this.params.autoRefresh*1000)
    }
  },
  destroyed() {
    clearInterval(this.interval)
  },
  methods: {
    getChartData (tmp, start, end) {
      let params = []
      if (this.params.sys) {
        this.chartInfo.endpoint.forEach((ep) => {
          this.chartInfo.metric.forEach((me) => {
            params.push({
              id: this.chartInfo.id,
              endpoint: ep,
              metric: me,
              time: this.params.time.toString(),
              start: start ? start : this.params.start + '',
              end: end ? end : this.params.end + ''
            })
          })
        })
      } else {
        params.push({
          id: this.chartInfo.id,
          endpoint: this.chartInfo.endpoint[0],
          metric: this.chartInfo.metric[0],
          time: this.params.time.toString(),
          start: start ? start : this.params.start + '',
          end: end ? end : this.params.end + '',
          compare_first_start: this.params.compare_first_start,
          compare_first_end: this.params.compare_first_end,
          compare_second_start: this.params.compare_second_start,
          compare_second_end: this.params.compare_second_end
        })
      }
      this.$httpRequestEntrance.httpRequestEntrance('POST', 'dashboard/newchart', params, responseData => {
        const chartConfig = {editTitle: true, zoomCallback: true}
        responseData.metric = this.chartInfo.metric[0]
        readyToDraw(this,responseData, this.chartIndex, chartConfig)
      }, { isNeedloading: false })
    }
  },
  components: {},
}
</script>

<style scoped lang="less">
  .single-chart {
    display: inline-block;
    padding: 5px;
    .echart {
       height: 300px;
       width: 580px;
       border-radius: 4px;
      //  background: @gray-f;
    }
    .echart-no-data-tip {
      text-align: center;
      vertical-align: middle;
      display: table-cell;
    }
  }
</style>
