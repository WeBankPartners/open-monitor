<template>
  <div class=" ">
    <section>
      <ul class="search-ul">
        <li class="search-li">
          <Select v-model="type" style="width:100px" @on-change="typeChange">
            <Option v-for="item in typeList" :value="item.value" :key="item.value">{{ $t(item.label) }}</Option>
          </Select>
        </li>
        <li class="search-li">
          <Select
            style="width:300px"
            v-model="targrtId"
            filterable
            clearable
            ref="select"
            @on-open-change="getTargrtList"
            @on-clear="clearTargrt"
            >
            <Option v-for="(option, index) in targetOptions" :value="option.guid" :key="index">{{option.display_name}}
            </Option>
          </Select>
        </li>
        <li class="search-li">
          <button type="button" class="btn btn-sm btn-confirm-f"
          :disabled="targrtId === ''"
          @click="search">
            <i class="fa fa-search" ></i>
            {{$t('button.search')}}
          </button>
        </li>
        <li class="search-li" style="cursor: pointer;">
          <span @click="openDoc">
            <i 
              class="fa fa-book" 
              aria-hidden="true" 
              style="font-size:20px;color:#58a0e6;vertical-align: middle;margin-left:20px">
            </i>
            {{$t('operationDoc')}}
          </span>
        </li>
      </ul>
    </section> 
    <section v-show="showTargetManagement" style="margin-top: 16px;">
      <template v-if="type === 'group'">
        <groupManagement ref="group"></groupManagement>
      </template>
      <template v-if="type === 'endpoint'">
        <endpointManagement ref="endpoint"></endpointManagement>
      </template>
    </section>
    <!-- <Input v-model="aaa" />
    <Input v-model="bbb" type="textarea" />
    <span v-html="xxx">

    </span>
    <Button @click="test">我有一只小</Button> -->
  </div>
</template>

<script>
import endpointManagement from './business-monitor-endpoint.vue'
import groupManagement from './business-monitor-group.vue'
export default {
  name: '',
  data() {
    return {
      aaa: `.*{"code":"./ams/account/query_user_info","biz":".*?"`,
      bbb: 'INFO|2021-12-01 14:56:56382|http-nio-18000-exec-4|cn.webank.weden.utils.MonitorLogUtils:53|[][][CGI][{"code":"./ams/account/query_user_info","biz":"21120112D08048GSL0000Xd6G2NIgOVx","service":"","sceneid":"","cost_t":"292","ret":"80480000","msg":"SUCC","point":"CGI","uid":"876CD9A0CBA5A7257DEC83E3C79E60C9B6047CB959D9D85045AC543242A5C768"}][]',
      xxx: '',
      type: 'endpoint',
      typeList: [
        {label: this.$t('tableKey.endpoint'), value: 'endpoint'},
        {label: this.$t('field.resourceLevel'), value: 'group'}
      ],
      targrtId: '',
      targetOptions: [],
      showTargetManagement: false
    }
  },
  
  async mounted () {
   
  },
  beforeDestroy () {
    this.$root.$store.commit('changeTableExtendActive', -1)
  },
  methods: {
    test () {
      // var str = 'INFO|2021-12-01 14:56:56382|http-nio-18000-exec-4|cn.webank.weden.utils.MonitorLogUtils:53|[][][CGI][{"code":"./ams/account/query_user_info","biz":"21120112D08048GSL0000Xd6G2NIgOVx","service":"","sceneid":"","cost_t":"292","ret":"80480000","msg":"SUCC","point":"CGI","uid":"876CD9A0CBA5A7257DEC83E3C79E60C9B6047CB959D9D85045AC543242A5C768"}][]';
      // eslint-disable-next-line no-useless-escape
      const reg = new RegExp(this.aaa, 'g')
      // str = str.replace(reg, "<span style='color:red'>" + this.filterParams + '</span>')
      // console.log(str);

      let execRes = this.bbb.match(reg)
      this.xxx = this.bbb.replace(reg, "<span style='color:red'>" + execRes[0] + '</span>')
      console.log(this.xxx)
      const params = {
        reg_string: this.aaa,
        test_context: this.bbb
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('POST', '/monitor/api/v2/regexp/test/match', params, (responseData) => {
        console.log(responseData)
      }, {isNeedloading:false})
    },
    typeChange () {
      this.clearTargrt()
    },
    getTargrtList () {
      const api = this.$root.apiCenter.getTargetByEndpoint + '/' + this.type
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', api, '', (responseData) => {
        this.targetOptions = responseData
      }, {isNeedloading:false})
    },
    clearTargrt () {
      this.targetOptions = []
      this.targrtId = ''
      this.showTargetManagement = false
      this.$refs.select.query = ''
    },
    search () {
      if (this.targrtId) {
        this.showTargetManagement = true
        this.$refs[this.type].getDetail(this.targrtId)
      }
    },
    openDoc () {
      window.open('http://webankpartners.gitee.io/wecube-docs/manual-open-monitor-config/#_6')
    }
  },
  components: {
    endpointManagement,
    groupManagement
  },
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
<style scoped lang="less">
  .is-danger {
    color: red;
    margin-bottom: 0px;
  }
  .search-input {
    height: 32px;
    padding: 4px 7px;
    font-size: 12px;
    border: 1px solid #dcdee2;
    border-radius: 4px;
    width: 230px;
  }

  .section-table-tip {
    margin: 24px 20px 0;
  }
  .search-input:focus {
    outline: 0;
    border-color: #57a3f3;
  }

  .search-input-content {
    display: inline-block;
    vertical-align: middle; 
  }
  .tag-width {
    cursor: auto;
    width: 55px;
    text-align: center;
  } 
</style>
