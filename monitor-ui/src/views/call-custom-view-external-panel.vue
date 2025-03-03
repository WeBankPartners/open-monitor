<template>
  <div>
    <div style="margin-top: 12px;">
      <!-- <Title :title="$t('m_menu_templateManagement')"></Title> -->
      <header>
        <div class="header-name">
          <i class="fa fa-th-large fa-18" aria-hidden="true"></i>
          <span style="margin:0 4px"> {{panalName}}</span>
        </div>
        <div class="search-container">
          <div>
            <div class="search-zone">
              <span class="params-title">{{$t('m_field_relativeTime')}}：</span>
              <Select filterable v-model="viewCondition.timeTnterval" :disabled="disableTime" style="width:80px"  @on-change="initPanals">
                <Option v-for="item in dataPick" :value="item.value" :key="item.value">{{ item.label }}</Option>
              </Select>
            </div>
            <div class="search-zone">
              <span class="params-title">{{$t('m_placeholder_refresh')}}：</span>
              <Select filterable clearable v-model="viewCondition.autoRefresh" :disabled="disableTime" style="width:100px" @on-change="initPanals" :placeholder="$t('m_placeholder_refresh')">
                <Option v-for="item in autoRefreshConfig" :value="item.value" :key="item.value">{{ item.label }}</Option>
              </Select>
            </div>
            <div class="search-zone">
              <span class="params-title">{{$t('m_field_timeInterval')}}：</span>
              <DatePicker
                type="datetimerange"
                :value="viewCondition.dateRange"
                format="yyyy-MM-dd HH:mm:ss"
                split-panels
                placement="bottom-start"
                @on-change="datePick"
                :placeholder="$t('m_placeholder_datePicker')"
                style="width: 320px"
              >
              </DatePicker>
            </div>
            <!-- <div class="search-zone">
              <span class="params-title">{{$t('m_field_aggType')}}：</span>
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
            <Button v-if="!showAlarm" @click="openAlarmDisplay()">
              <i style="font-size: 18px;color: #0080FF;" class="fa fa-eye-slash" aria-hidden="true"></i>
            </Button>
            <Button v-else @click="closeAlarmDisplay()">
              <i style="font-size: 18px;color: #0080FF;" class="fa fa-eye" aria-hidden="true"></i>
            </Button>
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
                       @resize="resizeEvent"
                       @resized="resizeEvent"
            >

              <div class="c-dark" style="display:flex;padding:0 32px;">
                <div class="header-grid header-grid-name">
                  <span>{{item.i}}</span>
                </div>
                <div class="header-grid header-grid-tools">
                  <Select v-model="item.group" style="width:100px;" disabled size="small" clearable filterable :placeholder="$t('m_group_name')">
                    <Option v-for="item in panel_group_list" :value="item" :key="item" style="float: left;">{{ item }}</Option>
                  </Select>
                  <Tooltip :content="$t('m_button_chart_dataView')" theme="light" transfer placement="top">
                    <i class="fa fa-eye" style="font-size: 16px;" v-if="isShowGridPlus(item)" aria-hidden="true" @click="gridPlus(item)"></i>
                  </Tooltip>
                </div>
              </div>
              <section>
                <div v-for="(chartInfo,chartIndex) in item._activeCharts" :key="chartIndex">
                  <CustomChart v-if="['line','bar'].includes(chartInfo.chartType)" :panel_group_list="panel_group_list" :chartInfo="chartInfo" :chartIndex="index" :params="viewCondition"></CustomChart>
                  <CustomPieChart v-if="chartInfo.chartType === 'pie'" :panel_group_list="panel_group_list" :chartInfo="chartInfo" :chartIndex="index" :params="viewCondition"></CustomPieChart>
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
    <Drawer :title="$t('m_view_details')" :width="zoneWidth" v-model="showMaxChart">
      <ViewChart ref="viewChart"></ViewChart>
    </Drawer>
    <Drawer :title="$t('m_placeholder_chartConfiguration')" :width="zoneWidth" :mask-closable="false" v-model="showChartConfig">
      <!-- <editPieView v-if="chartType === 'pie' && showChartConfig" ref="editPieView" :activeGridConfig="activeGridConfig" :parentRouteData="parentRouteData"></editPieView>
      <editLineView v-if="chartType !== 'pie' && showChartConfig" ref="editLineView" :activeGridConfig="activeGridConfig" :parentRouteData="parentRouteData"></editLineView> -->
    </Drawer>
    <ModalComponent :modelConfig="setChartTypeModel">
      <div slot="setChartType">
        <div style="display:flex;justify-content:center">
          <i @dblclick="dblChartType('bar')" @click="choiceChartType('bar')" :class="['fa', 'fa-line-chart', activeChartType === 'bar' ? 'active-tag' : '']" aria-hidden="true"></i>
          <i @dblclick="dblChartType('pie')" @click="choiceChartType('pie')" :class="['fa', 'fa-pie-chart', activeChartType === 'pie' ? 'active-tag' : '']" aria-hidden="true"></i>
        </div>
      </div>
    </ModalComponent>
    <Modal
      v-model="isShowWarning"
      :title="$t('m_delConfirm_title')"
      @on-ok="confirmRemoveGrid"
      @on-cancel="cancel"
    >
      <div class="modal-body" style="padding:30px">
        <div style="text-align:center">
          <p style="color: red">{{$t('m_delConfirm_tip')}}</p>
        </div>
      </div>
    </Modal>
    <Modal v-model="showGroupMgmt" :title="$t('m_group_mgmt')" :mask-closable="false">
      <div style="margin: 40px 0 60px 0">
        <Form :label-width="100">
          <FormItem :label="$t('m_group_name')">
            <Input v-model="groupName" placeholder="" style="width: 300px" />
          </FormItem>
        </Form>
      </div>
      <Divider style="margin-top:40px">{{ $t('m_add_chart_to_group') }}</Divider>
      <div>
        <Row>
          <Col span="12" v-for="panel in panelGroupInfo" :key="panel.name">
          <Checkbox v-model="panel.setGroup" :disabled="panel.hasGroup">{{ panel.label }}</Checkbox>
          </Col>
        </Row>
      </div>
      <template slot="footer">
        <Button @click="showGroupMgmt = false">{{ $t('m_button_cancel') }}</Button>
        <Button @click="confirmGroupMgmt" :disabled="!groupName" type="primary" class="primary-btn">{{ $t('m_button_save') }}</Button>
      </template>
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
export default {
  name: '',
  data() {
    return {
      parentRouteData: {},
      editData: null,
      isShowWarning: false,
      deleteConfirm: {
        id: '',
        method: ''
      },
      isEditPanal: false,
      permission: 'edit',
      viewId: null,
      panalName: '',
      viewCondition: {
        timeTnterval: -3600,
        dateRange: ['', ''],
        autoRefresh: 10,
        agg: 'none' // 聚合类型
      },
      disableTime: false,
      dataPick,
      autoRefreshConfig,
      viewData: [],
      activeGroup: 'All',
      showGroupMgmt: false,
      panelGroupInfo: [], // 存放新增/编辑组时的panel信息
      groupName: '', // 新增及编辑时的组名称
      groupNameIndex: -1, // 编辑时的组序号
      panel_group_list: [], // 存放视图拥有的组信息
      layoutData: [
        //   {'x':0,'y':0,'w':2,'h':2,'i':'0'},
        //   {'x':1,'y':1,'w':2,'h':2,'i':'1'},
      ],
      editChartId: null,
      setChartTypeModel: {
        modalId: 'set_chart_type_Modal',
        modalTitle: 'm_button_add',
        isAdd: true,
        config: [
          {
            name: 'setChartType',
            type: 'slot'
          }
        ],
        addRow: {
          type: null
        },
        modalFooter: [
          {
            name: '确定',
            Func: 'confirmChartType'
          }
        ]
      },
      activeGridConfig: null,
      activeChartType: 'bar',
      showAlarm: false, // 显示告警信息
      cutsomViewId: null,
      showMaxChart: false,
      zoneWidth: '800',
      showChartConfig: false,
      chartType: '',
      request: this.$root.$httpRequestEntrance.httpRequestEntrance,
      apiCenter: this.$root.apiCenter,
    }
  },
  props: ['id'],
  created() {
    this.zoneWidth = window.screen.width * 0.65
  },
  computed: {
    tmpLayoutData() { // 缓存切换分组后数据
      if (this.activeGroup === 'All') {
        return this.layoutData
      }
      return this.layoutData.filter(d => d.group === this.activeGroup)

    }
  },
  mounted() {
  },
  methods: {
    getDashData(viewId) {
      this.request('GET',this.apiCenter.template.singleDash, {id: viewId}, responseData => {
        this.viewId = responseData.id
        this.panalName = responseData.name
        this.editData = responseData
        this.viewData = JSON.parse(responseData.cfg)
        this.activeGroup = 'All'
        this.panel_group_list = responseData.panel_group_list || []
        this.initPanals()
        this.cutsomViewId = responseData.id
      })
    },
    openAlarmDisplay() {
      this.showAlarm = !this.showAlarm
      this.$refs.cutsomViewId.getAlarm(this.cutsomViewId, this.viewCondition, this.permission)
    },
    closeAlarmDisplay() {
      this.showAlarm = !this.showAlarm
      this.$refs.cutsomViewId.clearAlarmInterval()
    },
    datePick(data) {
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
    dateToTimestamp(date) {
      if (!date) {
        return 0
      }
      let timestamp = Date.parse(new Date(date))
      timestamp = timestamp / 1000
      return timestamp
    },
    initPanals() {
      const tmp = []
      this.viewData.forEach(item => {
        const params = {
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
        item.query.forEach(_ => {
          params.data.push(_)
        })
        const height = (item.viewConfig.h+1) * 30-8
        const _activeCharts = []
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
    isShowGridPlus(item) {
      // 新增及饼图时屏蔽放大功能
      if (!item._activeCharts || item._activeCharts[0].chartType === 'pie') {
        return false
      }
      return true
    },
    addItem() {
      this.activeGroup = 'All'
      generateUuid().then(elId => {
        const key = ((new Date()).valueOf())
          .toString()
          .substring(10)
        const item = {
          x: 0,
          y: 0,
          w: 6,
          h: 7,
          i: `default${key}`,
          id: `id_${elId}`
        }
        this.layoutData.push(item)
      })
    },
    setChartType(item) {
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
    choiceChartType(activeChartType) {
      this.activeChartType = activeChartType
      this.chartType = activeChartType
    },
    dblChartType(activeChartType) {
      this.activeChartType = activeChartType
      this.chartType = activeChartType
      this.confirmChartType()
    },
    confirmChartType() {
      if (!this.activeChartType) {
        this.$Message.warning('请先设置图标类型！')
        return
      }
      this.$root.JQ('#set_chart_type_Modal').modal('hide')
      this.editGrid()
    },
    editGrid(item) {
      this.modifyLayoutData().then(resViewData => {
        const parentRouteData = this.editData
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
          const findEditData = parentRouteData.cfg.find(xItem => xItem.viewConfig.id === item.id)
          findEditData.aggregate = aggregate
          findEditData.agg_step = agg_step
        } else {
          parentRouteData.cfg = resViewData
        }
        parentRouteData.cfg = JSON.stringify(parentRouteData.cfg)
        this.parentRouteData = parentRouteData
        if (['line','bar'].includes(this.activeChartType)) {
          this.chartType = 'line'
          // this.$refs.editLineView.initChart({templateData: parentRouteData, panal:this.activeGridConfig})
        } else {
          this.chartType = 'pie'
          // this.$refs.editPieView.initChart({templateData: parentRouteData, panal:this.activeGridConfig})
        }
        this.showChartConfig = true
      })
    },
    removeGrid(itemxxx) {
      this.isShowWarning = true
      this.deleteConfirm.id = itemxxx.id
    },
    confirmRemoveGrid() {
      this.layoutData.forEach((item,index) => {
        if (item.id === this.deleteConfirm.id) {
          this.layoutData.splice(index,1)
        }
      })
    },
    cancel() {
      this.isShowWarning = false
    },
    async gridPlus(item) {
      const resViewData = await this.modifyLayoutData()
      const parentRouteData = this.editData
      parentRouteData.cfg = JSON.stringify(resViewData)
      this.showMaxChart = true
      this.$refs.viewChart.initChart({
        templateData: parentRouteData,
        panal: item,
        parentData: this.editData
      })
    },
    async modifyLayoutData() {
      const resViewData = []
      this.layoutData.forEach(layoutDataItem => {
        const temp = {
          panalTitle: layoutDataItem.i,
          panalUnit: '',
          chartType: this.activeChartType,
          query: [],
          viewConfig: layoutDataItem
        }

        this.viewData.forEach(i => {
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
    resizeEvent(i, newH, newW, newHPx, newWPx){
      resizeEvent(this, i, newH, newW, newHPx, newWPx)
    },
    saveEdit() {
      const res = []
      this.layoutData.forEach(layoutDataItem => {
        this.viewData.forEach(i => {
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
      const params = {
        // name: this.panalName,
        // id: this.editData.id,
        name: this.panalName,
        panel_group_list: this.panel_group_list || [],
        id: this.viewId,
        cfg: JSON.stringify(res)
      }
      this.request('POST',this.apiCenter.template.save, params, () => {
        this.isEditPanal = false
        this.$Message.success(this.$t('m_tips_success'))
      })
    },
    goBack() {
      this.$router.push({name: 'viewConfigIndex'})
    },
    savePanalEdit() {
      if (!this.panalName) {
        this.$Message.warning(this.$t('m_tips_required'))
        return
      }
      this.saveEdit()
    },
    canclePanalEdit() {
      this.isEditPanal = false
      this.panalName = this.editData.name
    },
    // #region 组管理
    selectGroup(item) {
      this.activeGroup = item
      this.refreshNow = true
      this.$nextTick(() => {
        this.refreshNow = false
      })
    },
    addGroup() {
      this.groupName = ''
      this.groupNameIndex = -1
      this.getPanelGroupInfo()
      this.showGroupMgmt = true
    },
    editGroup(item ,index) {
      this.oriGroupName = item
      this.groupName = item
      this.groupNameIndex = index
      this.getPanelGroupInfo()
      this.showGroupMgmt = true
    },
    getPanelGroupInfo() {
      this.panelGroupInfo = []
      this.layoutData.forEach((d, dIndex) => {
        this.panelGroupInfo.push({
          index: dIndex,
          label: d.i,
          group: d.group,
          hasGroup: !!d.group,
          setGroup: false
        })
      })
    },
    removeGroup(item, index) {
      this.$delConfirm({
        msg: item,
        callback: () => {
          this.delF(item, index)
        }
      })
    },
    delF(item, index) {
      this.panel_group_list.splice(index, 1)
      this.layoutData.forEach(d => {
        if (d.group === item) {
          d.group = ''
        }
      })
      this.savePanalEdit()
      this.activeGroup = 'All'
    },
    confirmGroupMgmt() {
      if (this.panel_group_list.includes(this.groupName)) {
        this.$Message.warning(this.$t('m_group_name_exist'))
        return
      }
      if (this.groupNameIndex === -1) {
        this.panel_group_list.push(this.groupName)
        this.panelGroupInfo.forEach(p => {
          if (p.setGroup) {
            this.layoutData[p.index].group = this.groupName
          }
        })
      } else {
        this.panel_group_list[this.groupNameIndex] = this.groupName
        this.layoutData.forEach(d => {
          if (d.group === this.oriGroupName) {
            d.group = this.groupName
          }
        })
        this.panelGroupInfo.forEach(p => {
          if (p.setGroup) {
            this.layoutData[p.index].group = this.groupName
          }
        })
      }
      this.showGroupMgmt = false
      this.savePanalEdit()
    },
    // #endregion
  },
  components: {
    GridLayout: VueGridLayout.GridLayout,
    GridItem: VueGridLayout.GridItem,
    CustomChart,
    CustomPieChart,
    ViewConfigAlarm,
    ViewChart,
    // editPieView,
    // editLineView
  },
}
</script>

<style scoped lang="less">
.panal-edit-icon {
  margin-left:4px;
  padding: 4px;
  cursor: pointer;
}
.header-name {
  font-size: 16px;
  margin: 0 8px;
  font-weight: 600;
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
.primary-btn {
  color: #fff;
  background-color: #57a3f3;
  border-color: #57a3f3;
}
</style>
