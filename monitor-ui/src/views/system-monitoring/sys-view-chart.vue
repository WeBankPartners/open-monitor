<template>
  <div class>
    <header>
      <div style="display:flex;justify-content:space-between; font-size:16px;padding:8px 16px">
        <div class="header-name">
        </div>
        <div class="header-tools">
          <Button type="primary" @click="goBack()">{{$t('m_button_back')}}</Button>
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
          <span>~~~No Data!~~~</span>
        </div>
      </div>
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
import ChartLinesModal from '@/components/chart-lines-modal'
import { generateUuid } from '@/assets/js/utils'
import { readyToDraw } from '@/assets/config/chart-rely'
export default {
  name: '',
  data() {
    return {
      elId: null,
      noDataTip: false,
      panalTitle: '',
      panalUnit: '',
      isLineSelectModalShow: false,
      setChartConfigId: '',
      chartInstance: null,
      request: this.$root.$httpRequestEntrance.httpRequestEntrance,
      apiCenter: this.$root.apiCenter,
    }
  },
  created() {
    generateUuid().then(elId => {
      this.elId = `id_${elId}`
    })
  },
  mounted() {
    if (this.$root.$validate.isEmpty_reset(this.$route.params)) {
      this.$router.push({ path: 'systemMonitoring' })
    } else {
      if (!this.$root.$validate.isEmpty_reset(this.$route.params.templateData)) {
        window['view-config-selected-line-data'] = {}
        this.$on('editShowLines', this.handleEditShowLines)
        this.initPanal()
      }
    }
  },
  methods: {
    initPanal() {
      this.panalTitle = this.$route.params.templateData.panalTitle
      const params = []
      this.noDataTip = false
      this.$route.params.templateData.query.forEach(item => {
        params.push(
          {
            endpoint: item.endpoint,
            metric: item.metricLabel,
            time: '-1800'
          }
        )
      })
      if (params !== []) {
        this.request('POST', this.apiCenter.metricConfigView.api, params, responseData => {
          responseData.chartId = this.elId
          responseData.yaxis.unit = this.panalUnit
          const chartConfig = {
            eye: false,
            clear: true,
            lineBarSwitch: true,
            dataZoom: false,
            chartId: this.elId,
            canEditShowLines: true
          }
          this.chartInstance = readyToDraw(this,responseData, 1, chartConfig)
          if (this.chartInstance) {
            this.chartInstance.on('legendselectchanged', params => {
              window['view-config-selected-line-data'][this.elId] = cloneDeep(params.selected)
            })
          }
        })
      }
    },
    goBack() {
      this.$router.push({
        name: 'systemMonitoring',
        params: this.$route.params.parentData
      })
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
      this.initPanal()
    }
  },
  components: {
    ChartLinesModal
  }
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
