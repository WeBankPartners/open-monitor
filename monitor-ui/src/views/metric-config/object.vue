<template>
  <div ref="maxheight" class="monitor-level-group">
    <Row>
      <Col :span="8">
        <!--对象-->
        <span style="font-size: 14px;">
          {{$t('m_endpoint')}}:
        </span>
        <Select
          style="width:300px;"
          v-model="endpoint"
          filterable 
          @on-change="changeEndpointGroup"
        >
          <Option v-for="(option, index) in endpointOptions" :value="option.guid" :label="option.guid" :key="index">
            <TagShow :list="endpointOptions" name="type" :tagName="option.type" :index="index"></TagShow> 
            {{option.guid}}
          </Option>
        </Select>
      </Col>
      <Col :span="16">
        <div class="btn-group">
          <MetricChange ref="metricChangeRef" @reloadData="reloadData" ></MetricChange>
        </div>
      </Col>
    </Row>
    <Table size="small" :columns="tableColumns.filter(col=>col.showType.includes(metricType))" :data="tableData" class="level-table" />
    <AddGroupDrawer
      v-if="addVisible && metricType==='originalMetrics'"
      :visible.sync="addVisible"
      :monitorType="monitorType"
      :endpoint_group="endpoint"
      :serviceGroup="serviceGroup"
      :data="row"
      operator="edit"
    ></AddGroupDrawer>
    <YearOverYear
      ref="yearOverYearRef"
      v-if="addVisible && metricType==='comparisonMetrics'"
      :visible.sync="addVisible"
      :monitorType="monitorType"
      :originalMetricsId="originalMetricsId"
      :serviceGroup="serviceGroup"
      :data="row"
      operator="edit"
      :endpointGroup="endpoint"
    ></YearOverYear>
  </div>
</template>

<script>
import axios from 'axios'
import {baseURL_config} from '@/assets/js/baseURL'
import { getToken, getPlatFormToken } from '@/assets/js/cookies.ts'
import TagShow from '@/components/Tag-show.vue'
import AddGroupDrawer from './components/add-group.vue'
import YearOverYear from './components/year-over-year.vue'
import MetricChange from './components/metric-change.vue'
export default {
  components: {
    TagShow,
    MetricChange,
    AddGroupDrawer,
    YearOverYear
  },
  data () {
    return {
      metricType: 'originalMetrics', // 原始指标originalMetrics、同环比指标comparisonMetrics
      token: null,
      monitorType: '',
      serviceGroup: '',
      endpoint: '',
      endpointOptions: [],
      maxHeight: 500,
      tableData: [],
      tableColumns: [
      {
          title: this.$t('m_year_over_year_metrics'), // 指标
          key: 'guid',
          minWidth: 250,
          showType: ['comparisonMetrics']
        },
        {
          title: this.$t('m_metric'), // 指标
          key: 'metric',
          minWidth: 200,
          showType: ['originalMetrics', 'comparisonMetrics']
        },
        {
          title: this.$t('m_scope'), // 作用域
          key: 'workspace',
          width: 150,
          render: (h, params) => {
            return <Tag size="medium">{ this.workspaceMap[params.row.workspace] }</Tag>
          },
          showType: ['originalMetrics', 'comparisonMetrics']
        },
        {
          title: this.$t('m_configuration_page'), // 配置页面
          key: 'metric_type',
          width: 140,
          render: (h, params) => {
            const typeList = [
              { label: this.$t('m_basic_type'), value: 'common', color: '#2d8cf0' },
              { label: this.$t('m_business_configuration'), value: 'business', color: '#81b337' },
              { label: this.$t('m_metric_list'), value: 'custom', color: '#b886f8' }
            ]
            const find = typeList.find(item => item.value === params.row.metric_type) || {}
            return <Tag color={find.color} type="border" size="medium">{find.label || '-'}</Tag>
          },
          showType: ['originalMetrics', 'comparisonMetrics']
        },
        {
          title: this.$t('m_calc_method'), // 计算方法
          key: 'calcMethod',
          width: 100,
          render: (h, params) => {
            const calcMethodToDisplay = {
              'avg': this.$t('m_average'),
              'min': this.$t('m_min'),
              'max': this.$t('m_max'),
              'sum': this.$t('m_sum'),
              '-': '-'
            }
            return <span>{calcMethodToDisplay[params.row.calcMethod || '-']}</span>
          },
          showType: ['comparisonMetrics']
        },
        {
          title: this.$t('m_group_type'), // 组类型
          key: 'group_type',
          width: 90,
          render: (h, params) => {
            const groupTypeToDisplay = {
              'level': this.$t('m_field_resourceLevel'),
              'object': this.$t('m_object_group'),
              'system': this.$t('m_basic_type'),
              '-': '-'
            }
            return <span>{groupTypeToDisplay[params.row.group_type || '-']}</span>
          },
          showType: ['comparisonMetrics']
        },
        {
          title: this.$t('m_group_name_'), // 组名
          key: 'group_type',
          width: 80,
          render: (h, params) => {
            return <span>{params.row.group_name || '-'}</span>
          },
          showType: ['comparisonMetrics']
        },
        {
          title: this.$t('m_tableKey_expr'), // 表达式
          key: 'prom_expr',
          minWidth: 250,
          render: (h, params) => {
            return (
              <Tooltip max-width="300" content={params.row.prom_expr} transfer>
                <span class="eclipse">{params.row.prom_expr || '-'}</span>
              </Tooltip>
            )
          },
          showType: ['originalMetrics']
        },
        {
          title: this.$t('m_comparison_types'), // 对比类型
          key: 'comparisonType',
          width: 160,
          render: (h, params) => {
            const comparisonTypeToDisplay = {
              'day': this.$t('m_dod_comparison'),
              'week': this.$t('m_wow_comparison'),
              'month': this.$t('m_mom_comparison'),
              '-': '-'
            }
            return <span>{comparisonTypeToDisplay[params.row.comparisonType || '-']}</span>
          },
          showType: ['comparisonMetrics']
        },
        {
          title: this.$t('m_calc_value'), // 计算数值
          key: 'calcType',
          width: 160,
          render: (h, params) => {
            const calcTypeToDisplay = {
              'diff': this.$t('m_difference'),
              'diff_percent': this.$t('m_percentage_difference')
            }
            const calcTypeCache = (params.row.calcType || []).map(t => calcTypeToDisplay[t]).join('，')
            return <span>{calcTypeCache || '-'}</span>
          },
          showType: ['comparisonMetrics']
        },
        {
          title: this.$t('m_calc_method'), // 计算方法
          key: 'calcMethod',
          width: 100,
          render: (h, params) => {
            const calcMethodToDisplay = {
              'avg': this.$t('m_average'),
              'min': this.$t('m_min'),
              'max': this.$t('m_max'),
              'sum': this.$t('m_sum'),
              '-': '-'
            }
            return <span>{calcMethodToDisplay[params.row.calcMethod || '-']}</span>
          },
          showType: ['comparisonMetrics']
        },
        {
          title: this.$t('m_calculation_period'), // 计算周期
          key: 'calcPeriod',
          width: 100,
          render: (h, params) => {
            return <span>{params.row.calcPeriod || '-'}S</span>
          },
          showType: ['comparisonMetrics']
        },
        {
          title: this.$t('m_business_configuration'), // 业务配置
          key: 'log_metric_group_name',
          width: 200,
          render: (h, params) => {
            return <span>{params.row.log_metric_group_name || '-'}</span>
          },
          showType: ['originalMetrics']
        },
        {
          title: this.$t('m_update_time'), // 更新时间
          key: 'update_time',
          width: 150,
          render: (h, params) => {
            return <span>{params.row.update_time || '-'}</span>
          },
          showType: ['originalMetrics', 'comparisonMetrics']
        },
        {
          title: this.$t('m_updatedBy'), // 更新人
          key: 'update_user',
          width: 150,
          render: (h, params) => {
            return <span>{params.row.update_user || '-'}</span>
          },
          showType: ['originalMetrics', 'comparisonMetrics']
        },
        // {
        //   title: this.$t('m_table_action'),
        //   key: 'action',
        //   width: 100,
        //   fixed: 'right',
        //   render: (h, params) => {
        //     return (
        //       <div style="display:flex;">
        //          {
        //           /* 查看 */
        //           <Tooltip content={this.$t('m_button_view')} placement="bottom" transfer>
        //             <Button
        //               size="small"
        //               type="info"
        //               onClick={() => {
        //                 this.showConfigModal(params.row)
        //               }}
        //               style="margin-right:5px;"
        //             >
        //               <Icon type="md-eye" size="16"></Icon>
        //             </Button>
        //           </Tooltip>
        //         }
        //       </div>
        //     )
        //   },
        //   showType: ['originalMetrics', 'comparisonMetrics']
        // }
      ],
      workspaceMap: {
        all_object: this.$t('m_all_object'), // 全部对象
        any_object: this.$t('m_any_object') // 层级对象
      },
      row: {},
      addVisible: false, // 是否显示查看
    }
  },
  async mounted () {
    await this.getEndpointList()
    this.endpoint = this.endpointOptions[0].guid
    this.getList()
    this.token = (window.request ? 'Bearer ' + getPlatFormToken() : getToken())|| null
    const clientHeight = document.documentElement.clientHeight
    this.maxHeight = clientHeight - this.$refs.maxheight.getBoundingClientRect().top - 100
  },
  methods: {
    reloadData (metricType) {
      this.metricType = metricType
      this.getList()
    },
    changeEndpointGroup () {
      this.getList()
    },
    getList () {
      const params = {
        endpoint: this.endpoint
      }
      const api = this.metricType === 'originalMetrics' ? '/monitor/api/v2/monitor/metric/list' : '/monitor/api/v2/monitor/metric_comparison/list'
      this.$root.$httpRequestEntrance.httpRequestEntrance(
        'GET',
        api,
        params,
        responseData => {
          this.tableData = responseData
        },
        { isNeedloading: true }
      )
    },
    getEndpointList () {
      const api = '/monitor/api/v1/alarm/endpoint/list?__orders=-created_date&page=1&size=1000000'
      return this.$root.$httpRequestEntrance.httpRequestEntrance(
        'GET',
        api,
        '',
        responseData => {
          this.endpointOptions = responseData.data || []
        },
        { isNeedloading:false }
      )
    },
    showConfigModal (row) {
      this.row = row
      this.originalMetricsId = row.guid
      this.monitorType = row.monitor_type
      this.addVisible = true
    }
  }
}
</script>

<style lang="less">
.monitor-level-group {
  .eclipse {
    overflow: hidden;
    text-overflow: ellipsis;
    display: -webkit-box;
    -webkit-line-clamp: 2;
    -webkit-box-orient: vertical;
  }
}
</style>
<style lang="less" scoped>
.monitor-level-group {
  padding-bottom: 20px;
  .btn-group {
    display: flex;
    justify-content: flex-end;
  }
  .level-table {
    margin-top: 12px;
  }
}
</style>