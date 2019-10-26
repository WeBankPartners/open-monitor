<template>
  <div class=" ">
      <header>
        <div style="display:flex;justify-content:space-between; font-size:16px;padding:8px 16px">
            <div class="header-name">
                <i class="fa fa-th-large fa-18" aria-hidden="true"></i>
                <span> {{$route.params.name}}</span>
                <!-- <i class="fa fa-backward fa-18" @click="goBack" aria-hidden="true"></i> -->
            </div>
            <div class="header-tools"> 
              <button class="btn btn-sm btn-confirm-f" @click="addItem">{{$t('button.add')}}</button>
              <button class="btn btn-sm btn-cancle-f" @click="goBack()">{{$t('button.back')}}</button>
                <!-- <i class="fa fa-plus-square-o fa-18"  aria-hidden="true"></i> -->
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
            {{item.i}}
          </div>
          <div class="header-grid header-grid-tools"> 
            <!-- <i class="fa fa-eye" aria-hidden="true" @click="gridPlus(item)"></i> -->
            <i class="fa fa-cog" @click="editGrid(item)" aria-hidden="true"></i>
            <i class="fa fa-trash" @click="removeGrid(item)" aria-hidden="true"></i>
          </div>
        </div>
        <div class="">
          <section class="metric-section">
            <div v-if="!noDataTip">
              <div :id="item.id" class="echart" style="height:230px;width:560px"></div>
            </div>
             <div v-else class="echart echart-no-data-tip">
              <span>~~~暂无数据~~~</span>
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
      noDataTip: false
    }
  },
  mounted() {
    if(this.$validate.isEmpty_reset(this.$route.params)) {
      this.$router.push({path:'viewConfigIndex'})
    } else {
      if (!this.$validate.isEmpty_reset(this.$route.params.cfg)) {
        this.viewData = JSON.parse(this.$route.params.cfg)
        this.initPanals()
      }
    }
  },
  methods: {
    initPanals () {
      this.viewData.forEach((item) => {
        this.layoutData.push(item.viewConfig)
        this.requestChart(item.viewConfig.id, item.query)
      })
    },
    requestChart (id, query) {
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
        // const colorSet = ['#487e89', '#153863', '#395b79',  '#153250']
        responseData.series.forEach((item)=>{
          legend.push(item.name)
          item.symbol = 'none'
          item.smooth = true
          item.lineStyle = {
            width: 1
          }
          item.itemStyle = {
            normal:{
              // color: colorSet[index]
            }
          }
          item.areaStyle = {
            normal: {
              // color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [{
              //     offset: 0,
              //     color: colorSet[index]
              // }, {
              //     offset: 1,
              //     color: colorSet[index]
              // }])
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
    editGrid(item) {
      this.modifyLayoutData().then((resViewData)=>{
        let parentRouteData = this.$route.params
        parentRouteData.cfg = JSON.stringify(resViewData) 
        this.$router.push({name: 'editView', params:{templateData: parentRouteData, panal:item, parentData: this.$route.params}}) 
      })
    },
    removeGrid(itemxxx) {
      this.layoutData.forEach((item,index) => {
        if (item.id === itemxxx.id) {
         this.layoutData.splice(index,1)
         return
        }
      })
    },
    gridPlus() {

    },
    modifyLayoutData() {
      return new Promise(resolve => {
        var resViewData = []
        this.layoutData.forEach((layoutDataItem) =>{
          let temp = {
            panalTitle: layoutDataItem.i,
            panalUnit: '',
            query: [],
            viewConfig: layoutDataItem
          }
          this.viewData.forEach((i) =>{
            if (layoutDataItem.id === i.viewConfig.id) {
              temp.panalUnit = i.panalUnit
              temp.query = i.query
            }
          })
          resViewData.push(temp)
        })
        resolve(resViewData)
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
      // border: 1px solid red;
      // font-size: 14px
    } 
  }
</style>
<style scoped lang="less">

.columns {
    -moz-columns: 120px;
    -webkit-columns: 120px;
    columns: 120px;
}

.vue-grid-item:not(.vue-grid-placeholder) {
    background: @gray-f;
    border: 1px solid @gray-f;
}

.vue-grid-item.resizing {
    opacity: 0.9;
}

.vue-grid-item.static {
    background: #cce;
}

.vue-grid-item .text {
    font-size: 24px;
    text-align: center;
    position: absolute;
    top: 0;
    bottom: 0;
    left: 0;
    right: 0;
    margin: auto;
    height: 100%;
    width: 100%;
}

.vue-grid-item .no-drag {
    height: 100%;
    width: 100%;
}

.vue-grid-item .minMax {
    font-size: 12px;
}

.vue-grid-item .add {
    cursor: pointer;
}

.vue-draggable-handle {
    position: absolute;
    width: 20px;
    height: 20px;
    top: 0;
    left: 0;
    background: url("data:image/svg+xml;utf8,<svg xmlns='http://www.w3.org/2000/svg' width='10' height='10'><circle cx='5' cy='5' r='5' fill='#999999'/></svg>") no-repeat;
    background-position: bottom right;
    padding: 0 8px 8px 0;
    background-repeat: no-repeat;
    background-origin: content-box;
    box-sizing: border-box;
    cursor: pointer;
}
</style>
