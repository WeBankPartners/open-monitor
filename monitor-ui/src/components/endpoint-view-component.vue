<template>
  <div class="endpoint-view-component" id="endpointView">
    <Search ref="search" />
    <Charts v-if="showCharts" :charts='charts' @refreshConfig="refreshConfig" ref="parentCharts" />
    <div v-if="recursiveViewConfig.length && showRecursive">
      <Recursive :recursiveViewConfig="recursiveViewConfig" :params="params"></Recursive>
    </div>
    <Drawer
      v-model="showMaxChart"
      transfer
      :title="$t('m_view_details')"
      :width="zoneWidth"
      :closable="true"
      :mask="true"
      :mask-style="{'z-index': 2000}"
      class-name="custom-drawer"
      @on-close="onDrawerClose"
    >
      <MaxChart ref="maxChart"></MaxChart>
    </Drawer>
    <Modal
      v-model="historyAlarmModel"
      width="1400"
      :mask-closable="false"
      :footer-hide="true"
      :fullscreen="isfullscreen"
      :title="$t('m_button_historicalAlert')"
    >
      <div slot="header" class="custom-modal-header">
        <span>
          {{$t('m_alarmHistory')}}
        </span>
        <Icon v-if="isfullscreen" @click="fullscreenChange" class="fullscreen-icon" type="ios-contract" />
        <Icon v-else @click="fullscreenChange" class="fullscreen-icon" type="ios-expand" />
      </div>
      <Table class='history-alarm-config' :columns="historyAlarmPageConfig.table.tableEle" :height="fullscreenTableHight" :data="historyAlarmPageConfig.table.tableData" />
      <Page
        class="table-pagination"
        :total="pagination.total"
        @on-change="(e) => {pagination.page = e; this.getHistoryAlarmData()}"
        @on-page-size-change="(e) => {pagination.pageSize = e; this.getHistoryAlarmData()}"
        :current="pagination.page"
        :page-size="pagination.pageSize"
        show-total
        show-sizer
      />
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
    label: 'm_low',
    buttonType: 'green'
  },
  medium: {
    label: 'm_medium',
    buttonType: 'gold'
  },
  high: {
    label: 'm_high',
    buttonType: 'red'
  }
}
export default {
  name: 'endpoint-view',
  props: {
    data: Object
  },
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
              key: 'alarm_name',
              width: 170,
            },
            {
              title: this.$t('m_tableKey_status'),
              width: 80,
              key: 'status'
            },
            {
              title: this.$t('m_menu_configuration'),
              key: 'strategyGroupsInfo',
              render: (h, params) => (
                <div domPropsInnerHTML={params.row.strategyGroupsInfo}></div>
              )
            },
            {
              title: this.$t('m_field_endpoint'),
              key: 'endpoint'
            },
            {
              title: this.$t('m_alarmContent'),
              key: 'content',
              width: 200,
              render: (h, params) => (
                <Tooltip transfer={true} placement="bottom-start" max-width="300">
                  <div slot="content">
                    <div domPropsInnerHTML={params.row.content}></div>
                  </div>
                  <div class='column-eclipse'>{params.row.content || '-'}</div>
                </Tooltip>
              )
            },
            {
              title: this.$t('m_tableKey_s_priority'),
              key: 's_priority',
              width: 100,
              render: (h, params) => (
                <Tag color={alarmLevelMap[params.row.s_priority].buttonType}>{this.$t(alarmLevelMap[params.row.s_priority].label)}</Tag>
              )
            },
            {
              title: this.$t('m_field_metric'),
              key: 'alarm_metric_list_join',
              render: (h, params) => (
                <Tooltip transfer={true} placement="bottom-start" max-width="300">
                  <div slot="content">
                    <div domPropsInnerHTML={params.row.alarm_metric_list_join}></div>
                  </div>
                  <div class='column-eclipse'>{params.row.alarm_metric_list_join || '-'}</div>
                </Tooltip>
              )
            },
            {
              title: this.$t('m_field_threshold'),
              key: 'alarm_detail',
              width: 200,
              render: (h, params) => (
                <Tooltip transfer={true} placement="bottom-start" max-width="300">
                  <div slot="content">
                    <div domPropsInnerHTML={params.row.alarm_detail}></div>
                  </div>
                  <span class='column-eclipse'>{params.row.alarm_detail}</span>
                </Tooltip>
              )
            },
            {
              title: this.$t('m_tableKey_start'),
              key: 'start_string',
              width: 120,
            },
            {
              title: this.$t('m_tableKey_end'),
              key: 'end_string',
              width: 120,
              render: (h, params) => {
                let res = params.row.end_string
                if (params.row.end_string === '0001-01-01 00:00:00') {
                  res = '-'
                }
                return h('span', res)
              }
            },
            {
              title: this.$t('m_remark'),
              key: 'custom_message',
              width: 120,
              render: (h, params) => (
                <div>{params.row.custom_message || '-'}</div>
              )
            },
          ]
        }
      },
      strategyNameMaps: {
        endpointGroup: 'm_base_group',
        serviceGroup: 'm_field_resourceLevel'
      },
      pagination: {
        page: 1,
        pageSize: 10,
        total: 0
      },
      endPointItem: {},
      apiCenter: this.$root.apiCenter,
      request: this.$root.$httpRequestEntrance.httpRequestEntrance
    }
  },
  created() {
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
    disabledEndpoint(val) {
      if (this.$refs.search) {
        this.$refs.search.disabledEndpoint(val)
      }
    },
    refreshConfig(endpointObj) {
      if (this.$refs.search) {
        this.$refs.search.getChartsConfig(endpointObj)
      }
    },
    manageCharts(chartsConfig, params) {
      if (params.sys) {
        this.params = params
        this.showCharts = false
        this.recursiveView(params)
        return
      }
      this.charts.chartsConfig = []
      chartsConfig.forEach(item => {
        item.autoRefresh = params.autoRefresh
        const chart = {
          tabTape: {
            label: item.title,
            name: item.title + '_',
          },
          btns: item.tags.option,
          tagsUrl: item.tags.url,
          charts: item.charts,
          params
        }
        this.charts.chartsConfig.push(chart)
      })
      this.showCharts = true
      this.showRecursive = false
      this.$refs.parentCharts&&this.$refs.parentCharts.refreshCharts()
    },
    recursiveView(params) {
      this.recursiveViewConfig = []
      this.request('GET',this.apiCenter.recursive.api, params, responseData => {
        this.showRecursive = true
        this.recursiveViewConfig = [responseData]
      })
    },
    zoomChart(data) {
      this.showMaxChart = true
      setTimeout(() => {
        if (this.$refs.maxChart) {
          this.$refs.maxChart.enlargeChart(data)
        }
      }, 200)
    },
    // #region 历史告警
    historyAlarm(rowData) {
      this.pagination.page = 1
      this.pagination.pageSize = 10
      this.endPointItem = rowData
      this.pagination.total = 0
      this.historyAlarmPageConfig.table.tableData = []
      this.getHistoryAlarmData()
      this.isfullscreen = false
      this.historyAlarmModel = true
    },
    getHistoryAlarmData() {
      const params = {
        id: this.endPointItem.id,
        page: this.pagination.page,
        pageSize: this.pagination.pageSize,
        serviceGroup: this.endPointItem.option_value,
      }
      this.request('GET', this.apiCenter.alarm.history, params, responseData => {
        this.pagination.total = responseData.pageInfo.totalRows
        this.historyAlarmPageConfig.table.tableData = this.changeResultData(responseData.contents.problem_list || [])
      })
    },
    changeResultData(dataList) {
      if (dataList && !isEmpty(dataList)) {
        dataList.forEach(item => {
          item.strategyGroupsInfo = '-'
          item.alarm_metric_list_join = '-'
          if (!isEmpty(item.strategy_groups)) {
            item.strategyGroupsInfo = item.strategy_groups.reduce((res, cur) => res + this.$t(this.strategyNameMaps[cur.type]) + ':' + cur.name + '<br/> ', '')
          }

          if (!isEmpty(item.alarm_metric_list)) {
            item.alarm_metric_list_join = item.alarm_metric_list.join(',')
          }
        })
      }
      return dataList
    },
    fullscreenChange() {
      this.isfullscreen = !this.isfullscreen
      if (this.isfullscreen) {
        this.fullscreenTableHight = document.documentElement.clientHeight - 160
      } else {
        this.fullscreenTableHight = document.documentElement.clientHeight - 300
      }
    },
    onDrawerClose() {
      this.showMaxChart = false
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
<style lang="less" scoped>
.table-pagination {
  display: flex;
  justify-content: flex-end;
  margin-top: 10px;
}
.btn-jump {
  margin-left: 10px;
}
.custom-modal-header {
  line-height: 20px;
  font-size: 16px;
  color: #17233d;
  font-weight: 500;
  .fullscreen-icon {
    float: right;
    margin-right: 28px;
    font-size: 18px;
    cursor: pointer;
  }
}
</style>
<style lang="less">
.custom-drawer {
  z-index: 2000!important
}
.history-alarm-config {
  .column-eclipse {
    overflow: hidden;
    text-overflow: ellipsis;
    display: -webkit-box;
    -webkit-line-clamp: 3;
    -webkit-box-orient: vertical;
  }
}
</style>
