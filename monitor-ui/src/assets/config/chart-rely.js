// 引入柱状图
require('echarts/lib/chart/line');
// 引入提示框和标题组件
require('echarts/lib/component/tooltip');
require('echarts/lib/component/title');
require('echarts/lib/component/legend');
require('echarts/lib/component/toolbox');
require('echarts/lib/component/legendScroll');

const echarts = require('echarts/lib/echarts');

export const readyToDraw = function(that, responseData, viewIndex, chartConfig) {
  var legend = []
  if (responseData.series.length === 0) {
    that.chartTitle = responseData.title
    that.noDataTip = true
    return
  }
  const colorX = ['#33CCCC','#666699','#66CC66','#996633','#9999CC','#339933','#339966','#663333','#6666CC','#336699','#3399CC','#33CC66','#CC3333','#CC6666','#996699','#CC9933']
  let colorSet = []
  for (let i=0;i<colorX.length;i++) {
    let tmpIndex = viewIndex*3 + i
    tmpIndex = tmpIndex%colorX.length
    colorSet.push(colorX[tmpIndex])
  }
  responseData.series.forEach((item, index)=>{
    legend.push(item.name)
    item.symbol = 'none'
    item.smooth = true
    item.lineStyle = {
      width: 1
    }
    // 堆叠区域图开启
    // item.areaStyle = {}

    // 渐变图像开启
    item.itemStyle = {
      normal:{
        color: colorSet[index] ? colorSet[index] : '#666699' 
      }
    }
    item.areaStyle = {
      normal: {
        color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [{
            offset: 0,
            color: colorSet[index] ? colorSet[index] : '#666699' 
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
    yaxis: responseData.yaxis
  }
  drawChart(that, config, chartConfig)
}

export const drawChart = function(that,config,userConfig) {
  const chartTextColor = '#a1a1a2'
  let originConfig = {
    title: true,
    eye: true,
    dataZoom: true,
    clear: false
  }
  let finalConfig = Object.assign(originConfig, userConfig)

  // 基于准备好的dom，初始化echarts实例
  var myChart = echarts.init(document.getElementById(that.elId))
  if (originConfig.clear) {
    myChart.clear()
  }
  let option = {
    title: {
      textStyle: {
        fontSize: 16,
        fontWeight: 'bolder',
        color: chartTextColor          // 主标题文字颜色
      },
      // text: config.title,
      left:'10%',
      top: '10px'
    },
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
        // dataZoom: {
        //     yAxisIndex: 'none'
        // },
        // myTool:{
        //   show:true,
        //   title:'查看全部',
        //   icon: 'path://M432.45,595.444c0,2.177-4.661,6.82-11.305,6.82c-6.475,0-11.306-4.567-11.306-6.82s4.852-6.812,11.306-6.812C427.841,588.632,432.452,593.191,432.45,595.444L432.45,595.444z M421.155,589.876c-3.009,0-5.448,2.495-5.448,5.572s2.439,5.572,5.448,5.572c3.01,0,5.449-2.495,5.449-5.572C426.604,592.371,424.165,589.876,421.155,589.876L421.155,589.876z M421.146,591.891c-1.916,0-3.47,1.589-3.47,3.549c0,1.959,1.554,3.548,3.47,3.548s3.469-1.589,3.469-3.548C424.614,593.479,423.062,591.891,421.146,591.891L421.146,591.891zM421.146,591.891',   
        //   onclick: () => {
        //     that.$emit('sendConfig', that.chartItemx)
        //   }
        // }
      }
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
    // color: ['#7EB26D', '#EAB839', '#6ED0E0', '#EF843C', '#E24D42', '#1F78C1', '#BA43A9', '#705DA0', '#508642', '#CCA300', '#447EBC', '#C15C17'],
    // color: ['#61a0a8', '#2f4554', '#c23531', '#d48265', '#91c7ae', '#749f83', '#ca8622', '#bda29a', '#6e7074', '#546570', '#c4ccd3'],
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
    ],
    series: config.series
  }

  if (finalConfig.title) {
    option.title.text = config.title
  }

  if (finalConfig.eye) {
    option.toolbox.feature.myTool = {
      show:true,
      title: that.$t('button.chart.dataView'),
      icon: 'path://M432.45,595.444c0,2.177-4.661,6.82-11.305,6.82c-6.475,0-11.306-4.567-11.306-6.82s4.852-6.812,11.306-6.812C427.841,588.632,432.452,593.191,432.45,595.444L432.45,595.444z M421.155,589.876c-3.009,0-5.448,2.495-5.448,5.572s2.439,5.572,5.448,5.572c3.01,0,5.449-2.495,5.449-5.572C426.604,592.371,424.165,589.876,421.155,589.876L421.155,589.876z M421.146,591.891c-1.916,0-3.47,1.589-3.47,3.549c0,1.959,1.554,3.548,3.47,3.548s3.469-1.589,3.469-3.548C424.614,593.479,423.062,591.891,421.146,591.891L421.146,591.891zM421.146,591.891',   
      onclick: () => {
        that.$emit('sendConfig', that.chartItemx)
      }
    }
  }
  if (finalConfig.dataZoom) {
    option.toolbox.feature.dataZoom = {
      title: {
        zoom: that.$t('button.chart.zoom'),
        back: that.$t('button.chart.back'),
      },
      yAxisIndex: 'none'
    }
  }
  
  // 绘制图表
  myChart.setOption(option) 
  }