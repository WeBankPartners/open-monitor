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
        <div class="title">{{ (['add', 'copy'].includes(operator) ? $t('m_button_add') : $t('m_button_edit')) + $t('m_year_over_year_metrics') }}<span class="underline"></span></div>
        <slot name="sub-title"></slot>
      </div>
      <div class="content" :style="{maxHeight: maxHeight + 'px'}">
        <Form :label-width="110" label-position="left">
          <FormItem :label="$t('m_metric')" v-if="operator === 'edit' || viewOnly" required>
            <Input disabled v-model="metricConfigData.guid"/>
          </FormItem>
          <!--原始指标-->
          <FormItem :label="$t('m_original_metric_key')" required>
            <Select filterable v-model="metricConfigData.metricId" :disabled="operator === 'edit' || viewOnly" :transfer="true" @on-change="getMetricType">
              <Option v-for="item in metricList" :value="item.guid" :key="item.guid">{{ item.metric }}</Option>
            </Select>
          </FormItem>
          <!-- 原始指标类型 -->
          <FormItem :label="$t('m_original_metric_type')">
            <Tag size="medium" type="border" :color=metricTypeConfig.color>{{metricTypeConfig.label || '-'}}</Tag>
          </FormItem>
          <!-- 对比类型 -->
          <FormItem :label="$t('m_comparison_types')" required>
            <RadioGroup v-model="metricConfigData.comparisonType" @on-change="getChartData">
              <Radio label="day" :disabled="operator === 'edit'">
                <span>{{ $t('m_dod_comparison') }}</span>
              </Radio>
              <Radio label="week" :disabled="operator === 'edit'">
                <span>{{ $t('m_wow_comparison') }}</span>
              </Radio>
              <Radio label="month" :disabled="operator === 'edit'">
                <span>{{ $t('m_mom_comparison') }}</span>
              </Radio>
            </RadioGroup>
          </FormItem>
          <!-- 计算数值 -->
          <FormItem :label="$t('m_calc_value')" required>
            <CheckboxGroup v-model="metricConfigData.calcType" @on-change="getChartData">
              <Checkbox label="diff" :disabled="viewOnly">
                <span>{{ $t('m_difference') }}</span>
              </Checkbox>
              <Checkbox label="diff_percent" :disabled="viewOnly">
                <span>{{ $t('m_percentage_difference') }}</span>
              </Checkbox>
            </CheckboxGroup>
          </FormItem>
          <!--计算方法-->
          <FormItem :label="$t('m_calc_method')" required>
            <Select  filterable v-model="metricConfigData.calcMethod" @on-change="getChartData" :disabled="operator === 'edit' || viewOnly" :transfer="true">
              <Option v-for="item in calcMethodOption" :value="item.value" :key="item.value">{{ item.label }}</Option>
            </Select>
          </FormItem>
          <!--计算周期-->
          <FormItem :label="$t('m_calculation_period')" required>
            <Select
              filterable
              :disabled="operator === 'edit' || viewOnly"
              @on-change="getChartData"
              v-model="metricConfigData.calcPeriod"
            >
              <Option
                v-for="item in aggStepOptions"
                :value="item.value"
                :label="item.label"
                :key="item.value"
              >
                {{ item.label }}
              </Option>
            </Select>
          </FormItem>
          <!--预览对象-->
          <FormItem :label="$t('m_preview') + $t('m_endpoint')">
            <Select filterable clearable v-model="metricConfigData.endpoint" @on-change="getChartData">
              <Option v-for="item in endpointOptions" :value="item.guid" :key="item.guid">{{ item.name || item.guid }}</Option>
            </Select>
          </FormItem>
          <!--预览区-->
          <div id="echartId" class="echart"></div>
        </Form>
      </div>
      <div class="drawer-footer">
        <Button style="margin-right: 8px" @click="handleCancel">{{ $t('m_button_cancel') }}</Button>
        <Button type="primary" class="primary" :disabled="metricConfigData.metricId === '' || metricConfigData.calcType.length === 0 || viewOnly" @click="handleSubmit">{{ $t('m_button_save') }}</Button>
      </div>
    </Drawer>
  </div>
</template>

<script>
import { debounce} from '@/assets/js/utils'
import {cloneDeep} from 'lodash'
// import { readyToDraw } from '@/assets/config/chart-rely-yoy'
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
    endpointGroup: { // 组关键参数
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
    originalMetricsId: { // 原始指标中新增同环比传递的指标
      type: String,
      default: ''
    },
    // 仅查看
    viewOnly: {
      type: Boolean,
      default: false
    }
  },
  data() {
    return {
      metricList: [], // 原始指标列表
      metricTypeConfig: {},
      metricConfigData: {
        metricId: '', // 原始指标Id
        comparisonType: 'day', // 对比类型
        calcType: ['diff'], // 计算数值
        calcMethod: 'avg', // 计算方法
        calcPeriod: 60, // 计算周期
        endpoint: ''
      },
      endpoint: '',
      endpointOptions: [],
      echartId: '',
      maxHeight: 500,
      calcMethodOption: [
        {
          label: this.$t('m_average'),
          value: 'avg'
        },
        {
          label: this.$t('m_min'),
          value: 'min'
        },
        {
          label: this.$t('m_max'),
          value: 'max'
        },
        {
          label: this.$t('m_sum'),
          value: 'sum'
        }
      ],
      aggStepOptions: [
        {
          label: '60S',
          value: 60
        },
        {
          label: '300S',
          value: 300
        },
        {
          label: '600S',
          value: 600
        },
        {
          label: '1800S',
          value: 1800
        },
        {
          label: '3600S',
          value: 3600
        }
      ],
      previewObject: {}, // 预览对象，供查看时渲染预览对象值使用
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
  async mounted() {
    this.metricConfigData.metricId = this.originalMetricsId || ''
    await this.getMetricList()
    this.initDrawerHeight()
    if (['edit', 'copy'].includes(this.operator)) {
      this.getConfigData()
    }
    await this.getEndpoint()
  },
  methods: {
    setPreviewObject(obj) {
      this.previewObject = obj
      // this.getEndpoint()
    },
    // 获取原始指标列表
    async getMetricList() {
      let params = {
        monitorType: this.monitorType,
        onlyService: 'Y',
        serviceGroup: this.serviceGroup,
        endpointGroup: this.endpointGroup
      }
      // 对象类型查看使用特殊入参
      if (this.fromPage === 'object') {
        params = {
          guid: this.data.metricId
        }
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', '/monitor/api/v2/monitor/metric/list', params, responseData => {
        this.metricList = responseData
        this.getMetricType(this.metricConfigData.metricId)
      }, {isNeedloading: true})
    },
    // 获取回显数据
    getConfigData() {
      const params = {
        monitorType: this.monitorType,
        onlyService: 'Y',
        serviceGroup: this.serviceGroup,
        guid: this.data.guid,
        endpointGroup: this.endpointGroup
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', '/monitor/api/v2/monitor/metric_comparison/list', params, responseData => {
        if (responseData.length === 1) {
          this.metricConfigData = responseData[0]
          if (this.fromPage === 'object') {
            this.metricConfigData.endpoint = this.previewObject.guid
          }
          this.getMetricType(this.metricConfigData.metricId)
        }
      }, {isNeedloading: true})
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
    // 获取预览对象列表
    getEndpoint() {
      // 对象处查看，直接使用父页面中的对象
      if (this.fromPage === 'object') {
        this.metricConfigData.endpoint = this.previewObject.guid
        this.$nextTick(() => {
          this.$set(this.endpointOptions, 0, this.previewObject)
          this.getChartData()
        })
        return
      }
      const params = {
        type: this.monitorType, // 基础类型
        serviceGroup: this.serviceGroup, // 层级对象
        endpointGroup: this.endpointGroup, // 组
        workspace: 'all_object'
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance(
        'GET',
        this.$root.apiCenter.getEndpoint,
        params,
        responseData => {
          this.endpointOptions = responseData
          if (this.endpointOptions.length > 0) {
            this.metricConfigData.endpoint = this.endpointOptions[0].guid
            this.getChartData()
          }
        }
      )
    },
    // 选择预览对象
    async getMetricType(val) {
      if (!val) {
        return
      }
      const findOriMetric = this.metricList.find(item => item.guid === val)
      if (findOriMetric) {
        const typeList = [
          {
            label: this.$t('m_basic_type'),
            value: 'common',
            color: '#2d8cf0'
          },
          {
            label: this.$t('m_business_configuration'),
            value: 'business',
            color: '#81b337'
          },
          {
            label: this.$t('m_metric_list'),
            value: 'custom',
            color: '#b886f8'
          }
        ]
        this.metricTypeConfig = typeList.find(item => item.value === findOriMetric.metric_type) || {}
      }
      this.getChartData()
    },
    // 渲染echart
    async getChartData() {
      const {
        endpoint, metricId, comparisonType, calcType, calcMethod, calcPeriod
      } = this.metricConfigData
      if ([undefined, ''].includes(endpoint) || calcType.length === 0 || metricId === '') {
        const myChart = echarts.init(document.getElementById('echartId'))
        myChart.clear()
        return
      }
      const params = {
        endpoint,
        metricId,
        comparisonType,
        calcType,
        calcMethod,
        calcPeriod,
        timeSecond: -calcPeriod
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance(
        'POST',
        '/monitor/api/v1/dashboard/comparison_chart',
        params,
        responseData => {
          // const chartConfig = {eye: false,dataZoom:false, lineBarSwitch: true, chartType: 'twoYaxes', params: this.chartInfo.chartParams};
          const chartConfig = {
            clear: true,
            eye: false,
            dataZoom: false,
            lineBarSwitch: true,
            params: {
              lineType: 2
            }
          }
          readyToDraw(this, responseData, 1, chartConfig, 'echartId')
        },
        { isNeedloading: false }
      )
    },
    handleSubmit() {
      if (!this.metricConfigData.metricId) {
        return this.$Message.error(this.$t('m_original_metric_key') + this.$t('m_tips_required'))
      }
      const type = 'POST'
      if (this.operator === 'copy') {
        const keys = ['calcMethod', 'calcPeriod', 'calcType', 'comparisonType', 'endpoint', 'metricId']
        const params = {}
        keys.forEach(key => {
          params[key] = this.metricConfigData[key]
        })
        this.metricConfigData = cloneDeep(params)
      }
      this.metricConfigData.endpoint_group = this.endpointGroup
      this.$root.$httpRequestEntrance.httpRequestEntrance(
        type,
        this.$root.apiCenter.comparisonMetricMgmt,
        this.metricConfigData,
        () => {
          this.$Message.success(this.$t('m_tips_success'))
          this.$emit('update:visible', false)
          this.$emit('fetchList')
        }
      )
    },
    handleCancel() {
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
