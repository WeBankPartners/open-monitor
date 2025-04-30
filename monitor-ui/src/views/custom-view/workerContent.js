export const initPanalsWebWorker = dataObj => {
  const viewDataArr = dataObj.viewDataArr
  const viewCondition = dataObj.viewCondition
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
    parsedDisplayConfig = JSON.parse(item.displayConfig)

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
  return tmpArr
}
