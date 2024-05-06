<template>
  <div>
    <Modal v-model="showModel" :title="$t('menu.configuration')" :mask-closable="false" :width="1100" :fullscreen="isfullscreen">
      <div slot="header" class="custom-modal-header">
        <span>
          {{(isAdd ? $t('button.add') : $t('button.edit')) + $t('menu.configuration')}}
        </span>
        <Icon v-if="isfullscreen" @click="isfullscreen = !isfullscreen" class="fullscreen-icon" type="ios-contract" />
        <Icon v-else @click="isfullscreen = !isfullscreen" class="fullscreen-icon" type="ios-expand" />
      </div>
      <div :class="isfullscreen? 'modal-container-fullscreen':'modal-container-normal'">
        <Form :label-width="120" inline>
          <FormItem :label="$t('tableKey.name')">
            <Tooltip :content="businessConfig.name" transfer :disabled="businessConfig.name === ''" max-width="200">
              <Input v-model.trim="businessConfig.name" maxlength="30" show-word-limit style="width:220px"></Input>
              <span style="color: red">*</span>
            </Tooltip>
          </FormItem>
          <FormItem :label="$t('m_metric_code')">
            <Input v-model.trim="businessConfig.metric_prefix_code" maxlength="6" :disabled="!isAdd" show-word-limit style="width:220px"></Input>
            <span style="color: red">*</span>
          </FormItem>
        </Form>
        <Divider orientation="left" size="small">{{ $t('m_associated_template') }}</Divider>
        <div>
          <StandardRegexDisplay ref="standardRegexDisplayRef" v-if="configInfo.log_type==='regular'" :configInfo="configInfo"></StandardRegexDisplay>
          <JsonRegexDisplay ref="jsonRegexDisplayRef" v-if="configInfo.log_type==='json'" :configInfo="configInfo"></JsonRegexDisplay>
        </div>
        <div>
          <Divider orientation="left" size="small">{{ $t('m_configuration_information') }}</Divider>
          <span style="color:#5cadff">{{ $t('m_service_code') }}</span>
          <div>
            <Row>
              <Col span="4">{{ $t('m_match_type') }}</Col>
              <Col span="6">
                <span style="color:red">*</span>
                {{ $t('m_source_value_regular') }}</Col>
              <Col span="6">
                <span style="color:red">*</span>
                {{ $t('m_match_value') }}</Col>
            </Row>
            <Row v-for="(item, itemIndex) in businessConfig.code_string_map" :key="itemIndex" style="margin:6px 0">
              <Col span="4">
                <Select v-model="item.regulative" style="width:90%">
                  <Option :value="1" key="m_regular_match">{{ $t('m_regular_match') }}</Option>
                  <Option :value="0" key="m_irregular_matching">{{ $t('m_irregular_matching') }}</Option>
                </Select>
              </Col>
              <Col span="6">
                <Input v-model.trim="item.source_value" style="width:90%"></Input>
              </Col>
              <Col span="6">
                <Input v-model.trim="item.target_value" style="width:90%"></Input>
              </Col>
              <Col span="2">
                <Button
                  type="error"
                  ghost
                  @click="deleteItem('code_string_map',itemIndex)"
                  size="small"
                  style="vertical-align: sub;cursor: pointer"
                  icon="md-trash"
                ></Button>
              </Col>
            </Row>
            <div style="text-align: right;margin-right: 8px;cursor: pointer">
              <Button type="primary" ghost @click="addItem('code_string_map')" size="small" icon="md-add"></Button>
            </div>
          </div>
          <span style="color:#5cadff">{{ $t('m_return_code') }}</span>
          <div>
            <Row>
              <Col span="4">{{ $t('m_match_type') }}</Col>
              <Col span="6">
                <span style="color:red">*</span>
                {{ $t('m_source_value') }}</Col>
              <Col span="6">
                <span style="color:red">*</span>
                {{ $t('m_match_value') }}</Col>
            </Row>
            <Row v-for="(item, itemIndex) in businessConfig.retcode_string_map" :key="itemIndex" style="margin:6px 0">
              <Col span="4">
                <Select v-model="item.regulative" style="width:90%">
                  <Option :value="1" key="m_regular_match">{{ $t('m_regular_match') }}</Option>
                  <Option :value="0" key="m_irregular_matching">{{ $t('m_irregular_matching') }}</Option>
                </Select>
              </Col>
              <Col span="6">
                <Input v-model.trim="item.source_value" style="width:90%"></Input>
              </Col>
              <Col span="6">
                <Input v-model.trim="item.target_value" style="width:90%"></Input>
              </Col>
              <Col span="6">
                <span style="line-height: 32px;">{{ $t('m_'+item.value_type) }}</span>
                <Button
                  type="error"
                  ghost
                  v-if="itemIndex!==0"
                  @click="deleteItem('retcode_string_map',itemIndex)"
                  size="small"
                  style="margin-left:24px; cursor: pointer"
                  icon="md-trash"
                ></Button>
              </Col>
            </Row>
            <div style="text-align: right;margin-right: 8px;cursor: pointer">
              <Button type="primary" ghost @click="addItem('retcode_string_map')" size="small" icon="md-add"></Button>
            </div>
          </div>
        </div>
      </div>
      <template #footer>
        <Button @click="showModel=false">{{ $t('button.cancel') }}</Button>
        <Button @click="saveConfig" type="primary">{{ $t('button.save') }}</Button>
      </template>
    </Modal>
  </div>
</template>

<script>
import StandardRegexDisplay from '@/views/monitor-config/log-template-config/standard-regex-display.vue'
import JsonRegexDisplay from '@/views/monitor-config/log-template-config/json-regex-display.vue'
export default {
  name: '',
  data () {
    return {
      showModel: false,
      isAdd: true,
      isfullscreen: false,
      parentGuid: '', //上级唯一标识
      configInfo: {
        log_type: ''
      },
      businessConfig: {}
    }
  },
  methods: {
    loadPage (actionType, templateGuid, parentGuid, configGuid) {
      this.parentGuid = parentGuid
      // actionType add/edit
      // templateGuid, 模版id
      // parentGuid, 上级唯一标识
      // configGuid, 配置唯一标志 
      this.isAdd = actionType === 'add'

      if (configGuid) {
        this.getConfig(configGuid)
      } else {
        this.businessConfig = {
          name: '', // 名称
          metric_prefix_code: '', // 指标编码 1到6个字符的字母、数字、下划线或短横线
          log_metric_monitor_guid: '',
          log_monitor_template_guid: '',
          code_string_map: [
            // {
            //   regulative: 1,
            //   source_value: '',
            //   target_value: ''
            // }
          ],
          retcode_string_map: [
            {
              regulative: 1,
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
    getTemplateDetail(guid) {
      const api = this.$root.apiCenter.getConfigDetailByGuid + guid
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', api, {}, (resp) => {
        this.configInfo = resp
      })
    },
    getConfig(guid) {
      const api = this.$root.apiCenter.getLogMetricConfig + guid
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', api, {}, (resp) => {
        this.businessConfig = resp
      })
    },
    paramsValidate (tmpData) {
      if (tmpData.name === '') {
        this.$Message.warning(`${this.$t('tableKey.name')}: ${this.$t('m_cannot_be_empty')}`)
        return true
      }
      if (tmpData.metric_prefix_code === '') {
        this.$Message.warning(`${this.$t('m_metric_code')}: ${this.$t('m_cannot_be_empty')}`)
        return true
      }
      // eslint-disable-next-line no-useless-escape
      const regex = /^[a-zA-Z0-9_\-]{1,6}$/;
      
      if (!regex.test(tmpData.metric_prefix_code)) {
        this.$Message.warning(`${this.$t('m_metric_code')}: ${this.$t('m_metric_prefix_code_validate')}`)
        return true
      }
      const isCodeMapEmpty = tmpData.code_string_map.some((element) => {
        return element.source_value === '' || element.target_value === ''
      })
      if (isCodeMapEmpty) {
        this.$Message.warning(`${this.$t('m_service_code')}: ${this.$t('m_fields_cannot_be_empty')}`)
        return true
      }

      const isRetcodeMapEmpty = tmpData.retcode_string_map.some((element) => {
        return element.source_value === '' || element.target_value === ''
      })
      if (isRetcodeMapEmpty) {
        this.$Message.warning(`${this.$t('m_return_code')}: ${this.$t('m_fields_cannot_be_empty')}`)
        return true
      }

      return false
    },

    saveConfig () {
      let tmpData = JSON.parse(JSON.stringify(this.businessConfig))
      if (this.paramsValidate(tmpData)) return
      // delete tmpData.create_user
      // delete tmpData.create_time
      // delete tmpData.update_user
      // delete tmpData.update_time
      let methodType = this.isAdd ? 'POST' : 'PUT'
      this.$root.$httpRequestEntrance.httpRequestEntrance(methodType, this.$root.apiCenter.logMetricGroup, tmpData, () => {
        this.$Message.success(this.$t('tips.success'))
        this.showModel = false
        this.$emit('reloadMetricData', this.parentGuid)
      })
    },
    addItem (key) {
      let params = key === 'code_string_map' ? {
          regulative: 1,
          source_value: '',
          target_value: ''
        } : {
          regulative: 1,
          source_value: '',
          target_value: '',
          value_type: 'fail'
        }
      this.businessConfig[key].push(params)
    },
    deleteItem (key, itemIndex) {
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
  height: ~"calc(100vh - 100px)";
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
</style>
