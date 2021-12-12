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
            <Option v-for="(option, index) in targetOptions" :value="option.option_value" :key="index">{{option.option_text}}
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
      </ul>
    </section> 
    <section v-show="showTargetManagement" style="margin-top: 16px;">
      <template v-if="type === 'group'">
        <groupManagement ref="group"></groupManagement>
      </template>
      <template v-if="type === 'endpoint'">
        <endpointManagement ref="endpoint"></endpointManagement>
      </template>
      <template v-if="type === 'service'">
        <serviceManagement ref="service"></serviceManagement>
      </template>
    </section>
  </div>
</template>

<script>
import endpointManagement from './threshold-management-endpoint.vue'
import groupManagement from './threshold-management-group.vue'
import serviceManagement from './threshold-management-service.vue'
export default {
  name: '',
  data() {
    return {
      type: 'service',
      typeList: [
        {label: 'field.endpoint', value: 'endpoint'},
        {label: 'field.resourceLevel', value: 'service'},
        {label: 'field.group', value: 'group'}
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
    typeChange () {
      this.clearTargrt()
    },
    getTargrtList () {
      const api = `/monitor/api/v2/alarm/strategy/search?type=${this.type}&search=`
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
        const find = this.targetOptions.find(item => item.option_value === this.targrtId)
        this.$refs[this.type].getDetail(this.targrtId, find.type)
      }
    }
  },
  components: {
    endpointManagement,
    groupManagement,
    serviceManagement
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
