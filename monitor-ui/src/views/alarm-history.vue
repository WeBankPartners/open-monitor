<template>
  <div>
    <global-loading :isSpinShow='isSpinShow' :showText="$t('m_is_requesting')" />
    <div class="title-wrapper">
      <Title :title="$t('m_alarmHistory')"> </Title>
      <div class="title-form">
        <ul>
          <li class="filter-li">
            <DatePicker
              type="date"
              :value="startDate"
              @on-change="changeStartDate"
              format="yyyy-MM-dd HH:mm:ss"
              placement="bottom-start"
              :placeholder="$t('m_startDatePlaceholder')"
              style="width: 220px"
            >
            </DatePicker>
          </li>
          <li class="filter-li">
            <DatePicker
              type="date"
              :value="endDate"
              @on-change="changeEndDate"
              format="yyyy-MM-dd HH:mm:ss"
              placement="bottom-start"
              :placeholder="$t('m_endDatePlaceholder')"
              style="width: 220px"
            >
            </DatePicker>
          </li>
          <!-- <li class="filter-li">
            <Select v-model="filter" style="width: 220px">
              <Option value="all">{{ $t('all') }}</Option>
              <Option value="low">{{ $t('low') }}</Option>
              <Option value="medium">{{ $t('medium') }}</Option>
              <Option value="high">{{ $t('high')}}</Option>
            </Select>
          </li> -->
          <Button type="primary" @click="getAlarm" style="margin-left: 24px;">{{ $t("m_button_search") }}</Button>
        </ul>
        <div class='top-right-search'>
          <SearchBadge :tempFilters="JSON.stringify(filters)" @filtersChange='onFiltersChange' />
          <Button type="primary" @click="realTimeAlarm">{{ $t("m_realTimeAlarm") }}</Button>
        </div>
      </div>
    </div>
    <div class="data-stats-container" v-if="showGraph">
      <top-stats :lstats="leftStats" :rstats="rightStats" :rtitle="$t('m_alarmHistory')" :noData="noData" />
    </div>
    <div class="data-stats-container" v-if="showGraph">
      <transition name="slide-fade">
        <div class="content-stats-container">
          <div class="left" :class="{'cover': total === 0 || noData}">
            <alarm-assets-basic :total="total" :noData="total === 0 ? true : noData" :isRunning="false" />

            <template v-if="!noData && !loading && total > 0">
              <circle-label v-for="cr in circles" :key="cr.type" :data="cr" @onFilter="addParams" />
              <circle-rotate v-for="cr in circles" :key="cr.label" :data="cr" @onFilter="addParams" />
            </template>

            <metrics-bar :metrics="outerMetrics" :total="outerTotal" v-if="total > 0 && !noData" @onFilter="addParams" />
          </div>
          <div class="right" v-if="total > 0 && !noData">
            <section class="alarm-card-container">
              <alarm-card v-for="(item, alarmIndex) in resultData" :key="alarmIndex" :data="item"></alarm-card>
            </section>
            <div class='card-pagination'>
              <Page :total="paginationInfo.total" @on-change="pageIndexChange" @on-page-size-change="pageSizeChange" show-sizer show-total />
            </div>
          </div>
        </div>
      </transition>
    </div>
  </div>
</template>

<script>
import Vue from 'vue'
import {isEmpty, cloneDeep} from 'lodash'
import TopStats from '@/components/top-stats.vue'
import MetricsBar from '@/components/metrics-bar.vue'
import CircleRotate from '@/components/circle-rotate.vue'
import CircleLabel from '@/components/circle-label.vue'
import AlarmAssetsBasic from '@/components/alarm-assets-basic.vue'
import AlarmCard from '@/components/alarm-card.vue'
import SearchBadge from '../components/search-badge.vue'
import GlobalLoading from '../components/globalLoading.vue'

export default {
  name: '',
  components: {
    TopStats,
    MetricsBar,
    CircleRotate,
    CircleLabel,
    AlarmAssetsBasic,
    AlarmCard,
    SearchBadge,
    GlobalLoading
  },
  data() {
    return {
      startDate: new Date(new Date().toLocaleDateString()),
      endDate: new Date(),
      filter: 'all',
      loading: true,
      noData: false,
      showGraph: true,
      alramEmpty: true,
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
      tlow: 0,
      tmid: 0,
      thigh: 0,
      full_len: 280,
      outerMetrics: [],
      outerTotal: 0,
      paginationInfo: {
        total: 0,
        startIndex: 1,
        pageSize: 10
      },
      isSpinShow: false
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
          total: this.ttotal,
          value: this.ttotal,
          icon: require('../assets/img/icon_alarm_ttl.png')
        },
        {
          key: 'l_low',
          type: 'low',
          title: this.$t('m_low'),
          total: this.ttotal,
          value: this.tlow,
          icon: require('../assets/img/icon_alarm_L.png')
        },
        {
          key: 'l_medium',
          type: 'medium',
          title: this.$t('m_medium'),
          total: this.ttotal,
          value: this.tmid,
          icon: require('../assets/img/icon_alarm_M.png')
        },
        {
          key: 'l_high',
          type: 'high',
          title: this.$t('m_high'),
          total: this.ttotal,
          value: this.thigh,
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
          total: this.total,
          value: this.total
        },
        {
          key: 'r_low',
          type: 'low',
          title: this.$t('m_low'),
          total: this.total,
          value: this.low
        },
        {
          key: 'r_medium',
          type: 'medium',
          title: this.$t('m_medium'),
          total: this.total,
          value: this.mid
        },
        {
          key: 'r_high',
          type: 'high',
          title: this.$t('m_high'),
          total: this.total,
          value: this.high
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
          value: this.low,
          total: this.total,
          deg: '-60deg',
          tx: 0,
          ty: -0.5,
        },
        {
          type: 'mid',
          key: 'medium',
          label: this.$t('m_medium'),
          icon: require('../assets/img/peichart_M.png'),
          value: this.mid,
          total: this.total,
          deg: '60deg',
          tx: 0,
          ty: -0.5,
        },
        {
          type: 'high',
          key: 'high',
          label: this.$t('m_high'),
          icon: require('../assets/img/peichart_H.png'),
          value: this.high,
          total: this.total,
          deg: '0',
          tx: 0,
          ty: 0.5,
        },
      ]
    },
  },
  mounted() {
    this.getAlarm()
    this.getRealTimeAlarm()
  },
  methods: {
    changeStartDate(data) {
      this.startDate = data
    },
    changeEndDate(data) {
      if (data && data.indexOf('00:00:00') !== -1) {
        this.endDate = data.replace('00:00:00', '23:59:59')
      } else {
        this.endDate = data
      }
    },
    getRealTimeAlarm() {
      const params = {
        page: {
          startIndex: 1,
          pageSize: 10
        }
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance(
        'POST',
        '/monitor/api/v1/alarm/problem/page',
        params,
        responseData => {
          this.noData = false
          this.tlow = responseData.low
          this.tmid = responseData.mid
          this.thigh = responseData.high
        },
        {isNeedloading: false},
        () => {
          this.noData = true
        }
      )
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
      if (
        !this.startDate
        || !this.endDate
        || Date.parse(new Date(this.startDate))
          > Date.parse(new Date(this.endDate))
      ) {
        this.$Message.error(this.$t('m_timeIntervalWarn'))
        return
      }
      if (this.startDate === this.endDate) {
        this.endDate = this.endDate.replace('00:00:00', '23:59:59')
      }
      if (ifPageKeep !== 'keep') {
        this.paginationInfo = {
          total: 0,
          startIndex: 1,
          pageSize: this.paginationInfo.pageSize
        }
      }
      const start = Date.parse(this.startDate) / 1000
      const end = Date.parse(this.endDate) / 1000
      const params = {
        start,
        end,
        filter: this.filter,
        page: {
          startIndex: this.paginationInfo.startIndex,
          pageSize: this.paginationInfo.pageSize
        }
      }
      const filters = cloneDeep(this.filters)
      const endpointList = []
      !isEmpty(filters.endpoint) && filters.endpoint.forEach(val => {
        if (val.indexOf('$*$') > -1) {
          endpointList.push(val.split('$*$')[1])
        } else {
          endpointList.push(val)
        }
      })
      filters.endpoint = endpointList

      const keys = Object.keys(filters)
      this.filtersForShow = []
      for (let i = 0; i < keys.length; i++) {
        params[keys[i]] = filters[keys[i]]
        this.filtersForShow.push({
          key: keys[i],
          value: filters[keys[i]],
        })
      }
      if (this.isSpinShow === false) {
        this.isSpinShow = true
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance(
        'POST',
        '/monitor/api/v1/alarm/problem/history',
        params,
        responseData => {
          this.loading = false
          this.noData = false
          this.resultData = responseData.data
          this.low = responseData.low
          this.mid = responseData.mid
          this.high = responseData.high
          this.paginationInfo.total = responseData.page.totalRows
          this.paginationInfo.startIndex = responseData.page.startIndex
          this.paginationInfo.pageSize = responseData.page.pageSize
          this.alramEmpty = !!this.low || !!this.mid || !!this.high
          this.showSunburst(responseData)
          if (this.isSpinShow) {
            this.isSpinShow = false
          }
        },
        {isNeedloading: false},
        () => {
          if (this.isSpinShow) {
            this.isSpinShow = false
          }
          this.loading = false
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
        } else if (val1 > val2) {
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
            color: '#ed4014',
          },
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
            color: '#19be6b',
          },
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
            color: '#2d8cf0',
          },
        }
        legendData.push('medium')
        pieInner.push(mid)
      }

      const colorX = [
        '#33CCCC',
        '#666699',
        '#66CC66',
        '#996633',
        '#9999CC',
        '#339933',
        '#339966',
        '#663333',
        '#6666CC',
        '#336699',
        '#3399CC',
        '#33CC66',
        '#CC3333',
        '#CC6666',
        '#996699',
        '#CC9933',
      ]
      let index = 0
      let pieOuter = []
      const itemStyleSet = {}
      if (!isEmpty(originData.count)) {
        const metricInfo = originData.count
        const set = new Set()
        metricInfo.forEach(item => {
          if (set.has(item.name)) {
            item.itemStyle = itemStyleSet[item.name]
          } else {
            legendData.push(item.name)
            index++
            const itemStyle = {
              color: colorX[index],
            }
            itemStyleSet[item.name] = itemStyle
            item.itemStyle = itemStyle
          }
          set.add(item.name)
        })
        pieOuter = metricInfo.sort(this.compare('type'))
        this.outerMetrics = pieOuter
        this.outerTotal = pieOuter.reduce((n, m) => n + m.value, 0)
      }
    },
    realTimeAlarm() {
      this.$router.push('/alarmManagement')
    },
    addParams({key, value}) {
      Vue.set(this.filters, key, this.filters[key] || [])
      const singleArr = this.filters[key]
      if (singleArr.includes(value)) {
        singleArr.splice(singleArr.indexOf(value), 1)
      } else {
        singleArr.push(value)
      }
    },
    clearAll() {
      this.filters = {}
      this.getAlarm()
    },
    exclude(key) {
      delete this.filters[key]
      this.getAlarm()
    },
    getPercentage(val, total) {
      return ((parseInt(val, 10) * 100) / parseInt(total, 10) || 0).toFixed(2)
    },
    onFiltersChange(filters) {
      this.filters = filters
      this.getAlarm()
    }
  },
}
</script>

<style scoped lang="less">
.filter-li {
  display: inline-block;
  margin-left: 8px;
}
.echart {
  height: ~"calc(100vh - 180px)";
  width: ~"calc(100vw * 0.4)";
  background: #ffffff;
}
.alarm-empty {
  height: ~"calc(100vh - 180px)";
  width: ~"calc(100vw * 0.4)";
  text-align: center;
  padding: 50px;
  color: #2d8cf0;
}

.title-wrapper {
  display: flex;
  align-items: flex-end;
  margin-bottom: 24px;

  .title-form {
    margin-left: 21px;
    padding: 10px 0;
    flex: auto;
    border: 2px solid #f2f3f7;
    border-radius: 4px;
    display: flex;
    justify-content: space-between;

    /deep/.ivu-input {
      border: 1px solid #f2f3f7;
    }

    .top-right-search {
      display: flex;
      align-items: center;
      margin-right: 80px;
    }
  }
}

.data-stats-container {

  .content-stats-container {
    height: ~"calc(100vh - 250px)";
    width: 100%;
    display: flex;

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

      .bg {
        position: absolute;
        top: 0;
      }
    }
    .right {
      flex-basis: 40%;
      overflow-x: auto;

      .card-pagination {
        width: 40%;
        position: fixed;
        bottom: 0px;
        right: 0px;
        opacity: 1;
        background: #fff;
        padding-bottom: 20px;
      }

      .alarm-card-container {
        // height: ~"calc(100vh - 310px)";
        height: 740px;
        overflow-y: auto;
        padding-bottom: 40px;
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

    }
    .right::-webkit-scrollbar {
      display: none;
    }
  }
}

.flex-container {
  margin: 8px;
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
  height: ~"calc(100vh - 180px)";
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

/* 可以设置不同的进入和离开动画 */
/* 设置持续时间和动画函数 */
.slide-fade-enter-active {
  transition: all 0.3s ease;
}
.slide-fade-leave-active {
  transition: all 0.8s cubic-bezier(1, 0.5, 0.8, 1);
}
.slide-fade-enter, .slide-fade-leave-to
/* .slide-fade-leave-active for below version 2.1.8 */ {
  transform: translateX(10px);
  opacity: 0;
}
</style>
