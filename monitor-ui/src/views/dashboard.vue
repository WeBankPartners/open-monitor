<template>
  <div class=" ">
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
        class="c-dark"
        :x="item.x"
        :y="item.y"
        :w="item.w"
        :h="item.h"
        :i="item.i"
        :key="index"
        @resized="resizedEvent">
                 
        <div class="c-dark" style="display:flex;justify-content:flex-end;padding:0 32px;">
          <div class="header-grid header-grid-name">
            {{item.i}}
          </div>
        </div>
        <div class="">
          <section class="metric-section">
            <div>
              <div :id="item.id" class="echart" style=""></div>
            </div>
          </section>
        </div>
        </grid-item>
      </grid-layout>
  </div>
</template>

<script>
import VueGridLayout from 'vue-grid-layout'
import {readyToDraw} from  '@/assets/config/chart-rely'
const echarts = require('echarts/lib/echarts');
export default {
  name: '',
  data() {
    return {
      viewData: [],
      layoutData: [
        //   {'x':0,'y':0,'w':2,'h':2,'i':'0'},
        //   {'x':1,'y':1,'w':2,'h':2,'i':'1'},
      ]
    }
  },
  mounted() {
    this.getDashboardData()
  },
  methods: {
    getDashboardData () {
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET',this.$root.apiCenter.template.get, '', responseData => {
        if (responseData.cfg === '' || responseData.cfg === '[]') {
          this.$router.push({path: 'portal'}) 
        }else {
          this.viewData = JSON.parse(responseData.cfg) 
          this.initPanals()
        }
      })
    },
    initPanals () {
      this.viewData.forEach((item,viewIndex) => {
        this.layoutData.push(item.viewConfig)
        this.requestChart(item.viewConfig.id,item.panalUnit, item.query,viewIndex)
      })
    },
    requestChart (id, panalUnit, query,viewIndex) {
      let params = []
      query.forEach((item) => {
        params.push(JSON.stringify({
          endpoint: item.endpoint,
          metric: item.metricLabel,
          prom_ql: item.metric,
          time: '-1800'
        })) 
      })
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET',this.$root.apiCenter.metricConfigView.api, {config: `[${params.join(',')}]`}, responseData => {
        
        responseData.yaxis.unit =  panalUnit  
        this.elId = id
        const chartConfig = {eye: false,dataZoom:false}
        readyToDraw(this,responseData, viewIndex, chartConfig)
      })
    },
    resizeEvent: function(i, newH, newW, newHPx, newWPx){
      this.layoutData.forEach((item,index) => {
        if (item.i === i) {
          this.layoutData[index].h = newH
          this.layoutData[index].w = newW
          var myChart = echarts.init(document.getElementById(item.id))
          myChart.resize({height:newHPx-34+'px',width:newWPx+'px'})
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
    text-align: center;
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
</style>
