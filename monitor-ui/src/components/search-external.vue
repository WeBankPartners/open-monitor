<template>
  <div class="" style="display:inline-block">
   <ul class="search-ul">
      <li class="search-li">
        <span class="params-title">{{$t('field.endpoint')}}：</span>
        <Tag color="blue">VM_0_16_centos_192.168.0.16_host</Tag>
      </li>
      <li class="search-li">
        <span class="params-title">{{$t('field.relativeTime')}}：</span>
        <Select v-model="timeTnterval" style="width:80px" @on-change="getChartsConfig">
          <Option v-for="item in dataPick" :value="item.value" :key="item.value">{{ item.label }}</Option>
        </Select>
      </li>
      <li class="search-li">
        <span class="params-title">{{$t('field.timeInterval')}}：</span>
        <DatePicker type="daterange" placement="bottom-end" @on-change="datePick" :placeholder="$t('placeholder.datePicker')" style="width: 200px"></DatePicker>
      </li>
      <li class="search-li">
        <span class="params-title">{{$t('placeholder.refresh')}}：</span>
        <Select v-model="autoRefresh" style="width:100px" @on-change="getChartsConfig" :placeholder="$t('placeholder.refresh')">
          <Option v-for="item in autoRefreshConfig" :value="item.value" :key="item.value">{{ item.label }}</Option>
        </Select>
      </li>
   </ul>
  </div>
</template>

<script>
import {cookies} from '@/assets/js/cookieUtils'
import {dataPick, autoRefreshConfig} from '@/assets/config/common-config'
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
      dateRange: ['',''],
      autoRefresh: 0,
      autoRefreshConfig: autoRefreshConfig,
      params: {
        // time: this.timeTnterval,
        // group: 1,
        // endpoint: '192.168.0.16',
        // start: Date.parse(this.dateRange[0]),
        // end: Date.parse(this.dateRange[1])
      }
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
    }
  },
  mounted() {
    if (this.$root.$validate.isEmpty_reset(this.$route.params) && !this.$root.$validate.isEmpty_reset(this.$route.query)) {
      this.endpoint = this.$route.query.endpoint
      cookies.setAuthorization(`${this.$route.query.token}`)
      this.setLocale(this.$route.query.lang)
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
      if (this.dateRange[0] !== '') {
        this.dateRange[0] = this.dateRange[0] + ' 00:00:01'
      }
      if (this.dateRange[1] !== '') {
        this.dateRange[1] = this.dateRange[1] + ' 23:59:59'
      }
      this.getChartsConfig()
    },
    async getChartsConfig () {
      if (this.$root.$validate.isEmpty_reset(this.endpoint)) {
        return
      }
      if (this.$root.$validate.isEmpty_reset(this.endpointObject) && !this.$root.$validate.isEmpty_reset(this.$root.$store.state.ip)) {
        this.endpointObject = this.$root.$store.state.ip
        this.$root.$store.commit('storeip', {})
      }
      let params = {}
      if (this.endpointObject.type === 'sys' ) {
        params = {
          autoRefresh: this.autoRefresh,
          time: this.timeTnterval,
          endpoint: this.endpointObject.option_value,
          start: this.dateRange[0] ===''? '':Date.parse(this.dateRange[0])/1000,
          end: this.dateRange[1] ===''? '':Date.parse(this.dateRange[1])/1000,
          guid: this.endpointObject.option_value,
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
        start: this.dateRange[0] ===''? '':Date.parse(this.dateRange[0])/1000,
        end: this.dateRange[1] ===''? '':Date.parse(this.dateRange[1])/1000,
        sys: false
      }
      url = url.replace(`{${key}}`,params[key])
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET',url, params, responseData => {
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
    padding-left: 10px;
  }
  .params-title {
    font-size: 13px;
  }
</style>
