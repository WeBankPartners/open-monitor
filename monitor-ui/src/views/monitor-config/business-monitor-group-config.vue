<template>
  <div>
    <Modal v-model="showModel" :title="$t('m_menu_configuration')" :mask-closable="false" :width="1100" :fullscreen="isfullscreen">
      <div slot="header" class="custom-modal-header">
        <span>
          {{(view ? $t('m_button_view') : (isAdd ? $t('m_button_add') : $t('m_modify'))) + $t('m_menu_configuration')}}
        </span>
        <Icon v-if="isfullscreen" @click="isfullscreen = !isfullscreen" class="fullscreen-icon" type="ios-contract" />
        <Icon v-else @click="isfullscreen = !isfullscreen" class="fullscreen-icon" type="ios-expand" />
      </div>
      <div :class="isfullscreen ? 'modal-container-fullscreen' : 'modal-container-normal'">
        <div class="w-header" slot="title">
          <div class="title">
            {{ $t('m_configuration_information') }}
            <span class="underline"></span>
          </div>
        </div>
        <Row>
          <Col span="8">
          <Form :label-width="120">
            <FormItem :label="$t('m_tableKey_name')">
              <Tooltip :content="businessConfig.name" transfer :disabled="businessConfig.name === ''" max-width="200" style="width: 100%;">
                <Input v-model.trim="businessConfig.name" :disabled="view" maxlength="30" show-word-limit style="width:96%"></Input>
                <span style="color: red">*</span>
              </Tooltip>
            </FormItem>
            <FormItem :label="$t('m_metric_code') ">
              <Input
                v-model.trim="businessConfig.metric_prefix_code"
                maxlength="6"
                :disabled="!isAdd || view"
                show-word-limit
                :placeholder="$t('m_metric_code_placeholder')"
                style="width:96%"
              >
              </Input>
              <span style="color: red">*</span>
            </FormItem>
          </Form>
          </Col>
          <Col span="16" style="border-left: 2px solid rgb(232 234 236)">
          <div style="padding: 0 8px">
            <div>
              <Row>
                <Col span="3" style="margin-top: 30px">
                <span style="color:#5cadff">{{ $t('m_service_code') }}</span>
                </Col>
                <Col span="20">
                <Row>
                  <Col span="4">{{ $t('m_match_type') }}</Col>
                  <Col span="6">
                  <span style="color:red">*</span>
                  {{ $t('m_source_value') }}</Col>
                  <Col span="6">
                  <span style="color:red">*</span>
                  {{ $t('m_match_value') }}</Col>
                  <Col span="2">
                  {{ $t('m_field_type') }}</Col>
                  <Col span="2"></Col>
                </Row>
                <Row v-for="(item, itemIndex) in businessConfig.code_string_map" :key="itemIndex" style="margin:6px 0">
                  <Col span="4">
                  <Select v-model="item.regulative" :disabled="view" style="width:90%">
                    <Option :value="1" key="m_regular_match">{{ $t('m_regular_match') }}</Option>
                    <Option :value="0" key="m_irregular_matching">{{ $t('m_irregular_matching') }}</Option>
                  </Select>
                  </Col>
                  <Col span="6">
                  <Input v-model.trim="item.source_value" :disabled="view" style="width:90%"></Input>
                  </Col>
                  <Col span="6">
                  <Input v-model.trim="item.target_value" :disabled="view" style="width:90%"></Input>
                  </Col>
                  <Col span="2" offset="2">
                  <Button
                    type="error"
                    ghost
                    @click="deleteItem('code_string_map',itemIndex)"
                    size="small"
                    style="vertical-align: sub;cursor: pointer"
                    icon="md-trash"
                    :disabled="view"
                  ></Button>
                  </Col>
                </Row>
                </Col>
              </Row>
              <Row>
                <Col span="20" offset="3">
                <Row>
                  <Col span="2" offset="18">
                  <div style="cursor: pointer">
                    <Button type="success" :disabled="view" ghost @click="addItem('code_string_map')" size="small" icon="md-add"></Button>
                  </div>
                  </Col>
                </Row>
                </Col>
              </Row>
            </div>
            <div>
              <Row>
                <Col span="3" style="margin-top: 12px">
                <span style="color:#5cadff">{{ $t('m_return_code') }}</span>
                </Col>
                <Col span="20">
                <Row v-for="(item, itemIndex) in businessConfig.retcode_string_map" :key="itemIndex" style="margin:6px 0">
                  <Col span="4">
                  <Select v-model="item.regulative" :disabled="view" style="width:90%">
                    <Option :value="1" key="m_regular_match">{{ $t('m_regular_match') }}</Option>
                    <Option :value="0" key="m_irregular_matching">{{ $t('m_irregular_matching') }}</Option>
                  </Select>
                  </Col>
                  <Col span="6">
                  <Input v-model.trim="item.source_value" :disabled="view" style="width:90%"></Input>
                  </Col>
                  <Col span="6">
                  <Input v-model.trim="item.target_value" :disabled="view" style="width:90%"></Input>
                  </Col>
                  <Col span="2">
                  <span style="line-height: 32px;">{{ $t('m_' + item.value_type) }}</span>
                  </Col>
                  <Col span="2">
                  <Button
                    type="error"
                    ghost
                    v-if="itemIndex !== 0"
                    @click="deleteItem('retcode_string_map',itemIndex)"
                    size="small"
                    :disabled="view"
                    style="cursor: pointer"
                    icon="md-trash"
                  ></Button>
                  </Col>
                </Row>
                </Col>
              </Row>
              <Row>
                <Col span="20" offset="3">
                <Row>
                  <Col span="2" offset="18">
                  <div style="cursor: pointer">
                    <Button type="success" :disabled="view" ghost @click="addItem('retcode_string_map')" size="small" icon="md-add"></Button>
                  </div>
                  </Col>
                </Row>
                </Col>
              </Row>
            </div>
          </div>
          </Col>
        </Row>
        <div class="w-header" slot="title">
          <div class="title">
            {{ $t('m_associated_template') }}
            <span class="underline"></span>
          </div>
          <Button type="primary" @click="changeTemplateStatus" ghost size="small" >{{templeteStatus ? $t('m_hide_template') : $t('m_expand_template')}}</Button>
        </div>
        <div>
          <StandardRegexDisplay
            ref="standardRegexDisplayRef"
            v-if="configInfo.log_type === 'regular'"
            :prefixCode="businessConfig.metric_prefix_code"
            :configInfo="configInfo"
          ></StandardRegexDisplay>
          <JsonRegexDisplay ref="jsonRegexDisplayRef"
                            v-if="configInfo.log_type === 'json'"
                            :prefixCode="businessConfig.metric_prefix_code"
                            :configInfo="configInfo"
          ></JsonRegexDisplay>
        </div>
      </div>
      <template #footer>
        <Button @click="showModel = false">{{ $t('m_button_cancel') }}</Button>
        <Button :disabled="view" @click="saveConfig" type="primary">{{ $t('m_button_save') }}</Button>
      </template>
    </Modal>
  </div>
</template>

<script>
import StandardRegexDisplay from '@/views/monitor-config/log-template-config/standard-regex-display.vue'
import JsonRegexDisplay from '@/views/monitor-config/log-template-config/json-regex-display.vue'
export default {
  name: '',
  data() {
    return {
      showModel: false,
      isAdd: true,
      view: false,
      isfullscreen: true,
      parentGuid: '', // 上级唯一标识
      configInfo: {
        log_type: ''
      },
      businessConfig: {},
      templeteStatus: false
    }
  },
  methods: {
    loadPage(actionType, templateGuid, parentGuid, configGuid) {
      this.templeteStatus = false
      this.isfullscreen = true
      this.parentGuid = parentGuid
      // actionType add/edit
      // templateGuid, 模版id
      // parentGuid, 上级唯一标识
      // configGuid, 配置唯一标志
      this.isAdd = actionType === 'add'
      this.view = actionType === 'view'
      if (configGuid) {
        this.getConfig(configGuid)
      }
      else {
        this.businessConfig = {
          name: '', // 名称
          metric_prefix_code: '', // 指标编码 1到6个字符的字母、数字、下划线或短横线
          log_metric_monitor_guid: '',
          log_monitor_template_guid: '',
          code_string_map: [
            {
              regulative: 0,
              source_value: '',
              target_value: ''
            }
          ],
          retcode_string_map: [
            {
              regulative: 0,
              source_value: '',
              target_value: '',
              value_type: 'success'
            }
          ]
        }
      }

      if (templateGuid) {
        this.businessConfig.log_monitor_template_guid = templateGuid
        this.businessConfig.log_metric_monitor_guid = parentGuid
        this.getTemplateDetail(templateGuid)
      }

      this.$refs.standardRegexDisplayRef && this.$refs.standardRegexDisplayRef.hideTemplate()
      this.$refs.jsonRegexDisplayRef && this.$refs.jsonRegexDisplayRef.hideTemplate()
      this.showModel = true
    },
    changeTemplateStatus() {
      this.$refs.standardRegexDisplayRef && this.$refs.standardRegexDisplayRef.changeTemplateStatus()
      this.$refs.jsonRegexDisplayRef && this.$refs.jsonRegexDisplayRef.changeTemplateStatus()
      this.templeteStatus = this.getTemplateStatus() || false
    },
    getTemplateStatus() {
      const displayRef = this.configInfo.log_type==='regular' ? 'standardRegexDisplayRef' : 'jsonRegexDisplayRef'
      const res = this.$refs[displayRef] && this.$refs[displayRef].returnCurrentStatus()
      return res
    },
    getTemplateDetail(guid) {
      const api = this.$root.apiCenter.getConfigDetailByGuid + guid
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', api, {}, resp => {
        this.configInfo = resp
      })
    },
    getConfig(guid) {
      const api = this.$root.apiCenter.getLogMetricConfig + guid
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', api, {}, resp => {
        this.businessConfig = resp
      })
    },
    paramsValidate(tmpData) {
      if (tmpData.name === '') {
        this.$Message.warning(`${this.$t('m_tableKey_name')}: ${this.$t('m_cannot_be_empty')}`)
        return true
      }
      if (tmpData.metric_prefix_code === '') {
        this.$Message.warning(`${this.$t('m_metric_code')}: ${this.$t('m_cannot_be_empty')}`)
        return true
      }
      // eslint-disable-next-line no-useless-escape
      // const regex = /^[a-zA-Z0-9_\-]{1,6}$/;
      const regex = /^[A-Za-z][A-Za-z0-9]{0,5}$/

      if (!regex.test(tmpData.metric_prefix_code)) {
        this.$Message.warning(`${this.$t('m_metric_code')}: ${this.$t('m_metric_prefix_code_validate')}`)
        return true
      }
      const isCodeMapEmpty = tmpData.code_string_map.some(element => element.source_value === '' || element.target_value === '')
      if (isCodeMapEmpty) {
        this.$Message.warning(`${this.$t('m_service_code')}: ${this.$t('m_fields_cannot_be_empty')}`)
        return true
      }

      const isRetcodeMapEmpty = tmpData.retcode_string_map.some(element => element.source_value === '' || element.target_value === '')
      if (isRetcodeMapEmpty) {
        this.$Message.warning(`${this.$t('m_return_code')}: ${this.$t('m_fields_cannot_be_empty')}`)
        return true
      }

      return false
    },

    saveConfig() {
      const tmpData = JSON.parse(JSON.stringify(this.businessConfig))
      if (this.paramsValidate(tmpData)) {
        return
      }
      // delete tmpData.create_user
      // delete tmpData.create_time
      // delete tmpData.update_user
      // delete tmpData.update_time
      const methodType = this.isAdd ? 'POST' : 'PUT'
      this.$root.$httpRequestEntrance.httpRequestEntrance(methodType, this.$root.apiCenter.logMetricGroup, tmpData, () => {
        this.$Message.success(this.$t('m_tips_success'))
        this.showModel = false
        this.$emit('reloadMetricData', this.parentGuid)
      })
    },
    addItem(key) {
      const params = key === 'code_string_map' ? {
        regulative: 0,
        source_value: '',
        target_value: ''
      } : {
        regulative: 0,
        source_value: '',
        target_value: '',
        value_type: 'fail'
      }
      this.businessConfig[key].push(params)
    },
    deleteItem(key, itemIndex) {
      this.businessConfig[key].splice(itemIndex, 1)
    }
  },
  components: {
    StandardRegexDisplay,
    JsonRegexDisplay
  }
}
</script>

<style lang="less" scoped>
.modal-container-normal {
  height: ~"calc(100vh - 280px)";
  overflow: auto;
}
.modal-container-fullscreen {
  height: ~"calc(100vh - 150px)";
  overflow: auto;
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
