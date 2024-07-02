<template>
  <div ref="maxheight" class="monitor-level-group">
    <Row>
      <Col :span="8">
        <!--对象组-->
        <span style="font-size: 14px;">
          {{$t('m_object_group')}}:
        </span>
        <Select
          style="width:300px;"
          v-model="endpointGroup"
          filterable 
          @on-change="changeEndpointGroup"
        >
          <Option v-for="(option, index) in objectGroupOptions" :value="option.guid" :label="option.display_name" :key="index">
            <TagShow :list="objectGroupOptions" name="type" :tagName="option.monitor_type" :index="index"></TagShow> 
            {{option.display_name}}
          </Option>
        </Select>
      </Col>
      <Col :span="16">
        <div class="btn-group">
          <MetricChange ref="metricChangeRef" @reloadData="reloadData" ></MetricChange>
          <Button
            type="info"
            @click.stop="exportData"
          >
            <img src="@/assets/img/export.png" alt="" style="width:16px;" />
            {{ $t("m_export") }}
          </Button>
          <Upload 
          :action="uploadUrl" 
          :show-upload-list="false"
          :max-size="1000"
          with-credentials
          :headers="{'Authorization': token}"
          :on-success="uploadSucess"
          :on-error="uploadFailed">
            <Button type="primary">
              <img src="@/assets/img/import.png" alt="" style="width:16px;" />
              {{ $t('m_import') }}
            </Button>
          </Upload>
          <Button type="success" @click="handleAdd">{{$t('m_button_add')}}</Button>
        </div>
      </Col>
    </Row>
    <Table size="small" :columns="tableColumns.filter(col=>col.showType.includes(metricType))" :data="tableData" class="level-table" />
    <Modal
      v-model="deleteVisible"
      :title="$t('m_delConfirm_title')"
      @on-ok="submitDelete"
      @on-cancel="deleteVisible = false">
      <div class="modal-body" style="padding:10px 20px;">
        <p style="color: red">{{ $t('m_metric_deleteTips') }}</p>
      </div>
    </Modal>
    <AddGroupDrawer
      v-if="addVisible && metricType==='originalMetrics'"
      :visible.sync="addVisible"
      :monitorType="monitorType"
      :endpoint_group="endpointGroup"
      :serviceGroup="serviceGroup"
      :data="row"
      :operator="type"
      @fetchList="getList()"
    ></AddGroupDrawer>
    <YearOverYear
      ref="yearOverYearRef"
      v-if="addVisible && metricType==='comparisonMetrics'"
      :visible.sync="addVisible"
      :monitorType="monitorType"
      :originalMetricsId="originalMetricsId"
      :serviceGroup="serviceGroup"
      :data="row"
      :operator="type"
      :endpointGroup="endpointGroup"
      @fetchList="getList()"
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
    AddGroupDrawer,
    YearOverYear,
    MetricChange
  },
  data () {
    return {
      metricType: 'originalMetrics', // 原始指标originalMetrics、同环比指标comparisonMetrics
      token: null,
      monitorType: 'process',
      serviceGroup: '',
      endpointGroup: '',
      objectGroupOptions: [],
      maxHeight: 500,
      tableData: [],
      tableColumns: [
        {
          title: this.$t('m_field_metric'), // 指标
          key: 'metric',
          width: 250,
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
          width: 100,
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
        {
          title: this.$t('m_table_action'),
          key: 'action',
          width: 140,
          fixed: 'right',
          render: (h, params) => {
            return (
              <div style="display:flex;justify-content:center;">
                 {
                  this.metricType === 'originalMetrics' &&
                  /* 新增同环比指标 */
                  <Tooltip content={this.$t('m_button_add')+this.$t('m_year_over_year_metrics')} placement="bottom" transfer>
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
                  /* 编辑 */
                  <Tooltip content={this.$t('m_button_edit')} placement="bottom" transfer>
                    <Button
                      size="small"
                      type="primary"
                      onClick={() => {
                        this.handleEdit(params.row)
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
                    <Button
                      size="small"
                      type="error"
                      onClick={() => {
                        this.handleDelete(params.row)
                      }}
                      style="margin-right:5px;"
                    >
                      <Icon type="md-trash" size="16"></Icon>
                    </Button>
                  </Tooltip>
                }
              </div>
            )
          },
          showType: ['originalMetrics', 'comparisonMetrics']
        }
      ],
      workspaceMap: {
        all_object: this.$t('m_all_object'), // 全部对象
        any_object: this.$t('m_any_object') // 层级对象
      },
      row: {},
      type: '', // add、edit
      addVisible: false,
      deleteVisible: false,
      originalMetricsId: ''
    }
  },
  computed: {
    uploadUrl: function() {
      return baseURL_config + `${this.$root.apiCenter.metricImport}?serviceGroup=${this.serviceGroup}&monitorType=${this.monitorType}&endpointGroup=${this.endpointGroup}`
    }
  },
  async mounted () {
    await this.getObjectGroupList()
    this.endpointGroup = this.objectGroupOptions[0].guid
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
        endpointGroup: this.endpointGroup
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
    getObjectGroupList () {
      const api = '/monitor/api/v2/alarm/endpoint_group/query?__orders=-created_date&page=1&size=1000000'
      return this.$root.$httpRequestEntrance.httpRequestEntrance(
        'GET',
        api,
        '',
        responseData => {
          this.objectGroupOptions = responseData.data || []
        },
        { isNeedloading:false }
      )
    },
    exportData () {
      const api = `${this.$root.apiCenter.metricExport}?serviceGroup=${this.serviceGroup}&monitorType=${this.monitorType}&endpointGroup=${this.endpointGroup}`
      axios({
        method: 'GET',
        url: api,
        headers: {
          'Authorization': this.token
        }
      }).then((response) => {
        if (response.status < 400) {
          let content = JSON.stringify(response.data)
          let fileName = `${response.headers['content-disposition'].split(';')[1].trim().split('=')[1]}`
          let blob = new Blob([content])
          if('msSaveOrOpenBlob' in navigator){
            // Microsoft Edge and Microsoft Internet Explorer 10-11
            window.navigator.msSaveOrOpenBlob(blob, fileName)
          } else {
            if ('download' in document.createElement('a')) { // 非IE下载
              let elink = document.createElement('a')
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
    uploadSucess (val) {
      if (val.status === 'OK') {
        if (val.data) {
          if (Array.isArray(val.data.fail_list) && val.data.fail_list.length > 0) {
            this.$Notice.error({
              duration: 0,
              render: () => {
                return <div>
                  {this.$t('m_metric_export_errorTips')}
                  <span style="color:red;"> {val.data.fail_list.join('、')}</span>
                </div>
              }
            })
          } else {
            this.$Message.success(this.$t('m_tips_success'))
          }
        }
        this.getList()
      }
    },
    uploadFailed (error, file) {
      this.$Message.warning(file.message)
    },
    handleAdd () {
      this.type = 'add'
      this.addVisible = true
      this.originalMetricsId = ''
    },
    handleEdit (row) {
      this.type = 'edit'
      this.row = row
      this.addVisible = true
    },
    handleDelete (row) {
      this.row = row
      this.deleteVisible = true
    },
    submitDelete () {
      const api = this.metricType === 'originalMetrics' ? `${this.$root.apiCenter.metricManagement}?id=${this.row.guid}` : `/monitor/api/v1/dashboard/new/comparison_metric/${this.row.guid}`
      this.$root.$httpRequestEntrance.httpRequestEntrance(
        'DELETE',
        api,
        '',
        () => {
          this.$Message.success(this.$t('m_tips_success'))
          this.getList()
        })
    },
    // 同环比入口
    handleAddYearOverYear (row) {
      this.metricType = 'comparisonMetrics'
      this.addVisible = true
      this.operator = 'add'
      this.originalMetricsId = row.guid
      this.$refs.metricChangeRef.metricTypeChange(this.metricType)
    },
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