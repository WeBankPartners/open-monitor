<template>
  <div class="single-chart">
    <div v-if="!noDataTip" :id="elId" class="echart">
    </div>
    <div v-if="noDataTip" class="echart echart-no-data-tip">
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
  created (){
    generateUuid().then((elId)=>{
      this.elId =  `id_${elId}`; 
    })
  },
  mounted() {
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
      let params = {
        id: this.chartItemx.id,
        endpoint: [this.params.endpoint.split(':')[0]],
        metric: [this.chartItemx.metric[0]],
        time: this.params.time.toString(),
        start: this.params.start + '',
        end: this.params.end + ''
      }

      this.$httpRequestEntrance.httpRequestEntrance('GET', this.chartItemx.url, params, responseData => {
        // var legend = []
        // const colorx = ['#61a0a8', '#2f4554', '#c23531', '#d48265', '#91c7ae', '#749f83', '#ca8622', '#bda29a', '#6e7074', '#546570', '#c4ccd3']
        // const colorSet = colorx.concat(colorx,colorx,colorx,colorx).splice(this.chartIndex*3 + 4)
        // if (responseData.series.length === 0) {
        //   this.noDataTip = true
        //   return
        // }
        // responseData.series.forEach((item,index)=>{
        //   legend.push(item.name)
        //   item.symbol = 'none'
        //   item.smooth = true
        //   item.lineStyle = {
        //     width: 1
        //   }
        //   item.itemStyle = {
        //     normal:{
        //       color: colorSet[index]
        //     }
        //   }
        //   item.areaStyle = {
        //     normal: {
        //       color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [{
        //           offset: 0,
        //           color: colorSet[index]
        //       }, {
        //           offset: 1,
        //           color: 'white'
        //       }])
        //     }
        //   }
        // }) 
        // let config = {
        //   title: responseData.title,
        //   legend: legend,
        //   series: responseData.series,
        //   yaxis: responseData.yaxis,
        // }
        // this.config = config
        // drawChart(this, config)

        // responseData.yaxis.unit =  panalUnit  
        // this.elId = id
        // const chartConfig = {eye: false,dataZoom:false}

        readyToDraw(this,responseData, this.chartIndex)
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
