<template>
  <div class=" ">
    <div style="margin-bottom:24px;">
        <Select v-model="metricSelected" multiple style="width:260px" :label-in-value="true" 
            @on-change="selectMetric" placeholder="请选择监控指标">
            <Option v-for="item in metricList" :value="item.prom_ql" :key="item.metric">{{ item.metric }}</Option>
        </Select>
        <Select v-model="timeTnterval" style="width:80px">
          <Option v-for="item in dataPick" :value="item.value" :key="item.value">{{ item.label }}</Option>
        </Select>
        <button class="btn btn-sm btn-cancle-f" @click="addMetric">新增</button>
        <button class="btn btn-sm btn-cancle-f" @click="requestChart">查询</button>
        <button class="btn btn-sm btn-cancle-f" @click="saveConfig">保存</button>
    </div>
    
    <section>
      <ul>
        <template v-for="(metricItem, metricIndex) in totalMetric">
          <li :key="metricIndex" class="metric-display">
            <Tag color="primary" type="border" closable @on-close="delMetric(metricItem)">指标名称：{{metricItem.label}}</Tag>
            <div>
               <textarea v-model="metricItem.value" class="textareaSty"></textarea> 
            </div>
          </li>
        </template>
      </ul>
    </section>
  </div>
</template>

<script>
import {dataPick} from '@/assets/config/common-config'
export default {
  name: '',
  data() {
    return {
     metricSelected: [],
     metricSelectedOptions: [],
     metricList: [],

     timeTnterval: -1800,
     dataPick: dataPick,

     editMetric: []
    }
  },
  computed: {
     totalMetric: function () {
       return this.metricSelectedOptions.concat(this.editMetric)
     } 
  },
  mounted (){
    this.obtainMetricList()
  },
  methods: {
    addMetric() {
     this.editMetric.push({label: `default${Math.round(new Date() / 1000)}`, value: ''})
    },
    delMetric (metric) {
      if (metric.label.indexOf('default') > -1) {
       this.editMetric =  this.editMetric.filter((item)=>{
           return item.label !== metric.label
        })
      } else {
        this.metricSelectedOptions = this.metricSelectedOptions.filter((item)=>{
           return item.label !== metric.label
        })
        this.metricSelected = this.metricSelected.filter((item)=>{
           return item !== metric.value
        })
      }
    },
    selectMetric (option) {
      this.metricSelectedOptions = option
    },
    obtainMetricList () {
      let params = {type: this.$store.state.ip.type}
      this.$httpRequestEntrance.httpRequestEntrance('GET',this.apiCenter.metricList.api, params, responseData => {
        this.metricList = responseData
      })
    },
    requestChart () {
      if (this.$validate.isEmpty_reset(this.totalMetric)) {
        this.$Message.warning('请先设置监控指标!')
        return
      }
      let params = []
      this.totalMetric.forEach((item) => {
        params.push(JSON.stringify({
          endpoint: this.$store.state.ip.value.split(':')[0],
          prom_ql: item.value,
          time: this.timeTnterval + ''
        })) 
      })
      this.$httpRequestEntrance.httpRequestEntrance('GET',this.apiCenter.metricConfigView.api, {config: `[${params.join('')}]`}, responseData => {
        console.log(responseData)
      })

    },
    saveConfig () {
      console.log(234)
    }
  },
  components: {},
}
</script>

<style scoped lang="less">
  textarea:focus {
    outline: none;
  }
  .metric-display {
    margin: 16px;
  }
  .textareaSty{
    display: inline-block;
    vertical-align: top;
    width: 70%;
    border-radius: 4px;
    border-color: #dddee1;
    height: 100px;
    padding: 3px;
  }
</style>
