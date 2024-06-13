<template>
  <div class="page" id="endpointView">
    <Title :title="$t('menu.endpointView')"></Title>
    <Search ref="search" />
    <Charts v-if="showCharts" :charts='charts' @refreshConfig="refreshConfig" ref="parentCharts" />
    <div v-if="recursiveViewConfig.length && showRecursive">
      <Recursive :recursiveViewConfig="recursiveViewConfig" :params="params"></Recursive>
    </div>
    <Drawer title="View details" :width="zoneWidth" :closable="false" v-model="showMaxChart">
      <MaxChart ref="maxChart"></MaxChart>
    </Drawer>
    <Modal
      v-model="historyAlarmModel"
      width="1400"
      :mask-closable="false"
      :footer-hide="true"
      :fullscreen="isfullscreen"
      :title="$t('button.historicalAlert')">
      <div slot="header" class="custom-modal-header">
        <span>
          {{$t('alarmHistory')}}
        </span>
        <Icon v-if="isfullscreen" @click="fullscreenChange" class="fullscreen-icon" type="ios-contract" />
        <Icon v-else @click="fullscreenChange" class="fullscreen-icon" type="ios-expand" />
      </div>
      <Table :columns="historyAlarmPageConfig.table.tableEle" :height="fullscreenTableHight" :data="historyAlarmPageConfig.table.tableData"></Table>
    </Modal>
  </div>
</template>
<script>
import isEmpty from 'lodash/isEmpty'
import Search from '@/components/search'
import Charts from '@/components/charts'
import Recursive from '@/views/recursive-view/recursive'
import MaxChart from '@/components/max-chart'
const alarmLevelMap = {
  low: {
    label: "m_low",
    buttonType: "green"
  },
  medium: {
    label: "m_medium",
    buttonType: "gold"
  },
  high: {
    label: "m_high",
    buttonType: "red"
  }
}
export default {
  name: 'endpoint-view',
  data() {
    return {
      showCharts: false,
      showRecursive: false,
      params: null,
      charts: {
        chartsConfig: []
      },
      recursiveViewConfig: [],
      showMaxChart: false,
      zoneWidth: '800',
      historyAlarmModel: false,
      isfullscreen: false,
      fullscreenTableHight: document.documentElement.clientHeight - 300,
      historyAlarmPageConfig: {
        table: {
          tableData: [],
          tableEle: [
            {
              title: this.$t('m_alarmName'),
              key: 'alarm_name'
            },
            {
              title: this.$t('tableKey.status'),
              width: 80,
              key: 'status'
            },
            {
              title: this.$t('menu.configuration'),
              key: 'strategyGroupsInfo',
              render: (h, params) => {
                return (
                  <div domPropsInnerHTML={params.row.strategyGroupsInfo}></div>
                )
              }
            },
            {
              title: this.$t('field.endpoint'),
              key: 'endpoint'
            },
            {
              title: this.$t('alarmContent'),
              key: 'content'
            },
            {
              title: this.$t('tableKey.s_priority'),
              key: 's_priority',
              width: 100,
              render: (h, params) => {
                return (
                  <Tag color={alarmLevelMap[params.row.s_priority].buttonType}>{this.$t(alarmLevelMap[params.row.s_priority].label)}</Tag>
                )
              }
            },
            {
              title: this.$t('field.metric'),
              key: 'alarm_metric_list_join'
            },
            {
              title: this.$t('field.threshold'),
              key: 'alarm_detail',
              width: 200,
              ellipsis: true,
              tooltip: true,
              render: (h, params) => {
                return (
                  <Tooltip transfer={true} placement="bottom-start" max-width="300">
                    <div slot="content">
                      <div domPropsInnerHTML={params.row.alarm_detail}></div>
                    </div>
                    <div domPropsInnerHTML={params.row.alarm_detail}></div>
                  </Tooltip>
                )
              }
            },
            {
              title: this.$t('tableKey.start'),
              key: 'start_string',
              width: 120,
            },
            {
              title: this.$t('tableKey.end'),
              key: 'end_string',
              width: 120,
              render: (h, params) => {
                let res = params.row.end_string
                if (params.row.end_string === '0001-01-01 00:00:00') {
                  res = '-'
                }
                return h('span', res);
              }
            },
            {
              title: this.$t('m_remark'),
              key: 'custom_message',
              width: 120,
              render: (h, params) => {
                return(
                  <div>{params.row.custom_message || '-'}</div>
                )
              }
            },
          ]
        }
      },
      strategyNameMaps: {
        "endpointGroup": "m_base_group",
        "serviceGroup": "field.resourceLevel"
      }
    }
  },
  created () {
    this.$root.$eventBus.$on('callMaxChart', data => {
      this.zoomChart(data)
    })
    this.$root.$eventBus.$on('refreshRecursive', () => {
      this.refreshConfig()
    })
    this.zoneWidth = window.screen.width * 0.65
  },
  mounted() {
    this.refreshConfig()
  },
  methods: {
    refreshConfig () {
      if (this.$refs.search) {
        this.$refs.search.getChartsConfig() 
      } 
    },
    manageCharts (chartsConfig, params) {
      if (params.sys) {
        this.params = params
        this.showCharts = false
        this.recursiveView(params)
        return
      }
      this.charts.chartsConfig = []
      chartsConfig.forEach(item => {
        item.autoRefresh = params.autoRefresh
        let chart = {
          tabTape: {
            label: item.title,
            name: item.title + '_',
          },
          btns: item.tags.option,
          tagsUrl: item.tags.url,
          charts: item.charts,
          params: params
        }
        this.charts.chartsConfig.push(chart)
      })
      this.showCharts = true
      this.showRecursive = false
      this.$refs.parentCharts&&this.$refs.parentCharts.refreshCharts()
    },
    recursiveView (params) {
      this.recursiveViewConfig = []
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET',this.$root.apiCenter.recursive.api, params, responseData => {
        this.showRecursive = true
        this.recursiveViewConfig = [responseData]
      })
    },
    zoomChart (data) {
      this.showMaxChart = true
      this.$refs.maxChart.getChartData(data)
    },
    //#region 历史告警
    historyAlarm(rowData) {
      let params = {
        id: rowData.id
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.alarm.history, params, (responseData) => {
        this.historyAlarmPageConfig.table.tableData = this.changeResultData(responseData[0].problem_list)
      })
      this.isfullscreen = false
      this.historyAlarmModel = true
    },
    changeResultData(dataList) {
      if (dataList && !isEmpty(dataList)) {
        dataList.forEach(item => {
          item.strategyGroupsInfo = '-';
          item.alarm_metric_list_join = '-';
          if (!isEmpty(item.strategy_groups)) {
            item.strategyGroupsInfo = item.strategy_groups.reduce((res, cur)=> {
              return res + this.$t(this.strategyNameMaps[cur.type]) + ':' + cur.name + '<br/> '
            }, '')
          }

          if (!isEmpty(item.alarm_metric_list)) {
            item.alarm_metric_list_join = item.alarm_metric_list.join(',')
          }
        });
      }
      return dataList
    },
    fullscreenChange () {
      this.isfullscreen = !this.isfullscreen
      if (this.isfullscreen) {
        this.fullscreenTableHight = document.documentElement.clientHeight - 160
      } else {
        this.fullscreenTableHight = document.documentElement.clientHeight - 300
      }
    },
    //#endregion
  },
  components: {
    Search,
    Charts,
    Recursive,
    MaxChart
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
.btn-jump {
  margin-left: 10px;
}
</style>
