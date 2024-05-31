<template>
  <div class=" ">
    <section v-if="showManagement" style="margin-top: 16px;">
      <Tag color="blue">{{$t('m_log_file')}}</Tag>
      <button @click="add" type="button" class="btn btn-small success-btn" style="padding: 0 10px">
        <i class="fa fa-plus"></i>
        {{$t('button.add')}}
      </button>

      <button type="button" style="margin-left:16px" class="btn-cancel-f" @click="exportData">{{$t("m_export")}}</button>
      <div style="display: inline-block;margin-bottom: 3px;"> 
        <Upload 
        :action="uploadUrl" 
        :show-upload-list="false"
        :max-size="1000"
        with-credentials
        :headers="{'Authorization': token}"
        :on-success="uploadSucess"
        :on-error="uploadFailed">
          <Button icon="ios-cloud-upload-outline">{{$t('m_import')}}</Button>
        </Upload>
      </div>

      <PageTable :pageConfig="pageConfig">
        <div slot='tableExtend'>
          <div style="margin:8px;border:1px solid #19be6b">
            <button @click="addCustomMetric(pageConfig.table.isCustomMetricExtend.parentData)" type="button" style="margin-top:8px" class="btn btn-sm success-btn">
              <i class="fa fa-plus"></i>
              {{$t('title.logAdd')}}
            </button>
            <extendTable :detailConfig="pageConfig.table.isCustomMetricExtend.detailConfig"></extendTable>
          </div>
        </div>
      </PageTable>
    </section>
    <Modal
      v-model="addAndEditModal.isShow"
      :title="addAndEditModal.isAdd ? $t('button.add') : $t('button.edit')"
      :mask-closable="false"
      :width="730"
      >
      <div :style="{ 'max-height': MODALHEIGHT + 'px', overflow: 'auto' }">
        <div>
          <span>{{$t('field.type')}}:</span>
          <Select v-model="addAndEditModal.dataConfig.monitor_type" @on-change="getEndpoint(addAndEditModal.dataConfig.monitor_type, 'host')" style="width: 640px">
            <Option v-for="type in monitorTypeOptions" :key="type.value" :value="type.label">{{type.label}}</Option>
          </Select>
        </div>
        <div v-if="addAndEditModal.isAdd" style="margin: 4px 0px;padding:8px 12px;border:1px solid #dcdee2;border-radius:4px;width:680px">
          <template v-for="(item, index) in addAndEditModal.pathOptions">
            <p :key="index + 5">
              <Button
                v-if="addAndEditModal.isAdd"
                @click="deleteItem('path', index)"
                size="small"
                type="error"
                icon="md-trash"
              ></Button>
              <Tooltip :content="$t('tableKey.logPath')" :delay="1000">
                <Input v-model="item.path" style="width: 620px" :placeholder="$t('tableKey.logPath')" />
              </Tooltip>
            </p>
          </template>
          <Button
            @click="addEmptyItem('path')"
            type="success"
            size="small"
            style="width:650px"
            long
            >{{ $t('button.add') }}{{$t('tableKey.logPath')}}</Button
          >
        </div>
        <div v-else style="margin: 8px 0">
          <span>{{$t('tableKey.path')}}:</span>
          <Input style="width: 640px" v-model="addAndEditModal.dataConfig.log_path" />
        </div>
        <div style="margin: 4px 0px;padding:8px 12px;border:1px solid #dcdee2;border-radius:4px;width:680px">
          <template v-for="(item, index) in addAndEditModal.dataConfig.endpoint_rel">
            <p :key="index + 'c'">
              <Button
                @click="deleteItem('relate', index)"
                size="small"
                type="error"
                icon="md-trash"
              ></Button>
              <Tooltip :content="$t('m_business_object')" :delay="1000">
                <Select v-model="item.target_endpoint" style="width: 310px" :placeholder="$t('m_business_object')">
                  <Option v-for="type in targetEndpoints" :key="type.guid" :value="type.guid">{{type.display_name}}</Option>
                </Select>
              </Tooltip>
              <Tooltip :content="$t('m_log_server')" :delay="1000">
                <Select v-model="item.source_endpoint" style="width: 310px" :placeholder="$t('m_log_server')">
                  <Option v-for="type in sourceEndpoints" :key="type.guid" :value="type.guid">{{type.display_name}}</Option>
                </Select>
              </Tooltip>
            </p>
          </template>
          <Button
            @click="addEmptyItem('relate')"
            type="success"
            size="small"
            style="width:650px"
            long
            >{{$t('addStringMap')}}</Button
          >
        </div>
      </div>
      <div slot="footer">
        <Button @click="cancelAddAndEdit">{{$t('button.cancel')}}</Button>
        <Button @click="okAddAndEdit" type="primary">{{$t('button.save')}}</Button>
      </div>
    </Modal>
    <Modal
      v-model="isShowWarning"
      :title="$t('delConfirm.title')"
      @on-ok="ok"
      @on-cancel="cancel">
      <div class="modal-body" style="padding:30px">
        <div style="text-align:center">
          <p style="color: red">{{$t('delConfirm.tip')}}</p>
        </div>
      </div>
    </Modal>
    <Modal
      v-model="isShowWarningDelete"
      :title="$t('delConfirm.title')"
      @on-ok="okDelRow"
      @on-cancel="cancleDelRow">
      <div class="modal-body" style="padding:30px">
        <div style="text-align:center">
          <p style="color: red">{{$t('delConfirm.tip')}}</p>
        </div>
      </div>
    </Modal>
    <ModalComponent :modelConfig="customMetricsModelConfig">
      <div slot="ruleConfig" class="extentClass">
        <div class="marginbottom params-each">
          <label class="col-md-2 label-name">{{$t('m_regular')}}:</label>
          <Select v-model="customMetricsModelConfig.addRow.regulative" filterable clearable style="width:375px">
            <Option v-for="notify in customMetricsModelConfig.notifyEnableOption" :value="notify.value" :key="notify.value">{{
              notify.label
            }}</Option>
          </Select>
        </div>
        <div class="marginbottom params-each">
          <label class="col-md-2 label-name">{{$t('tableKey.s_priority')}}:</label>
          <Select filterable clearable v-model="customMetricsModelConfig.addRow.priority" style="width:375px">
            <Option v-for="item in customMetricsModelConfig.priorityList" :value="item.value" :key="item.value">{{ $t(item.label) }}</Option>
          </Select>
        </div>
        <div class="marginbottom params-each">
          <label class="col-md-2 label-name">{{$t('sendAlarm')}}:</label>
          <Select filterable clearable v-model="customMetricsModelConfig.addRow.notify_enable" style="width:375px">
            <Option v-for="item in customMetricsModelConfig.notifyEnableOption" :value="item.value" :key="item.value">{{ item.label }}</Option>
          </Select>
        </div>
      
        <div class="marginbottom params-each">
          <label class="col-md-2 label-name">{{$t('m_content')}}:</label>
          <Input type="textarea" v-model="customMetricsModelConfig.addRow.content" style="width: 375px"/>
        </div>

      </div>
    </ModalComponent>
  </div>
</template>

<script>
import { getToken, getPlatFormToken } from '@/assets/js/cookies.ts'
import {baseURL_config} from '@/assets/js/baseURL'
import {priorityList} from '@/assets/config/common-config.js'
import extendTable from '@/components/table-page/extend-table'
import axios from 'axios'
let tableEle = [
  {title: 'tableKey.logPath', value: 'log_path', display: true},
  {title: 'field.type', value: 'monitor_type', display: true},
]
const btn = [
  {btn_name: 'button.edit', btn_func: 'editF'},
  {btn_name: 'button.remove', btn_func: 'deleteConfirmModal', color: 'red'}
]

export default {
  name: '',
  data () {
    return {
      token: null,
      MODALHEIGHT: 300,
      isShowWarning: false,
      targrtId: '',
      targetDetail: {},
      showManagement: false,
      pageConfig: {
        table: {
          tableData: [],
          tableEle: tableEle,
          // filterMoreBtn: 'filterMoreBtn',
          primaryKey: 'id',
          btn: btn,
          handleFloat:true,
          isExtend: {
            parentData: null,
            func: 'getExtendInfo',
            data: {},
            slot: 'tableExtend',
            detailConfig: [{
              isExtendF: true,
              title: '',
              config: [
                {title: 'tableKey.regular', value: 'json_regular', display: true},
                {title: 'm_collection_interval', value: 'step', display: true},
                {title: 'tableKey.tags', value: 'tags', display: true},
                {title: 'table.action',btn:[
                  {btn_name: 'button.edit', btn_func: 'editRuleItem'},
                  {btn_name: 'button.remove', btn_func: 'delRuleconfirmModal', color: 'red'}
                ]}
              ],
              data: [1],
              scales: ['25%', '20%', '15%', '20%', '20%']
            }]
          },
          isCustomMetricExtend: {
            parentData: null,
            func: 'getExtendInfo',
            data: {},
            slot: 'rulesTableExtend',
            detailConfig: [{
              isExtendF: true,
              title: '',
              config: [
                {title: 'field.log', value: 'keyword', display: true},
                {title: 'sendAlarm', value: 'notify_enable', display: true},
                {title: 'tableKey.s_priority', value: 'priority', display: true},
                {title: 'm_regular', value: 'regulative', display: true},
                {title: 'table.action',btn:[
                  {btn_name: 'button.edit', btn_func: 'editCustomMetricItem'},
                  {btn_name: 'button.remove', btn_func: 'delCustomMetricConfirmModal', color: 'red'}
                ]}
              ],
              data: [1],
              scales: ['25%', '20%', '15%', '20%', '20%']
            }]
          }
        }
      },
      addAndEditModal: {
        isShow: false,
        isAdd: false,
        dataConfig: {
          service_group: '',
          log_path: '',
          monitor_type: '',
          endpoint_rel: []
        },
        pathOptions: [],
      },
      sourceEndpoints: [],
      targetEndpoints: [],
      showAddAndEditModal: false,
      activeData: {},
      showRegConfig: false,
      selectedData: null,
      selectedIndex: null,
      isShowWarningDelete: false,
      deleteType: '',
      showCustomRegConfig: false,
      customMetricsModelConfig: {
        modalId: 'custom_metrics',
        isAdd: true,
        modalStyle: 'min-width:550px',
        modalTitle: 'field.log',
        saveFunc: 'saveCustomMetric',
        config: [
          {label: 'tableKey.keyword', value: 'keyword', placeholder: '', disabled: false, type: 'text'},
          {name:'ruleConfig',type:'slot'}
        ],
        addRow: { // [通用]-保存用户新增、编辑时数据
          log_keyword_monitor: '',
          keyword: '',
          regulative: 0,
          notify_enable: 1,
          priority: 'low',
          content: ''
        },
        priorityList: priorityList,
        notifyEnableOption: [
          {label: 'Yes', value: 1},
          {label: 'No', value: 0}
        ]
      },
      modelTip: {
        key: '',
        value: 'keyword'
      },
      monitorTypeOptions: [
        {label: 'process', value: 'process'},
        {label: 'java', value: 'java'},
        {label: 'nginx', value: 'nginx'},
        {label: 'http', value: 'http'},
        {label: 'mysql', value: 'mysql'}
      ]
    }
  },
  computed: {
    uploadUrl: function() {
      return baseURL_config + `${this.$root.apiCenter.bussinessMonitorImport}?serviceGroup=${this.targrtId}`
    }
  },
  mounted () {
    this.MODALHEIGHT = document.body.scrollHeight - 300
    this.token = (window.request ? 'Bearer ' + getPlatFormToken() : getToken())|| null
  },
  methods: {
    exportData () {
      const api = `${this.$root.apiCenter.bussinessMonitorExport}?serviceGroup=${this.targrtId}`
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
      });
    },
    uploadSucess () {
      this.$Message.success(this.$t('tips.success'))
      this.getDetail(this.targrtId)
    },
    uploadFailed (error, file) {
      this.$Message.warning(file.message)
    },
    // other config
    editF (rowData) {
      this.getEndpoint(rowData.monitor_type, 'host')
      this.cancelAddAndEdit()
      this.addAndEditModal.isAdd = false
      this.activeData = rowData
      this.addAndEditModal.addRow = rowData
      this.modelTip.value = rowData.guid
      this.addAndEditModal.dataConfig.guid = rowData.guid
      this.addAndEditModal.dataConfig.service_group = rowData.service_group
      this.addAndEditModal.dataConfig.monitor_type = rowData.monitor_type
      this.addAndEditModal.dataConfig.log_path = rowData.log_path
      this.addAndEditModal.dataConfig.endpoint_rel = rowData.endpoint_rel
      this.addAndEditModal.isShow = true
    },
    addCustomMetricEmpty (type) {
      if (!this.customMetricsModelConfig.addRow[type]) {
        this.customMetricsModelConfig.addRow[type] = []
      }
      this.customMetricsModelConfig.addRow[type].push({
        source_value: '',
        regulative: 0,
        target_value: ''
      })
    },
    deleteCustomMetric(type, index) {
      this.customMetricsModelConfig.addRow[type].splice(index, 1)
    },
    addCustomMetric (rowData) {
      this.activeData = rowData
      this.customMetricsModelConfig.isAdd = true
      this.$root.JQ('#custom_metrics').modal('show')
    },
    saveCustomMetric () {
      let params = JSON.parse(JSON.stringify(this.customMetricsModelConfig.addRow))
      params.log_keyword_monitor = this.activeData.guid
      const requestType = this.customMetricsModelConfig.isAdd ? 'POST' : 'PUT'
      this.$root.$httpRequestEntrance.httpRequestEntrance(requestType, '/monitor/api/v2/service/log_keyword/log_keyword_config', params, () => {
        this.$Message.success(this.$t('tips.success'))
        this.$root.JQ('#custom_metrics').modal('hide')
        this.getDetail(this.targrtId)
      })
    },
    editCustomMetricItem (rowData) {
      this.customMetricsModelConfig.isAdd = false
      this.modelTip.value = rowData.keyword
      this.customMetricsModelConfig.addRow = JSON.parse(JSON.stringify(rowData))
      this.$root.JQ('#custom_metrics').modal('show')
    },
    delCustomMetricConfirmModal (rowData) {
      this.selectedData = rowData
      this.isShowWarningDelete = true
      this.deleteType = 'custom_metrics'
    },
    okDelRow () {
      if (this.deleteType === 'custom_metrics') {
        this.delCustomMericsItem(this.selectedData)
      }
    },
    delCustomMericsItem (rowData) {
      const api = '/monitor/api/v2/service/log_keyword/log_keyword_config/' + rowData.guid
      this.$root.$httpRequestEntrance.httpRequestEntrance('DELETE', api, '', () => {
        this.$Message.success(this.$t('tips.success'))
        this.getDetail(this.targrtId)
      })
    },
    cancleDelRow () {
      this.isShowWarningDelete = false
    },
    getExtendInfo(item){
      item.keyword_list.forEach(xx => xx.pId = item.guid)
      this.pageConfig.table.isCustomMetricExtend.detailConfig[0].data = item.keyword_list
      this.pageConfig.table.isCustomMetricExtend.parentData = item
    },
    deleteConfirmModal (rowData) {
      this.selectedData = rowData
      this.isShowWarning = true
    },
    ok () {
      this.delF(this.selectedData)
    },
    cancel () {
      this.isShowWarning = false
    },
    delF (rowData) {
      const api = '/monitor/api/v2/service/log_keyword/log_keyword_monitor' + '/' + rowData.guid
      this.$root.$httpRequestEntrance.httpRequestEntrance('DELETE', api, '', () => {
        this.$Message.success(this.$t('tips.success'))
        this.getDetail(this.targrtId)
      })
    },
    okAddAndEdit () {
      let params = JSON.parse(JSON.stringify(this.addAndEditModal.dataConfig))
      const methodType = this.addAndEditModal.isAdd ? 'POST' : 'PUT'
      params.service_group = this.targrtId
      if (this.addAndEditModal.isAdd) {
        params.log_path = this.addAndEditModal.pathOptions.map(p => p.path)
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance(methodType, '/monitor/api/v2/service/log_keyword/log_keyword_monitor', params, () => {
        this.$Message.success(this.$t('tips.success'))
        this.addAndEditModal.isShow = false
        this.getDetail(this.targrtId)
      }, {isNeedloading:false})
    },
    cancelAddAndEdit () {
      this.addAndEditModal.isShow = false
      this.addAndEditModal.pathOptions = []
      this.addAndEditModal.dataConfig = {
        service_group: '',
        log_path: [],
        monitor_type: '',
        endpoint_rel: []
      }
    },
    async getEndpoint (val, type) {
      this.addAndEditModal.dataConfig.endpoint_rel = []
      await this.getDefaultConfig(val, type)
      // get source Endpoint
      const sourceApi = this.$root.apiCenter.getEndpointsByType + '/' + this.targrtId + '/endpoint/' + type
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', sourceApi, '', (responseData) => {
        this.sourceEndpoints = responseData
      }, {isNeedloading:false})
      const targetApi = this.$root.apiCenter.getEndpointsByType + '/' + this.targrtId + '/endpoint/' + val
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', targetApi, '', (responseData) => {
        this.targetEndpoints = responseData
      }, {isNeedloading:false})
    },
    addEmptyItem (type) {
      switch (type) {
        case 'path': {
          const hasEmpty = this.addAndEditModal.pathOptions.every(item => item.path !== '')
          if (hasEmpty) {
            this.addAndEditModal.pathOptions.push(
              {path: ''}
            )
          } else {
            this.$Message.warning('Path Can Not Empty')
          }
          break
        }
        case 'relate': {
          const hasEmpty = this.addAndEditModal.dataConfig.endpoint_rel.every(item => item.source_endpoint !== '' && item.target_endpoint !== '')
          if (hasEmpty) {
            this.addAndEditModal.dataConfig.endpoint_rel.push(
              {source_endpoint: '', target_endpoint: ''}
            )
          } else {
            this.$Message.warning('Can Not Empty')
          }
          break
        }
      }
    },
    deleteItem(type, index) {
      switch (type) {
        case 'path': {
          this.addAndEditModal.pathOptions.splice(index, 1)
          break
        }
        case 'relate': {
          this.addAndEditModal.dataConfig.endpoint_rel.splice(index, 1)
          break
        }
      }
    },
    async add () {
      this.cancelAddAndEdit()
      this.addAndEditModal.isAdd = true
      this.addAndEditModal.isShow = true
    },
    getDefaultConfig (val, type) {
      const api = `/monitor/api/v2/service/service_group/endpoint_rel?serviceGroup=${this.targrtId}&sourceType=${type}&targetType=${val}`
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', api, '', (responseData) => {
        const tmp = responseData.map(r => {
            return {
              source_endpoint: r.source_endpoint,
              target_endpoint: r.target_endpoint
            }
          })
        if (type === 'host') {
          tmp.forEach(t => {
            const find = this.addAndEditModal.dataConfig.endpoint_rel.find(rel => rel.source_endpoint === t.source_endpoint && rel.target_endpoint === t.target_endpoint)
            if (find === undefined) {
              this.addAndEditModal.dataConfig.endpoint_rel.push(t)
            }
          })
        }
      })
    },
    getDetail (targrtId) {
      this.targrtId = targrtId
      const api = `/monitor/api/v2/service/log_keyword/list/service/${targrtId}`
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', api, '', (responseData) => {
        this.showManagement = true
        this.targetDetail = responseData[0]
        this.pageConfig.table.tableData = responseData[0].config
        this.$root.$store.commit('changeTableExtendActive', -1)
      }, {isNeedloading:true})
    }
  },
  components: {
    extendTable
  },
}
</script>

<style>
.ivu-form-item {
  margin-bottom: 4px;
}
.success-btn {
  color: #fff;
  background-color: #19be6b;
  border-color: #19be6b;
}
</style>
