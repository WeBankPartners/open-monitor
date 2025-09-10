<template>
  <div class="alarm-card-collapse">
    <Collapse v-model="expandCollapse">
      <Panel v-for="item in collapseData" :key="item.id" :name="item.id + ''">
        <template>
          <img
            v-if="item.s_priority === 'high'"
            class="bg"
            src="../assets/img/icon_alarm_H_cube.png"
            style="margin-right: 8px; cursor: pointer"
            @click="e => addParams('priority', item.s_priority, e)"
          />
          <img
            v-else-if="item.s_priority === 'medium'"
            class="bg"
            src="../assets/img/icon_alarm_M_cube.png"
            style="margin-right: 8px; cursor: pointer"
            @click="e => addParams('priority', item.s_priority, e)"
          />
          <img
            v-else
            class="bg"
            src="../assets/img/icon_alarm_L_cube.png"
            style="margin-right: 8px; cursor: pointer"
            @click="e => addParams('priority', item.s_priority, e)"
          />
          <template v-if="item.alarm_name">
            <Tooltip :content="item.alarm_name" max-width="600" >
              <div class="custom-title">
                <span class="custom-title-text">{{item.alarm_name}}</span>
                <img
                  class="filter-icon-flex"
                  @click="e => addParams('alarm_name', item.alarm_name, e)"
                  src="../assets/img/icon_filter.png"
                />
              </div>
            </Tooltip>
          </template>
          <div v-else>
            <span v-if="item.is_custom" v-html="item.title"></span>
            <span v-else v-html="item.content"></span>
          </div>
          <span class="start-time">{{$t('m_first_time') + ': ' + item.start_string}}</span>
        </template>
        <template slot="content">
          <div class="collapse-content">
            <div v-if="isCanAction" class="right-action-button">
              <Tooltip :content="$t('m_duplicate_alert_object')">
                <Icon
                  type="ios-copy-outline"
                  size="20"
                  class="fa-operate"
                  @click="copyEndpoint(item)"
                />
              </Tooltip>
              <Poptip trigger="hover" transfer>
                <div slot="title" style="white-space: normal;color: #5384FF">
                  <p>{{ $t('m_initiate_orchestration') }}: {{ item.notify_callback_name }}</p>
                </div>
                <div slot="content" style="white-space: normal;padding:12px">
                  <p>{{ $t('m_tableKey_description') }}: {{ item.notify_message }}</p>
                </div>
                <img v-if="item.notify_id !== ''"
                     @click="goToNotify(item)"
                     style="vertical-align: super;padding:0px 8px;cursor:pointer; margin-bottom: -7px"
                     src="../assets/img/icon_start_flow.png"
                />
              </Poptip>
              <Tooltip :content="$t('m_menu_endpointView')">
                <Icon
                  type="ios-stats"
                  size="18"
                  class="fa-operate"
                  v-if="!item.is_custom"
                  @click="goToEndpointView(item)"
                />
              </Tooltip>
              <Poptip
                confirm
                :title="$t('m_confirm_close_alarm')"
                placement="left"
                @on-ok="deleteConfirmModal(item, false)"
              >
                <Tooltip :content="$t('m_alarm_close')">
                  <Icon
                    type="ios-eye-off"
                    size="18"
                    class="fa-operate"
                  />
                </Tooltip>
              </Poptip>
              <Tooltip :content="$t('m_remark')">
                <Icon
                  type="ios-pricetags"
                  size="18"
                  class="fa-operate"
                  :color="item.custom_message !== '' ? '#5384FF' : ''"
                  @click="remarkModal(item)"
                />
              </Tooltip>
            </div>
            <ul>
              <li>
                <label class="card-label" v-html="$t('m_notify_content')"></label>
                <div class="card-content">
                  <div style="display:flex;align-items:center;width:100%;">
                    <div class="ellipsis">
                      <Tooltip :content="item.content" :max-width="300" placement="bottom-start">
                        <div slot="content">
                          <div v-html="item.content || '-'"></div>
                        </div>
                        <div v-html="item.content || '-'" class="ellipsis-text" style="max-width: 40vw"></div>
                      </Tooltip>
                    </div>
                  </div>
                </div>
              </li>
              <li>
                <label class="card-label" v-html="$t('m_log')"></label>
                <div class="card-content">
                  <div style="display:flex;align-items:center;width:100%;">
                    <div class="ellipsis">
                      <Tooltip :content="item.log" :max-width="300" placement="bottom-start">
                        <div slot="content">
                          <div v-html="item.log || '-'"></div>
                        </div>
                        <div v-html="item.log || '-'" class="ellipsis-text" style="max-width: 40vw"></div>
                      </Tooltip>
                    </div>
                  </div>
                </div>
              </li>
              <li v-if="item.system_id">
                <label class="card-label" v-html="$t('m_tableKey_system_id')"></label>
                <div class="card-content">
                  {{ item.system_id }}
                </div>
              </li>
              <li>
                <label class="card-label" v-html="$t('m_field_endpoint')"></label>
                <div class="card-content">
                  <div style="display:flex;align-items:center;width:100%;">
                    <div class="ellipsis">
                      <Tooltip :content="item.endpoint" :max-width="300">
                        {{ item.endpoint }}
                      </Tooltip>
                    </div>
                    <img
                      class="filter-icon-flex"
                      @click="addParams('endpoint', item.endpoint + '$*$' + item.endpoint_guid)"
                      src="../assets/img/icon_filter.png"
                    />
                  </div>
                </div>
              </li>
              <li>
                <label class="card-label" v-html="$t('m_field_metric')"></label>
                <div class="card-content" style="display: flex">
                  <div class="mr-2" v-for="(metric, index) in item.alarm_metric_list" :key=index>
                    {{ metric }}
                    <img
                      class="filter-icon"
                      @click="addParams('metric', metric)"
                      src="../assets/img/icon_filter.png"
                    />
                  </div>
                </div>
              </li>
              <li>
                <label class="card-label" v-html="$t('m_configuration')"></label>
                <div class="card-content">
                  <span class="mr-2" v-for="(item, index) in item.strategy_groups" :key=index>
                    {{$t(strategyNameMaps[item.type])}}: {{item.name}}
                  </span>
                </div>
              </li>

              <li>
                <label class="card-label" v-html="$t('m_tableKey_threshold')"></label>
                <div class="card-content">
                  <span v-html="item.alarm_detail"></span>
                </div>
              </li>
              <li v-if="item.is_custom">
                <label class="card-label" v-html="$t('m_content')"></label>
                <div class="card-content" v-html="item.content"></div>
              </li>
              <li v-if="item.is_custom">
                <label class="card-label" v-html="$t('m_frequency')"></label>
                <div class="card-content" v-html="item.alarm_total"></div>
              </li>
              <li>
                <label class="card-label" v-html="$t('m_update')"></label>
                <div class="card-content" v-html="item.update_string"></div>
              </li>
              <li>
                <label class="card-label" v-html="$t('m_duration')"></label>
                <div class="card-content" v-html="formatDuration(item.duration_sec)"></div>
              </li>
            </ul>
          </div>
        </template>
      </Panel>
    </Collapse>
    <Modal
      v-model="isShowStartFlow"
      :title="$t('m_initiate_orchestration')"
      @on-ok="confirmStartFlow"
      @on-cancel="isShowStartFlow = false"
    >
      <div class="modal-body" style="padding:30px">
        <div style="text-align:center">
          <p style="color: red;text-align: left;">{{startFlowTip}}</p>
        </div>
      </div>
    </Modal>
    <Modal
      v-model="showEndpointView"
      :mask-closable="false"
      :footer-hide="true"
      :fullscreen="true"
      :title="$t('m_menu_endpointView')"
      :mask="false"
      class-name="endpoint-view-modal"
    >
      <EndpointViewComponent v-if='showEndpointView' ref="endpointViewComponentRef"></EndpointViewComponent>
    </Modal>
  </div>
</template>

<script>
import Vue from 'vue'
import EndpointViewComponent from '@/components/endpoint-view-component'

export default {
  props: {
    collapseData: Array,
    isCanAction: {
      type: Boolean,
      default: true
    },
    isCollapseExpandAll: {
      type: Boolean,
      default: false
    }
  },
  data() {
    return {
      expandCollapse: [],
      isShowStartFlow: false,
      startFlowTip: '',
      alertId: '',
      test: 'system_id:5006 <br/> title:bdphdp010001: JournalNode10分钟之内ops次数大于10000 <br/> object: <br/> info:bdphdp010001在2022.05.16-00:14:14触发JournalNode10分钟之内ops次数大于10000 <br/> 【告警主机】 127.0.0.1[bdphdp010001] <br/> 【告警集群】 international_cluster <br/> 【附加信息】 请联系值班人:[admin]，资源池[admin]',
      strategyNameMaps: {
        endpointGroup: 'm_base_group',
        serviceGroup: 'm_field_resourceLevel'
      },
      showEndpointView: false, // 弹窗展示对象视图
      request: this.$root.$httpRequestEntrance.httpRequestEntrance,
      apiCenter: this.$root.apiCenter,
      imageMap: {
        low: 'icon_alarm_L_cube.png',
        medium: 'icon_alarm_M_cube.png',
        high: 'icon_alarm_H_cube.png'
      }
    }
  },
  watch: {
    showEndpointView(val) {
      if (val) {
        setTimeout(() => {
          this.$refs.endpointViewComponentRef.disabledEndpoint(val)
        }, 100)
      }
    },
    isCollapseExpandAll(val) {
      this.expandCollapse = []
      if (val) {
        this.collapseData.forEach(item => {
          this.expandCollapse.push(item.id + '')
        })
      }
    }
  },
  methods: {
    goToEndpointView(alarmItem) {
      const endpointObject = {
        option_name: alarmItem.endpoint,
        option_value: alarmItem.endpoint_guid,
        type: alarmItem.endpoint_guid.split('_').slice(-1)[0],
      }
      this.showEndpointView = true
      setTimeout(() => {
        this.$refs.endpointViewComponentRef.refreshConfig(endpointObject)
      }, 200)
    },
    goToNotify(item) {
      if (item.notify_permission === 'no' || !item.notify_permission) {
        return this.$Message.error(this.$t('m_noProcessPermission'))
      } else if (item.notify_status === 'notStart') {
        this.startFlowTip = `${this.$t('m_button_confirm')} ${this.$t('m_initiate_orchestration')}: [${item.notify_callback_name}]`
      } else if (item.notify_status === 'started') {
        this.startFlowTip = `${this.$t('m_already_initiated')}，${this.$t('m_button_confirm')} ${this.$t('m_reinitiate_orchestration')}: 【${item.notify_callback_name}】`
      }
      this.alertId = item.id
      this.isShowStartFlow = true
    },
    confirmStartFlow() {
      const params = {
        id: this.alertId
      }
      this.request('POST',this.apiCenter.startNotify, params, () => {
        this.$Message.success(this.$t('m_tips_success'))
      },{isNeedloading: false})
    },
    deleteConfirmModal(rowData, isBatch) {
      this.$parent.isBatch = isBatch
      this.$parent.removeAlarm(rowData)
    },
    remarkModal(item) {
      this.$emit('openRemarkModal', item)
    },
    addParams(key, value, e) {
      if (e) {
        e.stopPropagation()
      }
      Vue.set(this.$parent.filters, key, this.$parent.filters[key] || [])
      const singleArr = this.$parent.filters[key]
      if (singleArr.includes(value)) {
        singleArr.splice(singleArr.indexOf(value), 1)
      } else {
        singleArr.push(value)
      }
    },
    copyEndpoint(data) {
      const inputElement = document.createElement('input')
      inputElement.value = data.alarm_obj_name
      document.body.appendChild(inputElement)
      inputElement.select()
      document.execCommand('Copy')
      inputElement.remove()
      this.$Message.success(this.$t('m_copied_to_clipboard'))
    },
    formatDuration(seconds) {
      const units = [
        {
          value: Math.floor(seconds / 86400),
          label: '日'
        }, // 天数
        {
          value: Math.floor((seconds % 86400) / 3600),
          label: '时'
        }, // 小时数
        {
          value: Math.floor((seconds % 3600) / 60),
          label: '分'
        }, // 分钟数
        {
          value: seconds % 60,
          label: '秒'
        } // 秒数
      ]
      const result = []
      let hasHigherUnit = false
      for (const unit of units) {
        if (unit.value > 0 || hasHigherUnit) {
          result.push(`${unit.value}${unit.label}`)
          hasHigherUnit = true // 标记存在更大单位，后续单位需保留
        }
      }
      return result.length > 0 ? result.join('') : '0秒'
    },
    closeAllCollapse() {
      this.expandCollapse = []
    },
    expandAllCollapse() {
      this.$nextTick(() => {
        this.expandCollapse = []
        this.collapseData.forEach(item => {
          this.expandCollapse.push(item.id + '')
        })
      })
    }
  },
  components: {
    EndpointViewComponent
  }
}
</script>
<style lang="less">
.alarm-card-collapse {
  .ivu-collapse {
    border: 0px;
  }
  .ivu-collapse-content:last-of-type {
    border-bottom: 1px solid #CFD0D3;
    border-radius: 0px;
  }
  .ivu-collapse-header {
    background-color: #F2F3F7;
  }
  .start-time {
    color: rgb(126,128,134);
    position: absolute;
    right: 16px;
    top: 0px;
    font-size: 14px;
  }
  .collapse-content {
    position: relative;
    .right-action-button {
      position: absolute;
      top: -5px;
      right: 0px
    }
  }
}

.endpoint-view-modal {
  z-index: 1000!important;
}
</style>
<style scoped lang="less">
/deep/ .ivu-card-head {
  background: #f2f3f7;
  display: flex;
  align-items: center;
  padding: 8px 16px !important;
}

/deep/ .ivu-card-body {
  position: relative;
  line-height: 28px;
}

li {
  display: flex;
}
.card-label {
  width: 80px;
  color: #7e8086;
  font-size: 14px;
  text-align: left;
}
.card-content {
  width: ~"calc(100% - 80px)";
  color: #404144;
  overflow: hidden;
  font-size: 14px;
  .tags-box {
    // width: max-content;
    width: 100%;
    display: flex;
    flex-wrap: wrap;
  }
}
.filter-icon {
  position: relative;
  top: -2px;
  margin-left: 6px;
  cursor: pointer;
}

.filter-icon-flex {
  margin-left: 15px;
  cursor: pointer;
}

.fa-operate {
  margin: 8px;
  font-size: 16px;
  cursor: pointer;
}
.ellipsis-text {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
.ellipsis {
  &:extend(.ellipsis-text);
  max-width: ~"calc(100% - 120px)";
}
.copy-data {
  font-size: 16px;
  cursor: pointer
}
.custom-title {
  display: flex;
  align-items: center;
  .custom-title-text {
    color: rgb(64,65,68);
    max-width: 550px;
    // max-width: 440px;
    font-size: 16px;
    display: inline-block;
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;
  }
}
</style>
