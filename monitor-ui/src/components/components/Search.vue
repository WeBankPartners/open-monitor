<template>
  <div class="" style="display:inline-block">
   <ul class="search-ul">
      <li class="search-li">
        <Searchinput ref="searchInput" :parentConfig="searchInputConfig"></Searchinput> 
      </li>
      <li class="search-li">
        <button type="button" class="btn btn-sm btn-confirm-f"
            @click="getChartsConfig()">
            <i class="fa fa-search" ></i>
            搜索
          </button>
      </li>
      <li class="search-li">
          <Select v-model="timeTnterval" style="width:80px" @on-change="getChartsConfig">
            <Option v-for="item in dataPick" :value="item.value" :key="item.value">{{ item.label }}</Option>
          </Select>
      </li>
      <li class="search-li">
        <DatePicker type="daterange" placement="bottom-end" @on-change="datePick" placeholder="请选择日期" style="width: 200px"></DatePicker>
      </li>
      <li class="search-li">
        <Select v-model="autoRefresh" style="width:100px" @on-change="getChartsConfig" placeholder="定时刷新">
          <Option v-for="item in autoRefreshConfig" :value="item.value" :key="item.value">{{ item.label }}</Option>
        </Select>
      </li>
   </ul>
  </div>
</template>

<script>
import {dataPick, autoRefreshConfig} from '@/assets/config/common-config'
import Searchinput from './Search-input'
export default {
  name: '',
  data() {
    return {
      searchInputConfig: {
        poptipWidth: 300,
        placeholder: '请输入主机名或IP地址，可模糊匹配',
        inputStyle: "width:300px;",
        api: this.apiCenter.resourceSearch.api
      },
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
  methods: {
    getMainConfig () {
      return new Promise(resolve => {
        let params = {
          type: this.ip.value.split(':')[1]
        }
        this.$httpRequestEntrance.httpRequestEntrance('GET', this.apiCenter.mainConfig.api, params, (responseData) => {
            resolve(responseData)
          })
        })
    },
    datePick (data) {
      this.dateRange = data
      this.getChartsConfig()
    },
    getChartsConfig () {
      if (Object.keys(this.$store.state.ip).length === 0) {
        this.$Message.warning('请选择有效IP！')
        return
      } else {
        this.ip = this.$store.state.ip
      }
      this.getMainConfig().then((res)=>{
        let url = res.panels.url
        let key = res.search.name
        let params = {
          autoRefresh: this.autoRefresh,
          time: this.timeTnterval,
          endpoint: this.ip.value,
          start: this.dateRange[0] ===''? '':Date.parse(this.dateRange[0])/1000,
          end: this.dateRange[1] ===''? '':Date.parse(this.dateRange[1])/1000
        }
        url = url.replace(`{${key}}`,params[key].split(':')[0])
        this.$httpRequestEntrance.httpRequestEntrance('GET',url, params, responseData => {
          this.$parent.manageCharts(responseData, params)
        },{isNeedloading: false})
      })
      
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
