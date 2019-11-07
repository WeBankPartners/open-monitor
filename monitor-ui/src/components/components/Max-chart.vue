<template>
  <div class=" ">
       <div class="max-chart">
         <div class="hiddenBtn" @click="hideMaxChart">
            <i class="fa fa-angle-right" aria-hidden="true"></i>
         </div>
        <div class="condition-zone">
          <ul>
            <li>
              <div class="condition condition-title">{{$t('field.relativeTime')}}</div>
              <div class="condition">
                <RadioGroup v-model="chartCondition.timeTnterval" size="small" type="button">
                  <Radio label="-1800">30m</Radio>
                  <Radio label="-3600">1h</Radio>
                  <Radio label="-10800">3h</Radio>
                </RadioGroup>
              </div>
            </li>
            <li>
              <div class="condition condition-title">{{$t('field.timeInterval')}}</div>
              <div class="condition">
                <DatePicker type="daterange" placement="bottom-end" @on-change="datePick" :placeholder="$t('placeholder.datePicker')" style="width: 200px"></DatePicker>
              </div>
            </li>
            <li>
              <div class="condition condition-title">{{$t('field.aggType')}}</div>
              <div class="condition">
                <RadioGroup v-model="chartCondition.agg" size="small" type="button">
                  <Radio label="min">Min</Radio>
                  <Radio label="max">Max</Radio>
                  <Radio label="avg">Average </Radio>
                  <Radio label="p95">P95</Radio>
                  <Radio label="none">Original</Radio>
                </RadioGroup>
              </div>
            </li>
          </ul>
        </div>
        <div class="chart-zone" >
          <div :id="elId" class="echart" :style="chartStyle"></div>
        </div>
      </div>
  </div>
</template>

<script>
import {generateUuid} from '@/assets/js/utils'
// 引入 ECharts 主模块
import {drawChart} from  '@/assets/config/chart-rely'
const echarts = require('echarts/lib/echarts');
export default {
  name: '',
  data() {
    return {
      chartItem: {},
      elId: null,
      chartCondition: {
        timeTnterval: "-1800",
        dateRange: ['',''],
        agg: 'none' // 聚合类型
      },
      chartStyle: {
        minHeight: '400px',
        minWidth: '700px'
      }
    }
  },
  watch: {
    chartCondition: {
      handler: function () {
        this.getChartConfig()
      },
      deep: true
    }
  },
  created (){
    generateUuid().then((elId)=>{
      this.elId =  `id_${elId}`; 
      this.chartStyle.width = window.screen.width * 0.6 + 'px'
      this.chartStyle.height = window.screen.height * 0.4 + 'px'
    })

  },
  methods: {
    datePick (data) {
      this.chartCondition.dateRange = data
      if (this.chartCondition.dateRange[0] !== '') {
        this.chartCondition.dateRange[0] = this.chartCondition.dateRange[0] + ' 00:00:01'
      }
      if (this.chartCondition.dateRange[1] !== '') {
        this.chartCondition.dateRange[1] = this.chartCondition.dateRange[1] + ' 23:59:59'
      }
      this.getChartConfig()
    },
    getChartConfig (chartItem=this.chartItem) {
      this.chartItem = chartItem
      let params = {
        id: chartItem.id,
        endpoint: [chartItem.endpoint[0]],
        metric: [chartItem.metric[0]],
        time: this.chartCondition.timeTnterval,
        agg: this.chartCondition.agg,
      }
      if (this.chartCondition.dateRange.length !==0) {
        params.start = this.chartCondition.dateRange[0] ===''? '':Date.parse(this.chartCondition.dateRange[0])/1000 + '',
        params.end = this.chartCondition.dateRange[1] ===''? '':Date.parse(this.chartCondition.dateRange[1])/1000 + ''
      }
      this.$httpRequestEntrance.httpRequestEntrance('GET', chartItem.url, params, responseData => {
        var legend = []
        const colorSet = ['#CC9999','#CCCCFF','#FF66FF','#FFCCCC','#66CCCC','#996699','#99CC66','#FFCC99','#CC99CC','#FF9999','#FF9966','#FFCCFF']
        responseData.series.forEach((item,index)=>{
          legend.push(item.name)
          item.symbol = 'none'
          item.smooth = true
          item.lineStyle = {
            width: 1
          }
          item.itemStyle = {
            normal:{
              color: colorSet[index]
            }
          }
          item.areaStyle = {
            normal: {
              color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [{
                  offset: 0,
                  color: colorSet[index]
              }, {
                  offset: 1,
                  color: 'white'
              }])
            }
          }
        }) 
        let config = {
          title: responseData.title,
          legend: legend,
          series: responseData.series,
          yaxis: responseData.yaxis,
        }
        drawChart(this,config,{eye:false})
      })
    },
    hideMaxChart () {
      this.$parent.showMaxChart = false
    }
  },
  components: {},
}
</script>

<style scoped lang="less">
  .max-chart {
    // width:800px;
    min-height: 540px;
    height: 123vh;
    background: white;
    position: absolute;
    border: 1px solid @blue-lingt;
    right: 0;
    top: 60px;
    z-index: 2;
    padding: 12px;
  }

  .hiddenBtn {
    position: absolute;
    top: 50%;
    left: 0;
    width: 12px;
    padding: 8px 0;
    // height: 20px;
    text-align: center;
    background: @blue-lingt;
    i {
      font-size: 16px;
      color: white;
    }
  }
  .condition {
    display: inline-block;
  }
  .condition /deep/ .ivu-input {
    height: 24px;
  }
  .condition /deep/ .ivu-input-suffix i {
    line-height: 24px;
  }
  .condition-title {
    background: @gray-d;
    width: 110px;
    text-align: center;
    vertical-align: middle;
    margin: 4px 8px 4px 0;
    padding: 3px;
  }

  .chart-zone {
    margin-top: 12px;
  }
</style>

