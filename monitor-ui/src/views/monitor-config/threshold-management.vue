<template>
  <div class=" ">
    <section>
      <ul class="search-ul">
        <li class="search-li">
          <Select filterable clearable v-model="type" style="width:100px" @on-change="typeChange">
            <Option v-for="item in typeList" :value="item.value" :key="item.value">{{ $t(item.label) }}</Option>
          </Select>
        </li>
        <li class="search-li">
          <Select
            style="width:300px"
            v-model="endpointID"
            filterable
            remote
            clearable
            :remote-method="getEndpointList"
            @on-clear="clearEndpoint"
            >
            <Option v-for="(option, index) in endpointOptions" :value="option.id" :key="index">
            <TagShow v-if="type!== 'grp'" :tagName="option.option_type_name" :index="index"></TagShow> 
            {{option.option_text}}
            </Option>
          </Select>
        </li>
        <li class="search-li">
          <button type="button" class="btn btn-sm btn-confirm-f"
          @click="search">
            <i class="fa fa-search" ></i>
            {{$t('button.search')}}
          </button>
        </li>
      </ul> 
    </section>
    <section>
      <template v-for="(tableItem, tableIndex) in totalPageConfig">
        <section :key="tableIndex + 'f'">
          <div class="section-table-tip">
            <Tag color="blue" :key="tableIndex + 'a'" v-if="tableItem.obj_name">{{tableItem.obj_name}}</Tag>
            <button @click="add" type="button" v-if="tableItem.operation" class="btn btn-sm btn-cancel-f" :key="tableIndex + 'b'">
              <i class="fa fa-plus"></i>
              {{$t('button.add')}}
            </button>
          </div>
          <PageTable :pageConfig="tableItem"></PageTable>
          <section class="receiver-config">
            <section :key="tableIndex + 'a'" v-if="tableItem.showReceiver" style="margin: 16px 0">
              <span class="receiver-header">{{$t('button.receivers')}}:</span>
              <Tag 
                v-for="(receiver, receiverIdex) in tableItem.accept" 
                type="border" 
                :key="receiverIdex" 
                color="primary">
                {{receiver.option_text}}
              </Tag>
              <button @click="tableItem.showReceiver = !tableItem.showReceiver" class="btn btn-small btn-cancel-f">{{$t('button.edit')}}</button>
            </section>
            <div v-else  style="margin: 16px 0">
              <h5>{{$t('button.receiversConfiguration')}}:</h5>
              <div class="receiver-config-set" style="margin: 8px 0">
                <div>
                  <span style="margin: 0 16px">{{$t('button.receiversSelect')}}:</span>
                    <Select
                    style="width:200px"
                    v-model="tableItem.selectedReceiver"
                    @on-open-change="getSelectReceivers('r')"
                    filterable clearable
                    multiple
                    remote
                    :remote-method="getSelectReceivers"
                    >
                    <Option v-for="(option, index) in selectReceiversOptions" :value="option.type" :key="index">
                    {{option.option_text}}</Option>
                  </Select>
                </div>
                <div>
                  <span style="margin: 0 8px">{{$t('button.receiversInput')}}:</span>
                  <input type="text" v-model.trim="tableItem.inputReceiver" :placeholder="$t('button.receiversInputTip')" class="form-control research-input c-dark">
                </div>
                <div>
                  <button @click="saveReceivers(tableItem)" class="btn btn-confirm-f">{{$t('button.save')}}</button>
                </div>
              </div>
              <div class="receiver-config-set"  style="margin: 8px 0">
                <Tag 
                  v-for="(receiver, receiverIdex) in tableItem.accept" 
                  type="border" :key="receiverIdex" 
                  closable
                  @on-close="removeReceiver(tableItem,receiver,receiverIdex)"
                  color="primary">
                  {{receiver.option_text}}
                </Tag>
              </div>
            </div>
            <div class="partition"></div>
          </section>
          
        </section>
      </template>
      <ModalComponent :modelConfig="modelConfig">
        <div slot="metricSelect" class="extentClass">  
          <div class="marginbottom params-each">
            <label class="col-md-2 label-name">{{$t('tableKey.metricName')}}:</label>
            <Select v-model="modelConfig.addRow.expr" filterable clearable style="width:340px"
            :label-in-value="true" @on-change="selectMetric">
              <Option v-for="(item, index) in modelConfig.metricList" :value="item.prom_ql" :key="item.prom_ql+item.metric+index">{{ item.metric }}</Option>
            </Select>
          </div> 
        </div>
        <div slot="thresholdConfig" class="extentClass">  
          <div class="marginbottom params-each">
            <label class="col-md-2 label-name">{{$t('field.threshold')}}:</label>
            <Select filterable clearable v-model="modelConfig.threshold" style="width:100px">
              <Option v-for="(item, index) in modelConfig.thresholdList" :value="item.value" :key="index">{{ item.label }}</Option>
            </Select>
            <div class="search-input-content" style="margin-left: 8px">
              <input 
                v-validate="'required|isNumber'" 
                v-model="modelConfig.thresholdValue" 
                name="thresholdValue"
                :class="{ 'red-border': veeErrors.has('thresholdValue') }"
                type="text" 
                class="form-control model-input search-input c-dark"/>
              <label class="required-tip">*</label>
            </div>
            <div style="margin-left:120px">
              <label v-show="veeErrors.has('thresholdValue')" class="is-danger">{{ veeErrors.first('thresholdValue')}}</label>
            </div>
          </div>
          <div class="marginbottom params-each">
            <label class="col-md-2 label-name">{{$t('tableKey.s_last')}}:</label>
            <div class="search-input-content" style="margin-right: 8px">
              <input 
                v-validate="'required|isNumber'" 
                v-model="modelConfig.lastValue" 
                name="lastValue"
                :class="{ 'red-border': veeErrors.has('lastValue') }"
                type="text" 
                class="form-control model-input search-input c-dark"/>
              <label class="required-tip">*</label>
            </div>
            <Select filterable clearable v-model="modelConfig.last" style="width:100px">
              <Option v-for="item in modelConfig.lastList" :value="item.value" :key="item.value">{{ item.label }}</Option>
            </Select>
            <div style="margin-left:10px">
              <label v-show="veeErrors.has('lastValue')" class="is-danger">{{ veeErrors.first('lastValue')}}</label>
            </div>
          </div>
          <div class="marginbottom params-each">
            <label class="col-md-2 label-name">{{$t('tableKey.s_priority')}}:</label>
            <Select filterable clearable v-model="modelConfig.priority" style="width:340px">
              <Option v-for="item in modelConfig.priorityList" :value="item.value" :key="item.value">{{ $t(item.label) }}</Option>
            </Select>
          </div>
          <div class="marginbottom params-each">
            <label class="col-md-2 label-name">{{$t('sendAlarm')}}:</label>
            <Select filterable clearable v-model="modelConfig.addRow.notify_enable" style="width:340px">
              <Option v-for="item in modelConfig.notifyEnableOption" :value="item.value" :key="item.value">{{ item.label }}</Option>
            </Select>
          </div>
          <div class="marginbottom params-each">
            <label class="col-md-2 label-name">{{$t('delay')}}:</label>
            <Select filterable clearable v-model="modelConfig.addRow.notify_delay" style="width:340px">
              <Option v-for="item in modelConfig.notifyDelayOption" :value="item.value" :key="item.value">{{ item.label }}</Option>
            </Select>
          </div>
        </div>
      </ModalComponent>
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
    </section>
  </div>
</template>

<script>
import {thresholdList, lastList, priorityList} from '@/assets/config/common-config.js'
import TagShow from '@/components/Tag-show.vue'
let tableEle = [
  {title: 'ID', value: 'id', display: false},
  {title: 'tableKey.metricName', value: 'metric', display: true},
  {title: 'tableKey.expr', value: 'expr', display: true},
  {title: 'tableKey.s_cond', value: 'cond', display: true},
  {title: 'tableKey.s_last', value: 'last', display: true},
  {title: 'tableKey.s_priority', value: 'priority', display: true}
]
const btn = [
  {btn_name: 'button.edit', btn_func: 'editF'},
  {btn_name: 'button.remove', btn_func: 'deleteConfirmModal'},
]

export default {
  name: '',
  data() {
    return {
      isShowWarning: false,
      requestParams: null,
      isShowWarningDelete: false,
      type: '',
      typeValue: 'endpoint',
      typeList: [
        {label: 'field.endpoint', value: 'endpoint'},
        {label: 'field.group', value: 'grp'}
      ],
      paramsType: null, // For get thresholdList
      endpointID: null,
      endpointType: '',
      endpointOptions: [],

      totalPageConfig: [],
      pageConfig: {
        table: {
          tableData: [],
          tableEle: tableEle,
          // filterMoreBtn: 'filterMoreBtn',
          primaryKey: 'id',
          btn: btn,
          handleFloat:true,
        }
      },
      modelTip: {
        key: '',
        value: 'metric'
      },
      modelConfig: {
        modalId: 'add_edit_Modal',
        modalTitle: 'field.threshold',
        isAdd: true,
        config: [
          {name:'metricSelect',type:'slot'},
          {label: 'tableKey.expr', value: 'expr', placeholder: 'tips.required', v_validate: 'required:true', disabled: true, type: 'textarea'},
          {label: 'tableKey.content', value: 'content', placeholder: 'tips.required', v_validate: 'required:true', disabled: false, type: 'textarea'},
          {name:'thresholdConfig',type:'slot'}
        ],
        addRow: { // [通用]-保存用户新增、编辑时数据
          metric: null,
          expr: null,
          content: null,
          notify_enable: 1,
          notify_delay: 0
        },
        metricName: '',
        metricList: [],
        threshold: '>',
        thresholdList: thresholdList,
        thresholdValue: '',
        last: 's',
        lastList: lastList,
        lastValue: '',
        priority: 'low',
        priorityList: priorityList,
        notifyEnableOption: [
          {label: 'Yes', value: 1},
          {label: 'No', value: 0}
        ],
        notifyDelayOption: [
          {label: '0s', value: 0},
          {label: '30s', value: 30},
          {label: '60s', value: 60},
          {label: '90s', value: 90},
          {label: '120s', value: 120},
          {label: '180s', value: 180},
          {label: '300s', value: 300},
          {label: '600s', value: 600}
        ],
        slotConfig: {
          resourceSelected: [],
          resourceOption: []
        }
      },
      id: null,

      selectReceiversOptions: [] // 待选择接收人
    }
  },
  watch: {
    endpointID: function (id) {
      this.paramsType = null
      const selectedEndpoint = this.endpointOptions.find((endpoint) => {
          return endpoint.id === id
      })
      if (this.type === 'endpoint' && id) {
        this.paramsType = selectedEndpoint.option_value.split(':')[1]
      }
      this.endpointType = selectedEndpoint.type
    }
  },
  mounted () {
    if (!this.$root.$validate.isEmpty_reset(this.$route.params)) {
      this.$parent.activeTab = '/monitorConfigIndex/thresholdManagement'
      this.type = this.$route.params.type
      this.typeValue = this.$route.params.id
      this.paramsType = this.$route.params.paramsType
      this.requestData(this.type, this.typeValue)
    } else {
      this.type = 'endpoint'
      this.typeValue = ''
    }
    this.getEndpointList('.')
    this.$root.JQ('#add_edit_Modal').on('hidden.bs.modal', () => {
      this.modelConfig.thresholdValue = ''
      this.modelConfig.lastValue = ''
    })
  },
  methods: {
    getSelectReceivers (query='r') {
      const params = {search: query}
      this.$root.apiCenter.thresholdManagement.recevier.api
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.thresholdManagement.recevier.getReceivers, params, (responseData) => {
        this.selectReceiversOptions = responseData
      }, {isNeedloading:false})
    },
    saveReceivers (tableItem) {
      if (tableItem.selectedReceiver.length===0&&tableItem.inputReceiver===null) {
        return
      }
      let accept = []
      const regx_email = /^[a-zA-Z0-9_.-]+@[a-zA-Z0-9-]+(\.[a-zA-Z0-9-]+)*\.[a-zA-Z0-9]{2,6}$/gi
      const regs_phone = /^1[345678]\d{9}$/
      const regx_email_res = regx_email.test(tableItem.inputReceiver)
      const regs_phone_res = regs_phone.test(tableItem.inputReceiver)
      if (regx_email_res) {
        accept.push({option_value: tableItem.inputReceiver,option_text: '',type: "mail"})
      }
      if (regs_phone_res) {
        accept.push({option_value: tableItem.inputReceiver,option_text: '',type: "phone"})
      }
      tableItem.selectedReceiver.forEach(sr => {
        accept.push({type: sr})
      })
      accept = tableItem.accept.concat(accept)
      let params = {
        tpl_id: tableItem.tpl_id,
        accept
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.thresholdManagement.recevier.api, params, () => {
        this.$Message.success(this.$t('tips.success'))
        this.requestData(this.type, this.typeValue)
      }, {isNeedloading:false})
    },
    removeReceiver (tableItem, receiver, index) {
      this.isShowWarning = true
      tableItem.accept.splice(index,1)
      this.requestParams = {
        tpl_id: tableItem.tpl_id,
        accept: tableItem.accept
      }
    },
    ok () {
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.thresholdManagement.recevier.api, this.requestParams, () => {
        // this.$root.$eventBus.$emit('hideConfirmModal')
        this.$Message.success(this.$t('tips.success'))
        this.isShowWarning = false
        this.requestData(this.type, this.typeValue)
      }, {isNeedloading:false})
    },
    cancel () {
      this.isShowWarning = false
    },
    search () {
      if (this.endpointID === null) {
        return
      }
      this.typeValue = this.endpointID
      this.requestData(this.type, this.endpointID)
    },
    typeChange() {
      this.totalPageConfig = []
      this.getEndpointList('.')
    },
    clearEndpoint () {
      this.totalPageConfig = []
      this.getEndpointList('.')
    },
    getEndpointList (query) {
      const params = {type: this.type,search: query}
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.resourceSearch.strategyApi, params, (responseData) => {
        this.endpointOptions = responseData
      })
    },
    requestData (type, id) {
      let params = {}
      params.type = type
      params.id = id
      this.totalPageConfig = []
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.thresholdManagement.list.api, params, (responseData) => {
        responseData.forEach((item)=>{
          let config = this.$root.$validate.deepCopy(this.pageConfig.table)
          if (!item.operation) {
            config.btn = []
          }
          item.strategy.forEach(rowData => {
            rowData.type = item.obj_type
          })
          config.tableData = item.strategy
          this.totalPageConfig.push({
            table:config, 
            tpl_id: item.tpl_id, 
            accept: item.accept,
            showReceiver: true,
            selectedReceiver: [], // 选择的接收人
            inputReceiver: null,  // 输入的接收人
            choicedReceiver: [],  // 选中待展示的接收人
            obj_type: item.obj_type, 
            obj_name: item.obj_name, 
            operation:item.operation
          })
        })
      })
    },
    deleteConfirmModal (rowData) {
       this.selectedData = rowData
      this.isShowWarningDelete = true
    },
    okDelRow () {
      this.delF(this.selectedData)
    },
    cancleDelRow () {
      this.isShowWarningDelete = false
    },
    deleteConfirm (rowData) {
      this.$delConfirm({
        msg: rowData.name,
        callback: () => {
          this.delF(rowData)
        }
      })
    },
    delF (rowData) {
      let params = {id: rowData.id}
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.thresholdManagement.delete.api, params, () => {
        // this.$root.$eventBus.$emit('hideConfirmModal')
        this.$Message.success(this.$t('tips.success'))
        this.requestData(this.type, this.typeValue)
      })
    },
    paramsPrepare() {
      let modelParams = {
        cond: this.modelConfig.threshold + this.modelConfig.thresholdValue,
        last: this.modelConfig.lastValue + this.modelConfig.last,
        priority: this.modelConfig.priority,        
      }
      if (this.type === 'grp') {
        modelParams.grp_id = this.typeValue
        modelParams.endpoint_id = 0
      } else {
        modelParams.endpoint_id = parseInt(this.typeValue)
        modelParams.grp_id = 0
      }
      return Object.assign(modelParams, this.modelConfig.addRow)
    },
    add () {
      var params = {endpointType: this.endpointType}
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.getMetricByEndpointType, params, (responseData) => {
        this.modelConfig.metricList = responseData
      })
      this.modelConfig.isAdd = true
      this.$root.JQ('#add_edit_Modal').modal('show')
    },
    addPost () {
      this.$validator.validate().then(result => {
        if (!result) return
        if (this.$root.$validate.isEmpty_reset(this.modelConfig.addRow.content)) {
          this.$Message.warning(this.$t('tableKey.content')+this.$t('tips.required'))
          return
        }
        let params = this.paramsPrepare()
        this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.thresholdManagement.add.api, params, () => {
          this.$Message.success(this.$t('tips.success'))
          this.$root.JQ('#add_edit_Modal').modal('hide')
          this.requestData(this.type, this.typeValue)
        })
      })
    },
    editF (rowData) {
      let params = {type: this.paramsType}
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.getMetricByEndpointType, params, (responseData) => {
        this.modelConfig.metricList = responseData
        this.modelConfig.isAdd = false
        this.id = rowData.id
        this.modelTip.value = rowData.metric
        this.modelConfig.addRow = this.$root.$tableUtil.manageEditParams(this.modelConfig.addRow, rowData)
        let cond = rowData.cond.split('')
        if (cond.indexOf('=') >= 0) {
          this.modelConfig.threshold = cond.slice(0,2).join('')
          this.modelConfig.thresholdValue = cond.slice(2).join('')
        } else {
          this.modelConfig.threshold = cond.slice(0,1).join('')
          this.modelConfig.thresholdValue = cond.slice(1).join('')
        }
        let last = rowData.last
        this.modelConfig.last = last.substring(last.length-1)
        this.modelConfig.lastValue = last.substring(0,last.length-1)
        this.modelConfig.priority = rowData.priority
        this.$root.JQ('#add_edit_Modal').modal('show')
      })
    },
    editPost () {
      this.$validator.validate().then(result => {
        if (!result) return
        if (this.$root.$validate.isEmpty_reset(this.modelConfig.addRow.content)) {
          this.$Message.warning(this.$t('tableKey.content')+this.$t('tips.required'))
          return
        }
        let params = this.paramsPrepare()
        params.strategy_id = this.id
        this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.thresholdManagement.update.api, params, () => {
          this.$Message.success(this.$t('tips.success'))
          this.$root.JQ('#add_edit_Modal').modal('hide')
          this.requestData(this.type, this.typeValue)
        })
      })
    },
    selectMetric (option) {
      if (option) {
        this.modelConfig.addRow.metric = option.label
      }
    }
  },
  components: {
    TagShow
  },
}
</script>

<style scoped lang="less">
  .search-li {
    display: inline-block;
  }
</style>
<style scoped lang="less">
  .search-input {
    height: 32px;
    padding: 4px 7px;
    font-size: 12px;
    border: 1px solid #dcdee2;
    border-radius: 4px;
    width: 230px;
  }

  .section-table-tip {
    margin: 24px 20px 0;
  }
  .search-input:focus {
    outline: 0;
    border-color: #57a3f3;
  }

  .search-input-content {
    display: inline-block;
    vertical-align: middle; 
  }
  .receiver-config {
    margin: 8px 20px;
  }
  .receiver-config-set {
    display: flex;
  }
  .receiver-header {
    font-weight: 500;
    margin-right: 8px;
  }
  .form-control {
    // 解决IE11 输入框无法与按钮同行问题
    display: inline-block;
    padding: 4px 18px 4px 7px;
  }
  /*取消选中样式*/
  .form-control:focus {
    box-shadow: none;
  }
    /*input框样式*/
  .research-input {
    width: 200px;
    height: 32px;
    font-size: 12px;
    margin-left: 8px;
    margin-right: 8px;
  }
  .partition {
    height: 2px;
    background: @blue-lingt;
  }
</style>
