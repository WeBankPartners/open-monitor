<template>
  <div ref="maxheight" class="monitor-level-group">
    <Row>
      <Col :span="8">
        <!--层级对象-->
        <span style="font-size: 14px;">
          {{$t('field.resourceLevel')}}:
        </span>
        <Select
          style="width:300px;"
          v-model="serviceGroup"
          filterable 
          @on-change="changeServiceGroup"
        >
          <Option v-for="(option, index) in recursiveOptions" :value="option.guid" :label="option.display_name" :key="index">
            <TagShow :list="recursiveOptions" name="type" :tagName="option.type" :index="index"></TagShow> 
            {{option.display_name}}
          </Option>
        </Select>
      </Col>
      <Col :span="16">
        <div class="btn-group">
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
          <Button type="success" @click="handleAdd">{{$t('button.add')}}</Button>
        </div>
      </Col>
    </Row>
    <Table size="small" :columns="tableColumns" :data="tableData" class="level-table" />
    <Modal
      v-model="deleteVisible"
      :title="$t('delConfirm.title')"
      @on-ok="submitDelete"
      @on-cancel="deleteVisible = false">
      <div class="modal-body" style="padding: 10px 20px">
        <p style="color: red">{{ $t('m_metric_deleteTips') }}</p>
      </div>
    </Modal>
    <AddGroupDrawer
      v-if="addVisible"
      :visible.sync="addVisible"
      :monitorType="monitorType"
      :serviceGroup="serviceGroup"
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
import TagShow from '@/components/Tag-show.vue'
import AddGroupDrawer from './components/add-group.vue'
export default {
  components: {
    TagShow,
    AddGroupDrawer
  },
  data () {
    return {
      token: null,
      monitorType: 'process',
      serviceGroup: '',
      recursiveOptions: [],
      maxHeight: 500,
      tableData: [],
      tableColumns: [
        {
          title: this.$t('field.metric'), // 指标
          key: 'metric',
          width: 250
        },
        {
          title: this.$t('m_scope'), // 作用域
          key: 'workspace',
          width: 150,
          render: (h, params) => {
            return <Tag size="medium">{ this.workspaceMap[params.row.workspace] }</Tag>
          }
        },
        {
          title: this.$t('field.type'), // 类型
          key: 'metric_type',
          width: 150,
          render: (h, params) => {
            const typeList = [
              { label: this.$t('m_base_group'), value: 'common', color: '#2d8cf0' },
              { label: this.$t('m_business_configuration'), value: 'business', color: '#81b337' },
              { label: this.$t('m_customize'), value: 'custom', color: '#b886f8' }
            ]
            const find = typeList.find(item => item.value === params.row.metric_type) || {}
            return <Tag color={find.color} type="border" size="medium">{find.label || '-'}</Tag>
          }
        },
        {
          title: this.$t('tableKey.expr'), // 表达式
          key: 'prom_expr',
          minWidth: 250,
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
          width: 150,
          render: (h, params) => {
            return <span>{params.row.update_time || '-'}</span>
          }
        },
        {
          title: this.$t('m_updatedBy'), // 更新人
          key: 'update_user',
          width: 150,
          render: (h, params) => {
            return <span>{params.row.update_user || '-'}</span>
          }
        },
        {
          title: this.$t('m_business_configuration'), // 业务配置
          key: 'log_metric_group_name',
          width: 200,
          render: (h, params) => {
            return <span>{params.row.log_metric_group_name || '-'}</span>
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
      workspaceMap: {
        all_object: this.$t('m_all_object'), // 全部对象
        any_object: this.$t('m_any_object') // 层级对象
      },
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
  async mounted () {
    await this.getRecursiveList()
    this.serviceGroup = this.recursiveOptions[0].guid
    this.getList()
    this.token = (window.request ? 'Bearer ' + getPlatFormToken() : getToken())|| null
    const clientHeight = document.documentElement.clientHeight
    this.maxHeight = clientHeight - this.$refs.maxheight.getBoundingClientRect().top - 100
  },
  methods: {
    changeServiceGroup () {
      this.getList()
    },
    getList () {
      const params = {
        monitorType: this.monitorType,
        onlyService: 'Y',
        serviceGroup: this.serviceGroup
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance(
        'GET',
        '/monitor/api/v2/monitor/metric/list',
        params,
        responseData => {
          this.tableData = responseData
        },
        { isNeedloading: true }
      )
    },
    getRecursiveList () {
      const api = this.$root.apiCenter.getTargetByEndpoint + '/group'
      return this.$root.$httpRequestEntrance.httpRequestEntrance(
        'GET',
        api,
        '',
        responseData => {
          this.recursiveOptions = responseData || []
        },
        { isNeedloading:false }
      )
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
    uploadSucess (val) {
      if (val.status === 'OK') {
        if (val.data) {
          if (val.data.fail_list.length > 0) {
            this.$Notice.error({
              duration: 0,
              render: () => {
                return <div>
                  {this.$t('m_metric_export_errorTips')}
                  <span style="color:red;">{val.data.fail_list.join('、')}</span>
                  {this.$t('m_metric')}
                </div>
              }
            })
          } else {
            this.$Message.success(this.$t('tips.success'))
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