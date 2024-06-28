<template>
  <div ref="maxheight" class="monitor-general-group">
    <Row>
      <Col :span="8">
        <!--对象类型-->
        <span style="font-size: 14px;">
          {{$t('m_tableKey_endpoint_type')}}：
        </span>
        <Select filterable v-model="monitorType" @on-change="changeMonitorType" style="width:300px">
          <Option v-for="(i, index) in monitorTypeOptions" :value="i" :key="index">{{ i }}</Option>
        </Select>
      </Col>
      <Col :span="16">
        <div class="btn-group">
          <MetricChange @reloadData="reloadData"></MetricChange>
          <!--新增-->
          <Button type="success" @click="handleAdd">{{ $t('m_button_add') }}</Button>
        </div>
      </Col>
    </Row>
    <Table size="small" :columns="tableColumns" :data="tableData" class="general-table"/>
    <Modal
      v-model="deleteVisible"
      :title="$t('m_delConfirm_title')"
      @on-ok="submitDelete"
      @on-cancel="deleteVisible = false">
      <div class="modal-body" style="padding: 10px 20px">
        <p style="color: red">{{ $t('m_metric_deleteTips') }}</p>
      </div>
    </Modal>
    <AddGroupDrawer
      v-if="addVisible && metricType==='originalMetrics'"
      :visible.sync="addVisible"
      :monitorType="monitorType"
      :data="row"
      :operator="type"
      @fetchList="getList()"
    ></AddGroupDrawer>
    <YearOverYear
      v-if="addVisible && metricType==='comparisonMetrics'"
      :visible.sync="addVisible"
      :monitorType="monitorType"
      :data="row"
      :operator="type"
      :metricList="tableData"
      @fetchList="getList()"
    ></YearOverYear>
  </div>
</template>

<script>
import axios from 'axios'
import {baseURL_config} from '@/assets/js/baseURL'
import { getToken, getPlatFormToken } from '@/assets/js/cookies.ts'
import MetricChange from './components/metric-change.vue'
import AddGroupDrawer from './components/add-group.vue'
import YearOverYear from './components/year-over-year.vue'
export default {
  components: {
    AddGroupDrawer,
    YearOverYear,
    MetricChange
  },
  data () {
    return {
      metricType: 'originalMetrics', // 原始指标originalMetrics、同环比指标comparisonMetrics
      monitorType: '',
      serviceGroup: '',
      monitorTypeOptions: [],
      maxHeight: 500,
      tableData: [],
      tableColumns: [
        {
          title: this.$t('m_field_metric'), // 指标
          key: 'metric',
          width: 250
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
          }
        },
        {
          title: this.$t('m_field_type'), // 类型
          key: 'metric_type',
          width: 160,
          render: (h, params) => {
            const typeList = [
              { label: this.$t('m_base_group'), value: 'common', color: '#2d8cf0' },
              { label: this.$t('m_business_configuration'), value: 'business', color: '#81b337' },
              { label: this.$t('m_customize'), value: 'custom', color: '#b886f8' }
            ]
            const find = typeList.find(item => item.value === params.row.metric_type) || {}
            return <Tag size="medium" type="border" color={find.color}>{find.label || '-'}</Tag>
          }
        },
        {
          title: this.$t('m_tableKey_expr'), // 表达式
          key: 'prom_expr',
          minWidth: 400,
          render: (h, params) => {
            return (
              <Tooltip max-width="300" content={params.row.prom_expr} transfer>
                <span class="eclipse">{params.row.prom_expr || '-'}</span>
              </Tooltip>
            )
          }
        },
        {
          title: this.$t('m_update_time'), // 更新时间
          key: 'update_time',
          width: 160,
          render: (h, params) => {
            return <span>{params.row.update_time || '-'}</span>
          }
        },
        {
          title: this.$t('m_updatedBy'), // 更新人
          key: 'update_user',
          width: 160,
          render: (h, params) => {
            return <span>{params.row.update_user || '-'}</span>
          }
        },
        {
          title: this.$t('m_table_action'),
          key: 'action',
          width: 100,
          
          fixed: 'right',
          render: (h, params) => {
            return (
              <div style="display:flex;justify-content:center;">
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
          }
        }
      ],
      row: {},
      type: '', // add、edit
      addVisible: false,
      deleteVisible: false
    }
  },
  mounted () {
    this.getMonitorType()
    const clientHeight = document.documentElement.clientHeight
    this.maxHeight = clientHeight - this.$refs.maxheight.getBoundingClientRect().top - 100
  },
  methods: {
    reloadData (metricType) {
      this.metricType = metricType
      this.getList()
    },
    getList () {
      const params = {
        monitorType: this.monitorType,
        onlyService: 'Y',
        serviceGroup: this.serviceGroup
      }
      const api = this.metricType === 'originalMetrics' ? '/monitor/api/v2/monitor/metric/list' : '/monitor/api/v2/monitor/metric_comparison/list'
      console.log(api)
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', api, params, responseData => {
        this.tableData = responseData
      }, {isNeedloading: true})
    },
    getMonitorType () {
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.getEndpointType, '', (responseData) => {
        this.monitorTypeOptions = responseData || []
        this.monitorType = this.monitorTypeOptions[0]
        this.getList()
      }, {isNeedloading: false})
    },
    changeMonitorType () {
      this.getList()
    },
    handleAdd () {
      this.type = 'add'
      this.addVisible = true
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
      this.$root.$httpRequestEntrance.httpRequestEntrance(
        'DELETE',
        `${this.$root.apiCenter.metricManagement}?id=${this.row.guid}`,
        '',
        () => {
          this.$Message.success(this.$t('m_tips_success'))
          this.getList()
        })
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