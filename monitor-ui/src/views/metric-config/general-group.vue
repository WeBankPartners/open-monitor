<template>
  <div ref="maxheight" class="monitor-general-group">
    <Row>
      <Col :span="8">
        <span style="font-size: 14px;">
          {{$t('field.type')}}：
        </span>
        <Select filterable v-model="monitorType" @on-change="changeMonitorType" style="width:300px">
          <Option v-for="(i, index) in monitorTypeOptions" :value="i" :key="index">{{ i }}</Option>
        </Select>
      </Col>
      <Col :span="16">
        <div class="btn-group">
          <!--导出-->
          <Button @click.stop="exportData">{{$t("m_export")}}{{$t("m_metric")}}</Button>
          <!--导入-->
          <Upload 
          :action="uploadUrl" 
          :show-upload-list="false"
          :max-size="1000"
          with-credentials
          :headers="{'Authorization': token}"
          :on-success="uploadSucess"
          :on-error="uploadFailed">
            <Button icon="ios-cloud-upload-outline">{{$t('m_import')}}{{$t("m_metric")}}</Button>
          </Upload>
          <!--新增-->
          <Button type="success" @click="handleAdd">{{ $t('button.add') }}</Button>
        </div>
      </Col>
    </Row>
    <Table size="small" :columns="tableColumns" :data="tableData" class="table" />
    <Modal
      v-model="deleteVisible"
      :title="$t('delConfirm.title')"
      @on-ok="submitDelete"
      @on-cancel="deleteVisible = false">
      <div class="modal-body" style="padding:30px">
        <div style="text-align:center">
          <p style="color: red">{{$t('delConfirm.tip')}}</p>
        </div>
      </div>
    </Modal>
    <AddGroupDrawer
      v-if="addVisible"
      :visible.sync="addVisible"
      :monitorType="monitorType"
      :data="row"
      :operator="type"
      @fetchList="getList()"
    ></AddGroupDrawer>
  </div>
</template>

<script>
import axios from 'axios'
import {baseURL_config} from '@/assets/js/baseURL'
import { getToken, getPlatFormToken } from '@/assets/js/cookies.ts'
import AddGroupDrawer from './components/add-group.vue'
export default {
  components: {
    AddGroupDrawer
  },
  data () {
    return {
      token: null,
      monitorType: 'process',
      serviceGroup: '',
      monitorTypeOptions: [],
      maxHeight: 500,
      tableData: [],
      tableColumns: [
        {
          title: this.$t('field.metric'), // 指标
          key: 'metric',
          minWidth: 150
        },
        {
          title: this.$t('m_scope'), // 作用域
          key: 'workspace',
          minWidth: 150,
          render: (h, params) => {
            const workspaceMap = {
              all_object: this.$t('m_all_object'), // 全部对象
              any_object: this.$t('m_any_object') // 层级对象
            }
            return <Tag size="medium">{workspaceMap[params.row.workspace] || '-'}</Tag>
          }
        },
        {
          title: this.$t('field.type'), // 类型
          key: 'metric_type',
          minWidth: 150,
          render: (h, params) => {
            const typeList = [
              { label: this.$t('m_general_type'), value: 'common', color: '#2d8cf0' },
              { label: this.$t('m_business_configuration'), value: 'business', color: '#81b337' },
              { label: this.$t('m_customize'), value: 'custom', color: '#b886f8' }
            ]
            const find = typeList.find(item => item.value === params.row.metric_type) || {}
            return <Tag size="medium" color={find.color}>{find.label || '-'}</Tag>
          }
        },
        {
          title: this.$t('tableKey.expr'), // 表达式
          key: 'prom_expr',
          minWidth: 150,
          render: (h, params) => {
            return (
              <Tooltip max-width="300" content={params.row.prom_expr} transfer>
                <span class="eclipse">{params.row.prom_expr || '-'}</span>
              </Tooltip>
            )
          }
        },
        {
          title: this.$t('table.action'),
          key: 'action',
          width: 100,
          align: 'center',
          fixed: 'right',
          render: (h, params) => {
            return (
              <div style="display:flex;justify-content:center;">
                {
                  /* 编辑 */
                  <Tooltip content={this.$t('button.edit')} placement="bottom" transfer>
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
                  <Tooltip content={this.$t('button.remove')} placement="bottom" transfer>
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
  computed: {
    uploadUrl: function() {
      return baseURL_config + `${this.$root.apiCenter.metricImport}?serviceGroup=${this.serviceGroup}&monitorType=${this.monitorType}`
    }
  },
  mounted () {
    this.getMonitorType()
    this.getList()
    this.token = (window.request ? 'Bearer ' + getPlatFormToken() : getToken())|| null
    const clientHeight = document.documentElement.clientHeight
    this.maxHeight = clientHeight - this.$refs.maxheight.getBoundingClientRect().top - 100
  },
  methods: {
    getList () {
      const params = {
        monitorType: this.monitorType,
        onlyService: 'Y',
        serviceGroup: this.serviceGroup
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', '/monitor/api/v2/monitor/metric/list', params, responseData => {
        this.tableData = responseData
      }, {isNeedloading: true})
    },
    getMonitorType () {
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.getEndpointType, '', (responseData) => {
        this.monitorTypeOptions = responseData
      }, {isNeedloading: false})
    },
    changeMonitorType () {
      this.getList()
    },
    exportData () {
      const api = `${this.$root.apiCenter.metricExport}?serviceGroup=${this.serviceGroup}&monitorType=${this.monitorType}`
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
        this.$Message.warning(this.$t('tips.failed'))
      })
    },
    uploadSucess () {
      this.$Message.success(this.$t('tips.success'))
      this.getList()
    },
    uploadFailed (error, file) {
      this.$Message.warning(file.message)
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
          this.$Message.success(this.$t('tips.success'))
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
  .table td, .table th {
    vertical-align: center;
    border-top: none;
  }
  .table td {
    padding: 6px 0;
  }
  .table th {
    padding: 10px 0;
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
  .table {
    margin-top: 12px;
  }
}
</style>