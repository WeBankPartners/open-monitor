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
      width="1200"
      :mask-closable="false"
      :footer-hide="true"
      :title="$t('button.historicalAlert')">
      <Table height="500" row-key="id" :columns="historyAlarmPageConfig.table.tableEle" :data="historyAlarmPageConfig.table.tableData"></Table>
    </Modal>
  </div>
</template>
<script>
import isEmpty from 'lodash/isEmpty'
import Search from '@/components/search'
import Charts from '@/components/charts'
import Recursive from '@/views/recursive-view/recursive'
import MaxChart from '@/components/max-chart'
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
      historyAlarmPageConfig: {
        table: {
          tableData: [],
          tableEle: [
            {
              title: this.$t('tableKey.endpoint'),
              width: 220,
              key: 'endpoint',
              tree: true
            },
            {
              title: this.$t('m_alarmName'), 
              width: 100,
              key: 'alarm_name'
            },
            {
              title: this.$t('tableKey.status'),
              width: 80,
              key: 'status'
            },

            {
              title: this.$t('menu.configuration'),
              width: 120,
              key: 'strategyGroupsInfo'
            },
            {
              title: this.$t('alarmContent'), 
              key: 'content',
              width: 150
            },
            {
              title: this.$t('tableKey.s_priority'), 
              key: 's_priority', 
              width: 80
            },
            { 
              title: this.$t('tableKey.start'), 
              key: 'start_string', 
              width: 100
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
              title: this.$t('field.metric'), 
              key: 'alarm_metric_list', 
              width: 150,
              render: (h, params) => {
                let res = '-'
                if (!isEmpty(params.row.alarm_metric_list)) {
                  res = params.row.alarm_metric_list.join(';')
                }
                return h('span', res);
              }
            },
            { 
              title: this.$t('field.threshold'), 
              key: 'alarm_detail', 
              width: 200,
              renderContent: true
            }
          ],
          btn: [],
        },
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
    historyAlarm(endpointObject) {
      let params = {
        id: endpointObject.id,
        guid: endpointObject.option_value
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.alarm.history, params, (responseData) => {
        responseData.forEach((item) => {
          item.children = item.problem_list
          item.id = item.endpoint + '--'
          if (endpointObject.id !== -1) {
            item._showChildren = true
          }
          if (!isEmpty(item.children)) {
            item.children.forEach(child => {
              child.strategyGroupsInfo = '-'
              if (!isEmpty(child.strategy_groups)) {
                child.strategyGroupsInfo = item.strategy_groups.reduce((res, cur)=> {
                  return res + this.$t(this.strategyNameMaps[cur.type]) + ':' + cur.name + '; '
                }, '')
              }
            })
          }
          return item
        })
        this.historyAlarmPageConfig.table.tableData = responseData
      })
      this.historyAlarmModel = true
    }
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
