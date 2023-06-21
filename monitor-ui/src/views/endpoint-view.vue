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
              title: this.$t('tableKey.status'),
              width: 80,
              key: 'status'
            },
            {
              title: this.$t('tableKey.s_metric'),
              width: 200,
              key: 's_metric'
            },
            {
              title: this.$t('tableKey.start_value'),
              width: 120,
              key: 'start_value'
            },
            {
              title: this.$t('tableKey.s_cond'),
              width: 80,
              key: 's_cond'
            },
            {
              title: this.$t('tableKey.s_last'),
              width: 100,
              key: 's_last'
            },
            {
              title: this.$t('tableKey.s_priority'),
              width: 100,
              key: 's_priority'
            },
            {
              title: this.$t('tableKey.start'),
              width: 120,
              key: 'start_string'
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
            }
          ],
          btn: [],
        },
      },
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
