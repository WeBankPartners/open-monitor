<template>
  <div>
    <section v-if="isPlugin">
      <div style="margin: 100px 0;text-align: center;font-size:14px">
        {{$t('tips.dashboardEmpty')}}
      </div>
    </section>
    <section v-else>
      <header>
        <div class="search-zone">
          <span class="params-title">{{$t('field.relativeTime')}}：</span>
          <Select filterable v-model="viewCondition.timeTnterval" :disabled="disableTime" style="width:80px"  @on-change="initPanals">
            <Option v-for="item in dataPick" :value="item.value" :key="item.value">{{ item.label }}</Option>
          </Select>
        </div>
        <div class="search-zone">
          <span class="params-title">{{$t('placeholder.refresh')}}：</span>
          <Select filterable clearable v-model="viewCondition.autoRefresh" :disabled="disableTime" style="width:100px" @on-change="initPanals" :placeholder="$t('placeholder.refresh')">
            <Option v-for="item in autoRefreshConfig" :value="item.value" :key="item.value">{{ item.label }}</Option>
          </Select>
        </div>
        <div class="search-zone">
          <span class="params-title">{{$t('field.timeInterval')}}：</span>
          <DatePicker 
            type="datetimerange" 
            :value="viewCondition.dateRange" 
            format="yyyy-MM-dd HH:mm:ss" 
            placement="bottom-start" 
            @on-change="datePick" 
            :placeholder="$t('placeholder.datePicker')" 
            style="width: 320px">
          </DatePicker>
        </div>
        <div style="float:right" class="search-zone">
          <div class="header-tools">
            <button v-if="!showAlarm" class="btn btn-sm btn-cancel-f" @click="openAlarmDisplay()">
              <i style="font-size: 18px;color: #0080FF;" class="fa fa-eye-slash" aria-hidden="true"></i>
            </button>
            <button v-else class="btn btn-sm btn-cancel-f" @click="closeAlarmDisplay()">
              <i style="font-size: 18px;color: #0080FF;" class="fa fa-eye" aria-hidden="true"></i>
            </button>
          </div>
        </div>
      </header>
      <div>
        <div class="radio-group">
          <div
            class="radio-group-radio radio-group-optional"
            :style="activeGroup === 'All' ? 'background: rgba(129, 179, 55, 0.6)' : 'background: #fff'"
          >
            <span @click="selectGroup('All')" style="vertical-align: text-bottom;">All</span>
          </div>
          <div
            v-for="(item, index) in panel_group_list"
            :key="index"
            class="radio-group-radio radio-group-optional"
            :style="item === activeGroup ? 'background: rgba(129, 179, 55, 0.6)' : 'background: #fff'"
          >
            <span @click="selectGroup(item)" style="vertical-align: text-bottom;">
              {{ `${item}` }}
            </span>
          </div>
        </div>
      </div>
      <div style="display:flex">
        <div class="grid-style">
          <grid-layout
            :layout.sync="tmpLayoutData"
            :col-num="12"
            :row-height="30"
            :is-draggable="false"
            :is-resizable="false"
            :is-mirrored="false"
            :vertical-compact="true"
            :use-css-transforms="true"
            >
            <grid-item v-for="(item,index) in tmpLayoutData"
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
                <div class="header-grid">
                  <Select v-model="item.group" style="width:100px;" size="small" disabled clearable filterable :placeholder="$t('m_group_name')">
                    <Option v-for="item in panel_group_list" :value="item" :key="item" style="float: left;">{{ item }}</Option>
                  </Select>
                </div>
              </div>
              <section>
                <div v-for="(chartInfo,chartIndex) in item._activeCharts" :key="chartIndex">
                  <CustomChart v-if="['line','bar'].includes(chartInfo.chartType)" :refreshNow="refreshNow" :chartInfo="chartInfo" :chartIndex="index" :params="viewCondition"></CustomChart>
                  <CustomPieChart v-if="chartInfo.chartType === 'pie'" :refreshNow="refreshNow" :chartInfo="chartInfo" :chartIndex="index" :params="viewCondition"></CustomPieChart>
                </div>
              </section>
            </grid-item>
          </grid-layout>
        </div>
        <div v-show="showAlarm" class="alarm-style">
          <ViewConfigAlarm ref="cutsomViewId"></ViewConfigAlarm>
        </div>
      </div>
    </section>
  </div>
</template>

<script>
import VueGridLayout from 'vue-grid-layout'
import {dataPick, autoRefreshConfig} from '@/assets/config/common-config'
import {resizeEvent} from '@/assets/js/gridUtils.ts'
import CustomChart from '@/components/custom-chart'
import CustomPieChart from '@/components/custom-pie-chart'
import ViewConfigAlarm from '@/views/custom-view/view-config-alarm'
export default {
  name: '',
  data() {
    return {
      refreshNow: false,
      isPlugin: false,
      viewCondition: {
        timeTnterval: -3600,
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

      showAlarm: false, // 显示告警信息
      cutsomViewId: null,
      activeGroup: 'All',
      panel_group_list: [] // 存放视图拥有的组信息
    }
  },
  mounted() {
    // this.getDashboardData()
  },
  props: ['id'],
  watch: {
    id: {
      immediate:true,
      handler:function(val){
        if (val) {
          this.getDashData(val)
        }
      }
    }
  },
  computed: {
    tmpLayoutData() { // 缓存切换分组后数据
      if (this.activeGroup === 'All') {
        return this.layoutData
      } else {
        return this.layoutData.filter(d => d.group === this.activeGroup)
      }
    }
  },
  methods: {
    openAlarmDisplay () {
      this.showAlarm = !this.showAlarm
      this.$refs.cutsomViewId.getAlarm(this.cutsomViewId, this.viewCondition)
    },
    closeAlarmDisplay () {
      this.showAlarm = !this.showAlarm
      this.$refs.cutsomViewId.clearAlarmInterval()
    },
    getDashData (id) {
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET',this.$root.apiCenter.template.singleDash, {id: id}, responseData => {
        if (responseData.cfg === '' || responseData.cfg === '[]') {
          if (window.request) {
            this.isPlugin = true
          } else {
            this.$router.push({path: 'portal'})
          }
        }else {
          this.viewData = JSON.parse(responseData.cfg)
          this.activeGroup = 'All'
          this.panel_group_list = responseData.panel_group_list || []
          this.initPanals()
          this.cutsomViewId = responseData.id
          // this.$refs.cutsomViewId.getAlarm(this.cutsomViewId, this.viewCondition)
        }
      })
    },
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
    getDashboardData () {
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET',this.$root.apiCenter.template.get, '', responseData => {
        if (responseData.cfg === '' || responseData.cfg === '[]') {
          if (window.request) {
            this.isPlugin = true
          } else {
            this.$router.push({path: 'portal'})
          }
        }else {
          this.viewData = JSON.parse(responseData.cfg) 
          this.initPanals()
          this.cutsomViewId = responseData.id
          this.$refs.cutsomViewId.getAlarm(this.cutsomViewId, this.viewCondition)
        }
      })
    },
    dateToTimestamp (date) {
      if (!date) return 0
      let timestamp = Date.parse(new Date(date))
      timestamp = timestamp / 1000;
      return timestamp
    },
    initPanals () {
      let tmp = []
      this.viewData.forEach((item) => {
        let params = {
          aggregate: item.aggregate,
          agg_step: item.agg_step,
          lineType: item.lineType,
          time_second: this.viewCondition.timeTnterval,
          start: this.dateToTimestamp(this.viewCondition.dateRange[0]),
          end: this.dateToTimestamp(this.viewCondition.dateRange[1]),
          title: '',
          unit: '',
          data: []
        }
        item.query.forEach( _ => {
          params.data.push(_)
        })
        let height = (item.viewConfig.h) * 30
        let _activeCharts = []
        _activeCharts.push({
          style: `height:${height}px;`,
          panalUnit: item.panalUnit,
          elId: item.viewConfig.id,
          chartParams: params,
          chartType: item.chartType,
          aggregate: item.aggregate                                                      
        })
        item.viewConfig._activeCharts = _activeCharts
        tmp.push(item.viewConfig)
      })
      this.layoutData = tmp
    },
    resizeEvent: function(i, newH, newW, newHPx, newWPx){
      resizeEvent(this, i, newH, newW, newHPx, newWPx)
    },
    //#region 组管理
    selectGroup (item) {
      this.activeGroup = item
      this.refreshNow = true
      this.$nextTick(() => {
        this.refreshNow = false
      })
    },
    //#endregion
  },
  components: {
    GridLayout: VueGridLayout.GridLayout,
    GridItem: VueGridLayout.GridItem,
    CustomChart,
    CustomPieChart,
    ViewConfigAlarm
  },
}
</script>

<style scoped lang="less">
.grid-style {
  width: 100%;
  display: inline-block;
}
.alarm-style {
  width: 800px;
  display: inline-block;
}

header {
  margin: 16px 8px;
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
  text-align: right;
  line-height: 32px;
  i {
    margin: 0 4px;
    cursor: pointer;
  } 
}
.vue-grid-item {
  border-radius: 4px;
}
.vue-grid-item:not(.vue-grid-placeholder) {
    background: @gray-f;
    border: 1px solid @gray-f;
}

.radio-group {
  margin-bottom: 15px;
}
.radio-group-radio {
  padding: 5px 15px;
  border-radius: 32px;
  font-size: 12px;
  cursor: pointer;
  margin: 4px;
  display: inline-block;
}

.radio-group-optional {
  border: 1px solid #81b337;
  color: #81b337;
}
</style>
