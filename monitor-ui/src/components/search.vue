<template>
  <div class="" style="display:inline-block">
   <ul class="search-ul">
      <li class="search-li">
        <Select
          style="width:300px;"
          v-model="endpoint"
          filterable
          clearable 
          remote
          ref="select"
          :placeholder="$t('requestMoreData')"
          :remote-method="getEndpointList"
          @on-change="updateData"
          >
          <Option v-for="(option, index) in endpointList" :value="option.option_value" :key="index">
            <TagShow :list="endpointList" name="option_type_name" :tagName="option.option_type_name" :index="index"></TagShow> 
            {{option.option_text}}
          </Option>
          <Option value="moreTips" disabled>{{$t('tips.requestMoreData')}}</Option>
        </Select>
      </li>
      <template v-if="!is_mom_yoy">
        <li class="search-li">
          <Select filterable clearable v-model="timeTnterval" :disabled="disableTime" style="width:80px" @on-change="getChartsConfig">
            <Option v-for="item in dataPick" :value="item.value" :key="item.value">{{ item.label }}</Option>
          </Select>
        </li>
        <li class="search-li">
          <Select filterable clearable v-model="autoRefresh" :disabled="disableTime" style="width:100px" @on-change="getChartsConfig" :placeholder="$t('placeholder.refresh')">
            <Option v-for="item in autoRefreshConfig" :value="item.value" :key="item.value">{{ item.label }}</Option>
          </Select>
        </li>
        <li class="search-li">
          <DatePicker type="datetimerange" :value="dateRange" split-panels @on-change="datePick" format="yyyy-MM-dd HH:mm:ss" placement="bottom-start"  :placeholder="$t('placeholder.datePicker')" style="width: 320px"></DatePicker>
        </li>
      </template>
      <template v-else>
        <li class="search-li">
          <DatePicker type="datetimerange" :value="compareFirstDate" split-panels @on-change="pickFirstDate" format="yyyy-MM-dd" placement="bottom-start" :placeholder="$t('placeholder.datePicker')" style="width: 250px"></DatePicker>
        </li>
        <li class="search-li">
          <DatePicker type="datetimerange" :value="compareSecondDate" split-panels @on-change="pickSecondDate" format="yyyy-MM-dd" placement="bottom-start" :placeholder="$t('placeholder.comparedDatePicker')" style="width: 250px"></DatePicker>
        </li>
      </template>
      <li class="search-li">
        <Checkbox v-model="is_mom_yoy" @on-change="YoY">{{$t('button.MoM')}}</Checkbox>
      </li>
      <li class="search-li">
        <button type="button" class="btn btn-sm btn-confirm-f"
          @click="getChartsConfig()">
          <i class="fa fa-search" ></i>
          {{$t('button.search')}}
        </button>
      </li>
      <li class="search-li">
        <button type="button" v-if="isShow&&endpointObject.id !== -1" @click="changeRoute" class="btn btn-sm btn-cancel-f btn-jump">{{$t('button.endpointManagement')}}</button>
      </li>
      <li class="search-li">
        <button type="button" v-if="isShow" @click="historyAlarm" class="btn btn-sm btn-cancel-f btn-jump">{{$t('button.historicalAlert')}}</button>
      </li>
   </ul>
  </div>
</template>

<script>
import {dataPick, autoRefreshConfig} from '@/assets/config/common-config'
import TagShow from '@/components/Tag-show.vue'
export default {
  name: '',
  data() {
    return {
      endpoint: '',
      endpointObject: {},
      endpointList: [],
      ip: {},
      timeTnterval: -1800,
      dataPick: dataPick,
      dateRange: ['', ''],
      compareFirstDate: ['', ''],
      compareSecondDate: ['', ''],
      autoRefresh: 10,
      disableTime: false,
      autoRefreshConfig: autoRefreshConfig,
      is_mom_yoy: false,
      params: {}
    }
  },
  computed: {
    isShow: function () {
      return !this.$root.$validate.isEmpty_reset(this.endpoint)
    }
  },
  watch: {
    endpoint: function (val) {
      if (val) {
        this.endpointObject = this.endpointList.find(ep => {
          return ep.option_value === val
        })
      } else {
        this.endpointObject = {}
      }
    },
    isShow: function () {
      this.clearEndpoint = []
      this.getEndpointList('.')
      this.$parent.showCharts = false 
      this.$parent.showRecursive = false
    }
  },
  mounted() {
    this.getEndpointList('.')
    const jumpCallData = JSON.parse(localStorage.getItem('jumpCallData')) 
    localStorage.removeItem('jumpCallData')
    const outerData = jumpCallData || this.$route.params
    if (!this.$root.$validate.isEmpty_reset(outerData)) {
      const option_value = outerData.option_value
      const option_value_split = option_value.split('_')
      const option_text = option_value_split.slice(0, option_value_split.length - 1).join('_')
      this.endpointList = [{
        active: false,
        id: '',
        option_text: option_text,
        option_type_name: outerData.type,
        option_value: option_value,
        type: outerData.type
      }]
      this.endpoint = option_value
      this.endpointObject = outerData
    }
    if (this.$root.$validate.isEmpty_reset(outerData) && !this.$root.$validate.isEmpty_reset(this.$route.query)) {
      this.endpoint = this.$route.query.endpoint
      this.$root.$store.commit('storeip', {
        id: '',
        option_value: this.$route.query.endpoint,
        type: this.$route.query.type
      })
    }
  },
  methods: {
    getMainConfig () {
      if (this.$root.$validate.isEmpty_reset(this.endpointObject) && !this.$root.$validate.isEmpty_reset(this.$root.$store.state.ip)) {
        this.endpointObject = this.$root.$store.state.ip
        this.$root.$store.commit('storeip', {})
      }
      const type = this.endpointObject.type
      return new Promise(resolve => {
        let params = {
          type: type
        }
        this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.mainConfig.api, params, (responseData) => {
          resolve(responseData)
        })
      })
    },
    datePick (data) {
      this.dateRange = data
      this.disableTime = false
      if (this.dateRange[0] && this.dateRange[1]) {
        if (this.dateRange[0] === this.dateRange[1]) {
          this.dateRange[1] = this.dateRange[1].replace('00:00:00', '23:59:59')
        }
        this.disableTime = true
        this.autoRefresh = 0
      }
      this.getChartsConfig()
    },
    pickFirstDate(data) {
      this.compareFirstDate = data
    },
    pickSecondDate(data) {
      this.compareSecondDate = data
    },
    async getEndpointList(query) {
      let params = {
        search: query,
        page: 1,
        size: 1000
      }
      await this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.resourceSearch.api, params, (responseData) => {
        this.endpointList = responseData
      })
    },
    updateData () {
      this.$nextTick(() => {
        this.getChartsConfig()
      })
    },
    async getChartsConfig () {
      if (this.$root.$validate.isEmpty_reset(this.endpoint)) {
        return
      }
      if (this.$root.$validate.isEmpty_reset(this.endpointObject) && !this.$root.$validate.isEmpty_reset(this.$root.$store.state.ip)) {
        this.endpointObject = this.$root.$store.state.ip
        this.$root.$store.commit('storeip', {})
      }
      if (this.is_mom_yoy) {
        if (this.compareFirstDate[0] === '' || this.compareSecondDate[0] === '') {
          this.$Message.warning(this.$t('tips.selectDate'))
          return
        }
      }
      let params = {}
      if (this.endpointObject.type === 'sys' ) {
        params = {
          autoRefresh: this.autoRefresh,
          time: this.timeTnterval,
          endpoint: this.endpointObject.option_value,
          start: this.dateRange[0] ===''? '':Date.parse(this.dateRange[0].replace(/-/g, '/'))/1000,
          end: this.dateRange[1] ===''? '':Date.parse(this.dateRange[1].replace(/-/g, '/'))/1000,
          guid: this.endpointObject.option_value,
          compare_first_start: this.compareFirstDate[0],
          compare_first_end: this.compareFirstDate[1],
          compare_second_start: this.compareSecondDate[0],
          compare_second_end: this.compareSecondDate[1],
          sys: true
        }
        this.$parent.manageCharts({}, params)
        return
      }
      const res = await this.getMainConfig()
      let url = res.panels.url
      let key = res.search.name
      params = {
        autoRefresh: this.autoRefresh,
        time: this.timeTnterval,
        endpoint: this.endpoint,
        start: this.dateRange[0] ===''? '':Date.parse(this.dateRange[0].replace(/-/g, '/'))/1000,
        end: this.dateRange[1] ===''? '':Date.parse(this.dateRange[1].replace(/-/g, '/'))/1000,
        compare_first_start: this.compareFirstDate[0],
        compare_first_end: this.compareFirstDate[1],
        compare_second_start: this.compareSecondDate[0],
        compare_second_end: this.compareSecondDate[1],
        sys: false
      }
      url = '/monitor/api/v1' + url.replace(`{${key}}`,params[key])
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET',url, params, responseData => {
        this.$parent.manageCharts(responseData, params)
      },{isNeedloading: false})
    },
    YoY(status) {
      this.compareFirstDate = ['', '']
      this.compareSecondDate = ['', '']
      if (status) {
        this.disableTime = true
        this.$root.$eventBus.$emit('clearSingleChartInterval')
        this.autoRefresh = 0
        this.$parent.showCharts = false 
        this.$parent.showRecursive = false
      } else {
        this.disableTime = false
      }
    },
    changeRoute () {
      this.$router.push({name: 'endpointManagement', params: {search: this.endpointObject.option_value}})
    },
    historyAlarm () {
      this.$parent.historyAlarm(this.endpointObject)
    }
  },
  components: {
    TagShow
  }
}
</script>

<style scoped lang="less">
  .search-li {
    display: inline-block;
  }
  .search-ul>li:not(:first-child) {
    padding-left: 12px;
  }
</style>
