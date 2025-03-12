<template>
  <div class="monitor-custom-view-config">
    <div>
      <div class='view-config-top-content'>
        <header>
          <div class="header-name">
            <div v-if="pageType !== 'dashboard'">
              <Icon v-if="pageType !== 'link'" size="22" class="arrow-back" type="md-arrow-back" @click="returnPreviousPage" ></Icon>
              <Tag v-if='logMetricGroup' class='ml-2 mr-2' style="font-size: 14px; min-width: 46px" color='green'>auto</Tag>
              <template v-if="isEditPanal">
                <Input v-model.trim="panalName" style="width: 300px" type="text" :maxlength="30" show-word-limit/>
                <Icon class="panal-edit-icon" color="#5384FF" @click="savePanalEdit" type="md-checkmark" ></Icon>
                <Icon class="panal-edit-icon" color="red" @click="canclePanalEdit" type="md-trash" ></Icon>
              </template>
              <template v-else>
                <h5 class="d-inline-block"> {{panalName}}</h5>
                <Icon class="panal-edit-icon" color="#5384FF"  @click="isEditPanal = true" v-if="isEditStatus" type="md-create" ></Icon>
              </template>
              <span class="drop-icon-region" @click="onIconRegionClick">
                <Icon v-if="isActionRegionExpand" color="#5384FF" :size="22" type="ios-arrow-dropup" />
                <Icon v-else :size="22" color="#5384FF" type="ios-arrow-dropdown" />
              </span>
            </div>
            <div class="header-tools">
              <template v-if="isEditStatus">
                <Button class="btn-upload" @click.stop="exportPanel">
                  <img src="@/styles/icon/DownloadOutlined.png" class="upload-icon" />
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
        <div class="all-action-region" :style="{maxHeight: isActionRegionExpand ? '190px' : '0px', marginTop: pageType === 'dashboard' ? '35px' : '0px'}">

          <div class="search-container">
            <div>
              <div class="search-zone">
                <span class="params-title">{{$t('m_field_relativeTime')}}：</span>
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
                  style="width: 250px"
                ></DatePicker>
              </div>
              <div class="search-zone">
                <span class="params-title">{{$t('m_placeholder_refresh')}}：</span>
                <Select filterable clearable v-model="viewCondition.autoRefresh" :disabled="disableTime" style="width:100px" @on-change="initPanals" :placeholder="$t('m_placeholder_refresh')">
                  <Option v-for="item in autoRefreshConfig" :value="item.value" :key="item.value">{{ item.label }}</Option>
                </Select>
              </div>
            </div>
          </div>
          <div class="radio-group">
            <span class="ml-3 mr-3">{{$t('m_group_name')}}:</span>
            <div class='group-region'>
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
                <Icon v-if="isEditStatus" class="mr-2" @click="editGroup(item, index)" type="md-create" color="#5384FF" :size="15" ></Icon>
                <span @click="selectGroup(item)">
                  {{ `${item}` }}
                </span>
                <Poptip
                  confirm
                  :title="$t('m_delConfirm_tip')"
                  placement="left-end"
                  @on-ok="confirmDeleteGroup(item, index)"
                >
                  <Icon v-if="isEditStatus" class="ml-2" type="md-close" color="#FF4D4F" :size="15" ></Icon>
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
          </div>

          <div class="ml-2 mt-2 mb-2">
            <span class="params-title mr-3">{{$t('m_automatic_layout')}}：</span>
            <Poptip
              confirm
              transfer
              popper-class='chart-layout-poptip'
              :title="$t('m_layout_change_tips')"
              @on-ok="onLayoutPopTipConfirm"
            >
              <Button id='chartLayoutPopTipButton' style="display: none"></Button>
            </Poptip>
            <RadioGroup @on-change="onLayoutRadioChange" v-model="chartLayoutType" type="button" size="small">
              <Radio v-for="(item, idx) in layoutOptions" :label="item.value" :key="idx" :disabled="disableTime">{{ $t(item.label) }}</Radio>
            </RadioGroup>
          </div>

          <!-- 图表新增 -->
          <div class="chart-config-info" v-if="isEditStatus">
            <span class="fs-20 mr-3 ml-3">{{$t('m_graph')}}:</span>
            <Dropdown
              placement="bottom-start"
              trigger="click"
              v-for="(item, index) in allAddChartOptions"
              :key="index"
              transfer
              class="chart-option-menu"
              transfer-class-name='filter-chart-layer'
              @on-click="(info) => onAddChart(JSON.parse(info), item.type)"
            >
              <Button :type="item.colorType" @click="onCopyButtonClick">
                {{$t(item.name)}}
                <Icon type="ios-arrow-down"></Icon>
              </Button>
              <template slot='list' >
                <div>
                  <div v-if="item.type === 'add'">
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
                  </div>
                  <div v-else class='copy-drowdown-slot'>
                    <div class='copy-drowdown-slot-select'>
                      <Input v-model.trim="filterChartName"
                             clearable
                             :placeholder="$t('m_placeholder_input') + $t('m_graph_name')"
                             @on-change="(e) => {
                               filterChartName = e.target.value
                               debounceGetAllChartOptionList()
                             }"
                      />
                      <Select v-model="selectedDashBoardId"
                              clearable
                              filterable
                              :placeholder="$t('m_please_select') + $t('m_source_dashboard')"
                              @on-change="(id) => {
                                selectedDashBoardId = id
                                debounceGetAllChartOptionList()
                              }"
                      >
                        <Option v-for="single in allDashBoardList"
                                :value="single.id"
                                :key="single.id"
                        >
                          {{ single.name }}
                        </Option>
                      </Select>
                    </div>
                    <div class="copy-table-tips">{{$t('m_copy_table_tips')}}</div>
                    <Table
                      class="copy-detail-table"
                      size="small"
                      max-height="300"
                      :border="false"
                      :columns="copyTableColumns"
                      :data="allChartFilteredList"
                      @on-selection-change='onCopyTableSelected'
                    />

                    <Button
                      type="primary"
                      class='copy-drowdown-confirm-button'
                      @click="onAddChart(selectedChartList, item.type)"
                    >
                      {{$t('m_confirm')}}
                    </Button>
                  </div>

                </div>
              </template>
            </Dropdown>
          </div>
        </div>
      </div>

      <!-- 图表展示区域 -->
      <div v-if="tmpLayoutData.length > 0" style="display: flex" class=''>
        <div class="grid-window" :style="pageType === 'link' ? 'height: calc(100vh - 250px)' : ''" @scroll="onGridWindowScroll">
          <grid-layout
            :layout.sync="tmpLayoutData"
            :col-num="12"
            :row-height="30"
            :is-draggable="true"
            :is-resizable="isEditStatus"
            :is-mirrored="false"
            :vertical-compact="true"
            :use-css-transforms="true"
          >
            <grid-item v-for="(item,index) in tmpLayoutData"
                       style="cursor: auto; overflow-y: hidden;"
                       class="c-dark"
                       :x="item.x"
                       :y="item.y"
                       :w="item.w"
                       :h="item.h"
                       :i="item.i"
                       :minW="1"
                       :minH="5"
                       :key="index"
                       @resize="resizeEvent"
                       @resized="resizeEvent"
            >
              <template v-if="item.group === activeGroup || activeGroup === 'ALL'">
                <div class="c-dark grid-content">
                  <Tag v-if="item.logMetricGroup" class='grid-auto-tag-style' color='green'>auto</Tag>
                  <div class="header-grid header-grid-name">
                    <Tooltip v-if="editChartId !== item.id" :content="item.i" transfer :max-width='250' placement="bottom">
                      <div v-html="processHtmlText(item.i)" class='header-grid-name-text'></div>
                    </Tooltip>
                    <span v-else @click.stop="">
                      <Input v-model.trim="item.i" class="editChartId" autofocus :maxlength="100" show-word-limit style="width:150px" size="small" placeholder="" />
                    </span>
                    <Tooltip :content="$t('m_placeholder_editTitle')" theme="light" transfer placement="bottom">
                      <i v-if="isEditStatus && editChartId !== item.id && !noAllowChartChange(item)" class="fa fa-pencil-square" style="font-size: 16px;" @click.stop="startEditTitle(item)" aria-hidden="true"></i>
                      <Icon v-if="editChartId === item.id" size="20" type="md-checkmark" @click.stop="onChartTitleChange(item)" ></Icon>
                      <Icon v-if="editChartId === item.id" size="20" type="md-close" @click.stop="cancelEditTitle(item)" ></Icon>
                    </Tooltip>
                  </div>
                  <div class="header-grid header-grid-tools">
                    <Tag v-if="item.public" class="mr-1 mt-1 references-button">{{$t('m_shallow_copy')}}</Tag>
                    <span @click.stop="">
                      <Select v-model="item.group"
                              style="width:100px;"
                              size="small"
                              :disabled="permission !== 'edit'"
                              clearable
                              filterable
                              :placeholder="$t('m_group_name')"
                              @on-change="(e) => onSingleChartGroupChange(e, index, item)"
                              @on-clear="() => onSingleChartGroupClear(item)"
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
                    <Tooltip :content="$t('m_line_display_modification')" theme="light" transfer placement="top">
                      <Icon type="ios-funnel" size="16" @click="showLineSelectModal(item)" />
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
                    <CustomChart v-if="['line','bar'].includes(chartInfo.chartType)"
                                 :refreshNow="refreshNow"
                                 :scrollRefresh="scrollRefresh"
                                 :chartInfo="chartInfo"
                                 :chartIndex="index"
                                 :params="viewCondition"
                                 :hasNotRequestStatus="hasNotRequestStatus"
                    >
                    </CustomChart>
                    <CustomPieChart v-if="chartInfo.chartType === 'pie'" :refreshNow="refreshNow" :chartInfo="chartInfo" :chartIndex="index" :params="viewCondition"></CustomPieChart>
                  </div>
                </section>
              </template>
            </grid-item>
          </grid-layout>
        </div>
        <div class="view-config-alarm" :style="pageType === 'link' ? 'height: calc(100vh - 250px)' : ''" v-if="showAlarm">
          <ViewConfigAlarm ref="cutsomViewId"></ViewConfigAlarm>
        </div>
      </div>
      <div v-else class="no-data">
        {{ $t('m_noData') }}
      </div>
    </div>
    <Drawer :title="$t('m_view_details')" :width="zoneWidth" v-model="showMaxChart">
      <ViewChart v-if="showMaxChart" ref="viewChart"></ViewChart>
    </Drawer>

    <!-- 对于每个chart的抽屉详细信息 -->
    <Drawer :title="$t('m_placeholder_chartConfiguration')"
            :width="100"
            :mask-closable="false"
            v-model="showChartConfig"
            @on-close="closeChartInfoDrawer"
    >
      <editView id='edit-view' :chartId="setChartConfigId" v-if="showChartConfig"></editView>
    </Drawer>

    <!-- 分组新增 -->
    <Modal v-model="showGroupMgmt"
           :title="groupNameIndex === -1 ? $t('m_add_screen_group') : $t('m_edit_screen_group')"
           :mask-closable="false"
           :width="1000"
    >
      <div>
        <Form :label-width="90">
          <FormItem :label="$t('m_group_chart_name')">
            <Input v-model="groupName" placeholder="" style="width: 100%" :maxlength="20" show-word-limit />
          </FormItem>
          <FormItem :label="$t('m_use_charts')">
            <Row v-if="panelGroupInfo.length > 0" style="min-height: 200px; max-height: 400px;overflow-y: auto;">
              <Col span="12" v-for="panel in panelGroupInfo" :key="panel.name">
              <Checkbox v-model="panel.setGroup" :disabled="panel.hasGroup">
                <Tooltip :content="panel.label" transfer :max-width='200'>
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
      <template slot='footer'>
        <Button @click="showGroupMgmt = false">{{ $t('m_button_cancel') }}</Button>
        <Button @click="confirmGroupMgmt" :disabled="!groupName" type="primary">{{ $t('m_button_save') }}</Button>
      </template>
    </Modal>

    <!-- 实现线条是否展现弹窗 -->
    <ChartLinesModal
      :isLineSelectModalShow="isLineSelectModalShow"
      :chartId="setChartConfigId"
      @modalClose="onLineSelectChangeCancel"
    >
    </ChartLinesModal>
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
import {
  isEmpty, remove, cloneDeep, find, orderBy, maxBy, filter, debounce
} from 'lodash'
import {generateUuid} from '@/assets/js/utils'
import {
  dataPick, autoRefreshConfig, layoutOptions, layoutColumns
} from '@/assets/config/common-config'
import {resizeEvent} from '@/assets/js/gridUtils.ts'
import VueGridLayout from 'vue-grid-layout'
import CustomChart from '@/components/custom-chart'
import CustomPieChart from '@/components/custom-pie-chart'
import ChartLinesModal from '@/components/chart-lines-modal'
import ViewConfigAlarm from '@/views/custom-view/view-config-alarm'
import ViewChart from '@/views/custom-view/view-chart'
import EditView from '@/views/custom-view/edit-view'
import AuthDialog from '@/components/auth.vue'
import ExportChartModal from './export-chart-modal.vue'
// import { changeSeriesColor } from '@/assets/config/random-color'

const lineTypeNameMap = {
  line: 'm_line_chart_s',
  pie: 'm_pie_chart',
  bar: 'm_bar_chart'
}

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
        timeTnterval: -1800,
        dateRange: ['', ''],
        autoRefresh: 10,
        agg: 'none' // 聚合类型
      },
      disableTime: false,
      dataPick,
      autoRefreshConfig,
      layoutOptions,
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
          options: {
            chartOptions: [],
            sourceBoardOptions: []
          }
        },
        {
          name: 'm_shallow_copy',
          type: 'shallowCopy',
          colorType: 'success',
          options: {
            chartOptions: [],
            sourceBoardOptions: []
          }
        }
      ],
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
      isModalShow: false,
      chartLayoutType: 'customize',
      previousChartLayoutType: 'customize', // 用于记录radio点击前的type
      tempChartLayoutType: '', // 用于记录点击后的type,
      allPageLayoutData: [],
      selectedDashBoardId: 0,
      selectedChartList: [],
      logMetricGroup: '', // 该字段非空代表该看板为自动创建
      allDashBoardList: [],
      allChartFilteredList: [],
      filterChartName: '',
      copyTableColumns: [
        {
          type: 'selection',
          width: 60,
          align: 'center'
        },
        {
          title: this.$t('m_graph_name'),
          minWidth: 350,
          key: 'name',
        },
        {
          title: this.$t('m_screen_name'),
          width: 200,
          key: 'dashboardName',
        },
        {
          title: this.$t('m_graph_type'),
          minWidth: 220,
          key: 'lineType',
          render: (h, params) => (
            <div>
              {this.$t(lineTypeNameMap[params.row.lintType])}
            </div>
          )
        }
      ],
      isShowLoading: false,
      isLineSelectModalShow: false,
      isEmpty,
      scrollRefresh: false,
      hasNotRequestStatus: true,
      request: this.$root.$httpRequestEntrance.httpRequestEntrance,
      apiCenter: this.$root.apiCenter,
      isActionRegionExpand: true
    }
  },
  computed: {
    tmpLayoutData() { // 缓存切换分组后数据
      return this.layoutData
    },
    isEditStatus() {
      return this.permission === 'edit'
    }
  },
  mounted() {
    if (!this.pannelId) {
      return this.$router.push({path: '/viewConfigIndex/boardList'})
    }
    this.zoneWidth = window.screen.width * 0.65
    this.isShowLoading = true
    this.getAllChartOptionList()
    this.getPannelList()
    this.activeGroup = 'ALL'
    this.getAllRolesOptions()
    window['view-config-selected-line-data'] = {}
    setTimeout(() => {
      const domArr = document.querySelectorAll('.copy-drowdown-slot')
      !isEmpty(domArr) && domArr.forEach(dom => dom.addEventListener('click', e => e.stopPropagation()))
    }, 500)
  },
  methods: {
    getPannelList(activeGroup=this.activeGroup) {
      this.request('GET', this.apiCenter.template.deleteV2, {
        id: this.pannelId
      }, res => {
        if (isEmpty(res)) {
          this.$router.push({path: 'viewConfigIndex'})
        } else {
          this.viewCondition.timeTnterval = res.timeRange || -1800
          this.viewCondition.autoRefresh = res.refreshWeek || 60
          this.boardMgmtRoles = res.mgmtRoles
          this.boardUseRoles = res.useRoles
          this.panalName = res.name
          this.logMetricGroup = res.logMetricGroup
          this.activeGroup = activeGroup
          this.panel_group_list = res.panelGroupList || []
          this.viewData = res.charts || []
          this.initPanals('init')
          this.cutsomViewId = this.pannelId
        }
      })
    },
    onCopyButtonClick() {
      this.filterChartName = ''
      this.selectedDashBoardId = 0
      this.selectedChartList = []
      this.getAllChartOptionList()
    },
    debounceGetAllChartOptionList: debounce(function (){
      this.getAllChartOptionList()
    }, 500),

    getAllChartOptionList() {
      if (!this.isEditStatus) {return}
      this.selectedChartList = []
      const dashboardParams = {
        show: '',
        permission: '',
        name: '',
        id: 0,
        useRoles: [],
        mgmtRoles: [],
        updateUser: '',
        pageSize: 10000,
        startIndex: 0
      }
      this.request('POST', this.apiCenter.template.listV2, dashboardParams, res => {
        this.allDashBoardList = res.contents
      }, {isNeedloading: false})

      const params = {
        curDashboardId: this.pannelId,
        dashboardId: this.selectedDashBoardId,
        chartName: this.filterChartName
      }
      this.request('POST', this.apiCenter.chartSharedList, params, res => {
        this.allChartFilteredList = this.processChartOptions(res)
      }, {isNeedloading: false})
    },
    processChartOptions(rawData) {
      const options = []
      for (const key in rawData) {
        if (!isEmpty(rawData[key])) {
          rawData[key].forEach(item => {
            options.push(Object.assign({lintType: key}, item))
          })
        }
      }
      return options
    },
    openAlarmDisplay() {
      this.showAlarm = !this.showAlarm
      this.$emit('alarmStatueChange', this.showAlarm)
      setTimeout(() => {
        this.$refs.cutsomViewId.getAlarm(this.cutsomViewId, this.viewCondition, this.permission)
        this.refreshNow = !this.refreshNow
        this.calculateGridWindowHeight()
      }, 300)
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
        this.viewCondition.dateRange[1] = this.viewCondition.dateRange[1].replace('00:00:00', '23:59:59')
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
    async initPanals(type) {
      const tmpArr = []
      this.isShowLoading = false

      for (let k=0; k<this.viewData.length; k++) {
        const item = this.viewData[k]
        // 先对groupDisplayConfig进行初始化，防止异常值
        if (!this.isValidJson(item.groupDisplayConfig) || isEmpty(JSON.parse(item.groupDisplayConfig))) {
          item.groupDisplayConfig = ''
        }
        let parsedDisplayConfig = {}
        if (this.activeGroup === 'ALL') {
          parsedDisplayConfig = JSON.parse(item.displayConfig)
        } else {
          // 在各个分组中的情况
          if (this.isValidJson(item.groupDisplayConfig) && !isEmpty(JSON.parse(item.groupDisplayConfig))) {
            parsedDisplayConfig = JSON.parse(item.groupDisplayConfig)
          } else {
            item.groupDisplayConfig = ''
            parsedDisplayConfig = JSON.parse(item.displayConfig)
          }
        }
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
        for (let i=0; i<item.chartSeries.length; i++) {
          const single = item.chartSeries[i]
          single.defaultColor = single.colorGroup
          params.data.push(single)
        }
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
          end: this.dateToTimestamp(this.viewCondition.dateRange[1]),
          parsedDisplayConfig
        })
        tmpArr.push({
          _activeCharts,
          ...parsedDisplayConfig,
          i: item.name,
          id: item.id,
          group: item.group,
          public: item.public,
          sourceDashboard: item.sourceDashboard,
          allGroupDisplayConfig: JSON.parse(item.displayConfig),
          partGroupDisplayConfig: item.groupDisplayConfig === '' ? '' : JSON.parse(item.groupDisplayConfig),
          logMetricGroup: item.logMetricGroup
        })
      }
      if (isEmpty(this.layoutData) || type === 'init') {
        this.layoutData = tmpArr
      } else {
        tmpArr.forEach(tmpItem => {
          const item = find(this.layoutData, {
            id: tmpItem.id
          })
          if (!isEmpty(item)) {
            item._activeCharts = tmpItem._activeCharts
          }
        })
      }
      this.layoutData = this.sortLayoutData(cloneDeep(this.layoutData))
      if (type === 'init') {
        this.allPageLayoutData = cloneDeep(this.layoutData)
      }
      this.resetHasNotRequestStatus()
      this.filterLayoutData()
    },
    processBasicParams(metric, endpoint, serviceGroup, monitorType, tags, chartSeriesGuid = '', allItem = {}) {
      let tempTags = tags
      if (allItem.comparison && !isEmpty(tags) && tags[0].tagName === 'calc_type' && isEmpty(tags[0].tagValue)) {
        tempTags = []
      }
      return {
        metric,
        endpoint,
        serviceGroup,
        monitorType,
        tags: tempTags,
        chartSeriesGuid
      }
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
      } else {
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
      remove(this.layoutData, {id: this.deleteConfirm.id})
      remove(this.allPageLayoutData, {id: this.deleteConfirm.id})
      this.filterLayoutData()
      const params = this.processPannelParams()
      await this.request('PUT', this.apiCenter.template.deleteV2, params)
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
      setTimeout(() => {
        this.$refs.viewChart.initChart({
          templateData,
          panal: item,
          viewCondition: this.viewCondition
        })
      }, 50)
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
        } else {
          this.request('PUT', this.apiCenter.template.deleteV2, this.processPannelParams(), () => {
            this.refreshPannelNow()
            resolve()
          })
        }
      })
    },

    async savePanelInfo() {
      await this.submitPanelInfo()
      this.$Message.success(this.$t('m_save_success'))
      this.chartLayoutType = 'customize'
      this.getPannelList(this.activeGroup)
    },
    returnPreviousPage() {
      this.$router.push({
        name: 'viewConfigIndex',
        query: {
          needCache: 'yes'
        }
      })
    },
    async savePanalEdit() {
      if (!this.panalName) {
        this.$Message.warning(this.$t('m_tips_required'))
        return
      }
      await this.submitPanelInfo()
      this.$Message.success(this.$t('m_save_success'))
      this.isEditPanal = false
      this.getPannelList(this.activeGroup)
    },
    canclePanalEdit() {
      this.isEditPanal = false
      this.panalName = this.editData.name
    },
    // #region 组管理
    async selectGroup(item) {
      if (this.isEditStatus) {
        await this.submitPanelInfo() // 保存完毕后切换
        this.$Message.success(this.$t('m_save_success'))
      }
      this.activeGroup = item
      this.chartLayoutType = 'customize'
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
      this.allPageLayoutData.forEach((d, dIndex) => {
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
        } else { // 已绑定
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
      this.allPageLayoutData.forEach(d => {
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
            this.allPageLayoutData[p.index].group = this.groupName
          }
        })
      } else {
        this.panel_group_list[this.groupNameIndex] = this.groupName
        this.allPageLayoutData.forEach(d => {
          if (d.group === this.oriGroupName) {
            d.group = this.groupName
          }
        })
        this.panelGroupInfo.forEach(p => {
          if (p.hasGroup === false) {
            if (p.setGroup && p.group === '') {
              this.allPageLayoutData[p.index].group = this.groupName
            } else if (!p.setGroup) {
              this.allPageLayoutData[p.index].group = ''
            }
          }
        })
      }
      this.showGroupMgmt = false
      this.activeGroup = this.groupName
      this.savePanalEdit()
    },

    processInitialChartName() {
      let allName = []
      const initialList = this.layoutData.length > this.allPageLayoutData.length ? this.layoutData : this.allPageLayoutData
      if (!isEmpty(initialList)) {
        allName = initialList.map(item => {
          if (!isNaN(Number(item.i))) {
            return Number(item.i)
          }
          return '10000'
        })
        return Math.max(...allName) + 1 + ''
      }
      return '10000'
    },
    processSingleItem(lastItem, needSetItem) {
      if (lastItem.x + lastItem.w + needSetItem.w > 12) {
        needSetItem.x = 0
        needSetItem.y = lastItem.y + 7
      } else {
        needSetItem.x = lastItem.x + lastItem.w
        needSetItem.y = lastItem.y
      }
    },
    closeDropDownView(){
      const elements = document.querySelectorAll('.ivu-select-dropdown')
      elements.forEach(el => {
        if (getComputedStyle(el).display !== 'none') {
          el.style.display = 'none'
        }
      })
    },
    async onAddChart(copyInfo, type) {
      this.closeDropDownView()
      if (type === 'add') {
        // type为add的时候为新增，默认copyInfo只有一个
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
          groupDisplayConfig: '',
          displayConfig: {
            x: 0,
            y: 0,
            w: 6,
            h: 7,
          }
        }
        this.setChartConfigId = await this.request('POST', this.apiCenter.chartInfo, addChartParams)
        const item = {
          x: 0,
          y: 0,
          w: layoutColumns[this.chartLayoutType].w,
          h: 7,
          i: `${name}`,
          moved: false,
          id: `${this.setChartConfigId}`,
          group: this.activeGroup === 'ALL' ? '' : this.activeGroup,
          partGroupDisplayConfig: '',
          allGroupDisplayConfig: this.calculateDisplayConfig(),
          sourceDashboard: this.pannelId
        }
        if (this.layoutData.length) {
          const lastOne = this.layoutData[this.layoutData.length - 1]
          this.processSingleItem(lastOne, item)
        }
        this.layoutData.unshift(item)
        setTimeout(() => {
          this.request('PUT', this.apiCenter.dashboardCustom, this.processPannelParams(), () => {
            this.getPannelList(this.activeGroup)
            this.showChartConfig = true
          })
        }, 0)
      } else {
        // type不为add的时候，copyInfo为id的数组
        if (isEmpty(copyInfo)) {return}
        let chartId = ''
        for (let i=0; i<copyInfo.length; i++) {
          const group = type === 'copy' ? '' : (this.activeGroup === 'ALL' ? '' : this.activeGroup)
          const copyParams = {
            dashboardId: this.pannelId,
            ref: type === 'copy' ? false : true,
            originChartId: copyInfo[i].id,
            group,
            groupDisplayConfig: '',
            displayConfig: {
              x: 0,
              y: 0,
              w: 6,
              h: 7,
            }
          }
          chartId = await this.request('POST', this.apiCenter.chartCustomCopy, copyParams)
          const item = {
            x: 0,
            y: 0,
            w: layoutColumns[this.chartLayoutType].w,
            h: 7,
            i: this.processInitialChartName(),
            id: `${chartId}`,
            group,
            partGroupDisplayConfig: '',
            allGroupDisplayConfig: this.calculateDisplayConfig(),
            sourceDashboard: this.pannelId
          }
          if (this.layoutData.length) {
            const lastOne = this.layoutData[this.layoutData.length - 1]
            this.processSingleItem(lastOne, item)
          }
          this.layoutData.unshift(item)
        }
        if (copyInfo.length > 1) {
          this.initLayoutTypeByWidth(this.layoutData)
        }
        setTimeout(async () => {
          await this.request('PUT', this.apiCenter.dashboardCustom, this.processPannelParams())
          this.getPannelList(this.activeGroup)
          this.setChartConfigId = chartId
          if (type === 'copy') {
            this.showChartConfig = true
          }
          this.getAllChartOptionList()
        }, 50)
      }
      setTimeout(() => {
        document.querySelector('.vue-grid-layout').scrollIntoView({
          behavior: 'smooth',
          block: 'end'
        })
      }, 500)
    },
    extendTime() {
      return new Promise(resolve => {
        setTimeout(() => {
          resolve()
        }, 50)
      })
    },
    processPannelParams() {
      const finalCharts = []
      this.layoutData.forEach(item => {
        finalCharts.push({
          id: item.id,
          group: item.group,
          name: item.i,
          displayConfig: this.activeGroup === 'ALL' ? {
            x: item.x,
            y: item.y,
            w: item.w,
            h: item.h
          } : item.allGroupDisplayConfig,
          groupDisplayConfig: this.activeGroup === 'ALL' ? item.partGroupDisplayConfig : {
            x: item.x,
            y: item.y,
            w: item.w,
            h: item.h
          }
        })
      })

      this.allPageLayoutData.forEach(item => {
        const single = find(this.layoutData, {id: item.id})
        if (isEmpty(single)) {
          finalCharts.push(
            {
              id: item.id,
              group: item.group,
              name: item.i,
              displayConfig: item.allGroupDisplayConfig,
              groupDisplayConfig: item.partGroupDisplayConfig
            }
          )
        } else {
          // 解决分组修改后保存失效问题
          const singleChart = find(finalCharts, {id: item.id})
          if (!isEmpty(singleChart)) {
            singleChart.group = item.group
          }
        }
      })
      return {
        id: this.pannelId,
        name: this.panalName,
        charts: finalCharts,
        panelGroups: this.panel_group_list,
        timeRange: this.viewCondition.timeTnterval,
        refreshWeek: this.viewCondition.autoRefresh
      }
    },
    // requestReturnPromise(method, api, params, isNeedloading = true) {
    //   return new Promise(resolve => {
    //     this.request(method, api, params, res => {
    //       resolve(res)
    //     }, { isNeedloading })
    //   })
    // },
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
      await this.request('PUT', this.apiCenter.chartCustomName, {
        chartId: item.id,
        name: item.i
      })
      this.editChartId = null
      this.getPannelList(this.activeGroup)
      this.$Message.success(this.$t('m_save_success'))
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
        this.request('GET', this.apiCenter.chartCustomNameExist, params, res => {
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
        this.request('GET', this.apiCenter.changeChartCustomPermission, {
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
      this.request('GET', this.apiCenter.getUserRoleList, params, res => {
        this.userRolesOptions = this.processRolesList(res.data)
      })
      this.request('GET', this.apiCenter.getUserManageRole, {}, res => {
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
    async saveChartOrDashboardAuth(mgmtRoles, useRoles) {
      const params = {
        mgmtRoles,
        useRoles
      }
      if (this.saveAuthType === 'chart') {
        this.mgmtRoles = mgmtRoles
        this.useRoles = useRoles
        params.chartId = this.setChartConfigId
        await this.request('POST', this.apiCenter.changeChartCustomPermission, params)
      } else {
        this.boardMgmtRoles = mgmtRoles
        this.boardUseRoles = useRoles
        params.id = this.pannelId
        await this.request('POST', this.apiCenter.getDashboardCustomPermission, params)
      }
      this.$Message.success(this.$t('m_success'))
      this.getPannelList()
    },
    onSingleChartGroupClear(item) {
      const singleChart = find(this.allPageLayoutData, {id: item.id})
      if (!isEmpty(singleChart)) {
        singleChart.group = ''
      }
      this.request('PUT', this.apiCenter.dashboardCustom, this.processPannelParams(), () => {
        this.$Message.success(this.$t('m_success'))
        this.getPannelList(this.activeGroup)
      })
    },
    onSingleChartGroupChange(group, index, item) {
      if (isEmpty(group)) {
        return
      }
      const singleChart = find(this.allPageLayoutData, {id: item.id})
      if (!isEmpty(singleChart)) {
        singleChart.group = group
      }
      if (group !== 'ALL') {
        // 切换分组时，默认填充到最后一个，并保证w=6, h=7
        const layoutDataByGroup = this.layoutData.filter((d, m) => m !== index && d.group === group)
        const layoutByGroupNeedReset = layoutDataByGroup.some((item, i) => i !== index && item.partGroupDisplayConfig === '')
        this.layoutData[index].partGroupDisplayConfig = this.layoutData[index].partGroupDisplayConfig || {}
        const needChangeItem = this.layoutData[index].partGroupDisplayConfig
        if (isEmpty(layoutDataByGroup) || layoutByGroupNeedReset) {
          needChangeItem.w = 6
          needChangeItem.h = 7
          needChangeItem.x = 0
          needChangeItem.y = 0
        } else {
          // 先筛选出最下面的元素，再给这个元素填充
          const maxY = maxBy(layoutDataByGroup, item => item.partGroupDisplayConfig.y).partGroupDisplayConfig.y
          const sameMaxYData = filter(layoutDataByGroup, item => item.partGroupDisplayConfig.y === maxY)
          const maxX = maxBy(sameMaxYData, item => item.partGroupDisplayConfig.x).partGroupDisplayConfig.x
          const lastOneArr = filter(sameMaxYData, item => item.partGroupDisplayConfig.x === maxX)
          // 用于获取更改目标组的排列
          const isStandardArrangement = layoutDataByGroup.every((item, i) => i !== index && item.partGroupDisplayConfig.w === lastOneArr[0].partGroupDisplayConfig.w && item.partGroupDisplayConfig.h === lastOneArr[0].partGroupDisplayConfig.h)
          if (isStandardArrangement) {
            needChangeItem.w = lastOneArr[0].partGroupDisplayConfig.w
            needChangeItem.h = lastOneArr[0].partGroupDisplayConfig.h
          } else {
            needChangeItem.w = 6
            needChangeItem.h = 7
          }

          if (!isEmpty(lastOneArr) && lastOneArr.length === 1) {
            const lastOne = lastOneArr[0].partGroupDisplayConfig
            if (lastOne.x + lastOne.w + needChangeItem.w > 12) {
              needChangeItem.y = lastOne.y + 7
              needChangeItem.x = 0
            } else {
              needChangeItem.y = lastOne.y
              needChangeItem.x = lastOne.x + lastOne.w
            }
          }
        }
      }
      this.request('PUT', this.apiCenter.dashboardCustom, this.processPannelParams(), () => {
        this.$Message.success(this.$t('m_success'))
        this.getPannelList(this.activeGroup)
      })
    },
    closeChartInfoDrawer() {
      this.getPannelList()
    },
    exportPanel() {
      this.isModalShow = true
    },
    onLayoutRadioChange(type) {
      if (this.isEditStatus) {
        document.querySelector('#chartLayoutPopTipButton').click()
        this.tempChartLayoutType = type
        this.$nextTick(() => {
          this.chartLayoutType = this.previousChartLayoutType
        })
      } else {
        this.setChartLayoutType(type)
      }
    },

    onLayoutPopTipConfirm() {
      this.setChartLayoutType(this.tempChartLayoutType)
      this.resetHasNotRequestStatus()
      setTimeout(() => {
        this.scrollRefresh = !this.scrollRefresh
      }, 1000)
      setTimeout(() => {
        document.querySelector('.vue-grid-layout').scrollIntoView({
          behavior: 'smooth',
          block: 'start'
        })
      }, 100)
    },
    calculateLayout(data, type='customize') {
      if (isEmpty(data) || type==='customize') {
        return data
      }
      data.forEach((item, index) => {
        item.h = 7
        if (type === 'two') {
          item.w = 6
          item.x = (index % 2) * 6
          item.y = Math.floor(index / 2) * 7
        } else if (type === 'three') {
          item.w = 4
          item.x = (index % 3) * 4
          item.y = Math.floor(index / 3) * 7
        } else if (type === 'four') {
          item.w = 3
          item.x = (index % 4) * 3
          item.y = Math.floor(index / 4) * 7
        } else if (type === 'five') {
          item.w = 2.4
          item.x = (index % 5) * 2.4
          item.y = Math.floor(index / 5) * 7
        } else if (type === 'six') {
          item.w = 2
          item.x = (index % 6) * 2
          item.y = Math.floor(index / 6) * 7
        } else if (type === 'seven') {
          item.w = 1.7
          item.x = (index % 7) * 1.7
          item.y = Math.floor(index / 7) * 7
        } else if (type === 'eight') {
          item.w = 1.5
          item.x = (index % 8) * 1.5
          item.y = Math.floor(index / 8) * 7
        }
      })
      return data
    },
    setChartLayoutType(val = 'customize') {
      this.chartLayoutType = val
      this.previousChartLayoutType = val
      if (isEmpty(this.layoutData) || val === 'customize') {
        return
      }
      this.filterLayoutData()
    },
    filterLayoutData() {
      if (this.activeGroup === 'ALL') {
        this.layoutData = this.calculateLayout(this.layoutData, this.chartLayoutType)
      } else {
        this.layoutData = this.layoutData.filter(d => d.group === this.activeGroup)
        if (isEmpty(this.layoutData)) {
          return []
        }
        const layoutNeedReset = this.layoutData.some(item => item.partGroupDisplayConfig === '')
        // 假如其中有partGroupDisplayConfig为空，基于两列进行打平，假如没有为空，则基于partGroupDisplayConfig排列
        this.initLayoutTypeByWidth(this.layoutData)
        if (layoutNeedReset || ['two', 'three', 'four', 'five', 'six', 'seven', 'eight'].includes(this.chartLayoutType)) {
          this.sortLayoutData(this.layoutData)
          this.calculateLayout(this.layoutData, this.chartLayoutType)
        } else {
          this.layoutData.forEach(item => {
            Object.assign(item, item.partGroupDisplayConfig)
          })
          this.sortLayoutData(this.layoutData)
        }
      }
      this.chartLayoutType = this.confirmLayoutType(this.layoutData)
      this.previousChartLayoutType = this.chartLayoutType
      this.refreshPannelNow()
      return this.layoutData
    },
    refreshPannelNow() {
      this.$nextTick(() => {
        this.refreshNow = !this.refreshNow
      })
    },
    isValidJson(str) {
      try {
        JSON.parse(str)
        return true
      } catch (e) {
        return false
      }
    },
    sortLayoutData(data) {
      const sortedArr = orderBy(data, ['y', 'x'], ['asc', 'asc'])
      return sortedArr
    },
    initLayoutTypeByWidth(data) {
      if (isEmpty(data) || this.chartLayoutType !== 'customize') {
        return
      }
      const isTwo = data.every(item => item.h === 7 && item.w === 6)
      const isThree = data.every(item => item.h === 7 && item.w === 4)
      const isFour = data.every(item => item.h === 7 && item.w === 3)
      const isFive = data.every(item => item.h === 7 && item.w === 2.4)
      const isSix = data.every(item => item.h === 7 && item.w === 2)
      const isSeven = data.every(item => item.h === 7 && item.w === 1.7)
      const isEight = data.every(item => item.h === 7 && item.w === 1.5)

      isTwo ? this.chartLayoutType = 'two'
        : isThree ? this.chartLayoutType = 'three'
          : isFour ? this.chartLayoutType = 'four'
            : isFive ? this.chartLayoutType = 'five'
              : isSix ? this.chartLayoutType = 'six'
                : isSeven ? this.chartLayoutType = 'seven'
                  : isEight ? this.chartLayoutType = 'eight' : 'customize'
    },
    confirmLayoutType(data) {
      if (isEmpty(data)) {
        return 'customize'
      }
      let res = ''
      const isTwo = data.every((item, i) => item.x === (i % 2) * 6 && item.y === Math.floor(i / 2) * 7 && item.h === 7 && item.w === 6)
      const isThree = data.every((item, i) => (item.x === (i % 3) * 4) && item.y === Math.floor(i / 3) * 7 && item.h === 7 && item.w === 4)
      const isFour = data.every((item, i) => (item.x === (i % 4) * 3) && item.y === Math.floor(i / 4) * 7 && item.h === 7 && item.w === 3)
      const isFive = data.every((item, i) => (item.x === (i % 5) * 2.4) && item.y === Math.floor(i / 5) * 7 && item.h === 7 && item.w === 2.4)
      const isSix = data.every((item, i) => (item.x === (i % 6) * 2) && item.y === Math.floor(i / 6) * 7 && item.h === 7 && item.w === 2)
      const isSeven = data.every((item, i) => (item.x === (i % 7) * 1.7) && item.y === Math.floor(i / 7) * 7 && item.h === 7 && item.w === 1.7)
      const isEight = data.every((item, i) => (item.x === (i % 8) * 1.5) && item.y === Math.floor(i / 8) * 7 && item.h === 7 && item.w === 1.5)

      isTwo ? res = 'two'
        : isThree ? res = 'three'
          : isFour ? res = 'four'
            : isFive ? res = 'five'
              : isSix ? res = 'six'
                : isSeven ? res = 'seven'
                  : isEight ? res = 'eight' : res = 'customize'
      return res
    },
    // 在新增时计算DisplayConfig的值
    calculateDisplayConfig() {
      let finalItem = {}
      if (!isEmpty(this.allPageLayoutData)) {
        // 先筛选出最下面的元素，再给这个元素填充
        const maxY = maxBy(this.allPageLayoutData, item => item.allGroupDisplayConfig.y).allGroupDisplayConfig.y
        const sameMaxYData = filter(this.allPageLayoutData, item => item.allGroupDisplayConfig.y === maxY)
        const maxX = maxBy(sameMaxYData, item => item.allGroupDisplayConfig.x).allGroupDisplayConfig.x
        const lastOneArr = filter(sameMaxYData, item => item.allGroupDisplayConfig.x === maxX)
        // 用于获取更改目标组的排列
        const isStandardArrangement = this.allPageLayoutData.every(item => item.allGroupDisplayConfig.w === lastOneArr[0].allGroupDisplayConfig.w && item.allGroupDisplayConfig.h === lastOneArr[0].allGroupDisplayConfig.h)
        if (isStandardArrangement) {
          finalItem.w = lastOneArr[0].allGroupDisplayConfig.w
          finalItem.h = lastOneArr[0].allGroupDisplayConfig.h
        } else {
          finalItem.w = 6
          finalItem.h = 7
        }

        if (!isEmpty(lastOneArr) && lastOneArr.length === 1) {
          const lastOne = lastOneArr[0].allGroupDisplayConfig
          if (lastOne.x + lastOne.w + finalItem.w > 12) {
            finalItem.y = lastOne.y + 7
            finalItem.x = 0
          } else {
            finalItem.y = lastOne.y
            finalItem.x = lastOne.x + lastOne.w
          }
        } else {
          finalItem.x = 0
          finalItem.y = 0
        }
      } else {
        finalItem = {
          x: 0,
          y: 0,
          w: 6,
          h: 7
        }
      }
      return finalItem
    },
    processHtmlText(name = '') {
      if (name.split('/').length > 1) {
        return '<span style=\'font-size: 14px; line-height: 16px; display: inline-block\'>' + name.split('/').join('<br>') + '</span>'
      }
      return name

    },
    onCopyTableSelected(chartList) {
      this.selectedChartList = chartList
    },
    showLineSelectModal(item) {
      this.setChartConfigId = item.id
      this.isLineSelectModalShow = true
    },
    onLineSelectChangeCancel() {
      this.isLineSelectModalShow = false
      this.refreshNow = !this.refreshNow
    },
    onGridWindowScroll: debounce(function () {
      this.scrollRefresh = !this.scrollRefresh
    }, 1000),
    resetHasNotRequestStatus() {
      this.hasNotRequestStatus = !this.hasNotRequestStatus
    },
    onIconRegionClick() {
      this.isActionRegionExpand = !this.isActionRegionExpand
      this.$nextTick(() => {
        this.calculateGridWindowHeight()
      })
    },
    calculateGridWindowHeight() {
      const header = document.querySelector('.view-config-top-content')
      const gridContent = document.querySelector('.grid-window')
      const alarmContent = document.querySelector('.view-config-alarm')
      const alarmListContent = document.querySelector('.alarm-list')
      const headerHeight = header.offsetHeight
      gridContent && (gridContent.style.height = `calc(100vh - ${headerHeight + 100}px)`)
      alarmContent && (alarmContent.style.height = `calc(100vh - ${headerHeight + 100}px)`)
      alarmListContent && (alarmListContent.style.height = `calc(100vh - ${headerHeight + 100 + 120}px)`)
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
    ExportChartModal,
    ChartLinesModal
  },
}
</script>

<style lang="less">
.monitor-custom-view-config {
  .custom-title-text {
    max-width: 375px !important;
  }
}
.chart-layout-poptip {
  max-width: 350px !important;
  top: 75px !important;
  left: 230px !important
}

.chart-config-info {
  .ivu-dropdown-item-disabled {
    color: inherit
  }
}

.filter-chart-layer {
  max-height: 520px !important;
}

.filter-chart-layer::-webkit-scrollbar {
  display: none;
}

.references-button {
  font-size: 14px;
  min-width: 47px;
  background-color: #edf4fe !important;
  border-color: #b5d0fb !important;
}

.grid-content {
  display: flex;
  padding: 0 5px;
  font-size: 16px;
  align-items: flex-start;
  .grid-auto-tag-style {
    margin-top: 4px;
    font-size: 14px;
    min-width: 45px;
  }
}

.header-grid-tools {
  display: flex;
  align-items: center;
  flex-grow: 1;
  i {
    font-size: 18px !important
  }

  .grid-content {
    display: flex;
    padding: 0 5px;
    font-size: 16px;
    align-items: center;
    margin-top: 3px;
  }

  .header-grid-tools {
    display: flex;
    align-items: center;
    flex-grow: 1;
    i {
      font-size: 18px !important
    }
  }

  .header-grid-name {
    .ivu-tooltip-rel {
      display: flex;
      align-items: center;
      flex-grow: 1;
      i {
        font-size: 18px !important
      }
    }
  }
}

.copy-drowdown-slot-select > :nth-child(2) {
  .ivu-select-dropdown {
    max-height: 400px !important;
    overflow: auto !important;
  }
}
</style>

<style scoped lang="less">
/deep/ .ivu-form-item {
  margin-bottom: 0;
}

.all-action-region {
  position: relative;
  // margin-top: 18px;
  // border: 1px solid #cfd0d3;
}

.drop-icon-region {
  margin-left: 5px;
  cursor: pointer;
}

.view-config-top-content {
  max-height: 230px;
  overflow: scroll;
}
.view-config-top-content::-webkit-scrollbar {
  display: none;
}

.arrow-back {
  cursor: pointer;
  width: 28px;
  height: 24px;
  color: #fff;
  border-radius: 2px;
  background: #5384FF;
  margin-right: 12px;
  margin-bottom: 10px;
}
.header-name {
  position: relative;
  font-size: 16px;
  margin-left: 15px;
}
.header-tools {
  position: absolute;
  right: 0px;
  top: 0px;
}

.panal-edit-icon {
  margin-left:4px;
  padding: 4px;
}
.search-container {
  display: flex;
  justify-content: space-between;
  margin: 2px 8px 2px;
  font-size: 16px;
}
.search-zone {
  display: inline-block;
}
.params-title {
  margin-left: 7px;
  font-size: 13px;
}
.header-grid {
  display: flex;
  // flex-grow: 1;
  justify-content: flex-end;
  // line-height: 32px;
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
  display: flex;
  align-items: center;
  font-size: 14px;
  .group-region {
    display: inline-block;
    white-space: nowrap;
    max-width: ~'calc(100vw - 300px)';
    overflow: auto;
  }
}
.chart-config-info {
  font-size: 14px;
  padding-bottom: 10px;
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

.grid-window {
  height: ~"calc(100vh - 300px)";
  overflow: auto;
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
  min-width: 700px;
  height: ~"calc(100vh - 330px)";
  overflow: auto;
}
.view-config-alarm::-webkit-scrollbar {
  display: none;
}
.no-data {
  text-align: center;
  margin: 32px;
  font-size: 14px;
  color: gray;
}

.ellipsis-text {
  width: 350px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: inline-block;
  vertical-align: bottom;
}

.header-grid-name-text {
  display: inline-block;
  max-width: 100%;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  margin-top: 3px
}

.copy-drowdown-slot {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  width: 840px;

  .copy-drowdown-slot-select {
    display: flex;
    justify-content: flex-start;
    margin-bottom: 5px;
    width: 100%
  }
  .copy-drowdown-slot-select > :nth-child(1) {
    margin-left: 20px;
    width: 35%;
  }
  .copy-drowdown-slot-select > :nth-child(2) {
    margin-left: 20px;
    width: 35%;
  }
  .copy-drowdown-confirm-button {
    width: 100px;
    margin-right: 20px;
    margin-bottom: 10px;
    align-self: flex-end;
  }
  .copy-detail-table {
    margin-bottom: 10px;
    max-width: 850px;
  }
  .copy-table-tips {
    margin-left: 20px;
    margin-bottom: 5px;
    color: #535a6c
  }
}
</style>
