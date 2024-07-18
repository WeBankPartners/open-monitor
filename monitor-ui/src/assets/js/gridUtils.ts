const echarts = require('echarts/lib/echarts')
export const resizeEvent = function (that: any, i: string, newH: number, newW: number, newHPx: number, newWPx: number){
  that.layoutData.forEach((item: any,index: number) => {
    if (item.i === i) {
      that.layoutData[index].h = newH
      that.layoutData[index].w = newW
      that.layoutData[index]._activeCharts[0].style = ''
      const myChart = echarts.init(document.getElementById(item.id))
      myChart.resize({
        height: newHPx - 50+'px',
        width: newWPx - 6 +'px'
      })
      return
    }
  })
}
