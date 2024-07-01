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
        <div class="title">{{ (operator === 'add' ? $t('m_button_add'):$t('m_button_edit'))+$t('m_year_over_year_metrics') }}<span class="underline"></span></div>
        <slot name="sub-title"></slot>
      </div>
      <div class="content" :style="{ maxHeight: maxHeight + 'px' }">
        <Form :label-width="100" label-position="left">
          <!--原始指标-->
          <FormItem :label="$t('m_original_metric')">
            <Select filterable v-model="metricConfigData.metricId" :disabled="operator === 'edit'" :transfer="true" @on-change="getMetricType">
              <Option v-for="item in metricList" :value="item.guid" :key="item.guid">{{ item.metric }}</Option>
            </Select>
          </FormItem>
          <!-- 原始指标类型 -->
          <FormItem :label="$t('m_original_metric_type')">
            <Tag size="medium" type="border" :color=metricTypeConfig.color>{{metricTypeConfig.label || '-'}}</Tag>
          </FormItem>
          <!-- 对比类型 -->
          <FormItem :label="$t('m_comparison_types')">
            <RadioGroup v-model="metricConfigData.comparisonType">
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
          <FormItem :label="$t('m_calc_value')">
            <CheckboxGroup v-model="metricConfigData.calcType">
              <Checkbox label="diff">
                <span>{{ $t('m_difference') }}</span>
              </Checkbox>
              <Checkbox label="diff_percent">
                <span>{{ $t('m_percentage_difference') }}</span>
              </Checkbox>
            </CheckboxGroup>
          </FormItem>
          <!--计算方法-->
          <FormItem :label="$t('m_calc_method')">
            <Select filterable v-model="metricConfigData.calcMethod" :disabled="operator === 'edit'" :transfer="true">
              <Option v-for="item in calcMethodOption" :value="item.value" :key="item.value">{{ item.label }}</Option>
            </Select>
          </FormItem>
          <!--计算周期-->
          <FormItem :label="$t('m_calculation_period')">
            <Select 
              filterable
              :disabled="operator === 'edit'"
              v-model="metricConfigData.calcPeriod">
              <Option 
                v-for="item in aggStepOptions" 
                :value="item.value" 
                :label="item.label" 
                :key="item.value">
                {{ item.label }}
              </Option>
            </Select>
          </FormItem>
          <!--预览对象-->
          <!-- <FormItem :label="$t('m_preview') + $t('m_endpoint')">
            <Select filterable clearable v-model="metricConfigData.endpoint" @on-change="handleEndPointChange">
              <Option v-for="item in endpointOptions" :value="item.guid" :key="item.guid">{{ item.guid }}</Option>
            </Select>
          </FormItem> -->
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
    originalMetricsId: { // 原始指标中新增同环比传递的指标
      type: String,
      default: ''
    }
  },
  data () {
    return {
      metricList: [], // 原始指标列表
      metricTypeConfig: {},
      metricConfigData: {
        metricId: 'cpu.detail.percent2__host', // 原始指标Id
        comparisonType: 'day', // 对比类型
        calcType: ['diff'], // 计算数值
        calcMethod: 'avg', // 计算方法
        calcPeriod: 60, // 计算周期
      },
      endpoint: '',
      endpointOptions: [],
      echartId: '',
      maxHeight: 500,
      calcMethodOption: [
        {label: this.$t('m_average'), value: 'avg'},
        {label: this.$t('m_min'), value: 'min'},
        {label: this.$t('m_max'), value: 'max'},
        {label: this.$t('m_sum'), value: 'sum'}
      ],
      aggStepOptions: [
        {label: '60S', value: 60},
        {label: '300S', value: 300},
        {label: '600S', value: 600},
        {label: '1800S', value: 1800},
        {label: '3600S', value: 3600}
      ]
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
  async mounted () {
    this.metricConfigData.metricId = this.originalMetricsId
    await this.getEndpoint()
    await this.getMetricList()
    this.initDrawerHeight()
    if (this.operator === 'edit') {
      this.getConfigData()
    }
  },
  methods: {
    // 获取原始指标列表
    async getMetricList () {
      const params = {
        monitorType: this.monitorType,
        onlyService: 'Y',
        serviceGroup: this.serviceGroup
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', '/monitor/api/v2/monitor/metric/list', params, responseData => {
        this.metricList = responseData
        this.getMetricType(this.metricConfigData.metricId)
      }, {isNeedloading: true})
    },
    // 获取回显数据
    getConfigData () {
      const params = {
        monitorType: this.monitorType,
        onlyService: 'Y',
        serviceGroup: this.serviceGroup,
        guid: this.data.metricId
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', '/monitor/api/v2/monitor/metric_comparison/list', params, responseData => {
        if (responseData.length === 1) {
          this.metricConfigData = responseData[0]
        }
      }, {isNeedloading: true})
    },
    initDrawerHeight () {
      this.maxHeight = document.body.clientHeight - 150
      window.addEventListener(
        'resize',
        debounce(() => {
          this.maxHeight = document.body.clientHeight - 150
        }, 100)
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
    async getMetricType (val) {
      if (!val) return
      const findOriMetric = this.metricList.find(item => item.guid === val)
      const typeList = [
        { label: this.$t('m_basic_type'), value: 'common', color: '#2d8cf0' },
        { label: this.$t('m_business_configuration'), value: 'business', color: '#81b337' },
        { label: this.$t('m_metric_list'), value: 'custom', color: '#b886f8' }
      ]
      this.metricTypeConfig = typeList.find(item => item.value === findOriMetric.metric_type) || {}
      // this.getChartData()
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
      if (!this.metricConfigData.metricId) {
        return this.$Message.error(this.$t('m_original_metric') + this.$t('m_tips_required'))
      }
      const type = !this.metricConfigData.guid ? 'POST' : 'PUT'
      this.$root.$httpRequestEntrance.httpRequestEntrance(
        type,
        this.$root.apiCenter.comparisonMetricMgmt,
        this.metricConfigData,
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