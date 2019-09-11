<template>
  <div class="single-chart">
    <div :id="elId" class="echart"></div>
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
export default {
  name: '',
  data() {
    return {
      elId: null
    }
  },
  props: {
    chartItemx: Object,
    params: Object,
  },
  created (){
    this.elId =  `id_${this.guid()}`;
  },
  mounted() {
    this.getchartdata()
  },
  methods: {
    draw (config) {
      // 基于准备好的dom，初始化echarts实例
      var myChart = echarts.init(document.getElementById(this.elId));
      // 绘制图表
      myChart.setOption({
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
        legend: {
          y: 'bottom', 
          data: config.legend
        },
        calculable: false,
        color: ['#7EB26D', '#EAB839', '#6ED0E0', '#EF843C', '#E24D42', '#1F78C1', '#BA43A9', '#705DA0', '#508642', '#CCA300', '#447EBC', '#C15C17'],
        dataZoom: [{
          type: 'inside',
          // throttle: 1000,
          // start: 60,
          // end: 100
        }],
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
      });
    },
    guid() {
      return 'xxxxxxxx_xxxx_4xxx_yxxx_xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
          var r = Math.random()*16|0, v = c == 'x' ? r : (r&0x3|0x8);
          return v.toString(16);
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
  }
</style>
