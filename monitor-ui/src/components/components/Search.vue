<template>
  <div class=" ">
   <ul class="search-ul">
      <li class="search-li">
        <Searchinput @sendInputValue="obtainInputValue"></Searchinput> 
      </li>
      <li class="search-li">
        <Button type="primary" @click="getChartsConfig" icon="ios-search">搜索</Button>
      </li>
      <li class="search-li">
          <Select v-model="timeTnterval" style="width:80px" @on-change="getChartsConfig">
            <Option v-for="item in dataPick" :value="item.value" :key="item.value">{{ item.label }}</Option>
          </Select>
      </li>
      <li class="search-li">
        <DatePicker type="daterange" placement="bottom-end" @on-change="datePick" placeholder="请选择日期" style="width: 200px"></DatePicker>
      </li>
   </ul>
  </div>
</template>

<script>
import Searchinput from './Search-input'
export default {
  name: '',
  data() {
    return {
      ip: {},
      timeTnterval: -1800,
      dataPick: [
        {
            value: -1800,
            label: '30分钟'
        },
        {
            value: -3600,
            label: '1小时'
        },
        {
            value: -10800,
            label: '3小时'
        }
      ],
      dateRange: ['',''],
      params: {
        // time: this.timeTnterval,
        // group: 1,
        // endpoint: '192.168.0.16',
        // start: Date.parse(this.dateRange[0]),
        // end: Date.parse(this.dateRange[1])
      }
    }
  },
  mounted (){
  },
  
  methods: {
    datePick (data) {
      this.dateRange = data
      this.getChartsConfig()
    },
    getChartsConfig (ip=this.ip) {
      let params = {
        group: 1,
        time: this.timeTnterval,
        endpoint: ip.value,
        start: this.dateRange[0] ===''? '':Date.parse(this.dateRange[0])/1000,
        end: this.dateRange[1] ===''? '':Date.parse(this.dateRange[1])/1000
      }
      this.$httpRequestEntrance.httpRequestEntrance('GET','/dashboard/panels', params, responseData => {
        this.$parent.manageCharts(responseData, params)
      },{isNeedloading: false})
    },
    request () {
      if (!this.ip_name) {
        return
      }
      let params = {
        search: this.ip_name
      }
      this.$httpRequestEntrance.httpRequestEntrance('GET','/dashboard/search', params, responseData => {
        this.searchResult = responseData
      })
      this.showSearchTips = true
    },
    obtainInputValue (ip) {
      this.ip = ip
      this.getChartsConfig()
    } 
  },
  components: {
    Searchinput
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
</style>
