<template>
  <Card style="margin-bottom: 8px;" class="xxx">
    <template #title>
      <div
        class="col-md-9"
        style="padding: 0; color: #404144; font-size: 16px;display:flex;align-items:center;"
      >
        <img
          v-if="data.s_priority == 'high'"
          class="bg"
          src="../assets/img/icon_alarm_H_cube.png"
          style="margin-right: 8px; cursor: pointer"
          @click="addParams('priority', data.s_priority)"
        />
        <img
          v-else-if="data.s_priority == 'medium'"
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
        <div v-if="data.alarm_name">{{data.alarm_name}}
          <img
              class="filter-icon-flex"
              @click="addParams('alarm_name', data.alarm_name)"
              src="../assets/img/icon_filter.png"
            />
        </div>
        <div v-else>
          <span v-if="data.is_custom" v-html="data.title"></span>
          <span v-else v-html="data.content"></span>
        </div>
      </div>
      <div
        style="padding: 0; text-align: right; color: #7e8086;width:200px"
      >
        {{ data.start_string }}
      </div>
    </template>
    <div
      v-if="$attrs.button"
      style="position: absolute; top: 10px; right: 10px"
    >
      <Tooltip :content="$t('menu.endpointView')">
        <Icon
          type="ios-copy-outline"
          size="20"
          class="fa-operate"
          @click="copyEndpoint(data)"
        />
      </Tooltip>
      <Poptip trigger="hover">
        <div slot="title" style="white-space: normal;color: #2d8cf0">
          <p>{{ $t('m_initiate_orchestration') }}: {{ data.notify_callback_name }}</p>
        </div>
        <div slot="content" style="white-space: normal;padding:16px">
          <p>{{ $t('tableKey.description') }}: {{ data.notify_message }}</p>
        </div>
        <img v-if="data.notify_id !==''" @click="goToNotify(data)" style="vertical-align: super;padding:3px 8px;cursor:pointer" src="../assets/img/icon_start_flow.png" />
      </Poptip>
      <Tooltip :content="$t('menu.endpointView')">
        <Icon
          type="ios-stats"
          size="18"
          class="fa-operate"
          v-if="!data.is_custom"
          @click="goToEndpointView(data)"
        />
      </Tooltip>
      <Tooltip :content="$t('close')">
        <Icon
          type="ios-eye-off"
          size="18"
          class="fa-operate"
          @click="deleteConfirmModal(data, false)"
        />
      </Tooltip>
      <Tooltip :content="$t('m_remark')">
        <Icon
          type="ios-pricetags-outline"
          size="18"
          class="fa-operate"
          @click="remarkModal(data)"
        />
      </Tooltip>
    </div>
    <ul>
      <li>
        <label class="card-label" v-html="$t('tableKey.content')"></label>
        <div class="card-content">
          <div style="display:flex;align-items:center;width:100%;">
            <div class="ellipsis">
              <Tooltip :content="data.content">
                {{ data.content || '-' }}
              </Tooltip>
            </div>
          </div>
        </div>
      </li>
      <li v-if="data.system_id">
        <label class="card-label" v-html="$t('tableKey.system_id')"></label>
        <div class="card-content">
          {{ data.system_id }}
        </div>
      </li>
      <li>
        <label class="card-label" v-html="$t('field.endpoint')"></label>
        <div class="card-content">
          <div style="display:flex;align-items:center;width:100%;">
            <div class="ellipsis">
              <Tooltip :content="data.endpoint">
                {{ data.endpoint }}
              </Tooltip>
            </div>
            <img
              class="filter-icon-flex"
              @click="addParams('endpoint', data.endpoint)"
              src="../assets/img/icon_filter.png"
            />
          </div>
        </div>
      </li>
      <li>
        <label class="card-label" v-html="$t('field.metric')"></label>
        <div class="card-content" style="display: flex">
           <div class="mr-2" v-for="(metric, index) in data.alarm_metric_list" :key=index>
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
          <span class="mr-2" v-for="(item, index) in data.strategy_groups" :key=index>
            {{$t(strategyNameMaps[item.type])}}: {{item.name}}
          </span>
        </div>
      </li>

      <li>
        <label class="card-label" v-html="$t('tableKey.threshold')"></label>
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
      @on-cancel="isShowStartFlow = false">
      <div class="modal-body" style="padding:30px">
        <div style="text-align:center">
          <p style="color: red;text-align: left;">{{startFlowTip}}</p>
        </div>
      </div>
    </Modal>
  </Card>
</template>

<script>
export default {
  props: {
    data: Object,
  },
  data() {
    return {
      isShowStartFlow: false,
      startFlowTip: '',
      alertId: '',
      test: "system_id:5006 <br/> title:bdphdp010001: JournalNode10分钟之内ops次数大于10000 <br/> object: <br/> info:bdphdp010001在2022.05.16-00:14:14触发JournalNode10分钟之内ops次数大于10000 <br/> 【告警主机】 ***REMOVED***[bdphdp010001] <br/> 【告警集群】 international_cluster <br/> 【附加信息】 请联系值班人:[admin]，资源池[admin]",
      strategyNameMaps: {
        "endpointGroup": "m_base_group",
        "serviceGroup": "field.resourceLevel"
      }
    }
  },
  methods: {
    goToEndpointView(alarmItem) {
      const endpointObject = {
        option_value: alarmItem.endpoint,
        type: alarmItem.endpoint.split("_").slice(-1)[0],
      };
      localStorage.setItem("jumpCallData", JSON.stringify(endpointObject));
      this.$router.push({ path: "/endpointView" });
      // const news = this.$router.resolve({name: 'endpointView'})
      // window.open(news.href, '_blank')
    },
    goToNotify (item) {
      if (item.notify_status === 'notStart') {
        this.startFlowTip = `${this.$t('button.confirm')} ${this.$t('m_initiate_orchestration')}: [${item.notify_callback_name}]`
      } else if (item.notify_status === 'started') {
        this.startFlowTip = `${this.$t('m_already_initiated')}，${this.$t('button.confirm')} ${this.$t('m_reinitiate_orchestration')}: 【${item.notify_callback_name}】`
      }
      this.alertId = item.id
      this.isShowStartFlow = true
    },
    confirmStartFlow () {
      let params = {
        id: this.alertId
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST',this.$root.apiCenter.startNotify, params, () => {
        this.$Message.success(this.$t('tips.success'))
      },{isNeedloading: false})
    },
    deleteConfirmModal(rowData, isBatch) {
      this.$parent.isBatch = isBatch;
      this.$parent.selectedData = rowData;
      this.$parent.isShowWarning = true;
    },
    remarkModal(item) {
      this.$parent.modelConfig.addRow = {
        id: item.id,
        message: item.custom_message,
        is_custom: false,
      };
      this.$root.JQ("#remark_Modal").modal("show");
    },
    addParams(key, value) {
      this.$parent.filters[key] = value;
      this.$parent.getAlarm();
    },
    copyEndpoint (data) {
      let inputElement = document.createElement('input')
      inputElement.value = data.alarm_obj_name
      document.body.appendChild(inputElement)
      inputElement.select()
      document.execCommand('Copy')
      inputElement.remove()
      this.$Message.success(this.$t('m_copied_to_clipboard'))
    }
  },
};
</script>
<style lang="less">
.ivu-card-head {
  padding: 8px 16px !important;
}
</style>
<style scoped lang="less">
/deep/ .ivu-card-head {
  background: #f2f3f7;
  display: flex;
  align-items: center;
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
  margin-left: 6px;
  cursor: pointer;
}

.fa-operate {
  margin: 8px;
  float: right;
  font-size: 16px;
  cursor: pointer;
}
.ellipsis {
  max-width: ~"calc(100% - 120px)";
  text-overflow: ellipsis;
  overflow: hidden;
  white-space: nowrap;
}
.copy-data {
  font-size: 16px;
  cursor: pointer
}
</style>
