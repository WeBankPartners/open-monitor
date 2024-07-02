<template>
  <div>
    <Drawer
      v-model="drawerVisible"
      width="820"
      :mask-closable="false"
      :lock-scroll="true"
      @on-close="handleCancel"
      class="monitor-add-group"
    >
      <div slot="header" class="w-header">
        <div class="title">{{ (operator === 'add' ? $t('m_button_add'):$t('m_button_edit'))+$t('m_original_metric') }}<span class="underline"></span></div>
        <slot name="sub-title"></slot>
      </div>
      <div class="content" :style="{ maxHeight: maxHeight + 'px' }">
        <Form :label-width="100" label-position="left">
          <!--名称-->
          <FormItem :label="$t('m_tableKey_name')" required>
            <Input :disabled="operator === 'edit'" v-model="metricConfigData.metric"></Input>
          </FormItem>
          <!--作用域-->
          <FormItem :label="$t('m_scope')" required>
            <Select v-model="workspace" :disabled="metricConfigData.metric_type === 'business'" @on-change="changeWorkspace">
              <Option v-if="serviceGroup" value="all_object">{{ $t('m_all_object') }}</Option>
              <Option value="any_object">{{ $t('m_any_object') }}</Option>
            </Select>
          </FormItem>
          <!--推荐配置-->
          <FormItem :label="$t('m_recommend')">
            <Select v-model="templatePl" :disabled="metricConfigData.metric_type === 'business'" clearable @on-clear="clearTemplatePl" @on-change="changeTemplatePl">
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
                :disabled="metricConfigData.metric_type === 'business'"
                filterable
                :key="param.label"
                :placeholder="param.label"
                class="select-dropdown"
              >
                <Option 
                  style="white-space: normal;"
                  v-for="(item, itemIndex) in collectedMetricOptions" 
                  :value="item.option_value" 
                  :key="item.option_value + itemIndex">
                  {{ item.option_text }}
                </Option>
              </Select>
            </template>
          </FormItem>
          <!--表达式-->
          <FormItem :label="$t('m_field_metric')" required>
            <Input v-model="metricConfigData.prom_expr" :disabled="metricConfigData.metric_type === 'business'" type="textarea" :rows="5" style="margin:5px 0;" />
          </FormItem>
          <!--预览对象-->
          <FormItem :label="$t('m_preview') + $t('m_endpoint')">
            <Select filterable clearable v-model="metricConfigData.endpoint" @on-change="handleEndPointChange">
              <Option v-for="item in endpointOptions" :value="item.guid" :key="item.guid">{{ item.guid }}</Option>
            </Select>
          </FormItem>
          <!--预览区-->
          <div :id="echartId" class="echart"></div>
        </Form>
      </div>
      <div class="drawer-footer">
        <Button style="margin-right: 8px" @click="handleCancel">{{ $t('m_button_cancel') }}</Button>
        <Button type="primary" class="primary" @click="handleSubmit">{{ $t('m_button_save') }}</Button>
      </div>
    </Drawer>
  </div>
</template>

<script>
import { debounce , generateUuid} from '@/assets/js/utils'
import { readyToDraw } from '@/assets/config/chart-rely'
import * as echarts from 'echarts';
export default {
  props: {
    visible: {
      type: Boolean,
      default: false
    },
    monitorType: {
      type: String,
      default: ''
    },
    serviceGroup: {
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
    endpoint_group:{
      type: String,
      default: ''
    },
  },
  data () {
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
        endpoint_group: '', // 对象组
        panel_id: null,
        prom_expr: '', // 表达式
        endpoint: '' // 对象
      },
      endpoint: '',
      endpointOptions: [],
      echartId: '',
      maxHeight: 500
    }
  },
  computed: {
    drawerVisible: {
      get () {
        return this.visible
      },
      set (val) {
        this.$emit('update:visible', val)
      }
    }
  },
  watch: {
    workspace (val) {
      if (val) {
        this.getMetricTemplate()
      }
    }
  },
  mounted () {
    this.getEndpoint()
    this.initDrawerHeight()
    if (this.operator === 'edit') {
      this.workspace = this.data.workspace
      this.metricConfigData = {
        ...this.data
      }
    }
  },
  methods: {
    initDrawerHeight () {
      this.maxHeight = document.body.clientHeight - 150
      window.addEventListener(
        'resize',
        debounce(() => {
          this.maxHeight = document.body.clientHeight - 150
        }, 100)
      )
    },
    // 选择作用域
    changeWorkspace () {
      this.templatePl = ''
      this.metricTemplateParams = []
      this.metricConfigData.prom_expr = ''
    },
    // 选择推荐配置
    changeTemplatePl (val) {
      const find = this.metricTemplate.find(item => item.prom_expr === val)
      this.metricTemplateParams = find && find.param.split(',').map(p => {
        return {
          value: '',
          label: p
        }
      })
      if (find && find.name === 'custom') {
        // this.metricConfigData.prom_expr += val
      } else {
        this.metricConfigData.prom_expr = val
      }
    },
    // 清空推荐配置
    clearTemplatePl () {
      this.templatePl = ''
      this.metricConfigData.endpoint = ''
      this.metricTemplateParams = []
    },
    // 获取推荐配置列表
    getMetricTemplate () {
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
    changeCollectedMetric (val) {
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
    getCollectedMetric () {
      const params = {
        monitor_type: this.monitorType,
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
    getEndpoint () {
      const params = {
        type: this.monitorType,
        serviceGroup: this.serviceGroup
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance(
        'GET',
        this.$root.apiCenter.getEndpoint,
        params,
        responseData => {
          this.endpointOptions = responseData
        })
    },
    // 选择预览对象
    handleEndPointChange (val) {
      if (this.echartId) {
        const myChart = echarts.init(document.getElementById(this.echartId))
        myChart.clear()
      }
      if (!val) return
      this.getChartData()
      // if (this.workspace === 'all_object') {
      //   this.getChartData()
      // } else {
      //   this.getEndpoint()
      // }
    },
    // 渲染echart
    async getChartData () {
      generateUuid().then((elId)=>{
        this.echartId = `id_${elId}`
      })
      let params = {
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
        endpoint_type: this.monitorType,
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
            clear: true
          }
          readyToDraw(this, responseData, 1, chartConfig, this.echartId)
        },
        { isNeedloading: false }
      )
    },
    handleSubmit () {
      if (!this.metricConfigData.metric) {
        return this.$Message.error(this.$t('m_tableKey_name') + this.$t('m_tips_required'))
      }
      if (!this.workspace) {
        return this.$Message.error(this.$t('m_scope') + this.$t('m_tips_required'))
      }
      if (!this.metricConfigData.prom_expr) {
        return this.$Message.error(this.$t('m_field_metric') + this.$t('m_tips_required'))
      }
      const type = !this.metricConfigData.guid ? 'POST' : 'PUT'
      this.metricConfigData.monitor_type = this.monitorType
      this.metricConfigData.endpoint_group = this.endpoint_group
      
      this.metricConfigData.service_group = this.serviceGroup
      this.metricConfigData.workspace = this.workspace
      this.$root.$httpRequestEntrance.httpRequestEntrance(
        type,
        this.$root.apiCenter.metricManagement,
        [this.metricConfigData],
        () => {
          this.$Message.success(this.$t('m_tips_success'))
          this.$emit('update:visible', false)
          this.$emit('fetchList')
        })
    },
    handleCancel () {
      this.$emit('update:visible', false)
    }
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