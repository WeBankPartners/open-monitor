<template>
  <div class=" ">
    <template v-for="(tableItem, tableIndex) in totalPageConfig">
      <section :key="tableIndex + 'f'">
        <div class="section-table-tip">
          <Tag color="blue" :key="tableIndex + 'a'" v-if="tableItem.endpoint_group">{{tableItem.endpoint_group}}</Tag>
        </div>
        <PageTable :pageConfig="tableItem"></PageTable>
      </section>
      <section class="receiver-config" :key="tableIndex + 'g'">
        <div style="margin: 16px 0">
          <h5 style="display:inline-block">{{$t('button.receiversConfiguration')}}:</h5>
          <div class="receiver-config-set" style="margin: 8px 0">
          <template>
            <div style="margin: 4px 0px;padding:8px 12px;width:680px">
              <template v-for="(item, index) in tableItem.groupNotify">
                <p :key="index + 'S'">
                  <Select disabled v-model="item.alarm_action" style="width: 100px" :placeholder="$t('alarm_action')">
                    <Option v-for="type in ['firing', 'ok']" :key="type" :value="type">{{type}}</Option>
                  </Select>
                  <Select disabled v-model="item.proc_callback_key" style="width: 160px" :placeholder="$t('proc_callback_key')">
                    <Option v-for="(flow, flowIndex) in flows" :key="flowIndex" :value="flow.procDefId">{{flow.procDefName}}</Option>
                  </Select>
                  <Select disabled v-model="item.notify_roles" :max-tag-count="2" style="width: 360px" multiple filterable>
                    <Option v-for="item in allRole" :value="item.name" :key="item.value">{{ item.name }}</Option>
                  </Select>
                </p>
              </template>
            </div>
          </template>
          </div>
        </div>
      </section>
    </template>
    <ModalComponent :modelConfig="modelConfig">
      <div slot="metricSelect" class="extentClass">  
        <div class="marginbottom params-each">
          <label class="col-md-2 label-name">{{$t('tableKey.metricName')}}:</label>
          <Select disabled v-model="modelConfig.addRow.metric" filterable clearable style="width:514px"
          :label-in-value="true">
            <Option v-for="(item, index) in modelConfig.metricList" :value="item.guid" :key="item.metric+index">{{ item.metric }}</Option>
          </Select>
        </div> 
        <div class="marginbottom params-each">
          <label class="col-md-2 label-name">{{$t('tableKey.content')}}:</label>
          <Input disabled type="textarea" v-model="modelConfig.addRow.content" style="width:514px"></Input>
        </div> 
      </div>
      <div slot="thresholdConfig" class="extentClass">  
        <div class="marginbottom params-each">
          <label class="col-md-2 label-name">{{$t('field.threshold')}}:</label>
          <Select disabled filterable clearable v-model="modelConfig.threshold" style="width:100px">
            <Option v-for="(item, index) in modelConfig.thresholdList" :value="item.value" :key="index">{{ item.label }}</Option>
          </Select>
          <div class="search-input-content" style="margin-left: 8px">
            <input 
              v-validate="'required|isNumber'" 
              v-model="modelConfig.thresholdValue" 
              name="thresholdValue"
              disabled
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
              disabled
              name="lastValue"
               style="width: 400px"
              :class="{ 'red-border': veeErrors.has('lastValue') }"
              type="text" 
              class="form-control model-input search-input c-dark"/>
            <label class="required-tip">*</label>
          </div>
          <Select disabled filterable clearable v-model="modelConfig.last" style="width:100px">
            <Option v-for="item in modelConfig.lastList" :value="item.value" :key="item.value">{{ item.label }}</Option>
          </Select>
          <div style="margin-left:10px">
            <label v-show="veeErrors.has('lastValue')" class="is-danger">{{ veeErrors.first('lastValue')}}</label>
          </div>
        </div>
        <div class="marginbottom params-each">
          <label class="col-md-2 label-name">{{$t('tableKey.s_priority')}}:</label>
          <Select disabled filterable clearable v-model="modelConfig.priority" style="width:514px">
            <Option v-for="item in modelConfig.priorityList" :value="item.value" :key="item.value">{{ $t(item.label) }}</Option>
          </Select>
        </div>
        <div class="marginbottom params-each">
          <label class="col-md-2 label-name">{{$t('sendAlarm')}}:</label>
          <Select disabled filterable clearable v-model="modelConfig.addRow.notify_enable" style="width:514px">
            <Option v-for="item in modelConfig.notifyEnableOption" :value="item.value" :key="item.value">{{ item.label }}</Option>
          </Select>
        </div>
        <div class="marginbottom params-each">
          <label class="col-md-2 label-name">{{$t('delay')}}:</label>
          <Select disabled filterable clearable v-model="modelConfig.addRow.notify_delay_second" style="width:514px">
            <Option v-for="item in modelConfig.notifyDelayOption" :value="item.value" :key="item.value">{{ item.label }}</Option>
          </Select>
        </div>
      </div>
      <div slot="noticeConfig" class="extentClass">  
        <div v-if="modelConfig.addRow.notify.length > 0" style="margin: 4px 0px;padding:8px 12px;border:1px solid #dcdee2;border-radius:4px;">
          <template v-for="(item, index) in modelConfig.addRow.notify">
            <p :key="index + 'S'">
              <Tooltip :content="$t('alarm_action')" :delay="1000">
                <Select v-model="item.alarm_action" style="width: 100px" :placeholder="$t('alarm_action')">
                  <Option v-for="type in ['firing', 'ok']" :key="type" :value="type">{{type}}</Option>
                </Select>
              </Tooltip>
              <Tooltip :content="$t('proc_callback_key')" :delay="1000">
                <Select v-model="item.proc_callback_key" style="width: 160px" :placeholder="$t('proc_callback_key')">
                  <Option v-for="(flow, flowIndex) in flows" :key="flowIndex" :value="flow.procDefId">{{flow.procDefName}}</Option>
                </Select>
              </Tooltip>
              <Tooltip :content="$t('field.role')" :delay="1000">
                <Select v-model="item.notify_roles" :max-tag-count="2" style="width: 320px" multiple filterable :placeholder="$t('field.role')">
                  <Option v-for="item in allRole" :value="item.name" :key="item.value">{{ item.name }}</Option>
                </Select>
              </Tooltip>
            </p>
          </template>
        </div>
      </div>
      <div slot="btn">
        <Button style="float:right" @click="cancelModal">{{$t('button.cancel')}}</Button>
      </div>
    </ModalComponent>
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
  {btn_name: 'button.view', btn_func: 'editF'},
]
import {thresholdList, lastList, priorityList} from '@/assets/config/common-config.js'
export default {
  name: '',
  data () {
    return {
      targrtId: '',
      type: '',
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
      modelConfig: {
        modalId: 'add_edit_Modal',
        modalTitle: 'field.threshold',
        modalStyle: 'min-width:660px',
        isAdd: true,
        noBtn: true,
        config: [
          {name:'metricSelect',type:'slot'},
          {name:'thresholdConfig',type:'slot'},
          {name:'noticeConfig',type:'slot'},
          {name:'btn',type:'slot'}
        ],
        addRow: { // [通用]-保存用户新增、编辑时数据
          metric: null,
          content: null,
          notify_enable: 1,
          notify_delay_second: 0,
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
      modelTip: {
        key: '',
        value: 'metric'
      },
    }
  },
  mounted () {
    this.getWorkFlow()
    this.getAllRole()
  },
  methods: {
    cancelModal () {
      this.$root.JQ('#add_edit_Modal').modal('hide')
    },
    async editF (rowData) {
      await this.getAllRole()
      await this.getWorkFlow()
      const api = `/monitor/api/v2/monitor/metric/list?monitorType=${this.type}`
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', api, '', (responseData) => {
        this.modelConfig.metricList = responseData
        this.modelConfig.isAdd = false
        this.id = rowData.id
        this.modelTip.value = rowData.metric
        this.modelConfig.addRow = this.$root.$tableUtil.manageEditParams(this.modelConfig.addRow, rowData)
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
    getDetail (targrtId, type) {
      this.targrtId = targrtId
      this.type = type
      const api = this.$root.apiCenter.getThresholdEndpointDetail + '/' + targrtId
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', api, '', (responseData) => {
        this.showManagement = true
        responseData.forEach((item)=>{
          let config = this.$root.$validate.deepCopy(this.pageConfig.table)
          config.tableData = item.strategy
          this.totalPageConfig.push({
            table:config, 
            endpoint_group: item.endpoint_group,
            groupNotify: item.notify
          })
        })
      }, {isNeedloading:true})
    }
  },
  components: {},
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
  .success-btn {
    color: #fff;
    background-color: #19be6b;
    border-color: #19be6b;
  }
</style>

