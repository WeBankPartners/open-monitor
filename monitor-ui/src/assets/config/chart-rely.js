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
  if (responseData.series.length === 0) {
    that.chartTitle = responseData.title
    that.noDataTip = true
    return
  }
  let metricToColor = []
  let lineType = 1
  let metricEndpointColorInChartConfig = {}
  if (chartConfig.params) {
    lineType = chartConfig.params.lineType
    chartConfig.params.data.forEach(item => {
      metricEndpointColorInChartConfig[`${item.metric}:${item.endpoint}`] = item.defaultColor
      let nullColorIndex = []
      item.metricToColor.forEach((m, mIndex) => {
        if (m.color === '') {
          nullColorIndex.push(mIndex)
        }
      })

      if (nullColorIndex.length > 0 && item.defaultColor && item.defaultColor!== '') {
        let colors = generateAdjacentColors(item.defaultColor, nullColorIndex.length, 20)
        nullColorIndex.forEach((n, nIndex) => {
          item.metricToColor[n].color = colors[nIndex]
        })
      }
      metricToColor = metricToColor.concat(item.metricToColor)
    })

    // 处理在最初没数据，后面来数据 metricToColor 为空时的指标颜色处理
    responseData.series.forEach((item, itemIndex) => {
      const findIndex = metricToColor.findIndex(m => m.metric === item.name)
      if (findIndex === -1) {
        const keys = Object.keys(metricEndpointColorInChartConfig)
        keys.forEach(key => {
          if (item.name.startsWith(key)) {
            let color = generateAdjacentColors(metricEndpointColorInChartConfig[key], 1, 20 * (itemIndex - 0.3) )
            metricToColor.push({
              metric: item.name,
              color: color[0]
            })
          }
        })
      }
    })
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
    item.smooth = false
    item.lineStyle = {
      width: 1
    }
    let color = ''
    const find = metricToColor.find(m => m.metric === item.name)
    if (find && find.color !== '') {
      color = find.color
    } else {
      color = colorSet[index] ? colorSet[index] : '#666699'
    }
    // 堆叠区域图开启
    // item.areaStyle = {}

    // 渐变图像开启
    item.itemStyle = {
      normal:{
        color: color
      }
    }
    item.areaStyle = null
    if (lineType === 0) {
      item.areaStyle = {
        normal: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [{
              offset: 0,
              color: color
          }, {
              offset: 1,
              color: 'white'
          }])
        }
      }
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
  let originConfig = {
    title: true,
    eye: true,
    dataZoom: true,
    clear: false,
    editTitle: false,
    lineBarSwitch: false,
    chartType: 'line',
    zoomCallback: false // 选择区域后是否需要重新请求数据
  }
  let finalConfig = Object.assign(originConfig, userConfig)
  // 基于准备好的dom，初始化echarts实例
  var myChart = echarts.init(document.getElementById(elId || that.elId))
  myChart.resize()
  if (finalConfig.clear) {
    myChart.clear()
  }
  let option = {
    backgroundColor: '#f5f7f9',
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
          res = res+`<div><div style=' display: inline-block;width: 10px; 
          height: 10px;border: 1px solid transparent;border-radius:50%;
          background-color:${item.color};'  ></div>${Math.floor(item.data[1] * 1000) / 1000} ${seriesName}
          </div>`
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
  if (finalConfig.chartType !== config.series[0]) {
    config.series.forEach(se => {
      se.type = finalConfig.chartType
    })
  }
  if (finalConfig.title) {
    option.title.text = config.title
  }
  
  //切换为折线图，切换为柱状图
  if (finalConfig.lineBarSwitch) {
    option.toolbox.feature.magicType = {
      type: ['line', 'bar']
    }
  }
  if (finalConfig.eye) {
    option.toolbox.feature.myTool = {
      show:true,
      title: that.$t('button.chart.dataView'),
      icon: 'path://M432.45,595.444c0,2.177-4.661,6.82-11.305,6.82c-6.475,0-11.306-4.567-11.306-6.82s4.852-6.812,11.306-6.812C427.841,588.632,432.452,593.191,432.45,595.444L432.45,595.444z M421.155,589.876c-3.009,0-5.448,2.495-5.448,5.572s2.439,5.572,5.448,5.572c3.01,0,5.449-2.495,5.449-5.572C426.604,592.371,424.165,589.876,421.155,589.876L421.155,589.876z M421.146,591.891c-1.916,0-3.47,1.589-3.47,3.549c0,1.959,1.554,3.548,3.47,3.548s3.469-1.589,3.469-3.548C424.614,593.479,423.062,591.891,421.146,591.891L421.146,591.891zM421.146,591.891',   
      onclick: () => {
        that.$emit('sendConfig', that.chartInfo)
      }
    }
  }
  if (finalConfig.editTitle) {
    option.toolbox.feature.myEditTitle = {
      show:true,
      title: that.$t('button.chart.editTitle'),
      icon: 'path://M302.026 783.023q-0.761 0-2.282 0.761 88.25 0 352.999 0 6.847 0 11.412 4.565 4.565 4.565 4.565 10.651 0 6.847-4.565 11.412-4.565 4.565-11.412 4.565-156.72 0-470.919 0 0 0-0.761 0 0 0-0.761 0-1.522 0-2.282-0.761 0 0-0.761 0 0 0-0.761 0 0 0 0 0 0 0-0.761-0.761-0.761 0-1.522-0.761-0.761 0-1.522-0.761 0 0-0.761-0.761-0.761-0.761-2.282-2.282-0.761-1.522-1.522-2.282-0.761-0.761-0.761-1.522 0 0 0-0.761 0-0.761-0.761-0.761 0-0.761 0-1.522 0-0.761 0-2.282 0 0 0 0 0 0 0-0.761 0 0 0-0.761 0 0 0-1.522 0 0 0-0.761 0 0 0-0.761 0 0 0.761 0 0 0 0-0.761 7.608-28.909 30.431-115.638 1.522-3.804 4.565-6.847 160.523-159.002 481.57-477.006 4.565-4.565 10.651-4.565 6.847 0 11.412 4.565 4.565 4.565 4.565 10.651 0 6.847-4.565 11.412-159.763 158.241-478.527 473.962-6.086 21.302-23.584 85.968 22.062-5.325 86.728-22.823 119.442-118.681 478.527-473.962 4.565-4.565 10.651-4.565 6.847 0 11.412 4.565 4.565 4.565 4.565 10.651 0 6.847-4.565 11.412-160.523 159.002-482.331 477.006-2.282 3.043-6.847 3.804zM823.918 269.5q0.761-0.761 3.043-3.043 3.043-3.043 6.086-6.847 0.761-0.761 3.043-3.043 22.823-22.062 22.823-53.254 0-31.192-22.823-53.254-22.062-22.062-53.254-22.062-31.192 0-53.254 22.062-4.565 4.565-12.933 12.933 26.627 26.627 107.269 106.508zM166.608 798.999q0 0 0 0.761 0 0 0-0.761zM166.608 796.717q0 0 0 1.522 0 0 0-1.522zM166.608 801.281q0 0 0-1.522 0 0 0 1.522zM167.369 795.195q-0.761 0-0.761 0.761 0 0 0-0.761 0 0 0.761 0z',   
      iconStyle: {
        font: '22px',
        opacity: 0.7
      },
      onclick: () => {
        that.$emit('editTitle', config)
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
  // 清空所有事件重新绑定
  myChart.off()
  if (finalConfig.zoomCallback) {
    myChart.on('datazoom', function (params) {
      let startValue = null
      let endValue = null
      // TODO 多次放大后缩小无法恢复到最初状态，单次放大可以
      // 尚不知如何判断为放大还是缩小
      if (params.batch[0].endValue > 110) {
        startValue = parseInt(myChart.getModel().option.dataZoom[0].startValue/1000)
        endValue = parseInt(myChart.getModel().option.dataZoom[0].endValue/1000)
      }
      that.getChartData(null,startValue, endValue)
    })
  }
}

export const drawPieChart = function(that, responseData) {
  let option = {
    title: {
        // text: panalUnit,
        left: 'center'
    },
    tooltip: {
        confine: true, // tip控制在图像区内
        trigger: 'item',
        formatter: '{b} : {c} ({d}%)'
    },
    legend: {
        // orient: 'vertical',
        // top: 'middle',
        type: 'scroll',
        bottom: 5,
        left: 'center',
        data: responseData.legend
    },
    series: [
        {
            type: 'pie',
            radius: '65%',
            center: ['50%', '50%'],
            selectedMode: 'single',
            data: responseData.data,
            emphasis: {
                itemStyle: {
                    shadowBlur: 10,
                    shadowOffsetX: 0,
                    shadowColor: 'rgba(0, 0, 0, 0.5)'
                }
            }
        }
    ]
};
  var myChart = echarts.init(document.getElementById(that.elId))
  myChart.resize()
  myChart.setOption(option)
}