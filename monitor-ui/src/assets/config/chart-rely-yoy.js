// 引入柱状图
require('echarts/lib/chart/line');
require('echarts/lib/chart/pie');
require('echarts/lib/chart/bar');
// 引入提示框和标题组件
require('echarts/lib/component/tooltip');
require('echarts/lib/component/title');
require('echarts/lib/component/legend');
require('echarts/lib/component/toolbox');
require('echarts/lib/component/legendScroll');

import { generateAdjacentColors } from './random-color'
const echarts = require('echarts/lib/echarts');

export const readyToDraw = function(that, responseData, viewIndex, chartConfig, elId) {
  var legend = []
  
  responseData.series.forEach((item, index)=>{
    legend.push(item.name)
    item.symbol = 'none'
    item.smooth = false
    item.lineStyle = {
      width: 1
    }
  }) 
  let config = {
    ...responseData,
    legend: legend
  }
  drawChart(that, config, chartConfig, elId)
}

export const drawChart = function(that,config,userConfig, elId) {
  const chartTextColor = '#a1a1a2'
  // 基于准备好的dom，初始化echarts实例
  var myChart = echarts.init(document.getElementById(elId || that.elId))
  myChart.resize()
  let option = {
    tooltip: {
      trigger: 'axis',
      backgroundColor: 'rgba(245, 245, 245, 0.8)',
      borderWidth: 1,
      borderColor: '#ccc',
      padding: 10,
      confine: true, // tip控制在图像区内
      textStyle: {
        color: '#000'
      },
      // extraCssText:'width:160px;height:40px;background: red;',
      formatter: (params)=>{ 
        var str =''
        let date = new Date(params[0].data[0])
        let hours = date.getHours()>=10?date.getHours():'0'+date.getHours()
        let minutes = date.getMinutes()>=10?date.getMinutes():'0'+date.getMinutes()
        let seconds = date.getSeconds()>=10?date.getSeconds():'0'+date.getSeconds()
        str=hours+':'+minutes+':'+seconds
        var res = `<div>${str}</div>`
        params.forEach(item => {
          let str = item.seriesName
          let step = 100
          let strLen = str.length
          let arr = []
          for(var i=0; i<strLen; i=i+step){
              arr.push(str.substr(i, step));
          }
          arr.join(" ");
          const seriesName = arr.join('<br>')
          let valueDisplay = item.componentSubType === 'line' ? (Math.floor(item.data[1] * 1000) / 1000) : (item.data[1] + '%')
          res = res+`<div><div style=' display: inline-block;width: 10px; 
          height: 10px;border: 1px solid transparent;border-radius:50%;
          background-color:${item.color};'  ></div>${valueDisplay} ${seriesName}
          </div>`
        })
        return res
      },
    },
    legend: {
      textStyle: {
        color: chartTextColor          // 图例文字颜色
      },
      type: 'scroll',
      y: 'bottom',
      padding: 10,
      orient: 'horizontal',
      data: config.legend
    },
    calculable: false,
    grid: {
      top: '40',
      left: '3%',
      right: '5%',
      bottom: '40' ,
      containLabel: true
    },
    xAxis: {
      type: 'time',
      axisLabel: {
        textStyle: {
          color: chartTextColor
        },
        formatter: function (value) {
          return echarts.format.formatTime('MM-dd\nhh:mm:ss', value)
        }
      },
      boundaryGap : false,
      axisLine:{
        lineStyle:{
          color:'#a1a1a2'
        }
      }, 
      splitLine: {
        show: true,
        lineStyle:{
          color: ['#a1a1a2'],
          width: 1,
          type: 'solid'
        }
      }
    },
    yAxis: [
      {
        type: 'value',
        alignTicks: true,
        axisLabel: {
          textStyle: {
            color: chartTextColor
          },
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
        show: true,
        axisLine:{
          lineStyle:{
            color:'#a1a1a2'
          }
        }, 
        splitLine: {
          show: true,
          lineStyle:{
            color: ['#a1a1a2'],
            width: 1,
           type: 'solid'
          }
        }
      },
      {
        type: 'value',
        axisLabel: {
          textStyle: {
            color: chartTextColor
          },
          show: true,
          interval: 'auto',
          formatter: (value) => {
            return value + '%'
          }
        },
        show: true,
        axisLine:{
          lineStyle:{
            color:'#a1a1a2'
          }
        }, 
        splitLine: {
          show: true,
          lineStyle:{
            color: ['#a1a1a2'],
            width: 1,
           type: 'solid'
          }
        }
      }
    ],
    series: config.series
  }
  // 绘制图表
  myChart.clear();
  myChart.setOption(option)
  // 清空所有事件重新绑定
  myChart.off();
}