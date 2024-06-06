<template>
  <div class=" ">
    <section v-if="showManagement" style="margin-top: 16px;">
      <div class="w-header" slot="title" style="display: flex;justify-content: space-between; margin-bottom: 12px;">
        <div class="title">
          {{$t('m_log_file')}}
          <span class="underline"></span>
        </div>
        <div>
          <div style="display: inline-block;position: relative;bottom: 48px;left: 66px;">
            <Button
              type="info"
              class="btn-left"
              @click="exportData"
            >
              <img src="../../assets/img/export.png" class="btn-img" alt="" />
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
                :on-error="uploadFailed">
                  <Button type="primary" class="btn-left">
                    <img src="../../assets/img/import.png" class="btn-img" alt="" />
                    {{ $t('m_import') }}
                  </Button>
              </Upload>
            </div>
          </div>
          <Button type="success" class="btn-right" @click="add">
            {{ $t('button.add') }}
          </Button>
        </div>
      </div>

      <PageTable :pageConfig="pageConfig">
        <div slot='tableExtend'>
          <div style="margin: 12px 0;float: right;">
            <Button type="success" @click="addByCustom(pageConfig.table.isExtend.parentData)" style="margin: 0 12px;">{{ $t('m_use_custom') }}</Button>
            <Button type="success" @click="addMetricConfig(pageConfig.table.isExtend.parentData)">{{ $t('m_use_template') }}</Button>
          </div>
          <div>
            <extendTable :detailConfig="pageConfig.table.isExtend.detailConfig"></extendTable>
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
        <div v-if="addAndEditModal.isAdd" style="margin: 4px 0px;padding:8px 12px;border:1px solid #dcdee2;border-radius:4px;width:680px;text-align: center;">
          <template v-for="(item, index) in addAndEditModal.pathOptions">
            <p :key="index + 5">
              <Tooltip :content="$t('tableKey.logPath')" :delay="1000">
                <Input v-model="item.path" style="width: 620px" :placeholder="$t('tableKey.logPath')" />
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
            >{{ $t('button.add') }}{{$t('tableKey.logPath')}}</Button
          >
        </div>
        <div v-else style="margin: 8px 0">
          <span>{{$t('tableKey.path')}}:</span>
          <Input style="width: 640px" v-model="addAndEditModal.dataConfig.log_path" />
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
      :mask-closable="false"
      >
      <div :style="{ 'max-height': MODALHEIGHT + 'px', overflow: 'auto' }">
        <Form :label-width="100">
          <FormItem :label="$t('tableKey.tags')">
            <Input v-model="ruleModelConfig.addRow.tags" style="width:100%" />
          </FormItem>
          <FormItem :label="$t('tableKey.regular')">
            <Input type="textarea" v-model="ruleModelConfig.addRow.json_regular" style="width: 580px"/>
            <Button v-if="!showRegConfig" @click="showRegConfig = !showRegConfig">{{$t('menu.configuration')}}</Button>
          </FormItem>
        </Form>
        <RegTest v-if="showRegConfig" @updateReg="updateReg" @cancelReg="cancelReg"></RegTest>
        <div style="margin: 4px 0px;padding:8px 12px;border:1px solid #dcdee2;border-radius:4px;">
          <template v-for="(item, index) in ruleModelConfig.addRow.metric_list">
            <p :key="index + 3">
              <Tooltip :content="$t('m_key')" :delay="1000">
                <Input v-model="item.json_key" style="width: 190px" :placeholder="$t('m_key') + ' e.g:[.*][.*]'" />
              </Tooltip>
              <Tooltip :content="$t('field.metric') + ' , e.g:code'" :delay="1000">
                <Input v-model="item.metric" style="width: 190px" :placeholder="$t('field.metric') + ' , e.g:code'" />
              </Tooltip>
              <Tooltip :content="$t('field.aggType')" :delay="1000">
                <Select v-model="item.agg_type" filterable clearable style="width:100px">
                  <Option v-for="agg in ruleModelConfig.aggOption" :value="agg" :key="agg">{{
                    agg
                  }}</Option>
                </Select>
              </Tooltip>
              <Tooltip :content="$t('field.displayName')" :delay="1000">
                <Input v-model="item.display_name" style="width: 160px" :placeholder="$t('field.displayName')" />
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
                  <Tooltip :content="$t('tableKey.regular')" :delay="1000">
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
                >{{ $t('addStringMap') }}</Button
              >
            </div>
            <Divider :key="index + 'Q'" />
          </template>
          <Button
            @click="addEmptyItem('metric_list')"
            type="success"
            size="small"
            long
            >{{ $t('addMetricConfig') }}123</Button
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
      :mask-closable="false"
      @on-ok="ok"
      @on-cancel="cancel">
      <div class="modal-body" style="padding:30px">
        <div style="text-align:center">
          <p style="color: red">{{$t('delConfirm.tip')}}: {{ selectedData.log_path || ''}}</p>
        </div>
      </div>
    </Modal>
    <Modal
      v-model="isShowWarningDelete"
      :title="$t('delConfirm.title')"
      :mask-closable="false"
      @on-ok="okDelRow"
      @on-cancel="cancleDelRow">
      <div class="modal-body" style="padding:30px">
        <div style="text-align:center">
          <p style="color: red">{{$t('delConfirm.tip')}}: {{ selectedData.name || selectedData.display_name || '' }}</p>
        </div>
      </div>
    </Modal>
    <ModalComponent :modelConfig="customMetricsModelConfig">
      <div slot="ruleConfig" class="extentClass">
        <div class="marginbottom params-each">
          <label class="col-md-2 label-name">{{$t('tableKey.regular')}}:</label>
          <Input style="width: 70%" type="textarea" v-model="customMetricsModelConfig.addRow.regular" />
          <Button v-if="!showCustomRegConfig" size="small" @click="showCustomRegConfig = !showCustomRegConfig">{{$t('menu.configuration')}}</Button>
        </div>
        <RegTest v-if="showCustomRegConfig" @updateReg="updateCustomReg" @cancelReg="cancelCustomReg"></RegTest>
        <div class="marginbottom params-each">
          <label class="col-md-2 label-name">{{$t('field.aggType')}}:</label>
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
                <Tooltip :content="$t('tableKey.regular')" :delay="1000">
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
              >{{ $t('addStringMap') }}</Button
            >
          </div>
        </div>
        <!-- 新增标签 -->
        <div class="marginbottom params-each">
          <div style="margin: 4px 12px;padding:8px 12px;border:1px solid #dcdee2;border-radius:4px">
            <template v-for="(item, index) in customMetricsModelConfig.addRow.tag_config">
              <p :key="index">
                <Tooltip :content="$t('tableKey.tags')" :delay="1000">
                  <Input v-model="item.key" style="width: 250px" :placeholder="$t('tableKey.tags')" />
                </Tooltip>
                <Tooltip :content="$t('tableKey.regular')" :delay="1000">
                  <Input v-model="item.regular" style="width: 400px" :placeholder="$t('tableKey.regular')" />
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
              >{{ $t('m_add_tags') }}</Button
            >
          </div>
        </div>
      </div>
    </ModalComponent>
    <!-- DB config -->
    <section v-if="showManagement" style="margin-top: 16px;">
      <div class="w-header" slot="title" style="display: flex;justify-content: space-between;">
        <div class="title">
          {{$t('m_db')}}
          <span class="underline"></span>
        </div>
        <Button type="success" class="btn-right" @click="addDb" style="margin: 8px 0">
          {{ $t('button.add') }}
        </Button>
      </div>
      <PageTable :pageConfig="pageDbConfig" style="margin-top:8px"></PageTable>
    </section>
    <Modal
      v-model="dbModelConfig.isShow"
      :title="$t('m_db')"
      width="700"
      :mask-closable="false"
      >
      <div :style="{ 'max-height': MODALHEIGHT + 'px', overflow: 'auto' }">
        <Form :label-width="100">
          <FormItem :label="$t('field.displayName')">
            <Input v-model="dbModelConfig.addRow.display_name" style="width:520px"/>
          </FormItem>
          <FormItem :label="$t('field.metric')">
            <Input v-model="dbModelConfig.addRow.metric" style="width:520px" />
          </FormItem>
          <FormItem label="SQL">
            <Input v-model="dbModelConfig.addRow.metric_sql" type="textarea" style="width:520px" />
          </FormItem>
          <FormItem :label="$t('field.type')">
            <Select v-model="dbModelConfig.addRow.monitor_type" @on-change="getEndpoint(dbModelConfig.addRow.monitor_type, 'mysql')" style="width: 520px">
              <Option v-for="type in monitorTypeOptions" :key="type.value" :value="type.label">{{type.label}}</Option>
            </Select>
          </FormItem>
        </Form>
        <div style="margin: 4px 0px;padding:8px 12px;border:1px solid #dcdee2;border-radius:4px;text-align: center;">
          <template v-for="(item, index) in dbModelConfig.addRow.endpoint_rel">
            <p :key="index + 'S'">
              <Tooltip :content="$t('m_db')" :delay="1000">
                <Select v-model="item.target_endpoint" style="width: 290px" :placeholder="$t('m_target_value')">
                  <Option v-for="type in targetEndpoints" :key="type.guid" :value="type.guid">{{type.display_name}}</Option>
                </Select>
              </Tooltip>
              <Tooltip :content="$t('m_source_value')" :delay="1000">
                <Select v-model="item.source_endpoint" style="width: 290px" :placeholder="$t('m_source_value')">
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
            style="width:610px"
            >{{ $t('addMetricConfig') }}</Button
          >
        </div>
      </div>
      <div slot="footer">
        <Button @click="cancelDb">{{$t('button.cancel')}}</Button>
        <Button @click="saveDb" type="primary">{{$t('button.save')}}</Button>
      </div>
    </Modal>

    <Modal
      v-model="isShowGroupMetricUpload"
      :title="$t('m_import')"
      :mask-closable="false"
      @on-ok="isShowGroupMetricUpload = false"
      @on-cancel="isShowGroupMetricUpload = false">
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
          :on-error="uploadFailed">
            <Button icon="ios-cloud-upload-outline">{{$t('m_import')}}</Button>
          </Upload>
        </div>
      </div>
    </Modal>

    <!-- 新增指标配置 -->
    <Modal v-model="showTemplateSelect" :title="$t('m_select_template')" :mask-closable="false">
      <Form :label-width="120">
        <FormItem :label="$t('m_log_template')">
          <Select style="width: 80%" v-model="selectedTemp" filterable ref="selectRef" @on-open-change="clearQuery">
            <OptionGroup :label="$t('m_standard_json')">
              <Option v-for="item in templateList.json_list" :value="item.guid" :key="item.guid">{{ item.name }}</Option>
            </OptionGroup>
            <OptionGroup :label="$t('m_standard_regex')">
              <Option v-for="item in templateList.regular_list" :value="item.guid" :key="item.guid">{{ item.name }}</Option>
            </OptionGroup>
            <!-- <OptionGroup :label="$t('m_custom_regex')">
              <Option value="customGuid" key="customGuid">{{ $t('m_custom_regex')}}</Option>
            </OptionGroup> -->
          </Select>
        </FormItem>
      </Form>
      <template #footer>
        <Button @click="showTemplateSelect = false">{{ $t('button.cancel') }}</Button>
        <Button @click="okTempSelect" :disabled="selectedTemp === ''" type="primary">{{ $t('button.confirm') }}</Button>
      </template>
    </Modal>
    <CustomRegex ref="customRegexRef" @reloadMetricData="reloadMetricData"></CustomRegex>
    <BusinessMonitorGroupConfig ref="businessMonitorGroupConfigRef" @reloadMetricData="reloadMetricData"></BusinessMonitorGroupConfig>
  </div>
</template>

<script>
import { getToken, getPlatFormToken } from '@/assets/js/cookies.ts'
import {baseURL_config} from '@/assets/js/baseURL'
import RegTest from '@/components/reg-test'
import CustomRegex from '@/views/monitor-config/log-template-config/custom-regex.vue'
import BusinessMonitorGroupConfig from '@/views/monitor-config/business-monitor-group-config.vue'
import extendTable from '@/components/table-page/extend-table'
import axios from 'axios'
let tableEle = [
  {title: 'tableKey.logPath', value: 'log_path', display: true},
  {title: 'field.type', value: 'monitor_type', display: true},
]
const btn = [
  {btn_name: 'button.edit', btn_func: 'editF'},
  {btn_name: 'button.remove', btn_func: 'deleteConfirmModal', color: 'red'},
  // {btn_name: 'm_import', btn_func: 'importConfig'}
]

let tableDbEle = [
  {title: 'm_metric_name', value: 'display_name', display: true},
  {title: 'm_metric_key', value: 'metric', display: true},
  {title: 'field.type', value: 'monitor_type', display: true}
]
const btnDb = [
  {btn_name: 'button.edit', btn_func: 'editDbItem'},
  {btn_name: 'button.remove', btn_func: 'deleteDbConfirmModal', color: 'red'}
]
export default {
  name: '',
  data () {
    return {
      showTemplateSelect: false,
      selectedTemp: '', // 新增选中的模版
      parentGuid: '', // 新增在该数据下
      templateList: {
        json_list: [],
        regular_list: []
      },
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
                {title: 'm_configuration_name', value: 'name', display: true},
                {title: 'm_associated_template', value: 'log_monitor_template_name', display: true},
                {title: 'm_metric_config_type', value: 'log_type_display', display: true},
                {title: 'm_updatedBy', value: 'update_user', display: true},
                {title: 'title.updateTime', value: 'update_time', display: true},
                {title: 'table.action',btn:[
                  {btn_name: 'button.edit', btn_func: 'editRuleItem'},
                  {btn_name: 'button.remove', btn_func: 'delRuleconfirmModal', color: 'red'}
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
      regulationOption: [
        {label: this.$t('m_regular_match'), value: 1},
        {label: this.$t('m_irregular_matching'), value: 0}
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
      isShowWarningDelete: false,
      deleteType: '',
      showCustomRegConfig: false,
      customMetricsModelConfig: {
        modalId: 'custom_metrics',
        isAdd: true,
        modalStyle: 'min-width:750px',
        modalTitle: 'm_metric_regular',
        saveFunc: 'saveCustomMetric',
        config: [
          {label: 'field.metric', value: 'metric', placeholder: '', disabled: false, type: 'text', max: 50},
          {label: 'field.displayName', value: 'display_name', placeholder: '', disabled: false, type: 'text'},
          // {label: 'tableKey.regular', value: 'regular', placeholder: 'tips.required', v_validate: 'required:true', disabled: false, type: 'text'},
          {name:'ruleConfig',type:'slot'}
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
            {label: this.$t('m_regular_match'), value: 1},
            {label: this.$t('m_irregular_matching'), value: 0}
          ]
        }
      },
      modelTip: {
        key: '',
        value: 'metric'
      },
      // DB config 
      pageDbConfig: {
        table: {
          tableData: [],
          tableEle: tableDbEle,
          // filterMoreBtn: 'filterMoreBtn',
          primaryKey: 'id',
          btn: btnDb,
          handleFloat:true
        }
      },
      dbModelConfig: {
        isShow: false,
        isAdd: true,
        addRow: {
          service_group: '',
          metric_sql: '',
          metric: '',
          display_name: '',
          monitor_type: '',
          endpoint_rel: []
        }
      },
      monitorTypeOptions: [
        {label: 'process', value: 'process'},
        {label: 'java', value: 'java'},
        {label: 'nginx', value: 'nginx'},
        {label: 'http', value: 'http'},
        {label: 'mysql', value: 'mysql'}
      ],
      isShowGroupMetricUpload: false,
      groupMetricId: '',
      typeToName: { // 模版枚举
        custom: this.$t('m_custom_regex'),
        regular: this.$t('m_standard_regex'),
        json: this.$t('m_standard_json'),
      }
    }
  },
  computed: {
    uploadUrl: function() {
      return baseURL_config + `${this.$root.apiCenter.keywordImport}?serviceGroup=${this.targrtId}`
    },
    uploadGroupMetricUrl: function() {
      return baseURL_config + `/monitor/api/v2/service/log_metric/log_metric_import/excel/${this.groupMetricId}`
    }
  },
  mounted () {
    this.MODALHEIGHT = document.body.scrollHeight - 300
    this.token = (window.request ? 'Bearer ' + getPlatFormToken() : getToken())|| null
  },
  methods: {
    importConfig (rowData) {
      this.groupMetricId = rowData.guid
      this.isShowGroupMetricUpload = true
    },
    exportData () {
      const api = `${this.$root.apiCenter.keywordExport}?serviceGroup=${this.targrtId}`
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
    // BD config
    delDbItem (rowData) {
      const api = this.$root.apiCenter.saveTargetDb + '/' + rowData.guid
      this.$root.$httpRequestEntrance.httpRequestEntrance('DELETE', api, '', () => {
        this.$Message.success(this.$t('tips.success'))
        this.getDetail(this.targrtId)
      })
    },
    editDbItem (rowData) {
      this.getEndpoint(rowData.monitor_type, 'mysql')
      this.dbModelConfig.addRow = JSON.parse(JSON.stringify(rowData))
      this.dbModelConfig.isAdd = false
      this.dbModelConfig.isShow = true
    },
    addDb () {
      this.dbModelConfig.isAdd = true
      this.dbModelConfig.isShow = true
    },
    saveDb () {
      this.dbModelConfig.addRow.service_group = this.targrtId
      const requestType = this.dbModelConfig.isAdd ? 'POST' : 'PUT'
      this.$root.$httpRequestEntrance.httpRequestEntrance(requestType, this.$root.apiCenter.saveTargetDb, this.dbModelConfig.addRow, () => {
        this.$Message.success(this.$t('tips.success'))
        this.dbModelConfig.isShow = false
        this.getDbDetail(this.targrtId)
      }, {isNeedloading:false})
    },
    cancelDb () {
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
    getDbDetail (targetId) {
      const api = this.$root.apiCenter.getTargetDbDetail + '/group/' + targetId
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', api, '', (responseData) => {
        this.pageDbConfig.table.tableData = responseData
      }, {isNeedloading:false})
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
    updateReg (reg) {
      this.ruleModelConfig.addRow.json_regular = reg
      this.showRegConfig = false
    },
    cancelReg () {
      this.showRegConfig = false
    },
    updateCustomReg (reg) {
      this.customMetricsModelConfig.addRow.regular = reg
      this.showCustomRegConfig = false
    },
    cancelCustomReg () {
      this.showCustomRegConfig = false
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
      params.log_metric_monitor = this.activeData.guid
      if (!(params.regular.includes('(') && params.regular.includes(')'))) {
        this.$Message.error(this.$t('m_regular_tip'))
        return
      }
      const requestType = this.customMetricsModelConfig.isAdd ? 'POST' : 'PUT'
      this.$root.$httpRequestEntrance.httpRequestEntrance(requestType, this.$root.apiCenter.logMetricReg, params, () => {
        this.$Message.success(this.$t('tips.success'))
        this.$root.JQ('#custom_metrics').modal('hide')
        this.reloadMetricData(this.activeData.log_metric_monitor || this.activeData.guid)
        // this.getDetail(this.targrtId)
      })
    },
    editCustomMetricItem (rowData) {
      this.activeData = rowData
      this.customMetricsModelConfig.isAdd = false
      this.modelTip.value = rowData.display_name
      this.customMetricsModelConfig.addRow = JSON.parse(JSON.stringify(rowData))
      this.$root.JQ('#custom_metrics').modal('show')
    },
    delCustomMetricConfirmModal (rowData) {
      this.selectedData = rowData
      this.isShowWarningDelete = true
      this.deleteType = 'custom_metrics'
    },
    editRuleItem (rowData) {
      // this.cancelReg()
      // this.ruleModelConfig.isAdd = false
      // this.ruleModelConfig.addRow = JSON.parse(JSON.stringify(rowData))
      // this.ruleModelConfig.isShow = true

      if (rowData.log_type === 'custom') {
        this.$refs.customRegexRef.loadPage('edit', '', rowData.log_metric_monitor, rowData.guid)
      } else {
        this.$refs.businessMonitorGroupConfigRef.loadPage('edit', rowData.log_monitor_template, rowData.log_metric_monitor, rowData.guid)
      }
    },
    deleteDbConfirmModal (rowData) {
      this.selectedData = rowData
      this.isShowWarningDelete = true
      this.deleteType = 'db'
    },
    delRuleconfirmModal (rowData) {
      this.selectedData = rowData
      this.isShowWarningDelete = true
      this.deleteType = 'rules'
    },
    okDelRow () {
      if (this.deleteType === 'custom_metrics') {
        this.delCustomMericsItem(this.selectedData)
      } else if (this.deleteType === 'db') {
        this.delDbItem(this.selectedData)
      } else {
        this.delRuleItem(this.selectedData)
      }
    },
    delCustomMericsItem (rowData) {
      const api = this.$root.apiCenter.logMetricReg + '/' + rowData.guid
      this.$root.$httpRequestEntrance.httpRequestEntrance('DELETE', api, '', () => {
        this.$Message.success(this.$t('tips.success'))
        // this.getDetail(this.targrtId)
        this.reloadMetricData(rowData.log_metric_monitor)
      })
    },
    cancleDelRow () {
      this.isShowWarningDelete = false
    },
    delRuleItem (rowData) {
      const api = this.$root.apiCenter.deleteLogMetricGroup + rowData.guid
      this.$root.$httpRequestEntrance.httpRequestEntrance('DELETE', api, '', () => {
        this.$Message.success(this.$t('tips.success'))
        this.isShowWarningDelete = false
        this.reloadMetricData(rowData.log_metric_monitor)
        // this.getDetail(this.targrtId)
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
        this.reloadMetricData(this.activeData.guid || this.ruleModelConfig.addRow.pId)
        // this.getDetail(this.targrtId)
      })
    },
    reloadMetricData (guid) {
      const path = `${this.$root.apiCenter.getLogMetricByPath}/${guid}`
      this.$root.$httpRequestEntrance.httpRequestEntrance("GET", path, {}, (responseData) => {
        this.pageConfig.table.isExtend.detailConfig[0].data = responseData.metric_groups.map(group => {
        group.log_type_display = this.typeToName[group.log_type]
        return group
      })
      })
    },
    singleAddF (rowData) {
      this.cancelReg()
      this.cancelRule()
      this.activeData = rowData
      this.ruleModelConfig.isAdd = true
      this.ruleModelConfig.isShow = true
    },
    getExtendInfo(item){
      const path = `${this.$root.apiCenter.getLogMetricByPath}/${item.guid}`
      this.$root.$httpRequestEntrance.httpRequestEntrance("GET", path, {}, (responseData) => {
        this.pageConfig.table.isExtend.detailConfig[0].data = responseData.metric_groups.map(group => {
          group.log_type_display = this.typeToName[group.log_type]
          return group
        })
        this.pageConfig.table.isExtend.parentData = item
      })
      this.pageConfig.table.isExtend.parentData = item
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
      let params = JSON.parse(JSON.stringify(this.addAndEditModal.dataConfig))
      const methodType = this.addAndEditModal.isAdd ? 'POST' : 'PUT'
      params.service_group = this.targrtId
      if (this.addAndEditModal.isAdd) {
        params.log_path = this.addAndEditModal.pathOptions.map(p => p.path)
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance(methodType, this.$root.apiCenter.logMetricMonitor, params, () => {
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
      this.dbModelConfig.addRow.endpoint_rel = []
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
    addEmptyItem (type, index) {
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
    getDetail (targrtId) {
      this.targrtId = targrtId
      const api = this.$root.apiCenter.getTargetDetail + '/group/' + targrtId
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', api, '', (responseData) => {
        this.showManagement = true
        this.targetDetail = responseData
        this.pageConfig.table.tableData = responseData.config
        this.$root.$store.commit('changeTableExtendActive', -1)
      }, {isNeedloading:true})
      this.getDbDetail(targrtId)
    },
    // 新增指标配置--开始
    addMetricConfig (item) {
      this.parentGuid = item.guid
      this.selectedTemp = ''
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.logTemplateTableData, {}, (resp) => {
        this.templateList.json_list = resp.json_list
        this.templateList.regular_list = resp.regular_list
        this.showTemplateSelect = true
      })
    },
    // 新增自定指标指标
    addByCustom (item) {
      this.selectedTemp = 'customGuid'
      this.parentGuid = item.guid
      this.okTempSelect()
    },
    okTempSelect () {
      this.showTemplateSelect = false
      if (this.selectedTemp === 'customGuid') {
        this.$refs.customRegexRef.loadPage('add', '', this.parentGuid, '')
      } else {
        const tmpList = this.templateList.json_list.concat(this.templateList.regular_list);
        const findTarget = tmpList.find(tmp => tmp.guid === this.selectedTemp)
        this.$refs.businessMonitorGroupConfigRef.loadPage('add', findTarget.guid, this.parentGuid, '')
      }
    },
    clearQuery () {
      this.$refs.selectRef.query = ''
    },
    // 新增指标配置--结束
  },
  components: {
    extendTable,
    RegTest,
    CustomRegex,
    BusinessMonitorGroupConfig
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
.btn-img {
  width: 16px;
  vertical-align: middle;
}
.btn-left {
  margin-left: 8px;
}
.w-header {
  display: flex;
  align-items: center;
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
</style>
