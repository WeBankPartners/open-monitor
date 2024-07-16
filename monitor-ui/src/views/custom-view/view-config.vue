<template>
  <div class="all-content">
    <div>
      <header>
        <div class="header-name">
          <div v-if="pageType !== 'dashboard'">
            <Icon v-if="pageType !== 'link'" size="22" class="arrow-back" type="md-arrow-back" @click="returnPreviousPage" ></Icon>
            <template v-if="isEditPanal">
              <Input v-model.trim="panalName" style="width: 300px" type="text" :maxlength="30" show-word-limit/>
              <Icon class="panal-edit-icon" color="#2d8cf0" @click="savePanalEdit" type="md-checkmark" ></Icon>
              <Icon class="panal-edit-icon" color="red" @click="canclePanalEdit" type="md-trash" ></Icon>
            </template>
            <template v-else>
              <h5 class="d-inline-block"> {{panalName}}</h5>
              <Icon class="panal-edit-icon" color="#2d8cf0"  @click="isEditPanal = true" v-if="isEditStatus" type="md-create" ></Icon>
            </template>
          </div>
        </div>
        <div class="search-container">
          <div>
            <div class="search-zone">
              <span class="params-title">{{$t('m_field_relativeTime')}}：</span>
              <!-- <Select filterable v-model="viewCondition.timeTnterval" :disabled="disableTime" style="width:80px"  @on-change="initPanals">
                <Option v-for="item in dataPick" :value="item.value" :key="item.value">{{ item.label }}</Option>
              </Select> -->
              <RadioGroup @on-change="initPanals" v-model="viewCondition.timeTnterval" type="button" size="small">
                <Radio v-for="(item, idx) in dataPick" :label="item.value" :key="idx" :disabled="disableTime">{{ item.label }}</Radio>
              </RadioGroup>
            </div>
            <div class="search-zone ml-2">
              <span class="params-title">{{$t('m_field_timeInterval')}}：</span>
              <DatePicker
                type="datetimerange"
                :value="viewCondition.dateRange"
                format="yyyy-MM-dd HH:mm:ss"
                placement="bottom-start"
                split-panels
                @on-change="datePick"
                :placeholder="$t('m_placeholder_datePicker')"
                style="width: 320px"
              ></DatePicker>
            </div>
            <div class="search-zone ml-2">
              <span class="params-title">{{$t('m_placeholder_refresh')}}：</span>
              <Select filterable clearable v-model="viewCondition.autoRefresh" :disabled="disableTime" style="width:100px" @on-change="initPanals" :placeholder="$t('m_placeholder_refresh')">
                <Option v-for="item in autoRefreshConfig" :value="item.value" :key="item.value">{{ item.label }}</Option>
              </Select>
            </div>
          </div>

          <div class="header-tools">
            <template v-if="isEditStatus">
              <Button type="info" @click.stop="exportPanel">
                <Icon type="md-cloud-upload" size="20"></Icon>
                {{$t('m_export')}}
              </Button>
              <Button type="primary" @click="savePanelInfo">{{$t('m_save')}}</Button>
            </template>

            <Button type="warning" @click="showAlarm ? closeAlarmDisplay() : openAlarmDisplay()">
              {{$t('m_alert')}}
            </Button>
          </div>
        </div>
      </header>

      <!-- 分组 -->
      <div>
        <div class="radio-group">
          <span class="ml-3 mr-3">{{$t('m_group_name')}}:</span>
          <div
            :class="['radio-group-radio radio-group-optional', activeGroup === 'ALL' ? 'selected-radio' : 'is-not-selected-radio']"
          >
            <span @click="selectGroup('ALL')">{{$t('m_chart_all')}}</span>
          </div>
          <div
            v-for="(item, index) in panel_group_list"
            :key="index"
            :class="['radio-group-radio radio-group-optional', item === activeGroup ? 'selected-radio' : 'is-not-selected-radio']"
          >
            <Icon v-if="isEditStatus" class="mr-2" @click="editGroup(item, index)" type="md-create" color="#2d8cf0" :size="15" ></Icon>
            <span @click="selectGroup(item)">
              {{ `${item}` }}
            </span>
            <Poptip
              confirm
              :title="$t('m_delConfirm_tip')"
              placement="left-end"
              @on-ok="confirmDeleteGroup(item, index)"
            >
              <Icon v-if="isEditStatus" class="ml-2" type="md-close" color="#ed4014" :size="15" ></Icon>
            </Poptip>
          </div>
          <span>
            <Button
              v-if="isEditStatus"
              @click="addGroupItem"
              style="margin-top: -5px;"
              type="success"
              shape="circle"
              icon="md-add"
            ></Button>
          </span>
        </div>

        <!-- 图表新增 -->
        <div class="chart-config-info" v-if="isEditStatus" @mousemove="debounceGetAllChartOptionList" @click="getAllChartOptionList">
          <span class="fs-20 mr-3 ml-3">{{$t('m_graph')}}:</span>
          <Dropdown
            placement="bottom-start"
            v-for="(item, index) in allAddChartOptions"
            :key="index"
            class="chart-option-menu"
            @on-click="(info) => onAddChart(JSON.parse(info), item.type)"
          >
            <Button :type="item.colorType">
              {{$t(item.name)}}
              <Icon type="ios-arrow-down"></Icon>
            </Button>
            <template #list>
              <DropdownMenu v-if="item.options.length > 0">
                <DropdownItem v-for="(option, key) in item.options"
                              :name="JSON.stringify(option)"
                              :key="key"
                              :disabled="option.disabled"
                >
                  <Icon v-if="option.iconType" :type="option.iconType" ></Icon>
                  {{item.type === 'add' ? $t(option.name) : option.name}}
                </DropdownItem>
              </DropdownMenu>
              <DropdownMenu v-else>
                <DropdownItem>
                  {{ $t('m_add_chart_library') }}
                </DropdownItem>
              </DropdownMenu>
            </template>
          </Dropdown>
        </div>
      </div>

      <!-- 图表展示区域 -->
      <div v-if="tmpLayoutData.length > 0" style="display:flex">
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
                       style="cursor: auto;"
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
              <template v-if="item.group === activeGroup || activeGroup === 'ALL'">
                <div class="c-dark grid-content">
                  <div class="header-grid header-grid-name">
                    <span @click="onChartBodyClick(item)" v-if="editChartId !== item.id">{{item.i}}</span>
                    <span  v-else @click.stop="">
                      <Input v-model.trim="item.i" class="editChartId" autofocus :maxlength="30" show-word-limit style="width:200px" size="small" placeholder="" />
                    </span>
                    <Tooltip :content="$t('m_placeholder_editTitle')" theme="light" transfer placement="top">
                      <i v-if="isEditStatus && editChartId !== item.id && !noAllowChartChange(item)" class="fa fa-pencil-square" style="font-size: 16px;" @click.stop="startEditTitle(item)" aria-hidden="true"></i>
                      <Icon v-if="editChartId === item.id" size="20" type="md-checkmark" @click.stop="onChartTitleChange(item)" ></Icon>
                      <Icon v-if="editChartId === item.id" size="20" type="md-close" @click.stop="cancelEditTitle(item)" ></Icon>
                    </Tooltip>
                  </div>
                  <div class="header-grid header-grid-tools">
                    <Button v-if="item.public" size="small" class="mr-1 mt-1 references-button">{{$t('m_shallow_copy')}}</Button>
                    <span @click.stop="">
                      <Select v-model="item.group"
                              style="width:100px;"
                              size="small"
                              :disabled="permission !== 'edit'"
                              clearable
                              filterable
                              :placeholder="$t('m_group_name')"
                              @on-change="onSingleChartGroupChange"
                      >
                        <Option v-for="item in panel_group_list" :value="item" :key="item" style="float: left;">{{ item }}</Option>
                      </Select>
                    </span>
                    <Tooltip :content="$t('m_save_chart_library')" theme="light" transfer placement="top">
                      <Icon v-if="isEditStatus && !item.public" size="15" type="md-archive" @click.stop="showChartAuthDialog(item)" ></Icon>
                    </Tooltip>
                    <Tooltip :content="$t('m_button_chart_dataView')" theme="light" transfer placement="top">
                      <i class="fa fa-eye" style="font-size: 16px;" v-if="isShowGridPlus(item)" aria-hidden="true" @click.stop="gridPlus(item)"></i>
                    </Tooltip>
                    <Tooltip :content="$t('m_placeholder_chartConfiguration')" theme="light" transfer placement="top">
                      <i class="fa fa-cog" style="font-size: 16px;" v-if="isEditStatus && !noAllowChartChange(item)" @click.stop="setChartType(item)" aria-hidden="true"></i>
                    </Tooltip>
                    <Poptip
                      confirm
                      :title="$t('m_delConfirm_tip')"
                      placement="bottom-end"
                      @on-ok="confirmRemoveGrid(item)"
                    >
                      <i class="fa fa-trash" style="font-size: 16px;color:red" v-if="isEditStatus" aria-hidden="true"></i>
                    </Poptip>
                  </div>
                </div>
                <section style="height: 90%;">
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
      <div v-else class="no-data">
        {{ $t('m_noData') }}
      </div>
    </div>
    <Drawer title="View details" :width="zoneWidth" v-model="showMaxChart">
      <ViewChart ref="viewChart"></ViewChart>
    </Drawer>

    <!-- 对于每个chart的抽屉详细信息 -->
    <Drawer :title="$t('m_placeholder_chartConfiguration')"
            :width="100"
            :mask-closable="false"
            v-model="showChartConfig"
            @on-close="closeChartInfoDrawer"
    >
      <editView :chartId="setChartConfigId" v-if="showChartConfig"></editView>
    </Drawer>

    <!-- 分组新增 -->
    <Modal v-model="showGroupMgmt" :title="$t('m_edit_screen_group')" :mask-closable="false">
      <div>
        <Form :label-width="90">
          <FormItem :label="$t('m_group_chart_name')">
            <Input v-model="groupName" placeholder="" style="width: 100%" :maxlength="20" show-word-limit />
          </FormItem>
          <FormItem :label="$t('m_use_charts')">
            <Row v-if="panelGroupInfo.length > 0">
              <Col span="12" v-for="panel in panelGroupInfo" :key="panel.name">
              <Checkbox v-model="panel.setGroup" :disabled="panel.hasGroup">
                <Tooltip :content="panel.label">
                  <div class="ellipsis-text">{{ panel.label }}</div>
                </Tooltip>
              </Checkbox>
              </Col>
            </Row>
            <span v-else>
              {{ $t('m_noData') }}
            </span>
          </FormItem>
        </Form>
      </div>
      <template #footer>
        <Button @click="showGroupMgmt = false">{{ $t('m_button_cancel') }}</Button>
        <Button @click="confirmGroupMgmt" :disabled="!groupName" type="primary">{{ $t('m_button_save') }}</Button>
      </template>
    </Modal>
    <AuthDialog ref="authDialog" :useRolesRequired="true" @sendAuth="saveChartOrDashboardAuth" ></AuthDialog>
    <ExportChartModal
      :isModalShow="isModalShow"
      :pannelId="pannelId"
      :panalName="panalName"
      @close="() => isModalShow = false"
    ></ExportChartModal>
  </div>
</template>

<script>
import isEmpty from 'lodash/isEmpty'
import remove from 'lodash/remove'
import cloneDeep from 'lodash/cloneDeep'
import debounce from 'lodash/debounce'
import {generateUuid} from '@/assets/js/utils'
import {dataPick, autoRefreshConfig} from '@/assets/config/common-config'
import {resizeEvent} from '@/assets/js/gridUtils.ts'
import VueGridLayout from 'vue-grid-layout'
import CustomChart from '@/components/custom-chart'
import CustomPieChart from '@/components/custom-pie-chart'
import ViewConfigAlarm from '@/views/custom-view/view-config-alarm'
import ViewChart from '@/views/custom-view/view-chart'
import EditView from '@/views/custom-view/edit-view'
import AuthDialog from '@/components/auth.vue'
import ExportChartModal from './export-chart-modal.vue'

export default {
  name: '',
  props: {
    boardId: {
      type: Number,
      default: null
    },
    permissionType: {
      type: String,
      default: ''
    },
    pageType: {
      type: String,
      default: ''
    }
  },
  data() {
    return {
      refreshNow: false,
      parentRouteData: {},
      editData: null,
      deleteConfirm: {
        id: '',
        method: ''
      },
      isEditPanal: false,
      permission: this.permissionType || this.$route.params.permission,
      panalName: '',
      pannelId: this.boardId || this.$route.params.pannelId,
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
      activeGroup: 'ALL',
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
          name: 'm_button_add',
          type: 'add',
          colorType: 'success',
          options: [
            {
              name: 'm_line_chart_s',
              id: 'line',
              iconType: 'md-trending-up'
            },
            {
              name: 'm_pie_chart',
              id: 'pie',
              iconType: 'md-pie'
            }
          ],
        },
        {
          name: 'm_copy',
          type: 'copy',
          colorType: 'success',
          options: []
        },
        {
          name: 'm_shallow_copy',
          type: 'shallowCopy',
          colorType: 'success',
          options: []
        }
      ],
      request: this.$root.$httpRequestEntrance.httpRequestEntrance,
      lineTypeOption: {
        twoYaxes: 2,
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
      boardUseRoles: [],
      initTitle: '',
      isModalShow: false
    }
  },
  computed: {
    tmpLayoutData() { // 缓存切换分组后数据
      if (this.activeGroup === 'ALL') {
        return this.layoutData
      }
      return this.layoutData.filter(d => d.group === this.activeGroup)

    },
    isEditStatus() {
      return this.permission === 'edit'
    },
  },
  mounted() {
    if (!this.pannelId) {
      return this.$router.push({path: '/viewConfigIndex/boardList'})
    }
    this.zoneWidth = window.screen.width * 0.65
    this.getAllChartOptionList()
    this.getPannelList()
    this.activeGroup = 'ALL'
    this.getAllRolesOptions()
  },
  methods: {
    getPannelList(activeGroup='ALL') {
      this.request('GET', '/monitor/api/v2/dashboard/custom', {
        id: this.pannelId
      }, res => {
        if (isEmpty(res)) {
          this.$router.push({path: 'viewConfigIndex'})
        }
        else {
          this.viewCondition.timeTnterval = res.timeRange || -3600
          this.viewCondition.autoRefresh = res.refreshWeek || 10
          this.boardMgmtRoles = res.mgmtRoles
          this.boardUseRoles = res.useRoles
          this.panalName = res.name
          this.activeGroup = activeGroup
          this.panel_group_list = res.panelGroupList || []
          this.viewData = res.charts || []
          this.initPanals()
          this.cutsomViewId = this.pannelId
        }
      })
    },
    getAllChartOptionList() {
      this.request('GET', '/monitor/api/v2/chart/shared/list', {}, res => {
        this.allAddChartOptions[1].options = this.processChartOptions(res)
      })
      this.request('GET', '/monitor/api/v2/chart/shared/list', {
        dashboard_id: this.pannelId
      }, res => {
        this.allAddChartOptions[2].options = this.processChartOptions(res)
      })
    },
    debounceGetAllChartOptionList: debounce(function () {
      this.getAllChartOptionList()
    }, 500),
    processChartOptions(rawData) {
      const options = []
      const initialOption = {
        line: {
          name: '折线图',
          id: 'line',
          iconType: 'md-trending-up',
          disabled: true
        },
        pie: {
          name: '饼图',
          id: 'pie',
          iconType: 'md-pie',
          disabled: true
        },
        bar: {
          name: '柱状图',
          id: 'bar',
          iconType: 'ios-stats',
          disabled: true
        }
      }

      for (const key in rawData) {
        options.push(initialOption[key])
        if (!isEmpty(rawData[key])) {
          rawData[key].forEach(item => {
            options.push(Object.assign({disabled: false}, item))
          })
        }
      }
      return options
    },
    onChartBodyClick(item) {
      if (this.isEditStatus && !this.noAllowChartChange(item)) {
        this.setChartType(item)
      }
      else {
        if (!this.isEditStatus) {
          this.gridPlus(item)
        }
        else {
          this.$Message.warning('暂无编辑权限！')
        }
      }
    },
    openAlarmDisplay() {
      this.showAlarm = !this.showAlarm
      setTimeout(() => {
        this.$refs.cutsomViewId.getAlarm(this.cutsomViewId, this.viewCondition, this.permission)
        this.refreshNow = !this.refreshNow
      }, 0)
    },
    closeAlarmDisplay() {
      this.showAlarm = !this.showAlarm
      this.$refs.cutsomViewId.clearAlarmInterval()
      this.refreshNow = !this.refreshNow
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
      const tmpArr = []
      this.viewData.forEach(item => {
        const parsedDisplayConfig = JSON.parse(item.displayConfig)
        const params = {
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
        item.chartSeries.forEach(item => {
          item.defaultColor = item.colorGroup
          if (item.series && !isEmpty(item.series)) {
            item.metricToColor = cloneDeep(item.series).map(one => {
              one.metric = one.seriesName
              delete one.seriesName
              return one
            })
          }
          else {
            item.metricToColor = []
          }
          params.data.push(item)
        })

        const height = (parsedDisplayConfig.h + 1) * 30-8
        const _activeCharts = []
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
      setTimeout(() => {
        this.refreshNow = !this.refreshNow
      }, 100)
    },
    isShowGridPlus(item) {
      if (!item._activeCharts || item._activeCharts[0].chartType === 'pie') {
        return false
      }
      return true
    },
    addItem() {
      this.activeGroup = 'ALL'
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
      this.setChartConfigId = item.id
      const find = this.layoutData.find(i => i.id === item.id)
      this.activeGridConfig = find
      if (!find._activeCharts) {
        // this.$root.JQ('#set_chart_type_Modal').modal('show')
      }
      else {
        this.activeChartType = find._activeCharts[0].chartType
        this.chartType = find._activeCharts[0].chartType
        this.showChartConfig = true
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
    },
    async confirmRemoveGrid(item) {
      this.deleteConfirm.id = item.id
      const params = this.processPannelParams()
      remove(params.charts, {id: this.deleteConfirm.id})
      if (!isEmpty(params.charts)) {
        params.charts[params.charts.length -1].displayConfig.x = 0
      }
      await this.requestReturnPromise('PUT', '/monitor/api/v2/dashboard/custom', params)
      this.getPannelList()
      this.getAllChartOptionList()
    },
    async gridPlus(item) {
      if (!this.isShowGridPlus(item)) {
        return
      }
      const resViewData = await this.modifyLayoutData()
      this.showMaxChart = true
      const templateData = {
        cfg: JSON.stringify(resViewData)
      }
      this.$refs.viewChart.initChart({
        templateData,
        panal: item,
        viewCondition: this.viewCondition
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
    resizeEvent(i, newH, newW, newHPx, newWPx){
      resizeEvent(this, i, newH, newW, newHPx, newWPx)
    },
    submitPanelInfo() {
      return new Promise(resolve => {
        if (isEmpty(this.boardMgmtRoles) || isEmpty(this.boardUseRoles)) {
          this.saveAuthType = 'board'
          this.$refs.authDialog.startAuth(this.boardMgmtRoles, this.boardUseRoles, this.mgmtRolesOptions, this.userRolesOptions)
        }
        else {
          this.request('PUT', '/monitor/api/v2/dashboard/custom', this.processPannelParams(), () => {
            resolve()
          })
        }
      })
    },

    async savePanelInfo() {
      await this.submitPanelInfo()
      this.$Message.success(this.$t('m_tips_success'))
    },
    returnPreviousPage() {
      this.$router.push({name: 'viewConfigIndex'})
    },
    async savePanalEdit() {
      if (!this.panalName) {
        this.$Message.warning(this.$t('m_tips_required'))
        return
      }
      await this.submitPanelInfo()
      this.$Message.success(this.$t('m_tips_success'))
      this.isEditPanal = false
    },
    canclePanalEdit() {
      this.isEditPanal = false
      this.panalName = this.editData.name
    },
    // #region 组管理
    selectGroup(item) {
      this.activeGroup = item
      this.getPannelList(this.activeGroup)
    },
    addGroupItem() {
      this.groupName = ''
      this.groupNameIndex = -1
      this.getPanelGroupInfo('')
      this.showGroupMgmt = true
    },
    editGroup(item ,index) {
      this.oriGroupName = item
      this.groupName = item
      this.groupNameIndex = index
      this.getPanelGroupInfo(item)
      this.showGroupMgmt = true
    },
    getPanelGroupInfo(groupName) {
      this.panelGroupInfo = []
      this.layoutData.forEach((d, dIndex) => {
        const group = {
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
        }
        else { // 已绑定
          group.hasGroup = d.group === groupName ? false : true
          group.setGroup = d.group === groupName ? true : false
        }
        this.panelGroupInfo.push(group)
      })
    },
    confirmDeleteGroup(group, index) {
      this.panel_group_list.splice(index, 1)
      this.layoutData.forEach(d => {
        if (d.group === group) {
          d.group = ''
        }
      })
      this.savePanalEdit()
      this.activeGroup = 'ALL'
    },
    confirmGroupMgmt() {
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
      }
      else {
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
            }
            else if (!p.setGroup) {
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
      let allName = []
      if (!isEmpty(this.layoutData)) {
        allName = this.layoutData.map(item => {
          if (!isNaN(Number(item.i))) {
            return Number(item.i)
          }
          return 10000

        })
        return Math.max(...allName) + 1 + ''
      }
      return '10000'
    },
    async onAddChart(copyInfo, type) {
      if (type === 'add') {
        const name = this.processInitialChartName()
        const addChartParams = {
          dashboardId: this.pannelId,
          name,
          chartTemplate: 'one',
          chartType: copyInfo.id,
          lineType: copyInfo.id === 'line' ? 'line' : '',
          pieType: copyInfo.id === 'pie' ? 'tag' : '',
          aggregate: 'none',
          aggStep: 60,
          unit: '',
          group: this.activeGroup === 'ALL' ? '' : this.activeGroup,
          displayConfig: {
            x: 0,
            y: 0,
            w: 6,
            h: 7,
          }
        }
        this.setChartConfigId = await this.requestReturnPromise('POST', '/monitor/api/v2/chart/custom', addChartParams)
        const item = {
          x: 0,
          y: 0,
          w: 6,
          h: 7,
          i: `${name}`,
          id: `${this.setChartConfigId}`
        }
        if (this.layoutData.length) {
          const lastItem = this.layoutData[this.layoutData.length - 1]
          if (lastItem.x <= 6 && lastItem.w <= 6) {
            const popItem = this.layoutData.pop()
            popItem.x = 6
            this.layoutData.push(popItem)
          }
        }
        this.layoutData.push(item)
        setTimeout(() => {
          this.request('PUT', '/monitor/api/v2/dashboard/custom', this.processPannelParams(), () => {
            this.getPannelList()
            this.showChartConfig = true
          })
        }, 0)
      }
      else {
        const group = type === 'copy' ? '' : (this.activeGroup === 'ALL' ? '' : this.activeGroup)
        const copyParams = {
          dashboardId: this.pannelId,
          ref: type === 'copy' ? false : true,
          originChartId: copyInfo.id,
          group,
          displayConfig: {
            x: 0,
            y: 0,
            w: 6,
            h: 7,
          }
        }
        const chartId = await this.requestReturnPromise('POST', '/monitor/api/v2/chart/custom/copy', copyParams)
        const item = {
          x: 0,
          y: 0,
          w: 6,
          h: 7,
          i: '',
          id: `${chartId}`,
          group
        }
        this.layoutData.push(item)
        setTimeout(async () => {
          await this.requestReturnPromise('PUT', '/monitor/api/v2/dashboard/custom', this.processPannelParams())
          this.getPannelList()
          this.setChartConfigId = chartId
          if (type === 'copy') {
            this.showChartConfig = true
          }
          this.getAllChartOptionList()
        }, 0)
      }
    },
    processPannelParams() {
      const charts = []
      this.layoutData.forEach(item => {
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
        charts,
        panelGroups: this.panel_group_list,
        timeRange: this.viewCondition.timeTnterval,
        refreshWeek: this.viewCondition.autoRefresh
      }
    },
    requestReturnPromise(method, api, params) {
      return new Promise(resolve => {
        this.request(method, api, params, res => {
          resolve(res)
        })
      })
    },
    startEditTitle(item) {
      this.initTitle = item.i
      this.editChartId = item.id
    },

    cancelEditTitle(item) {
      item.i = this.initTitle
      this.editChartId = null
    },

    async onChartTitleChange(item) {
      const isDuplicateName = await this.chartDuplicateNameCheck(item.id, item.i)
      if (isDuplicateName) {
        return
      }
      await this.requestReturnPromise('PUT', '/monitor/api/v2/chart/custom/name', {
        chartId: item.id,
        name: item.i
      })
      this.editChartId = null
      this.$Message.success(this.$t('m_tips_success'))
    },
    chartDuplicateNameCheck(chartId, chartName, isPublic = 0) {
      return new Promise(resolve => {
        if (!chartName) {
          resolve(true)
          return
        }
        const params = {
          chart_id: chartId,
          name: chartName
        }
        if (isPublic) {
          params.public = 1 // 是否存入图表库，1表示是
        }
        this.request('GET', '/monitor/api/v2/chart/custom/name/exist', params, res => {
          if (res) {
            this.$Message.error(isPublic ? (this.$t('m_chart_library') + this.$t('m_name') + this.$t('m_cannot_be_repeated')) : (this.$t('m_graph_name') + this.$t('m_cannot_be_repeated')))
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
      this.chartAuthDialogShow = true
      if (await this.chartDuplicateNameCheck(item.id, item.i, 1)) {
        return
      }
      await this.getSingleChartAuth()
      this.saveAuthType = 'chart'
      this.$refs.authDialog.startAuth(this.mgmtRoles, this.useRoles, this.mgmtRolesOptions, this.userRolesOptions)
    },
    getSingleChartAuth() {
      return new Promise(resolve => {
        this.request('GET','/monitor/api/v2/chart/custom/permission', {
          chart_id: this.setChartConfigId
        }, res => {
          this.mgmtRoles = res.mgmtRoles
          this.useRoles = res.useRoles
          resolve()
        })
      })
    },
    getAllRolesOptions() {
      const params = {
        page: 1,
        size: 1000
      }
      this.request('GET','/monitor/api/v1/user/role/list', params, res => {
        this.userRolesOptions = this.processRolesList(res.data)
      })
      this.request('GET', '/monitor/api/v1/user/manage_role/list', {}, res => {
        this.mgmtRolesOptions = this.processRolesList(res)
      })
    },
    processRolesList(list = []) {
      if (isEmpty(list)) {
        return []
      }
      const resArr = cloneDeep(list).map(item => ({
        ...item,
        key: item.name,
        label: item.displayName || item.display_name
      }))
      return resArr
    },
    saveChartOrDashboardAuth(mgmtRoles, useRoles) {
      let path = ''
      const params = {
        mgmtRoles,
        useRoles
      }
      if (this.saveAuthType === 'chart') {
        this.mgmtRoles = mgmtRoles
        this.useRoles = useRoles
        path = '/monitor/api/v2/chart/custom/permission'
        params.chartId = this.setChartConfigId
      }
      else {
        this.boardMgmtRoles = mgmtRoles
        this.boardUseRoles = useRoles
        path = '/monitor/api/v2/dashboard/custom/permission'
        params.id = this.pannelId
      }
      this.request('POST', path, params, () => {
        this.$Message.success(this.$t('m_success'))
        this.getPannelList()
      })
    },
    onSingleChartGroupChange() {
      this.request('PUT', '/monitor/api/v2/dashboard/custom', this.processPannelParams(), () => {
        this.$Message.success(this.$t('m_success'))
        this.getPannelList(this.activeGroup)
        // this.activeGroup = 'ALL';
      })
    },
    closeChartInfoDrawer() {
      this.getPannelList()
      setTimeout(() => {
        this.refreshNow = !this.refreshNow
      }, 500)
    },
    exportPanel() {
      this.isModalShow = true
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
    AuthDialog,
    ExportChartModal
  },
}
</script>

<style lang="less">
.ivu-poptip-popper {
  color: #515a6e
}

.chart-config-info {
  .ivu-dropdown-item-disabled {
    color: inherit
  }
}

.chart-option-menu {
  .ivu-dropdown-menu {
    max-height: 600px;
    overflow: scroll;
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
  display: flex;
  align-items: center;
  i {
    font-size: 18px !important
  }
}
.header-tools {
  .ivu-btn-info {
    background-color: #aa8aea;
    border-color: #aa8aea;
  }
}
</style>

<style scoped lang="less">
/deep/ .ivu-form-item {
  margin-bottom: 0;
}

.arrow-back {
  cursor: pointer;
  width: 28px;
  height: 24px;
  color: #fff;
  border-radius: 2px;
  background: #2d8cf0;
  margin-right: 12px;
  margin-bottom: 10px;
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
  display: flex;
  flex-grow: 1;
  justify-content: flex-end;
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
    // width: 700px
  }

  .no-data {
    text-align: center;
    margin: 32px;
    font-size: 14px;
    color: gray;
  }

  .ellipsis-text {
    width: 170px;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    display: inline-block;
    vertical-align: bottom;
  }

.all-content {
  ::-webkit-scrollbar {
    display: none;
  }
}
</style>
