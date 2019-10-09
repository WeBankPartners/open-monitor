<template>
  <div class="single-chart">
    <div v-if="!noDataTip" :id="elId" class="echart"></div>
    <div v-if="noDataTip" class="echart echart-no-data-tip">
      <span>~~~暂无数据~~~</span>
    </div>
  </div>
</template>

<script>
import {generateUuid} from '@/assets/js/utils'

// 引入 ECharts 主模块
import {drawChart} from  '@/assets/config/chart-rely'

export default {
  name: '',
  data() {
    return {
      elId: null,
      noDataTip: false,
      config: '',
      myChart: '',
      interval: ''
    }
  },
  props: {
    chartItemx: Object,
    params: Object,
  },
  created (){
    generateUuid().then((elId)=>{
      this.elId =  `id_${elId}`; 
    })
  },
  mounted() {
    this.getchartdata()
    if (this.params.autoRefresh > 0) {
      this.interval = setInterval(()=>{
        this.refreshChart()
      },this.params.autoRefresh*1000)
    }
  },
  destroyed() {
    clearInterval(this.interval)
  },
  methods: {
    refreshChart() {
      let params = {
        id: this.chartItemx.id,
        endpoint: [this.params.endpoint.split(':')[0]],
        metric: [this.chartItemx.metric[0]],
        time: this.params.time.toString(),
        start: this.params.start + '',
        end: this.params.end + ''
      }
      this.$httpRequestEntrance.httpRequestEntrance('GET', this.chartItemx.url, params, responseData => {
        this.config.series = responseData.series
         drawChart(this, this.config)
      })
    },
    getchartdata () {
      let params = {
        id: this.chartItemx.id,
        endpoint: [this.params.endpoint.split(':')[0]],
        metric: [this.chartItemx.metric[0]],
        time: this.params.time.toString(),
        start: this.params.start + '',
        end: this.params.end + ''
      }
      this.$httpRequestEntrance.httpRequestEntrance('GET', this.chartItemx.url, params, responseData => {
        var legend = []
        if (responseData.series.length === 0) {
          this.noDataTip = true
          return
        }
        responseData.series.forEach((item)=>{
          legend.push(item.name)
          item.symbol = 'none'
          item.smooth = true
          item.lineStyle = {
            width: 1
          }
        }) 
        let config = {
          title: responseData.title,
          legend: legend,
          series: responseData.series,
          yaxis: responseData.yaxis,
        }
        this.config = config
        drawChart(this, config)
      })
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
       background: @gray-f;
    }
    .echart-no-data-tip {
      text-align: center;
      vertical-align: middle;
      display: table-cell;
    }
  }
</style>
