<template>
  <div>
    <Modal
      v-model="showModal"
      :mask-closable="false"
      :fullscreen="isfullscreen"
      :width="1100"
    >
      <div slot="header" class="custom-modal-header">
        <span>
          {{(isAdd ? $t('button.add') : $t('button.edit')) + $t('m_custom_regex')}}
        </span>
        <Icon v-if="isfullscreen" @click="isfullscreen = !isfullscreen" class="fullscreen-icon" type="ios-contract" />
        <Icon v-else @click="isfullscreen = !isfullscreen" class="fullscreen-icon" type="ios-expand" />
      </div>
      <div :class="isfullscreen? 'modal-container-fullscreen':'modal-container-normal'">
        <Row>
          <Col span="8">
            <Form :label-width="120">
              <FormItem :label="$t('tableKey.name')">
                <Input
                  v-model="configInfo.name"
                  maxlength="30"
                  show-word-limit
                  style="width: 96%"
                />
                <span style="color: red">*</span>
                <div
                  v-if="isParmasChanged && (configInfo.name.length === 0 || configInfo.name.length > 30)"
                  style="color: red"
                >
                  {{ $t('m_template_name') }}{{ $t('tw_limit_30') }}
                </div>
              </FormItem>
              <FormItem :label="$t('m_log_example')">
                <Input
                  v-model="configInfo.demo_log"
                  type="textarea"
                  :rows="15"
                  style="width: 96%"
                />
                <div v-if="isParmasChanged && configInfo.demo_log.length === 0" style="color: red">
                  {{ $t('m_log_example') }} {{ $t('tips.required') }}
                </div>
              </FormItem>
            </Form>
          </Col>
          <Col span="16" style="border-left: 2px solid rgb(232 234 236)">
            <div style="margin-left: 8px">
              <!-- 采集参数 -->
              <div>
                <Divider orientation="left" size="small">{{ $t('m_parameter_collection') }}</Divider>
                <Table
                  style="position: inherit;"
                  size="small"
                  :columns="columnsForParameterCollection"
                  :data="configInfo.param_list"
                  width="100%"
                ></Table>
                <Button type="primary" @click="addParameterCollection" ghost size="small" style="float:left;margin:12px">{{ $t('m_add_parameter_collection') }}</Button>
                <Button type="primary" @click="generateBackstageTrial" ghost size="small" style="float:right;margin:12px">{{ $t('m_match') }}</Button>
              </div>
              <!-- 计算指标 -->
              <div>
                <Divider orientation="left" size="small">{{ $t('m_compute_metrics') }}</Divider>
                <Table
                  style="position: inherit;"
                  size="small"
                  :columns="columnsForComputeMetrics"
                  :data="configInfo.metric_list"
                  width="100%"
                ></Table>
                <Button type="primary" @click="addComputeMetrics" ghost size="small" style="float:left;margin:12px">{{ $t('m_add_compute_metrics') }}</Button>
              </div>
            </div>
          </Col>
        </Row>
      </div>
      <div slot="footer">
        <Button @click="showModal = false">{{ $t('button.cancel') }}</Button>
        <Button @click="saveConfig" type="primary">{{ $t('button.save') }}</Button>
      </div>
    </Modal>
    <TagMapConfig ref="tagMapConfigRef" @setTagMap="setTagMap"></TagMapConfig>
  </div>
</template>

<script>
import TagMapConfig from './tag-map-config.vue'
export default {
  name: "standard-regex",
  data() {
    return {
      showModal: false,
      isfullscreen: false,
      isParmasChanged: false,
      parentGuid: '', //上级唯一标识
      isAdd: true,
      configInfo: {},
      columnsForParameterCollection: [
        {
          title: this.$t('field.displayName'),
          key: 'display_name',
          render: (h, params) => {
            return (
              <Input
                value={params.row.display_name}
                onInput={v => {
                  this.changeVal('param_list', params.index, 'display_name', v)
                }}
              />
            )
          }
        },
        {
          title: this.$t('m_parameter_key'),
          key: 'name',
          render: (h, params) => {
            return (
              <Input
                value={params.row.name}
                onInput={v => {
                  this.changeVal('param_list', params.index, 'name', v)
                }}
              />
            )
          }
        },
        {
          title: this.$t('m_extract_regular'),
          key: 'regular',
          render: (h, params) => {
            return (
              <Input
                value={params.row.regular}
                onInput={v => {
                  this.changeVal('param_list', params.index, 'regular', v)
                }}
              />
            )
          }
        },
        {
          title: this.$t('m_matching_result'),
          ellipsis: true,
          tooltip: true,
          key: 'demo_match_value',
        },
        {
          title: this.$t('m_tag_mapping'),
          ellipsis: true,
          tooltip: true,
          key: 'string_map',
          render: (h, params) => {
            const val = params.row.string_map.map(item => item.target_value).join(',')
            return (
              <div>
                <Input disabled style="width:80%" value={val}/>
                <Button
                  size="small"
                  type="success"
                  onClick={() => this.editTagMapping(params.index)}
                >
                  <Icon type="ios-create-outline" size="16"></Icon>
                </Button>
              </div>
            )
          }
        },
        {
          title: this.$t('table.action'),
          key: 'action',
          width: 80,
          align: 'left',
          render: (h, params) => {
            return (
              <div style="text-align: left; cursor: pointer;display: inline-flex;">
                <Button
                  size="small"
                  type="error"
                  style="margin-right:5px;"
                  onClick={() => this.deleteAction('param_list', params.index)}
                >
                  <Icon type="md-trash" size="16"></Icon>
                </Button>
              </div>
            )
          }
        }
      ],
      columnsForComputeMetrics: [
        {
          title: this.$t('field.displayName'),
          key: 'display_name',
          render: (h, params) => {
            return (
              <Input
                value={params.row.display_name}
                onInput={v => {
                  this.changeVal('metric_list', params.index, 'display_name', v)
                }}
              />
            )
          }
        },
        {
          title: this.$t('m_metric_key'),
          key: 'metric',
          render: (h, params) => {
            return (
              <Input
                value={params.row.metric}
                onInput={v => {
                  this.changeVal('metric_list', params.index, 'metric', v)
                }}
              />
            )
          }
        },
        {
          title: this.$t('m_statistical_parameters'),
          key: 'log_param_name',
          render: (h, params) => {
            const keys = this.configInfo.param_list.map(p => {
              return p.name
            })
            const selectOptions = [...new Set(keys)]
            return (
              <Select
                filterable
                value={params.row.log_param_name}
                on-on-change={(v) => {
                  this.changeVal('metric_list', params.index, 'log_param_name', v)
                }}
              >
                {selectOptions.map(option => (
                  <Option key={option} value={option}>
                    {option}
                  </Option>
                ))}
              </Select>
            )
          }
        },
        {
          title: this.$t('m_filter_label'),
          key: 'tag_config',
          render: (h, params) => {
            const keys = this.configInfo.param_list.map(p => {
              return p.name
            })
            const selectOptions = [...new Set(keys)]
            return (
              <Select
                filterable
                value={params.row.tag_config}
                multiple
                on-on-change={(v) => {
                  this.changeVal('metric_list', params.index, 'tag_config', v)
                }}
              >
                {selectOptions.map(option => (
                  <Option key={option} value={option}>
                    {option}
                  </Option>
                ))}
              </Select>
            )
          }
        },
        {
          title: this.$t('m_computed_type'),
          key: 'agg_type',
          render: (h, params) => {
            const selectOptions = ['avg', 'count', 'max', 'min', 'sum']
            return (
              <Select
                filterable
                value={params.row.agg_type}
                on-on-change={(v) => {
                  this.changeVal('metric_list', params.index, 'agg_type', v)
                }}
              >
                {selectOptions.map(option => (
                  <Option key={option} value={option}>
                    {option}
                  </Option>
                ))}
              </Select>
            )
          }
        },
        {
          title: this.$t('table.action'),
          key: 'action',
          width: 80,
          align: 'left',
          render: (h, params) => {
            return (
              <div style="text-align: left; cursor: pointer;display: inline-flex;">
                <Button
                  size="small"
                  type="error"
                  style="margin-right:5px;"
                  onClick={() => this.deleteAction('metric_list', params.index)}
                >
                  <Icon type="md-trash" size="16"></Icon>
                </Button>
              </div>
            )
          }
        }
      ],
      editTagMappingIndex: -1, // 正在编辑的参数采集
      
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
        this.configInfo = {
          guid: '',
          log_metric_monitor: '',
          name: '',
          log_type: 'custom',
          demo_log: '',
          param_list: [
            // {
            //   guid: '',
            //   name: '',
            //   display_name: '',
            //   json_key: '',
            //   regular: '',
            //   demo_match_value: '',
            //   string_map: [
            //     {
            //       regulative: 1,  //匹配类型： 0 是非正则，1是正则
            //       source_value: '', // 源值
            //       target_value: '', // 映射值
            //     }
            //   ]
            // }
          ],
          metric_list: [
            // {
            //   log_param_name: 'code',
            //   metric: 'req_count',
            //   display_name: this.$t('m_request_volume'),
            //   agg_type: 'count',
            //   tag_config: [
            //     'code'
            //   ]
            // }
          ]
        }
        this.configInfo.log_metric_monitor = parentGuid
      }

      this.showModal = true
    },
    paramsValidate (tmpData) {
      if (tmpData.name === '') {
        this.$Message.warning(`${this.$t('tableKey.name')}${this.$t('m_cannot_be_empty')}`)
        return true
      }
      const is_param_list_empty = tmpData.param_list.some((element) => {
        return element.name === '' || element.display_name === '' || element.regular === ''
      })
      if (is_param_list_empty) {
        this.$Message.warning(`${this.$t('m_parameter_collection')}: ${this.$t('m_fields_cannot_be_empty')}`)
        return true
      }
      const is_demo_match_value_empty = tmpData.param_list.some((element) => {
        return element.demo_match_value === ''
      })
      if (is_demo_match_value_empty) {
        this.$Message.warning(`${this.$t('m_matching_result')}: ${this.$t('m_cannot_be_empty')}`)
        return true
      }

      const hasDuplicatesParamList = tmpData.param_list.some((element, index) => {
        return tmpData.param_list.findIndex((item) => item.name === element.name) !== index
      })
      if (hasDuplicatesParamList) {
        this.$Message.warning(`${this.$t('m_parameter_key')}${this.$t('m_cannot_be_repeated')}`)
        return true
      }

      const is_metric_list_empty = tmpData.metric_list.some((element) => {
        return element.display_name === '' || element.metric === '' || element.log_param_name === '' || element.agg_type === ''
      })
      if (is_metric_list_empty) {
        this.$Message.warning(`${this.$t('m_compute_metrics')}: ${this.$t('m_fields_cannot_be_empty')}`)
        return true
      }
      const hasDuplicatesMetricList = tmpData.metric_list.some((element, index) => {
        return tmpData.metric_list.findIndex((item) => item.metric === element.metric) !== index
      })
      if (hasDuplicatesMetricList) {
        this.$Message.warning(`${this.$t('m_metric_key')}${this.$t('m_cannot_be_repeated')}`)
        return true
      }
      return false
    },
    saveConfig () {
      let tmpData = JSON.parse(JSON.stringify(this.configInfo))
      if (this.paramsValidate(tmpData)) return
      delete tmpData.create_user
      delete tmpData.create_time
      delete tmpData.update_user
      delete tmpData.update_time
      let methodType = this.isAdd ? 'POST' : 'PUT'
      this.$root.$httpRequestEntrance.httpRequestEntrance(methodType, this.$root.apiCenter.customLogMetricConfig, tmpData, () => {
        this.$Message.success(this.$t('tips.success'))
        this.showModal = false
        this.$emit('reloadMetricData', this.parentGuid)
      })
    },
    getConfig(guid) {
      const api = this.$root.apiCenter.customLogMetricConfig + '/' + guid
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', api, {}, (resp) => {
        this.configInfo = resp
        this.showModal = true
      })
    },
    changeVal (params, index, key , val) {
      this.configInfo[params][index][key] = val
    },
    generateBackstageTrial () {
      if (this.configInfo.demo_log === '') {
        this.$Message.warning(`${this.$t('m_log_example')}${this.$t('m_cannot_be_empty')}`)
        return
      }
      const hasDuplicatesParamList = this.configInfo.param_list.some((element, index) => {
        return this.configInfo.param_list.findIndex((item) => item.name === element.name) !== index
      })
      if (hasDuplicatesParamList) {
        this.$Message.warning(`${this.$t('m_parameter_key')}${this.$t('m_cannot_be_repeated')}`)
        return true
      }
      const params = {
        demo_log: this.configInfo.demo_log,
        param_list: this.configInfo.param_list
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.standardLogRegexMatch, params, (responseData) => {
        this.configInfo.param_list = responseData
      }, {isNeedloading:false})
    },
    //#region 参数采集
    addParameterCollection () {
      this.configInfo.param_list.push({
        guid: '',
        name: '',
        display_name: '',
        json_key: '',
        regular: '',
        demo_match_value: '',
        string_map: [
          // {
          //   regulative: 1,  //匹配类型： 0 是非正则，1是正则
          //   source_value: '', // 源值
          //   target_value: '', // 映射值
          //   value_type: '', //值类型： success 成功，fail 失败
          // }
        ]
      })
    },
    // 编辑标签映射
    editTagMapping (index) {
      this.editTagMappingIndex = index
      let tagMap = this.configInfo.param_list[index].string_map || []
      this.$refs.tagMapConfigRef.loadPage(tagMap)
    },
    setTagMap (arr) {
      this.configInfo.param_list[this.editTagMappingIndex].string_map = arr
    },
    //#endregion
    //#region 指标计算 
    addComputeMetrics () {
      this.configInfo.metric_list.push({
        log_param_name: '',
        metric: '',
        display_name: '',
        agg_type: '',
        tag_config: []
      })
    },
    //#endregion
    deleteAction (key, index) {
      this.configInfo[key].splice(index, 1)
    }
  },
  components: {
    TagMapConfig
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
.ivu-form-item {
  margin-bottom: 0px;
}
</style>