<template>
  <div class>
    <header>
      <div class="search-container">
        <div>
          <div class="search-zone">
            <span class="params-title">{{$t('m_field_relativeTime')}}：</span>
            <Select filterable v-model="viewCondition.timeTnterval" :disabled="disableTime" style="width:80px"  @on-change="initPanal">
              <Option v-for="item in dataPick" :value="item.value" :key="item.value">{{ item.label }}</Option>
            </Select>
          </div>
          <div class="search-zone">
            <span class="params-title">{{$t('m_placeholder_refresh')}}：</span>
            <Select v-model="viewCondition.autoRefresh" :disabled="disableTime" style="width:100px" @on-change="refreshInterval" :placeholder="$t('m_placeholder_refresh')">
              <Option v-for="item in autoRefreshConfig" :value="item.value" :key="item.value">{{ item.label }}</Option>
            </Select>
          </div>
          <div class="search-zone">
            <span class="params-title">{{$t('m_field_timeInterval')}}：</span>
            <DatePicker
              type="datetimerange"
              :value="viewCondition.dateRange"
              split-panels
              format="yyyy-MM-dd HH:mm:ss"
              placement="bottom-start"
              @on-change="datePick"
              :placeholder="$t('m_placeholder_datePicker')"
              style="width: 320px"
            />
          </div>
          <div class="search-zone">
            <span class="params-title">{{$t('m_field_aggType')}}：</span>
            <RadioGroup v-model="viewCondition.agg" @on-change="initPanal" size="small" type="button">
              <Radio label="min">Min</Radio>
              <Radio label="max">Max</Radio>
              <Radio label="avg">Average </Radio>
              <Radio label="p95">P95</Radio>
              <Radio label="sum">Sum</Radio>
              <Radio label="none">Original</Radio>
              <Radio label="avg_nonzero">AverageNZ</Radio>
            </RadioGroup>
          </div>
        </div>
      </div>
    </header>
    <div class="zone zone-chart c-dark">
      <div class="col-md-12">
        <div class="zone-chart-title">{{panalTitle}}</div>
        <div v-if="!noDataTip">
          <div :id="elId" class="echart"  style="height:80vh"></div>
        </div>
        <div v-else class="echart echart-no-data-tip">
          <span>{{this.$t('m_noData')}}</span>
        </div>
      </div>
    </div>
  </div>
</template>
<script>
import Vue from 'vue'
import cloneDeep from 'lodash/cloneDeep'
import { generateUuid } from '@/assets/js/utils'
import { readyToDraw } from '@/assets/config/chart-rely'
import {dataPick, autoRefreshConfig} from '@/assets/config/common-config'
export default {
  name: '',
  data() {
    return {
      viewCondition: {
        timeTnterval: -1800,
        dateRange: ['', ''],
        autoRefresh: 10,
        agg: 'none'
      },
      dataPick,
      autoRefreshConfig,

      viewData: null,
      panalData: null,

      elId: null,
      noDataTip: false,
      panalTitle: '',
      panalUnit: '',
      interval: null,
      allParams: null
    }
  },
  created() {
    generateUuid().then(elId => {
      this.elId = `id_${elId}`
    })
  },
  destroyed() {
    clearInterval(this.interval)
  },
  computed: {
    disableTime() {
      return this.viewCondition.dateRange[0] !== '' && this.viewCondition.dateRange[1] !== ''
    }
  },
  methods: {
    initChart(params) {
      this.allParams = params
      for (const key in this.viewCondition) {
        Vue.set(this.viewCondition, key, cloneDeep(params.viewCondition[key]))
      }
      if (params.templateData.cfg) {
        this.panalDataList = JSON.parse(params.templateData.cfg)
        const temp = this.panalDataList.filter(item => item.viewConfig.id === params.panal.id)
        this.panalData = temp[0]
        Vue.set(this.viewCondition, 'agg', this.panalData.aggregate)
        this.initPanal()
        this.scheduledRequest()
      }
    },
    scheduledRequest() {
      if (this.viewCondition.autoRefresh > 0) {
        clearInterval(this.interval)
        this.interval = setInterval(() => {
          this.initPanal()
        },this.viewCondition.autoRefresh*1000)
      }
      else {
        clearInterval(this.interval)
      }
    },
    refreshInterval() {
      this.initPanal()
      this.scheduledRequest()
    },
    datePick(data) {
      this.viewCondition.dateRange = data
      if (this.viewCondition.dateRange[0] && this.viewCondition.dateRange[1]) {
        if (this.viewCondition.dateRange[0] === this.viewCondition.dateRange[1]) {
          this.viewCondition.dateRange[1] = this.viewCondition.dateRange[1].replace('00:00:00', '23:59:59')
        }
        this.viewCondition.autoRefresh = 0
        clearInterval(this.interval)
      }
      this.initPanal()
    },
    initPanal() {
      this.panalTitle = this.panalData.panalTitle
      this.panalUnit = this.panalData.panalUnit
      this.noDataTip = false
      if (this.$root.$validate.isEmpty_reset(this.panalData.query)) {
        return
      }
      const params = {
        aggregate: this.viewCondition.agg,
        agg_step: this.panalData.agg_step,
        time_second: this.viewCondition.timeTnterval,
        start: this.viewCondition.dateRange[0] ===''
          ?0 :Date.parse(this.viewCondition.dateRange[0].replace(/-/g, '/'))/1000,
        end: this.viewCondition.dateRange[1] ===''
          ?0 :Date.parse(this.viewCondition.dateRange[1].replace(/-/g, '/'))/1000,
        title: '',
        unit: '',
        data: [],
        custom_chart_guid: this.panalData.viewConfig.id,
        lineType: this.allParams.panal._activeCharts[0].lineType
      }
      this.panalData.query.forEach(item => {
        params.data.push(item)
      })
      if (params !== []) {
        this.$root.$httpRequestEntrance.httpRequestEntrance(
          'POST',
          this.$root.apiCenter.metricConfigView.api,
          params,
          responseData => {
            responseData.yaxis.unit = this.panalUnit
            const chartConfig = {
              title: false,
              eye: false,
              clear: true,
              lineBarSwitch: true,
              chartType: this.panalData.chartType,
              params
            }
            readyToDraw(this,responseData, 1, chartConfig)
          }
        )
      }
    }
  },
  components: {}
}
</script>

<style scoped lang="less">
.zone {
  margin: 0 auto;
  background: @gray-f;
  border-radius: 4px;
}
.zone-chart-title {
  padding: 20px 40%;
  font-size: 14px;
}

.echart-no-data-tip {
  text-align: center;
  vertical-align: middle;
  display: table-cell;
}
</style>

<style scoped lang="less">
.search-container {
  display: flex;
  justify-content: space-between;
  margin: 8px;
  font-size: 16px;
}
.search-zone {
  display: inline-block;
}
.params-title {
  margin-left: 4px;
  font-size: 13px;
}
</style>
