<template>
  <div class=" " style="padding-top:20px;">
    <section style="margin-left:8px">
      <Tag color="warning">数据获取时间：{{timeForDataAchieve}}</Tag>
      <template v-for="(filterItem, filterIndex) in filtersForShow">
        <Tag color="success" type="border" closable @on-close="exclude(filterItem.key)" :key="filterIndex">{{filterItem.key}}：{{filterItem.value}}</Tag>
      </template>
    </section>
    <template v-for="(alarmItem, alarmIndex) in resultData">
      <section :key="alarmIndex" class="alarm-item" :class="'alarm-item-border-'+ alarmItem.s_priority">
        <ul>
           <li>
              <label class="col-md-1">对象:</label>
              <Tag type="border" closable @on-close="addParams('endpoint',alarmItem.endpoint)" color="primary">{{alarmItem.endpoint}}</Tag>
            </li>
            <li>
              <label class="col-md-1">指标:</label>
              <Tag type="border" closable @on-close="addParams('metric',alarmItem.s_metric)" color="primary">{{alarmItem.s_metric}}</Tag>
            </li>
            <li>
              <label class="col-md-1">级别:</label>
              <Tag type="border" closable @on-close="addParams('priority',alarmItem.s_priority)" color="primary">{{alarmItem.s_priority}}</Tag>
            </li>
            <li>
              <label class="col-md-1">开始时间:</label><span>{{alarmItem.start}}</span>
            </li>
            <li v-if="alarmIndex != actveAlarmIndex">
              <label class="col-md-1"></label><span><Icon @click="actveAlarmIndex = alarmIndex" type="ios-arrow-dropdown" size=16 /></span>
            </li>
          <template v-if="alarmIndex === actveAlarmIndex">
            <li>
              <label class="col-md-1">告警值:</label><span>{{alarmItem.start_value}}</span>
            </li>
            <li>
              <label class="col-md-1">阀值:</label><span>{{alarmItem.s_cond}}</span>
            </li>
            <li>
              <label class="col-md-1">持续时间:</label><span>{{alarmItem.s_last}}</span>
            </li>
          </template>
        </ul>
      </section>
    </template>  
  </div>
</template>

<script>
export default {
  name: '',
  data() {
    return {
      timeForDataAchieve: null,
      filters: {},
      filtersForShow: [],
      actveAlarmIndex: null,
      resultData: []
    }
  },
  mounted(){
    this.getAlarm()
  },
  methods: {
    getAlarm() {
      let temp = []
      let keys = Object.keys(this.filters)
      this.filtersForShow = []
      for (let i = 0; i< keys.length ;i++) {
        temp.push(`${keys[i]}=${this.filters[keys[i]]}`)
        this.filtersForShow.push({key:keys[i], value:this.filters[keys[i]]})
      }
      let params = {filter:temp}
      this.timeForDataAchieve = new Date().toLocaleString()
      this.$httpRequestEntrance.httpRequestEntrance('GET', 'alarm/problem/list', params, (responseData) => {
        this.resultData = responseData
      })
    },
    addParams (key, value) {
      this.filters[key] = value
      this.getAlarm()
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
label {
  margin-bottom: 0;
  text-align: right;
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
  border: 1px solid @color-orange-F;
  color: @color-orange-F;
}
.alarm-item-border-medium {
  border: 1px solid @blue-2;
  color: @blue-2;
}
.alarm-item-border-low {
  border: 1px solid @gray-d;
}

.alarm-item:hover {
  box-shadow: 0 0 12px @gray-d;
}

.alarm-item /deep/.ivu-icon-ios-close:before {
    content: "\F102";
}
</style>
