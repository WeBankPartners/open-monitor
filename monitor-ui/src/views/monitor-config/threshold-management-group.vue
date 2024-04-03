<template>
  <div class=" ">
    <section v-if="showManagement" style="margin-top: 16px;">
      <Tag color="blue">{{targetDetail.display_name || ''}}</Tag>
      <button @click="add" type="button" class="btn btn-small success-btn" style="padding: 0 10px">
        <i class="fa fa-plus"></i>
        {{$t('button.add')}}
      </button>
      <PageTable :pageConfig="pageConfig">
      </PageTable>
      <section class="receiver-config">
        <div style="margin: 16px 0">
          <h5 style="display:inline-block">{{$t('button.receiversConfiguration')}}:</h5>
          <button @click="addEmptyItem('group')" class="btn btn-small success-btn" style="padding: 0 10px">{{$t('button.add')}}</button>
          <button @click="updateNotify" class="btn btn-small btn-cancel-f">{{$t('button.save')}}</button>
          <div class="receiver-config-set" style="margin: 8px 0">
          <template>
            <div style="margin: 4px 0px;padding:8px 12px;">
              <template v-for="(item, index) in groupNotify">
                <p :key="index + 'S'" class="receiver-config-item" style="margin-bottom: 8px;">
                  <Button
                    @click="deleteItem('group', index)"
                    size="small"
                    style="background-color: #ff9900;border-color: #ff9900;"
                    type="error"
                    icon="md-close"
                  ></Button>
                  <Tooltip :content="$t('alarm_action')" :delay="1000">
                    <Select v-model="item.alarm_action" style="width: 100px" :placeholder="$t('alarm_action')">
                      <Option v-for="type in ['firing', 'ok']" :key="type" :value="type">{{type}}</Option>
                    </Select>
                  </Tooltip>
                  <Tooltip :content="$t('resourceLevel.role')" :delay="1000">
                    <Select v-model="item.notify_roles" :max-tag-count="2" style="width: 200px" multiple filterable :placeholder="$t('field.role')">
                      <Option v-for="item in allRole" :value="item.name" :key="item.value">{{ item.name }}</Option>
                    </Select>
                  </Tooltip>
                  <Tooltip :content="$t('proc_callback_key')" :delay="1000">
                    <Select v-model="item.proc_callback_key" @on-change="procCallbackKeyChange(item.proc_callback_key, index)" style="width: 160px" :placeholder="$t('proc_callback_key')">
                      <Option v-for="(flow, flowIndex) in flows" :key="flowIndex" :value="flow.procDefKey" :label="flow.procDefName + ' [' + flow.procDefVersion + ']'"><span>{{ flow.procDefName }} [{{ flow.procDefVersion }}]</span></Option>
                    </Select>
                  </Tooltip>
                  <Tooltip :content="$t('m_callback_mode')" :delay="1000">
                    <Select v-model="item.proc_callback_mode" style="width: 200px" :placeholder="$t('m_callback_mode')">
                      <Option v-for="item in callbackMode" :value="item.value" :key="item.value">{{ $t(item.label) }}</Option>
                    </Select>
                  </Tooltip>
                  <Tooltip :content="$t('tableKey.description')" :delay="1000">
                    <input 
                      v-model="item.description" 
                      style="width: 200px"
                      type="text" 
                      :placeholder="$t('tableKey.description')"
                      class="form-control model-input search-input c-dark"/>
                  </Tooltip>
                </p>
              </template>
            </div>
          </template>
          </div>
        </div>
      </section>
    </section>
    <ModalComponent :modelConfig="modelConfig">
      <div slot="metricSelect" class="extentClass">  
        <div class="marginbottom params-each">
          <label class="col-md-2 label-name">{{$t('tableKey.metricName')}}:</label>
          <Select v-model="modelConfig.addRow.metric" filterable clearable style="width:514px"
          :label-in-value="true">
            <Option v-for="(item, index) in modelConfig.metricList" :value="item.guid" :key="item.metric+index">{{ item.metric }}</Option>
          </Select>
        </div> 
        <div class="marginbottom params-each">
          <label class="col-md-2 label-name">{{$t('tableKey.content')}}:</label>
          <Input type="textarea" v-model="modelConfig.addRow.content" style="width:514px"></Input>
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
              style="width: 408px"
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
               style="width: 400px"
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
          <Select filterable clearable v-model="modelConfig.priority" style="width:514px">
            <Option v-for="item in modelConfig.priorityList" :value="item.value" :key="item.value">{{ $t(item.label) }}</Option>
          </Select>
        </div>
        <div class="marginbottom params-each">
          <label class="col-md-2 label-name">{{$t('sendAlarm')}}:</label>
          <Select filterable clearable v-model="modelConfig.addRow.notify_enable" style="width:514px">
            <Option v-for="item in modelConfig.notifyEnableOption" :value="item.value" :key="item.value">{{ item.label }}</Option>
          </Select>
        </div>
        <div class="marginbottom params-each">
          <label class="col-md-2 label-name">{{$t('delay')}}:</label>
          <Select filterable clearable v-model="modelConfig.addRow.notify_delay_second" style="width:514px">
            <Option v-for="item in modelConfig.notifyDelayOption" :value="item.value" :key="item.value">{{ item.label }}</Option>
          </Select>
        </div>
        <div class="marginbottom params-each">
          <label class="col-md-2 label-name">{{$t('m_active_window')}}:</label>
          <TimePicker v-model="modelConfig.addRow.active_window" :clearable="false" format="HH:mm" type="timerange" placement="bottom-end" style="width: 168px"></TimePicker>
        </div>
      </div>
      <div slot="noticeConfig" class="extentClass">  
        <div style="margin: 4px 0px;padding:8px 12px;border:1px solid #dcdee2;border-radius:4px;">
          <template v-for="(item, index) in modelConfig.addRow.notify">
            <p :key="index + 'S'">
              <Button
                @click="deleteItem('metric', index)"
                size="small"
                style="background-color: #ff9900;border-color: #ff9900;"
                type="error"
                icon="md-close"
              ></Button>
              <Tooltip :content="$t('alarm_action')" :delay="1000">
                <Select v-model="item.alarm_action" style="width: 100px" :placeholder="$t('alarm_action')">
                  <Option v-for="type in ['firing', 'ok']" :key="type" :value="type">{{type}}</Option>
                </Select>
              </Tooltip>
              <Tooltip :content="$t('proc_callback_key')" :delay="1000">
                <Select v-model="item.proc_callback_key" style="width: 160px" :placeholder="$t('proc_callback_key')">
                  <Option v-for="(flow, flowIndex) in flows" :key="flowIndex" :value="flow.procDefKey" :label="flow.procDefName + ' [' + flow.procDefVersion + ']'"><span>{{ flow.procDefName }} [{{ flow.procDefVersion }}]</span></Option>
                </Select>
              </Tooltip>
              <Tooltip :content="$t('field.role')" :delay="1000">
                <Select v-model="item.notify_roles" :max-tag-count="2" style="width: 320px" multiple filterable :placeholder="$t('field.role')">
                  <Option v-for="item in allRole" :value="item.name" :key="item.value">{{ item.name }}</Option>
                </Select>
              </Tooltip>
            </p>
          </template>
          <Button
            @click="addEmptyItem('metric')"
            type="success"
            size="small"
            style="background-color: #0080FF;border-color: #0080FF;"
            long
            >{{ $t('button.add') }}{{$t('tableKey.noticeConfig')}}</Button
          >
        </div>
      </div>
    </ModalComponent>
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
  </div>
</template>

<script>
let tableEle = [
  {title: 'tableKey.metricName', value: 'metric_name', display: true},
  {title: 'tableKey.s_cond', value: 'condition', display: true},
  {title: 'tableKey.s_last', value: 'last', display: true},
  {title: 'tableKey.s_priority', value: 'priority', display: true}
]
const btn = [
  {btn_name: 'button.edit', btn_func: 'editF'},
  {btn_name: 'button.remove', btn_func: 'deleteConfirmModal'},
]

import {thresholdList, lastList, priorityList} from '@/assets/config/common-config.js'
export default {
  name: '',
  data () {
    return {
      targrtId: '',
      type: '',
      showManagement: false,
      targetDetail: {},
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
      modelConfig: {
        modalId: 'add_edit_Modal',
        modalTitle: 'field.threshold',
        modalStyle: 'min-width:660px',
        isAdd: true,
        config: [
          {name:'metricSelect',type:'slot'},
          // {label: 'tableKey.content', value: 'content', placeholder: 'tips.required', v_validate: 'required:true', disabled: false, type: 'textarea'},
          {name:'thresholdConfig',type:'slot'},
          {name:'noticeConfig',type:'slot'}
        ],
        addRow: { // [通用]-保存用户新增、编辑时数据
          metric: null,
          content: null,
          notify_enable: 1,
          notify_delay_second: 0,
          active_window: [ '00:00', '23:59' ],
          notify: []
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
      allRole: [],
      flows: [],
      id: null,
      modelTip: {
        key: '',
        value: 'metric'
      },
      isShowWarningDelete: false,
      selectedData: {},

      groupNotify: [],
      callbackMode: [
        {label: 'm_manual', value: 'manual'},
        {label: 'm_auto', value: 'auto'}
      ]
    }
  },
  mounted () {
    this.getAllRole()
    this.getWorkFlow()
  },
  methods: {
    updateNotify () {
      const api = `/monitor/api/v2/alarm/endpoint_group/${this.targrtId}/notify/update`
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', api, this.groupNotify, () => {
        this.$Message.success(this.$t('tips.success'))
        this.getDetail(this.targrtId, this.type)
      })
    },
    editPost () {
      this.$validator.validate().then(result => {
        if (!result) return
        if (this.$root.$validate.isEmpty_reset(this.modelConfig.addRow.content)) {
          this.$Message.warning(this.$t('tableKey.content')+this.$t('tips.required'))
          return
        }
        let params = JSON.parse(JSON.stringify(this.paramsPrepare()))
        params.active_window = params.active_window.join('-')
        this.$root.$httpRequestEntrance.httpRequestEntrance('PUT', '/monitor/api/v2/alarm/strategy', params, () => {
          this.$Message.success(this.$t('tips.success'))
          this.$root.JQ('#add_edit_Modal').modal('hide')
          this.getDetail(this.targrtId, this.type)
        })
      })
    },
    delF (rowData) {
      const api = `/monitor/api/v2/alarm/strategy/${rowData.guid}`
      this.$root.$httpRequestEntrance.httpRequestEntrance('DELETE', api, '', () => {
        this.$Message.success(this.$t('tips.success'))
        this.getDetail(this.targrtId, this.type)
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
    async editF (rowData) {
      this.selectedData = rowData
      const api = `/monitor/api/v2/monitor/metric/list?monitorType=${this.type}`
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', api, '', (responseData) => {
        this.modelConfig.metricList = responseData
        this.modelConfig.isAdd = false
        this.id = rowData.id
        this.modelTip.value = rowData.metric
        this.modelConfig.addRow = this.$root.$tableUtil.manageEditParams(this.modelConfig.addRow, rowData)
        this.modelConfig.addRow.active_window = rowData.active_window === '' ? ['00:00', '23:59'] : rowData.active_window.split('-')
        let condition = rowData.condition.split('')
        if (condition.indexOf('=') >= 0) {
          this.modelConfig.threshold = condition.slice(0,2).join('')
          this.modelConfig.thresholdValue = condition.slice(2).join('')
        } else {
          this.modelConfig.threshold = condition.slice(0,1).join('')
          this.modelConfig.thresholdValue = condition.slice(1).join('')
        }
        let last = rowData.last
        this.modelConfig.last = last.substring(last.length-1)
        this.modelConfig.lastValue = last.substring(0,last.length-1)
        this.modelConfig.priority = rowData.priority
        this.$root.JQ('#add_edit_Modal').modal('show')
      })
      await this.getAllRole()
      await this.getWorkFlow()
    },
    deleteItem(type, index) {
      if (type === 'group') {
        this.groupNotify.splice(index, 1)
      } else {
        this.modelConfig.addRow.notify.splice(index, 1)
      }
    },
    procCallbackKeyChange(proc_callback_key, index) {
      const findFlow = this.flows.find(f => f.procDefKey === proc_callback_key)
      if (findFlow) {
        this.groupNotify[index].proc_callback_name = `${findFlow.procDefName}[${findFlow.procDefVersion}]`
      } else {
        this.groupNotify[index].proc_callback_name = ''
      }
    },
    addEmptyItem (type) {
      const tmp = {
        alarm_action: 'firing',
        proc_callback_key: '',
        proc_callback_name: '',
        notify_roles: [],
        proc_callback_mode: 'manual',
        description: ''
      }
      if (type === 'group') { 
        this.groupNotify.push(tmp)
      } else {
        this.modelConfig.addRow.notify.push(tmp)
      }
    },
    async add () {
      const api = `/monitor/api/v2/monitor/metric/list?monitorType=${this.type}`
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', api, '', (responseData) => {
        this.modelConfig.metricList = responseData
      })
      await this.getAllRole()
      await this.getWorkFlow()
      this.modelConfig.addRow = {
        metric: null,
        content: null,
        notify_enable: 1,
        notify_delay_second: 0,
        active_window: [ '00:00', '23:59' ],
        notify: []
      }
      this.modelConfig.isAdd = true
      this.$root.JQ('#add_edit_Modal').modal('show')
    },
    getWorkFlow () {
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', '/monitor/api/v2/alarm/event/callback/list', '', (responseData) => {
        this.flows = responseData
      })
    },
    getAllRole () {
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', '/monitor/api/v1/user/role/list?page=1&size=1000', '', (responseData) => {
        this.allRole = responseData.data.map((_) => {
          return {
            ..._,
            value: _.id
          }
        })
      })
    },
    addPost () {
      this.$validator.validate().then(result => {
        if (!result) return
        if (this.$root.$validate.isEmpty_reset(this.modelConfig.addRow.content)) {
          this.$Message.warning(this.$t('tableKey.content')+this.$t('tips.required'))
          return
        }
        let params = JSON.parse(JSON.stringify(this.paramsPrepare()))
        params.active_window = params.active_window.join('-')
        this.$root.$httpRequestEntrance.httpRequestEntrance('POST', '/monitor/api/v2/alarm/strategy', params, () => {
          this.$Message.success(this.$t('tips.success'))
          this.$root.JQ('#add_edit_Modal').modal('hide')
          this.getDetail(this.targrtId, this.type)
        })
      })
    },
    paramsPrepare() {
      let modelParams = {
        guid: this.selectedData.guid,
        endpoint_group: this.targrtId,
        condition: this.modelConfig.threshold + this.modelConfig.thresholdValue,
        last: this.modelConfig.lastValue + this.modelConfig.last,
        priority: this.modelConfig.priority,        
      }
      return Object.assign(modelParams, this.modelConfig.addRow)
    },
    getDetail (targrtId, type) {
      this.targrtId = targrtId
      this.type = type
      const api = this.$root.apiCenter.getThresholdGroupDetail + '/' + targrtId
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', api, '', (responseData) => {
        this.showManagement = true
        if (responseData.length > 0) {
          this.targetDetail = responseData[0]
          this.pageConfig.table.tableData = responseData[0].strategy
          this.groupNotify = responseData[0].notify
        } else {
          this.targetDetail = {}
          this.pageConfig.table.tableData = []
        }
      }, {isNeedloading:true})
    }
  },
  components: {},
}
</script>

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
  .success-btn {
    color: #fff;
    background-color: #19be6b;
    border-color: #19be6b;
  }

  .receiver-config-item > * {
    margin-left: 8px;
  }

  input::placeholder {
    color: #c9ccd2;
    font-size: 12px;
  }
</style>

