<template>
  <div class=" ">
      <header>
        <div style="display:flex;justify-content:space-between; font-size:16px;padding:8px 16px">
            <div class="header-name">
                <span>{{$t('tableKey.systemName')}}:</span>
                <span> {{sysConfig.systemName}}</span>
            </div>
            <!-- <div class="header-tools"> 
              <Select v-model="metricMulti" multiple style="width:200px">
                <Option v-for="item in metricLabelList" :value="item.value" :key="item.value">{{ item.label }}</Option>
              </Select>
            </div> -->
        </div>
      </header>
      <grid-layout
        :layout.sync="layoutData"
        :col-num="12"
        :row-height="30"
        :is-draggable="true"
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
        :key="index"
        @resize="resizeEvent"
        @resized="resizedEvent">
                 
        <div style="display:flex;justify-content:flex-end;padding:0 32px;">
          <div class="header-grid header-grid-name">
            <span>{{item.i}}</span>
          </div>
          <div class="header-grid header-grid-tools"> 
            <Tooltip :content="$t('placeholder.viewChart')" placement="top">
              <i class="fa fa-eye" aria-hidden="true" @click="gridPlus(item)"></i>
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
        systemName: 'test',
        ips: ['192.168.0.16','192.168.0.5'],
        endpointList: [],
      },
      metricMulti:['cpu.used.percent','mem.used.percent','load.1min'],
      metricLabelList: [
        {
          value: 'cpu.used.percent',
          label: 'cpu.used.percent'
        },
        {
          value: 'mem.used.percent',
          label: 'mem.used.percent'
        },
        {
          value: 'load.1min',
          label: 'load.1min'
        },
        {
          value: 'disk.read.bytes',
          label: 'disk.read.bytes'
        },
        {
          value: 'disk.write.bytes',
          label: 'disk.write.bytes'
        },
        {
          value: 'disk.iops',
          label: 'disk.iops'
        },
        {
          value: 'net.if.out.bytes',
          label: 'net.if.out.bytes'
        }
      ],
      array1: []
    }
  },
  mounted() {
    // systemMonitoring?systemName=test&ips=192.168.0.16,192.168.0.5
    // this.sysConfig.systemName = this.$route.query.systemName
    // this.sysConfig.ips = this.$route.query.ips.split(',')
    this.getMetric()
  },
  methods: {
    getMetric () {
      let url = '/dashboard/custom/endpoint/get?'
      let xx = this.sysConfig.ips.map((item) => {
        return 'ip=' + item
      })
      url += xx.join('&')
      this.$httpRequestEntrance.httpRequestEntrance('GET',url, '', responseData => {
        responseData.forEach((i)=>{
          this.sysConfig.endpointList.push(i.guid)
        })
        this.initData()
      })
    },
    initData () {
      this.viewData = []
      let res = []
      const num = this.metricMulti.length
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
      this.metricMulti.forEach((metric, index)=> {
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
        let config = {}
        config = {
          title: responseData.title,
          legend: legend,
          series: responseData.series,
          yaxis: responseData.yaxis
        }
        this.elId = id
        drawChart(this, config, {eye: false,dataZoom:false})
      })
    },
    resizeEvent: function(i, newH, newW, newHPx, newWPx){
      this.layoutData.forEach((item,index) => {
        if (item.i === i) {
          this.layoutData[index].h = newH
          this.layoutData[index].w = newW
          var myChart = echarts.init(document.getElementById(item.id))
          myChart.resize({height:newHPx-64+'px',width:newWPx+'px'})
          return
        }
      })
    },
    resizedEvent: function(i, newH, newW, newHPx, newWPx){
      this.resizeEvent(i, newH, newW, newHPx, newWPx)
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
