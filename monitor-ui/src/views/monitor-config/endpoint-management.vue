<template>
  <div class="main-content">
    <div class="content-seatch">
      <Input
        v-model="searchForm.search"
        style="width: 15%"
        clearable
        :placeholder="$t('m_enter_name_tips')"
        @on-change="onFilterConditionChange"
      >
      </Input>
      <Select
        v-model="searchForm.endpointGroup"
        filterable
        clearable
        multiple
        style="width: 25%"
        :placeholder="$t('m_please_select') + $t('m_object_group')"
        @on-change="onFilterConditionChange"
      >
        <Option v-for="name in objectGroupList" :value="name" :key="name">
          {{name}}
        </Option>
      </Select>

      <Select
        v-model="searchForm.basicType"
        multiple
        filterable
        clearable
        style="width: 25%"
        :placeholder="$t('m_please_select') + $t('m_basic_type')"
        @on-change="onFilterConditionChange"
      >
        <Option v-for="name in objectTypeList" :value="name" :key="name">
          {{name}}
        </Option>
      </Select>
      <Button class="add-content-item" @click="onAddButtonClick"  type="success" >{{$t('m_add')}}</Button>
    </div>

    <div class="content-table">
      <Table
        size="small"
        :columns="objectTableColumns"
        :data="objectTableData"
      />
    </div>
    <Page
      class="table-pagination"
      :total="pagination.total"
      @on-change="(e) => {pagination.page = e; this.getTableList()}"
      @on-page-size-change="(e) => {pagination.size = e; this.getTableList()}"
      :current="pagination.page"
      :page-size="pagination.size"
      show-total
      show-sizer
    />

    <!-- <PageTable :pageConfig="pageConfig"></PageTable> -->

    <ModalComponent :modelConfig="modelConfig">
      <div slot="advancedConfig" class="extentClass">
        <div class="marginbottom params-each">
          <label class="col-md-2 label-name">{{$t('m_field_endpoint')}}:</label>
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
      :title="$t('m_button_historicalAlert')"
    >
      <div slot="header" class="custom-modal-header">
        <span>
          {{$t('m_alarmHistory')}}
        </span>
        <Icon v-if="isfullscreen" @click="fullscreenChange" class="fullscreen-icon" type="ios-contract" />
        <Icon v-else @click="fullscreenChange" class="fullscreen-icon" type="ios-expand" />
      </div>
      <Table class='history-alarm-config' :columns="historyAlarmPageConfig.table.tableEle" :height="fullscreenTableHight" :data="historyAlarmPageConfig.table.tableData" />
      <Page
        class="history-pagination"
        :total="historyPagination.total"
        @on-change="(e) => {historyPagination.page = e; this.getHistoryAlarmData()}"
        @on-page-size-change="(e) => {historyPagination.size = e; this.getHistoryAlarmData()}"
        :current="historyPagination.page"
        :page-size="historyPagination.size"
        show-total
        show-sizer
      />
    </Modal>

    <ModalComponent :modelConfig="endpointRejectModel">
      <div slot="endpointReject">
        <div class="marginbottom params-each">
          <label class="col-md-2 label-name">{{$t('m_field_type')}}:</label>
          <Select filterable clearable :disabled="!endpointRejectModel.isAdd || isReviewMode" v-model="endpointRejectModel.addRow.type" @on-change="typeChange" style="width: 513px">
            <Option v-for="item in endpointRejectModel.endpointType" :label="item.label" :value="item.value" :key="item.value">
              {{item.label}}
            </Option>
          </Select>
        </div>
        <div v-if="endpointRejectModel.supportStep" class="marginbottom params-each">
          <label class="col-md-2 label-name">{{$t('m_collection_interval')}}:</label>
          <Select filterable clearable v-model="endpointRejectModel.addRow.step" :disabled="['mysql','host','ping','telnet','http','process'].includes(endpointRejectModel.addRow.type) || isReviewMode" style="width: 513px">
            <Option v-for="item in endpointRejectModel.stepOptions" :value="item.value" :label="item.label" :key="item.value">
              {{item.label}}
            </Option>
          </Select>
        </div>
        <div class="marginbottom params-each" v-if="!(['host','windows'].includes(endpointRejectModel.addRow.type)) && ['0', '1'].includes(systemType)">
          <label class="col-md-2 label-name">{{$t('m_field_instance')}}:</label>
          <input v-validate="'required'" :disabled="!endpointRejectModel.isAdd || isReviewMode" v-model="endpointRejectModel.addRow.name" name="name" :class="{'red-border': veeErrors.has('name')}" type="text" class="col-md-9 form-control model-input c-dark" />
          <label class="required-tip">*</label>
          <label v-show="veeErrors.has('name')" class="is-danger">{{ veeErrors.first('name')}}</label>
        </div>
        <div class="marginbottom params-each" v-if="['mysql','redis','java','nginx'].includes(endpointRejectModel.addRow.type)">
          <label class="col-md-2 label-name">{{$t('m_button_trusteeship')}}:</label>
          <Checkbox v-model="endpointRejectModel.addRow.agent_manager" :disabled="isReviewMode"></Checkbox>
        </div>
        <section v-if="['mysql','redis','java','nginx'].includes(endpointRejectModel.addRow.type) && endpointRejectModel.addRow.agent_manager">
          <div v-if="['mysql','java','nginx'].includes(endpointRejectModel.addRow.type)" class="marginbottom params-each">
            <label class="col-md-2 label-name">{{$t('m_button_username')}}:</label>
            <input v-validate="'required'" v-model="endpointRejectModel.addRow.user" :disabled="isReviewMode" name="user" :class="{'red-border': veeErrors.has('user')}" type="text" class="col-md-9 form-control model-input c-dark" />
            <label class="required-tip">*</label>
            <label v-show="veeErrors.has('user')" class="is-danger">{{ veeErrors.first('user')}}</label>
          </div>
          <div class="marginbottom params-each">
            <label class="col-md-2 label-name">{{$t('m_button_password')}}:</label>
            <input v-validate="'required'" v-model="endpointRejectModel.addRow.password" :disabled="isReviewMode" name="password" :class="{'red-border': veeErrors.has('password')}" type="text" class="col-md-9 form-control model-input c-dark" />
            <label class="required-tip">*</label>
            <label v-show="veeErrors.has('password')" class="is-danger">{{ veeErrors.first('password')}}</label>
          </div>
        </section>
        <section v-if="endpointRejectModel.addRow.type === 'http'">
          <div class="marginbottom params-each">
            <label class="col-md-2 label-name">Method:</label>
            <input v-validate="'required'" v-model="endpointRejectModel.addRow.method" :disabled="isReviewMode" name="method" :class="{'red-border': veeErrors.has('method')}" type="text" class="col-md-9 form-control model-input c-dark" />
            <label class="required-tip">*</label>
            <label v-show="veeErrors.has('method')" class="is-danger">{{ veeErrors.first('method')}}</label>
          </div>
          <div class="marginbottom params-each">
            <label class="col-md-2 label-name">URL:</label>
            <input v-validate="'required'" v-model="endpointRejectModel.addRow.url" :disabled="isReviewMode" name="url" :class="{'red-border': veeErrors.has('url')}" type="text" class="col-md-9 form-control model-input c-dark" />
            <label class="required-tip">*</label>
            <label v-show="veeErrors.has('url')" class="is-danger">{{ veeErrors.first('url')}}</label>
          </div>
        </section>
        <div class="marginbottom params-each" v-if="endpointRejectModel.addRow.type === 'other'">
          <label class="col-md-2 label-name">exporter_type: </label>
          <input v-validate="'required'" v-model="endpointRejectModel.addRow.exporter_type" :disabled="isReviewMode" name="exporter_type" :class="{'red-border': veeErrors.has('exporter_type')}" type="text" class="col-md-9 form-control model-input c-dark" />
          <label class="required-tip">*</label>
          <label v-show="veeErrors.has('exporter_type')" class="is-danger">{{ veeErrors.first('exporter_type')}}</label>
        </div>
        <div class="marginbottom params-each" v-if="!(['ping','http', 'snmp', 'process', 'pod'].includes(endpointRejectModel.addRow.type))">
          <label class="col-md-2 label-name">{{$t('m_button_port')}}:</label>
          <input v-validate="'required|isNumber'" v-model="endpointRejectModel.addRow.port" :disabled="isReviewMode" name="port" :class="{'red-border': veeErrors.has('port')}" type="text" class="col-md-9 form-control model-input c-dark" />
          <label class="required-tip">*</label>
          <label v-show="veeErrors.has('port')" class="is-danger">{{ veeErrors.first('port')}}</label>
        </div>
        <div class="marginbottom params-each" v-if="(['ping','http','telnet'].includes(endpointRejectModel.addRow.type))">
          <label class="col-md-2 label-name">{{$t('m_exporter')}}:</label>
          <Checkbox v-model="endpointRejectModel.addRow.exporter" :disabled="isReviewMode" />
        </div>
        <div v-if="endpointRejectModel.addRow.exporter">
          <label class="col-md-2 label-name">{{$t('m_exporter_address')}}:</label>
          <input v-validate="'required'" :placeholder="$t('m_exporter_address_placeholder')" :disabled="isReviewMode" v-model="endpointRejectModel.addRow.export_address" name="export_address" :class="{'red-border': veeErrors.has('export_address')}" type="text" class="col-md-9 form-control model-input c-dark" />
          <label class="required-tip">*</label>
          <label v-show="veeErrors.has('export_address')" class="is-danger">{{ veeErrors.first('export_address')}}</label>
        </div>
        <template v-if="endpointRejectModel.addRow.type === 'process'">
          <div>
            <label class="col-md-2 label-name">{{$t('m_processName')}}:</label>
            <input v-validate="'required'" :placeholder="$t('m_processName')" :disabled="isReviewMode" v-model="endpointRejectModel.addRow.process_name" name="process_name" :class="{'red-border': veeErrors.has('process_name')}" type="text" class="col-md-9 form-control model-input c-dark" />
            <label class="required-tip">*</label>
            <label v-show="veeErrors.has('process_name')" class="is-danger">{{ veeErrors.first('process_name')}}</label>
          </div>
          <div>
            <label class="col-md-2 label-name">{{$t('m_processTags')}}:</label>
            <input :placeholder="$t('m_processTags')" v-model="endpointRejectModel.addRow.tags" :disabled="isReviewMode" type="text" class="col-md-9 form-control model-input c-dark" />
          </div>
        </template>
        <template>
          <div v-if="endpointRejectModel.addRow.type !== 'process'">
            <label class="col-md-2 label-name">{{$t('m_field_ip')}}:</label>
            <input v-validate="'required'" :placeholder="$t('m_field_ip')" :disabled="isReviewMode || disabledIp" v-model="endpointRejectModel.addRow.ip" name="ip" :class="{'red-border': veeErrors.has('ip')}" type="text" class="col-md-9 form-control model-input c-dark" />
            <label class="required-tip">*</label>
            <label v-show="veeErrors.has('ip')" class="is-danger">{{ veeErrors.first('ip')}}</label>
          </div>
          <div v-else>
            <label class="col-md-2 label-name">{{$t('m_field_ip')}}:</label>
            <Select filterable v-model="endpointRejectModel.addRow.ip" :disabled="isReviewMode || disabledIp" @on-change="changeIp" @on-open-change="getIpList()" style="width: 513px">
              <Option v-for="item in endpointRejectModel.ipOptions" :value="item.ip" :key="item.guid">
                {{item.guid}}
              </Option>
            </Select>
          </div>
        </template>
        <div class="marginbottom params-each" v-if="endpointRejectModel.addRow.type === 'pod'">
          <label class="col-md-2 label-name">{{$t('m_cluster')}}:</label>
          <Select filterable clearable v-model="endpointRejectModel.addRow.kubernetes_cluster" :disabled="isReviewMode" style="width: 513px">
            <Option v-for="item in endpointRejectModel.clusterList" :value="item.id" :key="item.id">
              {{item.cluster_name}}
            </Option>
          </Select>
          <label class="required-tip">*</label>
          <label v-show="veeErrors.has('kubernetes_cluster')" class="is-danger">{{ veeErrors.first('kubernetes_cluster')}}</label>
        </div>
        <div class="marginbottom params-each" v-if="endpointRejectModel.addRow.type === 'pod'">
          <label class="col-md-2 label-name">{{$t('m_pod_name')}}:</label>
          <input v-validate="'required'" v-model="endpointRejectModel.addRow.pod_name" :disabled="isReviewMode" name="pod_name" :class="{'red-border': veeErrors.has('pod_name')}" type="text" class="col-md-9 form-control model-input c-dark" />
          <label class="required-tip">*</label>
          <label v-show="veeErrors.has('pod_name')" class="is-danger">{{ veeErrors.first('pod_name')}}</label>
        </div>
      </div>
    </ModalComponent>

    <ModalComponent :modelConfig="processConfigModel">
      <div slot="processConfig">
        <div class="marginbottom params-each">
          <div style="color:#fa7821">
            <span>{{$t('m_button_processConfiguration_tip1')}}</span>
          </div>
        </div>
        <section>
          <div style="display: flex;">
            <div class="port-title">
              <span>{{$t('m_processName')}}:</span>
            </div>
            <div class="port-title">
              <span>{{$t('m_processTags')}}:</span>
            </div>
            <div class="port-title">
              <span>{{$t('m_displayName')}}:</span>
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
              <span>{{$t('m_field_port')}}:</span>
            </div>
            <div class="port-title">
              <span>{{$t('m_tableKey_description')}}:</span>
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
            <i class="fa fa-plus-square-o port-config-icon" @click="addPort" :style="{'visibility': pmIndex + 1 === portModel.portMsg.length ?  'unset' : 'hidden'}" aria-hidden="true"></i>
          </div>
        </section>
      </div>
    </ModalComponent>

    <Modal v-model="isShowDataMonitor" :title="$t('m_button_dataMonitoring')" :styles="{top: '100px',width: '1000px'}" footer-hide>
      <DataMonitor :endpointId="dbEndpointId" ref="dataMonitor"></DataMonitor>
    </Modal>

    <ModalComponent :modelConfig="maintenanceWindowModel">
      <template slot='maintenanceWindow'>
        <div style="margin: 4px 12px;padding:8px 12px;border:1px solid #dcdee2;border-radius:4px">
          <template v-for="(item, index) in maintenanceWindowModel.result" style="margin:6px 0">
            <p :key="index">
              <Button
                @click="deleteMaintenanceWindow(index)"
                size="small"
                type="error"
                icon="md-trash"
                style="background-color: #FF4D4F !important"
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
            style="background-color: #00CB91 !important"
          >{{ $t('m_button_add') }}</Button>
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
          filterable
        >
        </Transfer>
      </div>
    </ModalComponent>
  </div>
</template>

<script>
import { collectionInterval, cycleOption } from '@/assets/config/common-config'
import {
  interceptParams, showPoptipOnTable
} from '@/assets/js/utils'
import DataMonitor from '@/views/monitor-config/data-monitor'
import CryptoJS from 'crypto-js'
import debounce from 'lodash/debounce'
import isEmpty from 'lodash/isEmpty'

export const custom_api_enum = [
  {
    monitorEndpointQuery: 'get'
  },
  {
    getMonitorEndpointById: 'get'
  }
]

const alarmLevelMap = {
  low: {
    label: 'm_low',
    buttonType: 'green'
  },
  medium: {
    label: 'm_medium',
    buttonType: 'gold'
  },
  high: {
    label: 'm_high',
    buttonType: 'red'
  }
}
export default {
  name: '',
  components: {
    DataMonitor
  },
  data() {
    return {
      isfullscreen: false,
      isShowDataMonitor: false,
      dbEndpointId: '',
      dbMonitorData: [],
      historyAlarmPageConfig: {
        table: {
          tableData: [],
          tableEle: [
            {
              title: this.$t('m_alarmName'),
              key: 'alarm_name',
              width: 150,
            },
            // {
            //   title: this.$t('m_tableKey_status'),
            //   width: 80,
            //   key: 'status'
            // },
            {
              title: this.$t('m_menu_configuration'),
              key: 'strategyGroupsInfo',
              render: (h, params) => (
                <div domPropsInnerHTML={params.row.strategyGroupsInfo}></div>
              ),
              width: 80,
            },
            {
              title: this.$t('m_field_endpoint'),
              key: 'endpoint',
              width: 150,
            },
            {
              title: this.$t('m_alarmContent'),
              key: 'content',
              width: 200,
              render: (h, params) => (
                <Tooltip transfer={true} placement="bottom-start" max-width="300">
                  <div slot="content">
                    <div domPropsInnerHTML={params.row.content}></div>
                  </div>
                  <div class='column-eclipse'>{params.row.content || '-'}</div>
                </Tooltip>
              )
            },
            {
              title: this.$t('m_tableKey_s_priority'),
              key: 's_priority',
              width: 80,
              render: (h, params) => (
                <Tag color={alarmLevelMap[params.row.s_priority].buttonType}>{this.$t(alarmLevelMap[params.row.s_priority].label)}</Tag>
              )
            },
            {
              title: this.$t('m_field_metric'),
              key: 'alarm_metric_list_join',
              minWidth: 100,
              render: (h, params) => (
                <Tooltip transfer={true} placement="bottom-start" max-width="300">
                  <div slot="content">
                    <div domPropsInnerHTML={params.row.alarm_metric_list_join}></div>
                  </div>
                  <span class='column-eclipse'>{params.row.alarm_metric_list_join || '-'}</span>
                </Tooltip>
              )
            },
            {
              title: this.$t('m_field_threshold'),
              key: 'alarm_detail',
              minWidth: 150,
              tooltip: true,
              render: (h, params) => (
                <Tooltip transfer={true} placement="bottom-start" max-width="300">
                  <div slot="content">
                    <div domPropsInnerHTML={params.row.alarm_detail}></div>
                  </div>
                  <span class='column-eclipse' domPropsInnerHTML={params.row.alarm_detail}></span>
                </Tooltip>
              )
            },
            {
              title: this.$t('m_tableKey_start'),
              key: 'start_string',
              width: 120,
            },
            {
              title: this.$t('m_frequency'),
              key: 'alarm_total',
              width: 80,
              render: (h, params) => (
                <div>{params.row.alarm_total}</div>
              )
            },
            // {
            //   title: this.$t('m_tableKey_end') + 'wee',
            //   key: 'end_string',
            //   width: 120,
            //   render: (h, params) => {
            //     let res = params.row.end_string
            //     if (params.row.end_string === '0001-01-01 00:00:00') {
            //       res = '-'
            //     }
            //     return h('span', res)
            //   }
            // },
            {
              title: this.$t('m_remark'),
              key: 'custom_message',
              width: 120,
              render: (h, params) => (
                <div>{params.row.custom_message || '-'}</div>
              )
            },
          ]
        }
      },
      modelConfig: {
        modalId: 'add_object_Modal',
        modalTitle: 'm_button_add',
        isAdd: true,
        modalStyle: 'min-width:700px',
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
        modalTitle: 'm_button_portConfiguration',
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
        modalTitle: this.$t('m_field_endpoint'),
        supportStep: true,
        isAdd: true,
        modalStyle: 'min-width:700px',
        modalFooter: [],
        saveFunc: 'endpointRejectSave',
        config: [{
          name: 'endpointReject',
          type: 'slot'
        },
        {
          label: 'm_field_proxy_exporter',
          value: 'proxy_exporter',
          option: 'proxy_exporter',
          hide: true,
          disabled: false,
          type: 'select'
        }
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
          tags: '',
          kubernetes_cluster: null,
          pod_name: ''
        },
        v_select_configs: {
          proxy_exporter: []
        },
        stepOptions: collectionInterval,
        ipOptions: [],
        endpointType: [],
        clusterList: [],
      },
      processConfigModel: {
        modalId: 'process_config_model',
        modalTitle: 'm_button_processConfiguration',
        isAdd: true,
        modalStyle: 'min-width:700px',
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
        cycleOption
      },
      id: null,
      showGroupMsg: false,
      groupMsg: {},
      groupModel: {
        modalId: 'group_modal',
        modalTitle: 'm_tableKey_endpoint',
        saveFunc: 'managementEndpoint',
        modalStyle: 'min-width:900px',
        isAdd: true,
        config: [
          {
            name: 'endpointOperate',
            type: 'slot'
          }
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
        endpointGroup: 'm_base_group',
        serviceGroup: 'm_field_resourceLevel'
      },
      searchForm: {
        search: '',
        endpointGroup: [],
        basicType: []
      },
      objectGroupList: [],
      objectTypeList: [],
      pagination: {
        __orders: '-created_date',
        total: 0,
        page: 1,
        size: 10
      },
      objectTableData: [],
      objectTableColumns: [
        {
          title: this.$t('m_tableKey_endpoint'),
          width: 250,
          key: 'guid',
          tooltip: true
        },
        {
          title: this.$t('m_object_group'),
          width: 300,
          key: 'groups',
          render: (h, params) => (
            <div>
              {
                params.row.groups && params.row.groups.length
                  ? params.row.groups.map(item => (
                    <Tag>{item.name}</Tag>
                  )) : <div>-</div>
              }
            </div>
          )
        },
        {
          title: this.$t('m_basic_type'),
          minWidth: 200,
          key: 'type',
          render: (h, params) => (
            <div>
              {
                params.row.type ? (<TagShow tagName={params.row.type}></TagShow>) : <div>-</div>
              }
            </div>
          )
        },
        {
          title: this.$t('m_creator'),
          key: 'create_user',
          minWidth: 150,
          render: (h, params) => (
            <div>
              {
                params.row.create_user ? <span>{params.row.create_user}</span> : <span>-</span>
              }
            </div>
          )
        },
        {
          title: this.$t('m_updatedBy'),
          key: 'update_user',
          minWidth: 150,
          render: (h, params) => (
            <div>
              {
                params.row.update_user ? <span>{params.row.update_user}</span> : <span>-</span>
              }
            </div>
          )
        },
        {
          title: this.$t('m_update_time'),
          key: 'update_time',
          minWidth: 150,
        },
        {
          title: this.$t('m_table_action'),
          key: 'action',
          width: 210,
          align: 'center',
          fixed: 'right',
          render: (h, params) => (
            <div style="display: flex;">
              <Tooltip placement="top" max-width="400" transfer content={this.$t('m_button_view')}>
                <Button class="mr-1" size="small" type="success" on-click={() => {
                  this.endpointRejectModel.modalFooter = []
                  this.isReviewMode = true
                  this.editF(params.row)
                }}>
                  <Icon type="md-eye" />
                </Button>
              </Tooltip>
              <Tooltip placement="top" max-width="400" transfer content={this.$t('m_button_edit')}>
                <Button class="mr-1" size="small" type="primary" on-click={() => {
                  this.endpointRejectModel.modalFooter = null
                  this.isReviewMode = false
                  this.editF(params.row)
                }}>
                  <Icon type="md-create" />
                </Button>
              </Tooltip>
              <Tooltip placement="top" max-width="400" transfer content={this.$t('m_button_maintenanceWindow')}>
                <Button class="mr-1" size="small" on-click={() => this.maintenanceWindow(params.row)} type="primary">
                  <Icon type="ios-build" />
                </Button>
              </Tooltip>
              <Tooltip placement="top" max-width="400" transfer content={this.$t('m_button_historicalAlert')}>
                <Button class="mr-1" size="small" type="warning" on-click={() => this.historyAlarm(params.row)}>
                  <Icon type="md-warning" />
                </Button>
              </Tooltip>
              <Poptip
                confirm
                transfer
                title={this.$t('m_delConfirm_tip')}
                placement="left-end"
                on-on-ok={() => this.deleteConfirmModal(params.row)}>
                <Button size="small" type="error" on-click={() => {
                  showPoptipOnTable()
                }}>
                  <Icon type="md-trash" />
                </Button>
              </Poptip>
            </div>
          )
        }
      ],
      isReviewMode: false,
      encryptKey: '', // 加密key
      systemType: '', // 0 自定义 1 系统
      historyPagination: {
        total: 0,
        page: 1,
        size: 10
      },
      endPointItem: {},
      request: this.$root.$httpRequestEntrance.httpRequestEntrance,
      apiCenter: this.$root.apiCenter,
    }
  },
  computed: {
    disabledIp() {
      if (['process', 'host'].includes(this.endpointRejectModel.addRow.type) && this.endpointRejectModel.isAdd === false) {
        return true
      }
      return false

    }
  },
  mounted() {
    // this.pageConfig.CRUD = this.$root.apiCenter.endpointManagement.list.api
    if (this.$root.$validate.isEmpty_reset(this.$route.params)) {
      this.groupMsg = {}
      this.showGroupMsg = false
    } else {
      this.$parent.activeTab = '/monitorConfigIndex/endpointManagement'
      if (Object.prototype.hasOwnProperty.call(this.$route.params, 'group')) {
        this.groupMsg = this.$route.params.group
        this.showGroupMsg = true
      }
    }
    this.searchForm.search = this.$route.query.name || ''
    this.getTableList()
    this.getAllOptions()
    this.getIpList()
  },
  filters: {
    interceptParams(val) {
      return interceptParams(val, 55)
    }
  },
  methods: {
    fullscreenChange() {
      this.isfullscreen = !this.isfullscreen
      if (this.isfullscreen) {
        this.fullscreenTableHight = document.documentElement.clientHeight - 160
      } else {
        this.fullscreenTableHight = document.documentElement.clientHeight - 300
      }
    },
    changeIp(val) {
      const process = this.endpointRejectModel.ipOptions.find(i => i.ip === val)
      this.endpointRejectModel.addRow.step = process.step
    },
    getIpList() {
      const api = this.apiCenter.monitorEndpointQuery + '?monitorType=host'
      this.request('GET', api, '', res => {
        this.endpointRejectModel.ipOptions = res.data
      })
    },
    async editF(rowData) {
      const params = {
        page: 1,
        size: 10000,
      }
      await this.request('GET', this.apiCenter.getEndpointTypeNew, params, res => {
        this.endpointRejectModel.endpointType = res.map(item => ({
          label: item.guid,
          value: item.guid,
          systemType: item.system_type // 0 自定义 1 系统
        }))
        // this.endpointRejectModel.endpointType.push({
        //   label: 'm_other',
        //   value: 'other'
        // })
      })
      this.modelTip.value = rowData.guid
      this.endpointRejectModel.isAdd = false
      const api = `/monitor/api/v2/monitor/endpoint/get/${rowData.guid}`
      this.request('GET', api, '', res => {
        this.endpointRejectModel.addRow = res
        const obj = this.endpointRejectModel.endpointType.find(i => i.value === this.endpointRejectModel.addRow.type) || {}
        this.systemType = obj.systemType
        // 如果类型是pod，加载集群列表
        // if (this.endpointRejectModel.addRow.type === 'pod') {
        //   this.getClusterList()
        // }
        this.$root.JQ('#endpoint_reject_model').modal('show')
      })
    },
    managementEndpoint() {
      const params = {
        endpoint_id: Number(this.id),
        group_ids: this.groupModel.group.map(Number)
      }
      this.request('POST', this.apiCenter.endpointManagement.groupUpdate, params, () => {
        this.$Message.success(this.$t('m_tips_success'))
        this.$root.JQ('#group_modal').modal('hide')
      })
    },
    handleChange(newTargetKeys) {
      this.groupModel.group = newTargetKeys
    },
    async groupManagement(rowData) {
      this.id = rowData.id
      const params = {
        page: 1,
        size: 10000,
      }
      await this.request('GET', this.apiCenter.groupManagement.list.api, params, res => {
        this.groupModel.groupOptions = res.data.map(item => ({
          label: item.name,
          key: item.id
        }))
      })
      this.groupModel.group = rowData.groups.map(item => item.id)
      this.$root.JQ('#group_modal').modal('show')
    },
    maintenanceWindowSave() {
      this.maintenanceWindowModel.result.forEach(item => {
        item.weekday = item.weekday.join(',')
      })
      const params = {
        endpoint: this.id,
        data: this.maintenanceWindowModel.result
      }
      this.request('POST', this.apiCenter.endpointManagement.maintenanceWindow.update, params, () => {
        this.$Message.success(this.$t('m_tips_success'))
        this.$root.JQ('#maintenance_window_model').modal('hide')
        this.getTableList()
      })
    },
    maintenanceWindow(rowData) {
      this.id = rowData.guid
      const params = {
        endpoint: rowData.guid
      }
      this.request('GET', this.apiCenter.endpointManagement.maintenanceWindow.get, params, responseData => {
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
    deleteMaintenanceWindow(index) {
      this.maintenanceWindowModel.result.splice(index, 1)
    },
    addEmptyMaintenanceWindow() {
      this.maintenanceWindowModel.result.push({
        time_list: ['00:00', '00:00'],
        weekday: ['All']
      })
    },
    dataMonitor(row) {
      this.dbEndpointId = row.id
      const params = {
        endpoint_id: this.dbEndpointId
      }
      this.request('GET', this.apiCenter.endpointManagement.db.dbMonitor, params, responseData => {
        this.$refs.dataMonitor.managementData(responseData)
        this.dbMonitorData = responseData
        this.isShowDataMonitor = true
      })

    },
    typeChange(type) {
      const obj = this.endpointRejectModel.endpointType.find(i => i.value === type) || {}
      this.systemType = obj.systemType
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
        exporter_type: '',
        kubernetes_cluster: null,
        pod_name: ''
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
      const proxy_exporter = this.endpointRejectModel.config.find(item => item.value === 'proxy_exporter')
      proxy_exporter.hide = true
      this.endpointRejectModel.supportStep = true
      if (type && type === 'snmp') {
        proxy_exporter.hide = false
        this.endpointRejectModel.supportStep = false
        this.request('GET', this.apiCenter.newSnmpConfig, {}, responseData => {
          this.endpointRejectModel.v_select_configs.proxy_exporter = responseData.map(item => ({
            label: item.id,
            value: item.id
          }))
        })
      }
      // 当类型为pod时，获取集群列表
      if (type && type === 'pod') {
        this.endpointRejectModel.supportStep = false
        this.getClusterList()
      }
    },
    add() {
      this.modelConfig.slotConfig.resourceOption = []
      this.modelConfig.slotConfig.resourceSelected = []
      const params = {
        search: '',
        page: 1,
        size: 300
      }
      this.request('POST', this.apiCenter.endpointManagement.list.api, params, responseData => {
        responseData.data.forEach(item => {
          if (item.id !== -1) {
            this.modelConfig.slotConfig.resourceOption.push(item)
          }
        })
        this.$root.JQ('#add_object_Modal').modal('show')
      })
    },
    addPost() {
      const params = {
        grp: this.groupMsg.id,
        endpoints: this.modelConfig.slotConfig.resourceSelected.map(Number),
        operation: 'add'
      }
      this.request('POST', this.apiCenter.endpointManagement.update.api, params, () => {
        this.$Message.success(this.$t('m_tips_success'))
        this.$root.JQ('#add_object_Modal').modal('hide')
      })
    },
    deleteConfirmModal(rowData) {
      const params = {
        guid: rowData.guid
      }
      this.request('POST', this.apiCenter.endpointManagement.deregister.api, params, () => {
        this.$Message.success(this.$t('m_tips_success'))
        this.getTableList()
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
      this.endPointItem = rowData
      this.historyPagination.size = 10
      this.historyPagination.page = 1
      this.historyPagination.total = 0
      this.historyAlarmPageConfig.table.tableData = []
      this.getHistoryAlarmData()
      this.isfullscreen = false
      this.historyAlarmModel = true
    },
    getHistoryAlarmData() {
      const params = {
        id: this.endPointItem.guid,
        page: this.historyPagination.page,
        pageSize: this.historyPagination.size,
      }
      this.request('GET', this.apiCenter.alarm.history, params, responseData => {
        this.historyPagination.total = responseData.pageInfo.totalRows
        this.historyAlarmPageConfig.table.tableData = this.changeResultData(responseData.contents.problem_list)
      })
    },
    changeResultData(dataList) {
      if (dataList && !isEmpty(dataList)) {
        dataList.forEach(item => {
          item.strategyGroupsInfo = '-'
          item.alarm_metric_list_join = '-'
          if (!isEmpty(item.strategy_groups)) {
            item.strategyGroupsInfo = item.strategy_groups.reduce((res, cur) => res + this.$t(this.strategyNameMaps[cur.type]) + ':' + cur.name + '<br/> ', '')
          }

          if (!isEmpty(item.alarm_metric_list)) {
            item.alarm_metric_list_join = item.alarm_metric_list.join(',')
          }
        })
      }
      return dataList
    },
    async endpointReject() {
      const params = {
        page: 1,
        size: 10000,
      }
      await this.request('GET', this.apiCenter.getEndpointTypeNew, params, res => {
        this.endpointRejectModel.endpointType = res.map(item => ({
          label: item.guid,
          value: item.guid,
          systemType: item.system_type // 0 自定义 1 系统
        }))
        // this.endpointRejectModel.endpointType.push(
        //   {
        //     label: 'other',
        //     value: 'other'
        //   }
        // )
      })
      this.endpointRejectModel.isAdd = true
      this.isReviewMode = false
      this.endpointRejectModel.addRow.type = 'host'
      this.systemType = '1'
      this.endpointRejectModel.addRow.step = 10
      this.endpointRejectModel.addRow.port = 9100
      this.$root.JQ('#endpoint_reject_model').modal('show')
    },
    async endpointRejectSave() {
      await this.getEncryptKey()
      this.endpointRejectModel.addRow.port += ''
      const params = this.$root.$validate.isEmptyReturn_JSON(this.endpointRejectModel.addRow)
      this.$validator.validate().then(async result => {
        if (!result) {
          return
        }
        if (this.endpointRejectModel.addRow.exporter_type && ['host', 'mysql', 'redis', 'java', 'windows', 'ping', 'telnet', 'http'].includes(this.endpointRejectModel.addRow.exporter_type)) {
          this.$Message.warning('Export port existed!')
          return
        }
        if (this.endpointRejectModel.addRow.exporter_type) {
          params.type = this.endpointRejectModel.addRow.exporter_type
        }
        if (Object.keys(params).includes('password') && params.password !== '') {
          const key = CryptoJS.enc.Utf8.parse(this.encryptKey)
          const config = {
            iv: CryptoJS.enc.Utf8.parse(Math.trunc(new Date() / 100000) * 100000000),
            mode: CryptoJS.mode.CBC
          }
          params.password = CryptoJS.AES.encrypt(params.password, key, config).toString()
        }

        if (this.endpointRejectModel.isAdd) {
          await this.request('POST', this.apiCenter.endpointManagement.register.api, params)
        } else {
          await this.request('PUT', this.apiCenter.monitorEndpointUpdate, params)
        }
        this.$root.$validate.emptyJson(this.endpointRejectModel.addRow)
        this.$root.JQ('#endpoint_reject_model').modal('hide')
        this.$Message.success(this.$t('m_tips_success'))
        this.getTableList()
      })
    },
    async getEncryptKey() {
      await this.request('GET', this.apiCenter.getEncryptKey, '', responseData => {
        this.encryptKey = responseData
      })
    },
    processManagement(rowData) {
      this.processConfigModel.processName = ''
      this.id = rowData.id
      this.processConfigModel.addRow.processSet = []
      this.request('GET', this.apiCenter.alarmProcessList, {
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
      const emptyPath = this.processConfigModel.process_list.some(t => !t.process_name)
      if (emptyPath) {
        this.$Message.warning(this.$t('m_tableKey_name') + this.$t('m_tips_required'))
        return
      }
      const params = {
        endpoint_id: +this.id,
        process_list: this.processConfigModel.process_list,
        check: true
      }
      this.request('POST', this.apiCenter.alarmProcessUpdate, params, () => {
        this.$Message.success(this.$t('m_tips_success'))
      })
      this.$root.JQ('#process_config_model').modal('hide')
    },
    addProcess() {
      const emptyPath = this.processConfigModel.process_list.some(t => !t.process_name)
      if (emptyPath) {
        this.$Message.warning(this.$t('m_tableKey_name') + this.$t('m_tips_required'))
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
      this.$router.push({
        name: 'businessMonitor',
        params: rowData
      })
    },
    addBusiness() {
      const emptyPath = this.businessConfigModel.pathMsg.some(t => !t.path)
      if (emptyPath) {
        this.$Message.warning(this.$t('m_tableKey_path') + this.$t('m_tips_required'))
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
      const params = {
        guid: rowData.guid
      }
      this.request('GET', this.apiCenter.getAgentEndpointTelnet, params, responseData => {
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
      const emptyPort = this.portModel.portMsg.some(t => !t.port === true)
      if (emptyPort) {
        this.$Message.warning(this.$t('m_field_port') + this.$t('m_tips_required'))
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
      const temp = JSON.parse(JSON.stringify(this.portModel.portMsg.filter(t => {
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
      this.request('POST', this.apiCenter.endpointTelnetUpdate, params, () => {
        this.$Message.success(this.$t('m_tips_success'))
        this.$root.JQ('#port_Modal').modal('hide')
        this.getTableList()
      })
    },
    getTableList() {
      const params = Object.assign({}, this.searchForm, this.pagination)
      // const path = '/monitor/api/v1/alarm/endpoint/list'
      this.request('POST', this.apiCenter.endpointManagement.list.api, params, res => {
        this.pagination.total = res.num
        this.pagination.page = parseInt(params.page)
        this.objectTableData = isEmpty(res.data) ? [] : res.data
      })
    },
    onFilterConditionChange: debounce(function () {
      this.pagination.page = 1
      this.pagination.size = 10
      this.getTableList()
    }, 300),
    onAddButtonClick() {
      this.endpointRejectModel.modalFooter = null
      if (this.$route.params.group) {
        this.add()
      } else {
        this.endpointReject()
      }
    },
    getAllOptions() {
      this.request('GET', this.apiCenter.alarmEndpointOptions, {}, res => {
        this.objectGroupList = res.endpointGroup
        this.objectTypeList = res.basicType
      })
    },
    getClusterList() {
      this.request('GET', this.apiCenter.getK8sClusterList, {}, res => {
        this.endpointRejectModel.clusterList = Array.isArray(res) ? res : []
      })
    }
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

.content-seatch {
  display: flex;
  flex-direction: row;
  justify-content: flex-start;
  align-items: center;
  margin: 15px 0;
  .add-content-item {
    margin-left: auto;
  }
}
.content-seatch > * {
  margin-right: 20px;
}
.content-table {
  max-height: ~'calc(100vh - 200px)';
  overflow-y: auto;
}
.table-pagination {
  position: fixed;
  right: 20px;
  bottom: 20px;
  z-index: 10
}
.history-pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 10px;
}
</style>

<style lang='less'>
.ivu-table-wrapper {
  overflow: inherit;
}
.history-alarm-config {
  .column-eclipse {
    overflow: hidden;
    text-overflow: ellipsis;
    display: -webkit-box;
    -webkit-line-clamp: 3;
    -webkit-box-orient: vertical;
  }
}

</style>
