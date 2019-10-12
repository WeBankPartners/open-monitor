<template>
  <div class=" ">
      <header>
        <div style="display:flex;justify-content:space-between">
            <div class="header-name">
              <i class="fa fa-th-large fa-2x" aria-hidden="true"></i>
              <span> 12123123</span>
            </div>
            <div class="header-tools"> 
              <i class="fa fa-floppy-o fa-2x" aria-hidden="true"></i>
            </div>
        </div>
      </header>
      <div class="zone zone-chart" >
        <div :id="elId" class="echart"></div>
      </div>
      <div class="zone zone-config" >
        <div style="display:flex">
          <section>
            <ul>
              <li>
                <Tooltip content="指标配置" placement="bottom">
                  <div class="step-icon" @click="activeStep='chat_query'">
                    <i class="fa fa-line-chart" aria-hidden="true"></i>
                  </div>
                </Tooltip>
              </li>
              <li>
                <div class="step-link"></div>
              </li>
              <li>
                <Tooltip content="全局配置" placement="bottom">
                  <div class="step-icon" @click="activeStep='chat_general'">
                    <i class="fa fa-cog" aria-hidden="true"></i>
                  </div>
                </Tooltip>
              </li>
            </ul>
          </section>
          <section class="zone-config-operation">
            <button class="btn btn-sm btn-cancle-f" >新增指标</button>
            <button class="btn btn-sm btn-cancle-f" >保存配置</button>
            <div v-if="activeStep==='chat_query'">
              <template v-for="(queryItem,queryIndex) in chartQueryList">
                <div class="condition-zone" :key="queryIndex">
                  <ul>
                    <li>
                      <div class="condition condition-title">对象</div>
                      <div class="condition">
                        <Select
                          style="width:300px"
                          v-model="chartQueryList[queryIndex].entpointModel"
                          filterable
                          remote
                          :remote-method="entpointList">
                          <Option v-for="(option, index) in options" :value="option.option_value" :key="index">{{option.option_text}}</Option>
                        </Select>
                      </div>
                    </li>
                    <li>
                      <div class="condition condition-title">指标</div>
                      <div class="condition">
                        <Select v-model="chartQueryList[queryIndex].metricModel" style="width:300px" @on-open-change="metricSelectOpen(queryItem.entpointModel)">
                          <Option v-for="(item,index) in metricList" :value="item.prom_ql" :key="item.prom_ql+index">{{ item.metric }}</Option>
                        </Select>
                      </div>
                    </li>
                  </ul>
                </div>
              </template>
                <!-- <div class="condition-zone" >
                  <ul>
                    <li>
                      <div class="condition condition-title">主机</div>
                      <div class="condition">
                        <Select
                          style="width:300px"
                          v-model="model"
                          filterable
                          remote
                          :remote-method="entpointList">
                          <Option v-for="(option, index) in options" :value="option.option_value" :key="index">
                          {{option.option_text}}</Option>
                        </Select>
                      </div>
                    </li>
                  </ul>
                </div> -->
            </div>
            <div v-if="activeStep==='chat_general'">
              2
            </div>
          </section>
        </div>
        
      </div>
  </div>
</template>

<script>
import {generateUuid} from '@/assets/js/utils'
import {drawChart} from  '@/assets/config/chart-rely'
export default {
  name: '',
  data() {
    return {
      elId: null,
      activeStep: 'chat_query',
      initQuery: {
        entpointModel: '',
        metricList: [],
        metricModel: ''
      },
      chartQueryList:[{
        entpointModel: '',
        metricModel: ''
      }],
      
      options: [],
      metricList: []
      // model: ''
    }
  },
  created (){
    generateUuid().then((elId)=>{
      this.elId =  `id_${elId}`; 
    })
  },
  mounted() {
    this.chartData()
    // this.entpointList()
  },
  methods: {
    entpointList(query) {
      let params = {
        search: query,
        page: 1,
        size: 100
      }
      this.$httpRequestEntrance.httpRequestEntrance('GET', this.apiCenter.resourceSearch.api, params, (responseData) => {
       this.options = responseData
      })
    },
    metricSelectOpen(metricModel) {
      if (this.$validate.isEmpty_reset(metricModel)) {
        this.$Message.warning('请先选择主机！')
      } else {
        let params = {type: metricModel.split(':')[1]}
        this.$httpRequestEntrance.httpRequestEntrance('GET',this.apiCenter.metricList.api, params, responseData => {
          this.metricList = responseData
        })
      } 
    },
    chartData () {
        let params = {
            agg: 'none',
            endpoint: ['VM_0_14_centos_192.168.0.14_host'],
            id: 1,
            metric: ['cpu.used.percent'],
            time: '-1800'
        }
        this.$httpRequestEntrance.httpRequestEntrance('GET', '/dashboard/chart', params, responseData => {
            var legend = []
            responseData.series.forEach((item)=>{
            legend.push(item.name)
            item.symbol = 'none'
            item.smooth = true
            item.lineStyle = {
                width: 1
            }
            }) 
            let config = {
            title: responseData.title,
            legend: legend,
            series: responseData.series,
            yaxis: responseData.yaxis,
            }
            drawChart(this,config)
        })
    }
  },
  components: {},
}
</script>

<style scoped lang="less">
.zone {
    width:1100px;
    margin: 0 auto;
    background: @gray-f;
    border-radius: 4px;
}
.zone-chart {
    margin-top: 16px;
    margin-bottom: 16px;

}
.zone-config {
  padding: 8px;
}
.echart {
    height:400px;
    width:1100px;
}

.step-icon {
    i {
        height: 24px;
        width: 24px;
        font-size: 18px;
        color: @blue-lingt;
    }
    .fa-line-chart {
        margin: 7px 6px;
    }
    .fa-cog {
        margin: 8px;
    }
    width: 36px;
    height: 36px;
    border: 2px solid @blue-lingt;
    border-radius: 18px;
    cursor: pointer;
}
.step-link {
    height:36px;
    border-left:2px solid @blue-lingt;
    margin-left:16px;
}

.zone-config-operation {
  margin-left: 24px;
}
</style>

<style scoped lang="less">

  .condition {
    margin: 2px;
    display: inline-block;
  }
  .condition-title {
    background: @gray-d;
    width: 100px;
    text-align: center;
    vertical-align: middle;
    margin-right: 8px;
    padding: 6px;
  }
  .condition-zone {
    border: 1px solid @blue-2;
    padding: 4px;
    margin: 4px;
  }
</style>

