<template>
  <div>
    <div>
      <div class="header-title">
        <Icon size="22" type="md-arrow-back" class="icon" @click="returnPreviousPage" />
        <h4>{{$t('menu.screenConfiguration')}}</h4>
      </div>
      <header>
        <div class="header-name">
          <i class="fa fa-th-large fa-18 mr-2" aria-hidden="true"></i>
          <template v-if="isEditPanal">
            <Input v-model.trim="panalName" style="width: 100px" type="text"></Input>
            <Icon class="panal-edit-icon" @click="savePanalEdit" type="md-checkmark" />
            <Icon class="panal-edit-icon" @click="canclePanalEdit" type="md-trash" />
          </template>
          <template v-else>
            <h5 class="d-inline-block"> {{panalName}}</h5>
            <Icon class="panal-edit-icon" @click="isEditPanal=true" v-if="isEditStatus" type="md-create" />
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
            <div class="search-zone ml-2">
              <span class="params-title">{{$t('placeholder.refresh')}}：</span>
              <Select filterable clearable v-model="viewCondition.autoRefresh" :disabled="disableTime" style="width:100px" @on-change="initPanals" :placeholder="$t('placeholder.refresh')">
                <Option v-for="item in autoRefreshConfig" :value="item.value" :key="item.value">{{ item.label }}</Option>
              </Select>
            </div>
            <div class="search-zone ml-2">
              <span class="params-title">{{$t('field.timeInterval')}}：</span>
              <DatePicker 
                type="datetimerange" 
                :value="viewCondition.dateRange" 
                format="yyyy-MM-dd HH:mm:ss" 
                placement="bottom-start" 
                split-panels
                @on-change="datePick" 
                :placeholder="$t('placeholder.datePicker')" 
                style="width: 320px">
              </DatePicker>
            </div>
          </div>

          <div class="header-tools">
            <template v-if="isEditStatus">
              <!-- <button class="btn btn-sm btn-cancel-f" style="margin-right:60px" @click="addItem">{{$t('m_new_graph')}}</button> -->
              <button class="btn btn-sm btn-confirm-f" @click="savePanelInfo">{{$t('button.saveConfig')}}</button>
            </template> 
            
            <button v-if="!showAlarm" class="btn btn-sm btn-cancel-f" @click="openAlarmDisplay()">
              <i style="font-size: 18px;color: #0080FF;" class="fa fa-eye-slash" aria-hidden="true"></i>
            </button>
            <button v-else class="btn btn-sm btn-cancel-f" @click="closeAlarmDisplay()">
              <i style="font-size: 18px;color: #0080FF;" class="fa fa-eye" aria-hidden="true"></i>
            </button>
          </div>
        </div>
      </header>

      <!-- 分组 -->
      <div>
        <div class="radio-group">
          <span class="ml-3 mr-3">{{$t('m_group_name')}}:</span>
          <div
            :class="['radio-group-radio radio-group-optional', activeGroup === 'All' ? 'selected-radio' : 'is-not-selected-radio']"
          >
            <span @click="selectGroup('All')" style="vertical-align: text-bottom;">{{$t('m_chart_all')}}</span>
          </div>
          <div
            v-for="(item, index) in panel_group_list"
            :key="index"
            :class="['radio-group-radio radio-group-optional', item === activeGroup ? 'selected-radio' : 'is-not-selected-radio']"
          >
            <Icon v-if="isEditStatus" class="mr-2" @click="editGroup(item, index)" type="md-create" color="#2d8cf0" :size="15" />
            <span @click="selectGroup(item)" style="vertical-align: text-bottom;">
              {{ `${item}` }}
            </span>
            <Icon v-if="isEditStatus" class="ml-2" @click="removeGroup(item, index)" type="md-close" color="#ed4014" :size="15" />
          </div>
          <span>
            <Button
              v-if="isEditStatus"
              @click="addGroupItem"
              class="primary-btn"
              style="margin-top: -5px;"
              type="success"
              shape="circle"
              icon="md-add"
            ></Button>
          </span>
        </div>

        <!-- 图表新增 -->
        <div class="chart-config-info">
          <span class="fs-20 mr-3 ml-3">{{$t('m_graph')}}:</span>
          <Dropdown 
            v-for="(item, index) in allAddChartOptions"
            :key="index"
            class="mr-3 chart-option-menu" 
            @on-click="(info) => onAddChart(JSON.parse(info), item.type)">
            <a href="javascript:void(0)">
                {{$t(item.name)}}
                <Icon type="ios-arrow-down"></Icon>
            </a>
            <template #list>
                <DropdownMenu>
                    <DropdownItem v-for="(option, key) in item.options"
                      :name="JSON.stringify(option)" 
                      :key="key"
                      :disabled="option.disabled">
                      <Icon v-if="option.iconType" :type="option.iconType" />
                      {{item.type === 'add' ? $t(option.name) : option.name}}
                    </DropdownItem>
                </DropdownMenu>
            </template>
          </Dropdown>
        </div>
      </div>

      <!-- 图表展示区域 -->
      <div style="display:flex">
        <div class="grid-style">
          <grid-layout 
          :layout.sync="tmpLayoutData"
          :col-num="12"
          :row-height="30"
          :is-draggable="true"
          :is-resizable="true"
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
              @resized="resizeEvent">
              <template v-if="item.group === activeGroup || activeGroup === 'All'">
                <div class="c-dark grid-content">
                  <div class="header-grid header-grid-name">
                    <span v-if="editChartId !== item.id">{{item.i}}</span>
                    <Input v-else v-model="item.i" class="editChartId" autofocus style="width:100px" size="small" placeholder="" />
                    <Tooltip :content="$t('placeholder.editTitle')" theme="light" transfer placement="top">
                      <i v-if="isEditStatus && editChartId !== item.id && !noAllowChartChange(item)" class="fa fa-pencil-square" style="font-size: 16px;" @click="editChartId = item.id" aria-hidden="true"></i>
                      <Icon v-if="editChartId === item.id" size="16" type="md-checkmark" @click="onChartTitleChange(item)" />
                    </Tooltip>
                  </div>
                  <div class="header-grid header-grid-tools">
                    <Button v-if="item.public" size="small" class="mr-1 references-button">{{$t('m_shallow_copy')}}</Button>
                    <Select v-model="item.group" 
                      style="width:100px;" 
                      size="small" 
                      :disabled="permission !== 'edit'" 
                      clearable 
                      filterable 
                      :placeholder="$t('m_group_name')"
                      @on-change="onSingleChartGroupChange">
                      <Option v-for="item in panel_group_list" :value="item" :key="item" style="float: left;">{{ item }}</Option>
                    </Select>
                    <Tooltip :content="$t('m_save_chart_library')" theme="light" transfer placement="top">
                      <Icon v-if="!item.public" size="15" type="md-archive" @click="showChartAuthDialog(item)" />
                    </Tooltip>
                    <Tooltip :content="$t('button.chart.dataView')" theme="light" transfer placement="top">
                      <i class="fa fa-eye" style="font-size: 16px;" v-if="isShowGridPlus(item)" aria-hidden="true" @click="gridPlus(item)"></i>
                    </Tooltip>
                    <Tooltip :content="$t('placeholder.chartConfiguration')" theme="light" transfer placement="top">
                      <i class="fa fa-cog" style="font-size: 16px;" v-if="isEditStatus && !noAllowChartChange(item)" @click="setChartType(item)" aria-hidden="true"></i>
                    </Tooltip>
                    <Tooltip :content="$t('placeholder.deleteChart')" theme="light" transfer placement="top">
                      <i class="fa fa-trash" style="font-size: 16px;color:red" v-if="isEditStatus" @click="removeGrid(item)" aria-hidden="true"></i>
                    </Tooltip>
                  </div>
                </div>
                <section style="height: 90%" @click="setChartType(item)">
                  <div v-for="(chartInfo,chartIndex) in item._activeCharts" :key="chartIndex">
                    <CustomChart v-if="['line','bar'].includes(chartInfo.chartType)" :refreshNow="refreshNow" :chartInfo="chartInfo" :chartIndex="index" :params="viewCondition"></CustomChart>
                    <CustomPieChart v-if="chartInfo.chartType === 'pie'" :refreshNow="refreshNow" :chartInfo="chartInfo" :chartIndex="index" :params="viewCondition"></CustomPieChart>
                  </div>
                </section>
              </template>       
            </grid-item>
          </grid-layout>
        </div>
        <div class="view-config-alarm" v-if="showAlarm">
          <ViewConfigAlarm ref="cutsomViewId"></ViewConfigAlarm>
        </div>
      </div>
    </div>
    <Drawer title="View details" :width="zoneWidth" v-model="showMaxChart">
      <ViewChart ref="viewChart"></ViewChart>
    </Drawer>

    <!-- 对于每个chart的抽屉详细信息 -->
    <Drawer :title="$t('placeholder.chartConfiguration')" :width="90" :mask-closable="false" v-model="showChartConfig">
      <editView :chartId="setChartConfigId" v-if="showChartConfig"></editView>
    </Drawer>
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

    <!-- 分组新增 -->
    <Modal v-model="showGroupMgmt" :title="$t('m_edit_screen_group')" :mask-closable="false">
      <div>
        <Form :label-width="90">
          <FormItem :label="$t('m_group_chart_name')">
            <Input v-model="groupName" placeholder="" style="width: 300px" />
          </FormItem>
        </Form>
      </div>
      <div style="display: flex" class="ml-4">
        <div style="width: 100px">{{$t('m_use_charts')}}</div>
        <Row>
          <Col span="12" v-for="panel in panelGroupInfo" :key="panel.name">
            <Checkbox v-model="panel.setGroup" :disabled="panel.hasGroup"
              >{{ panel.label }}</Checkbox
            >
          </Col>
        </Row>
      </div>
      <template #footer>
        <Button @click="showGroupMgmt = false">{{ $t('button.cancel') }}</Button>
        <Button @click="confirmGroupMgmt" :disabled="!groupName" type="primary">{{ $t('button.save') }}</Button>
      </template>
    </Modal>
    <!-- 删除组 -->
    <Modal
      v-model="isShowDeleteGroupWarning"
      :title="$t('delConfirm.title')"
      @on-ok="confirmDeleteGroup"
      @on-cancel="isShowDeleteGroupWarning = false">
      <div class="modal-body" style="padding:30px">
        <div style="text-align:center">
          <p style="color: red">{{$t('m_delete_tip')}}: {{ deleteGroup }}</p>
        </div>
      </div>
    </Modal>
    <AuthDialog ref="authDialog" :useRolesRequired="true" @sendAuth="saveChartOrDashboardAuth" />
  </div>
</template>

<script>
import isEmpty from 'lodash/isEmpty';
import remove from 'lodash/remove';
import cloneDeep from 'lodash/cloneDeep';
import {generateUuid} from '@/assets/js/utils'
import {dataPick, autoRefreshConfig} from '@/assets/config/common-config'
import {resizeEvent} from '@/assets/js/gridUtils.ts'
import VueGridLayout from 'vue-grid-layout'
import CustomChart from '@/components/custom-chart'
import CustomPieChart from '@/components/custom-pie-chart'
import ViewConfigAlarm from '@/views/custom-view/view-config-alarm'
import ViewChart from '@/views/custom-view/view-chart'
import EditView from '@/views/custom-view/edit-view/edit-view'
import AuthDialog from '@/components/auth.vue';
import { nextTick } from 'vue';
export default {
  name: '',
  data() {
    return {
      isShowDeleteGroupWarning: false,
      deleteGroup: null,
      deleteGroupIndex: -1,
      refreshNow: false,
      parentRouteData: {},
      editData: null,
      isShowWarning: false,
      deleteConfirm: {
        id: '',
        method: ''
      },
      isEditPanal: false,
      permission: this.$route.params.permission || 'edit',
      panalName: '',
      pannelId: this.$route.params.pannelId || 29,
      viewCondition: {
        timeTnterval: -3600,
        dateRange: ['', ''],
        autoRefresh: 10,
        agg: 'none' // 聚合类型
      },
      disableTime: false,
      dataPick: dataPick,
      autoRefreshConfig: autoRefreshConfig,
      viewData: [],
      activeGroup: 'All',
      showGroupMgmt: false,
      panelGroupInfo: [], // 存放新增/编辑组时的panel信息
      groupName: '', // 新增及编辑时的组名称
      groupNameIndex: -1, // 编辑时的组序号
      panel_group_list: [], // 存放视图拥有的组信息
      layoutData: [],
      editChartId: null,
      activeGridConfig: null,
      activeChartType: 'bar',
      showAlarm: false, // 显示告警信息
      cutsomViewId: null,
      showMaxChart: false,
      zoneWidth: '800',
      showChartConfig: false,
      chartType: '',
      allAddChartOptions: [
        {
          name: 'button.add',
          type: 'add',
          options: [
            {
              name: "m_line_chart_s",
              id: "line",
              iconType: "md-trending-up"
            },
            {
              name: "m_pie_chart",
              id: "pie",
              iconType: "md-pie"
            }
          ],
        },
        {
          name: 'm_copy',
          type: 'copy',
          options: []
        },
        {
          name: 'm_shallow_copy',
          type: 'shallowCopy',
          options: []
        }
      ],
      request: this.$root.$httpRequestEntrance.httpRequestEntrance,
      lineTypeOption: {
        line: 1,
        area: 0
      },
      setChartConfigId: '',
      chartAuthDialogShow: false,
      mgmtRoles: [],
      useRoles: [],
      mgmtRolesOptions: [], 
      userRolesOptions: [],
      boardMgmtRoles: [],
      boardUseRoles: []
    }
  },
  computed: {
    tmpLayoutData() { // 缓存切换分组后数据
      if (this.activeGroup === 'All') {
        return this.layoutData
      } else {
        return this.layoutData.filter(d => d.group === this.activeGroup)
      }
    },
    isEditStatus() {
      return this.permission === 'edit'
    },
  },
  created () {
    this.zoneWidth = window.screen.width * 0.65;
    this.getAllChartOptionList();
    this.getPannelList();
    this.activeGroup = 'All';
    this.getAllRolesOptions();
  },
  methods: {
    getPannelList() {
      this.request('GET', '/monitor/api/v2/dashboard/custom', {
        id: this.pannelId
      }, res => {
        if(isEmpty(res)) {
          this.$router.push({path:'viewConfigIndex'})
        } else {
          this.boardMgmtRoles = res.mgmtRoles;
          this.boardUseRoles = res.useRoles;
          this.panalName = res.name;
          this.activeGroup = 'All';
          this.panel_group_list = res.panelGroupList || [];
          this.viewData = res.charts || [];
          this.initPanals();
          this.cutsomViewId = this.pannelId;
        }
      })
    },
    getAllChartOptionList() {
      this.request('GET', '/monitor/api/v2/chart/shared/list', {}, res => {
        const copyChartsOptions = this.processChartOptions(res);
        this.allAddChartOptions[1].options = copyChartsOptions;
        this.allAddChartOptions[2].options = copyChartsOptions;
      })
    },
    processChartOptions(rawData) {
      const options = [];
      const initialOption = {
        line: {
          name: "折线图",
          id: "line",
          iconType: "md-trending-up",
          disabled: true
        },
        pie: {
          name: "饼图",
          id: "pie",
          iconType: "md-pie",
          disabled: true
        },
        bar: {
          name: "柱状图",
          id: "bar",
          iconType: "ios-stats",
          disabled: true
        }
      }

      for(let key in rawData) {
        options.push(initialOption[key]);
        if (!isEmpty(rawData[key])) {
          rawData[key].forEach(item => {
            options.push(Object.assign({disabled: false}, item))
          })
        }
      }
      return options
    },
    openAlarmDisplay () {
      this.showAlarm = !this.showAlarm
      nextTick(() => {
        this.$refs.cutsomViewId.getAlarm(this.cutsomViewId, this.viewCondition, this.permission);
        this.refreshNow = !this.refreshNow;
      })
    },
    closeAlarmDisplay () {
      this.showAlarm = !this.showAlarm;
      this.$refs.cutsomViewId.clearAlarmInterval();
      this.refreshNow = !this.refreshNow;
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
      let tmpArr = []
      this.viewData.forEach((item) => {
        const parsedDisplayConfig = JSON.parse(item.displayConfig);
        let params = {
          aggregate: item.aggregate,
          agg_step: item.aggStep,
          lineType: this.lineTypeOption[item.lineType],
          time_second: this.viewCondition.timeTnterval,
          start: this.dateToTimestamp(this.viewCondition.dateRange[0]),
          end: this.dateToTimestamp(this.viewCondition.dateRange[1]),
          title: '',
          unit: '',
          data: []
        }
        item.chartSeries.forEach( _ => {
          params.data.push(_)
        })

        let height = (parsedDisplayConfig.h + 1) * 30-8
        let _activeCharts = []
        _activeCharts.push({
          style: `height:${height}px;`,
          panalUnit: item.unit,
          elId: item.id,
          chartParams: params,
          chartType: item.chartType,
          aggregate: item.aggregate,
          agg_step: item.aggStep,
          lineType: this.lineTypeOption[item.lineType],
          time_second: this.viewCondition.timeTnterval,
          start: this.dateToTimestamp(this.viewCondition.dateRange[0]),
          end: this.dateToTimestamp(this.viewCondition.dateRange[1])                                  
        })
        tmpArr.push({
          _activeCharts,
          ...parsedDisplayConfig,
          i: item.name,
          id: item.id,
          group: item.group,
          public: item.public,
          sourceDashboard: item.sourceDashboard
        })
      })
      this.layoutData = tmpArr
    },
    isShowGridPlus (item) {
      if (!item._activeCharts || item._activeCharts[0].chartType === 'pie') {
        return false
      }
      return true
    },
    addItem() {
      this.activeGroup = 'All'
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
      this.setChartConfigId = item.id;
      const find = this.layoutData.find(i => i.id === item.id)
      this.activeGridConfig = find
      if (!find._activeCharts) {
        // this.$root.JQ('#set_chart_type_Modal').modal('show')
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
      this.editGrid()
    },
    editGrid(item) {
      this.modifyLayoutData().then((resViewData)=>{
        let parentRouteData = this.editData
        this.parentRouteData = parentRouteData
        if (['line','bar'].includes(this.activeChartType)) {
          this.chartType = 'line'
        } else {
          this.chartType = 'pie'
        }
        this.showChartConfig = true
      })
    },
    removeGrid(item) {
      this.isShowWarning = true
      this.deleteConfirm.id = item.id
    },
    async confirmRemoveGrid () {
      const params = this.processPannelParams()
      remove(params.charts, {id: this.deleteConfirm.id});
      await this.requestReturnPromise('PUT', '/monitor/api/v2/dashboard/custom', params);
      this.getPannelList();
    },
    cancel () {
      this.isShowWarning = false
    },
    async gridPlus(item) {
      const resViewData = await this.modifyLayoutData()
      this.showMaxChart = true
      const templateData = {
        cfg: JSON.stringify(resViewData)
      }
      this.$refs.viewChart.initChart({templateData, panal:item})
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
          if (layoutDataItem.id === i.id) {
            temp.panalUnit = i.unit
            temp.query = i.chartSeries
            temp.chartType = i.chartType
            temp.aggregate = i.aggregate
            temp.agg_step = i.aggStep
            temp.lineType = this.lineTypeOption[i.lineType]
          }
        })
        resViewData.push(temp)
      })
      return resViewData
    },
    resizeEvent: function(i, newH, newW, newHPx, newWPx){
      resizeEvent(this, i, newH, newW, newHPx, newWPx)
    },
    submitPanelInfo() {
      return new Promise(resolve => {
          if (isEmpty(this.boardMgmtRoles) || isEmpty(this.boardUseRoles)) {
          this.saveAuthType = 'board';
          this.$refs.authDialog.startAuth(this.boardMgmtRoles, this.boardUseRoles, this.mgmtRolesOptions, this.userRolesOptions);
        } else {
          this.request('PUT', '/monitor/api/v2/dashboard/custom', this.processPannelParams(), res => {
            resolve()
          });
        }
      })
    },

    async savePanelInfo() {
      await this.submitPanelInfo();
      this.$Message.success(this.$t('tips.success'));
    },

    returnPreviousPage() {
      this.$router.push({name:'viewConfigIndex'})
    },
    async savePanalEdit () {
      if (!this.panalName) {
        this.$Message.warning(this.$t('tips.required'))
        return
      }
      await this.submitPanelInfo();
      this.$Message.success(this.$t('tips.success'));
      this.isEditPanal = false;
    },
    canclePanalEdit () {
      this.isEditPanal = false
      this.panalName = this.editData.name
    },
    //#region 组管理
    selectGroup (item) {
      this.activeGroup = item
      this.refreshNow = true
      this.$nextTick(() => {
        this.refreshNow = false
      })
    },
    addGroupItem () {
      this.groupName = ''
      this.groupNameIndex = -1
      this.getPanelGroupInfo('')
      this.showGroupMgmt = true
    },
    editGroup (item ,index) {
      this.oriGroupName = item
      this.groupName = item
      this.groupNameIndex = index
      this.getPanelGroupInfo(item)
      this.showGroupMgmt = true
    },
    getPanelGroupInfo (groupName) {
      this.panelGroupInfo = []
      this.layoutData.forEach((d, dIndex) => {
        let group = {
          index: dIndex,
          label: d.i,
          group: d.group || '',
          hasGroup: !!d.group,
          setGroup: false
        }
        // 未绑定组
        if (!d.group) {
          group.hasGroup = false
          group.setGroup = false
        } else { // 已绑定
          group.hasGroup = d.group === groupName ? false : true
          group.setGroup = d.group === groupName ? true : false
        }
        this.panelGroupInfo.push(group)
      })
    },
    removeGroup (item, index) {
      this.isShowDeleteGroupWarning = true
      this.deleteGroup = item
      this.deleteGroupIndex = index
    },
    confirmDeleteGroup () {
      this.panel_group_list.splice(this.deleteGroupIndex, 1)
      this.layoutData.forEach(d => {
        if (d.group === this.deleteGroup) {
          d.group = ''
        }
      })
      this.savePanalEdit()
      this.activeGroup = 'All'
    },
    confirmGroupMgmt () {
      if (this.groupNameIndex === -1 && this.panel_group_list.includes(this.groupName)) {
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
          if (p.hasGroup === false) {
            if (p.setGroup && p.group === '') {
              this.layoutData[p.index].group = this.groupName
            } else if (!p.setGroup ) {
              this.layoutData[p.index].group = ''
            }
          }
        })
      }
      this.showGroupMgmt = false
      this.savePanalEdit()
      this.activeGroup = this.groupName
    },

    processInitialChartName() {
      let allName = [];
      if (!isEmpty(this.layoutData)) {
        allName = this.layoutData.map(item => {
          if (!isNaN(Number(item.i))) {
            return Number(item.i)
          } else {
            return 10000
          }
        })
        return Math.max(...allName) + 1 + ''
      }
      return '10000'
    },
    async onAddChart(copyInfo, type) {
      if (type === 'add') {
        const name = this.processInitialChartName();
        const addChartParams = {
          dashboardId: this.pannelId,
          name: name,
          chartTemplate: "",
          chartType: copyInfo.id,
          lineType: copyInfo.id === 'line' ? 'line' : "",
          pieType: copyInfo.id === 'pie' ? 'tag' : "",
          aggregate: "none",
          aggStep: 60,
          unit: '',
          group: this.activeGroup === 'ALL' ? "" : this.activeGroup,
          displayConfig: {
              x:0,
              y:0,
              w:6,
              h:7,
            }
        }
        this.setChartConfigId = await this.requestReturnPromise('POST', '/monitor/api/v2/chart/custom', addChartParams);
        let item = {
          x:0,
          y:0,
          w:6,
          h:7,
          i: `${name}`,
          id: `${this.setChartConfigId}`
        }
        this.layoutData.push(item);
        this.editGrid()

        nextTick(async () => {
          await this.requestReturnPromise('PUT', '/monitor/api/v2/dashboard/custom', this.processPannelParams());
          this.getPannelList();
          this.showChartConfig = true;
        })
      } else {
        const group = type === 'copy' ? '' : (this.activeGroup === 'ALL' ? "" : this.activeGroup);
        let item = {
          x:0,
          y:0,
          w:6,
          h:7,
          i: '',
          id: `${copyInfo.id}`, 
          group
        }
        this.layoutData.push(item);
        this.editGrid()

        nextTick(async () => {
          const copyParams = {
            dashboardId: this.pannelId,
            ref: type === 'copy' ? false : true,
            originChartId: copyInfo.id,
            group,
            displayConfig: {
              x:0,
              y:0,
              w:6,
              h:7,
            }
          }
          await this.requestReturnPromise('POST', '/monitor/api/v2/chart/custom/copy', copyParams);
          await this.requestReturnPromise('PUT', '/monitor/api/v2/dashboard/custom', this.processPannelParams());
          this.getPannelList();
          this.setChartConfigId = copyInfo.id;
          this.showChartConfig = true
        })
      }
    },

    processPannelParams() {
      const charts = [];
      this.layoutData.forEach(item =>{
        charts.push({
          id: item.id,
          group: item.group,
          name: item.i,
          displayConfig: {
            x: item.x,
            y: item.y,
            w: item.w,
            h: item.h
          }
        })
      })
      return {
        id: this.pannelId,
        name: this.panalName,
        charts
      }
    },

    requestReturnPromise(method, api, params) {
      return new Promise(resolve => {
        this.request(method, api, params, res => {
            resolve(res)
          })
      })
    },

    async onChartTitleChange(item) {
      const isDuplicateName = await this.chartDuplicateNameCheck(item.id, item.i);
      if (isDuplicateName) return
      await this.requestReturnPromise('PUT', '/monitor/api/v2/chart/custom/name', {
        chartId: item.id,
        name: item.i
      })
      this.editChartId = null;
      this.$Message.success(this.$t('tips.success'));
    },
    chartDuplicateNameCheck(chartId, chartName) {
      return new Promise(resolve => {
        this.request('GET', '/monitor/api/v2/chart/custom/name/exist', {
          chart_id: chartId,
          name: chartName
        }, res => {
          if (res) {
            this.$Message.error(this.$t('m_graph_name') + this.$t('m_cannot_be_repeated'))
          }
          resolve(res)
        })
      })
    },
    noAllowChartChange(item) {
      return item.public === true && item.sourceDashboard !== this.pannelId
    }, 
    async showChartAuthDialog(item) {
      this.setChartConfigId = item.id
      this.chartAuthDialogShow = true;
      await this.getSingleChartAuth();
      this.saveAuthType = 'chart';
      this.$refs.authDialog.startAuth(this.mgmtRoles, this.useRoles, this.mgmtRolesOptions, this.userRolesOptions);
    },
    getSingleChartAuth() {
      return new Promise(resolve => {
        this.request('GET','/monitor/api/v2/chart/custom/permission', {
          chart_id: this.setChartConfigId
        }, res => {
          this.mgmtRoles = res.mgmtRoles;
          this.useRoles = res.useRoles;
          resolve()
        })
      })
    },
    getAllRolesOptions () {
      const params = {
        page: 1,
        size: 1000
      }
      this.request('GET','/monitor/api/v1/user/role/list', params, res => {
        this.userRolesOptions = this.processRolesList(res.data)
      })
      this.request('GET', '/monitor/api/v1/user/manage_role/list', {}, res => {
        this.mgmtRolesOptions = this.processRolesList(res);
      })
    },
    processRolesList(list = []) {
      if (isEmpty(list)) return [];
      const resArr = cloneDeep(list).map(item => {
        return {
          ...item,
          key: item.name,
          label: item.displayName || item.display_name
        }
      })
      return resArr
    },
    saveChartOrDashboardAuth(mgmtRoles, useRoles) {
      let path = ''
      const params = {
        mgmtRoles,
        useRoles
      }
      if (this.saveAuthType === 'chart') {
        this.mgmtRoles = mgmtRoles;
        this.useRoles = useRoles;
        path = '/monitor/api/v2/chart/custom/permission';
        params.chartId = this.setChartConfigId
      } else {
        this.boardMgmtRoles = mgmtRoles;
        this.boardUseRoles = useRoles;
        path = '/monitor/api/v2/dashboard/custom/permission';
        params.id = this.pannelId
      }
      this.request('POST', path, params, () => {
        this.$Message.success(this.$t('m_success'));
        this.getPannelList()
      })
    },
    onSingleChartGroupChange() {
      this.request('PUT', '/monitor/api/v2/dashboard/custom', this.processPannelParams(), res => {
        this.$Message.success(this.$t('m_success'));
        this.getPannelList();
        this.activeGroup = 'All';
      });
    }
  },
  components: {
    GridLayout: VueGridLayout.GridLayout,
    GridItem: VueGridLayout.GridItem,
    CustomChart,
    CustomPieChart,
    ViewConfigAlarm,
    ViewChart,
    EditView,
    AuthDialog
  },
}
</script>

<style lang="less">
.chart-config-info {
  .ivu-dropdown-item-disabled {
    color: inherit
  }
}

.references-button {
  background-color: #edf4fe;
  border-color: #b5d0fb;
}

.chart-option-menu {
  .ivu-select-dropdown {
    max-height: 100vh !important;
  }
}

.grid-content {
  display: flex;
  padding: 0 32px;
  font-size: 16px;
}

.header-grid-tools {
  i {
    font-size: 18px !important
  }
}

</style>

<style scoped lang="less">
.header-title {
  display: flex;
  align-items: center;
  margin-left: 15px;
  .icon {
    cursor: pointer;
    width: 28px;
    height: 24px;
    color: #fff;
    border-radius: 2px;
    background: #2d8cf0;
    margin-right: 20px;
    margin-bottom: 10px;
  }
}

.header-name {
  font-size: 16px; 
  margin-left: 15px;
}

.panal-edit-icon {
  margin-left:4px;
  padding: 4px;
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
  font-size: 14px;
  margin-bottom: 15px;
}
.chart-config-info {
  font-size: 14px;
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

.selected-radio {
  background-color: rgba(129, 179, 55, 0.6)
}

.is-not-selected-radio {
  background: #fff
}
.primary-btn {
  color: #fff;
  background-color: #57a3f3;
  border-color: #57a3f3;
}

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

  .view-config-alarm {
    width: 700px
  }
</style>
