<template>
  <div class="page" id="endpointView">
    <Title :title="$t('menu.endpointView')"></Title>
    <Search ref="search" />
    <Charts v-if="showCharts" :charts='charts' ref="parentCharts" />
    <div v-if="recursiveViewConfig.length && showRecursive">
      <Recursive :recursiveViewConfig="recursiveViewConfig" :params="params"></Recursive>
    </div>
    <Drawer title="View details" :width="zoneWidth" :closable="false" v-model="showMaxChart">
        <MaxChart ref="maxChart"></MaxChart>
    </Drawer>
    <ModalComponent :modelConfig="historyAlarmModel">
      <div slot="historyAlarm" style="max-height:400px;overflow-y:auto">
        <Table height="400" width="900" :columns="historyAlarmPageConfig.table.tableEle" :data="historyAlarmPageConfig.table.tableData"></Table>
      </div>
    </ModalComponent>
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

      historyAlarmModel: {
        modalId: 'history_alarm_Modal',
        modalTitle: 'button.historicalAlert',
        modalStyle: 'width:930px;max-width: none;',
        noBtn: true,
        isAdd: true,
        config: [{
          name: 'historyAlarm',
          type: 'slot'
        }]
      },
      historyAlarmPageConfig: {
        table: {
          tableData: [],
          tableEle: [{
              title: this.$t('tableKey.status'),
              key: 'status'
            },
            {
              title: this.$t('tableKey.s_metric'),
              width: 200,
              key: 's_metric'
            },
            {
              title: this.$t('tableKey.start_value'),
              key: 'start_value'
            },
            {
              title: this.$t('tableKey.s_cond'),
              key: 's_cond'
            },
            {
              title: this.$t('tableKey.s_last'),
              key: 's_last'
            },
            {
              title: this.$t('tableKey.s_priority'),
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
      this.$refs.search.getChartsConfig() 
    })
    this.zoneWidth = window.screen.width * 0.65
  },
  mounted() {
    this.$refs.search.getChartsConfig() 
  },
  methods: {
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
      this.$refs.parentCharts.refreshCharts()
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
        id: endpointObject.id
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.alarm.history, params, (responseData) => {
        this.historyAlarmPageConfig.table.tableData = responseData
      })
      this.$root.JQ('#history_alarm_Modal').modal('show')
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
