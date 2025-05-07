// 引入柱状图
require('echarts/lib/chart/line')
require('echarts/lib/chart/pie')
require('echarts/lib/chart/bar')
// 引入提示框和标题组件
require('echarts/lib/component/tooltip')
require('echarts/lib/component/title')
require('echarts/lib/component/legend')
require('echarts/lib/component/toolbox')
require('echarts/lib/component/legendScroll')
const {isEmpty} = require('lodash')

import { generateAdjacentColors, stringToNumber } from './random-color'
const echarts = require('echarts/lib/echarts')

export const readyToDraw = function (that, responseData, viewIndex, chartConfig, elId) {
  const legend = []
  let myChart = null
  if (responseData.series.length === 0) {
    that.chartTitle = responseData.title
    that.noDataTip = true
    myChart = elId && echarts.init(document.getElementById(elId))
    myChart && myChart.clear()
    return
  }
  let metricToColor = []
  let lineType = 1
  let isHostOrSys = false
  const metricEndpointColorInChartConfig = {}
  const metricSysColorInChartConfig = {}
  if (chartConfig.params) {
    lineType = chartConfig.params.lineType
    chartConfig.params.data&&chartConfig.params.data.forEach(item => {
      // 通过endpoint中‘.’的个数，判定是主机还是层级对象
      if (item.endpoint.split('.').length >= 3) {
        isHostOrSys = true
      }
      metricEndpointColorInChartConfig[`${item.metric}:${item.endpoint}`] = item.defaultColor || ''
      metricSysColorInChartConfig[`${item.metric}`] = item.defaultColor || ''
      const nullColorIndex = []
      item.metricToColor = item.metricToColor || []
      item.metricToColor.forEach((m, mIndex) => {
        if (m.color === '') {
          nullColorIndex.push(mIndex)
        }
      })

      if (nullColorIndex.length > 0 && item.defaultColor && item.defaultColor!== '') {
        const colors = generateAdjacentColors(item.defaultColor, nullColorIndex.length, 20)
        nullColorIndex.forEach((n, nIndex) => {
          item.metricToColor[n].color = colors[nIndex]
        })
      }
      metricToColor = metricToColor.concat(item.metricToColor)
    })
    // 处理在最初没数据，后面来数据 metricToColor 为空时的指标颜色处理
    if (isHostOrSys) {
      responseData.series.forEach(item => {
        const findIndex = metricToColor.findIndex(m => m.metric === item.name)
        if (findIndex === -1) {
          const keys = Object.keys(metricEndpointColorInChartConfig)
          keys.forEach(key => {
            if (item.name.startsWith(key)) {
              const color = generateAdjacentColors(metricEndpointColorInChartConfig[key], 1, stringToNumber(item.name))
              metricToColor.push({
                metric: item.name,
                color: color[0]
              })
            }
          })
        }
      })
    } else {
      // window.metricToColor = metricToColor
      responseData.series.forEach(item => {
        const findIndex = metricToColor.findIndex(m => m.metric === item.name)
        if (findIndex === -1) {
          const keys = Object.keys(metricSysColorInChartConfig)
          keys.forEach(key => {
            if (item.name.includes(key)) {
              const color = generateAdjacentColors(metricSysColorInChartConfig[key], 1, stringToNumber(item.name))
              metricToColor.push({
                metric: item.name,
                color: color[0]
              })
            }
          })
        }
      })
    }
  }
  const colorX = ['#33CCCC','#666699','#66CC66','#996633','#9999CC','#339933','#339966','#663333','#6666CC','#336699','#3399CC','#33CC66','#CC3333','#CC6666','#996699','#CC9933']
  const colorSet = []
  for (let i=0; i < colorX.length; i++) {
    let tmpIndex = viewIndex*3 + i
    tmpIndex = tmpIndex%colorX.length
    colorSet.push(colorX[tmpIndex])
  }
  responseData.series.forEach((item, index) => {
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
      normal: {
        color
      }
    }
    item.areaStyle = null
    if (lineType === 0) {
      item.areaStyle = {
        normal: {
          color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [{
            offset: 0,
            color
          }, {
            offset: 1,
            color: 'white'
          }])
        }
      }
    }
  })
  const config = {
    ...responseData,
    legend
  }
  myChart = drawChart(that, config, chartConfig, elId)
  return myChart
}

export const drawChart = function (that,config,userConfig, elId) {
  const chartTextColor = '#a1a1a2'
  const originConfig = {
    title: true,
    eye: true,
    dataZoom: true,
    clear: false,
    editTitle: false,
    lineBarSwitch: false,
    chartType: 'line',
    zoomCallback: false // 选择区域后是否需要重新请求数据
  }
  const finalConfig = Object.assign(originConfig, userConfig)
  // 基于准备好的dom，初始化echarts实例
  const myChart = echarts.init(document.getElementById(elId || that.elId))
  myChart.resize()
  // 最新的参数，假如isNeedClear为true, 则需要重新刷新
  if (finalConfig.isNeedClear) {
    myChart.clear()
  }
  // if (finalConfig.clear) {
  //   myChart.clear()
  // }
  let isTwoYaxes = false
  if (finalConfig.params && finalConfig.params.lineType === 2) {
    isTwoYaxes = true
  }
  const minMax = mgmtYAxesMinMax(config.series)
  const option = {
    backgroundColor: '#f5f7f9',
    title: {
      textStyle: {
        fontSize: 16,
        fontWeight: 'bolder',
        color: chartTextColor // 主标题文字颜色
      },
      // text: config.title,
      left: '10%',
      top: '10px'
    },
    tooltip: {
      enterable: false,
      appendToBody: true,
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
      formatter: params => {
        let str =''
        const date = new Date(params[0].data[0])
        // let year =  date.getFullYear()
        // let month = (date.getMonth() + 1)>=10?(date.getMonth() + 1):'0'+(date.getMonth() + 1)
        // let day = date.getDate()>=10?date.getDate():'0'+date.getDate()
        const hours = date.getHours()>=10?date.getHours():'0'+date.getHours()
        const minutes = date.getMinutes()>=10?date.getMinutes():'0'+date.getMinutes()
        const seconds = date.getSeconds()>=10?date.getSeconds():'0'+date.getSeconds()
        str=hours+':'+minutes+':'+seconds
        let res = `<div>${str}</div>`
        const regex = /{.*}/
        const isAllSeriesNameContainBrackets = params.every(item => regex.test(item.seriesName))
        // 所有指标中均包含大括号启用分组、否则不原样显示
        if (isAllSeriesNameContainBrackets) {
          const titleSet = {}
          params.forEach(item => {
            const metricSplit = item.seriesName.split('{')
            if (Object.keys(titleSet).includes(metricSplit[0])) {
              titleSet[metricSplit[0]].push({
                color: item.color,
                data: item.data,
                metric: `{${metricSplit[1]}`
              })
            } else {
              titleSet[metricSplit[0]] = [{
                color: item.color,
                data: item.data,
                metric: `{${metricSplit[1]}`
              }]
            }
          })
          const keys = Object.keys(titleSet)
          keys.forEach(key => {
            res = res + `<div style="color:#5384FF">${key}</div>`
            titleSet[key].forEach(item => {
              res = res+`<div><div style=' display: inline-block;width: 10px; 
                height: 10px;border: 1px solid transparent;border-radius:50%;
                background-color:${item.color};'  ></div>${Math.floor(item.data[1] * 1000) / 1000} ${item.metric}
                </div>`
            })
          })
        } else {
          params.forEach(item => {
            const str = item.seriesName
            const step = 100
            const strLen = str.length
            const arr = []
            for (let i=0; i<strLen; i=i+step){
              arr.push(str.substr(i, step))
            }
            arr.join(' ')
            const seriesName = arr.join('<br>')
            res = res+`<div><div style=' display: inline-block;width: 10px;
            height: 10px;border: 1px solid transparent;border-radius:50%;
            background-color:${item.color};'  ></div>${Math.floor(item.data[1] * 1000) / 1000} ${seriesName}
            </div>`
          })
        }
        res = `<div class="echarts-custom-tooltip-${finalConfig.chartId || ''}">` + res + '</div>'
        return res
      },
    },
    toolbox: {
      right: isTwoYaxes ? '170px' : '4%',
      top: '4px',
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
        color: chartTextColor // 图例文字颜色
      },
      type: 'scroll',
      y: 'bottom',
      padding: 25,
      orient: 'horizontal',
      data: config.legend
    },
    calculable: false,
    grid: {
      top: '40',
      left: '3%',
      right: '5%',
      bottom: '40',
      containLabel: true
    },
    xAxis: {
      type: 'time',
      axisLabel: {
        textStyle: {
          color: chartTextColor
        },
        formatter(value) {
          window.maxFormatTimeStamp = window.maxFormatTimeStamp || value
          if (value > window.maxFormatTimeStamp) {
            window.maxFormatTimeStamp = value
            window.maxFormatTimeFormat = echarts.format.formatTime('MM-dd\nhh:mm:ss', value)
          }
          return echarts.format.formatTime('MM-dd\nhh:mm:ss', value)
        }
      },
      boundaryGap: false,
      max(value) {
        window.xAxisMaxValue = new Date(value.max) + '【' + value.max + '】'
        return value.max
      },
      min(value) {
        window.xAxisMinValue = new Date(value.min)
        return value.min
      },
      axisLine: {
        lineStyle: {
          color: '#a1a1a2'
        }
      },
      splitLine: {
        show: true,
        lineStyle: {
          color: ['#a1a1a2'],
          type: 'dotted',
          width: 0.3
        }
      }
    },
    yAxis: [
      {
        type: 'value',
        name: isTwoYaxes ? that.$t('m_difference') : '',
        alignTicks: true,
        max: minMax.y1Max,
        min: minMax.y1Min,
        axisLabel: {
          textStyle: {
            color: chartTextColor
          },
          show: true,
          interval: 'auto',
          formatter: value => {
            let val = value
            let unit = ''
            if (val > 1024*1024*1024*1024) {
              val = val / (1024*1024*1024*1024)
              unit = 'T'
            } else if (val > 1024*1024*1024) {
              val = val / (1024*1024*1024)
              unit = 'G'
            } else if (val > 1024*1024) {
              val = val / (1024*1024)
              unit = 'M'
            } else if (val > 1024) {
              val = val / 1024
              unit = 'K'
            } else {
              return val + ' ' + config.yaxis.unit
            }
            const newValue = Number.isInteger(val) ? val : val.toFixed(3)
            return newValue + ' ' + unit + config.yaxis.unit
          }
        },
        show: true,
        axisLine: {
          lineStyle: {
            color: '#a1a1a2'
          }
        },
        splitLine: {
          show: true,
          lineStyle: {
            color: ['#a1a1a2'],
            type: 'dotted',
            width: 0.3
          }
        }
      },
      {
        type: 'value',
        name: isTwoYaxes ? that.$t('m_percentage_difference') : '',
        max: minMax.y2Max,
        min: minMax.y2Min,
        axisLabel: {
          textStyle: {
            color: chartTextColor
          },
          show: true,
          interval: 'auto',
          formatter: value => value + '%'
        },
        show: true,
        axisLine: {
          lineStyle: {
            color: '#a1a1a2'
          }
        },
        splitLine: {
          show: true,
          lineStyle: {
            color: ['#a1a1a2'],
            type: 'dotted',
            width: 0.3
          }
        }
      }
    ],
    series: config.series
  }
  if (!isTwoYaxes && finalConfig.chartType !== config.series[0]) {
    config.series.forEach(se => {
      se.type = finalConfig.chartType
    })
  }
  if (finalConfig.title) {
    option.title.text = config.title
  }

  // 切换为折线图，切换为柱状图
  if (finalConfig.lineBarSwitch) {
    option.toolbox.feature.magicType = {
      type: ['line', 'bar']
    }
  }
  if (finalConfig.eye) {
    option.toolbox.feature.myTool = {
      show: true,
      title: that.$t('m_button_chart_dataView'),
      icon: 'path://M432.45,595.444c0,2.177-4.661,6.82-11.305,6.82c-6.475,0-11.306-4.567-11.306-6.82s4.852-6.812,11.306-6.812C427.841,588.632,432.452,593.191,432.45,595.444L432.45,595.444z M421.155,589.876c-3.009,0-5.448,2.495-5.448,5.572s2.439,5.572,5.448,5.572c3.01,0,5.449-2.495,5.449-5.572C426.604,592.371,424.165,589.876,421.155,589.876L421.155,589.876z M421.146,591.891c-1.916,0-3.47,1.589-3.47,3.549c0,1.959,1.554,3.548,3.47,3.548s3.469-1.589,3.469-3.548C424.614,593.479,423.062,591.891,421.146,591.891L421.146,591.891zM421.146,591.891',
      onclick: () => {
        that.$emit('sendConfig', that.chartInfo)
      }
    }
  }
  if (finalConfig.editTitle) {
    option.toolbox.feature.myEditTitle = {
      show: true,
      title: that.$t('m_button_chart_editTitle'),
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
        zoom: that.$t('m_button_chart_zoom'),
        back: that.$t('m_button_chart_back'),
      },
      yAxisIndex: 'none'
    }
  }
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
  if (finalConfig.canEditShowLines) {
    option.toolbox.feature.myEditLines = {
      show: true,
      title: that.$t('m_line_display_modification'),
      icon: 'image://data:image/png;base64,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',
      iconStyle: {
        font: '22px',
        opacity: 0.7
      },
      onclick: () => {
        that.$emit('editShowLines', config)
      }
    }
  }
  if (!isEmpty(window['view-config-selected-line-data']) && finalConfig.chartId && !isEmpty(window['view-config-selected-line-data'][finalConfig.chartId])) {
    option.legend.selected = window['view-config-selected-line-data'][finalConfig.chartId]
  }
  // 绘制图表
  // myChart.clear()
  myChart.setOption(option)
  // 清空所有事件重新绑定
  myChart.off()
  setTimeout(() => {
    myChart.resize()
  }, 200)

  return myChart
}

export const drawPieChart = function (that, responseData) {
  const option = {
    title: {
      // text: panalUnit,
      left: 'center'
    },
    tooltip: {
      appendToBody: true,
      confine: true, // tip控制在图像区内
      trigger: 'item',
      formatter: '{b} : {c} ({d}%)'
    },
    legend: {
      // orient: 'vertical',
      // top: 'middle',
      type: 'scroll',
      bottom: 0,
      left: 'center',
      data: responseData.legend
    },
    series: [
      {
        type: 'pie',
        radius: '55%',
        center: ['50%', '45%'],
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
  }
  if (!isEmpty(window['view-config-selected-line-data']) && responseData.chartId && !isEmpty(window['view-config-selected-line-data'][responseData.chartId])) {
    option.legend.selected = window['view-config-selected-line-data'][responseData.chartId]
  }
  const myChart = echarts.init(document.getElementById(that.elId))
  myChart.resize()
  myChart.setOption(option)
  return myChart
}

const mgmtYAxesMinMax = function (series) {
  const { maxValue: max1 = 1, minValue: min1 } = findMinMaxForYAxisIndexOne(series, 0)
  const { maxValue: max2, minValue: min2 } = findMinMaxForYAxisIndexOne(series, 1)
  if (min1 >= 0 || min2 >= 0) {
    return false
  }
  const ratio = (max1 - min1) / (max2 - min2)
  const minMax = {}
  if (max1 < max2 * ratio) {
    minMax.y1Max = max2 * ratio
    minMax.y2Max = max2
  } else {
    minMax.y1Max = max1
    minMax.y2Max = max1 / ratio
  }
  if (min1 < min2 * ratio) {
    minMax.y1Min = min1
    minMax.y2Min = min1 / ratio
  } else {
    minMax.y1Min = min2 * ratio
    minMax.y2Min = min2
  }
  minMax.y1Min = (minMax.y1Min * 1.5).toFixed(2)
  minMax.y2Min = (minMax.y2Min * 1.5).toFixed(2)
  minMax.y1Max = (minMax.y1Max * 1.5).toFixed(2)
  minMax.y2Max = (minMax.y2Max * 1.5).toFixed(2)
  return minMax
}

const findMinMaxForYAxisIndexOne = function (data, yAxisIndex) {
  // 过滤出 yAxisIndex 为 1 的数据系列
  const yAxisIndexOneSeries = data.filter(series => series.yAxisIndex === yAxisIndex)

  // 初始化最大值和最小值变量
  let maxValue = 1
  let minValue = 0

  // 遍历所有符合条件的数据系列
  yAxisIndexOneSeries.forEach(series => {
    // 获取该系列的数据点
    const values = series.data

    // 遍历数据点，更新最大值和最小值
    values.forEach(value => {
      if (value[1] > maxValue) {
        maxValue = value[1]
      }
      if (value[1] < minValue) {
        minValue = value[1]
      }
    })
  })

  return {
    maxValue,
    minValue
  }
}
