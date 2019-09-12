<template>
  <div class=" ">
       <div class="max-chart">
         <div class="hiddenBtn" @click="hideMaxChart">
            <i class="fa fa-angle-right" aria-hidden="true"></i>
         </div>
        <div class="condition-zone">
          <ul>
            <li>
              <div class="condition condition-title">时间段</div>
              <div class="condition">
                <RadioGroup v-model="chartCondition.timeTnterval" size="small" type="button">
                  <Radio label="-1800">30分钟</Radio>
                  <Radio label="-3600">1小时</Radio>
                  <Radio label="-10800">3小时</Radio>
                </RadioGroup>
              </div>
            </li>
            <li>
              <div class="condition condition-title">时间区间</div>
              <div class="condition">
                <DatePicker type="daterange" placement="bottom-end" @on-change="datePick" placeholder="请选择日期" style="width: 200px"></DatePicker>
              </div>
            </li>
            <li>
              <div class="condition condition-title">聚合类型</div>
              <div class="condition">
                <RadioGroup v-model="chartCondition.agg" size="small" type="button">
                  <Radio label="min">最小值</Radio>
                  <Radio label="max">最大值</Radio>
                  <Radio label="avg">平均值</Radio>
                  <Radio label="p95">P95值</Radio>
                  <Radio label="none">原始值</Radio>
                </RadioGroup>
              </div>
            </li>
          </ul>
        </div>
        <div class="chart-zone" >
          <div :id="elId" class="echart" style="height:400px;width:600px"></div>
        </div>
      </div>
  </div>
</template>

<script>
// 引入 ECharts 主模块
var echarts = require('echarts/lib/echarts');
// 引入柱状图
require('echarts/lib/chart/line');
// 引入提示框和标题组件
require('echarts/lib/component/tooltip');
require('echarts/lib/component/title');
require('echarts/lib/component/legend');
require('echarts/lib/component/toolbox');
export default {
  name: '',
  data() {
    return {
      chartItem: {},
      elId: null,
      chartCondition: {
        timeTnterval: "-1800",
        dateRange: '',
        agg: 'none' // 聚合类型
      },
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
    this.elId =  `id_${this.guid()}`;
  },
  methods: {
    datePick (data) {
      this.chartCondition.dateRange = data
    },
    guid() {
      return 'xxxxxxxx_xxxx_4xxx_yxxx_xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
          var r = Math.random()*16|0, v = c == 'x' ? r : (r&0x3|0x8);
          return v.toString(16);
      })
    },
    draw (config) {
      // 基于准备好的dom，初始化echarts实例
      var myChart = echarts.init(document.getElementById(this.elId));

      // 绘制图表
      myChart.setOption(
        {
          title: {
            text: config.title,
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
                background-color:${item.color};'  ></div> ${item.seriesName} <br>  &nbsp; &nbsp;
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
            y: 'bottom', 
            data: config.legend
          },
          calculable: false,
          color: ['#7EB26D', '#EAB839', '#6ED0E0', '#EF843C', '#E24D42', '#1F78C1', '#BA43A9', '#705DA0', '#508642', '#CCA300', '#447EBC', '#C15C17'],
          grid: {
            left: '3%',
            right: '5%',
            bottom: '8%' ,
            containLabel: true
          },
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
                    return value + ' ' + config.yaxis.unit
                  }
                  let newValue = Number.isInteger(value) ? value : value.toFixed(3)
                  return newValue + ' ' + unit + config.yaxis.unit
                }
              },
              show: true
            }
          ],
          series: config.series
        }
      )
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
        this.draw(config)
      })
    },
    hideMaxChart () {
      this.$parent.showMaxChart = false
    }
  },
  components: {},
}
</script>
<style>
 
</style>
<style scoped lang="less">
  .max-chart {
    width:610px;
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
    width: 100px;
    text-align: center;
    vertical-align: middle;
    margin: 4px 8px 4px 0;
    padding: 3px;
  }

  .chart-zone {
    margin-top: 12px;
  }
</style>

