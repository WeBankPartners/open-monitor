<template>
  <div>
    <div class="select-tag-radio">
      <RadioGroup v-model="selectedType"
                  type="button"
                  size="small"
                  @on-change="onSelectedTypeChange"
      >
        <Radio v-for="(tag, idx) in allTypeList"
               :label="tag"
               :key="idx"
               :disabled="endpointExternal"
        >
          {{ tag }}
        </Radio>
      </RadioGroup>
    </div>

    <div style="display:inline-block">
      <ul class="search-ul">
        <li class="search-li">
          <Select
            style="width:300px;"
            v-model="endpoint"
            filterable
            clearable
            remote
            ref="select"
            :disabled="endpointExternal"
            :placeholder="$t('m_requestMoreData')"
            @on-query-change="debounceGetEndpointList"
            @on-change="updateData"
          >
            <Option v-for="(option, index) in endpointList" :value="option.option_value" :label="option.option_text" :key="index">
              <TagShow :list="endpointList" name="option_type_name" :tagName="option.option_type_name" :index="index"></TagShow>
              {{option.option_text}}
            </Option>
            <Option value="moreTips" disabled>{{$t('m_tips_requestMoreData')}}</Option>
          </Select>
        </li>
        <template v-if="!is_mom_yoy">
          <li class="search-li">
            <Select filterable v-model="timeTnterval" :disabled="disableTime" style="width:80px" @on-change="getChartsConfig()">
              <Option v-for="item in dataPick" :value="item.value" :key="item.value">{{ item.label }}</Option>
            </Select>
          </li>
          <li class="search-li">
            <Select filterable v-model="autoRefresh" :disabled="disableTime" style="width:100px" @on-change="getChartsConfig()" :placeholder="$t('m_placeholder_refresh')">
              <Option v-for="item in autoRefreshConfig" :value="item.value" :key="item.value">{{ item.label }}</Option>
            </Select>
          </li>
          <li class="search-li view-pannal-date-picker">
            <DatePicker
              type="datetimerange"
              :value="dateRange"
              split-panels
              @on-change="datePick"
              @on-ok="onDatePickOk"
              @on-open-change="onDatePickChange"
              format="yyyy-MM-dd HH:mm:ss"
              placement="bottom-start"
              :placeholder="$t('m_placeholder_datePicker')"
              style="width: 320px"
            ></DatePicker>
          </li>
        </template>
        <template v-else>
          <li class="search-li">
            <DatePicker type="datetimerange" :value="compareFirstDate" split-panels @on-change="pickFirstDate" format="yyyy-MM-dd" placement="bottom-start" :placeholder="$t('m_placeholder_datePicker')" style="width: 250px"></DatePicker>
          </li>
          <li class="search-li">
            <DatePicker type="datetimerange" :value="compareSecondDate" split-panels @on-change="pickSecondDate" format="yyyy-MM-dd" placement="bottom-start" :placeholder="$t('m_placeholder_comparedDatePicker')" style="width: 250px"></DatePicker>
          </li>
        </template>
        <li class="search-li">
          <Checkbox v-model="is_mom_yoy" @on-change="YoY">{{$t('m_button_MoM')}}</Checkbox>
        </li>
        <li class="search-li">
          <Button type="primary" @click="getChartsConfig()">{{$t('m_button_search')}}</Button>
        </li>
        <li class="search-li">
          <Button v-if="isShow && endpointObject && endpointObject.id !== -1 && !endpointExternal" @click="changeRoute">{{$t('m_button_endpointManagement')}}</Button>
        </li>
        <li class="search-li">
          <Button v-if="isShow && !endpointExternal" @click="historyAlarm">{{$t('m_button_historicalAlert')}}</Button>
        </li>
      </ul>
    </div>
  </div>
</template>

<script>
import {find, isEmpty, debounce} from 'lodash'
import {dataPick, autoRefreshConfig} from '@/assets/config/common-config'
import TagShow from '@/components/Tag-show.vue'

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
      compareFirstDate: ['', ''],
      compareSecondDate: ['', ''],
      autoRefresh: 10,
      disableTime: false,
      autoRefreshConfig,
      is_mom_yoy: false,
      params: {},
      endpointExternal: false,
      request: this.$root.$httpRequestEntrance.httpRequestEntrance,
      apiCenter: this.$root.apiCenter,
      selectedType: 'all',
      allTypeList: []
    }
  },
  computed: {
    isShow() {
      return !this.$root.$validate.isEmpty_reset(this.endpoint)
    }
  },
  watch: {
    endpoint(val) {
      if (val) {
        this.endpointObject = this.endpointList.find(ep => ep.option_value === val)
      } else {
        this.endpointObject = {}
      }
    },
    async isShow(val) {
      if (!val) {
        this.clearEndpoint = []
        this.$parent.showCharts = false
        this.$parent.showRecursive = false
        await this.getEndpointList('.')
      }
    }
  },
  async mounted() {
    await this.getEndpointList('.')
    this.getAllSelectedTypeList()
    const jumpCallData = JSON.parse(localStorage.getItem('jumpCallData'))
    localStorage.removeItem('jumpCallData')
    let outerData = jumpCallData || this.$route.params
    if (this.$route.name === 'viewConfig') {
      outerData = null
    }
    if (!this.$root.$validate.isEmpty_reset(outerData)) {
      this.endpointObject = find(this.endpointList, {option_text: outerData.option_name})
      this.endpoint = this.endpointObject.option_value
    }
    if (this.$root.$validate.isEmpty_reset(outerData) && !this.$root.$validate.isEmpty_reset(this.$route.query)) {
      this.endpoint = this.$route.query.endpoint || ''
      this.$root.$store.commit('storeip', {
        id: '',
        option_value: this.$route.query.endpoint,
        type: this.$route.query.type
      })
    }
    this.updateData()
  },
  methods: {
    getMainConfig() {
      if (this.$root.$validate.isEmpty_reset(this.endpointObject) && !this.$root.$validate.isEmpty_reset(this.$root.$store.state.ip)) {
        this.endpointObject = this.$root.$store.state.ip
        this.$root.$store.commit('storeip', {})
      }
      const type = this.endpointObject.type
      return new Promise(resolve => {
        const params = {
          type
        }
        this.request('GET', this.apiCenter.mainConfig.api, params, responseData => {
          resolve(responseData)
        })
      })
    },
    datePick(data) {
      if (data[0] === this.dateRange[0] && data[1] === this.dateRange[1]) {
        return
      }
      this.dateRange = data
      this.disableTime = false
      if (this.dateRange[0] && this.dateRange[1]) {
        this.dateRange[1] = this.dateRange[1].replace('00:00:00', '23:59:59')
        this.disableTime = true
        this.autoRefresh = 0
      }
      this.getChartsConfig()
    },
    onDatePickChange(flag) {
      if (!flag) {
        this.onDatePickOk()
      }
    },
    onDatePickOk() {
      const datePickDomList = document.querySelectorAll('.view-pannal-date-picker .ivu-input.ivu-input-default.ivu-input-with-suffix')
      const valStr = datePickDomList[0].value
      if (valStr) {
        const matches = valStr.match(/\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}/g)
        if (matches.length === 2) {
          this.datePick(matches)
        }
      }
    },
    pickFirstDate(data) {
      this.compareFirstDate = data
    },
    pickSecondDate(data) {
      this.compareSecondDate = data
    },
    debounceGetEndpointList: debounce(function (query) {
      if (query) {
        this.getEndpointList(query)
      }
    }, 500),
    async getEndpointList(query = '.') {
      return new Promise(async resolve => {
        const params = {
          search: query,
          page: 1,
          size: 1000,
          optionTypeName: this.selectedType
        }
        await this.request('GET', this.apiCenter.resourceSearch.api, params, responseData => {
          this.endpointList = responseData
          resolve(responseData)
        })
      })

    },
    updateData() {
      this.$nextTick(() => {
        this.getChartsConfig()
      })
    },
    disabledEndpoint(val) {
      this.endpointExternal = val
    },
    async getChartsConfig(endpointObj) {
      if (!isEmpty(endpointObj) && endpointObj.option_name) {
        await this.getEndpointList()
        this.endpointObject = find(this.endpointList, {option_text: endpointObj.option_name})
        this.endpoint = this.endpointObject.option_value
      }
      if (this.$root.$validate.isEmpty_reset(this.endpoint)) {
        return
      }
      if (this.$root.$validate.isEmpty_reset(this.endpointObject) && !this.$root.$validate.isEmpty_reset(this.$root.$store.state.ip)) {
        this.endpointObject = this.$root.$store.state.ip
        this.$root.$store.commit('storeip', {})
      }
      if (this.is_mom_yoy) {
        if (this.compareFirstDate[0] === '' || this.compareSecondDate[0] === '') {
          this.$Message.warning(this.$t('m_tips_selectDate'))
          return
        }
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
      const key = res.search.name
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
      this.request('GET',url, params, responseData => {
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
    changeRoute() {
      this.$router.push({
        name: 'endpointManagement',
        params: {search: this.endpointObject.option_value}
      })
    },
    historyAlarm() {
      this.$parent.historyAlarm(this.endpointObject)
    },
    onSelectedTypeChange() {
      this.getEndpointList('.')
    },
    getAllSelectedTypeList() {
      this.request('GET', this.apiCenter.getOptionTypeNameList, {}, responseData => {
        this.allTypeList = responseData
      })
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
  .select-tag-radio {
    margin-top: 5px;
    margin-bottom: 15px;
  }
</style>
<style lang="less">
.select-tag-radio {
  .ivu-radio-group {
    display: flex;
    overflow-x: auto;
    max-width: 98vw;
    overflow-y: hidden;
  }
}

</style>
