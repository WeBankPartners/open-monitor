<template>
  <div class="single-chart">
    <div v-if="!noDataTip" :id="elId" class="echart"></div>
    <div v-if="noDataTip" class="echart echart-no-data-tip">
    <span class="no-data-tip">~~~暂无数据~~~</span>
    </div>
  </div>
</template>

<script>
import {generateUuid} from '@/assets/js/utils'
// 引入 ECharts 主模块
var echarts = require('echarts/lib/echarts');
// 引入柱状图
require('echarts/lib/chart/line');
// 引入提示框和标题组件
require('echarts/lib/component/tooltip');
require('echarts/lib/component/title');
require('echarts/lib/component/legend');
require('echarts/lib/component/toolbox');
require('echarts/lib/component/legendScroll');
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
    console.log(this.params.autoRefresh)
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
        this.draw()
      })
    },
    draw () {
      // 基于准备好的dom，初始化echarts实例
      this.myChart = echarts.init(document.getElementById(this.elId))
      // 绘制图表
      this.myChart.setOption(
        {
          title: {
            text: this.config.title,
            left:'10%',
            top: '10px'
          },
          tooltip: {
            trigger: 'axis',
            backgroundColor: 'rgba(245, 245, 245, 0.8)',
            borderWidth: 1,
            borderColor: '#ccc',
            padding: 10,
            textStyle: {
              color: '#000'
            },
            formatter: (params)=>{ 
              var str =''
              let date = new Date(params[0].data[0])
              // let year =  date.getFullYear()
              // let month = (date.getMonth() + 1)>=10?(date.getMonth() + 1):'0'+(date.getMonth() + 1)
              // let day = date.getDate()>=10?date.getDate():'0'+date.getDate()
              let hours = date.getHours()>=10?date.getHours():'0'+date.getHours()
              let minutes = date.getMinutes()>=10?date.getMinutes():'0'+date.getMinutes()
              let seconds = date.getSeconds()>=10?date.getSeconds():'0'+date.getSeconds()
              str=hours+':'+minutes+':'+seconds
              var res = `<div>${str}</div>`
              params.forEach(item=>{
                res = res+`<div><div style=' display: inline-block;width: 10px;
                height: 10px;border: 1px solid transparent;border-radius:50%;
                background-color:${item.color};'  ></div> ${item.seriesName}
                ${Math.floor(item.data[1] * 1000) / 1000}</div>`
              })
              return res
            },
          },  
          toolbox: {
            right: '4%',
            top: '10px',
            feature: {
              dataZoom: {
                  yAxisIndex: 'none'
              },
              myTool:{
                show:true,
                title:'查看全部',
                icon: 'path://M432.45,595.444c0,2.177-4.661,6.82-11.305,6.82c-6.475,0-11.306-4.567-11.306-6.82s4.852-6.812,11.306-6.812C427.841,588.632,432.452,593.191,432.45,595.444L432.45,595.444z M421.155,589.876c-3.009,0-5.448,2.495-5.448,5.572s2.439,5.572,5.448,5.572c3.01,0,5.449-2.495,5.449-5.572C426.604,592.371,424.165,589.876,421.155,589.876L421.155,589.876z M421.146,591.891c-1.916,0-3.47,1.589-3.47,3.549c0,1.959,1.554,3.548,3.47,3.548s3.469-1.589,3.469-3.548C424.614,593.479,423.062,591.891,421.146,591.891L421.146,591.891zM421.146,591.891',   
                onclick: () => {
                  this.$emit('sendConfig', this.chartItemx)
                }
              }
            }
          },
          legend: {
            type: 'scroll',
            y: 'bottom',
            padding: 10,
            orient: 'horizontal',
            data: this.config.legend
          },
          calculable: false,
          color: ['#7EB26D', '#EAB839', '#6ED0E0', '#EF843C', '#E24D42', '#1F78C1', '#BA43A9', '#705DA0', '#508642', '#CCA300', '#447EBC', '#C15C17'],
          grid: {
            left: '3%',
            right: '5%',
            bottom: '8%' ,
            containLabel: true
          },
          // dataZoom: [{
          //   type: 'inside',
          // }],
          xAxis: {
            type: 'time',
            axisLabel: {
            formatter: function (value) {
              return echarts.format.formatTime('MM-dd\nhh:mm:ss', value)
            }
            },
            boundaryGap : false,
            splitLine: {
              show: true
            }
          },
          yAxis: [
            {
              type: 'value',
              axisLabel: {
                show: true,
                interval: 'auto',
                formatter: (value) => {
                  let unit = ''
                  if (value > 1024*1024*1024*1024) {
                    value = value / (1024*1024*1024*1024)  
                    unit = 'T'
                  } else if (value > 1024*1024*1024) {
                    value = value / (1024*1024*1024)  
                    unit = 'G'
                  } else if (value > 1024*1024) {
                    value = value / (1024*1024)  
                    unit = 'M'
                  } else if (value > 1024) {
                    value = value / 1024  
                    unit = 'K'
                  } else {
                    return value + ' ' + this.config.yaxis.unit
                  }
                  let newValue = Number.isInteger(value) ? value : value.toFixed(3)
                  return newValue + ' ' + unit + this.config.yaxis.unit
                }
              },
              show: true
            }
          ],
          series: this.config.series
        }
      )
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
        this.draw()
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
