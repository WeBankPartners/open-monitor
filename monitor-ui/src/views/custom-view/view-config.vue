<template>
  <div>
    <div class=" ">
      <Title :title="$t('menu.templateManagement')"></Title>
      <header>
        <div class="header-name">
          <i class="fa fa-th-large fa-18" aria-hidden="true"></i>
          <template v-if="isEditPanal">
            <Input v-model.trim="panalName" style="width: 100px" type="text"></Input>
            <Icon class="panal-edit-icon" @click="savePanalEdit" type="md-checkmark" />
            <Icon class="panal-edit-icon" @click="canclePanalEdit" type="md-close" />
          </template>
          <template v-else>
            <span> {{panalName}}</span>
            <Icon class="panal-edit-icon" @click="isEditPanal=true" v-if="permission === 'edit'" type="md-create" />
          </template>
        </div>
        <div class="search-container">
          <div>
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
            <!-- <div class="search-zone">
              <span class="params-title">{{$t('field.aggType')}}：</span>
              <RadioGroup v-model="viewCondition.agg" @on-change="initPanals" size="small" type="button">
                <Radio disabled label="min">Min</Radio>
                <Radio disabled label="max">Max</Radio>
                <Radio disabled label="avg">Average </Radio>
                <Radio disabled label="p95">P95</Radio>
                <Radio disabled label="sum">Sum</Radio>
                <Radio disabled label="none">Original</Radio>
              </RadioGroup>
            </div> -->
          </div>

          <div class="header-tools">
            <template v-if="permission === 'edit'">
              <button class="btn btn-sm btn-cancel-f" style="margin-right:60px" @click="addItem">{{$t('m_new_graph')}}</button>
              <button class="btn btn-sm btn-confirm-f" @click="saveEdit">{{$t('button.saveConfig')}}</button>
            </template> 
            
            <button class="btn btn-sm btn-cancel-f" @click="goBack()">{{$t('button.back')}}</button>
            <button v-if="!showAlarm" class="btn btn-sm btn-cancel-f" @click="openAlarmDisplay()">
              <i style="font-size: 18px;color: #0080FF;" class="fa fa-eye-slash" aria-hidden="true"></i>
            </button>
            <button v-else class="btn btn-sm btn-cancel-f" @click="closeAlarmDisplay()">
              <i style="font-size: 18px;color: #0080FF;" class="fa fa-eye" aria-hidden="true"></i>
            </button>
          </div>
        </div>
      </header>
      <div style="display:flex">
        <div class="grid-style">
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
                        
              <div class="c-dark" style="display:flex;padding:0 32px;">
                <div class="header-grid header-grid-name">
                  <span v-if="editChartId !== item.id">{{item.i}}</span>
                  <Input v-else v-model="item.i" class="editChartId" style="width:100px" @on-blur="editChartId = null" size="small" placeholder="small size" />
                  <Tooltip :content="$t('placeholder.editTitle')" theme="light" transfer placement="top">
                    <i class="fa fa-pencil-square" style="font-size: 16px;" v-if="permission === 'edit'" @click="editChartId = item.id" aria-hidden="true"></i>
                  </Tooltip>
                </div>
                <div class="header-grid header-grid-tools"> 
                  <Tooltip :content="$t('button.chart.dataView')" theme="light" transfer placement="top">
                    <i class="fa fa-eye" style="font-size: 16px;" v-if="isShowGridPlus(item)" aria-hidden="true" @click="gridPlus(item)"></i>
                  </Tooltip>
                  <Tooltip :content="$t('placeholder.chartConfiguration')" theme="light" transfer placement="top">
                    <i class="fa fa-cog" style="font-size: 16px;" v-if="permission === 'edit'" @click="setChartType(item)" aria-hidden="true"></i>
                  </Tooltip>
                  <Tooltip :content="$t('placeholder.deleteChart')" theme="light" transfer placement="top">
                    <i class="fa fa-trash" style="font-size: 16px;" v-if="permission === 'edit'" @click="removeGrid(item)" aria-hidden="true"></i>
                  </Tooltip>
                </div>
              </div>
              <section>
                <div v-for="(chartInfo,chartIndex) in item._activeCharts" :key="chartIndex">
                  <CustomChart v-if="['line','bar'].includes(chartInfo.chartType)" :chartInfo="chartInfo" :chartIndex="index" :params="viewCondition"></CustomChart>
                  <CustomPieChart v-if="chartInfo.chartType === 'pie'" :chartInfo="chartInfo" :chartIndex="index" :params="viewCondition"></CustomPieChart>
                </div>
              </section>
            </grid-item>
          </grid-layout>
        </div>
        <div v-show="showAlarm" class="alarm-style">
          <ViewConfigAlarm ref="cutsomViewId"></ViewConfigAlarm>
        </div>
      </div>
    </div>
    <Drawer title="View details" :width="zoneWidth" v-model="showMaxChart">
      <ViewChart ref="viewChart"></ViewChart>
    </Drawer>
    <Drawer :title="$t('placeholder.chartConfiguration')" :width="zoneWidth" :mask-closable="false" v-model="showChartConfig">
      <editPieView v-if="chartType === 'pie'" ref="editPieView"></editPieView>
      <editLineView v-else ref="editLineView"></editLineView>
    </Drawer>
    <ModalComponent :modelConfig="setChartTypeModel">
      <div slot="setChartType">
        <div style="display:flex;justify-content:center">
          <i @dblclick="dblChartType('bar')" @click="choiceChartType('bar')" :class="['fa', 'fa-line-chart', activeChartType==='bar' ? 'active-tag': '']" aria-hidden="true"></i>
          <i @dblclick="dblChartType('pie')" @click="choiceChartType('pie')" :class="['fa', 'fa-pie-chart', activeChartType==='pie' ? 'active-tag': '']" aria-hidden="true"></i>
        </div>
      </div>
    </ModalComponent>
    <Modal
      v-model="isShowWarning"
      :title="$t('delConfirm.title')"
      @on-ok="confirmRemoveGrid"
      @on-cancel="cancel">
      <div class="modal-body" style="padding:30px">
        <div style="text-align:center">
          <p style="color: red">{{$t('delConfirm.tip')}}</p>
        </div>
      </div>
    </Modal>
  </div>

</template>
<style lang="less" scoped>
  .grid-style {
    width: 100%;
    display: inline-block;
  }
  .alarm-style {
    width: 800px;
    display: inline-block;
  }
  .fa-line-chart, .fa-pie-chart {
    cursor: pointer;
    font-size: 36px;
    padding: 24px 48px;
    border: 1px solid @gray-d;
    margin: 8px;
    border-radius: 4px;
  }
  .fa-line-chart:hover, .fa-pie-chart:hover {
    box-shadow: 0 1px 8px @gray-d;
    border-color: @blue-2;
  }
  .active-tag {
    color: @blue-2;
    border-color: @blue-2;
  } 
  .i-icon-menu-fold:before {
    content: "\E600";
  }
</style>
<script>
import {generateUuid} from '@/assets/js/utils'
import {dataPick, autoRefreshConfig} from '@/assets/config/common-config'
import {resizeEvent} from '@/assets/js/gridUtils.ts'
import VueGridLayout from 'vue-grid-layout'
import CustomChart from '@/components/custom-chart'
import CustomPieChart from '@/components/custom-pie-chart'
import ViewConfigAlarm from '@/views/custom-view/view-config-alarm'
import ViewChart from '@/views/custom-view/view-chart'
import editLineView from '@/views/custom-view/edit-line-view'
import editPieView from '@/views/custom-view/edit-pie-view'
export default {
  name: '',
  data() {
    return {
      editData: null,
      isShowWarning: false,
      deleteConfirm: {
        id: '',
        method: ''
      },
      isEditPanal: false,
      permission: '',
      panalName: this.$route.params.panalItem.name,
      viewCondition: {
        timeTnterval: -3600,
        dateRange: ['', ''],
        autoRefresh: 5,
        agg: 'none' // 聚合类型
      },
      disableTime: false,
      dataPick: dataPick,
      autoRefreshConfig: autoRefreshConfig,
      viewData: [],
      layoutData: [
        //   {'x':0,'y':0,'w':2,'h':2,'i':'0'},
        //   {'x':1,'y':1,'w':2,'h':2,'i':'1'},
      ],
      editChartId: null,
      setChartTypeModel: {
        modalId: 'set_chart_type_Modal',
        modalTitle: 'button.add',
        isAdd: true,
        config: [
          {name:'setChartType',type:'slot'}
        ],
        addRow: {
          type: null
        },
        modalFooter: [
          {name: '确定', Func: 'confirmChartType'}
        ]
      },
      activeGridConfig: null,
      activeChartType: 'bar',

      showAlarm: false, // 显示告警信息
      cutsomViewId: null,

      showMaxChart: false,
      zoneWidth: '800',
      showChartConfig: false,
      chartType: ''
    }
  },
  created () {
    this.zoneWidth = window.screen.width * 0.65
  },
  mounted() {
    this.reloadPanal(this.$route.params.panalItem)
  },
  methods: {
    reloadPanal (params) {
      this.permission = this.$route.params.permission
      this.editData = params
      if(this.$root.$validate.isEmpty_reset(params)) {
        this.$router.push({path:'viewConfigIndex'})
      } else {
        if (!this.$root.$validate.isEmpty_reset(params.cfg)) {
          this.viewData = JSON.parse(params.cfg)
          this.initPanals()
          this.cutsomViewId = params.id
          // this.$refs.cutsomViewId.getAlarm(this.cutsomViewId, this.viewCondition)
        }
      }
    },
    openAlarmDisplay () {
      this.showAlarm = !this.showAlarm
      this.$refs.cutsomViewId.getAlarm(this.cutsomViewId, this.viewCondition, this.permission)
    },
    closeAlarmDisplay () {
      this.showAlarm = !this.showAlarm
      this.$refs.cutsomViewId.clearAlarmInterval()
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
        let height = (item.viewConfig.h+1) * 30-8
        let _activeCharts = []
        _activeCharts.push({
          style: `height:${height}px;`,
          panalUnit: item.panalUnit,
          elId: item.viewConfig.id,
          chartParams: params,
          chartType: item.chartType,
          aggregate: item.aggregate,
          agg_step: item.agg_step,
          lineType: item.lineType,
          time_second: this.viewCondition.timeTnterval,
          start: this.dateToTimestamp(this.viewCondition.dateRange[0]),
          end: this.dateToTimestamp(this.viewCondition.dateRange[1])                                  
        })
        item.viewConfig._activeCharts = _activeCharts
        tmp.push(item.viewConfig)
      })
      this.layoutData = tmp
    },
    isShowGridPlus (item) {
      // 新增及饼图时屏蔽放大功能
      if (!item._activeCharts || item._activeCharts[0].chartType === 'pie') {
        return false
      }
      return true
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
    setChartType (item) {
      const find = this.layoutData.find(i => i.id === item.id)
      this.activeGridConfig = find
      if (!find._activeCharts) {
        this.$root.JQ('#set_chart_type_Modal').modal('show')
      } else {
        this.activeChartType = find._activeCharts[0].chartType
        this.chartType = find._activeCharts[0].chartType
        this.editGrid(item)
      }
    },
    choiceChartType (activeChartType) {
      this.activeChartType = activeChartType
      this.chartType = activeChartType
    },
    dblChartType (activeChartType) {
      this.activeChartType = activeChartType
      this.chartType = activeChartType
      this.confirmChartType()
    },
    confirmChartType () {
      if (!this.activeChartType) {
        this.$Message.warning('请先设置图标类型！')
        return
      }
      this.$root.JQ('#set_chart_type_Modal').modal('hide')
      this.editGrid()
    },
    editGrid(item) {
      this.modifyLayoutData().then((resViewData)=>{
        let parentRouteData = this.editData
        const cfg = JSON.parse(parentRouteData.cfg)
        parentRouteData.cfg = JSON.parse(parentRouteData.cfg)
        const oriConfig = JSON.parse(JSON.stringify(cfg))
        let aggregate = 'none'
        let agg_step = 60
        if (item) {
          const find = oriConfig.find(xItem => xItem.viewConfig.id === item.id)
          if (find) {
            aggregate = find.aggregate || 'none'
            agg_step = find.agg_step || 60
          }
          let findEditData = parentRouteData.cfg.find(xItem => xItem.viewConfig.id === item.id)
          findEditData.aggregate = aggregate
          findEditData.agg_step = agg_step
        } else {
          parentRouteData.cfg = resViewData
        }
        parentRouteData.cfg = JSON.stringify(parentRouteData.cfg)
        if (['line','bar'].includes(this.activeChartType)) {
          this.chartType = 'line'
          this.$refs.editLineView.initChart({templateData: parentRouteData, panal:this.activeGridConfig})
        } else {
          this.chartType = 'pie'
          this.$refs.editPieView.initChart({templateData: parentRouteData, panal:this.activeGridConfig})
        }
        this.showChartConfig = true
      })
    },
    removeGrid(itemxxx) {
      this.isShowWarning = true
      this.deleteConfirm.id = itemxxx.id
    },
    confirmRemoveGrid () {
      this.layoutData.forEach((item,index) => {
        if (item.id ===  this.deleteConfirm.id) {
          this.layoutData.splice(index,1)
        }
      })
    },
    cancel () {
      this.isShowWarning = false
    },
    async gridPlus(item) {
      const resViewData = await this.modifyLayoutData()
      let parentRouteData = this.editData
      parentRouteData.cfg = JSON.stringify(resViewData)
      this.showMaxChart = true
      this.$refs.viewChart.initChart({templateData: parentRouteData, panal:item, parentData: this.editData})
    },
    async modifyLayoutData() {
      var resViewData = []
      this.layoutData.forEach((layoutDataItem) =>{
        let temp = {
          panalTitle: layoutDataItem.i,
          panalUnit: '',
          chartType: this.activeChartType,
          query: [],
          viewConfig: layoutDataItem
        }

        this.viewData.forEach((i) =>{
          if (layoutDataItem.id === i.viewConfig.id) {
            temp.panalUnit = i.panalUnit
            temp.query = i.query
            temp.chartType = i.chartType
            temp.aggregate = i.aggregate
            temp.agg_step = i.agg_step
            temp.lineType = i.lineType
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
              chartType: i.chartType,
              aggregate: i.aggregate,
              agg_step: i.agg_step,
              lineType: i.lineType,
              viewConfig: layoutDataItem
            })
          }
        })
      })
      let params = {
        name: this.panalName,
        id: this.editData.id,
        cfg: JSON.stringify(res)
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST',this.$root.apiCenter.template.save, params, () => {
        this.isEditPanal = false
        this.$Message.success(this.$t('tips.success'))
      })
    },
    goBack () {
      this.$router.push({name:'viewConfigIndex'})
    },
    savePanalEdit () {
      if (!this.panalName) {
        this.$Message.warning(this.$t('tips.required'))
        return
      }
      this.saveEdit()
    },
    canclePanalEdit () {
      this.isEditPanal = false
      this.panalName = this.editData.name
    }
  },
  components: {
    GridLayout: VueGridLayout.GridLayout,
    GridItem: VueGridLayout.GridItem,
    CustomChart,
    CustomPieChart,
    ViewConfigAlarm,
    ViewChart,
    editPieView,
    editLineView
  },
}
</script>

<style scoped lang="less">
.panal-edit-icon {
  margin-left:4px;
  padding: 4px;
}
.header-name {
  font-size: 16px; 
}
.search-container {
  display: flex;
  justify-content: space-between;
  margin: 8px;
  font-size: 16px;
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
