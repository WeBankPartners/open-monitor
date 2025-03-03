<template>
  <div class="monitor-general-group">
    <Row>
      <Col :span="16">
      <!--对象类型-->
      <span style="font-size: 14px;">
        {{$t('m_basic_type')}}：
      </span>
      <Select
        v-model="monitorType"
        filterable
        clearable
        @on-change="() => {
          metric = ''
          onFilterChange()
        }"
        style="width:300px"
      >
        <Option v-for="(i, index) in monitorTypeOptions" :value="i" :key="index">{{ i }}</Option>
      </Select>
      <Input
        v-model="metric"
        clearable
        style="width: 250px; margin-left: 10px"
        :placeholder="$t('m_placeholder_input') + (metricType === 'comparisonMetrics' ? $t('m_button_MoM') : '' ) + $t('m_metric')"
        @on-change='onFilterChange'
      />
      </Col>
      <Col :span="8">
      <div class="btn-group">
        <Button
          class="btn-upload"
          @click.stop="exportData"
        >
          <img src="@/styles/icon/DownloadOutlined.png" class="upload-icon" />
          {{ $t('m_export') }}
        </Button>
        <Upload
          :action="uploadUrl"
          :show-upload-list="false"
          :max-size="1000"
          with-credentials
          :headers="{'Authorization': token}"
          :on-success="uploadSucess"
          :on-error="uploadFailed"
        >
          <Button class="btn-upload">
            <img src="@/styles/icon/UploadOutlined.png" class="upload-icon" />
            {{ $t('m_import') }}
          </Button>
        </Upload>
        <Button type="success" @click="handleAdd">{{$t('m_button_add')}}</Button>
      </div>
      </Col>
    </Row>
    <Table
      ref="maxHeight"
      size="small"
      :columns="tableColumns.filter(col=>col.showType.includes(metricType))"
      :data="tableData"
      :max-height="maxHeight"
      class="general-table"
    />
    <AddGroupDrawer
      v-if="addVisible && showDrawer === 'originalMetrics'"
      :visible.sync="addVisible"
      :monitorType="monitorType"
      :data="row"
      :operator="type"
      :viewOnly="viewOnly"
      fromPage="generalGroup"
      @fetchList="reloadData('originalMetrics')"
    ></AddGroupDrawer>
    <YearOverYear
      ref="yearOverYearRef"
      v-if="addVisible && showDrawer === 'comparisonMetrics'"
      :visible.sync="addVisible"
      :monitorType="monitorType"
      :originalMetricsId="originalMetricsId"
      serviceGroup=""
      :data="row"
      :operator="type"
      :viewOnly="viewOnly"
      @fetchList="reloadData('comparisonMetrics')"
    ></YearOverYear>
  </div>
</template>

<script>
import {debounce} from 'lodash'
import axios from 'axios'
import {baseURL_config} from '@/assets/js/baseURL'
import { getToken, getPlatFormToken } from '@/assets/js/cookies.ts'
import AddGroupDrawer from './components/add-group.vue'
import YearOverYear from './components/year-over-year.vue'
export const custom_api_enum = [ // 对于当前文档中没有放入this.request中的path,为了获取路由
  {
    'metricList.api': 'get'
  },
  {
    metricComparisonList: 'get'
  },
  {
    metricManagement: 'delete'
  },
  {
    comparisonMetricItem: 'delete'
  },
  {
    metricExport: 'get'
  },
  {
    metricImport: 'post'
  }
]

export default {
  components: {
    AddGroupDrawer,
    YearOverYear
  },
  data() {
    return {
      metricType: 'originalMetrics', // 原始指标originalMetrics、同环比指标comparisonMetrics
      monitorType: '',
      serviceGroup: '',
      monitorTypeOptions: [],
      maxHeight: 500,
      tableData: [],
      tableColumns: [
        {
          title: this.$t('m_year_over_year_metrics'), // 指标
          key: 'metric',
          minWidth: 250,
          showType: ['comparisonMetrics']
        },
        {
          title: this.$t('m_metric'), // 原始指标-指标
          key: 'metric',
          minWidth: 200,
          showType: ['originalMetrics']
        },
        {
          title: this.$t('m_metric'), // 同环比指标-指标
          key: 'origin_metric',
          minWidth: 200,
          showType: ['comparisonMetrics']
        },
        {
          title: this.$t('m_scope'), // 作用域
          key: 'workspace',
          width: 160,
          render: (h, params) => {
            const workspaceMap = {
              all_object: this.$t('m_all_object'), // 全部对象
              any_object: this.$t('m_any_object') // 层级对象
            }
            return <Tag size="medium">{workspaceMap[params.row.workspace] || '-'}</Tag>
          },
          showType: ['originalMetrics', 'comparisonMetrics']
        },
        {
          title: this.$t('m_configuration_page'), // 配置页面
          key: 'metric_type',
          width: 140,
          render: (h, params) => {
            const typeList = [
              {
                label: this.$t('m_basic_type'),
                value: 'common',
                color: '#5384FF'
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
            const find = typeList.find(item => item.value === params.row.metric_type) || {}
            return <Tag size="medium" type="border" color={find.color}>{find.label || '-'}</Tag>
          },
          showType: ['originalMetrics', 'comparisonMetrics']
        },
        {
          title: this.$t('m_tableKey_expr'), // 表达式
          key: 'prom_expr',
          minWidth: 400,
          render: (h, params) => (
            <Tooltip max-width="300" content={params.row.prom_expr} transfer>
              <span class="eclipse">{params.row.prom_expr || '-'}</span>
            </Tooltip>
          ),
          showType: ['originalMetrics']
        },
        {
          title: this.$t('m_comparison_types'), // 对比类型
          key: 'comparisonType',
          width: 100,
          render: (h, params) => {
            const comparisonTypeToDisplay = {
              day: this.$t('m_dod_comparison'),
              week: this.$t('m_wow_comparison'),
              month: this.$t('m_mom_comparison'),
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
              diff: this.$t('m_difference'),
              diff_percent: this.$t('m_percentage_difference')
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
              avg: this.$t('m_average'),
              min: this.$t('m_min'),
              max: this.$t('m_max'),
              sum: this.$t('m_sum'),
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
          render: (h, params) => <span>{params.row.calcPeriod || '-'}S</span>,
          showType: ['comparisonMetrics']
        },
        {
          title: this.$t('m_update_time'), // 更新时间
          key: 'update_time',
          width: 160,
          render: (h, params) => <span>{params.row.update_time || '-'}</span>,
          showType: ['originalMetrics', 'comparisonMetrics']
        },
        {
          title: this.$t('m_updatedBy'), // 更新人
          key: 'update_user',
          width: 160,
          render: (h, params) => <span>{params.row.update_user || '-'}</span>,
          showType: ['originalMetrics', 'comparisonMetrics']
        },
        {
          title: this.$t('m_table_action'),
          key: 'action',
          width: 220,
          showType: ['originalMetrics', 'comparisonMetrics'],
          fixed: 'right',
          render: (h, params) => (
            <div style="display:flex;justify-content:center;">
              <Tooltip max-width={400} placement="top" transfer content={this.$t('m_copy')}>
                <Button size="small" class="mr-1" type="success" on-click={() => this.handleCopyItem(params.row)}>
                  <Icon type="md-document" size="16"></Icon>
                </Button>
              </Tooltip>
              {
                this.metricType === 'originalMetrics'
                  /* 新增同环比指标 */
                  && <Tooltip content={this.$t('m_button_add')+this.$t('m_year_over_year_metrics')} placement="bottom" transfer>
                    <Button
                      size="small"
                      type="success"
                      onClick={() => {
                        this.handleAddYearOverYear(params.row)
                      }}
                      style="margin-right:5px;"
                    >
                      <Icon type="md-add" size="16"></Icon>
                    </Button>
                  </Tooltip>
              }
              {
                /* 查看 */
                <Tooltip content={this.$t('m_button_view')} placement="top" transfer={true}>
                  <Button
                    size="small"
                    type="info"
                    onClick={() => this.handleEdit(params.row, true)}
                    style="margin-right:5px;"
                  >
                    <Icon type="ios-eye" size="16"></Icon>
                  </Button>
                </Tooltip>
              }
              {
                /* 编辑 */
                <Tooltip content={this.$t('m_button_edit')} placement="bottom" transfer>
                  <Button
                    size="small"
                    type="primary"
                    onClick={() => {
                      this.handleEdit(params.row, false)
                    }}
                    style="margin-right:5px;"
                  >
                    <Icon type="md-create" size="16"></Icon>
                  </Button>
                </Tooltip>
              }
              {
                /* 删除 */
                <Tooltip content={this.$t('m_button_remove')} placement="bottom" transfer>
                  <Poptip
                    confirm
                    transfer
                    title={this.$t('m_delConfirm_tip')}
                    placement="left-end"
                    on-on-ok={() => {
                      this.submitDelete(params.row)
                    }}>
                    <Button size="small" type="error">
                      <Icon type="md-trash" size="16" />
                    </Button>
                  </Poptip>
                </Tooltip>
              }
            </div>
          )
        }
      ],
      row: {},
      type: '', // add、edit
      addVisible: false,
      originalMetricsId: '',
      showDrawer: '', // 控制显示抽屉的类型
      viewOnly: false, // 仅查看
      token: null,
      metric: '',
      request: this.$root.$httpRequestEntrance.httpRequestEntrance,
      apiCenter: this.$root.apiCenter,
    }
  },
  computed: {
    uploadUrl() {
      return baseURL_config + `${this.apiCenter.metricImport}?serviceGroup=${this.serviceGroup}&monitorType=${this.monitorType}&comparison=${this.metricType === 'originalMetrics' ? 'N' : 'Y'}`
    }
  },
  mounted() {
    this.getMonitorType()
    this.token = (window.request ? 'Bearer ' + getPlatFormToken() : getToken())|| null
    this.maxHeight = document.documentElement.clientHeight - this.$refs.maxHeight.$el.getBoundingClientRect().top - 60
  },
  methods: {
    reloadData(metricType) {
      this.metricType = metricType
      this.getList()
    },
    getList() {
      const params = {
        monitorType: this.monitorType,
        onlyService: 'Y',
        serviceGroup: this.serviceGroup,
        metric: this.metric
      }
      const api = this.metricType === 'originalMetrics' ? this.apiCenter.metricList.api : this.apiCenter.metricComparisonList
      this.request('GET', api, params, responseData => {
        this.tableData = responseData
      }, {isNeedloading: true})
      this.getMetricTotalNumber()
    },
    getMetricTotalNumber() {
      const params = {
        monitorType: this.monitorType,
        onlyService: 'Y',
        serviceGroup: this.serviceGroup,
        metric: this.metric
      }
      this.request('GET', this.apiCenter.metricListCount, params, response => {
        this.$emit('totalCount', response, this.metricType)
      }, {isNeedloading: false})
    },
    getMonitorType() {
      this.request('GET', this.apiCenter.getEndpointType, '', responseData => {
        this.monitorTypeOptions = responseData || []
        this.monitorType = this.monitorTypeOptions[0]
        this.getList()
      }, {isNeedloading: false})
    },
    onFilterChange: debounce(function () {
      this.getList()
    }, 500),
    handleAdd() {
      this.type = 'add'
      this.viewOnly = false
      this.showDrawer = this.metricType
      this.addVisible = true
      this.originalMetricsId = ''
    },
    handleEdit(row, viewOnly) {
      this.showDrawer = this.metricType
      this.type = 'edit'
      this.viewOnly = viewOnly
      this.row = row
      this.addVisible = true
    },
    handleCopyItem(row) {
      this.showDrawer = this.metricType
      this.type = 'copy'
      this.viewOnly = false
      this.row = row
      this.addVisible = true
    },
    submitDelete(row) {
      const api = this.metricType === 'originalMetrics' ? `${this.apiCenter.metricManagement}?id=${row.guid}` : `/monitor/api/v1/dashboard/new/comparison_metric/${row.guid}`
      this.request(
        'DELETE',
        api,
        '',
        () => {
          this.$Message.success(this.$t('m_tips_success'))
          this.getList()
        }
      )
    },
    // 同环比入口
    handleAddYearOverYear(row) {
      this.showDrawer = 'comparisonMetrics'
      this.addVisible = true
      this.type = 'add'
      this.viewOnly = false
      this.originalMetricsId = row.guid
    },
    exportData() {
      const api = `${this.apiCenter.metricExport}?serviceGroup=${this.serviceGroup}&monitorType=${this.monitorType}&comparison=${this.metricType === 'originalMetrics' ? 'N' : 'Y'}`
      axios({
        method: 'GET',
        url: api,
        headers: {
          Authorization: this.token
        }
      }).then(response => {
        if (response.status < 400) {
          const content = JSON.stringify(response.data)
          const fileName = `${response.headers['content-disposition'].split(';')[1].trim().split('=')[1]}`
          const blob = new Blob([content])
          if ('msSaveOrOpenBlob' in navigator){
            // Microsoft Edge and Microsoft Internet Explorer 10-11
            window.navigator.msSaveOrOpenBlob(blob, fileName)
          } else {
            if ('download' in document.createElement('a')) { // 非IE下载
              const elink = document.createElement('a')
              elink.download = fileName
              elink.style.display = 'none'
              elink.href = URL.createObjectURL(blob)
              document.body.appendChild(elink)
              elink.click()
              URL.revokeObjectURL(elink.href) // 释放URL 对象
              document.body.removeChild(elink)
            } else { // IE10+下载
              navigator.msSaveOrOpenBlob(blob, fileName)
            }
          }
        }
      })
        .catch(() => {
          this.$Message.warning(this.$t('m_tips_failed'))
        })
    },
    uploadSucess(val) {
      if (val.status === 'OK') {
        if (val.data) {
          if (Array.isArray(val.data.fail_list) && val.data.fail_list.length > 0) {
            this.$Notice.error({
              duration: 0,
              render: () => <div>
                {this.$t('m_metric_export_errorTips')}
                <span style="color:red;"> {val.data.fail_list.join('、')}</span>
              </div>
            })
          } else {
            this.$Message.success(this.$t('m_tips_success'))
          }
        }
        this.getList()
      }
    },
    uploadFailed(file) {
      this.$Message.warning(file.message)
    }
  }
}
</script>

<style lang="less">
.monitor-general-group {
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
.monitor-general-group {
  padding-bottom: 20px;
  .btn-group {
    display: flex;
    justify-content: flex-end;
  }
  .general-table {
    margin-top: 12px;
  }
}
</style>
