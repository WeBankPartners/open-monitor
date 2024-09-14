<template>
  <div class="all-content">
    <section>
      <div class='upload-content'>
        <Button
          type="info"
          class="btn-left"
          @click="exportData"
        >
          <img src="../../assets/img/import.png" class="btn-img" alt="" />
          {{ $t('m_export') }}
        </Button>
        <div style="display: inline-block;margin-bottom: 3px;">
          <Upload
            :action="uploadUrl"
            :show-upload-list="false"
            :max-size="1000"
            with-credentials
            :headers="{'Authorization': token}"
            :on-success="uploadSucess"
            :on-error="uploadFailed"
          >
            <Button type="primary" class="btn-left">
              <img src="../../assets/img/export.png" class="btn-img" alt="" />
              {{ $t('m_import') }}
            </Button>
          </Upload>
        </div>
      </div>

      <div v-for="(single, i) in logAndDataBaseAllDetail" :key="i">
        <div class='w-header'>
          <div class="title">
            {{$t('m_log_file')}}
            <span class="underline"></span>
          </div>
          <Button type="success" class="btn-right mr-4" @click="addLogFile">
            {{ $t('m_button_add') }}
          </Button>
        </div>

        <div>
          <Collapse v-model="logFileCollapseValue" v-if='!isEmpty(single.logFile) && !isEmpty(single.logFile.config)'>
            <Panel v-for="(item, index) in single.logFile.config"
                   :key="index"
                   :name="index + ''"
            >
              <div class="log-file-collapse-content">
                <div>
                  <div class="use-underline-title mr-4">
                    {{item.log_path}}
                    <span class="underline"></span>
                  </div>
                  <Tag color="blue">{{ item.monitor_type }}</Tag>
                </div>
                <div class="log-file-collapse-button" @click="(e) => {e.stopPropagation()}">
                  <Tooltip :content="$t('m_use_custom')" placement="bottom" transfer>
                    <Button
                      size="small"
                      type="success"
                      @click.stop="addByCustom(item)"
                    >
                      <Icon type="ios-add-circle" size="16" />
                    </Button>
                  </Tooltip>
                  <Dropdown
                    placement="left-start"
                    :key="index"
                    class="chart-option-menu"
                    @on-click="(index) => {
                      selectedTemp = allTemplateList[index].guid;
                      parentGuid = item.guid;
                      addConfigType = allTemplateList[index].log_type
                      okTempSelect()
                    }"
                  >
                    <Button type="success" size="small">
                      <Icon type="ios-link" size="16" />
                    </Button>
                    <template slot='list'>
                      <DropdownMenu>
                        <DropdownItem v-for="(option, key) in allTemplateList"
                                      :name="key"
                                      :key="key"
                                      :disabled="option.disabled"
                        >
                          {{option.name}}
                        </DropdownItem>
                      </DropdownMenu>
                    </template>
                  </Dropdown>
                  <Button size="small" class="mr-1"  type="primary" @click.stop="editF(item)">
                    <Icon type="md-create" size="16" />
                  </Button>
                  <Poptip
                    confirm
                    :title="$t('m_delConfirm_tip')"
                    placement="left-end"
                    @on-ok="deleteLogFiltItem(item)"
                  >
                    <Button size="small" type="error" class="mr-2">
                      <Icon type="md-trash" size="16" />
                    </Button>
                  </Poptip>
                </div>
              </div>
              <template slot='content'>
                <Table
                  class="log-file-table"
                  size="small"
                  :columns="logFileTableColumns"
                  :data="item.metric_groups"
                />
              </template>
            </Panel>
          </Collapse>
          <div v-else class='no-data-class'>{{$t('m_table_noDataTip')}}</div>
        </div>

        <div style="margin-top: 16px; padding-bottom: 30px">
          <div class="w-header" slot="title">
            <div class="title">
              {{$t('m_db')}}
              <span class="underline"></span>
            </div>
            <Button type="success" class="btn-right mr-4" @click="addDb">
              {{ $t('m_button_add') }}
            </Button>
          </div>
          <Table
            class="log-file-table"
            size="small"
            :columns="dataBaseTableColumns"
            :data="!isEmpty(single.database) && !isEmpty(single.database.config) ? single.database.config : []"
          />
        </div>
      </div>

    </section>
    <Modal
      v-model="addAndEditModal.isShow"
      :title="addAndEditModal.isAdd ? $t('m_button_add') : $t('m_button_edit')"
      :mask-closable="false"
      :width="730"
    >
      <div :style="{'max-height': MODALHEIGHT + 'px', overflow: 'auto'}">
        <div>
          <span>{{$t('m_field_type')}}:</span>
          <Select v-model="addAndEditModal.dataConfig.monitor_type" @on-change="getEndpoint(addAndEditModal.dataConfig.monitor_type, 'host')" style="width: 640px">
            <Option v-for="type in monitorTypeOptions" :key="type.value" :value="type.label">{{type.label}}</Option>
          </Select>
        </div>
        <div v-if="addAndEditModal.isAdd" style="margin: 4px 0px;padding:8px 12px;border:1px solid #dcdee2;border-radius:4px;width:680px;text-align: center;">
          <template v-for="(item, index) in addAndEditModal.pathOptions">
            <p :key="index + 5">
              <Tooltip :content="$t('m_tableKey_logPath')" :delay="1000">
                <Input v-model.trim="item.path" style="width: 620px" :placeholder="$t('m_tableKey_logPath')" />
              </Tooltip>
              <Button
                v-if="addAndEditModal.isAdd"
                @click="deleteItem('path', index)"
                size="small"
                type="error"
                icon="md-trash"
              ></Button>
            </p>
          </template>
          <Button
            @click="addEmptyItem('path')"
            type="success"
            size="small"
            style="width:650px"
            long
          >{{ $t('m_button_add') }}{{$t('m_tableKey_logPath')}}</Button>
        </div>
        <div v-else style="margin: 8px 0">
          <span>{{$t('m_tableKey_path')}}:</span>
          <Input style="width: 640px" v-model.trim="addAndEditModal.dataConfig.log_path" />
        </div>
        <div style="margin: 4px 0px;padding:8px 12px;border:1px solid #dcdee2;border-radius:4px;width:680px;text-align: center;">
          <template v-for="(item, index) in addAndEditModal.dataConfig.endpoint_rel">
            <p :key="index + 'c'">
              <Tooltip :content="$t('m_type_object')" :delay="1000">
                <Select v-model="item.target_endpoint" style="width: 310px" :placeholder="$t('m_type_object')">
                  <Option v-for="type in targetEndpoints" :key="type.guid" :value="type.guid">{{type.display_name}}</Option>
                </Select>
              </Tooltip>
              <Tooltip :content="$t('m_host_object')" :delay="1000">
                <Select v-model="item.source_endpoint" style="width: 310px" :placeholder="$t('m_host_object')">
                  <Option v-for="type in sourceEndpoints" :key="type.guid" :value="type.guid">{{type.display_name}}</Option>
                </Select>
              </Tooltip>
              <Button
                @click="deleteItem('relate', index)"
                size="small"
                type="error"
                icon="md-trash"
              ></Button>
            </p>
          </template>
          <Button
            @click="addEmptyItem('relate')"
            type="success"
            size="small"
            style="width:650px"
            long
          >{{$t('m_addStringMap')}}</Button>
        </div>
      </div>
      <div slot="footer">
        <Button @click="cancelAddAndEdit">{{$t('m_button_cancel')}}</Button>
        <Button @click="okAddAndEdit" type="primary">{{$t('m_button_save')}}</Button>
      </div>
    </Modal>
    <Modal
      v-model="ruleModelConfig.isShow"
      :title="$t('m_json_regular')"
      width="840"
      :mask-closable="false"
    >
      <div :style="{'max-height': MODALHEIGHT + 'px', overflow: 'auto'}">
        <Form :label-width="100">
          <FormItem :label="$t('m_tableKey_tags')">
            <Input v-model="ruleModelConfig.addRow.tags" style="width:100%" />
          </FormItem>
          <FormItem :label="$t('m_tableKey_regular')">
            <Input type="textarea" v-model="ruleModelConfig.addRow.json_regular" style="width: 580px"/>
            <Button v-if="!showRegConfig" @click="showRegConfig = !showRegConfig">{{$t('m_menu_configuration')}}</Button>
          </FormItem>
        </Form>
        <RegTest v-if="showRegConfig" @updateReg="updateReg" @cancelReg="cancelReg"></RegTest>
        <div style="margin: 4px 0px;padding:8px 12px;border:1px solid #dcdee2;border-radius:4px;">
          <template v-for="(item, index) in ruleModelConfig.addRow.metric_list">
            <p :key="index + 3">
              <Tooltip :content="$t('m_key')" :delay="1000">
                <Input v-model="item.json_key" style="width: 190px" :placeholder="$t('m_key') + ' e.g:[.*][.*]'" />
              </Tooltip>
              <Tooltip :content="$t('m_field_metric') + ' , e.g:code'" :delay="1000">
                <Input v-model="item.metric" style="width: 190px" :placeholder="$t('m_field_metric') + ' , e.g:code'" />
              </Tooltip>
              <Tooltip :content="$t('m_field_aggType')" :delay="1000">
                <Select v-model="item.agg_type" filterable clearable style="width:100px">
                  <Option v-for="agg in ruleModelConfig.aggOption" :value="agg" :key="agg">{{
                    agg
                  }}</Option>
                </Select>
              </Tooltip>
              <Tooltip :content="$t('m_field_displayName')" :delay="1000">
                <Input v-model="item.display_name" style="width: 160px" :placeholder="$t('m_field_displayName')" />
              </Tooltip>
              <Button
                @click="deleteItem('metric_list', index)"
                size="small"
                type="error"
                icon="md-trash"
              ></Button>
            </p>
            <div :key="index + 1" style="margin: 4px 0px;padding:8px 12px;border:1px solid #dcdee2;border-radius:4px;text-align: end;">
              <template v-for="(stringMapItem, stringMapIndex) in item.string_map">
                <p :key="stringMapIndex + 2">
                  <Tooltip :content="$t('m_tableKey_regular')" :delay="1000">
                    <Select v-model="stringMapItem.regulative" filterable clearable style="width:230px">
                      <Option v-for="regulation in regulationOption" :value="regulation.value" :key="regulation.value">{{
                        regulation.label
                      }}</Option>
                    </Select>
                  </Tooltip>
                  <Tooltip :content="$t('m_target_value')" :delay="1000">
                    <Input v-model="stringMapItem.target_value" style="width: 230px" :placeholder="$t('m_target_value')" />
                  </Tooltip>
                  <Tooltip :content="$t('m_source_value')" :delay="1000">
                    <Input v-model="stringMapItem.source_value" style="width: 230px" :placeholder="$t('m_source_value')" />
                  </Tooltip>
                  <Button
                    @click="deleteItem('string_map', index)"
                    size="small"
                    type="error"
                    icon="md-trash"
                  ></Button>
                </p>
              </template>
              <Button
                @click="addEmptyItem('string_map', index)"
                type="success"
                size="small"
              >{{ $t('m_addStringMap') }}</Button>
            </div>
            <Divider :key="index + 'Q'" />
          </template>
          <Button
            @click="addEmptyItem('metric_list')"
            type="success"
            size="small"
            long
          >{{ $t('m_addMetricConfig') }}</Button>
        </div>
      </div>
      <div slot="footer">
        <Button @click="cancelRule">{{$t('m_button_cancel')}}</Button>
        <Button @click="saveRule" type="primary">{{$t('m_button_save')}}</Button>
      </div>
    </Modal>
    <ModalComponent :modelConfig="customMetricsModelConfig">
      <div slot="ruleConfig" class="extentClass">
        <div class="marginbottom params-each">
          <label class="col-md-2 label-name">{{$t('m_tableKey_regular')}}:</label>
          <Input style="width: 70%" type="textarea" v-model="customMetricsModelConfig.addRow.regular" />
          <Button v-if="!showCustomRegConfig" size="small" @click="showCustomRegConfig = !showCustomRegConfig">{{$t('m_menu_configuration')}}</Button>
        </div>
        <RegTest v-if="showCustomRegConfig" @updateReg="updateCustomReg" @cancelReg="cancelCustomReg"></RegTest>
        <div class="marginbottom params-each">
          <label class="col-md-2 label-name">{{$t('m_field_aggType')}}:</label>
          <Select v-model="customMetricsModelConfig.addRow.agg_type" filterable clearable style="width:510px">
            <Option v-for="agg in customMetricsModelConfig.slotConfig.aggOption" :value="agg" :key="agg">{{
              agg
            }}</Option>
          </Select>
        </div>
        <div class="marginbottom params-each">
          <div style="margin: 4px 12px;padding:8px 12px;border:1px solid #dcdee2;border-radius:4px">
            <template v-for="(item, index) in customMetricsModelConfig.addRow.string_map">
              <p :key="index">
                <Tooltip :content="$t('m_tableKey_regular')" :delay="1000">
                  <Select v-model="item.regulative" filterable clearable style="width:150px">
                    <Option v-for="regulation in regulationOption" :value="regulation.value" :key="regulation.value">{{
                      regulation.label
                    }}</Option>
                  </Select>
                </Tooltip>
                <Tooltip :content="$t('m_target_value')" :delay="1000">
                  <Input v-model="item.target_value" style="width: 250px" :placeholder="$t('m_target_value')" />
                </Tooltip>
                <Tooltip :content="$t('m_source_value')" :delay="1000">
                  <Input v-model="item.source_value" style="width: 250px" :placeholder="$t('m_source_value')" />
                </Tooltip>
                <Button
                  @click="deleteCustomMetric('string_map', index)"
                  size="small"
                  type="error"
                  icon="md-trash"
                ></Button>
              </p>
            </template>
            <Button
              @click="addCustomMetricEmpty('string_map')"
              type="success"
              size="small"
              long
            >{{ $t('m_addStringMap') }}</Button>
          </div>
        </div>
        <!-- 新增标签 -->
        <div class="marginbottom params-each">
          <div style="margin: 4px 12px;padding:8px 12px;border:1px solid #dcdee2;border-radius:4px">
            <template v-for="(item, index) in customMetricsModelConfig.addRow.tag_config">
              <p :key="index">
                <Tooltip :content="$t('m_tableKey_tags')" :delay="1000">
                  <Input v-model="item.key" style="width: 250px" :placeholder="$t('m_tableKey_tags')" />
                </Tooltip>
                <Tooltip :content="$t('m_tableKey_regular')" :delay="1000">
                  <Input v-model="item.regular" style="width: 400px" :placeholder="$t('m_tableKey_regular')" />
                </Tooltip>
                <Button
                  @click="deleteCustomMetric('tag_config', index)"
                  size="small"
                  type="error"
                  icon="md-trash"
                ></Button>
              </p>
            </template>
            <Button
              @click="addCustomMetricEmpty('tag_config')"
              type="success"
              size="small"
              long
            >{{ $t('m_add_tags') }}</Button>
          </div>
        </div>
      </div>
    </ModalComponent>

    <!--数据库-->
    <BaseDrawer
      :title="$t('m_db')"
      :visible.sync="dbModelConfig.isShow"
      :realWidth="1000"
      :scrollable="true"
    >
      <template slot-scope="{maxHeight}" slot="content">
        <div :style="{'max-height': maxHeight + 'px', overflow: 'auto'}">
          <Form :label-width="100">
            <!-- <FormItem :label="$t('m_field_displayName')" required>
              <Input v-model.trim="dbModelConfig.addRow.display_name" />
            </FormItem> -->
            <FormItem :label="$t('m_metric_key')" required>
              <Input v-model.trim="dbModelConfig.addRow.metric" :placeholder="$t('m_metric_key_placeholder_second')" />
            </FormItem>
            <FormItem :label="$t('m_sql_script')" required>
              <Input v-model.trim="dbModelConfig.addRow.metric_sql" type="textarea" />
            </FormItem>
            <FormItem :label="$t('m_field_type')" style="margin-top: 12px;" required>
              <Select v-model="dbModelConfig.addRow.monitor_type" @on-change="getEndpoint(dbModelConfig.addRow.monitor_type, 'mysql')" transfer>
                <Option v-for="type in monitorTypeOptions" :key="type.value" :value="type.label">{{type.label}}</Option>
              </Select>
            </FormItem>
            <FormItem :label="$t('m_collection_interval')">
              <Select v-model="dbModelConfig.addRow.step" style="width: 520px" transfer>
                <Option
                  v-for="item in stepOptions"
                  :key="item.value"
                  :value="item.value"
                  :label="item.name"
                >
                  {{item.name}}
                </Option>
              </Select>
            </FormItem>
          </Form>
          <div style="margin: 4px 0px;padding:8px 12px;border:1px solid #dcdee2;border-radius:4px;text-align: center;">
            <template v-for="(item, index) in dbModelConfig.addRow.endpoint_rel">
              <p :key="index + 'S'">
                <Tooltip :content="$t('m_db')" :delay="1000">
                  <Select v-model="item.target_endpoint" :placeholder="$t('m_target_value')" style="width:420px;">
                    <Option v-for="type in targetEndpoints" :key="type.guid" :value="type.guid">{{type.display_name}}</Option>
                  </Select>
                </Tooltip>
                <Tooltip :content="$t('m_source_value')" :delay="1000">
                  <Select v-model="item.source_endpoint" :placeholder="$t('m_source_value')" style="width:420px;">
                    <Option v-for="type in sourceEndpoints" :key="type.guid" :value="type.guid">{{type.display_name}}</Option>
                  </Select>
                </Tooltip>
                <Button
                  @click="deleteItem('endpoint_rel', index)"
                  size="small"
                  type="error"
                  icon="md-trash"
                ></Button>
              </p>
            </template>
            <Button
              @click="addEmptyItem('endpoint_rel')"
              type="success"
              size="small"
              long
              style="width:800px"
            >{{ $t('m_addMetricConfig') }}</Button>
          </div>
        </div>
      </template>
      <template slot="footer">
        <Button @click="cancelDb">{{$t('m_button_cancel')}}</Button>
        <Button @click="saveDb" type="primary">{{$t('m_button_save')}}</Button>
      </template>
    </BaseDrawer>
    <Modal
      v-model="isShowGroupMetricUpload"
      :title="$t('m_import')"
      :mask-closable="false"
      @on-ok="isShowGroupMetricUpload = false"
      @on-cancel="isShowGroupMetricUpload = false"
    >
      <div class="modal-body" style="padding:30px">
        <div style="display: inline-block;margin-bottom: 3px;">
          <Upload
            :action="uploadGroupMetricUrl"
            accept=".xlsx,.csv"
            :show-upload-list="false"
            :max-size="1000"
            with-credentials
            :headers="{Authorization: token}"
            :on-success="uploadSucess"
            :on-error="uploadFailed"
          >
            <Button icon="ios-cloud-upload-outline">{{$t('m_import')}}</Button>
          </Upload>
        </div>
      </div>
    </Modal>
    <CustomRegex ref="customRegexRef" @reloadMetricData="reloadMetricData"></CustomRegex>
    <BusinessMonitorGroupConfig ref="businessMonitorGroupConfigRef" @reloadMetricData="reloadMetricData"></BusinessMonitorGroupConfig>
  </div>
</template>

<script>
import {
  uniq, filter, cloneDeep, map, isEmpty
} from 'lodash'
// import Vue from 'vue'
import { getToken, getPlatFormToken } from '@/assets/js/cookies.ts'
import {baseURL_config} from '@/assets/js/baseURL'
import RegTest from '@/components/reg-test'
import CustomRegex from '@/views/monitor-config/log-template-config/custom-regex.vue'
import BusinessMonitorGroupConfig from '@/views/monitor-config/business-monitor-group-config.vue'
import axios from 'axios'
import {showPoptipOnTable} from '@/assets/js/utils.js'

export default {
  name: '',
  data() {
    return {
      selectedTemp: '', // 新增选中的模版
      parentGuid: '', // 新增在该数据下
      templateList: {
        json_list: [],
        regular_list: [],
        custom_list: []
      },
      token: null,
      MODALHEIGHT: 300,
      targetId: '',
      logFileDetail: [],
      showManagement: false,
      addAndEditModal: {
        isShow: false,
        isAdd: false,
        dataConfig: {
          service_group: '',
          log_path: '',
          monitor_type: 'process',
          endpoint_rel: []
        },
        pathOptions: [],
      },
      sourceEndpoints: [],
      targetEndpoints: [],
      showAddAndEditModal: false,
      activeData: {},
      regulationOption: [
        {
          label: this.$t('m_regular_match'),
          value: 1
        },
        {
          label: this.$t('m_irregular_matching'),
          value: 0
        }
      ],
      showRegConfig: false,
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
      selectedData: {},
      selectedIndex: null,
      deleteType: '',
      showCustomRegConfig: false,
      customMetricsModelConfig: {
        modalId: 'custom_metrics',
        isAdd: true,
        modalStyle: 'min-width:750px',
        modalTitle: 'm_metric_regular',
        saveFunc: 'saveCustomMetric',
        config: [
          {
            label: 'm_field_metric',
            value: 'metric',
            placeholder: '',
            disabled: false,
            type: 'text',
            max: 50
          },
          {
            label: 'm_field_displayName',
            value: 'display_name',
            placeholder: '',
            disabled: false,
            type: 'text'
          },
          // {label: 'm_tableKey_regular', value: 'regular', placeholder: 'm_tips_required', v_validate: 'required:true', disabled: false, type: 'text'},
          {
            name: 'ruleConfig',
            type: 'slot'
          }
        ],
        addRow: { // [通用]-保存用户新增、编辑时数据
          log_metric_monitor: '',
          display_name: '',
          agg_type: 'min',
          metric: null,
          regular: '',
          string_map: [],
          tag_config: []
        },
        slotConfig: {
          aggOption: ['sum', 'avg', 'count', 'max', 'min'],
          regulationOption: [
            {
              label: this.$t('m_regular_match'),
              value: 1
            },
            {
              label: this.$t('m_irregular_matching'),
              value: 0
            }
          ]
        }
      },
      modelTip: {
        key: '',
        value: 'metric'
      },
      dbModelConfig: {
        isShow: false,
        isAdd: true,
        addRow: {
          service_group: '',
          metric_sql: '',
          metric: '',
          display_name: '',
          monitor_type: 'process',
          step: 10,
          endpoint_rel: []
        }
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
      stepOptions: [
        {
          name: '10s',
          value: 10
        },
        {
          name: '30s',
          value: 30
        },
        {
          name: '1min',
          value: 60
        },
        {
          name: '5min',
          value: 300
        },
        {
          name: '10min',
          value: 600
        },
        {
          name: '1h',
          value: 3600
        },
        {
          name: '2h',
          value: 7200
        },
        {
          name: '12h',
          value: 43200
        },
        {
          name: '24h',
          value: 86400
        }
      ],
      isShowGroupMetricUpload: false,
      groupMetricId: '',
      typeToName: { // 模版枚举
        custom: this.$t('m_custom_regex'),
        regular: this.$t('m_standard_regex'),
        json: this.$t('m_standard_json'),
      },
      logFileCollapseValue: ['0'],
      logFileTableColumns: [
        {
          title: this.$t('m_configuration_name'),
          width: 150,
          key: 'name',
          render: (h, params) => (
            params.row.name
              ? <Tooltip placement="right" max-width="300" content={params.row.name + (params.row.metric_prefix_code ? (' [' + params.row.metric_prefix_code + ']') : '') }>
                <div class="table-ellipsis" style="width: 130px">{params.row.name + (params.row.metric_prefix_code ? (' [' + params.row.metric_prefix_code + ']') : '') } </div>
              </Tooltip> : <div>-</div>
          )
        },
        {
          title: this.$t('m_associated_template'),
          width: 200,
          ellipsis: true,
          tooltip: true,
          key: 'log_monitor_template_name',
          render: (h, params) => (
            params.row.log_monitor_template_name
              ? <Tooltip placement="top" max-width="300" content={params.row.log_monitor_template_name}>
                <div class="table-ellipsis" style="width: 180px">{params.row.log_monitor_template_name}</div>
              </Tooltip> : <div>-</div>
          )
        },
        {
          title: this.$t('m_metric_key'),
          minWidth: 300,
          ellipsis: true,
          tooltip: true,
          key: 'log_type_display',
          render: (h, params) => {
            let metricStr = ''
            let metricHtml = ''
            if (params.row.metric_list.length) {
              const metricArr = map(params.row.metric_list, 'full_metric')
              metricStr = metricArr.join(',')
              metricHtml = '<p>' + metricArr.join('<br>') + '</p>'
            }
            return (
              metricStr
                ? <Tooltip placement="top" max-width="400">
                  <div slot="content" style="white-space: normal;">
                    <div domPropsInnerHTML={metricHtml}></div>
                  </div>
                  <div class="table-ellipsis" style="width: 280px">{metricStr}</div>
                </Tooltip> : <div>-</div>
            )
          }
        },
        {
          title: this.$t('m_metric_config_type'),
          width: 150,
          key: 'log_type_display'
        },
        {
          title: this.$t('m_updatedBy'),
          key: 'update_user',
          width: 100,
        },
        {
          title: this.$t('m_title_updateTime'),
          minWidth: 100,
          key: 'update_time'
        },
        {
          title: this.$t('m_table_action'),
          width: 150,
          key: 'index',
          fixed: 'right',
          render: (h, params) => (
            <div style='display: flex'>
              <Tooltip placement="top" transfer content={this.$t('m_copy')}>
                <Button size="small" class="mr-1" type="success" on-click={() => this.copySingleItem(params.row)}>
                  <Icon type="md-document" size="16"></Icon>
                </Button>
              </Tooltip>
              <Tooltip placement="top" transfer content={this.$t('m_button_edit')}>
                <Button size="small" class="mr-1" type="primary" on-click={() => {
                  this.editRuleItem(params.row)
                }}>
                  <Icon type="md-create" size="16"></Icon>
                </Button>
              </Tooltip>
              <Poptip
                confirm
                title={this.$t('m_delConfirm_tip')}
                placement="left-end"
                on-on-ok={() => {
                  this.deleteType = 'rules'; this.okDelRow(params.row)
                }}>
                <Button size="small" type="error" class="mr-2" on-click={() => {
                  showPoptipOnTable()
                }}>
                  <Icon type="md-trash" size="16" />
                </Button>
              </Poptip>
            </div>
          )
        }
      ],
      dataBaseTableColumns: [
        {
          title: this.$t('m_metric_key'),
          width: 350,
          key: 'metric'
        },
        {
          title: this.$t('m_field_type'),
          key: 'monitor_type',
          minWidth: 100,
        },
        {
          title: this.$t('m_updatedBy'),
          key: 'update_user',
          width: 100,
        },
        {
          title: this.$t('m_title_updateTime'),
          minWidth: 100,
          key: 'update_time'
        },
        {
          title: this.$t('m_table_action'),
          width: 150,
          key: 'index',
          render: (h, params) => (
            <div style='display: flex'>
              <Tooltip placement="top" transfer content={this.$t('m_copy')}>
                <Button size="small" class="mr-1" type="success" on-click={() => this.copyDbItem(params.row)}>
                  <Icon type="md-document" size="16"></Icon>
                </Button>
              </Tooltip>
              <Tooltip placement="top" transfer content={this.$t('m_button_edit')}>
                <Button size="small" class="mr-1" type="primary" on-click={() => {
                  this.editDbItem(params.row)
                }}>
                  <Icon type="md-create" size="16"></Icon>
                </Button>
              </Tooltip>
              <Poptip
                confirm
                title={this.$t('m_delConfirm_tip')}
                placement="left-end"
                on-on-ok={() => {
                  this.deleteType = 'db'; this.okDelRow(params.row)
                }}>
                <Button size="small" type="error" class="mr-2">
                  <Icon type="md-trash" size="16" />
                </Button>
              </Poptip>
            </div>
          )
        }
      ],
      dataBaseTableData: [],
      allTemplateList: [],
      logAndDataBaseAllDetail: [],
      isEmpty,
      metricKey: '',
      addConfigType: ''
    }
  },
  computed: {
    uploadUrl() {
      return baseURL_config + `${this.$root.apiCenter.keywordImport}?serviceGroup=${this.targetId}`
    },
    uploadGroupMetricUrl() {
      return baseURL_config + `/monitor/api/v2/service/log_metric/log_metric_import/excel/${this.groupMetricId}`
    }
  },
  mounted() {
    this.MODALHEIGHT = document.body.scrollHeight - 300
    this.token = (window.request ? 'Bearer ' + getPlatFormToken() : getToken())|| null
    this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.logTemplateTableData, {}, resp => {
      this.templateList.json_list = resp.json_list || []
      this.templateList.regular_list = resp.regular_list || []
      this.templateList.custom_list = resp.custom_list || []
      this.allTemplateList = [{
        name: this.$t('m_standard_json'),
        value: 'm_standard_json',
        disabled: true
      }, ...resp.json_list, {
        name: this.$t('m_standard_regex'),
        value: 'm_standard_regex',
        disabled: true
      }, ...resp.regular_list, {
        name: this.$t('m_custom_regex'),
        value: 'm_custom_regex',
        disabled: true
      }, ...this.templateList.custom_list]
    })
  },
  methods: {
    importConfig(rowData) {
      this.groupMetricId = rowData.guid
      this.isShowGroupMetricUpload = true
    },
    exportData() {
      const api = `${this.$root.apiCenter.keywordExport}?serviceGroup=${this.targetId}`
      axios({
        method: 'GET',
        url: api,
        headers: {
          Authorization: this.token
        }
      }).then(response => {
        if (response.status < 400) {
          const content = JSON.stringify(response.data)
          const fileName = `${response.headers['content-disposition'].split(';')[1].trim().split('=')[1]}`
          const blob = new Blob([content])
          if ('msSaveOrOpenBlob' in navigator){
            // Microsoft Edge and Microsoft Internet Explorer 10-11
            window.navigator.msSaveOrOpenBlob(blob, fileName)
          } else {
            if ('download' in document.createElement('a')) { // 非IE下载
              const elink = document.createElement('a')
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
          this.$Message.warning(this.$t('m_tips_failed'))
        })
    },
    uploadSucess() {
      this.$Message.success(this.$t('m_tips_success'))
      this.getDetail(this.targetId)
    },
    uploadFailed(file) {
      this.$Message.warning(file.message)
    },
    // BD config
    delDbItem(rowData) {
      const api = this.$root.apiCenter.saveTargetDb + '/' + rowData.guid
      this.$root.$httpRequestEntrance.httpRequestEntrance('DELETE', api, '', () => {
        this.$Message.success(this.$t('m_tips_success'))
        this.getDetail(this.targetId)
      })
    },
    editDbItem(rowData) {
      this.getEndpoint(rowData.monitor_type, 'mysql')
      this.dbModelConfig.addRow = JSON.parse(JSON.stringify(rowData))
      this.dbModelConfig.isAdd = false
      this.dbModelConfig.isShow = true
    },
    getEndpointDefaultValue(monitorType) {
      this.dbModelConfig.addRow.endpoint_rel = this.dbModelConfig.addRow.endpoint_rel || []
      const api = `/monitor/api/v2/service/service_group/endpoint_rel?serviceGroup=${this.targetId}&sourceType=mysql&targetType=${monitorType}`
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', api, '', responseData => {
        if (!isEmpty(responseData)) {
          responseData.forEach(item => {
            this.dbModelConfig.addRow.endpoint_rel.push({
              target_endpoint: item.target_endpoint,
              source_endpoint: item.source_endpoint
            })
          })
        }
      })
    },
    copyDbItem(rowData) {
      this.getEndpoint(rowData.monitor_type, 'mysql')
      this.dbModelConfig.addRow = JSON.parse(JSON.stringify(rowData))
      this.dbModelConfig.addRow.metric += '1'
      this.dbModelConfig.isAdd = true
      this.dbModelConfig.isShow = true
    },
    addDb() {
      this.dbModelConfig.addRow = {
        service_group: '',
        metric_sql: '',
        metric: '',
        display_name: '',
        monitor_type: 'process',
        step: 10,
        endpoint_rel: []
      }
      this.getEndpoint('process', 'mysql', true)
      // this.getEndpointDefaultValue('process')
      this.dbModelConfig.isAdd = true
      this.dbModelConfig.isShow = true
    },
    saveDb() {
      // if (!this.dbModelConfig.addRow.display_name) {
      //   return this.$Message.error('显示名不能为空')
      // }
      if (!(/^[A-Za-z][A-Za-z0-9_]{0,48}[A-Za-z0-9]$/.test(this.dbModelConfig.addRow.metric))) {
        return this.$Message.error(this.$t('m_metric_key') + ':' + this.$t('m_regularization_check_failed_tips'))
      }
      if (!this.dbModelConfig.addRow.metric_sql) {
        return this.$Message.error('SQL不能为空')
      }
      if (!this.dbModelConfig.addRow.monitor_type) {
        return this.$Message.error('类型不能为空')
      }
      const endpointRelFlag = this.dbModelConfig.addRow.endpoint_rel.every(item => item.source_endpoint !== '' && item.target_endpoint !== '')
      if (!endpointRelFlag || this.dbModelConfig.addRow.endpoint_rel.length === 0) {
        return this.$Message.error('指标配置不能为空')
      }
      this.dbModelConfig.addRow.service_group = this.targetId
      const requestType = this.dbModelConfig.isAdd ? 'POST' : 'PUT'
      this.$root.$httpRequestEntrance.httpRequestEntrance(requestType, this.$root.apiCenter.saveTargetDb, this.dbModelConfig.addRow, () => {
        this.$Message.success(this.$t('m_tips_success'))
        this.dbModelConfig.isShow = false
        this.getDetail(this.targetId)
        this.getDbDetail(this.targetId)
      }, {isNeedloading: false})
    },
    cancelDb() {
      this.dbModelConfig.isShow = false
      this.dbModelConfig.addRow = {
        service_group: '',
        metric_sql: '',
        metric: '',
        display_name: '',
        monitor_type: '',
        endpoint_rel: []
      }
    },
    // other config
    editF(rowData) {
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
    updateReg(reg) {
      this.ruleModelConfig.addRow.json_regular = reg
      this.showRegConfig = false
    },
    cancelReg() {
      this.showRegConfig = false
    },
    updateCustomReg(reg) {
      this.customMetricsModelConfig.addRow.regular = reg
      this.showCustomRegConfig = false
    },
    cancelCustomReg() {
      this.showCustomRegConfig = false
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
      params.log_metric_monitor = this.activeData.guid
      if (!(params.regular.includes('(') && params.regular.includes(')'))) {
        this.$Message.error(this.$t('m_regular_tip'))
        return
      }
      const requestType = this.customMetricsModelConfig.isAdd ? 'POST' : 'PUT'
      this.$root.$httpRequestEntrance.httpRequestEntrance(requestType, this.$root.apiCenter.logMetricReg, params, () => {
        this.$Message.success(this.$t('m_tips_success'))
        this.$root.JQ('#custom_metrics').modal('hide')
        this.reloadMetricData(this.activeData.log_metric_monitor || this.activeData.guid)
        // this.getDetail(this.targetId)
      })
    },
    editCustomMetricItem(rowData) {
      this.activeData = rowData
      this.customMetricsModelConfig.isAdd = false
      this.modelTip.value = rowData.display_name
      this.customMetricsModelConfig.addRow = JSON.parse(JSON.stringify(rowData))
      this.$root.JQ('#custom_metrics').modal('show')
    },
    editRuleItem(rowData) {
      if (rowData.log_type === 'custom') {
        this.$refs.customRegexRef.loadPage('edit', '', rowData.log_metric_monitor, rowData.guid)
      } else {
        this.$refs.businessMonitorGroupConfigRef.loadPage('edit', rowData.log_monitor_template, rowData.log_metric_monitor, rowData.guid)
      }
    },
    copySingleItem(rowData) {
      if (rowData.log_type === 'custom') {
        this.$refs.customRegexRef.loadPage('copy', '', rowData.log_metric_monitor, rowData.guid)
      } else {
        this.$refs.businessMonitorGroupConfigRef.loadPage('copy', rowData.log_monitor_template, rowData.log_metric_monitor, rowData.guid)
      }
    },
    okDelRow(item) {
      if (this.deleteType === 'custom_metrics') {
        this.delCustomMericsItem(item)
      } else if (this.deleteType === 'db') {
        this.delDbItem(item)
      } else {
        this.delRuleItem(item)
      }
    },
    delCustomMericsItem(rowData) {
      const api = this.$root.apiCenter.logMetricReg + '/' + rowData.guid
      this.$root.$httpRequestEntrance.httpRequestEntrance('DELETE', api, '', () => {
        this.$Message.success(this.$t('m_tips_success'))
        this.reloadMetricData(rowData.log_metric_monitor)
      })
    },
    delRuleItem(rowData) {
      const api = this.$root.apiCenter.deleteLogMetricGroup + rowData.guid
      this.$root.$httpRequestEntrance.httpRequestEntrance('DELETE', api, '', () => {
        this.$Message.success(this.$t('m_tips_success'))
        this.reloadMetricData(rowData.log_metric_monitor)
      })
    },
    cancelRule() {
      this.ruleModelConfig.addRow = {
        log_metric_monitor: null,
        json_regular: null,
        tags: null,
        metric_list: []
      }
      this.ruleModelConfig.isShow = false
    },
    saveRule() {
      this.ruleModelConfig.addRow.log_metric_monitor = this.activeData.guid
      const requestType = this.ruleModelConfig.isAdd ? 'POST' : 'PUT'
      this.$root.$httpRequestEntrance.httpRequestEntrance(requestType, this.$root.apiCenter.logMetricJson, this.ruleModelConfig.addRow, () => {
        this.$Message.success(this.$t('m_tips_success'))
        this.ruleModelConfig.isShow = false
        this.reloadMetricData(this.activeData.guid || this.ruleModelConfig.addRow.pId)
      })
    },
    reloadMetricData() {
      this.getDetail(this.targetId)
    },
    singleAddF(rowData) {
      this.cancelReg()
      this.cancelRule()
      this.activeData = rowData
      this.ruleModelConfig.isAdd = true
      this.ruleModelConfig.isShow = true
    },
    deleteLogFiltItem(item) {
      this.delF(item)
    },
    delF(rowData) {
      const api = this.$root.apiCenter.deletePath + '/' + rowData.guid
      this.$root.$httpRequestEntrance.httpRequestEntrance('DELETE', api, '', () => {
        this.$Message.success(this.$t('m_tips_success'))
        this.getDetail(this.targetId)
      })
    },
    okAddAndEdit() {
      if (!this.addAndEditModal.dataConfig.monitor_type) {
        return this.$Message.error('类型不能为空')
      }
      if (this.addAndEditModal.isAdd) {
        const pathFlag = this.addAndEditModal.pathOptions.every(item => item.path !== '')
        if (!pathFlag || this.addAndEditModal.pathOptions.length === 0) {
          return this.$Message.error('日志路径不能为空')
        }
      } else {
        if (!this.addAndEditModal.dataConfig.log_path) {
          return this.$Message.error('日志路径不能为空')
        }
      }
      const endpointRelFlag = this.addAndEditModal.dataConfig.endpoint_rel.every(item => item.source_endpoint !== '' && item.target_endpoint !== '')
      if (!endpointRelFlag || this.addAndEditModal.dataConfig.endpoint_rel.length === 0) {
        return this.$Message.error('映射不能为空')
      }
      const params = JSON.parse(JSON.stringify(this.addAndEditModal.dataConfig))
      const methodType = this.addAndEditModal.isAdd ? 'POST' : 'PUT'
      params.service_group = this.targetId
      if (this.addAndEditModal.isAdd) {
        params.log_path = this.addAndEditModal.pathOptions.map(p => p.path)
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance(methodType, this.$root.apiCenter.logMetricMonitor, params, () => {
        this.$Message.success(this.$t('m_tips_success'))
        this.addAndEditModal.isShow = false
        this.getDetail(this.targetId)
      }, {isNeedloading: false})
    },
    cancelAddAndEdit() {
      this.addAndEditModal.isShow = false
      this.addAndEditModal.pathOptions = []
      this.addAndEditModal.dataConfig = {
        service_group: '',
        log_path: [],
        monitor_type: 'process',
        endpoint_rel: []
      }
    },
    async getEndpoint(val, type, needGetDefault = false) {
      this.addAndEditModal.dataConfig.endpoint_rel = []
      this.dbModelConfig.addRow.endpoint_rel = []
      if (needGetDefault) {
        await this.getDefaultConfig(val, type)
      }
      // get source Endpoint
      const sourceApi = this.$root.apiCenter.getEndpointsByType + '/' + this.targetId + '/endpoint/' + type
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', sourceApi, '', responseData => {
        this.sourceEndpoints = responseData
      }, {isNeedloading: false})
      const targetApi = this.$root.apiCenter.getEndpointsByType + '/' + this.targetId + '/endpoint/' + val
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', targetApi, '', responseData => {
        this.targetEndpoints = responseData
      }, {isNeedloading: false})
    },
    addEmptyItem(type, index) {
      switch (type) {
        case 'path': {
          this.addAndEditModal.pathOptions.push(
            {path: ''}
          )
          break
        }
        case 'relate': {
          this.addAndEditModal.dataConfig.endpoint_rel.push(
            {
              source_endpoint: '',
              target_endpoint: ''
            }
          )
          break
        }
        case 'metric_list': {
          this.ruleModelConfig.addRow[type].push({
            json_key: '',
            display_name: '',
            metric: '',
            agg_type: 'avg',
            string_map: [],
            tag_config: []
          })
          break
        }
        case 'string_map': {
          this.ruleModelConfig.addRow.metric_list[index][type].push({
            source_value: '',
            regulative: 0,
            target_value: ''
          })
          break
        }
        case 'tag_config': {
          this.ruleModelConfig.addRow.tag_config[index][type].push({
            key: '',
            regular: ''
          })
          break
        }
        case 'endpoint_rel': {
          this.dbModelConfig.addRow[type].push({
            source_endpoint: '',
            target_endpoint: ''
          })
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
        case 'metric_list': {
          this.ruleModelConfig.addRow[type].splice(index, 1)
          break
        }
        case 'string_map': {
          this.ruleModelConfig.addRow.metric_list[index][type].splice(index, 1)
          break
        }
        case 'tag_config': {
          this.ruleModelConfig.addRow.tag_config[index][type].splice(index, 1)
          break
        }
        case 'endpoint_rel': {
          this.dbModelConfig.addRow.endpoint_rel.splice(index, 1)
        }
      }
    },
    async addLogFile() {
      this.cancelAddAndEdit()
      this.getEndpoint(this.addAndEditModal.dataConfig.monitor_type, 'host', true)
      this.addAndEditModal.isAdd = true
      this.addAndEditModal.isShow = true
    },
    getDefaultConfig(val, type) {
      const api = `/monitor/api/v2/service/service_group/endpoint_rel?serviceGroup=${this.targetId}&sourceType=${type}&targetType=${val}`
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
        } else {
          tmp.forEach(t => {
            const find = this.dbModelConfig.addRow.endpoint_rel.find(rel => rel.source_endpoint === t.source_endpoint && rel.target_endpoint === t.target_endpoint)
            if (find === undefined) {
              this.dbModelConfig.addRow.endpoint_rel.push(t)
            }
          })
        }
      })
    },
    getLogKeyWordDetail() {
      return new Promise(resolve => {
        let api = this.$root.apiCenter.getTargetDetail + '/group/' + this.targetId
        if (this.metricKey) {
          api += `?metricKey=${this.metricKey}`
        }
        this.$root.$httpRequestEntrance.httpRequestEntrance('GET', api, '', responseData => {
          this.showManagement = true
          if (Array.isArray(responseData)) {
            this.logFileDetail = responseData
          } else {
            this.logFileDetail = [responseData]
          }

          this.logFileDetail.forEach(single => {
            single.config.forEach((item, index) => {
              const groups = item.metric_groups
              groups.forEach(group => {
                group.log_type_display = this.typeToName[group.log_type]
                group.logFileIndex = index
              })
            })
          })

          // this.logFileCollapseData = cloneDeep(responseData.config)
          // this.logFileCollapseData.forEach((item, index) => {
          //   const groups = item.metric_groups
          //   groups.forEach(group => {
          //     group.log_type_display = this.typeToName[group.log_type]
          //     group.logFileIndex = index
          //   })
          // })
          this.$root.$store.commit('changeTableExtendActive', -1)
          resolve(this.logFileDetail)
        }, {isNeedloading: true})
      })
    },
    async getDetail(targetId, metricKey) {
      if ((metricKey || metricKey === '') && this.metricKey !== metricKey) {
        this.metricKey = metricKey
        this.logFileCollapseValue = ['0', '1', '2', '3', '4', '5', '6', '7', '8', '9', '10']
      } else {
        this.logFileCollapseValue = ['0']
      }
      this.targetId = targetId
      await this.getLogKeyWordDetail()
      await this.getDbDetail()
      this.processAllInfo()
    },
    processAllInfo() {
      this.logAndDataBaseAllDetail = []
      const allDetail = [...cloneDeep(this.logFileDetail), ...cloneDeep(this.dataBaseTableData)]
      const allGuid = uniq(map(allDetail, 'guid')) || []

      allGuid.forEach(guid => {
        const tempInfo = {
          logFile: filter(this.logFileDetail, item => item.guid === guid)[0],
          database: filter(this.dataBaseTableData, item => item.guid === guid)[0]
        }
        this.logAndDataBaseAllDetail.push(tempInfo)
      })
    },
    getDbDetail() {
      return new Promise(resolve => {
        let api = this.$root.apiCenter.getTargetDbDetail + '/group/' + this.targetId
        if (this.metricKey) {
          api += `?metricKey=${this.metricKey}`
        }
        this.$root.$httpRequestEntrance.httpRequestEntrance('GET', api, '', responseData => {
          if (Array.isArray(responseData)) {
            this.dataBaseTableData = responseData
          } else {
            this.dataBaseTableData = [responseData]
          }
          // this.dataBaseTableData = responseData
          resolve(this.dataBaseTableData)
        }, {isNeedloading: false})
      })
    },
    // 新增自定指标指标
    addByCustom(item) {
      this.selectedTemp = 'customGuid'
      this.parentGuid = item.guid
      this.$refs.customRegexRef.loadPage('add', '', this.parentGuid, '')
      // this.okTempSelect()
    },
    okTempSelect() {
      if (this.addConfigType === 'custom') {
        this.$refs.customRegexRef.loadPage('add', '', this.parentGuid, this.selectedTemp, true)
      } else {
        const tmpList = this.templateList.json_list.concat(this.templateList.regular_list).concat(this.templateList.custom_list)
        const findTarget = tmpList.find(tmp => tmp.guid === this.selectedTemp)
        this.$refs.businessMonitorGroupConfigRef.loadPage('add', findTarget.guid, this.parentGuid, '')
      }
    },
    clearQuery() {
      this.$refs.selectRef.query = ''
    },
    // 新增指标配置--结束
  },
  components: {
    RegTest,
    CustomRegex,
    BusinessMonitorGroupConfig
  },
}
</script>

<style lang="less">
.ivu-table-wrapper {
  overflow: inherit;
}
.ivu-form-item {
  margin-bottom: 4px;
}
.success-btn {
  color: #fff;
  background-color: #19be6b;
  border-color: #19be6b;
}
.btn-img {
  width: 16px;
  vertical-align: middle;
}
.btn-left {
  margin-left: 8px;
}

.ivu-collapse-header {
  display: flex;
  align-items: center;
}

.log-file-table {
  .table-ellipsis {
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
}

.log-file-collapse-content {
  .ivu-dropdown-menu {
    max-height: 300px;
    overflow: auto;
  }
}

</style>

<style lang="less" scoped>
.w-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin: 10px 0;
  .title {
    font-size: 16px;
    font-weight: bold;
    margin: 0 10px;
    .underline {
      display: block;
      margin-top: -10px;
      margin-left: -6px;
      width: 100%;
      padding: 0 6px;
      height: 12px;
      border-radius: 12px;
      background-color: #c6eafe;
      box-sizing: content-box;
    }
  }
}

.use-underline-title {
  display: inline-block;
  font-size: 16px;
  font-weight: 700;
  margin: 0 10px;
  .underline {
    display: block;
    margin-top: -20px;
    margin-left: -6px;
    width: 100%;
    padding: 0 6px;
    height: 12px;
    border-radius: 12px;
    background-color: #c6eafe;
    -webkit-box-sizing: content-box;
    box-sizing: content-box;
  }
}

.log-file-collapse-content {
  width: 100%;
  display: flex;
  justify-content: space-between;
  align-items: center;
  .use-underline-title {
    font-size: 14px;
  }
}

.all-content {
  ::-webkit-scrollbar {
    position: relative;
    display: none;
  }
  .upload-content {
    display: flex;
    position: absolute;
    top: 70px;
    right: 26px;
    .btn-img {
      width: 16px;
      vertical-align: middle;
    }
  }
}

.no-data-class {
  display: flex;
  justify-content: center;
}
</style>
