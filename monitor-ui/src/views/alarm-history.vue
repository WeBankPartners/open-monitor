<template>
  <div>
    <div class="title-wrapper">
      <Title :title="$t('alarmHistory')">
      </Title>
      <div class="title-form">
        <ul>
          <li class="filter-li">
            <DatePicker 
              type="date" 
              :value="startDate" 
              @on-change="changeStartDate"
              format="yyyy-MM-dd HH:mm:ss" 
              placement="bottom-start" 
              :placeholder="$t('startDatePlaceholder')" 
              style="width: 220px">
            </DatePicker>
          </li>
          <li class="filter-li">
            <DatePicker 
              type="date" 
              :value="endDate" 
              @on-change="changeEndDate"
              format="yyyy-MM-dd HH:mm:ss" 
              placement="bottom-start" 
              :placeholder="$t('endDatePlaceholder')" 
              style="width: 220px">
            </DatePicker>
          </li>
          <li class="filter-li">
            <Select filterable clearable v-model="filter" style="width:80px">
              <Option v-for="item in filterList" :value="item.value" :key="item.value">{{ item.label }}</Option>
            </Select>
          </li>
          <li class="filter-li">
            <button class="btn btn-sm btn-confirm-f" @click="getAlarm">
              <i class="fa fa-search"></i>
              {{$t('button.search')}}
            </button>
          </li>
        </ul>
        <button class="btn btn-sm btn-confirm-f" @click="realTimeAlarm">
          {{$t('realTimeAlarm')}}
        </button>
      </div>
    </div>
    <div class="data-stats-container" v-if="showGraph">
      <div class="top-stats-container">
        <div class="left">
          <div class="metics-metal">
            <div class="col">
              <div class="title">{{$t('realTimeAlarm')}}</div>
              <img class="time-icon" src="../assets/img/icon_rltm.png" />
            </div>
          </div>
          <circle-item type="total" :title="$t('m_total')" :total="total" :value="total" :icon="totalIcon" />
          <circle-item type="low" :title="$t('m_low')" :total="total" :value="low" :icon="lowIcon" />
          <circle-item type="medium" :title="$t('m_medium')" :total="total" :value="mid" :icon="midIcon" />
          <circle-item type="high" :title="$t('m_high')" :total="total" :value="high" :icon="highIcon" />
        </div>
        <div class="right">
          <div class="metics-metal">
            <div class="col">
              <div class="title">{{$t('todayAlarm')}}</div>
              <img class="time-icon" src="../assets/img/icon_rltm.png" />
            </div>
          </div>
          <circle-item type="total" :title="$t('m_total')" :total="total" :value="total" />
          <circle-item type="low" :title="$t('m_low')" :total="total" :value="low" />
          <circle-item type="medium" :title="$t('m_medium')" :total="total" :value="mid" />
          <circle-item type="high" :title="$t('m_high')" :total="total" :value="high" />
        </div>
      </div>
    </div>
    <div class="data-stats-container" v-if="showGraph">
      <transition name="slide-fade">
        <div class="content-stats-container">
          <div class="left">
            <img class="bg" src="../assets/img/bgd_main_cube.png" />
            <img class="cube" width="640" height="640" src="../assets/img/the_cube.png" />
            <circle-rotate
              :icon="lowCircle"
              :value="low"
              :total="total"
              :deg="'-60deg'"
              :tx="0"
              :ty="-0.5"
            />
            <circle-rotate
              :icon="midCircle"
              :value="mid"
              :total="total"
              :deg="'60deg'"
              :tx="0"
              :ty="-0.5"
            />
            <circle-rotate
              :icon="highCircle"
              :value="high"
              :total="total"
              :deg="'0'"
              :tx="0"
              :ty="0.5"
            />
            <div class="cir low">
              <div class="text">
                <div class="title">{{ $t('m_low') }}</div>
                <div class="value">{{ getPercentage(low, total) }}%</div>
              </div>
            </div>
            <div class="cir mid">
              <div class="text">
                <div class="title">{{ $t('m_medium') }}</div>
                <div class="value">{{ getPercentage(mid, total) }}%</div>
              </div>
            </div>
            <div class="cir high">
              <div class="text">
                <div class="title">{{ $t('m_high') }}</div>
                <div class="value">{{ getPercentage(high, total) }}%</div>
              </div>
            </div>
            <div class="metrics-bar" v-show="outerMetrics && outerMetrics.length > 0">
              <div class="bar-item" v-for="(mtc, idx) in outerMetrics" :key="mtc.name + mtc.type" :style="{ background: barColors[idx % 13], height: '15px', width: `${100 * mtc.value / outerTotal}%` }">
                <Tooltip :content="`${mtc.name}: ${mtc.value}`" placement="top">
                  <div class="content">&nbsp;&nbsp;</div>
                </Tooltip>
              </div>
            </div>
          </div>
          <div class="right">
            
          </div>
        </div>
      </transition>
    </div>
  </div>
</template>

<script>
import CircleItem from "../components/circle-item.vue";
import CircleRotate from "../components/circle-rotate.vue";

export default {
  name: '',
  components: {
    CircleItem,
    CircleRotate
  },
  data() {
    return {
      startDate: '',
      endDate: '',
      filter:'start',
      filterList: [
        {label: 'all', value: 'all'},
        {label: 'start', value: 'start'}
      ],

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
      barColors: ['#DE4B7D', '#E57A50', '#D8CF6B', '#AFC8E4', '#002B55', '#EC6820', '#98B63F', '#0199D3', '#03519F', '#535557', '#60C7C4', '#A7D9BF', '#FFDB3B'],

      totalIcon: require('../assets/img/icon_alarm_ttl.png'),
      lowIcon: require('../assets/img/icon_alarm_L.png'),
      midIcon: require('../assets/img/icon_alarm_M.png'),
      highIcon: require('../assets/img/icon_alarm_H.png'),

      lowCircle: require('../assets/img/peichart_L.png'),
      midCircle: require('../assets/img/peichart_M.png'),
      highCircle: require('../assets/img/peichart_H.png'),
    }
  },
  computed: {
    total() {
      return this.low + this.mid + this.high
    },

  },
  mounted() {
    this.getTodayAlarm()
  },
  methods: {
    changeStartDate (data) {
      this.startDate = data
    },
    changeEndDate (data) {
      this.endDate = data
    },
    getTodayAlarm() {
      const start = new Date(new Date().toLocaleDateString()).getTime();
      const end = new Date(new Date().toLocaleDateString()).getTime() + 24 * 60 * 60 * 1000 - 1;
      const params = {
        start: parseInt(start / 1000, 10),
        end: parseInt(end / 1000, 10),
        filter: this.filter
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', '/monitor/api/v1/alarm/problem/history', params, (responseData) => {
        this.tlow = responseData.low
        this.tmid = responseData.mid
        this.thigh = responseData.high
      })
    },
    getAlarm() {
      if (!this.startDate || !this.endDate || Date.parse(new Date(this.startDate)) > Date.parse(new Date(this.endDate))) {
        this.$Message.error(this.$t('timeIntervalWarn'))
        return
      }
      if (this.startDate === this.endDate) {
        this.endDate = this.endDate.replace('00:00:00', '23:59:59')
      }
      const start = Date.parse(this.startDate)/1000
      const end = Date.parse(this.endDate)/1000
      let params = {
        start,
        end,
        filter: this.filter
      }
      let keys = Object.keys(this.filters)
      this.filtersForShow = []
      for (let i = 0; i< keys.length ;i++) {
        params[keys[i]] = this.filters[keys[i]]
        this.filtersForShow.push({key:keys[i], value:this.filters[keys[i]]})
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', '/monitor/api/v1/alarm/problem/history', params, (responseData) => {
        this.resultData = responseData.data
        this.low = responseData.low
        this.mid = responseData.mid
        this.high = responseData.high
        this.alramEmpty = !!this.low || !!this.mid ||!!this.high
        this.showSunburst(responseData)
      })
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
      this.outerMetrics = pieOuter
      this.outerTotal = pieOuter.reduce((n, m) => (n + m.value), 0)
    },
    realTimeAlarm () {
      this.$router.push('/alarmManagement')
    },
    addParams (key, value) {
      this.filters[key] = value
      this.getAlarm()
    },
    clearAll () {
      this.filters = []
      this.getAlarm()
    },
    exclude (key) {
      delete this.filters[key]
      this.getAlarm()
    },
    getPercentage(val, total) {
      return ((parseInt(val, 10) * 100 / parseInt(total, 10)) || 0).toFixed(2)
    }
  }
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
    margin-left: 21px;
    padding: 10px 0;
    flex: auto;
    border: 2px solid #F2F3F7;
    border-radius: 4px;
    display: flex;
    justify-content: space-between;

    /deep/.ivu-input {
      border: 1px solid #F2F3F7;
    }

    .btn-sm {
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
    margin: 12px 0;

    .left {
      position: relative;
      flex-basis: 60%;
      display: flex;
      justify-content: center;
      align-items: center;
      padding-top: 53.5px;
      padding-bottom: 94px;
      .bg {
        position: absolute;
        top: 0;
      }
      .low {
        position: absolute;
      }
      .mid {
        position: absolute;
        transform: translate(86.6px, -50px);
      }
      .high {
        position: absolute;
      }

      .cir {
        position: absolute;
        width: 16px;
        height: 16px;
        background: #FFFFFF;
        border-radius: 50%;

        .text {
          .title {
            font-size: 16px;
            color: #404144;
          }
          .value {
            font-size: 36px;
            font-weight: 500;
          }
        }

        &.low {
          border: 2px solid #6ED06D;
          transform: translate(-173px, -100px);

          .text {
            margin-left: -135px;

            .value {
              color: #6ED06D;
            }
          }
        }
        &.mid {
          border: 2px solid #F19D38;
          transform: translate(173px, -100px);

          .text {
            margin-left: 85px;

            .value {
              color: #F19D38;
            }
          }
        }
        &.high {
          border: 2px solid #DA4E2B;
          transform: translate(0, 200px);

          .text {
            margin-left: 10px;
            margin-top: 65px;

            .value {
              color: #DA4E2B;
            }
          }
        }
      }

      .metrics-bar {
        position: absolute;
        top: 719px;
        width: 750px;
        height: 31px;
        background: #FFFFFF;
        box-shadow: 0px 8px 15px 0px rgba(17,110,249,0.15);
        border-radius: 15px;
        display: flex;
        padding: 8px 10px;

        .bar-item {
          height: 15px;

          /deep/ .ivu-tooltip {
            width: 100%;
          }

          .content {
            width: 100%;
          }
        }

        .bar-item:nth-child(1) {
          border-radius: 7px 0 0 7px;
        }
        .bar-item:last-child {
          border-radius: 0 7px 7px 0;
        }
      }
    }
    .right {
      flex-basis: 40%;
      height: 100%;
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
