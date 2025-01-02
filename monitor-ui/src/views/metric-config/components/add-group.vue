<template>
  <div>
    <Drawer
      v-model="drawerVisible"
      width="60%"
      :mask-closable="false"
      :lock-scroll="true"
      @on-close="handleCancel"
      class="monitor-add-group"
    >
      <div slot="header" class="w-header">
        <div class="title">{{ (['add', 'copy'].includes(operator) ? $t('m_button_add') : $t('m_button_edit')) + $t('m_original_metric') }}<span class="underline"></span></div>
        <slot name="sub-title"></slot>
      </div>
      <div class="content" :style="{maxHeight: maxHeight + 'px'}">
        <Form :label-width="100" label-position="left">
          <!--名称-->
          <FormItem :label="$t('m_metric_key')" required>
            <Input
              :disabled="operator === 'edit' || viewOnly"
              :maxlength='40'
              show-word-limit
              v-model="metricConfigData.metric"
              :placeholder="$t('m_metric_key_placeholder_second')"
            ></Input>
          </FormItem>
          <!--作用域-->
          <FormItem :label="$t('m_scope')" required>
            <Select v-model="workspace" :disabled="operator !== 'copy' && (metricConfigData.metric_type === 'business' || viewOnly)" @on-change="changeWorkspace">
              <Option v-if="['level', 'objectGroup'].includes(fromPage)" value="all_object">{{ $t('m_all_object') }}</Option>
              <Option value="any_object">{{ $t('m_any_object') }}</Option>
            </Select>
          </FormItem>
          <!--对象类型-->
          <FormItem v-if="fromPage === 'level'" :label="$t('m_tableKey_endpoint_type')" required>
            <Select filterable clearable :disabled="operator === 'edit' || viewOnly" v-model="metricConfigData.endpoint_type" style="width:100%">
              <Option v-for="item in monitorTypeOptions" :value="item" :key="item">{{ item }}</Option>
            </Select>
          </FormItem>
          <!--推荐配置-->
          <FormItem :label="$t('m_recommend')">
            <Select  filterable clearable v-model="templatePl" :disabled="operator !== 'copy' && (metricConfigData.metric_type === 'business' || viewOnly)" @on-clear="clearTemplatePl" @on-change="changeTemplatePl">
              <Option v-for="item in metricTemplate" :value="item.prom_expr" :key="item.prom_expr">{{ item.name }}</Option>
            </Select>
          </FormItem>
          <!--采集数据-->
          <FormItem v-if="templatePl" :label="$t('m_collected_data')">
            {{templatePl}}
            <template v-for="param in metricTemplateParams">
              <Select
                v-model="param.value"
                @on-open-change="getCollectedMetric"
                @on-change="changeCollectedMetric(param)"
                :disabled="metricConfigData.metric_type === 'business' || viewOnly"
                filterable
                clearable
                :key="param.label"
                :placeholder="param.label"
                class="select-dropdown"
              >
                <Option
                  style="white-space: normal;"
                  v-for="(item, itemIndex) in collectedMetricOptions"
                  :value="item.option_value"
                  :key="item.option_value + itemIndex"
                >
                  {{ item.option_text }}
                </Option>
              </Select>
            </template>
          </FormItem>
          <!--表达式-->
          <FormItem :label="$t('m_field_metric')" required>
            <Input v-model="metricConfigData.prom_expr" :disabled="operator !== 'copy' && (metricConfigData.metric_type === 'business' || viewOnly)" type="textarea" :rows="5" style="margin:5px 0;" />
          </FormItem>
          <!--预览对象-->
          <FormItem :label="$t('m_preview') + ' ' + $t('m_endpoint')">
            <Select
              filterable
              clearable
              v-model="metricConfigData.endpoint"
              @on-change="handleEndPointChange"
            >
              <Option v-for="item in endpointOptions" :value="item.guid" :key="item.guid">{{ item.name || item.guid }}</Option>
            </Select>
          </FormItem>
          <!--预览区-->
          <div :id="echartId" class="echart"></div>
        </Form>
      </div>
      <div class="drawer-footer">
        <Button style="margin-right: 8px" @click="handleCancel">{{ $t('m_button_cancel') }}</Button>
        <Button type="primary" class="primary" :disabled="viewOnly" @click="handleSubmit">{{ $t('m_button_save') }}</Button>
      </div>
    </Drawer>
    <ChartLinesModal
      :isLineSelectModalShow="isLineSelectModalShow"
      :chartId="setChartConfigId"
      @modalClose="onLineSelectChangeCancel"
    >
    </ChartLinesModal>
  </div>
</template>

<script>
import {isEmpty, cloneDeep} from 'lodash'
import ChartLinesModal from '@/components/chart-lines-modal'
import { debounce , generateUuid} from '@/assets/js/utils'
import { readyToDraw } from '@/assets/config/chart-rely'
import * as echarts from 'echarts'
export default {
  props: {
    fromPage: {
      type: String,
      default: ''
    },
    visible: {
      type: Boolean,
      default: false
    },
    monitorType: { // 基础类型关键参数，其他类型传process
      type: String,
      default: ''
    },
    serviceGroup: { // 层级对象关键参数
      type: String,
      default: ''
    },
    data: {
      type: Object,
      default: () => {}
    },
    // add添加，edit编辑
    operator: {
      type: String,
      default: 'add'
    },
    // 仅查看
    viewOnly: {
      type: Boolean,
      default: false
    },
    endpoint_group: { // 组关键参数
      type: String,
      default: ''
    }
  },
  data() {
    return {
      workspace: '', // 作用域
      templatePl: '', // 推荐配置
      metricTemplate: [], // 推荐配置下拉列表
      metricTemplateParams: [],
      collectedMetricOptions: [], // 采集数据下拉列表
      metricConfigData: {
        guid: null,
        metric: '', // 名称
        monitor_type: '', // 类型
        endpoint_type: 'process', // 在层级对象页面需要对象类型
        endpoint_group: '', // 对象组
        panel_id: null,
        prom_expr: '', // 表达式
        endpoint: '' // 对象
      },
      endpoint: '',
      endpointOptions: [],
      echartId: '',
      maxHeight: 500,
      monitorTypeOptions: [],
      previewObject: {}, // 预览对象，供查看时渲染预览对象值使用
      isLineSelectModalShow: false,
      setChartConfigId: '',
      chartInstance: null
    }
  },
  computed: {
    drawerVisible: {
      get() {
        return this.visible
      },
      set(val) {
        this.$emit('update:visible', val)
      }
    }
  },
  watch: {
    workspace(val) {
      if (val) {
        this.getMetricTemplate()
      }
    }
  },
  mounted() {
    this.getEndpointType()
    this.initDrawerHeight()
    if (this.operator === 'edit') {
      this.workspace = this.data.workspace
      this.metricConfigData = {
        ...this.data
      }
      this.metricConfigData.endpoint_type = this.metricConfigData.monitor_type
    }
    if (this.operator === 'copy') {
      this.workspace = this.data.workspace
      this.metricConfigData = {
        ...this.data
      }
      this.metricConfigData.endpoint_type = this.metricConfigData.monitor_type
      this.metricConfigData.metric += '1'
    }
    this.getEndpoint()
    generateUuid().then(elId => {
      this.echartId = `id_${elId}`
    })
    window['view-config-selected-line-data'] = {}
    this.$on('editShowLines', this.handleEditShowLines)
  },
  methods: {
    setPreviewObject(obj) {
      this.previewObject = obj
    },
    getEndpointType() {
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.getEndpointType, '', responseData => {
        this.monitorTypeOptions = responseData
      }, {isNeedloading: false})
    },
    initDrawerHeight() {
      this.maxHeight = document.body.clientHeight - 150
      window.addEventListener(
        'resize',
        debounce(() => {
          this.maxHeight = document.body.clientHeight - 150
        }, 100)
      )
    },
    // 选择作用域
    changeWorkspace() {
      // this.templatePl = ''
      // this.metricTemplateParams = []
      // this.metricConfigData.prom_expr = ''
      this.getEndpoint()
    },
    // 选择推荐配置
    changeTemplatePl(val) {
      const find = this.metricTemplate.find(item => item.prom_expr === val)
      this.metricTemplateParams = find && find.param.split(',').map(p => ({
        value: '',
        label: p
      }))
      if (find && find.name === 'custom') {
        // this.metricConfigData.prom_expr += val
      } else {
        this.metricConfigData.prom_expr = val
      }
    },
    // 清空推荐配置
    clearTemplatePl() {
      this.templatePl = ''
      this.metricConfigData.endpoint = ''
      this.metricTemplateParams = []
    },
    // 获取推荐配置列表
    getMetricTemplate() {
      const params = {
        workspace: this.workspace
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance(
        'GET',
        '/monitor/api/v2/sys/parameter/metric_template',
        params,
        res => {
          this.metricTemplate = res
        },
        { isNeedloading: false }
      )
    },
    // 选择采集数据
    changeCollectedMetric(val) {
      if (val.value) {
        const find = this.metricTemplate.find(m => m.prom_expr === this.templatePl)
        if (find && find.name === 'custom') {
          this.metricConfigData.prom_expr += val.value
        }
        this.metricConfigData.prom_expr = this.metricConfigData.prom_expr.replaceAll('undefined', '')
        this.metricConfigData.prom_expr = this.metricConfigData.prom_expr.replaceAll(val.label, val.value)
      }
    },
    // 获取采集数据列表
    getCollectedMetric() {
      const params = {
        // 在层级对象页面需要使用页面中选择的对象类型
        monitor_type: this.fromPage === 'level' ? this.metricConfigData.endpoint_type: this.monitorType,
        guid: this.endpoint,
        service_group: this.serviceGroup,
        workspace: this.workspace
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance(
        'POST',
        this.$root.apiCenter.getMetricOptions,
        params,
        responseData => {
          this.collectedMetricOptions = responseData
        },
        { isNeedloading: false }
      )
    },
    // 获取预览对象列表
    getEndpoint() {
      // 对象处查看，直接使用父页面中的对象
      if (this.fromPage === 'object') {
        this.$nextTick(() => {
          this.endpointOptions = [this.previewObject]
          this.metricConfigData.endpoint = this.previewObject.guid
          this.getChartData()
          return
        })
      }
      const params = {
        // 在层级对象页面需要使用页面中选择的对象类型
        type: this.fromPage === 'level' ? this.metricConfigData.endpoint_type: this.monitorType,
        serviceGroup: this.serviceGroup,
        endpointGroup: this.endpoint_group,
        workspace: this.workspace
      }
      this.metricConfigData.endpoint = ''
      this.$root.$httpRequestEntrance.httpRequestEntrance(
        'GET',
        this.$root.apiCenter.getEndpoint,
        params,
        responseData => {
          if (Array.isArray(responseData) && !isEmpty(responseData)) {
            this.endpointOptions = [...this.endpointOptions, ...responseData]
          }
          if (this.endpointOptions.length > 0) {
            this.metricConfigData.endpoint = this.endpointOptions[0].guid
            this.getChartData()
          }
        }
      )
    },
    // 选择预览对象
    handleEndPointChange(val) {
      if (this.echartId) {
        const myChart = echarts.init(document.getElementById(this.echartId))
        myChart.clear()
      }
      if (!val) {
        return
      }
      this.getChartData()
    },
    // 渲染echart
    async getChartData() {
      const params = {
        calc_service_group_enable: true,
        aggregate: 'none',
        time_second: -1800,
        start: 0,
        end: 0,
        title: '',
        unit: '',
        data: []
      }
      const find = this.endpointOptions.find(e => e.guid === this.metricConfigData.endpoint)
      params.data = [{
        endpoint: (find && find.guid) || '',
        app_object: this.serviceGroup,
        // 在层级对象页面需要使用页面中选择的对象类型
        endpoint_type: this.fromPage === 'level' ? this.metricConfigData.endpoint_type: this.monitorType,
        prom_ql: this.metricConfigData.prom_expr,
        metric: this.metricConfigData.prom_expr === '' ? this.metricConfigData.metric : ''
      }]
      this.$root.$httpRequestEntrance.httpRequestEntrance(
        'POST',
        this.$root.apiCenter.metricConfigView.api,
        params,
        responseData => {
          const chartConfig = {
            eye: false,
            clear: true,
            lineBarSwitch: true,
            canEditShowLines: true,
            dataZoom: false,
            chartId: this.echartId
          }
          responseData.chartId = this.echartId
          this.chartInstance = readyToDraw(this, responseData, 1, chartConfig, this.echartId)
          if (this.chartInstance) {
            this.chartInstance.on('legendselectchanged', params => {
              window['view-config-selected-line-data'][this.echartId] = cloneDeep(params.selected)
            })
          }
        },
        { isNeedloading: false }
      )
    },
    handleSubmit() {
      if (!(/^[A-Za-z][A-Za-z0-9_]{0,48}[A-Za-z0-9]$/.test(this.metricConfigData.metric))) {
        return this.$Message.error(this.$t('m_metric_key') + ':' + this.$t('m_regularization_check_failed_tips'))
      }
      if (!this.workspace) {
        return this.$Message.error(this.$t('m_scope') + this.$t('m_tips_required'))
      }
      if (!this.metricConfigData.prom_expr) {
        return this.$Message.error(this.$t('m_field_metric') + this.$t('m_tips_required'))
      }
      const type = ['add', 'copy'].includes(this.operator) ? 'POST' : 'PUT'
      // 在层级对象页面需要使用页面中选择的对象类型
      this.metricConfigData.monitor_type = this.fromPage === 'level' ? this.metricConfigData.endpoint_type: this.monitorType
      this.metricConfigData.endpoint_group = this.endpoint_group

      this.metricConfigData.service_group = this.serviceGroup
      this.metricConfigData.workspace = this.workspace
      if (this.operator === 'copy') {
        delete this.metricConfigData.guid
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance(
        type,
        this.$root.apiCenter.metricManagement,
        [this.metricConfigData],
        () => {
          this.$Message.success(this.$t('m_tips_success'))
          this.$emit('update:visible', false)
          this.$emit('fetchList')
        }
      )
    },
    handleCancel() {
      this.$emit('update:visible', false)
    },
    handleEditShowLines(config) {
      this.setChartConfigId = config.chartId
      if (isEmpty(window['view-config-selected-line-data'][this.setChartConfigId])) {
        window['view-config-selected-line-data'][this.setChartConfigId] = {}
        config.legend.forEach(one => {
          window['view-config-selected-line-data'][this.setChartConfigId][one] = true
        })
      }
      this.isLineSelectModalShow = true
    },
    onLineSelectChangeCancel() {
      this.isLineSelectModalShow = false
      this.handleEndPointChange(this.metricConfigData.endpoint)
    },
  },
  components: {
    ChartLinesModal
  }
}
</script>

<style lang="less" scoped>
.monitor-add-group {
  .w-header {
    display: flex;
    align-items: center;
    .title {
      font-size: 16px;
      font-weight: bold;
      color: #515a6e;
      margin: 0 6px;
      width: fit-content;
      .underline {
        display: block;
        margin-top: -12px;
        margin-left: -8px;
        width: 100%;
        padding: 0 8px;
        height: 12px;
        border-radius: 12px;
        background-color: #c6eafe;
        box-sizing: content-box;
      }
    }
  }
  .content {
    padding: 8px;
    overflow-y: auto;
    .echart {
      width: 100%;
      height: 280px;
    }
  }
  .drawer-footer {
    width: 100%;
    position: absolute;
    bottom: 0;
    left: 0;
    border-top: 1px solid #e8e8e8;
    padding: 10px 16px;
    text-align: center;
    background: #fff;
  }
  /deep/.ivu-form-item {
    margin-bottom: 8px;
  }
}
</style>
