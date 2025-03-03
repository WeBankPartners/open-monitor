<template>
  <div class="" style="display:inline-block">
    <ul class="search-ul">
      <li class="search-li">
        <span class="params-title">{{$t('m_field_endpoint')}}：</span>
        <Tag color="blue">{{endpointObject.option_value}}</Tag>
      </li>
      <li class="search-li">
        <span class="params-title">{{$t('m_field_relativeTime')}}：</span>
        <Select filterable v-model="timeTnterval" :disabled="disableTime" style="width:80px" @on-change="getChartsConfig">
          <Option v-for="item in dataPick" :value="item.value" :key="item.value">{{ item.label }}</Option>
        </Select>
      </li>
      <li class="search-li">
        <span class="params-title">{{$t('m_placeholder_refresh')}}：</span>
        <Select filterable clearable v-model="autoRefresh" :disabled="disableTime" style="width:100px" @on-change="getChartsConfig" :placeholder="$t('m_placeholder_refresh')">
          <Option v-for="item in autoRefreshConfig" :value="item.value" :key="item.value">{{ item.label }}</Option>
        </Select>
      </li>
      <li class="search-li"  style="margin-left:20px">
        <span class="params-title">{{$t('m_field_timeInterval')}}：</span>
        <DatePicker type="datetimerange" :value="dateRange" split-panels format="yyyy-MM-dd HH:mm:ss" placement="bottom-start" @on-change="datePick" :placeholder="$t('m_placeholder_datePicker')" style="width: 320px"></DatePicker>
      </li>
    </ul>
  </div>
</template>

<script>
import { setToken} from '@/assets/js/cookies.ts'
import {dataPick, autoRefreshConfig} from '@/assets/config/common-config'

export const custom_api_enum = [
  {
    getDashboardPanels: 'get'
  }
]

export default {
  name: '',
  data() {
    return {
      endpoint: '',
      endpointObject: {},
      endpointList: [],
      ip: {},
      timeTnterval: -1800,
      dataPick,
      dateRange: ['', ''],
      autoRefresh: 10,
      disableTime: false,
      autoRefreshConfig,
      params: {
        // time: this.timeTnterval,
        // group: 1,
        // endpoint: '192.168.0.16',
        // start: Date.parse(this.dateRange[0]),
        // end: Date.parse(this.dateRange[1])
      },
      request: this.$root.$httpRequestEntrance.httpRequestEntrance,
      apiCenter: this.$root.apiCenter,
    }
  },
  watch: {
    // endpoint: function (val) {
    //   if (val) {
    //     this.endpointObject = this.endpointList.find(ep => {
    //       return ep.option_value === val
    //     })
    //   } else {
    //     this.endpointObject = {}
    //   }
    // }
  },
  mounted() {
    if (this.$root.$validate.isEmpty_reset(this.$route.params) && !this.$root.$validate.isEmpty_reset(this.$route.query)) {
      this.endpoint = this.$route.query.endpoint
      setToken(`${this.$route.query.token}`)
      this.setLocale(this.$route.query.lang)
      this.dateRange = [this.$route.query.startTime,this.$route.query.endTime]
      this.disableTime = true
      this.endpointObject = {
        id: '',
        option_value: this.$route.query.endpoint,
        type: this.$route.query.type
      }
      this.$root.$store.commit('storeip', this.endpointObject)
      this.getMainConfig()
    }
  },
  methods: {
    getMainConfig() {
      if (this.$root.$validate.isEmpty_reset(this.endpointObject) && !this.$root.$validate.isEmpty_reset(this.$root.$store.state.ip)) {
        this.endpointObject = this.$root.$store.state.ip
        this.$root.$store.commit('storeip', {})
      }
      this.getChartsConfig()
    },
    datePick(data) {
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
    async getChartsConfig() {
      if (this.$root.$validate.isEmpty_reset(this.endpoint)) {
        return
      }
      if (this.$root.$validate.isEmpty_reset(this.endpointObject) && !this.$root.$validate.isEmpty_reset(this.$root.$store.state.ip)) {
        this.endpointObject = this.$root.$store.state.ip
        this.$root.$store.commit('storeip', {})
      }
      let params = {}
      if (this.endpointObject.type === 'sys') {
        params = {
          autoRefresh: this.autoRefresh,
          time: this.timeTnterval,
          endpoint: this.endpointObject.option_value,
          start: this.dateRange[0] ===''? '':Date.parse(this.dateRange[0].replace(/-/g, '/'))/1000,
          end: this.dateRange[1] ===''? '':Date.parse(this.dateRange[1].replace(/-/g, '/'))/1000,
          guid: this.endpointObject.option_value,
          sys: true
        }
        this.$parent.manageCharts({}, params)
        return
      }
      const res = await this.getMainConfig()
      let url = res.panels.url
      const key = res.search.name
      params = {
        autoRefresh: this.autoRefresh,
        time: this.timeTnterval,
        endpoint: this.endpoint,
        start: this.dateRange[0] ===''? '':Date.parse(this.dateRange[0].replace(/-/g, '/'))/1000,
        end: this.dateRange[1] ===''? '':Date.parse(this.dateRange[1].replace(/-/g, '/'))/1000,
        sys: false
      }
      url = url.replace(`{${key}}`,params[key])
      this.request('GET',url, params, responseData => {
        this.$parent.manageCharts(responseData, params)
      },{isNeedloading: false})
    },
    setLocale(lang) {
      localStorage.setItem('lang', lang)
      this.$i18n.locale = lang
      this.$validator.locale = lang
    },
  },
  components: {
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
  .params-title {
    font-size: 13px;
  }
</style>
