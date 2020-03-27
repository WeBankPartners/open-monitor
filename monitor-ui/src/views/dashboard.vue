<template>
  <div>
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
        @resized="resizeEvent">
                 
        <div class="c-dark" style="display:flex;justify-content:flex-end;padding:0 32px;">
          <div class="header-grid header-grid-name">
            {{item.i}}
          </div>
        </div>
        <section class="metric-section">
            <div :id="item.id" class="echart" :style="{height: ((item.h +1)*30)+'px'}"></div>
        </section>
        </grid-item>
      </grid-layout>
  </div>
</template>

<script>
import VueGridLayout from 'vue-grid-layout'
import {readyToDraw} from  '@/assets/config/chart-rely'
import {resizeEvent} from '@/assets/js/gridUtils'
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
        params.push({
          endpoint: item.endpoint,
          metric: item.metricLabel,
          prom_ql: item.metric,
          time: '-1800'
        })
      })
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST',this.$root.apiCenter.metricConfigView.api, params, responseData => {
        
        responseData.yaxis.unit =  panalUnit  
        this.elId = id
        const chartConfig = {eye: false,dataZoom:false}
        readyToDraw(this,responseData, viewIndex, chartConfig)
      })
    },
    resizeEvent: function(i, newH, newW, newHPx, newWPx){
      resizeEvent(this, i, newH, newW, newHPx, newWPx)
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
.vue-grid-item {
  border-radius: 4px;
}
.vue-grid-item:not(.vue-grid-placeholder) {
    background: @gray-f;
    border: 1px solid @gray-f;
}
</style>
