<template>
  <Card style="margin-bottom: 8px;">
    <template slot="title">
      <div
        style="padding: 0; color: #404144; font-size: 16px;display:flex;align-items:center;"
      >
        <img
          v-if="data.s_priority === 'high'"
          class="bg"
          src="../assets/img/icon_alarm_H_cube.png"
          style="margin-right: 8px; cursor: pointer"
          @click="addParams('priority', data.s_priority)"
        />
        <img
          v-else-if="data.s_priority === 'medium'"
          class="bg"
          src="../assets/img/icon_alarm_M_cube.png"
          style="margin-right: 8px; cursor: pointer"
          @click="addParams('priority', data.s_priority)"
        />
        <img
          v-else
          class="bg"
          src="../assets/img/icon_alarm_L_cube.png"
          style="margin-right: 8px; cursor: pointer"
          @click="addParams('priority', data.s_priority)"
        />
        <template v-if="data.alarm_name">
          <Tooltip :content=data.alarm_name max-width="300" >
            <div class="custom-title">
              <span class="custom-title-text">{{data.alarm_name}}</span>
              <img
                v-if="!$attrs.hideFilter"
                class="filter-icon-flex"
                @click="addParams('alarm_name', data.alarm_name)"
                src="../assets/img/icon_filter.png"
              />
            </div>
          </Tooltip>
        </template>
        <div v-else>
          <span v-if="data.is_custom" v-html="data.title"></span>
          <span v-else v-html="data.content"></span>
        </div>
      </div>
    </template>
    <div
      slot="extra"
      style="padding: 0; color: #7e8086;"
    >
      {{ data.start_string }}
    </div>
    <div
      v-if="$attrs.button"
      style="position: absolute; top: 10px; right: 10px"
    >
      <Tooltip :content="$t('m_duplicate_alert_object')">
        <Icon
          type="ios-copy-outline"
          size="20"
          class="fa-operate"
          @click="copyEndpoint(data)"
        />
      </Tooltip>
      <Poptip trigger="hover" transfer>
        <div slot="title" style="white-space: normal;color: #2d8cf0">
          <p>{{ $t('m_initiate_orchestration') }}: {{ data.notify_callback_name }}</p>
        </div>
        <div slot="content" style="white-space: normal;padding:12px">
          <p>{{ $t('m_tableKey_description') }}: {{ data.notify_message }}</p>
        </div>
        <img v-if="data.notify_id !== ''" @click="goToNotify(data)" style="vertical-align: super;padding:3px 8px;cursor:pointer" src="../assets/img/icon_start_flow.png" />
      </Poptip>
      <Tooltip :content="$t('m_menu_endpointView')">
        <Icon
          type="ios-stats"
          size="18"
          class="fa-operate"
          v-if="!data.is_custom"
          @click="goToEndpointView(data)"
        />
      </Tooltip>
      <Poptip
        confirm
        :title="$t('m_confirm_close_alarm')"
        placement="left"
        @on-ok="deleteConfirmModal(data, false)"
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
          :color="data.custom_message !== '' ? '#2d8cf0' : ''"
          @click="remarkModal(data)"
        />
      </Tooltip>
    </div>
    <ul>
      <li>
        <label class="card-label" v-html="$t('m_tableKey_content')"></label>
        <div class="card-content">
          <div style="display:flex;align-items:center;width:100%;">
            <div class="ellipsis">
              <Tooltip :content="data.content" :max-width="300" placement="bottom-start">
                <div slot="content">
                  <div v-html="data.content || '-'"></div>
                </div>
                <div v-html="data.content || '-'" class="ellipsis-text" style="width: 450px;"></div>
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
              <Tooltip :content="data.log" :max-width="300" placement="bottom-start">
                <div slot="content">
                  <div v-html="data.log || '-'"></div>
                </div>
                <div v-html="data.log || '-'" class="ellipsis-text" style="width: 300px"></div>
              </Tooltip>
            </div>
          </div>
        </div>
      </li>
      <li v-if="data.system_id">
        <label class="card-label" v-html="$t('m_tableKey_system_id')"></label>
        <div class="card-content">
          {{ data.system_id }}
        </div>
      </li>
      <li>
        <label class="card-label" v-html="$t('m_field_endpoint')"></label>
        <div class="card-content">
          <div style="display:flex;align-items:center;width:100%;">
            <div class="ellipsis">
              <Tooltip :content="data.endpoint" :max-width="300">
                {{ data.endpoint }}
              </Tooltip>
            </div>
            <img
              v-if="!$attrs.hideFilter"
              class="filter-icon-flex"
              @click="addParams('endpoint', data.endpoint + '$*$' + data.endpoint_guid)"
              src="../assets/img/icon_filter.png"
            />
          </div>
        </div>
      </li>
      <li>
        <label class="card-label" v-html="$t('m_field_metric')"></label>
        <div class="card-content" style="display: flex">
          <div class="mr-2" v-for="(metric, index) in data.alarm_metric_list" :key=index>
            {{ metric }}
            <img
              v-if="!$attrs.hideFilter"
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
          <span class="mr-2" v-for="(item, index) in data.strategy_groups" :key=index>
            {{$t(strategyNameMaps[item.type])}}: {{item.name}}
          </span>
        </div>
      </li>

      <li>
        <label class="card-label" v-html="$t('m_tableKey_threshold')"></label>
        <div class="card-content">
          <span v-html="data.alarm_detail"></span>
        </div>
      </li>
      <li v-if="data.is_custom">
        <label class="card-label" v-html="$t('m_content')"></label>
        <div class="card-content" v-html="data.content"></div>
      </li>
    </ul>
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
    >
      <EndpointViewComponent v-if='showEndpointView' ref="endpointViewComponentRef"></EndpointViewComponent>
    </Modal>
  </Card>
</template>

<script>
import Vue from 'vue'
import EndpointViewComponent from '@/components/endpoint-view-component'
export default {
  props: {
    data: Object,
  },
  data() {
    return {
      isShowStartFlow: false,
      startFlowTip: '',
      alertId: '',
      test: 'system_id:5006 <br/> title:bdphdp010001: JournalNode10分钟之内ops次数大于10000 <br/> object: <br/> info:bdphdp010001在2022.05.16-00:14:14触发JournalNode10分钟之内ops次数大于10000 <br/> 【告警主机】 127.0.0.1[bdphdp010001] <br/> 【告警集群】 international_cluster <br/> 【附加信息】 请联系值班人:[admin]，资源池[admin]',
      strategyNameMaps: {
        endpointGroup: 'm_base_group',
        serviceGroup: 'm_field_resourceLevel'
      },
      showEndpointView: false // 弹窗展示对象视图
    }
  },
  watch: {
    showEndpointView(val) {
      if (val) {
        setTimeout(() => {
          this.$refs.endpointViewComponentRef.disabledEndpoint(val)
        }, 100)
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
      }, 100)
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
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST',this.$root.apiCenter.startNotify, params, () => {
        this.$Message.success(this.$t('m_tips_success'))
      },{isNeedloading: false})
    },
    deleteConfirmModal(rowData, isBatch) {
      this.$parent.isBatch = isBatch
      this.$parent.removeAlarm(rowData)
    },
    remarkModal(item) {
      this.$emit('openRemarkModal', item)
      // this.$parent.modelConfig.addRow = {
      //   id: item.id,
      //   message: item.custom_message,
      //   is_custom: false,
      // };
      // this.$root.JQ("#remark_Modal").modal("show");
    },
    addParams(key, value) {
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
    }
  },
  components: {
    EndpointViewComponent
  }
}
</script>
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
  float: right;
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
    display: inline-block;
    max-width: ~"calc(40vw - 285px)";
    white-space: nowrap;
    overflow: hidden;
    text-overflow: ellipsis;

  }
}
</style>
