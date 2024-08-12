<template>
  <div class="all-content">
    <div class="title-wrapper">
      <div class="title-form">
        <ul>
          <li class="filter-li">
            <span class="label">{{$t('m_title_updateTime')}}：</span>{{timeForDataAchieve}}
          </li>
          <li class="filter-li">
            <span class="label">{{$t('alarmStatistics')}}：</span>
            <i-switch size="large" v-model="showGraph">
              <span slot="open">ON</span>
              <span slot="close">OFF</span>
            </i-switch>
          </li>
          <li class="filter-li">
            <span class="label">{{$t('classic_mode')}}：</span>
            <i-switch size="large" v-model="isClassicModel">
              <span slot="open">ON</span>
              <span slot="close">OFF</span>
            </i-switch>
          </li>
          <li class="filter-li" v-if="filtersForShow.length">
            <button @click="clearAll" class="btn btn-sm btn-cancel-f">{{$t('m_reset_condition')}}</button>
          </li>
        </ul>
        <div class="top-right-search">
          <SearchBadge :tempFilters="JSON.stringify(filters)" @filtersChange='onFiltersChange' />
          <Poptip
            confirm
            :title="$t('m_confirm_close_alarm')"
            placement="left-end"
            @on-ok="deleteConfirmModal()"
          >
            <Button
              :disabled="isEmpty(filters) || (isEmpty(filters.alarm_name) && isEmpty(filters.metric) && isEmpty(filters.endpoint)) || resultData.length === 0"
              class="btn btn-sm btn-cancel-f"
            >
              {{$t('m_batch_close')}}
            </Button>
          </Poptip>
          <button @click="alarmHistory" class="btn btn-sm btn-confirm-f">{{$t('alarmHistory')}}</button>
        </div>
      </div>
    </div>
    <div class="data-stats-container">
      <top-stats :lstats="leftStats" :rstats="rightStats" :rtitle="$t('todayAlarm')" :noData="noData" />
    </div>
    <div class="data-stats-container" v-show="!isClassicModel">
      <transition name="slide-fade">
        <div class="content-stats-container">
          <div class="left" :class="{'cover': total === 0 || noData}" v-if="showGraph">
            <alarm-assets-basic :total="total" :noData="noData" :isRunning="true" />

            <template v-if="!noData">
              <circle-label v-for="cr in circles" :key="cr.type" :data="cr" />
              <circle-rotate v-for="cr in circles" :key="cr.label" :data="cr" @onFilter="addParams" />
            </template>

            <metrics-bar :metrics="outerMetrics" :total="outerTotal" v-if="total > 0 && !noData" @onFilter="addParams" />
          </div>
          <div class="right" :class="{'cover': !showGraph}" v-if="total > 0 && !noData">
            <!-- <section style="margin-left:8px;margin-bottom:10px" class="c-dark-exclude-color"> -->
            <!-- <template v-for="(filterItem, filterIndex) in filtersForShow">
                <Tag color="success" type="border" closable @on-close="exclude(filterItem.key)" :key="filterIndex">{{filterItem.key}}：{{filterItem.value}}</Tag>
              </template> -->
            <!-- <button v-if="filtersForShow.length" @click="clearAll" class="btn btn-small btn-cancel-f">{{$t('clearAll')}}</button> -->
            <!-- </section> -->
            <section class="alarm-card-container">
              <alarm-card v-for="(item, alarmIndex) in resultData" @openRemarkModal="remarkModal" :key="alarmIndex" :data="item" :button="true"/>
            </section>
            <div style="margin: 4px 0; text-align:right">
              <Page :total="paginationInfo.total" @on-change="pageIndexChange" @on-page-size-change="pageSizeChange" show-elevator show-sizer show-total />
            </div>
          </div>
        </div>
      </transition>
    </div>
    <ClassicAlarm ref="classicAlarm" v-show="isClassicModel">
      <template v-slot:pagination>
        <div class="page-left">
          <Page :total="paginationInfo.total" @on-change="pageIndexChange" @on-page-size-change="pageSizeChange" show-elevator show-sizer show-total />
        </div>
      </template>
    </ClassicAlarm>
    <Modal
      :width="600"
      :title="$t('m_remark')"
      v-model="showRemarkModal"
    >
      <div>
        <Input v-model="modelConfig.addRow.message" type="textarea" placeholder="" />
      </div>
      <div slot="footer">
        <Button :disabled="modelConfig.addRow.message === ''" type="primary" @click="remarkAlarm">{{$t('m_button_save')}}</Button>
        <Button @click="cancelRemark">{{$t('m_button_cancel')}}</Button>
      </div>
    </Modal>
  </div>
</template>

<script>
import isEmpty from 'lodash/isEmpty'
import TopStats from '@/components/top-stats.vue'
import MetricsBar from '@/components/metrics-bar.vue'
import CircleRotate from '@/components/circle-rotate.vue'
import CircleLabel from '@/components/circle-label.vue'
import AlarmAssetsBasic from '@/components/alarm-assets-basic.vue'
import ClassicAlarm from '@/views/alarm-management-classic'
import AlarmCard from '@/components/alarm-card.vue'
import SearchBadge from '../components/search-badge.vue'

export default {
  name: '',
  components: {
    TopStats,
    MetricsBar,
    CircleRotate,
    CircleLabel,
    AlarmAssetsBasic,
    ClassicAlarm,
    AlarmCard,
    SearchBadge
  },
  data() {
    return {
      noData: false,
      showGraph: true,
      alramEmpty: true,
      isClassicModel: false,
      interval: null,
      timeForDataAchieve: null,
      filters: {},
      filtersForShow: [],
      actveAlarmIndex: null,
      resultData: [],
      outerMetrics: [],
      // selectedData: '', // 存放选中数据

      low: 0,
      mid: 0,
      high: 0,

      tlow: 0,
      tmid: 0,
      thigh: 0,

      outerTotal: 0,

      showRemarkModal: false,
      modelConfig: {
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
      isBatch: false,
      request: this.$root.$httpRequestEntrance.httpRequestEntrance,
      isEmpty
    }
  },
  computed: {
    total() {
      return this.low + this.mid + this.high
    },
    ttotal() {
      return this.tlow + this.tmid + this.thigh
    },
    leftStats() {
      return [
        {
          key: 'l_total',
          type: 'total',
          title: this.$t('m_total'),
          total: this.total,
          value: this.total,
          icon: require('../assets/img/icon_alarm_ttl.png')
        },
        {
          key: 'l_low',
          type: 'low',
          title: this.$t('m_low'),
          total: this.total,
          value: this.low,
          icon: require('../assets/img/icon_alarm_L.png')
        },
        {
          key: 'l_medium',
          type: 'medium',
          title: this.$t('m_medium'),
          total: this.total,
          value: this.mid,
          icon: require('../assets/img/icon_alarm_M.png')
        },
        {
          key: 'l_high',
          type: 'high',
          title: this.$t('m_high'),
          total: this.total,
          value: this.high,
          icon: require('../assets/img/icon_alarm_H.png')
        }
      ]
    },
    rightStats() {
      return [
        {
          key: 'r_total',
          type: 'total',
          title: this.$t('m_total'),
          total: this.ttotal,
          value: this.ttotal
        },
        {
          key: 'r_low',
          type: 'low',
          title: this.$t('m_low'),
          total: this.ttotal,
          value: this.tlow
        },
        {
          key: 'r_medium',
          type: 'medium',
          title: this.$t('m_medium'),
          total: this.ttotal,
          value: this.tmid
        },
        {
          key: 'r_high',
          type: 'high',
          title: this.$t('m_high'),
          total: this.ttotal,
          value: this.thigh
        }
      ]
    },
    circles() {
      return [
        {
          type: 'low',
          key: 'low',
          label: this.$t('m_low'),
          icon: require('../assets/img/peichart_L.png'),
          value: this.noData ? 0 : this.low,
          total: this.total,
          deg: '-60deg',
          tx: 0,
          ty: -0.5
        },
        {
          type: 'mid',
          key: 'medium',
          label: this.$t('m_medium'),
          icon: require('../assets/img/peichart_M.png'),
          value: this.noData ? 0 : this.mid,
          total: this.total,
          deg: '60deg',
          tx: 0,
          ty: -0.5
        },
        {
          type: 'high',
          key: 'high',
          label: this.$t('m_high'),
          icon: require('../assets/img/peichart_H.png'),
          value: this.noData ? 0 : this.high,
          total: this.total,
          deg: '0',
          tx: 0,
          ty: 0.5
        }
      ]
    }
  },
  mounted(){
    this.getTodayAlarm()
    this.getAlarm()
    this.interval = setInterval(() => {
      this.getAlarm('keep')
    }, 10000)
    this.$once('hook:beforeDestroy', () => {
      clearInterval(this.interval)
    })
  },
  methods: {
    getTodayAlarm() {
      const start = new Date(new Date().toLocaleDateString()).getTime()
      const end = new Date(new Date().toLocaleDateString()).getTime() + 24 * 60 * 60 * 1000 - 1
      const params = {
        start: parseInt(start / 1000, 10),
        end: parseInt(end / 1000, 10),
        filter: 'all'
      }
      this.request(
        'POST',
        '/monitor/api/v1/alarm/problem/history',
        params,
        responseData => {
          this.tlow = responseData.low
          this.tmid = responseData.mid
          this.thigh = responseData.high
        }
      )
    },
    remarkModal(item) {
      this.modelConfig.addRow = {
        id: item.id,
        message: item.custom_message,
        is_custom: false
      }
      this.showRemarkModal = true
    },
    remarkAlarm() {
      this.request('POST', this.apiCenter.remarkAlarm, this.modelConfig.addRow, () => {
        this.$Message.success(this.$t('m_tips_success'))
        this.getAlarm()
        this.showRemarkModal = false
      })
    },
    cancelRemark() {
      this.showRemarkModal = false
    },
    pageIndexChange(pageIndex) {
      this.paginationInfo.startIndex = pageIndex
      this.getAlarm('keep')
    },
    pageSizeChange(pageSize) {
      this.paginationInfo.startIndex = 1
      this.paginationInfo.pageSize = pageSize
      this.getAlarm('keep')
    },
    getAlarm(ifPageKeep) {
      if (ifPageKeep !== 'keep') {
        this.paginationInfo = {
          total: 0,
          startIndex: 1,
          pageSize: 10
        }
      }
      const params = {
        page: {
          startIndex: this.paginationInfo.startIndex,
          pageSize: this.paginationInfo.pageSize
        }
      }
      const keys = Object.keys(this.filters)
      this.filtersForShow = []
      for (let i = 0; i< keys.length; i++) {
        params[keys[i]] = this.filters[keys[i]]
        this.filtersForShow.push({
          key: keys[i],
          value: this.filters[keys[i]]
        })
      }

      this.timeForDataAchieve = new Date().toLocaleString()
      this.timeForDataAchieve = this.timeForDataAchieve.replace('上午', 'AM ')
      this.timeForDataAchieve = this.timeForDataAchieve.replace('下午', 'PM ')
      this.request(
        'POST',
        '/monitor/api/v1/alarm/problem/page',
        params,
        responseData => {
          this.noData = false
          this.resultData = responseData.data || []
          this.paginationInfo.total = responseData.page.totalRows
          this.paginationInfo.startIndex = responseData.page.startIndex
          this.paginationInfo.pageSize = responseData.page.pageSize
          this.low = responseData.low
          this.mid = responseData.mid
          this.high = responseData.high
          this.alramEmpty = !!this.low || !!this.mid ||!!this.high
          this.showSunburst(responseData)
          this.$refs.classicAlarm.getAlarm(this.resultData)
        },
        {isNeedloading: false},
        () => {
          this.noData = true
        }
      )
    },
    compare(prop) {
      return function (obj1, obj2) {
        let val1 = obj1[prop]
        let val2 = obj2[prop]
        if (!isNaN(Number(val1)) && !isNaN(Number(val2))) {
          val1 = Number(val1)
          val2 = Number(val2)
        }
        if (val1 < val2) {
          return -1
        }
        else if (val1 > val2) {
          return 1
        }
        return 0

      }
    },
    showSunburst(originData) {
      const legendData = []
      const pieInner = []
      if (originData.high) {
        const high = {
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
        const low = {
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
        const mid = {
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
      const itemStyleSet = {}
      const metricInfo = originData.count
      const set = new Set()
      metricInfo.forEach(item => {
        if (set.has(item.name)) {
          item.itemStyle = itemStyleSet[item.name]
        }
        else {
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

      this.outerMetrics = pieOuter
      this.outerTotal = pieOuter.reduce((n, m) => (n + m.value), 0)
    },
    addParams({key, value}) {
      this.filters[key] = value
      this.getAlarm()
    },
    deleteConfirmModal() {
      this.isBatch = true
      this.removeAlarm()
    },
    removeAlarm(alarmItem={}) {
      const params = {
        id: 0,
        custom: true,
        metric: [],
        priority: ''
      }
      if (this.isBatch) {
        const find = this.filtersForShow.find(f => f.key === 'metric')
        if (find) {
          params.metric = find.value
        }
        const priority = this.filtersForShow.find(f => f.key === 'priority')
        if (priority) {
          params.priority = priority.value
        }
        params.alarmName = this.filters.alarm_name || []
        params.endpoint = this.filters.endpoint || []
        params.metric = this.filters.metric || []
      }
      else {
        params.id = alarmItem.id
      }
      if (!alarmItem.is_custom) {
        params.custom = false
      }
      this.request('POST', this.$root.apiCenter.alarmManagement.close.api, params, () => {
        this.clearAll()
      })
    },
    clearAll() {
      this.filters = {}
      this.getAlarm()
    },
    exclude(key) {
      delete this.filters[key]
      this.getAlarm()
    },
    alarmHistory() {
      this.$router.push({name: 'alarmHistory'})
    },
    onFiltersChange(filters) {
      this.filters = filters
      this.getAlarm()
    }
  }
}
</script>

<style lang='less'>
.drop-down-content {
  .ivu-select-dropdown {
    overflow: scroll;
  }
}
.all-content {
  ::-webkit-scrollbar {
    display: none;
  }
}
</style>

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

.title-wrapper {
  display: flex;
  align-items: flex-end;
  margin-bottom: 24px;

  .title-form  {
    padding: 10px 18px;
    flex: auto;
    border-radius: 4px;
    border: 2px solid #f2f3f7;
    display: flex;
    justify-content: space-between;
    align-items: center;
    .top-right-search {
      display: flex;
      align-items: center;
    }

    .label {
      color: #116EF9;
      font-size: 14px;
      font-weight: bold;
    }

    ul {
      display: flex;
      align-items: center;
      li {
        color: #7E8086;
        font-size: 12px;
        margin-right: 28px;
      }
    }

    .btn-confirm-f {
      background: #116EF9;
    }
  }
}

.data-stats-container {

  .top-stats-container {
    width: 100%;
    height: 90px;
    background: #FFFFFF;
    border: 2px solid #F2F3F7;
    border-radius: 4px;
    display: flex;

    .metics-metal {
      height: 100%;
      background: linear-gradient(90deg, #F5F8FE 0%, rgba(234,242,253,0) 100%);

      .col {
        position: relative;
        width: 180px;
        height: 100%;
        display: flex;
        flex-direction: column;
        align-items: flex-start;
        justify-content: center;

        .title {
          font-size: 16px;
        }

        .time-icon {
          width: 32px;
          height: 32px;
          margin-top: 14px;
        }

        &::after {
          content: "";
          position: absolute;
          width: 2px;
          height: 63px;
          right: 0;
          background: #F2F3F7;
        }
      }
    }

    .left {
      flex-basis: 60%;
      height: 100%;
      display: flex;
    }
    .right {
      flex-basis: 40%;
      height: 100%;
      display: flex;
      justify-content: center;
      align-items: center;
    }
  }

  .content-stats-container {
    width: 100%;
    display: flex;
    // margin: 12px 0;

    .left {
      position: relative;
      flex-basis: 60%;
      display: flex;
      justify-content: center;
      align-items: center;
      padding-top: 53.5px;
      padding-bottom: 94px;

      &.cover {
        flex-basis: 100%;
      }
    }
    .right {
      flex-basis: 40%;
      overflow-x: auto;

      .alarm-card-container {
        // height: 740px;
        height: ~"calc(100vh - 310px)";
        overflow-y: auto;

        &::-webkit-scrollbar {
          width: 6px;
          height: 6px;
        }

        &::-webkit-scrollbar-thumb {
          // border-radius: 1em;
          background-color: rgba(0, 21, 41, 0.2);
        }

        &::-webkit-scrollbar-track {
          // border-radius: 1em;
          background-color: rgba(181, 164, 164, 0.2);
        }
      }

      &.cover {
        flex-basis: 100%;
      }
    }
  }
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

.page-left {
  display: flex;
  justify-content: flex-end;
  margin-top: 10px;
  margin-bottom: 10px;
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
