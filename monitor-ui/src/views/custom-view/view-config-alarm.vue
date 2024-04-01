<template>
  <div class=" ">
    <div class="alarm-total">
      <Tag color="success"><span style="font-size:14px">{{$t('m_low')}}:{{this.low}}</span></Tag>
      <Tag color="warning"><span style="font-size:14px">{{$t('m_medium')}}:{{this.mid}}</span></Tag>
      <Tag color="error"><span style="font-size:14px">{{$t('m_high')}}:{{this.high}}</span></Tag>
    </div>
    <div class="alarm-list">
      <template v-for="(alarmItem, alarmIndex) in resultData">
        <section :key="alarmIndex" class="alarm-item c-dark-exclude-color" :class="'alarm-item-border-'+ alarmItem.s_priority">
          <div style="float:right">
            <Poptip trigger="hover">
              <div slot="title" style="white-space: normal;color: #2d8cf0">
                <p>{{ $t('m_initiate_orchestration') }}: {{ data.notify_callback_name }}</p>
              </div>
              <div slot="content" style="white-space: normal;padding:16px">
                <p>{{ $t('tableKey.description') }}: {{ data.notify_message }}</p>
              </div>
              <Icon
                type="ios-megaphone"
                size="18"
                class="fa-operate"
                v-if="data.notify_id !==''"
                @click="goToNotify(data)"
              />
            </Poptip>
            <Tooltip :content="$t('menu.endpointView')">
              <Icon type="ios-stats" size="18" class="fa-operate" v-if="!alarmItem.is_custom" @click="goToEndpointView(alarmItem)"/>
            </Tooltip>
            <Tooltip :content="$t('close')">
              <Icon type="ios-eye-off" size="18" class="fa-operate" v-if="permission === 'edit'" @click="deleteConfirmModal(alarmItem)"/>
            </Tooltip>
            <Tooltip :content="$t('m_remark')">
              <Icon type="ios-pricetags-outline" size="18" class="fa-operate" slot="" v-if="permission === 'edit'" @click="remarkModal(alarmItem)" />
            </Tooltip>
          </div>
          <ul>
            <li>
              <label class="alarm-item-label">{{$t('field.endpoint')}}:</label>
              <Tag type="border" color="primary">{{alarmItem.endpoint}}</Tag>
            </li>
            <li v-if="!alarmItem.is_custom">
              <label class="alarm-item-label">{{$t('field.metric')}}:</label>
              <Tag type="border" color="primary">{{alarmItem.s_metric}}</Tag>
            </li>
            <li>
              <label class="alarm-item-label">{{$t('tableKey.s_priority')}}:</label>
              <Tag type="border" color="primary">{{alarmItem.s_priority}}</Tag>
            </li>
            <li v-if="!alarmItem.is_custom && alarmItem.tags">
              <label class="alarm-item-label">{{$t('tableKey.tags')}}:</label>
              <Tag type="border" v-for="(t,tIndex) in alarmItem.tags.split('^')" :key="tIndex" color="cyan">{{t}}</Tag>
            </li>
            <li v-if="alarmItem.custom_message">
              <label class="alarm-item-label">{{$t('m_remark')}}:</label>
              <Tooltip max-width="300">
                <div style="border: 1px solid #2d8cf0;padding:2px;border-radius:4px; color: #2d8cf0">
                {{alarmItem.custom_message.length > 60 ? alarmItem.custom_message.substring(0,60) + '...' : alarmItem.custom_message}}
                </div>
                <div slot="content" style="white-space: normal;">
                  <p>{{alarmItem.custom_message}}</p>
                </div>
              </Tooltip>
            </li>
            <li>
              <label class="alarm-item-label">{{$t('tableKey.start')}}:</label><span>{{alarmItem.start_string}}</span>
            </li>
            <li>
              <label class="alarm-item-label">{{$t('details')}}:</label>
              <span>
                <Tag color="default">{{$t('tableKey.start_value')}}:{{alarmItem.start_value}}</Tag>
                <Tag color="default" v-if="alarmItem.s_cond">{{$t('tableKey.threshold')}}:{{alarmItem.s_cond}}</Tag>
                <Tag color="default" v-if="alarmItem.s_last">{{$t('tableKey.s_last')}}:{{alarmItem.s_last}}</Tag>
                <Tag color="default" v-if="alarmItem.path">{{$t('tableKey.path')}}:{{alarmItem.path}}</Tag>
                <Tag color="default" v-if="alarmItem.keyword">{{$t('tableKey.keyword')}}:{{alarmItem.keyword}}</Tag>
              </span>
            </li>
            <li>
              <label class="alarm-item-label" style="vertical-align: top;">{{$t('alarmContent')}}:</label>
              <div class="col-md-9" style="display: inline-block;padding:0" v-html="alarmItem.content"></div>
            </li>
          </ul>
        </section>
      </template>
    </div>
    <Modal
      v-model="isShowWarning"
      :title="$t('closeConfirm.title')"
      @on-ok="ok"
      @on-cancel="cancel">
      <div class="modal-body" style="padding:30px">
        <div style="text-align:center">
          <p style="color: red">{{$t('closeConfirm.tip')}}</p>
        </div>
      </div>
    </Modal>
    <ModalComponent :modelConfig="modelConfig"></ModalComponent>
  </div>
</template>

<script>
export default {
  name: '',
  data() {
    return {
      cutsomViewId: null,
      interval: '',

      resultData: [],
      selectedData: '', // 存放选中数据
      isShowWarning: false,
      low: 0,
      mid: 0,
      high: 0,
      modelConfig: {
        modalId: 'remark_Modal',
        modalTitle: 'm_remark',
        saveFunc: 'remarkAlarm',
        isAdd: true,
        config: [
          {label: 'm_remark', value: 'message', placeholder: '', v_validate: '', disabled: false, type: 'textarea'}
        ],
        addRow: { // [通用]-保存用户新增、编辑时数据
          id: '',
          message: '',
          is_custom: false
        }
      },
      cacheParams: {
        id: '',
        viewCondition: ''
      }
    }
  },
  mounted () {
    window.addEventListener("visibilitychange", this.isTabActive, true)
  },
  destroyed() {
    this.clearAlarmInterval()
    window.removeEventListener("visibilitychange", this.isTabActive, true)
  },
  methods: {
    isTabActive () {
       if (document.hidden) {
        this.clearAlarmInterval()
      } else {
        if (this.cacheParams.id) {
          this.getAlarm(this.cacheParams.id, this.cacheParams.viewCondition)
        }
      }
    },
    clearAlarmInterval () {
      clearInterval(this.interval)
    },
    getAlarm (id, viewCondition, permission) {
      this.permission = permission
      this.cacheParams.id = id
      this.cacheParams.viewCondition = viewCondition
      this.getAlarmdata(id)
      this.interval = setInterval(()=>{
        this.getAlarmdata(id)
      }, (viewCondition.autoRefresh || 10) * 1000)

    },
    getAlarmdata (id) {
      const parmas = {id}
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', '/monitor/api/v1/dashboard/custom/alarm/list', parmas, (responseData) => {
        this.resultData = responseData.data
        this.low = responseData.low
        this.mid = responseData.mid
        this.high = responseData.high
      }, {isNeedloading: false})
    },
    goToEndpointView (alarmItem) {
      const endpointObject = {
        option_value: alarmItem.endpoint,
        type: alarmItem.endpoint.split('_').slice(-1)[0]
      }
      localStorage.setItem('jumpCallData', JSON.stringify(endpointObject))
      this.$router.push({path: '/endpointView'})
      // const news = this.$router.resolve({name: 'endpointView'})
      // window.open(news.href, '_blank')
    },
    goToNotify (item) {
      let params = {
        id: item.id
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST',this.$root.apiCenter.startNotify, params, () => {
        this.$Message.success(this.$t('tips.success'))
      },{isNeedloading: false})
    },
    deleteConfirmModal (rowData) {
      this.selectedData = rowData
      this.isShowWarning = true
    },
    ok () {
      this.removeAlarm(this.selectedData)
    },
    cancel () {
      this.isShowWarning = false
    },
    removeAlarm(alarmItem) {
      let params = {
        id: alarmItem.id,
        custom: true
      }
      if (!alarmItem.is_custom) {
        params.custom = false
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.alarmManagement.close.api, params, () => {
        // this.$root.$eventBus.$emit('hideConfirmModal')
        this.getAlarm(this.cacheParams.id, this.cacheParams.viewCondition)
      })
    },
    remarkModal (item) {
      this.modelConfig.addRow = {
        id: item.id,
        message: item.custom_message,
        is_custom: false
      }
      this.$root.JQ('#remark_Modal').modal('show')
    },
    remarkAlarm () {
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.apiCenter.remarkAlarm, this.modelConfig.addRow, () => {
        this.$Message.success(this.$t('tips.success'))
        this.getAlarm(this.cacheParams.id, this.cacheParams.viewCondition)
        this.$root.JQ('#remark_Modal').modal('hide')
      })
    }
  },
  components: {},
}
</script>

<style scoped lang="less">
.alarm-item-label {
  width: 70px;
}
.alarm-empty {
  height: ~"calc(100vh - 180px)";
  width: ~"calc(100vw * 0.4)";
  text-align: center;
  padding:50px;
  color: #2d8cf0;
}
.flex-container {
  display: flex;
}
li {
  list-style: none;
} 

label {
  margin-bottom: 0;
  text-align: right;
}
.alarm-total {
  // float: right;
  margin-left: 8px;
  font-size: 18px;
}
.alarm-list {
  height: ~"calc(100vh - 250px)";
  width: 100%;
  overflow-y: auto;
}
.alarm-item {
  border: 1px solid @gray-d;
  margin: 8px;
  padding: 4px;
  border-radius: 4px;
  li {
    padding: 2px;
  }
}
.alarm-item-border-high {
  // border: 1px solid @color-orange-F;
  color: @color-orange-F;
}
.alarm-item-border-medium {
  // border: 1px solid @blue-2;
  color: @blue-2;
}
// .alarm-item-border-low {
//   // border: 1px solid @gray-d;
// }

.alarm-item:hover {
  box-shadow: 0 0 12px @gray-d;
}

.alarm-item /deep/.ivu-icon-ios-close:before {
  content: "\F102";
}

.fa-operate {
  margin: 8px;
  float: right;
  font-size: 16px;
  cursor: pointer;
}

/* 可以设置不同的进入和离开动画 */
/* 设置持续时间和动画函数 */
.slide-fade-enter-active {
  transition: all .3s ease;
}
.slide-fade-leave-active {
  transition: all .8s cubic-bezier(1.0, 0.5, 0.8, 1.0);
}
.slide-fade-enter, .slide-fade-leave-to
/* .slide-fade-leave-active for below version 2.1.8 */ {
  transform: translateX(10px);
  opacity: 0;
}
</style>
