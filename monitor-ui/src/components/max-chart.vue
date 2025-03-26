<template>
  <div class="">
    <div class="c-dark max-chart">
      <div class="condition-zone">
        <ul>
          <li>
            <div class="condition condition-title c-black-gray">{{$t('m_button_MoM')}}</div>
            <div class="condition">
              <Checkbox v-model="is_mom_yoy" @on-change="YoY">{{$t('m_button_MoM')}}</Checkbox>
            </div>
          </li>
          <template v-if="is_mom_yoy">
            <li>
              <div class="condition condition-title c-black-gray">{{$t('m_field_timeInterval')}}</div>
              <div class="condition">
                <DatePicker type="daterange" split-panels :value="chartCondition.compareFirstDate" placement="bottom-start" @on-change="pickFirstDate" :placeholder="$t('m_placeholder_datePicker')" style="width: 200px"></DatePicker>
              </div>
            </li>
            <li>
              <div class="condition condition-title c-black-gray">{{$t('m_field_comparedTimeInterval')}}</div>
              <div class="condition">
                <DatePicker type="daterange" :value="chartCondition.compareSecondDate" split-panels placement="bottom-start" @on-change="pickSecondDate" :placeholder="$t('m_placeholder_comparedDatePicker')" style="width: 200px"></DatePicker>
              </div>
            </li>
          </template>
          <template v-else>
            <li>
              <div class="condition condition-title c-black-gray">{{$t('m_field_relativeTime')}}</div>
              <div class="condition">
                <RadioGroup v-model="chartCondition.timeTnterval" size="small" type="button">
                  <Radio label="-1800">30m</Radio>
                  <Radio label="-3600">1h</Radio>
                  <Radio label="-10800">3h</Radio>
                </RadioGroup>
              </div>
            </li>
            <li>
              <div class="condition condition-title c-black-gray">{{$t('m_field_timeInterval')}}</div>
              <div class="condition">
                <DatePicker
                  type="daterange"
                  :value="chartCondition.dateRange"
                  split-panels
                  placement="bottom-start"
                  :placeholder="$t('m_placeholder_datePicker')"
                  style="width: 200px"
                  :transfer="false"
                  @on-change="datePick"
                >
                </DatePicker>
              </div>
            </li>
            <li>
              <div class="condition condition-title c-black-gray">{{$t('m_field_aggType')}}</div>
              <div class="condition">
                <RadioGroup v-model="chartCondition.agg" size="small" type="button">
                  <Radio label="min">Min</Radio>
                  <Radio label="max">Max</Radio>
                  <Radio label="avg">Average </Radio>
                  <Radio label="p95">P95</Radio>
                  <Radio label="sum">Sum</Radio>
                  <Radio label="none">Original</Radio>
                  <Radio label="avg_nonzero">AverageNZ</Radio>
                </RadioGroup>
              </div>
            </li>
          </template>
        </ul>
      </div>
      <div class="chart-zone" v-if="isShowChart">
        <div :id="elId" class="echart" :style="chartStyle"></div>
      </div>
      <div v-else class="no-data-style">{{this.$t('m_noData')}}</div>
    </div>
    <ChartLinesModal
      :isLineSelectModalShow="isLineSelectModalShow"
      :chartId="setChartConfigId"
      @modalClose="onLineSelectChangeCancel"
    >
    </ChartLinesModal>
  </div>
</template>

<script>
import {isEmpty, cloneDeep} from 'lodash'
import {generateUuid} from '@/assets/js/utils'
import ChartLinesModal from '@/components/chart-lines-modal'
import {readyToDraw} from '@/assets/config/chart-rely'
export default {
  name: '',
  data() {
    return {
      isShowChart: true,
      chartItem: {},
      elId: null,
      is_mom_yoy: false,
      chartCondition: {
        timeTnterval: -1800,
        dateRange: ['',''],
        compareFirstDate: ['', ''],
        compareSecondDate: ['', ''],
        agg: 'none' // 聚合类型
      },
      chartStyle: {
        minHeight: '400px',
        minWidth: '700px'
      },
      isLineSelectModalShow: false,
      setChartConfigId: '',
      chartInstance: null,
      apiCenter: this.$root.apiCenter,
      request: this.$root.$httpRequestEntrance.httpRequestEntrance
    }
  },
  watch: {
    chartCondition: {
      handler(val) {
        this.isShowChart = true
        if (this.is_mom_yoy && (val.compareFirstDate[0] === '' || val.compareSecondDate[0] === '')) {
          this.isShowChart = false
          return
        }
        this.getChartData(this.chartItem)
      },
      deep: true
    }
  },
  mounted(){
    this.$on('editShowLines', this.handleEditShowLines)
    this.generatUuid()
  },
  methods: {
    datePick(data) {
      this.chartCondition.dateRange = data
      if (this.chartCondition.dateRange[0] !== '') {
        this.chartCondition.dateRange[0] = this.chartCondition.dateRange[0] + ' 00:00:01'
      }
      if (this.chartCondition.dateRange[1] !== '') {
        this.chartCondition.dateRange[1] = this.chartCondition.dateRange[1] + ' 23:59:59'
      }
      this.getChartData(this.chartItem)
    },
    pickFirstDate(data) {
      this.chartCondition.compareFirstDate = data
    },
    pickSecondDate(data) {
      this.chartCondition.compareSecondDate = data
    },
    getChartData(chartItem, start, end) {
      // 为兼容放大区域调用
      if (chartItem) {
        this.chartItem = chartItem
      }
      const params = {
        aggregate: this.chartCondition.agg,
        time_second: Number(this.chartCondition.timeTnterval),
        start: 0,
        end: 0,
        title: this.chartItem.title,
        unit: '',
        chart_id: this.chartItem.id,
        compare: {
          compare_first_start: this.chartCondition.compareFirstDate[0],
          compare_first_end: this.chartCondition.compareFirstDate[1],
          compare_second_start: this.chartCondition.compareSecondDate[0],
          compare_second_end: this.chartCondition.compareSecondDate[1]
        },
        data: []
      }
      this.chartItem.endpoint.forEach(ep => {
        this.chartItem.metric.forEach(me => {
          params.data.push({
            endpoint: ep,
            metric: me,
            compare_first_start: this.chartCondition.compareFirstDate[0],
            compare_first_end: this.chartCondition.compareFirstDate[1],
            compare_second_start: this.chartCondition.compareSecondDate[0],
            compare_second_end: this.chartCondition.compareSecondDate[1]
          })
        })
      })
      // params.data = [{
      //   endpoint: this.chartItem.endpoint[0],
      //   metric: this.chartItem.metric[0],
      //   compare_first_start: this.chartCondition.compareFirstDate[0],
      //   compare_first_end: this.chartCondition.compareFirstDate[1],
      //   compare_second_start: this.chartCondition.compareSecondDate[0],
      //   compare_second_end: this.chartCondition.compareSecondDate[1]
      // }]
      // 外部有时间传入(放大)，以传入时间为准
      if (this.chartCondition.dateRange.length !==0) {
        params.start = start ? start : (this.chartCondition.dateRange[0] ===''
          ?0 :Date.parse(this.chartCondition.dateRange[0].replace(/-/g, '/'))/1000),
        params.end = end ? end : (this.chartCondition.dateRange[1] ===''
          ?0 :Date.parse(this.chartCondition.dateRange[1].replace(/-/g, '/'))/1000)
      }
      this.request('POST', this.apiCenter.metricConfigView.api, params, responseData => {
        // const chartConfig = {eye: false,clear:true, zoomCallback: true}
        const chartConfig = {
          eye: false,
          clear: true,
          lineBarSwitch: true,
          canEditShowLines: true,
          dataZoom: false,
          chartId: this.elId
        }
        responseData.chartId = this.elId
        if (!isEmpty(responseData.legend)) {
          this.chartInstance = readyToDraw(this,responseData, 1, chartConfig)
          if (this.chartInstance) {
            this.chartInstance.on('legendselectchanged', params => {
              window['view-config-selected-line-data'][this.elId] = cloneDeep(params.selected)
            })
          }
        } else {
          this.isShowChart = false
        }

      })
    },
    YoY(status) {
      this.chartCondition.dateRange = ['', '']
      this.chartCondition.compareFirstDate = ['', '']
      this.chartCondition.compareSecondDate = ['', '']
      this.isShowChart = false
      if (!status) {
        this.isShowChart = true
      }
    },
    handleEditShowLines(config) {
      this.setChartConfigId = config.chartId
      if (isEmpty(window['view-config-selected-line-data'][this.setChartConfigId])) {
        window['view-config-selected-line-data'][this.setChartConfigId] = {}
        config.legend.forEach(one => {
          window['view-config-selected-line-data'][this.setChartConfigId][one] = true
        })
      }
      this.isLineSelectModalShow = true
    },
    onLineSelectChangeCancel() {
      this.isLineSelectModalShow = false
      this.getChartData(this.chartItem)
    },
    generatUuid() {
      generateUuid().then(elId => {
        this.elId = `id_${elId}`
        this.chartStyle.width = window.screen.width * 0.6 + 'px'
        this.chartStyle.height = window.screen.height * 0.4 + 'px'
      })
    },
    enlargeChart(data) {
      window['view-config-selected-line-data'] = {}
      this.generatUuid()
      this.getChartData(data)
    }
  },
  components: {
    ChartLinesModal
  },
}
</script>

<style scoped lang="less">
  .max-chart {
    min-height: 540px;
    background: white;
    z-index: 2;
    padding: 12px;
  }
  li {
    list-style: none;
  }
  .condition {
    display: inline-block;
  }
  .condition /deep/ .ivu-input {
    height: 24px;
  }
  .condition /deep/ .ivu-input-suffix i {
    line-height: 24px;
  }
  .condition-title {
    background: @gray-d;
    width: 110px;
    text-align: center;
    vertical-align: middle;
    margin: 4px 8px 4px 0;
    padding: 3px;
  }

  .chart-zone {
    margin-top: 12px;
  }
  .no-data-style {
    display: flex;
    width: 100%;
    justify-content: center;
    margin-top: 30px
  }
</style>
