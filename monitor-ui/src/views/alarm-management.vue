<template>
  <div>
    <Title :title="$t('menu.alert')"></Title>
    <Modal
      v-model="isShowWarning"
      title="Delete confirmation"
      @on-ok="ok"
      @on-cancel="cancel">
      <div class="modal-body" style="padding:30px">
        <div style="text-align:center">
          <p style="color: red">Will you delete it?</p>
        </div>
      </div>
    </Modal>
    <div class="alarm-total">
      <Tag color="primary">Low:{{this.low}}</Tag>
      <Tag color="success">Medium:{{this.mid}}</Tag>
      <Tag color="error">High:{{this.high}}</Tag>
    </div>
    <section style="margin-left:8px" class="c-dark-exclude-color">
      <Tag color="warning">{{$t('title.updateTime')}}：{{timeForDataAchieve}}</Tag>
      <template v-for="(filterItem, filterIndex) in filtersForShow">
        <Tag color="success" type="border" closable @on-close="exclude(filterItem.key)" :key="filterIndex">{{filterItem.key}}：{{filterItem.value}}</Tag>
      </template>
      <template v-if="!resultData.length">
        <Tag color="primary">{{$t('table.noDataTip')}}！</Tag>
      </template>
    </section>
    <div class="alarm-list">
      <template v-for="(alarmItem, alarmIndex) in resultData">
        <section :key="alarmIndex" class="alarm-item c-dark-exclude-color" :class="'alarm-item-border-'+ alarmItem.s_priority">
          <i class="fa fa-times" @click="deleteConfirmModal(alarmItem)" aria-hidden="true"></i>
          <ul>
            <li>
              <label class="col-md-1">{{$t('field.endpoint')}}:</label>
              <Tag type="border" closable @on-close="addParams('endpoint',alarmItem.endpoint)" color="primary">{{alarmItem.endpoint}}</Tag>
            </li>
            <li v-if="!alarmItem.is_custom">
              <label class="col-md-1">{{$t('field.metric')}}:</label>
              <Tag type="border" closable @on-close="addParams('metric',alarmItem.s_metric)" color="primary">{{alarmItem.s_metric}}</Tag>
            </li>
            <li>
              <label class="col-md-1">{{$t('tableKey.s_priority')}}:</label>
              <Tag type="border" closable @on-close="addParams('priority',alarmItem.s_priority)" color="primary">{{alarmItem.s_priority}}</Tag>
            </li>
            <li v-if="!alarmItem.is_custom && alarmItem.tags">
              <label class="col-md-1">Tags:</label>
              <Tag type="border" v-for="(t,tIndex) in alarmItem.tags.split('^')" :key="tIndex" color="cyan">{{t}}</Tag>
            </li>
            <li>
              <label class="col-md-1">{{$t('tableKey.start')}}:</label><span>{{alarmItem.start_string}}</span>
            </li>
            <li v-if="alarmIndex != actveAlarmIndex">
              <label class="col-md-1"></label><span><Icon @click="actveAlarmIndex = alarmIndex" type="ios-arrow-dropdown" size=16 /></span>
            </li>
            <template v-if="alarmIndex === actveAlarmIndex">
              <template v-if="alarmItem.is_log_monitor">
                <li>
                  <label class="col-md-1">{{$t('tableKey.path')}}:</label><span>{{alarmItem.path}}</span>
                </li>
                <li>
                  <label class="col-md-1">{{$t('tableKey.keyword')}}:</label><span>{{alarmItem.keyword}}</span>
                </li>
                <li>
                  <label class="col-md-1">{{$t('tableKey.description')}}:</label><span>{{alarmItem.content}}</span>
                </li>
              </template>
              <template v-else-if="alarmItem.is_custom">
                <li>
                  <label class="col-md-1">Log:</label>
                  <div class="col-md-10" style="display: inline-flex;padding: 0px;font-size: 15px;" v-html="alarmItem.content"></div>
                </li>
              </template>
              <template v-else>
                <li>
                  <label class="col-md-1">{{$t('tableKey.start_value')}}:</label><span>{{alarmItem.start_value}}</span>
                </li>
                <li>
                  <label class="col-md-1">{{$t('field.threshold')}}:</label><span>{{alarmItem.s_cond}}</span>
                </li>
                <li>
                  <label class="col-md-1">{{$t('tableKey.s_last')}}:</label><span>{{alarmItem.s_last}}</span>
                </li>
              </template>
            </template>
          </ul>
        </section>
      </template>
    </div>
  </div>
</template>

<script>
export default {
  name: '',
  data() {
    return {
      isShowWarning: false,
      interval: null,
      timeForDataAchieve: null,
      filters: {},
      filtersForShow: [],
      actveAlarmIndex: null,
      resultData: [],
      selectedData: '', // 存放选中数据

      low: 0,
      mid: 0,
      high: 0
    }
  },
  mounted(){
    this.getAlarm()
    this.interval = setInterval(()=>{
      this.getAlarm()
    }, 10000)
    this.$once('hook:beforeDestroy', () => {
      clearInterval(this.interval)
    })
  },
  methods: {
    getAlarm() {
      let params = {}
      let keys = Object.keys(this.filters)
      this.filtersForShow = []
      for (let i = 0; i< keys.length ;i++) {
        params[keys[i]] = this.filters[keys[i]]
        this.filtersForShow.push({key:keys[i], value:this.filters[keys[i]]})
      }
      
      this.timeForDataAchieve = new Date().toLocaleString()
      this.timeForDataAchieve = this.timeForDataAchieve.replace('上午', 'AM ')
      this.timeForDataAchieve = this.timeForDataAchieve.replace('下午', 'PM ')
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', 'alarm/problem/query', params, (responseData) => {
        this.resultData = responseData.data
        this.low = responseData.low
        this.mid = responseData.mid
        this.high = responseData.high
      })
    },
    addParams (key, value) {
      this.filters[key] = value
      this.getAlarm()
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
    removeConfirm (alarmItem) {
      this.$delConfirm({
        msg: alarmItem.endpoint,
        callback: () => {
          this.removeAlarm(alarmItem)
        }
      })
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
        this.getAlarm()
      })
    },
    exclude (key) {
      delete this.filters[key]
      this.getAlarm()
    }
  },
  components: {},
}
</script>

<style scoped lang="less">
li {
  list-style: none;
} 

label {
  margin-bottom: 0;
  text-align: right;
}
.alarm-total {
  float: right;
  font-size: 18px;
}
.alarm-list {
  height: ~"calc(100vh - 150px)";
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

.fa-times {
  margin: 8px;
  float: right;
  font-size: 16px;
  cursor: pointer;
}

</style>
