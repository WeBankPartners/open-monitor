<template>
  <div class=" ">
      <header>
        <div style="display:flex;justify-content:space-between; font-size:16px;padding:8px 16px">
            <div class="header-name">
                <span>系统名称：</span>
                <span> 双子星系统</span>
            </div>
            <div class="header-tools"> 
              <button class="btn btn-sm btn-cancle-f" @click="goBack()">{{$t('button.back')}}</button>
            </div>
        </div>
      </header>
      <grid-layout
        :layout.sync="layoutData"
        :col-num="12"
        :row-height="30"
        :is-draggable="true"
        :is-resizable="true"
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
              <!-- <i class="fa fa-eye" aria-hidden="true" @click="gridPlus(item)"></i> -->
              <i class="fa fa-eye" aria-hidden="true"></i>
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
import {generateUuid} from '@/assets/js/utils'
import VueGridLayout from 'vue-grid-layout'
import {drawChart} from  '@/assets/config/chart-rely'
const echarts = require('echarts/lib/echarts');
export default {
  name: '',
  data() {
    return {
      viewData: [],
      layoutData: [
        //   {'x':0,'y':0,'w':2,'h':2,'i':'0'},
        //   {'x':1,'y':1,'w':2,'h':2,'i':'1'},
      ],
      noDataTip: false,
      sysConfig: {
        systemName: '双子星系统',
        ip: ['192.168.0.16','192.168.0.5'],
        endpointList: ['VM_0_16_centos_192.168.0.16_host','VM_0_5_centos_192.168.0.5_host'],
      },
      metricLabelList: ['mem.used.percent','load.1min','cpu.used.percent'],
      array1: []
    }
  },
  mounted() {
    let res = []
    const num = this.metricLabelList.length
    for (let i=0;i<num; i++) {
      generateUuid().then((elId)=>{
        const key = ((new Date()).valueOf()).toString().substring(10)
        this.array1.push({
          x: i%2*6,
          y:Math.ceil((i/2+0.000001)-1)*7,
          w:6,
          h:7,
          i: `default${key}`,
          id: `id_${elId}`,
          "moved": false
        })
      })
    }
    console.log(this.array1)
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
      singleChart.viewConfig = this.array1[index]
      res.push(singleChart)
    })

    console.log(res)
    this.viewData = [{
        "panalTitle": "default824",
        "query": [{
            "endpoint": 'VM_0_16_centos_192.168.0.16_host',
            "metricLabel": "mem.used.percent",
        },
        {
            "endpoint": 'VM_0_5_centos_192.168.0.5_host',
            "metricLabel": "mem.used.percent",
        }],
        "viewConfig": {
            "x": 0,
            "y": -0,
            "w": 6,
            "h": 7,
            "i": "default824",
            "id": "id_9244fc70_79b8_4c95_876d_f1d27aec9283",
            "moved": false
        }
    },
    {
        "panalTitle": "default740",
        "query": [
          {
            "endpoint": 'VM_0_16_centos_192.168.0.16_host',
            "metricLabel": "load.1min",
        },
        {
            "endpoint": 'VM_0_5_centos_192.168.0.5_host',
            "metricLabel": "load.1min",
        }],
        "viewConfig": {
            "x": 6,
            "y": 0,
            "w": 6,
            "h": 7,
            "i": "default740",
            "id": "id_5bf38763_afdd_41d4_ae66_339304821870",
            "moved": false
        }
    },
    {
        "panalTitle": "default653",
        "query": [
          {
            "endpoint": 'VM_0_16_centos_192.168.0.16_host',
            "metricLabel": "cpu.used.percent",
        },
        {
            "endpoint": 'VM_0_5_centos_192.168.0.5_host',
            "metricLabel": "cpu.used.percent",
        }],
        "viewConfig": {
            "x": 0,
            "y": 7,
            "w": 6,
            "h": 7,
            "i": "default653",
            "id": "id_5ab4bdb7_aaf8_47da_8cbe_cd5e65e7aa34",
            "moved": false
        }
    },
    {
      "panalTitle": "default6531111",
      "query": [
        {
          "endpoint": 'VM_0_16_centos_192.168.0.16_host',
          "metricLabel": "cpu.used.percent",
      },
      {
          "endpoint": 'VM_0_5_centos_192.168.0.5_host',
          "metricLabel": "cpu.used.percent",
      }],
      "viewConfig": {
          "x": 6,
          "y": 7,
          "w": 6,
          "h": 7,
          "i": "default6531111",
          "id": "id_5ab4bdb7_aaf8_47da_8cbe_cd5e65e7aasfd34",
          "moved": false
      }
    }]
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
    },
    addItem() {
      generateUuid().then((elId)=>{
        const key = ((new Date()).valueOf()).toString().substring(10)
        let item = {
          x:0,
          y:0,
          w:6,
          h:7,
          i: `default${key}`,
          id: `id_${elId}`
        }
        this.layoutData.push(item)
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
    },
    goBack () {
      this.$router.push({name:'viewConfigIndex'})
    },
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
