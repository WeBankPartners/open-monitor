
const echarts = require('echarts/lib/echarts');
export const resizeEvent = function(that, i, newH, newW, newHPx, newWPx){
  that.layoutData.forEach((item,index) => {
    if (item.i === i) {
      that.layoutData[index].h = newH
      that.layoutData[index].w = newW
      var myChart = echarts.init(document.getElementById(item.id))
      myChart.resize({height:newHPx - 34+'px',width:newWPx - 6 +'px'})
      return
    }
  })
}