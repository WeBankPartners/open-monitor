<template>
  <div class="charts charts-click">
    <Tabs :value="activeTab" @on-click="changeTab">
      <template v-for="(chartItem, chartIndex) in charts.chartsConfig">
        <TabPane :label="chartItem.tabTape.label" :name="chartItem.tabTape.name" :key="chartIndex">
        </TabPane>
      </template>
    </Tabs>
    <section>
      <template v-if="btns.length">
        <div class="btn-content">
          <RadioGroup v-model="currentParameter" size="small" type="button">
            <Radio v-for="(btnItem,btnIndex) in btns" :label="btnItem.option_value" :key="btnIndex" >{{btnItem.option_text}}</Radio>
          </RadioGroup>
        </div>
      </template>
      <div class="box">
        <div v-for="(chartInfo,chartIndex) in activeCharts" :key="chartIndex" class="list">
          <SingleChart
            :chartInfo="chartInfo"
            :chartIndex="chartIndex"
            :params="params"
            :refreshNow="refreshNow"
            @sendConfig="receiveConfig"
            @editTitle="editTitle"
            @editShowLines="handleEditShowLines"
          >
          </SingleChart>
        </div>
        <div v-for="(ph) in phZone" class="list" :key="ph"></div>
      </div>
    </section>
    <ChartLinesModal
      :isLineSelectModalShow="isLineSelectModalShow"
      :chartId="setChartConfigId"
      @modalClose="onLineSelectChangeCancel"
    >
    </ChartLinesModal>
    <ModalComponent :modelConfig="modelConfig"></ModalComponent>
  </div>
</template>

<script>
import {isEmpty} from 'lodash'
import SingleChart from '@/components/single-chart'
import ChartLinesModal from '@/components/chart-lines-modal'

export const custom_api_enum = [
  {
    getTagDetail: 'get'
  },
  {
    getDashboardTags: 'get'
  }
]

export default {
  name: '',
  data() {
    return {
      activeTab: '',
      activeCharts: [],
      phZone: [], // 占位数据
      btns: [],
      tagsUrl: '',
      params: {},
      currentParameter: null,
      editChartConfig: null,
      modelConfig: {
        modalId: 'edit_Modal',
        modalTitle: 'm_button_chart_editTitle',
        saveFunc: 'titleSave',
        isAdd: true,
        config: [
          {
            label: 'm_tableKey_name',
            value: 'name',
            placeholder: 'm_tips_inputRequired',
            v_validate: 'required:true|min:2|max:60',
            disabled: false,
            type: 'text'
          }
        ],
        addRow: { // [通用]-保存用户新增、编辑时数据
          name: null
        },
      },
      isLineSelectModalShow: false,
      setChartConfigId: '',
      refreshNow: false,
      apiCenter: this.$root.apiCenter,
      request: this.$root.$httpRequestEntrance.httpRequestEntrance
    }
  },
  props: {
    charts: Object
  },
  watch: {
    currentParameter() {
      this.pitchOnBtn()
    },
    activeCharts(val) {
      this.phZone = []
      const len = val.length
      if (!len) {
        return
      }
      const remainder = 6 - len%6
      if (remainder) {
        for (let i = 0; i < remainder; i++) {
          this.phZone.push(Math.random())
        }
      }
    }
  },
  mounted() {
    window['view-config-selected-line-data'] = {}
    if (this.charts.chartsConfig.length !== 0) {
      this.activeCharts = this.charts.chartsConfig[0].charts
      this.refreshCharts()
    }
  },
  methods: {
    refreshCharts() {
      if (this.$root.$validate.isEmpty_reset(this.activeTab)
      || this.charts.chartsConfig.findIndex(element => (element.tabTape.name === this.activeTab)) === -1) {
        this.activeTab = this.charts.chartsConfig[0].tabTape.name
      }
      this.changeTab(this.activeTab)
    },
    changeTab(name) {
      this.params = this.charts.chartsConfig[0].params
      this.activeTab = name
      this.activeCharts = []
      this.btns = []
      this.charts.chartsConfig.forEach(item => {
        if (item.tabTape.name === name) {
          this.btns = item.btns
          if (this.btns.length !== 0) {
            this.currentParameter = this.btns[0].option_value
          }
          this.tagsUrl = '/monitor/api/v1'+ item.tagsUrl
          this.$nextTick(() => {
            this.activeCharts = item.charts
          })
        }
      })
    },
    pitchOnBtn() {
      this.request('GET',this.tagsUrl + this.currentParameter, '', responseData => {
        this.activeCharts.forEach((element,index) => {
          element.metric = responseData[index].metric
        })
        this.activeCharts = []
        this.charts.chartsConfig.forEach(item => {
          if (item.tabTape.name === this.activeTab) {
            this.$nextTick(() => {
              this.activeCharts = item.charts
            })
          }
        })
      })
    },
    receiveConfig(chartItem) {
      this.$root.$eventBus.$emit('clearSingleChartInterval')
      this.$parent.showMaxChart = true
      this.$parent.$refs.maxChart.getChartData(chartItem)
      return
    },
    editTitle(config) {
      this.modelConfig.addRow.name = config.title
      this.editChartConfig = config
      this.$root.JQ('#edit_Modal').modal('show')
    },
    titleSave() {
      const params = {
        chart_id: this.editChartConfig.id,
        metric: this.editChartConfig.metric,
        name: this.modelConfig.addRow.name
      }
      this.request('POST', this.apiCenter.editTitle.api, params, () => {
        this.$root.JQ('#edit_Modal').modal('hide')
        // this.refreshCharts()
        this.$emit('refreshConfig')
      })
    },
    onLineSelectChangeCancel() {
      this.isLineSelectModalShow = false
      this.refreshNow = !this.refreshNow
    },
    handleEditShowLines(config) {
      this.setChartConfigId = config.id + ''
      if (isEmpty(window['view-config-selected-line-data'][config.id])) {
        window['view-config-selected-line-data'][config.id + ''] = {}
        config.legend.forEach(one => {
          window['view-config-selected-line-data'][config.id][one] = true
        })
      }
      this.isLineSelectModalShow = true
    }
  },
  components: {
    SingleChart,
    ChartLinesModal
  }
}
</script>
<style scoped lang="less">
  .charts {
    padding-top: 20px;
  }

  .btn-content {
    padding: 2px;
  }

 .box {
	display:flex;
	flex-wrap: wrap;
	justify-content: space-around;
}
.box .list{
	width: 580px;
}

</style>
