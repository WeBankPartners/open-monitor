<template>
  <div class=" ">
      <header>
        <div style="display:flex;justify-content:space-between; font-size:16px;padding:8px 16px">
            <div class="header-name">
                <span>{{$t('tableKey.systemName')}}:</span>
                <span> Wecube-monitor</span>
            </div>
            <!-- <div class="header-tools"> 
              <button class="btn btn-sm btn-cancle-f" @click="goBack()">{{$t('button.back')}}</button>
            </div> -->
        </div>
      </header>
      <grid-layout
        :layout.sync="layoutData"
        :col-num="12"
        :row-height="30"
        :is-draggable="false"
        :is-resizable="false"
        :is-mirrored="false"
        :vertical-compact="true"
        :use-css-transforms="true"
        >
      <grid-item v-for="(item,index) in layoutData"
        :x="item.x"
        :y="item.y"
        :w="item.w"
        :h="item.h"
        :i="item.i"
        :key="index">
                 
        <div style="display:flex;justify-content:flex-end;padding:0 32px;">
          <div class="header-grid header-grid-name">
            <span>{{item.i}}</span>
          </div>
          <div class="header-grid header-grid-tools"> 
            <Tooltip :content="$t('placeholder.viewChart')" placement="top">
              <!-- <i class="fa fa-eye" aria-hidden="true" @click="gridPlus(item)"></i> -->
              <!-- <i class="fa fa-eye" aria-hidden="true"></i> -->
            </Tooltip>
          </div>
        </div>
        <div class="">
          <section class="metric-section">
            <div v-if="!noDataTip">
              <div :id="item.id" class="echart" style="height:230px;width:560px"></div>
            </div>
             <div v-else class="echart echart-no-data-tip">
              <span>~~~No Data!~~~</span>
            </div>
          </section>
        </div>
        </grid-item>
      </grid-layout>
  </div>
</template>

<script>
import VueGridLayout from 'vue-grid-layout'
import {drawChart} from  '@/assets/config/chart-rely'
const echarts = require('echarts/lib/echarts');
export default {
  name: '',
  data() {
    return {
      viewData: [],
      layoutData: [
      ],
      noDataTip: false,
      sysConfig: {
        systemName: '双子星系统',
        ip: ['192.168.0.16','192.168.0.5'],
        endpointList: ['VM_0_16_centos_192.168.0.16_host','VM_0_5_centos_192.168.0.5_host'],
      },
      metricLabelList: ['mem.used.percent','load.1min','cpu.used.percent','disk.write.bytes'],
      array1: []
    }
  },
  mounted() {
    let res = []
    const num = this.metricLabelList.length
    for (let i=0;i<num; i++) {
      const key = ((new Date()).valueOf()).toString().substring(10)
      let xx = {
        x: i%2*6,
        y:Math.ceil((i/2+0.000001)-1)*7,
        w:6,
        h:7,
        i: '',
        id: `id_${key}${i}`,
        "moved": false
      }
      this.array1.push(xx)
    }
    this.metricLabelList.forEach((metric, index)=> {
      let singleChart = {}
      singleChart.panalTitle = metric
      singleChart.query = []
      for (let endpoint of this.sysConfig.endpointList) {
        let condition = {}
        condition.endpoint = endpoint
        condition.metricLabel = metric
        singleChart.query.push(condition)
      }
      this.array1[index].i = metric
      singleChart.viewConfig = this.array1[index]
      res.push(singleChart)
    })
    this.viewData = res
    this.initPanals()
  },
  methods: {
    initPanals () {
      this.viewData.forEach((item,viewIndex) => {
        this.layoutData.push(item.viewConfig)
        this.requestChart(item.viewConfig.id, item.query,viewIndex)
      })
    },
    requestChart (id, query,viewIndex) {
      let params = []
      query.forEach((item) => {
        params.push(JSON.stringify({
          endpoint: item.endpoint,
          metric: item.metricLabel,
          prom_ql: item.metric,
          time: '-1800'
        })) 
      })
      this.$httpRequestEntrance.httpRequestEntrance('GET',this.apiCenter.metricConfigView.api, {config: `[${params.join(',')}]`}, responseData => {
        var legend = []
        if (responseData.series.length === 0) {
          this.noDataTip = true
          return
        }
        const colorX = ['#339933','#33CCCC','#666699','#66CC66','#996633','#9999CC','#339966','#663333','#6666CC','#336699','#3399CC','#33CC66','#CC3333','#CC6666','#996699','#CC9933']
        let colorSet = []
        for (let i=0;i<colorX.length;i++) {
          let tmpIndex = viewIndex*3 + i
          tmpIndex = tmpIndex%colorX.length
          colorSet.push(colorX[tmpIndex])
        }
        responseData.series.forEach((item,index)=>{
          legend.push(item.name)
          item.symbol = 'none'
          item.smooth = true
          item.lineStyle = {
            width: 1
          }
          item.itemStyle = {
            normal:{
              color: colorSet[index]
            }
          }
          item.areaStyle = {
            normal: {
              color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [{
                  offset: 0,
                  color: colorSet[index]
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
        this.elId = id
        drawChart(this, config, {eye: false,dataZoom:false})
      })
    }
  },
  components: {
    GridLayout: VueGridLayout.GridLayout,
    GridItem: VueGridLayout.GridItem,
  },
}
</script>

<style scoped lang="less">
  .header-grid {
    flex-grow: 1;
    text-align: end;
    line-height: 32px;
    i {
      margin: 0 4px;
      cursor: pointer;
    } 
  }
</style>
<style scoped lang="less">

.vue-grid-item:not(.vue-grid-placeholder) {
    background: @gray-f;
    border: 1px solid @gray-f;
}
.echart-no-data-tip {
  text-align: center;
}
</style>
