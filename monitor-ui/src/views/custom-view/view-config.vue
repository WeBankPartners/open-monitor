<template>
  <div class=" ">
    <Title :title="$t('menu.templateManagement')"></Title>
    <header>
      <div class="header-name" >
        <i class="fa fa-th-large fa-18" aria-hidden="true"></i>
        <span> {{$route.params.name}}</span>
      </div>
      <div class="search-container">
          <div>
            <div class="search-zone">
              <span class="params-title">{{$t('field.relativeTime')}}：</span>
              <Select v-model="viewCondition.timeTnterval" :disabled="disableTime" style="width:80px"  @on-change="initPanals">
                <Option v-for="item in dataPick" :value="item.value" :key="item.value">{{ item.label }}</Option>
              </Select>
            </div>
            <div class="search-zone">
              <span class="params-title">{{$t('placeholder.refresh')}}：</span>
              <Select v-model="viewCondition.autoRefresh" :disabled="disableTime" style="width:100px" @on-change="initPanals" :placeholder="$t('placeholder.refresh')">
                <Option v-for="item in autoRefreshConfig" :value="item.value" :key="item.value">{{ item.label }}</Option>
              </Select>
            </div>
            <div class="search-zone">
              <span class="params-title">{{$t('field.timeInterval')}}：</span>
              <DatePicker 
                type="datetimerange" 
                :value="viewCondition.dateRange" 
                format="yyyy-MM-dd HH:mm:ss" 
                placement="bottom-end" 
                @on-change="datePick" 
                :placeholder="$t('placeholder.datePicker')" 
                style="width: 320px">
              </DatePicker>
            </div>
          </div>

          <div class="header-tools"> 
            <button class="btn btn-sm btn-cancle-f" @click="addItem">{{$t('button.add')}}</button>
            <button class="btn btn-sm btn-confirm-f" @click="saveEdit">{{$t('button.saveEdit')}}</button>
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
      class="c-dark"
      :x="item.x"
      :y="item.y"
      :w="item.w"
      :h="item.h"
      :i="item.i"
      :key="index"
      @resize="resizeEvent"
      @resized="resizeEvent">
                
      <div class="c-dark" style="display:flex;justify-content:flex-end;padding:0 32px;">
        <div class="header-grid header-grid-name">
          <span v-if="editChartId !== item.id">{{item.i}}</span>
          <Input v-else v-model="item.i" class="editChartId" style="width:100px" @on-blur="editChartId = null" size="small" placeholder="small size" />
          <Tooltip :content="$t('placeholder.editTitle')" theme="light" transfer placement="top">
            <i class="fa fa-pencil-square" @click="editChartId = item.id" aria-hidden="true"></i>
          </Tooltip>
        </div>
        <div class="header-grid header-grid-tools"> 
          <Tooltip :content="$t('button.chart.dataView')" theme="light" transfer placement="top">
            <i class="fa fa-eye" aria-hidden="true" @click="gridPlus(item)"></i>
          </Tooltip>
          <Tooltip :content="$t('placeholder.chartConfiguration')" theme="light" transfer placement="top">
            <i class="fa fa-cog" @click="editGrid(item)" aria-hidden="true"></i>
          </Tooltip>
          <Tooltip :content="$t('placeholder.deleteChart')" theme="light" transfer placement="top">
            <i class="fa fa-trash" @click="removeGrid(item)" aria-hidden="true"></i>
          </Tooltip>
        </div>
      </div>
      <section>
        <div v-for="(chartItemx,chartIndexx) in item.activeCharts" :key="chartIndexx">
          <CustomChart :chartItemx="chartItemx" :chartIndex="index" :key="chartIndexx" :params="viewCondition"></CustomChart>
        </div>
      </section>
      </grid-item>
    </grid-layout>
  </div>
</template>

<script>
import {generateUuid} from '@/assets/js/utils'
import {dataPick, autoRefreshConfig} from '@/assets/config/common-config'
import {resizeEvent} from '@/assets/js/gridUtils'
import VueGridLayout from 'vue-grid-layout'
import CustomChart from '@/components/custom-chart'
export default {
  name: '',
  data() {
    return {
      viewCondition: {
        timeTnterval: -1800,
        dateRange: ['', ''],
        autoRefresh: 10,
      },
      disableTime: false,
      dataPick: dataPick,
      autoRefreshConfig: autoRefreshConfig,
      viewData: [],
      layoutData: [
        //   {'x':0,'y':0,'w':2,'h':2,'i':'0'},
        //   {'x':1,'y':1,'w':2,'h':2,'i':'1'},
      ],
      editChartId: null
    }
  },
  mounted() {
    if(this.$root.$validate.isEmpty_reset(this.$route.params)) {
      this.$router.push({path:'viewConfigIndex'})
    } else {
      if (!this.$root.$validate.isEmpty_reset(this.$route.params.cfg)) {
        this.viewData = JSON.parse(this.$route.params.cfg)
        this.initPanals()
      }
    }
  },
  methods: {
    datePick (data) {
      this.viewCondition.dateRange = data
      this.disableTime = false
      if (this.viewCondition.dateRange[0] && this.viewCondition.dateRange[1]) {
        if (this.viewCondition.dateRange[0] === this.viewCondition.dateRange[1]) {
          this.viewCondition.dateRange[1] = this.viewCondition.dateRange[1].replace('00:00:00', '23:59:59')
        }
        this.disableTime = true
        this.viewCondition.autoRefresh = 0
      }
      this.initPanals()
    },
    initPanals () {
      let tmp = []
      this.viewData.forEach((item) => {
        let params = []
        item.query.forEach( _ => {
          params.push({
            endpoint: _.endpoint,
            metric: _.metricLabel,
            prom_ql: _.metric,
          })
        })
        let height = (item.viewConfig.h) * 30
        let activeCharts = []
        activeCharts.push({
          style: `height:${height}px;`,
          panalUnit: item.panalUnit,
          elId: item.viewConfig.id,
          chartParams: params                                                      
        })
        item.viewConfig.activeCharts = activeCharts
        tmp.push(item.viewConfig)
      })
      this.layoutData = tmp
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
        this.$router.push({name: 'editView', params:{templateData: parentRouteData, panal:item}}) 
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
    async gridPlus(item) {
      const resViewData = await this.modifyLayoutData()
      let parentRouteData = this.$route.params
      parentRouteData.cfg = JSON.stringify(resViewData) 
      this.$router.push({name: 'viewChart', params:{templateData: parentRouteData, panal:item, parentData: this.$route.params}}) 
    },
    async modifyLayoutData() {
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
      return resViewData
    },
    resizeEvent: function(i, newH, newW, newHPx, newWPx){
      resizeEvent(this, i, newH, newW, newHPx, newWPx)
    },
    saveEdit() {
      let res = []
      this.layoutData.forEach((layoutDataItem) =>{
        this.viewData.forEach((i) =>{
          if (layoutDataItem.id === i.viewConfig.id) {
            res.push({
              panalTitle: i.panalTitle,
              panalUnit: i.panalUnit,
              query: i.query,
              viewConfig: layoutDataItem
            })
          }
        })
      })
      let params = {
        name: this.$route.params.name,
        id: this.$route.params.id,
        cfg: JSON.stringify(res)
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST',this.$root.apiCenter.template.save, params, () => {
        this.$Message.success(this.$t('tips.success'))
      })
    },
    goBack () {
      this.$router.push({name:'viewConfigIndex'})
    },
  },
  components: {
    GridLayout: VueGridLayout.GridLayout,
    GridItem: VueGridLayout.GridItem,
    CustomChart
  },
}
</script>

<style scoped lang="less">
  .header-name {
    font-size: 16px; 
  }
  .search-container {
    display: flex;
    justify-content: space-between;
    font-size: 16px;
    padding: 0 10px;
  }
  .search-zone {
    display: inline-block;
  }
  .params-title {
    margin-left: 4px;
    font-size: 13px;
  }
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
.vue-grid-item {
  border-radius: 4px;
}
.vue-grid-item:not(.vue-grid-placeholder) {
    background: @gray-f;
    border: 1px solid @gray-f;
}
.echart-no-data-tip {
  text-align: center;
}
</style>
