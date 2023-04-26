<template>
  <div>
    <Title :title="$t('menu.alert')"></Title>
    <Tabs value="default">
      <TabPane :label="$t('recommended_mode')" name="default">
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
        <div class="flex-container">
          <transition name="slide-fade">
            <div class="flex-item" v-show="showGraph">
              <div>
                <Tag color="success"><span style="font-size:14px">{{$t('m_low')}}:{{this.low}}</span></Tag>
                <Tag color="warning"><span style="font-size:14px">{{$t('m_medium')}}:{{this.mid}}</span></Tag>
                <Tag color="error"><span style="font-size:14px">{{$t('m_high')}}:{{this.high}}</span></Tag>
                <button v-if="filtersForShow.length" @click="clearAll" style="float:right;margin-right:50px" class="btn btn-small btn-cancel-f">{{$t('m_reset_condition')}}</button>
                <div v-show="alramEmpty" style="display:none" id="elId" class="echart"></div>
                <div v-if="!alramEmpty"  class="alarm-empty">
                  <span style="font-size:14px"></span>
                </div>
              </div>
            </div>
          </transition>
          <div class="flex-item" style="width: 100%">
            <div class="alarm-total" v-if="!showGraph">
              <Tag color="success"><span style="font-size:14px">{{$t('m_low')}}:{{this.low}}</span></Tag>
              <Tag color="warning"><span style="font-size:14px">{{$t('m_medium')}}:{{this.mid}}</span></Tag>
              <Tag color="error"><span style="font-size:14px">{{$t('m_high')}}:{{this.high}}</span></Tag>
            </div>
            <section style="margin-left:8px" class="c-dark-exclude-color">
              <div style="display: inline-block;margin-right:16px">
                <span>{{$t('alarmStatistics')}}：</span>
                <i-switch size="large" v-model="showGraph">
                  <span slot="open">ON</span>
                  <span slot="close">OFF</span>
                </i-switch>
              </div>
              <Tag color="warning">{{$t('title.updateTime')}}：{{timeForDataAchieve}}</Tag>
              <template>
                <Tag v-for="(filterItem, filterIndex) in filtersForShow" color="success" type="border" closable @on-close="exclude(filterItem.key)" :key="filterIndex">{{filterItem.key}}：{{filterItem.value}}</Tag>
              </template>
              <button v-if="filtersForShow.length" @click="clearAll" class="btn btn-small btn-cancel-f">{{$t('clearAll')}}</button>
              <template v-if="!resultData.length">
                <Tag color="primary">{{$t('table.noDataTip')}}！</Tag>
              </template>
              <button @click="alarmHistory" style="float: right;margin-right: 25px;" class="btn btn-sm btn-cancel-f">{{$t('alarmHistory')}}</button>
              <button :disabled="!filtersForShow.some(f => f.key === 'metric')" @click="deleteConfirmModal({}, true)" style="float: right;margin-right: 25px;" class="btn btn-sm btn-cancel-f">{{$t('m_batch_close')}}</button>
            </section>
            <div class="alarm-list">
              <template>
                <section v-for="(alarmItem, alarmIndex) in resultData" :key="alarmIndex" class="alarm-item c-dark-exclude-color" :class="'alarm-item-border-'+ alarmItem.s_priority">
                  <div style="float:right">
                    <Tooltip :content="$t('menu.endpointView')">
                      <Icon type="ios-stats" size="18" class="fa-operate" v-if="!alarmItem.is_custom" @click="goToEndpointView(alarmItem)"/>
                    </Tooltip>
                    <Tooltip :content="$t('close')">
                      <Icon type="ios-eye-off" size="18" class="fa-operate" @click="deleteConfirmModal(alarmItem, false)"/>
                    </Tooltip>
                    <Tooltip :content="$t('m_remark')">
                      <Icon type="ios-pricetags-outline" size="18" class="fa-operate" @click="remarkModal(alarmItem)" />
                    </Tooltip>
                  </div>
                  <ul>
                    <li>
                      <label class="col-md-2" style="vertical-align: top;line-height: 24px;">{{$t('field.endpoint')}}&{{$t('tableKey.s_priority')}}:</label>
                      <Tag type="border" closable @on-close="addParams('endpoint',alarmItem.endpoint)" color="primary">{{alarmItem.endpoint}}</Tag>
                      <Tag type="border" closable @on-close="addParams('priority',alarmItem.s_priority)" color="primary">{{alarmItem.s_priority}}</Tag>
                      <Tag type="border" color="warning">{{alarmItem.start_string}}</Tag>
                    </li>
                    <li v-if="!alarmItem.is_custom">
                      <label class="col-md-2" style="vertical-align: top;line-height: 24px;">
                        <span>{{$t('field.metric')}}</span>
                        <span v-if="alarmItem.tags">&{{$t('tableKey.tags')}}</span>
                        :</label>
                        <Tag type="border" closable @on-close="addParams('metric',alarmItem.s_metric)" color="primary">{{alarmItem.s_metric}}</Tag>
                        <template v-if="alarmItem.tags">
                          <Tag type="border" v-for="(t,tIndex) in alarmItem.tags.split('^')" :key="tIndex" color="cyan">{{t}}</Tag>
                        </template>
                    </li>
                    <li v-if="alarmItem.custom_message">
                      <label class="col-md-2" style="vertical-align: top;line-height: 24px;">
                        <span>{{$t('m_remark')}}:</span></label>
                        <Tooltip max-width="300">
                          <div style="border: 1px solid #2d8cf0;padding:2px;border-radius:4px; color: #2d8cf0">
                          {{alarmItem.custom_message.length > 100 ? alarmItem.custom_message.substring(0,100) + '...' : alarmItem.custom_message}}
                          </div>
                          <div slot="content" style="white-space: normal;">
                            <p>{{alarmItem.custom_message}}</p>
                          </div>
                        </Tooltip>
                    </li>
                    <li  v-if="!alarmItem.is_custom">
                      <label class="col-md-2" style="vertical-align: top;line-height: 24px;">{{$t('details')}}:</label>
                      <div class="col-md-9" style="display: inline-block;padding:0">
                        <span>
                          <Tag color="default">{{$t('tableKey.start_value')}}:{{alarmItem.start_value}}</Tag>
                          <Tag color="default" v-if="alarmItem.s_cond">{{$t('tableKey.threshold')}}:{{alarmItem.s_cond}}</Tag>
                          <Tag color="default" v-if="alarmItem.s_last">{{$t('tableKey.s_last')}}:{{alarmItem.s_last}}</Tag>
                          <Tag color="default" v-if="alarmItem.path">{{$t('tableKey.path')}}:{{alarmItem.path}}</Tag>
                          <Tag color="default" v-if="alarmItem.keyword">{{$t('tableKey.keyword')}}:{{alarmItem.keyword}}</Tag>
                        </span>
                      </div>
                    </li>
                    <li>
                      <label class="col-md-2" style="vertical-align: top;">{{$t('alarmContent')}}:</label>
                      <div class="col-md-9" style="display: inline-block;padding:0">
                        <span style="word-break: break-all;" v-html="alarmItem.content"></span>
                      </div>
                    </li>
                  </ul>
                </section>
              </template>
            </div>
            <div style="margin: 4px 0; text-align:right">
              <Page :total="paginationInfo.total" @on-change="pageIndexChange" @on-page-size-change="pageSizeChange" show-elevator show-sizer show-total />
            </div>
          </div>
        </div>
      </TabPane>
      <TabPane :label="$t('classic_mode')" name="classic">
        <ClassicAlarm ref="classicAlarm"></ClassicAlarm>
      </TabPane>
    </Tabs>
    <ModalComponent :modelConfig="modelConfig"></ModalComponent>
  </div>
</template>

<script>
import echarts from 'echarts'
import ClassicAlarm from '@/views/alarm-management-classic'
export default {
  name: '',
  data() {
    return {
      showGraph: true,
      alramEmpty: true,
      myChart: null,
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
      paginationInfo: {
        total: 0,
        startIndex: 1,
        pageSize: 10
      },
      isBatch: false
    }
  },
  mounted(){
    this.myChart = echarts.init(document.getElementById('elId'))
    this.getAlarm()
    this.interval = setInterval(()=>{
      this.getAlarm()
    }, 10000)
    this.$once('hook:beforeDestroy', () => {
      clearInterval(this.interval)
    })
  },
  methods: {
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
        this.getAlarm()
        this.$root.JQ('#remark_Modal').modal('hide')
      })
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
    pageIndexChange(pageIndex) {
      this.paginationInfo.startIndex = pageIndex
      this.getAlarm()
    },
    pageSizeChange(pageSize) {
      this.paginationInfo.startIndex = 1
      this.paginationInfo.pageSize = pageSize
      this.getAlarm()
    },
    getAlarm() {
      let params = {
        page: {
          startIndex: this.paginationInfo.startIndex,
          pageSize: this.paginationInfo.pageSize
        }
      }
      let keys = Object.keys(this.filters)
      this.filtersForShow = []
      for (let i = 0; i< keys.length ;i++) {
        params[keys[i]] = this.filters[keys[i]]
        this.filtersForShow.push({key:keys[i], value:this.filters[keys[i]]})
      }
      
      this.timeForDataAchieve = new Date().toLocaleString()
      this.timeForDataAchieve = this.timeForDataAchieve.replace('上午', 'AM ')
      this.timeForDataAchieve = this.timeForDataAchieve.replace('下午', 'PM ')
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', '/monitor/api/v1/alarm/problem/page', params, (responseData) => {
        this.resultData = responseData.data
        this.paginationInfo.total = responseData.page.totalRows
        this.paginationInfo.startIndex = responseData.page.startIndex
        this.paginationInfo.pageSize = responseData.page.pageSize
        this.low = responseData.low
        this.mid = responseData.mid
        this.high = responseData.high
        this.alramEmpty = !!this.low || !!this.mid ||!!this.high
        this.showSunburst(responseData)
        this.$refs.classicAlarm.getAlarm(this.resultData)
      }, {isNeedloading: false})
    },
    compare (prop) {
      return function (obj1, obj2) {
        var val1 = obj1[prop];
        var val2 = obj2[prop];
        if (!isNaN(Number(val1)) && !isNaN(Number(val2))) {
          val1 = Number(val1);
          val2 = Number(val2);
        }
        if (val1 < val2) {
          return -1;
        } else if (val1 > val2) {
          return 1;
        } else {
          return 0;
        }  
      } 
    },
    showSunburst (originData) {
      this.myChart.off()
      let legendData = []
      let pieInner = []
      if (originData.high) {
        let high = {
          name: 'high',
          value: originData.high,
          filterType: 'priority',
          itemStyle: {
            color: '#ed4014'
          }
        }
        legendData.push('high')
        pieInner.push(high)
      }
      if (originData.low) {
        let low = {
          name: 'low',
          value: originData.low,
          filterType: 'priority',
          itemStyle: {
            color: '#19be6b'
          }
        }
        legendData.push('low')
        pieInner.push(low)
      }
      if (originData.mid) {
        let mid = {
          name: 'medium',
          value: originData.mid,
          filterType: 'priority',
          itemStyle: {
            color: '#2d8cf0'
          }
        }
        legendData.push('medium')
        pieInner.push(mid)
      }

      const colorX = ['#33CCCC','#666699','#66CC66','#996633','#9999CC','#339933','#339966','#663333','#6666CC','#336699','#3399CC','#33CC66','#CC3333','#CC6666','#996699','#CC9933']
      let index = 0
      let pieOuter = []
      let itemStyleSet = {}
      const metricInfo = originData.count
      let set = new Set()
      metricInfo.forEach(item => {
        if (set.has(item.name)) {
          item.itemStyle = itemStyleSet[item.name]
        } else {
          legendData.push(item.name)
          index++
          const itemStyle = {
            color: colorX[index]
          }
          itemStyleSet[item.name] = itemStyle
          item.itemStyle = itemStyle
        }
        set.add(item.name)
      })
      pieOuter = metricInfo.sort(this.compare('type'))
      let option = {
        backgroundColor: '#ffffff',
          tooltip: {
              trigger: 'item',
              formatter: '{b}: {c}'
          },
          legend: {
            bottom: '15%',
            selectedMode: false,
            data: legendData
          },
          series: [
              {
                  type: 'pie',
                  selectedMode: 'single',
                  radius: [0, '30%'],
                  center: ['50%', '40%'],
                  label: {
                    formatter: '{b}:{c}',
                    position: 'inner',
                    rich: {
                      b: {
                        fontSize: 16,
                        lineHeight: 33
                      }
                    }
                  },
                  labelLine: {
                      show: false
                  },
                  data: pieInner
              },
              {
                  type: 'pie',
                  radius: ['40%', '55%'],
                  center: ['50%', '40%'],
                  label: {
                      formatter: ' {b|{b}:}{c} ',
                      backgroundColor: '#ffffff',
                      borderColor: '#2d8cf0',
                      borderWidth: 1,
                      borderRadius: 4,
                      position: 'outer',
                      alignTo: 'edge',
                      margin: 8,
                      rich: {
                        b: {
                          fontSize: 12,
                          lineHeight: 28
                        }
                      }
                  },
                  data: pieOuter
              }
          ]
      }

      this.myChart.setOption(option)
      this.myChart.on('click', params => {
        this.addParams(params.data.filterType, params.data.name)
      })
    },
    addParams (key, value) {
      this.filters[key] = value
      this.getAlarm()
    },
    deleteConfirmModal (rowData, isBatch) {
      this.isBatch = isBatch
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
        id: 0,
        custom: true,
        metric: ""
      }
      if (this.isBatch) {
        let find = this.filtersForShow.find(f => f.key === 'metric')
        if (find) {
          params.metric = find.value
        }
      } else {
        params.id = alarmItem.id
      }
      if (!alarmItem.is_custom) {
        params.custom = false
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', this.$root.apiCenter.alarmManagement.close.api, params, () => {
        // this.$root.$eventBus.$emit('hideConfirmModal')
        this.getAlarm()
      })
    },
    clearAll () {
      this.filters = []
      this.getAlarm()
    },
    exclude (key) {
      delete this.filters[key]
      this.getAlarm()
    },
    alarmHistory () {
      this.$router.push({name: 'alarmHistory'})
    }
  },
  components: {
    ClassicAlarm
  }
}
</script>

<style scoped lang="less">
.echart {
  height: ~"calc(100vh - 200px)";
  width: ~"calc(100vw * 0.4)";
  background:#ffffff;
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
  float: right;
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
