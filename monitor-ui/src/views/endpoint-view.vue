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
        <TableTemp :table="historyAlarmPageConfig.table" :pageConfig="historyAlarmPageConfig"></TableTemp>
      </div>
    </ModalComponent>
  </div>
</template>
<script>
import Search from '@/components/search'
import TableTemp from '@/components/table-page/table'
import Charts from '@/components/charts'
import Recursive from '@/views/recursive-view/recursive'
import MaxChart from '@/components/max-chart'
const historyAlarmEle = [{
    title: 'tableKey.status',
    value: 'status',
    style: 'min-width:70px',
    display: true
  },
  {
    title: 'tableKey.s_metric',
    value: 's_metric',
    display: true
  },
  {
    title: 'tableKey.start_value',
    value: 'start_value',
    display: true
  },
  {
    title: 'tableKey.s_cond',
    value: 's_cond',
    style: 'min-width:70px',
    display: true
  },
  {
    title: 'tableKey.s_last',
    value: 's_last',
    style: 'min-width:65px',
    display: true
  },
  {
    title: 'tableKey.s_priority',
    value: 's_priority',
    display: true
  },
  {
    title: 'tableKey.start',
    value: 'start_string',
    style: 'min-width:200px',
    display: true
  },
  {
    title: 'tableKey.end',
    value: 'end_string',
    style: 'min-width:200px',
    display: true,
    'render': (item) => {
      if (item.end_string === '0001-01-01 00:00:00') {
        return '-'
      } else {
        return item.end_string
      }
    }
  }
]
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
          tableEle: historyAlarmEle,
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
    MaxChart,
    TableTemp
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
.btn-jump {
  margin-left: 10px;
}
</style>
