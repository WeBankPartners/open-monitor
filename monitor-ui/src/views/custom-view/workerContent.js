const orderBy = require('lodash/orderBy')
const find = require('lodash/find')
export const initPanalsWebWorker = dataObj => {
  const viewDataArr = dataObj.viewDataArr
  const viewCondition = dataObj.viewCondition
  const activeGroup = dataObj.activeGroup

  const actionType = dataObj.actionType
  let tempLayoutData = dataObj.layoutData
  const calculateLayout = (data = [], type='customize') => {
    if (data.length === 0 || type==='customize') {
      return data
    }
    data.forEach((item, index) => {
      item.h = 7
      if (type === 'two') {
        item.w = 6
        item.x = (index % 2) * 6
        item.y = Math.floor(index / 2) * 7
      } else if (type === 'three') {
        item.w = 4
        item.x = (index % 3) * 4
        item.y = Math.floor(index / 3) * 7
      } else if (type === 'four') {
        item.w = 3
        item.x = (index % 4) * 3
        item.y = Math.floor(index / 4) * 7
      } else if (type === 'five') {
        item.w = 2.4
        item.x = (index % 5) * 2.4
        item.y = Math.floor(index / 5) * 7
      } else if (type === 'six') {
        item.w = 2
        item.x = (index % 6) * 2
        item.y = Math.floor(index / 6) * 7
      } else if (type === 'seven') {
        item.w = 1.7
        item.x = (index % 7) * 1.7
        item.y = Math.floor(index / 7) * 7
      } else if (type === 'eight') {
        item.w = 1.5
        item.x = (index % 8) * 1.5
        item.y = Math.floor(index / 8) * 7
      }
    })
    return data
  }

  const confirmLayoutType = data => {
    if (!data || !data.length) {
      return 'customize'
    }
    let res = ''
    data.every((item, i) => (item.x === (i % 3) * 4) && item.y === Math.floor(i / 3) * 7 && item.h === 7 && item.w === 4) ? res = 'three'
      : data.every((item, i) => item.x === (i % 2) * 6 && item.y === Math.floor(i / 2) * 7 && item.h === 7 && item.w === 6) ? res = 'two'
        : data.every((item, i) => (item.x === (i % 4) * 3) && item.y === Math.floor(i / 4) * 7 && item.h === 7 && item.w === 3) ? res = 'four'
          : data.every((item, i) => (item.x === (i % 5) * 2.4) && item.y === Math.floor(i / 5) * 7 && item.h === 7 && item.w === 2.4) ? res = 'five'
            : data.every((item, i) => (item.x === (i % 6) * 2) && item.y === Math.floor(i / 6) * 7 && item.h === 7 && item.w === 2) ? res = 'six'
              : data.every((item, i) => (item.x === (i % 7) * 1.7) && item.y === Math.floor(i / 7) * 7 && item.h === 7 && item.w === 1.7) ? res = 'seven'
                : data.every((item, i) => (item.x === (i % 8) * 1.5) && item.y === Math.floor(i / 8) * 7 && item.h === 7 && item.w === 1.5) ? res = 'eight' : res = 'customize'
    return res
  }

  const sortLayoutData = data => {
    const sortedArr = orderBy(data, ['y', 'x'], ['asc', 'asc'])
    return sortedArr
  }

  const initLayoutTypeByWidth = data => {
    let chartLayoutType = ''
    if (!data || !data.length) {
      return 'customize'
    }
    data.every(item => item.h === 7 && item.w === 4) ? chartLayoutType = 'three'
      : data.every(item => item.h === 7 && item.w === 6) ? chartLayoutType = 'two'
        : data.every(item => item.h === 7 && item.w === 3) ? chartLayoutType = 'four'
          : data.every(item => item.h === 7 && item.w === 2.4) ? chartLayoutType = 'five'
            : data.every(item => item.h === 7 && item.w === 2) ? chartLayoutType = 'six'
              : data.every(item => item.h === 7 && item.w === 1.7) ? chartLayoutType = 'seven'
                : data.every(item => item.h === 7 && item.w === 1.5) ? chartLayoutType = 'eight' : 'customize'
    return chartLayoutType
  }

  const filterLayoutData = (layoutData, activeGroup, chartLayoutType = 'customize') => {
    if (activeGroup === 'ALL') {
      layoutData = calculateLayout(layoutData, chartLayoutType) // eslint-disable-line no-param-reassign
    } else {
      layoutData = layoutData.filter(d => d.group === activeGroup) // eslint-disable-line no-param-reassign
      if (layoutData && !layoutData.length) {
        return []
      }
      const layoutNeedReset = layoutData.some(item => item.partGroupDisplayConfig === '')
      // 假如其中有partGroupDisplayConfig为空，基于两列进行打平，假如没有为空，则基于partGroupDisplayConfig排列
      chartLayoutType = initLayoutTypeByWidth(layoutData) // eslint-disable-line no-param-reassign
      if (layoutNeedReset || ['two', 'three', 'four', 'five', 'six', 'seven', 'eight'].includes(chartLayoutType)) {
        sortLayoutData(layoutData)
        calculateLayout(layoutData, chartLayoutType)
      } else {
        layoutData.forEach(item => {
          Object.assign(item, item.partGroupDisplayConfig)
        })
        sortLayoutData(layoutData)
      }
    }
    chartLayoutType = confirmLayoutType(layoutData) // eslint-disable-line no-param-reassign
    return {
      layoutData,
      chartLayoutType
    }
  }

  const dateToTimestamp = date => {
    if (!date) {
      return 0
    }
    let timestamp = Date.parse(new Date(date))
    timestamp = timestamp / 1000
    return timestamp
  }
  const isValidJson = str => {
    try {
      JSON.parse(str)
      return true
    } catch (e) {
      return false
    }
  }
  const lineTypeOption = {
    twoYaxes: 2,
    line: 1,
    area: 0
  }
  const tmpArr = []

  for (let k=0; k<viewDataArr.length; k++) {
    const item = viewDataArr[k]
    // 先对groupDisplayConfig进行初始化，防止异常值
    if (!isValidJson(item.groupDisplayConfig) || Object.keys(JSON.parse(item.groupDisplayConfig)).length === 0) {
      item.groupDisplayConfig = ''
    }
    let parsedDisplayConfig = {}
    if (activeGroup === 'ALL') {
      parsedDisplayConfig = JSON.parse(item.displayConfig)
    } else {
      // 在各个分组中的情况
      if (isValidJson(item.groupDisplayConfig) && Object.keys(JSON.parse(item.groupDisplayConfig)) !== 0) {
        parsedDisplayConfig = JSON.parse(item.groupDisplayConfig)
      } else {
        item.groupDisplayConfig = ''
        parsedDisplayConfig = JSON.parse(item.displayConfig)
      }
    }

    const params = {
      aggregate: item.aggregate,
      agg_step: item.aggStep,
      lineType: lineTypeOption[item.lineType],
      time_second: viewCondition.timeTnterval,
      start: dateToTimestamp(viewCondition.dateRange[0]),
      end: dateToTimestamp(viewCondition.dateRange[1]),
      title: '',
      unit: '',
      data: []
    }
    for (let i=0; i<item.chartSeries.length; i++) {
      const single = item.chartSeries[i]
      single.defaultColor = single.colorGroup
      params.data.push(single)
    }
    const height = (parsedDisplayConfig.h + 1) * 30-8
    const _activeCharts = []
    _activeCharts.push({
      style: `height:${height}px;`,
      panalUnit: item.unit,
      elId: item.id,
      chartParams: params,
      chartType: item.chartType,
      aggregate: item.aggregate,
      agg_step: item.aggStep,
      lineType: lineTypeOption[item.lineType],
      time_second: viewCondition.timeTnterval,
      start: dateToTimestamp(viewCondition.dateRange[0]),
      end: dateToTimestamp(viewCondition.dateRange[1]),
      parsedDisplayConfig
    })
    const one = Object.assign({}, {
      _activeCharts,
      i: item.name,
      id: item.id,
      group: item.group,
      public: item.public,
      sourceDashboard: item.sourceDashboard,
      allGroupDisplayConfig: JSON.parse(item.displayConfig),
      partGroupDisplayConfig: item.groupDisplayConfig === '' ? '' : JSON.parse(item.groupDisplayConfig),
      logMetricGroup: item.logMetricGroup
    }, parsedDisplayConfig)
    tmpArr.push(one)
  }

  if (!tempLayoutData.length || actionType === 'init') {
    tempLayoutData = tmpArr
  } else {
    tmpArr.forEach(tmpItem => {
      const item = find(tempLayoutData, {
        id: tmpItem.id
      })
      if (item) {
        item._activeCharts = tmpItem._activeCharts
      }
    })
  }
  tempLayoutData = sortLayoutData(JSON.parse(JSON.stringify(tempLayoutData)))
  return filterLayoutData(tempLayoutData, activeGroup)
  // return tmpArr
}
