<template>
<div class="main-content">
  <PageTable :pageConfig="pageConfig"></PageTable>
  <ModalComponent :modelConfig="modelConfig">
    <div slot="advancedConfig" class="extentClass">
      <div class="marginbottom params-each">
        <label class="col-md-2 label-name">{{$t('field.endpoint')}}:</label>
        <Select v-model="modelConfig.slotConfig.resourceSelected" multiple filterable style="width:300px">
          <Option v-for="item in modelConfig.slotConfig.resourceOption" :value="item.id" :key="item.id">
            {{ item.guid }}
          </Option>
        </Select>
      </div>
    </div>
  </ModalComponent>
  <Modal
    v-model="historyAlarmModel"
    width="1400"
    :mask-closable="false"
    :footer-hide="true"
    :fullscreen="isfullscreen"
    :title="$t('button.historicalAlert')">
    <div slot="header" class="custom-modal-header">
      <span>
        {{$t('alarmHistory')}}
      </span>
      <Icon v-if="isfullscreen" @click="fullscreenChange" class="fullscreen-icon" type="ios-contract" />
      <Icon v-else @click="fullscreenChange" class="fullscreen-icon" type="ios-expand" />
    </div>
    <Table :columns="historyAlarmPageConfig.table.tableEle" :height="fullscreenTableHight" :data="historyAlarmPageConfig.table.tableData"></Table>
  </Modal>
  <ModalComponent :modelConfig="endpointRejectModel">
    <div slot="endpointReject">
      <div class="marginbottom params-each">
        <label class="col-md-2 label-name">{{$t('field.type')}}:</label>
        <Select filterable clearable :disabled="!endpointRejectModel.isAdd" v-model="endpointRejectModel.addRow.type" style="width:338px" @on-change="typeChange">
          <Option v-for="item in endpointRejectModel.endpointType" :value="item.value" :key="item.value">
            {{item.label}}
          </Option>
        </Select>
      </div>
      <div v-if="endpointRejectModel.supportStep" class="marginbottom params-each">
        <label class="col-md-2 label-name">{{$t('m_collection_interval')}}:</label>
        <Select filterable clearable v-model="endpointRejectModel.addRow.step" style="width:338px" :disabled="['mysql','host','ping','telnet','http','process'].includes(endpointRejectModel.addRow.type)">
          <Option v-for="item in endpointRejectModel.stepOptions" :value="item.value" :key="item.value">
            {{item.label}}
          </Option>
        </Select>
      </div>
      <div class="marginbottom params-each" v-if="!(['host','windows'].includes(endpointRejectModel.addRow.type))">
        <label class="col-md-2 label-name">{{$t('field.instance')}}:</label>
        <input v-validate="'required'" :disabled="!endpointRejectModel.isAdd" v-model="endpointRejectModel.addRow.name" name="name" :class="{ 'red-border': veeErrors.has('name') }" type="text" class="col-md-7 form-control model-input c-dark" />
        <label class="required-tip">*</label>
        <label v-show="veeErrors.has('name')" class="is-danger">{{ veeErrors.first('name')}}</label>
      </div>
      <div class="marginbottom params-each" v-if="['mysql','redis','java','nginx'].includes(endpointRejectModel.addRow.type)">
        <label class="col-md-2 label-name">{{$t('button.trusteeship')}}:</label>
        <Checkbox v-model="endpointRejectModel.addRow.agent_manager"></Checkbox>
      </div>
      <section v-if="['mysql','redis','java','nginx'].includes(endpointRejectModel.addRow.type) && endpointRejectModel.addRow.agent_manager">
        <div v-if="['mysql','java','nginx'].includes(endpointRejectModel.addRow.type)" class="marginbottom params-each">
          <label class="col-md-2 label-name">{{$t('button.username')}}:</label>
          <input v-validate="'required'" v-model="endpointRejectModel.addRow.user" name="user" :class="{ 'red-border': veeErrors.has('user') }" type="text" class="col-md-7 form-control model-input c-dark" />
          <label class="required-tip">*</label>
          <label v-show="veeErrors.has('user')" class="is-danger">{{ veeErrors.first('user')}}</label>
        </div>
        <div class="marginbottom params-each">
          <label class="col-md-2 label-name">{{$t('button.password')}}:</label>
          <input v-validate="'required'" v-model="endpointRejectModel.addRow.password" name="password" :class="{ 'red-border': veeErrors.has('password') }" type="text" class="col-md-7 form-control model-input c-dark" />
          <label class="required-tip">*</label>
          <label v-show="veeErrors.has('password')" class="is-danger">{{ veeErrors.first('password')}}</label>
        </div>
      </section>
      <section v-if="endpointRejectModel.addRow.type === 'http'">
        <div class="marginbottom params-each">
          <label class="col-md-2 label-name">Method:</label>
          <input v-validate="'required'" v-model="endpointRejectModel.addRow.method" name="method" :class="{ 'red-border': veeErrors.has('method') }" type="text" class="col-md-7 form-control model-input c-dark" />
          <label class="required-tip">*</label>
          <label v-show="veeErrors.has('method')" class="is-danger">{{ veeErrors.first('method')}}</label>
        </div>
        <div class="marginbottom params-each">
          <label class="col-md-2 label-name">URL:</label>
          <input v-validate="'required'" v-model="endpointRejectModel.addRow.url" name="url" :class="{ 'red-border': veeErrors.has('url') }" type="text" class="col-md-7 form-control model-input c-dark" />
          <label class="required-tip">*</label>
          <label v-show="veeErrors.has('url')" class="is-danger">{{ veeErrors.first('url')}}</label>
        </div>
      </section>
      <div class="marginbottom params-each" v-if="endpointRejectModel.addRow.type === 'other'">
        <label class="col-md-2 label-name">exporter_type: </label>
        <input v-validate="'required'" v-model="endpointRejectModel.addRow.exporter_type" name="exporter_type" :class="{ 'red-border': veeErrors.has('exporter_type') }" type="text" class="col-md-7 form-control model-input c-dark" />
        <label class="required-tip">*</label>
        <label v-show="veeErrors.has('exporter_type')" class="is-danger">{{ veeErrors.first('exporter_type')}}</label>
      </div>
      <div class="marginbottom params-each" v-if="!(['ping','http', 'snmp', 'process'].includes(endpointRejectModel.addRow.type))">
        <label class="col-md-2 label-name">{{$t('button.port')}}:</label>
        <input v-validate="'required|isNumber'" v-model="endpointRejectModel.addRow.port" name="port" :class="{ 'red-border': veeErrors.has('port') }" type="text" class="col-md-7 form-control model-input c-dark" />
        <label class="required-tip">*</label>
        <label v-show="veeErrors.has('port')" class="is-danger">{{ veeErrors.first('port')}}</label>
      </div>
      <div class="marginbottom params-each" v-if="(['ping','http','telnet'].includes(endpointRejectModel.addRow.type))">
        <label class="col-md-2 label-name">{{$t('exporter')}}:</label>
        <Checkbox v-model="endpointRejectModel.addRow.exporter"></Checkbox>
      </div>
      <div v-if="endpointRejectModel.addRow.exporter">
        <label class="col-md-2 label-name">{{$t('exporter_address')}}:</label>
        <input v-validate="'required'" :placeholder="$t('exporter_address_placeholder')" v-model="endpointRejectModel.addRow.export_address" name="export_address" :class="{ 'red-border': veeErrors.has('export_address') }" type="text" class="col-md-7 form-control model-input c-dark" />
        <label class="required-tip">*</label>
        <label v-show="veeErrors.has('export_address')" class="is-danger">{{ veeErrors.first('export_address')}}</label>
      </div>
      <template v-if="endpointRejectModel.addRow.type === 'process'">
        <div>
          <label class="col-md-2 label-name">{{$t('processName')}}:</label>
          <input v-validate="'required'" :placeholder="$t('processName')" v-model="endpointRejectModel.addRow.process_name" name="process_name" :class="{ 'red-border': veeErrors.has('process_name') }" type="text" class="col-md-7 form-control model-input c-dark" />
          <label class="required-tip">*</label>
          <label v-show="veeErrors.has('process_name')" class="is-danger">{{ veeErrors.first('process_name')}}</label>
        </div>
        <div>
          <label class="col-md-2 label-name">{{$t('processTags')}}:</label>
          <input :placeholder="$t('processTags')" v-model="endpointRejectModel.addRow.tags" type="text" class="col-md-7 form-control model-input c-dark" />
        </div>
      </template>
      <template>
        <div v-if="endpointRejectModel.addRow.type !== 'process'">
          <label class="col-md-2 label-name">{{$t('field.ip')}}:</label>
          <input v-validate="'required'" :placeholder="$t('field.ip')" v-model="endpointRejectModel.addRow.ip" name="ip" :class="{ 'red-border': veeErrors.has('ip') }" type="text" class="col-md-7 form-control model-input c-dark" />
          <label class="required-tip">*</label>
          <label v-show="veeErrors.has('ip')" class="is-danger">{{ veeErrors.first('ip')}}</label>
        </div>
        <div v-else>
          <label class="col-md-2 label-name">{{$t('field.ip')}}:</label>
          <Select filterable v-model="endpointRejectModel.addRow.ip" @on-change="changeIp" style="width:338px">
            <Option v-for="item in endpointRejectModel.ipOptions" :value="item.ip" :key="item.guid">
              {{item.guid}}
            </Option>
          </Select>
        </div>
      </template>
    </div>
  </ModalComponent>
  <ModalComponent :modelConfig="processConfigModel">
    <div slot="processConfig">
      <div class="marginbottom params-each">
        <div style="color:#fa7821">
          <span>{{$t('button.processConfiguration_tip1')}}</span>
        </div>
      </div>
      <section>
        <div style="display: flex;">
          <div class="port-title">
            <span>{{$t('processName')}}:</span>
          </div>
          <div class="port-title">
            <span>{{$t('processTags')}}:</span>
          </div>
          <div class="port-title">
            <span>{{$t('displayName')}}:</span>
          </div>
          <i class="fa fa-plus-square-o port-config-icon" @click="addProcess" aria-hidden="true"></i>
        </div>
      </section>
      <section v-for="(pl, plIndex) in processConfigModel.process_list" :key="plIndex">
        <div class="port-config">
          <div style="width: 55%">
            <input type="text" v-model.trim="pl.process_name" class="search-input" style="width: 93%" />
            <label class="required-tip">*</label>
          </div>
          <div style="width: 51%">
            <input type="text" v-model.trim="pl.tags" class="search-input" style="width: 93%" />
          </div>
          <div style="width: 47%">
            <input type="text" v-model.trim="pl.display_name" class="search-input" style="width: 93%" />
          </div>
          <span style="float: right" >
            <i class="fa fa-trash-o port-config-icon" @click="delProcess(plIndex)" aria-hidden="true"></i>
          </span>
        </div>
      </section>
    </div>
  </ModalComponent>
  <ModalComponent :modelConfig="portModel">
    <div slot="port">
      <section>
        <div style="display: flex;">
          <div class="port-title">
            <span>{{$t('field.port')}}:</span>
          </div>
          <div class="port-title">
            <span>{{$t('tableKey.description')}}:</span>
          </div>
        </div>
      </section>

      <section v-for="(pm, pmIndex) in portModel.portMsg" :key="pmIndex">
        <div class="port-config">
          <div style="width: 48%">
            <InputNumber v-model.number="pm.port" type="number" :min=1 :max=65535 style="width: 100%" />
          </div>
          <div style="width: 48%">
            <input type="text" v-model="pm.note" class="search-input" style="width: 100%" />
          </div>
          <i class="fa fa-trash-o port-config-icon" v-if="portModel.portMsg.length > 1" @click="removePort(pmIndex)" aria-hidden="true"></i>
          <i class="fa fa-plus-square-o port-config-icon" @click="addPort" :style="{'visibility': pmIndex+1===portModel.portMsg.length?  'unset' : 'hidden'}" aria-hidden="true"></i>
        </div>
      </section>
    </div>
  </ModalComponent>
  <Modal v-model="isShowWarning" 
    :title="$t('delConfirm.title')"
    @on-ok="ok" 
    @on-cancel="cancel">
    <div class="modal-body" style="padding:30px">
      <div style="text-align:center">
        <p style="color: red">{{$t('delConfirm.tip')}}</p>
      </div>
    </div>
  </Modal>
  <Modal v-model="isShowDataMonitor" :title="$t('button.dataMonitoring')" :styles="{top: '100px',width: '1000px'}" footer-hide>
    <DataMonitor :endpointId="dbEndpointId" ref="dataMonitor"></DataMonitor>
  </Modal>

  <ModalComponent :modelConfig="maintenanceWindowModel">
    <template #maintenanceWindow>
      <div style="margin: 4px 12px;padding:8px 12px;border:1px solid #dcdee2;border-radius:4px">
        <template v-for="(item, index) in maintenanceWindowModel.result">
          <p :key="index" style="margin:6px 0">
            <Button
              @click="deleteMaintenanceWindow(index)"
              size="small"
              type="error"
              icon="md-trash"
            ></Button>
            <TimePicker format="HH:mm" type="timerange" v-model="item.time_list" :clearable="false" style="width: 200px"></TimePicker>
            <Select v-model="item.weekday" multiple filterable style="width:200px">
              <Option v-for="cycle in maintenanceWindowModel.cycleOption" :value="cycle.value" :key="cycle.value">{{
                $t(cycle.label)
              }}</Option>
            </Select>
          </p>
        </template>
        <Button
          @click="addEmptyMaintenanceWindow"
          type="success"
          size="small"
          long
          >{{ $t('button.add') }}</Button
        >
      </div>
    </template>
  </ModalComponent>
  <ModalComponent :modelConfig="groupModel">
    <div slot="endpointOperate">  
      <Transfer
      :data="groupModel.groupOptions"
      :target-keys="groupModel.group"
      :titles="groupModel.titles"
      :list-style="groupModel.listStyle"
      @on-change="handleChange"
      filterable>
      </Transfer>
    </div>
  </ModalComponent>
</div>
</template>

<script>
import DataMonitor from '@/views/monitor-config/data-monitor'
import { cycleOption, collectionInterval } from '@/assets/config/common-config'
import isEmpty from 'lodash/isEmpty'
import {
  interceptParams
} from '@/assets/js/utils'
let tableEle = [{
    title: 'tableKey.endpoint',
    value: 'guid',
    style: 'width:300px',
    display: true
  },
  {
    title: 'tableKey.group',
    display: true,
    tags: {
      // style: 'width: 300px;'
    },
    'render': (item) => {
      let res = item.groups&&item.groups.map((i) => {
        return {
          label: i.name,
          value: i.name
        }
      })
      return res
    }
  }
]
const alarmLevelMap = {
  low: {
    label: "m_low",
    buttonType: "green"
  },
  medium: {
    label: "m_medium",
    buttonType: "gold"
  },
  high: {
    label: "m_high",
    buttonType: "red"
  }
}

const btn = [{
    btn_name: 'button.thresholdManagement',
    btn_func: 'thresholdConfig'
  },
  {
    btn_name: 'button.edit',
    btn_func: 'editF'
  },
  {
    btn_name: 'button.historicalAlert',
    btn_func: 'historyAlarm'
  },
  {
    btn_name: 'field.group',
    btn_func: 'groupManagement'
  },
  {
    btn_name: 'button.remove',
    btn_func: 'deleteConfirmModal',
    color: 'red'
  },
  {
    btn_name: 'button.logConfiguration',
    btn_func: 'logManagement'
  },
  {
    btn_name: 'button.portConfiguration',
    btn_func: 'portManagement'
  },
  {
    btn_name: 'button.processConfiguration',
    btn_func: 'processManagement'
  },
  {
    btn_name: 'button.businessConfiguration',
    btn_func: 'businessManagement'
  },
  {
    btn_name: 'button.dataMonitoring',
    btn_func: 'dataMonitor'
  },
  {
    btn_name: 'm_button_maintenanceWindow',
    btn_func: 'maintenanceWindow'
  }
]
export default {
  name: '',
  data() {
    return {
      isfullscreen: false,
      isShowDataMonitor: false,
      dbEndpointId: '',
      dbMonitorData: [],

      isShowWarning: false,
      pageConfig: {
        CRUD: '',
        researchConfig: {
          input_conditions: [{
            value: 'search',
            type: 'input',
            placeholder: 'placeholder.input',
            style: ''
          }],
          btn_group: [{
            btn_name: 'button.search',
            btn_func: 'search',
            class: 'btn-confirm-f',
            btn_icon: 'fa fa-search'
          }],
          filters: {
            search: ''
          }
        },
        table: {
          tableData: [],
          tableEle: tableEle,
          filterMoreBtn: 'filterMoreBtn',
          primaryKey: 'guid',
          btn: btn,
          pagination: this.pagination,
          handleFloat: true,
        },
        pagination: {
          __orders: '-created_date',
          total: 0,
          page: 1,
          size: 10
        }
      },
      historyAlarmPageConfig: {
        table: {
          tableData: [],
          tableEle: [
            {
              title: this.$t('m_alarmName'),
              key: 'alarm_name'
            },
            {
              title: this.$t('tableKey.status'),
              width: 80,
              key: 'status'
            },
            {
              title: this.$t('menu.configuration'),
              key: 'strategyGroupsInfo',
              render: (h, params) => {
                return (
                  <div domPropsInnerHTML={params.row.strategyGroupsInfo}></div>
                )
              }
            },
            {
              title: this.$t('field.endpoint'),
              key: 'endpoint'
            },
            {
              title: this.$t('alarmContent'),
              key: 'content'
            },
            {
              title: this.$t('tableKey.s_priority'),
              key: 's_priority',
              width: 100,
              render: (h, params) => {
                return (
                  <Tag color={alarmLevelMap[params.row.s_priority].buttonType}>{this.$t(alarmLevelMap[params.row.s_priority].label)}</Tag>
                )
              }
            },
            {
              title: this.$t('field.metric'),
              key: 'alarm_metric_list_join'
            },
            {
              title: this.$t('field.threshold'),
              key: 'alarm_detail',
              width: 200,
              ellipsis: true,
              tooltip: true,
              render: (h, params) => {
                return (
                  <Tooltip transfer={true} placement="bottom-start" max-width="300">
                    <div slot="content">
                      <div domPropsInnerHTML={params.row.alarm_detail}></div>
                    </div>
                    <div domPropsInnerHTML={params.row.alarm_detail}></div>
                  </Tooltip>
                )
              }
            },
            {
              title: this.$t('tableKey.start'),
              key: 'start_string',
              width: 120,
            },
            {
              title: this.$t('tableKey.end'),
              key: 'end_string',
              width: 120,
              render: (h, params) => {
                let res = params.row.end_string
                if (params.row.end_string === '0001-01-01 00:00:00') {
                  res = '-'
                }
                return h('span', res);
              }
            },
            {
              title: this.$t('m_remark'),
              key: 'custom_message',
              width: 120,
              render: (h, params) => {
                return(
                  <div>{params.row.custom_message || '-'}</div>
                )
              }
            },
          ]
        }
      },
      modelConfig: {
        modalId: 'add_object_Modal',
        modalTitle: 'button.add',
        isAdd: true,
        config: [{
          name: 'advancedConfig',
          type: 'slot'
        }],
        addRow: { // [通用]-保存用户新增、编辑时数据
          name: null,
          description: null,
        },
        slotConfig: {
          resourceSelected: [],
          resourceOption: []
        }
      },
      historyAlarmModel: false,
      portModel: {
        modalId: 'port_Modal',
        modalTitle: 'button.portConfiguration',
        saveFunc: 'portSave',
        isAdd: true,
        config: [{
          name: 'port',
          type: 'slot'
        }],
        portMsg: []
      },
      endpointRejectModel: {
        modalId: 'endpoint_reject_model',
        modalTitle: this.$t('field.endpoint'),
        supportStep: true,
        isAdd: true,
        saveFunc: 'endpointRejectSave',
        config: [{
            name: 'endpointReject',
            type: 'slot'
          },
          {label: 'field.proxy_exporter', value: 'proxy_exporter', option: 'proxy_exporter', hide: true, disabled: false, type: 'select'}
        ],
        addRow: {
          name: '',
          type: 'host',
          step: 10,
          ip: null,
          port: 9100,
          agent_manager: false,
          user: '',
          password: '',
          method: '',
          url: '',
          exporter_type: '',
          exporter: false,
          export_address: '',
          proxy_exporter: null,
          process_name: '',
          tags: ''
        },
        v_select_configs: {
            proxy_exporter: []
        },
        stepOptions: collectionInterval,
        ipOptions: [],
        endpointType: [{
            label: 'host',
            value: 'host'
          },
          {
            label: 'mysql',
            value: 'mysql'
          },
          {
            label: 'redis',
            value: 'redis'
          },
          {
            label: 'java',
            value: 'java'
          },
          {
            label: 'process',
            value: 'process'
          },
          {
            label: 'nginx',
            value: 'nginx'
          },
          {
            label: 'windows',
            value: 'windows'
          },
          {
            label: 'ping',
            value: 'ping'
          },
          {
            label: 'telnet',
            value: 'telnet'
          },
          {
            label: 'http',
            value: 'http'
          },
          {
            label: 'snmp',
            value: 'snmp'
          },
          {
            label: 'other',
            value: 'other'
          }
        ],
      },
      processConfigModel: {
        modalId: 'process_config_model',
        modalTitle: 'button.processConfiguration',
        isAdd: true,
        saveFunc: 'processConfigSave',
        config: [{
          name: 'processConfig',
          type: 'slot'
        }],
        addRow: {
          businessSet: [],
        },
        process_list: [],
      },

      maintenanceWindowModel: {
        modalId: 'maintenance_window_model',
        modalTitle: 'm_button_maintenanceWindow',
        isAdd: true,
        saveFunc: 'maintenanceWindowSave',
        config: [{
          name: 'maintenanceWindow',
          type: 'slot'
        }],
        addRow: {
          // businessSet: [],
        },
        result: [],
        cycleOption: cycleOption
      },
      id: null,
      showGroupMsg: false,
      groupMsg: {},
      groupModel: {
        modalId: 'group_modal',
        modalTitle: 'tableKey.endpoint',
        saveFunc: 'managementEndpoint',
        modalStyle: 'min-width:900px',
        isAdd: true,
        config: [
          {name:'endpointOperate',type:'slot'}
        ],
        group: [],
        groupOptions: [],
        titles: [this.$t('m_value_to_be_selected'), this.$t('m_selected_value')],
        listStyle: {
          width: '400px',
          height: '400px'
        }
      },
      modelTip: {
        key: 'guid',
        value: null
      },
      fullscreenTableHight: document.documentElement.clientHeight - 300,
      strategyNameMaps: {
        "endpointGroup": "m_base_group",
        "serviceGroup": "field.resourceLevel"
      }
    }
  },
  mounted() {
    this.pageConfig.CRUD = this.$root.apiCenter.endpointManagement.list.api
    if (this.$root.$validate.isEmpty_reset(this.$route.params)) {
      this.groupMsg = {}
      this.showGroupMsg = false
      this.pageConfig.researchConfig.btn_group.push({
        btn_name: 'button.add',
        btn_func: 'endpointReject',
        class: 'btn-cancel-f',
        btn_icon: 'fa fa-plus'
      })
    } else {
      this.$parent.activeTab = '/monitorConfigIndex/endpointManagement'
      if (Object.prototype.hasOwnProperty.call(this.$route.params, 'group')) {
        this.groupMsg = this.$route.params.group
        this.showGroupMsg = true
        this.pageConfig.researchConfig.btn_group.push({
          btn_name: 'button.add',
          btn_func: 'add',
          class: 'btn-cancel-f',
          btn_icon: 'fa fa-plus'
        })
        this.pageConfig.researchConfig.filters.grp = this.groupMsg.id
      }
      if (Object.prototype.hasOwnProperty.call(this.$route.params, 'search')) {
        this.pageConfig.researchConfig.filters.search = this.$route.params.search
      }
    }
    this.initData(this.pageConfig.CRUD, this.pageConfig)
    this.getIpList()
  },
  filters: {
    interceptParams(val) {
      return interceptParams(val, 55)
    }
  },
  methods: {
    fullscreenChange () {
      this.isfullscreen = !this.isfullscreen
      if (this.isfullscreen) {
        this.fullscreenTableHight = document.documentElement.clientHeight - 160
      } else {
        this.fullscreenTableHight = document.documentElement.clientHeight - 300
      }
    },
    changeIp (val) {
      const process = this.endpointRejectModel.ipOptions.find(i => i.ip === val)
      this.endpointRejectModel.addRow.step = process.step
    },
    getIpList () {
      const api = '/monitor/api/v2/monitor/endpoint/query?monitorType=host'
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', api, '', (res) => {
        this.endpointRejectModel.ipOptions = res.data
      })
    },
    editF (rowData) {
      this.endpointRejectModel.endpointType = [{
            label: 'host',
            value: 'host'
          },
          {
            label: 'mysql',
            value: 'mysql'
          },
          {
            label: 'redis',
            value: 'redis'
          },
          {
            label: 'java',
            value: 'java'
          },
          {
            label: 'process',
            value: 'process'
          },
          {
            label: 'nginx',
            value: 'nginx'
          },
          {
            label: 'windows',
            value: 'windows'
          },
          {
            label: 'ping',
            value: 'ping'
          },
          {
            label: 'telnet',
            value: 'telnet'
          },
          {
            label: 'http',
            value: 'http'
          },
          {
            label: 'snmp',
            value: 'snmp'
          },
          {
            label: 'other',
            value: 'other'
        }]
      this.modelTip.value = rowData.guid
      this.endpointRejectModel.isAdd = false
      const api = `/monitor/api/v2/monitor/endpoint/get/${rowData.guid}`
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', api, '', (res) => {
        this.endpointRejectModel.addRow = res
        this.$root.JQ('#endpoint_reject_model').modal('show')
      })
    },
    managementEndpoint() {
      let params = {
        endpoint_id: Number(this.id),
        group_ids: this.groupModel.group.map(Number)
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.endpointManagement.groupUpdate, params, () => {
        this.$Message.success(this.$t('tips.success'))
        this.$root.JQ('#group_modal').modal('hide')
        this.initData(this.pageConfig.CRUD, this.pageConfig)
      })
    },
    handleChange (newTargetKeys) {
      this.groupModel.group = newTargetKeys
    },
    async groupManagement (rowData) {
      this.id = rowData.id
      let params = {
        page: 1,
        size: 10000,
      }
      await this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.groupManagement.list.api, params, res => {
        this.groupModel.groupOptions = res.data.map(item => {
          return {
            label: item.name,
            key: item.id
          }
        })
      })
      this.groupModel.group = rowData.groups.map(item => item.id)
      this.$root.JQ('#group_modal').modal('show')
    },
    maintenanceWindowSave () {
      this.maintenanceWindowModel.result.forEach(item => {
        item.weekday = item.weekday.join(',')
      })
      const params = {
        endpoint: this.id,
        data: this.maintenanceWindowModel.result
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.endpointManagement.maintenanceWindow.update, params, () => {
        this.$Message.success(this.$t('tips.success'))
        this.$root.JQ('#maintenance_window_model').modal('hide')
      })
    },
    maintenanceWindow (rowData) {
      this.id = rowData.guid
      const params = {
        endpoint: rowData.guid
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.endpointManagement.maintenanceWindow.get, params, (responseData) => {
        this.maintenanceWindowModel.result = []
        responseData.forEach(item => {
          this.maintenanceWindowModel.result.push({
            time_list: item.time_list,
            weekday: item.weekday.split(',')
          })
        })
        this.$root.JQ('#maintenance_window_model').modal('show')
      })
    },
    deleteMaintenanceWindow (index) {
      this.maintenanceWindowModel.result.splice(index, 1)
    },
    addEmptyMaintenanceWindow () {
      this.maintenanceWindowModel.result.push({time_list: ['00:00', '00:00'], weekday: ['All']})
    },
    dataMonitor(row) {
      this.dbEndpointId = row.id
      const params = {
        endpoint_id: this.dbEndpointId
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.endpointManagement.db.dbMonitor, params, responseData => {
        this.$refs.dataMonitor.managementData(responseData)
        this.dbMonitorData = responseData
        this.isShowDataMonitor = true
      })

    },
    typeChange(type) {
      // 
      this.endpointRejectModel.addRow = Object.assign(this.endpointRejectModel.addRow, {
        name: '',
        type,
        ip: null,
        step: 10,
        port: 9100,
        agent_manager: false,
        user: '',
        password: '',
        method: '',
        url: '',
        exporter_type: ''
      })
      if (['ping', 'telnet', 'http'].includes(type)) {
        this.endpointRejectModel.addRow.step = 30 
      }
      const typeToPort = {
        host: 9100,
        mysql: 9104,
        redis: 9121,
        java: 9151,
        windows: 9182
      }
      this.endpointRejectModel.addRow.port = typeToPort[type]
      let proxy_exporter = this.endpointRejectModel.config.find(item => item.value === 'proxy_exporter')
      proxy_exporter.hide = true
      this.endpointRejectModel.supportStep = true
      if (type && type === 'snmp') {
        proxy_exporter.hide = false
        this.endpointRejectModel.supportStep = false
        this.$root.$httpRequestEntrance.httpRequestEntrance('GET', '/monitor/api/v1/config/new/snmp', {},
        responseData => {
          this.endpointRejectModel.v_select_configs.proxy_exporter = responseData.map(item => {
            return {label: item.id, value: item.id}
          })
        })
      }
    },
    initData(url = this.pageConfig.CRUD, params) {
      this.$root.$tableUtil.initTable(this, 'GET', url, params)
    },
    filterMoreBtn(rowData) {
      // let moreBtnGroup = ['thresholdConfig', 'historyAlarm', 'maintenanceWindow', 'deleteConfirmModal']
      let moreBtnGroup = ['historyAlarm', 'editF', 'maintenanceWindow', 'deleteConfirmModal']
      if (rowData.type === 'host') {
        // moreBtnGroup.push('processManagement', 'businessManagement', 'logManagement', 'portManagement')
        moreBtnGroup.push('portManagement')
      }
      // if (rowData.type === 'mysql') {
      //   moreBtnGroup.push('dataMonitor')
      // }
      // if (this.showGroupMsg) {
      //   moreBtnGroup.push('deleteConfirm')
      // }
      return moreBtnGroup
    },
    add() {
      this.modelConfig.slotConfig.resourceOption = []
      this.modelConfig.slotConfig.resourceSelected = []
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', '/monitor/api/v1/alarm/endpoint/list', {
        search: '.',
        page: 1,
        size: 300
      }, responseData => {
        responseData.data.forEach((item) => {
          if (item.id !== -1) {
            this.modelConfig.slotConfig.resourceOption.push(item)
          }
        })
        this.$root.JQ('#add_object_Modal').modal('show')
      })
    },
    addPost() {
      let params = {
        grp: this.groupMsg.id,
        endpoints: this.modelConfig.slotConfig.resourceSelected.map(Number),
        operation: 'add'
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.endpointManagement.update.api, params, () => {
        this.$Message.success(this.$t('tips.success'))
        this.$root.JQ('#add_object_Modal').modal('hide')
        this.initData(this.pageConfig.CRUD, this.pageConfig)
      })
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
    deleteConfirm(rowData) {
      this.$delConfirm({
        msg: rowData.guid,
        callback: () => {
          this.delF(rowData)
        }
      })
    },
    delF(rowData) {
      let params = {
        guid: rowData.guid
      }
      let url = this.$root.apiCenter.endpointManagement.deregister.api
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', url, params, () => {
        this.$Message.success(this.$t('tips.success'))
        this.initData(this.pageConfig.CRUD, this.pageConfig)
      })
    },
    thresholdConfig(rowData) {
      this.$router.push({
        name: 'thresholdManagement',
        params: {
          id: rowData.id,
          type: 'endpoint',
          paramsType: rowData.type
        }
      })
    },
    logManagement(rowData) {
      this.$router.push({
        name: 'logManagement',
        params: {
          id: rowData.id,
          type: 'endpoint'
        }
      })
    },
    historyAlarm(rowData) {
      let params = {
        id: rowData.id
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.alarm.history, params, (responseData) => {
        this.historyAlarmPageConfig.table.tableData = this.changeResultData(responseData[0].problem_list)
      })
      this.isfullscreen = false
      this.historyAlarmModel = true
    },
    changeResultData(dataList) {
      if (dataList && !isEmpty(dataList)) {
        dataList.forEach(item => {
          item.strategyGroupsInfo = '-';
          item.alarm_metric_list_join = '-';
          if (!isEmpty(item.strategy_groups)) {
            item.strategyGroupsInfo = item.strategy_groups.reduce((res, cur)=> {
              return res + this.$t(this.strategyNameMaps[cur.type]) + ':' + cur.name + '<br/> '
            }, '')
          }

          if (!isEmpty(item.alarm_metric_list)) {
            item.alarm_metric_list_join = item.alarm_metric_list.join(',')
          }
        });
      }
      return dataList
    },
    endpointReject() {
      this.endpointRejectModel.isAdd = true
      this.endpointRejectModel.addRow.type = 'host'
      this.endpointRejectModel.addRow.step = 10
      this.endpointRejectModel.addRow.port = 9100
      this.$root.JQ('#endpoint_reject_model').modal('show')
    },
    endpointRejectSave() {
      this.endpointRejectModel.addRow.port += ''
      let params = this.$root.$validate.isEmptyReturn_JSON(this.endpointRejectModel.addRow)
      this.$validator.validate().then(result => {
        if (!result) return
        if (this.endpointRejectModel.addRow.exporter_type && ['host', 'mysql', 'redis', 'java', 'windows', 'ping', 'telnet', 'http'].includes(this.endpointRejectModel.addRow.exporter_type)) {
          this.$Message.warning('Export port existed!')
          return
        } else {
          if (this.endpointRejectModel.addRow.exporter_type) {
            params.type = this.endpointRejectModel.addRow.exporter_type
          }
        }
        const methodType = this.endpointRejectModel.isAdd ? 'POST' : 'PUT'
        const api = this.endpointRejectModel.isAdd ? this.$root.apiCenter.endpointManagement.register.api : '/monitor/api/v2/monitor/endpoint/update'
        this.$root.$httpRequestEntrance.httpRequestEntrance(methodType, api, params, () => {
          this.$root.$validate.emptyJson(this.endpointRejectModel.addRow)
          this.$root.JQ('#endpoint_reject_model').modal('hide')
          this.$Message.success(this.$t('tips.success'))
          this.initData(this.pageConfig.CRUD, this.pageConfig)
        })
      })
    },
    processManagement(rowData) {
      this.processConfigModel.processName = ''
      this.id = rowData.id
      this.processConfigModel.addRow.processSet = []
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', '/monitor/api/v1/alarm/process/list', {
        id: this.id
      }, responseData => {
        if (!responseData.length) {
          responseData.push({
            process_name: '',
            display_name: '',
            tags: ''
          })
        }
        this.processConfigModel.process_list = responseData
        this.$root.JQ('#process_config_model').modal('show')
      })
    },
    processConfigSave() {
      const emptyPath = this.processConfigModel.process_list.some(t => {
        return !t.process_name
      })
      if (emptyPath) {
        this.$Message.warning(this.$t('tableKey.name') + this.$t('tips.required'))
        return
      }
      const params = {
        endpoint_id: +this.id,
        process_list: this.processConfigModel.process_list,
        check: true
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', '/monitor/api/v1/alarm/process/update', params, () => {
        this.$Message.success(this.$t('tips.success'))
      })
      this.$root.JQ('#process_config_model').modal('hide')
    },
    addProcess() {
      const emptyPath = this.processConfigModel.process_list.some(t => {
        return !t.process_name
      })
      if (emptyPath) {
        this.$Message.warning(this.$t('tableKey.name') + this.$t('tips.required'))
        return
      }
      this.processConfigModel.process_list.push({
        process_name: '',
        display_name: '',
        tags: ''
      })
    },
    delProcess(plIndex) {
      this.processConfigModel.process_list.splice(plIndex, 1)
    },

    businessManagement(rowData) {
      this.$router.push({name: 'businessMonitor', params: rowData})
    },
    addBusiness() {
      const emptyPath = this.businessConfigModel.pathMsg.some(t => {
        return !t.path
      })
      if (emptyPath) {
        this.$Message.warning(this.$t('tableKey.path') + this.$t('tips.required'))
        return
      }
      this.businessConfigModel.pathMsg.push({
        owner_endpoint: null,
        path: null
      })
    },
    delBusiness(pmIndex) {
      this.businessConfigModel.pathMsg.splice(pmIndex, 1)
    },
    portManagement(rowData) {
      this.id = rowData.guid
      let params = {
        guid: rowData.guid
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', '/monitor/api/v1/agent/endpoint/telnet/get', params, (responseData) => {
        if (!responseData.length) {
          responseData.push({
            port: null,
            note: '',
          })
        }
        this.portModel.portMsg = responseData
      })
      this.$root.JQ('#port_Modal').modal('show')
    },
    addPort() {
      const emptyPort = this.portModel.portMsg.some(t => {
        return !t.port === true
      })
      if (emptyPort) {
        this.$Message.warning(this.$t('field.port') + this.$t('tips.required'))
        return
      }
      this.portModel.portMsg.push({
        port: null,
        note: '',
      })
    },
    removePort(pmIndex) {
      this.portModel.portMsg.splice(pmIndex, 1)
    },
    portSave() {
      let temp = JSON.parse(JSON.stringify(this.portModel.portMsg.filter(t => {
        if (!t.port === false) {
          return t
        }
      })))
      temp.map(t => {
        t.port += ''
        return t
      })
      const params = {
        guid: this.id,
        config: temp
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', '/monitor/api/v1/agent/endpoint/telnet/update', params, () => {
        this.$Message.success(this.$t('tips.success'))
        this.$root.JQ('#port_Modal').modal('hide')
      })
    }
  },
  components: {
    DataMonitor
  }
}
</script>

<style lang="less" scoped>
.params-each /deep/ .ivu-checkbox {
  top: 3px !important;
}

.red-border {
  border: 1px solid red !important;
}

.is-danger {
  color: red;
  margin-left: 80px;
  margin-bottom: 0px;
}

.search-input {
  display: inline-block;
  height: 32px;
  padding: 4px 7px;
  font-size: 12px;
  border: 1px solid #dcdee2;
  border-radius: 4px;
  color: #515a6e;
  background-color: #fff;
  background-image: none;
  position: relative;
  cursor: text;

  width: 310px;
}

.search-input:focus {
  outline: 0;
  border-color: #57a3f3;
}

.search-input-content {
  display: inline-block;
  vertical-align: middle;
}

.port-title {
  width: 48%;
  font-size: 14px;
  padding: 2px 0 2px 4px;
  // border: 1px solid @blue-2;
}

.port-config {
  display: flex;
  margin-top: 4px;
}

.port-config-icon {
  font-size: 16px;
  margin: 7px 2px;
}

.fa-trash-o {
  color: @color-red;
}

.fa-plus-square-o {
  color: @color-blue;
}

.custom-modal-header {
  line-height: 20px;
  font-size: 16px;
  color: #17233d;
  font-weight: 500;
  .fullscreen-icon {
    float: right;
    margin-right: 28px;
    font-size: 18px;
    cursor: pointer;
  }
}
</style>
