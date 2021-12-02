<template>
  <div class=" ">
    <section v-if="showManagement" style="margin-top: 16px;">
      <Tag color="blue">{{targetDetail.display_name}}</Tag>
      <button @click="add" type="button" class="btn btn-sm btn-cancel-f">
        <i class="fa fa-plus"></i>
        {{$t('button.add')}}
      </button>
      <PageTable :pageConfig="pageConfig">
        <div slot='tableExtend'>
          <div style="margin:8px;border:1px solid #2db7f5">
            <button @click="singleAddF(pageConfig.table.isExtend.parentData)" type="button" style="margin-top:8px" class="btn btn-sm btn-cancel-f">
              <i class="fa fa-plus"></i>
              {{$t('m_add_json_regular')}}
            </button>
            <extendTable :detailConfig="pageConfig.table.isExtend.detailConfig"></extendTable>
          </div>
          <div style="margin:8px;border:1px solid #19be6b">
            <button @click="addCustomMetric(pageConfig.table.isCustomMetricExtend.parentData)" type="button" style="margin-top:8px" class="btn btn-sm btn-cancel-f">
              <i class="fa fa-plus"></i>
              {{$t('m_add_metric_regular')}}
            </button>
            <extendTable :detailConfig="pageConfig.table.isCustomMetricExtend.detailConfig"></extendTable>
          </div>
        </div>
      </PageTable>
    </section>
    <Modal
      v-model="addAndEditModal.isShow"
      :title="$t('button.add')"
      >
      <div>
        <div>
          <span>{{$t('field.type')}}:</span>
          <Select v-model="addAndEditModal.dataConfig.monitor_type" @on-change="getEndpoint" style="width: 445px">
            <Option v-for="type in addAndEditModal.monitorTypeOptions" :key="type.value" :value="type.label">{{type.label}}</Option>
          </Select>
        </div>
        <div style="margin: 4px 0px;padding:8px 12px;border:1px solid #dcdee2;border-radius:4px">
          <template v-for="(item, index) in addAndEditModal.pathOptions">
            <p :key="index + 5">
              <Button
                @click="deleteItem('path', index)"
                size="small"
                style="background-color: #ff9900;border-color: #ff9900;"
                type="error"
                icon="md-close"
              ></Button>
              <Input v-model="item.path" style="width: 432px" :placeholder="$t('tableKey.path')" />
            </p>
          </template>
          <Button
            @click="addEmptyItem('path')"
            type="success"
            size="small"
            style="background-color: #0080FF;border-color: #0080FF;"
            long
            >{{ $t('button.add') }}{{$t('tableKey.path')}}</Button
          >
        </div>
        <div style="margin: 4px 0px;padding:8px 12px;border:1px solid #dcdee2;border-radius:4px">
          <template v-for="(item, index) in addAndEditModal.dataConfig.endpoint_rel">
            <p :key="index + 'c'">
              <Button
                @click="deleteItem('relate', index)"
                size="small"
                style="background-color: #ff9900;border-color: #ff9900;"
                type="error"
                icon="md-close"
              ></Button>
              <Select v-model="item.source_endpoint" style="width: 215px" :placeholder="$t('m_log_server')">
                <Option v-for="type in addAndEditModal.sourceEndpoints" :key="type.guid" :value="type.guid">{{type.display_name}}</Option>
              </Select>
              <Select v-model="item.target_endpoint" style="width: 215px" :placeholder="$t('m_business_object')">
                <Option v-for="type in addAndEditModal.targetEndpoints" :key="type.guid" :value="type.guid">{{type.display_name}}</Option>
              </Select>
            </p>
          </template>
          <Button
            @click="addEmptyItem('relate')"
            type="success"
            size="small"
            style="background-color: #0080FF;border-color: #0080FF;"
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
      v-model="ruleModelConfig.isShow"
      :title="$t('m_json_regular')"
      width="840"
      >
      <div :style="{ 'max-height': MODALHEIGHT + 'px', overflow: 'auto' }">
        <Form :label-width="100">
          <FormItem :label="$t('tableKey.regular')">
            <Input v-model="ruleModelConfig.addRow.json_regular" style="width:100%"/>
          </FormItem>
          <FormItem :label="$t('tableKey.tags')">
            <Input v-model="ruleModelConfig.addRow.tags" style="width:100%" />
          </FormItem>
        </Form>
        <div style="margin: 4px 0px;padding:8px 12px;border:1px solid #dcdee2;border-radius:4px">
          <template v-for="(item, index) in ruleModelConfig.addRow.metric_list">
            <p :key="index + 3">
              <Button
                @click="deleteItem('metric_list', index)"
                size="small"
                style="background-color: #ff9900;border-color: #ff9900;"
                type="error"
                icon="md-close"
              ></Button>
              <Input v-model="item.json_key" style="width: 190px" :placeholder="$t('m_key') + ' e.g:[.*][.*]'" />
              <Input v-model="item.metric" style="width: 190px" :placeholder="$t('field.metric') + ' , e.g:code'" />
              <Select v-model="item.agg_type" filterable clearable style="width:190px">
                <Option v-for="agg in ruleModelConfig.aggOption" :value="agg" :key="agg">{{
                  agg
                }}</Option>
              </Select>
              <Input v-model="item.display_name" style="width: 160px" :placeholder="$t('tableKey.description')" />
            </p>
            <div :key="index + 1" style="margin: 4px 0px;padding:8px 12px;border:1px solid #dcdee2;border-radius:4px;text-align: end;">
              <template v-for="(stringMapItem, stringMapIndex) in item.string_map">
                <p :key="stringMapIndex + 2">
                  <Button
                    @click="deleteItem('string_map', index)"
                    size="small"
                    style="background-color: #ff9900;border-color: #ff9900;"
                    type="error"
                    icon="md-close"
                  ></Button>
                  <Select v-model="stringMapItem.regulative" filterable clearable style="width:230px">
                    <Option v-for="regulation in regulationOption" :value="regulation.value" :key="regulation.value">{{
                      regulation.label
                    }}</Option>
                  </Select>
                  <Input v-model="stringMapItem.source_value" style="width: 230px" :placeholder="$t('m_log_server')" />
                  <Input v-model="stringMapItem.target_value" style="width: 230px" :placeholder="$t('m_business_object')" />
                </p>
              </template>
              <Button
                @click="addEmptyItem('string_map', index)"
                type="success"
                size="small"
                style="background-color: #19be6b;border-color: #19be6b;"
                >{{ $t('addStringMap') }}</Button
              >
            </div>
            <Divider :key="index + 'Q'" />
          </template>
          <Button
            @click="addEmptyItem('metric_list')"
            type="success"
            size="small"
            style="background-color: #0080FF;border-color: #0080FF;"
            long
            >{{ $t('addMetricConfig') }}</Button
          >
        </div>
      </div>
      <div slot="footer">
        <Button @click="cancelRule">{{$t('button.cancel')}}</Button>
        <Button @click="saveRule" type="primary">{{$t('button.save')}}</Button>
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
          <label class="col-md-2 label-name">{{$t('field.aggType')}}:</label>
          <Select v-model="customMetricsModelConfig.addRow.agg_type" filterable clearable style="width:375px">
            <Option v-for="agg in customMetricsModelConfig.slotConfig.aggOption" :value="agg" :key="agg">{{
              agg
            }}</Option>
          </Select>
        </div>
        <div class="marginbottom params-each">
          <div style="margin: 4px 12px;padding:8px 12px;border:1px solid #dcdee2;border-radius:4px">
            <template v-for="(item, index) in customMetricsModelConfig.addRow.string_map">
              <p :key="index">
                <Button
                  @click="deleteCustomMetric('string_map', index)"
                  size="small"
                  style="background-color: #ff9900;border-color: #ff9900;"
                  type="error"
                  icon="md-close"
                ></Button>
                <Select v-model="item.regulative" filterable clearable style="width:150px">
                  <Option v-for="regulation in regulationOption" :value="regulation.value" :key="regulation.value">{{
                    regulation.label
                  }}</Option>
                </Select>
                <Input v-model="item.source_value" style="width: 150px" :placeholder="$t('m_log_server')" />
                <Input v-model="item.target_value" style="width: 150px" :placeholder="$t('m_business_object')" />
              </p>
            </template>
            <Button
              @click="addCustomMetricEmpty('string_map')"
              type="success"
              size="small"
              style="background-color: #0080FF;border-color: #0080FF;"
              long
              >{{ $t('addStringMap') }}</Button
            >
          </div>
        </div>
      </div>
    </ModalComponent>
  </div>
</template>

<script>
import extendTable from '@/components/table-page/extend-table'
let tableEle = [
  {title: 'tableKey.path', value: 'log_path', display: true},
  {title: 'field.type', value: 'monitor_type', display: true},
]
const btn = [
  {btn_name: 'button.remove', btn_func: 'deleteConfirmModal'}
]
export default {
  name: '',
  data () {
    return {
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
                {title: 'tableKey.tags', value: 'tags', display: true},
                {title: 'table.action',btn:[
                  {btn_name: 'button.edit', btn_func: 'editRuleItem'},
                  {btn_name: 'button.remove', btn_func: 'delRuleconfirmModal'}
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
                {title: 'tableKey.regular', value: 'regular', display: true},
                {title: 'field.metric', value: 'metric', display: true},
                {title: 'field.aggType', value: 'agg_type', display: true},
                {title: 'table.action',btn:[
                  {btn_name: 'button.edit', btn_func: 'editCustomMetricItem'},
                  {btn_name: 'button.remove', btn_func: 'delCustomMetricConfirmModal'}
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
          log_path: [],
          monitor_type: '',
          endpoint_rel: []
        },
        pathOptions: [],
        monitorTypeOptions: [
          {label: 'process', value: 'process'},
          {label: 'java', value: 'java'},
          {label: 'nginx', value: 'nginx'},
          {label: 'http', value: 'http'}
        ],
        sourceEndpoints: [],
        targetEndpoints: [],
      },
      showAddAndEditModal: false,
      activeData: {},
      regulationOption: [
        {label: this.$t('m_regular_match'), value: 1},
        {label: this.$t('m_irregular_matching'), value: 0}
      ],
      ruleModelConfig: {
        isShow: false,
        isAdd: true,
        addRow: {
          log_metric_monitor: null,
          json_regular: null,
          tags: null,
          metric_list: []
        },
        aggOption: ['sum', 'avg', 'count', 'max', 'min']
      },
      selectedData: null,
      selectedIndex: null,
      isShowWarningDelete: false,
      deleteType: '',
      customMetricsModelConfig: {
        modalId: 'custom_metrics',
        isAdd: true,
        noBtn: true,
        modalStyle: 'min-width:550px',
        modalTitle: 'm_metric_regular',
        saveFunc: 'saveCustomMetric',
        config: [
          {label: 'field.metric', value: 'metric', placeholder: '', disabled: false, type: 'text'},
          {label: 'tableKey.description', value: 'display_name', placeholder: '', disabled: false, type: 'text'},
          {label: 'tableKey.regular', value: 'regular', placeholder: 'tips.required', v_validate: 'required:true', disabled: false, type: 'text'},
          {name:'ruleConfig',type:'slot'}
        ],
        addRow: { // [通用]-保存用户新增、编辑时数据
          log_metric_monitor: '',
          display_name: '',
          agg_type: 'min',
          metric: null,
          regular: '',
          string_map: []
        },
        slotConfig: {
          aggOption: ['sum', 'avg', 'count', 'max', 'min'],
          regulationOption: [
            {label: this.$t('m_regular_match'), value: 1},
            {label: this.$t('m_irregular_matching'), value: 0}
          ]
        }
      },
      modelTip: {
        key: '',
        value: 'metric'
      },
    }
  },
  mounted () {
    this.MODALHEIGHT = document.body.scrollHeight - 300
  },
  methods: {
    addCustomMetricEmpty (type) {
      if (!this.customMetricsModelConfig.addRow[type]) {
        this.customMetricsModelConfig.addRow[type] = []
      }
      this.customMetricsModelConfig.addRow[type].push({
        source_value: '',
        regulative: 1,
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
      params.log_metric_monitor = this.activeData.guid
      const requestType = this.customMetricsModelConfig.isAdd ? 'POST' : 'PUT'
      this.$root.$httpRequestEntrance.httpRequestEntrance(requestType, this.$root.apiCenter.logMetricReg, params, () => {
        this.$Message.success(this.$t('tips.success'))
        this.$root.JQ('#custom_metrics').modal('hide')
        this.getDetail(this.targrtId)
      })
    },
    editCustomMetricItem (rowData) {
      this.customMetricsModelConfig.isAdd = false
      this.modelTip.value = rowData.display_name
      this.customMetricsModelConfig.addRow = rowData
      this.$root.JQ('#custom_metrics').modal('show')
    },
    delCustomMetricConfirmModal (rowData) {
      this.selectedData = rowData
      this.isShowWarningDelete = true
      this.deleteType = 'custom_metrics'
    },
    editRuleItem (rowData) {
      this.ruleModelConfig.isAdd = false
      this.ruleModelConfig.addRow = rowData
      this.ruleModelConfig.isShow = true
    },
    delRuleconfirmModal (rowData) {
      this.selectedData = rowData
      this.isShowWarningDelete = true
      this.deleteType = 'rules'
    },
    okDelRow () {
      if (this.deleteType === 'custom_metrics') {
        this.delCustomMericsItem(this.selectedData)
      } else {
        this.delRuleItem(this.selectedData)
      }
    },
    delCustomMericsItem (rowData) {
      const api = this.$root.apiCenter.logMetricReg + '/' + rowData.guid
      this.$root.$httpRequestEntrance.httpRequestEntrance('DELETE', api, '', () => {
        this.$Message.success(this.$t('tips.success'))
        this.getDetail(this.targrtId)
      })
    },
    cancleDelRow () {
      this.isShowWarningDelete = false
    },
    delRuleItem (rowData) {
      const api = this.$root.apiCenter.logMetricJson + '/' + rowData.guid
      this.$root.$httpRequestEntrance.httpRequestEntrance('DELETE', api, '', () => {
        this.$Message.success(this.$t('tips.success'))
        this.isShowWarningDelete = false
        this.getDetail(this.targrtId)
      })
    },
    cancelRule () {
      this.ruleModelConfig.addRow = {
        log_metric_monitor: null,
        json_regular: null,
        tags: null,
        metric_list: []
      }
      this.ruleModelConfig.isShow = false
    },
    saveRule () {
      this.ruleModelConfig.addRow.log_metric_monitor = this.activeData.guid
      const requestType = this.ruleModelConfig.isAdd ? 'POST' : 'PUT'
      this.$root.$httpRequestEntrance.httpRequestEntrance(requestType, this.$root.apiCenter.logMetricJson, this.ruleModelConfig.addRow, () => {
        this.$Message.success(this.$t('tips.success'))
        this.ruleModelConfig.isShow = false
        this.getDetail(this.targrtId)
      })
    },
    singleAddF (rowData) {
      this.cancelRule()
      this.activeData = rowData
      this.ruleModelConfig.isAdd = true
      this.ruleModelConfig.isShow = true
    },
    getExtendInfo(item){
      item.json_config_list.forEach(xx => xx.pId = item.guid)
      this.pageConfig.table.isExtend.detailConfig[0].data = item.json_config_list
      this.pageConfig.table.isExtend.parentData = item

      item.metric_config_list.forEach(xx => xx.pId = item.guid)
      this.pageConfig.table.isCustomMetricExtend.detailConfig[0].data = item.metric_config_list
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
      const api = this.$root.apiCenter.deletePath + '/' + rowData.guid
      this.$root.$httpRequestEntrance.httpRequestEntrance('DELETE', api, '', () => {
        this.$Message.success(this.$t('tips.success'))
        this.getDetail(this.targrtId)
      })
    },
    okAddAndEdit () {
      this.addAndEditModal.dataConfig.service_group = this.targrtId
      this.addAndEditModal.dataConfig.log_path = this.addAndEditModal.pathOptions.map(p => p.path)
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.logMetricMonitor, this.addAndEditModal.dataConfig, () => {
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
    getEndpoint (val) {
      // get source Endpoint
      const sourceApi = this.$root.apiCenter.getEndpointsByType + '/' + this.targrtId + '/endpoint/host'
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', sourceApi, '', (responseData) => {
        this.addAndEditModal.sourceEndpoints = responseData
      }, {isNeedloading:false})
      const targetApi = this.$root.apiCenter.getEndpointsByType + '/' + this.targrtId + '/endpoint/' + val
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', targetApi, '', (responseData) => {
        this.addAndEditModal.targetEndpoints = responseData
      }, {isNeedloading:false})

    },
    addEmptyItem (type, index) {
      if (type === 'path') {
        const hasEmpty = this.addAndEditModal.pathOptions.every(item => item.path !== '')
        if (hasEmpty) {
          this.addAndEditModal.pathOptions.push(
            {path: ''}
          )
        } else {
          this.$Message.warning('Path Can Not Empty')
        }
      }
      if (type === 'relate') {
        const hasEmpty = this.addAndEditModal.dataConfig.endpoint_rel.every(item => item.source_endpoint !== '' && item.target_endpoint !== '')
        if (hasEmpty) {
          this.addAndEditModal.dataConfig.endpoint_rel.push(
            {source_endpoint: '', target_endpoint: ''}
          )
        } else {
          this.$Message.warning('Can Not Empty')
        }
      }
      if (type === 'metric_list') {
        this.ruleModelConfig.addRow[type].push({
          json_key: '',
          display_name: '',
          metric: '',
          agg_type: 'avg',
          string_map: []
        })
      }
      if (type === 'string_map') {
        this.ruleModelConfig.addRow.metric_list[index][type].push({
          source_value: '',
          regulative: 1,
          target_value: ''
        })
      }
      
    },
    deleteItem(type, index) {
      if (type === 'path') {
        this.addAndEditModal.pathOptions.splice(index, 1)
      }
      if (type === 'relate') {
        this.addAndEditModal.dataConfig.endpoint_rel.splice(index, 1)
      }
      if (type === 'metric_list') {
        this.ruleModelConfig.addRow[type].splice(index, 1)
      }
      if (type === 'string_map') {
        this.ruleModelConfig.addRow.metric_list[index][type].splice(index, 1)
      }
    },
    add () {
      this.cancelAddAndEdit()
      this.addAndEditModal.isAdd = true
      this.addAndEditModal.isShow = true
    },
    getDetail (targrtId) {
      this.targrtId = targrtId
      const api = this.$root.apiCenter.getTargetDetail + '/group/' + targrtId
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', api, '', (responseData) => {
        this.showManagement = true
        this.targetDetail = responseData
        this.pageConfig.table.tableData = responseData.config
        this.$root.$store.commit('changeTableExtendActive', -1)
      }, {isNeedloading:false})
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
</style>
