<template>
  <div class=" ">
    <section v-if="showManagement" style="margin-top: 16px;">
      <Tag color="blue">{{$t('m_log_file')}}</Tag>
      <PageTable :pageConfig="pageConfig">
        <div slot='tableExtend'>
          <div style="margin:8px;border:1px solid #19be6b">
            <extendTable :detailConfig="pageConfig.table.isCustomMetricExtend.detailConfig"></extendTable>
          </div>
        </div>
      </PageTable>
    </section>
    <Modal
      v-model="addAndEditModal.isShow"
      :title="$t('m_button_view')"
      :width="730"
    >
      <div :style="{'max-height': MODALHEIGHT + 'px', overflow: 'auto'}">
        <div>
          <span>{{$t('m_field_type')}}:</span>
          <Select v-model="addAndEditModal.dataConfig.monitor_type" disabled @on-change="getEndpoint(addAndEditModal.dataConfig.monitor_type, 'host')" style="width: 640px">
            <Option v-for="type in monitorTypeOptions" :key="type.value" :value="type.label">{{type.label}}</Option>
          </Select>
        </div>
        <div v-if="addAndEditModal.isAdd" disabled style="margin: 4px 0px;padding:8px 12px;border:1px solid #dcdee2;border-radius:4px;width:680px">
          <template v-for="(item, index) in addAndEditModal.pathOptions">
            <p :key="index + 5">
              <Tooltip :content="$t('m_tableKey_logPath')" :delay="1000">
                <Input v-model="item.path" style="width: 620px" :placeholder="$t('m_tableKey_logPath')" />
              </Tooltip>
            </p>
          </template>
        </div>
        <div v-else style="margin: 8px 0">
          <span>{{$t('m_tableKey_path')}}:</span>
          <Input style="width: 640px" disabled v-model="addAndEditModal.dataConfig.log_path" />
        </div>
        <div style="margin: 4px 0px;padding:8px 12px;border:1px solid #dcdee2;border-radius:4px;width:680px">
          <template v-for="(item, index) in addAndEditModal.dataConfig.endpoint_rel">
            <p :key="index + 'c'">
              <Tooltip :content="$t('m_business_object')" :delay="1000">
                <Select v-model="item.target_endpoint" disabled style="width: 310px" :placeholder="$t('m_business_object')">
                  <Option v-for="type in targetEndpoints" :key="type.guid" :value="type.guid">{{type.display_name}}</Option>
                </Select>
              </Tooltip>
              <Tooltip :content="$t('m_log_server')" :delay="1000">
                <Select v-model="item.source_endpoint" disabled style="width: 310px" :placeholder="$t('m_log_server')">
                  <Option v-for="type in sourceEndpoints" :key="type.guid" :value="type.guid">{{type.display_name}}</Option>
                </Select>
              </Tooltip>
            </p>
          </template>
        </div>
      </div>
      <div slot="footer">
        <Button @click="cancelAddAndEdit">{{$t('m_button_cancel')}}</Button>
      </div>
    </Modal>
    <ModalComponent :modelConfig="customMetricsModelConfig">
      <div slot="ruleConfig" class="extentClass">
        <div class="marginbottom params-each">
          <label class="col-md-2 label-name">{{$t('m_regular')}}:</label>
          <Select v-model="customMetricsModelConfig.addRow.regulative" disabled filterable clearable style="width:375px">
            <Option v-for="notify in customMetricsModelConfig.notifyEnableOption" :value="notify.value" :key="notify.value">{{
              notify.label
            }}</Option>
          </Select>
        </div>
        <div class="marginbottom params-each">
          <label class="col-md-2 label-name">{{$t('m_tableKey_s_priority')}}:</label>
          <Select filterable clearable disabled v-model="customMetricsModelConfig.addRow.priority" style="width:375px">
            <Option v-for="item in customMetricsModelConfig.priorityList" :value="item.value" :key="item.value">{{ $t(item.label) }}</Option>
          </Select>
        </div>
        <div class="marginbottom params-each">
          <label class="col-md-2 label-name">{{$t('sendAlarm')}}:</label>
          <Select filterable clearable disabled v-model="customMetricsModelConfig.addRow.notify_enable" style="width:375px">
            <Option v-for="item in customMetricsModelConfig.notifyEnableOption" :value="item.value" :key="item.value">{{ item.label }}</Option>
          </Select>
        </div>
      </div>
      <div slot="btn">
        <Button style="float:right" @click="cancelModal">{{$t('m_button_cancel')}}</Button>
      </div>
    </ModalComponent>
  </div>
</template>

<script>
import {priorityList} from '@/assets/config/common-config.js'
import extendTable from '@/components/table-page/extend-table'
const tableEle = [
  {
    title: 'm_tableKey_logPath',
    value: 'log_path',
    display: true
  },
  {
    title: 'm_field_type',
    value: 'monitor_type',
    display: true
  },
]
const btn = [
  {
    btn_name: 'm_button_view',
    btn_func: 'editF'
  }
]

export default {
  name: '',
  data() {
    return {
      MODALHEIGHT: 300,
      isShowWarning: false,
      targrtId: '',
      targetDetail: {},
      showManagement: false,
      pageConfig: {
        table: {
          tableData: [],
          tableEle,
          // filterMoreBtn: 'filterMoreBtn',
          primaryKey: 'id',
          btn,
          handleFloat: true,
          isExtend: {
            parentData: null,
            func: 'getExtendInfo',
            data: {},
            slot: 'tableExtend',
            detailConfig: [{
              isExtendF: true,
              title: '',
              config: [
                {
                  title: 'm_tableKey_regular',
                  value: 'json_regular',
                  display: true
                },
                {
                  title: 'm_collection_interval',
                  value: 'step',
                  display: true
                },
                {
                  title: 'm_tableKey_tags',
                  value: 'tags',
                  display: true
                },
                {
                  title: 'm_table_action',
                  btn: [
                    {
                      btn_name: 'm_button_edit',
                      btn_func: 'editRuleItem'
                    }
                  ]
                }
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
                {
                  title: 'm_field_log',
                  value: 'keyword',
                  display: true
                },
                {
                  title: 'sendAlarm',
                  value: 'notify_enable',
                  display: true
                },
                {
                  title: 'm_tableKey_s_priority',
                  value: 'priority',
                  display: true
                },
                {
                  title: 'm_regular',
                  value: 'regulative',
                  display: true
                },
                {
                  title: 'm_table_action',
                  btn: [
                    {
                      btn_name: 'm_button_view',
                      btn_func: 'editCustomMetricItem'
                    },
                  ]
                }
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
          log_path: [],
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
        modalTitle: 'm_field_log',
        saveFunc: 'saveCustomMetric',
        noBtn: true,
        config: [
          {
            label: 'm_tableKey_keyword',
            value: 'keyword',
            placeholder: '',
            disabled: true,
            type: 'text'
          },
          {
            name: 'ruleConfig',
            type: 'slot'
          },
          {
            name: 'btn',
            type: 'slot'
          }
        ],
        addRow: { // [通用]-保存用户新增、编辑时数据
          log_keyword_monitor: '',
          keyword: '',
          regulative: 0,
          notify_enable: 1,
          priority: 'low'
        },
        priorityList,
        notifyEnableOption: [
          {
            label: 'Yes',
            value: 1
          },
          {
            label: 'No',
            value: 0
          }
        ]
      },
      modelTip: {
        key: '',
        value: 'keyword'
      },
      monitorTypeOptions: [
        {
          label: 'process',
          value: 'process'
        },
        {
          label: 'java',
          value: 'java'
        },
        {
          label: 'nginx',
          value: 'nginx'
        },
        {
          label: 'http',
          value: 'http'
        },
        {
          label: 'mysql',
          value: 'mysql'
        }
      ],
      service_group: ''
    }
  },
  mounted() {
    this.MODALHEIGHT = document.body.scrollHeight - 300
  },
  methods: {
    // other config
    cancelModal() {
      this.$root.JQ('#custom_metrics').modal('hide')
    },
    editF(rowData) {
      this.service_group = rowData.service_group
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
    addCustomMetricEmpty(type) {
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
    addCustomMetric(rowData) {
      this.activeData = rowData
      this.customMetricsModelConfig.isAdd = true
      this.$root.JQ('#custom_metrics').modal('show')
    },
    saveCustomMetric() {
      const params = JSON.parse(JSON.stringify(this.customMetricsModelConfig.addRow))
      params.log_keyword_monitor = this.activeData.guid
      const requestType = this.customMetricsModelConfig.isAdd ? 'POST' : 'PUT'
      this.$root.$httpRequestEntrance.httpRequestEntrance(requestType, '/monitor/api/v2/service/log_keyword/log_keyword_config', params, () => {
        this.$Message.success(this.$t('m_tips_success'))
        this.$root.JQ('#custom_metrics').modal('hide')
        this.getDetail(this.targrtId)
      })
    },
    editCustomMetricItem(rowData) {
      this.customMetricsModelConfig.isAdd = false
      this.modelTip.value = rowData.keyword
      this.customMetricsModelConfig.addRow = JSON.parse(JSON.stringify(rowData))
      this.$root.JQ('#custom_metrics').modal('show')
    },
    delCustomMetricConfirmModal(rowData) {
      this.selectedData = rowData
      this.isShowWarningDelete = true
      this.deleteType = 'custom_metrics'
    },
    okDelRow() {
      if (this.deleteType === 'custom_metrics') {
        this.delCustomMericsItem(this.selectedData)
      }
    },
    delCustomMericsItem(rowData) {
      const api = '/monitor/api/v2/service/log_keyword/log_keyword_config/' + rowData.guid
      this.$root.$httpRequestEntrance.httpRequestEntrance('DELETE', api, '', () => {
        this.$Message.success(this.$t('m_tips_success'))
        this.getDetail(this.targrtId)
      })
    },
    cancleDelRow() {
      this.isShowWarningDelete = false
    },
    getExtendInfo(item){
      item.keyword_list.forEach(xx => xx.pId = item.guid)
      this.pageConfig.table.isCustomMetricExtend.detailConfig[0].data = item.keyword_list
      this.pageConfig.table.isCustomMetricExtend.parentData = item
    },
    deleteConfirmModal(rowData) {
      this.selectedData = rowData
      this.isShowWarning = true
    },
    ok() {
      this.delF(this.selectedData)
    },
    cancel() {
      this.isShowWarning = false
    },
    delF(rowData) {
      const api = '/monitor/api/v2/service/log_keyword/log_keyword_monitor' + '/' + rowData.guid
      this.$root.$httpRequestEntrance.httpRequestEntrance('DELETE', api, '', () => {
        this.$Message.success(this.$t('m_tips_success'))
        this.getDetail(this.targrtId)
      })
    },
    okAddAndEdit() {
      const params = JSON.parse(JSON.stringify(this.addAndEditModal.dataConfig))
      const methodType = this.addAndEditModal.isAdd ? 'POST' : 'PUT'
      params.service_group = this.targrtId
      if (this.addAndEditModal.isAdd) {
        params.log_path = this.addAndEditModal.pathOptions.map(p => p.path)
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance(methodType, '/monitor/api/v2/service/log_keyword/log_keyword_monitor', params, () => {
        this.$Message.success(this.$t('m_tips_success'))
        this.addAndEditModal.isShow = false
        this.getDetail(this.targrtId)
      }, {isNeedloading: false})
    },
    cancelAddAndEdit() {
      this.addAndEditModal.isShow = false
      this.addAndEditModal.pathOptions = []
      this.addAndEditModal.dataConfig = {
        service_group: '',
        log_path: [],
        monitor_type: '',
        endpoint_rel: []
      }
    },
    async getEndpoint(val, type) {
      this.addAndEditModal.dataConfig.endpoint_rel = []
      await this.getDefaultConfig(val, type)
      // get source Endpoint
      const sourceApi = this.$root.apiCenter.getEndpointsByType + '/' + this.service_group + '/endpoint/' + type
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', sourceApi, '', responseData => {
        this.sourceEndpoints = responseData
      }, {isNeedloading: false})
      const targetApi = this.$root.apiCenter.getEndpointsByType + '/' + this.service_group + '/endpoint/' + val
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', targetApi, '', responseData => {
        this.targetEndpoints = responseData
      }, {isNeedloading: false})
    },
    addEmptyItem(type) {
      switch (type) {
        case 'path': {
          const hasEmpty = this.addAndEditModal.pathOptions.every(item => item.path !== '')
          if (hasEmpty) {
            this.addAndEditModal.pathOptions.push(
              {path: ''}
            )
          }
          else {
            this.$Message.warning('Path Can Not Empty')
          }
          break
        }
        case 'relate': {
          const hasEmpty = this.addAndEditModal.dataConfig.endpoint_rel.every(item => item.source_endpoint !== '' && item.target_endpoint !== '')
          if (hasEmpty) {
            this.addAndEditModal.dataConfig.endpoint_rel.push(
              {
                source_endpoint: '',
                target_endpoint: ''
              }
            )
          }
          else {
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
    async add() {
      this.cancelAddAndEdit()
      this.addAndEditModal.isAdd = true
      this.addAndEditModal.isShow = true
    },
    getDefaultConfig(val, type) {
      const api = `/monitor/api/v2/service/service_group/endpoint_rel?serviceGroup=${this.service_group}&sourceType=${type}&targetType=${val}`
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', api, '', responseData => {
        const tmp = responseData.map(r => ({
          source_endpoint: r.source_endpoint,
          target_endpoint: r.target_endpoint
        }))
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
    getDetail(targrtId) {
      this.targrtId = targrtId
      this.targetDetail = []
      this.pageConfig.table.tableData = []
      const api = `/monitor/api/v2/service/log_keyword/list/endpoint/${targrtId}`
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', api, '', responseData => {
        this.showManagement = true
        if (responseData.length > 0) {
          this.targetDetail = responseData[0]
          this.pageConfig.table.tableData = responseData[0].config
        }
        this.$root.$store.commit('changeTableExtendActive', -1)
      }, {isNeedloading: true})
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
