<template>
  <div class="business-monitor">
    <div style="display: flex;justify-content: space-between;margin-bottom: 8px">
      <ul class="search-ul">
        <li class="search-li">
          <RadioGroup
            v-model="type"
            type="button"
            button-style="solid"
            @on-change="typeChange"
            style="margin-right: 5px"
          >
            <Radio label="group">{{ $t('m_field_resourceLevel') }}</Radio>
            <Radio label="endpoint">{{ $t('m_tableKey_endpoint') }}</Radio>
          </RadioGroup>
        </li>
        <li class="search-li">
          <Select
            style="width:300px;"
            v-model="targrtId"
            filterable
            clearable
            remote
            ref="select"
            @on-change="search"
            @on-clear="typeChange"
          >
            <Option v-for="(option, index) in targetOptions"
                    :value="option.guid"
                    :key="index"
                    :label="option.display_name"
            >
              <TagShow :list="targetOptions" name="type" :tagName="option.type" :index="index"></TagShow>
              {{option.display_name}}
            </Option>
          </Select>
        </li>
        <li class="search-li" style="cursor: pointer;">
          <span style="font-size: 14px;" @click="openDoc">
            <i
              class="fa fa-book"
              aria-hidden="true"
              style="font-size:20px;color:#58a0e6;vertical-align: middle;margin-left:20px"
            >
            </i>
            {{$t('operationDoc')}}
          </span>
        </li>
      </ul>
    </div>
    <section v-show="showTargetManagement" style="margin-top: 16px;">
      <template v-if="type === 'group'">
        <groupManagement ref="group"></groupManagement>
      </template>
      <template v-if="type === 'endpoint'">
        <endpointManagement ref="endpoint"></endpointManagement>
      </template>
    </section>
  </div>
</template>

<script>
import endpointManagement from './business-monitor-endpoint.vue'
import groupManagement from './business-monitor-group.vue'
import TagShow from '@/components/Tag-show.vue'
export default {
  name: '',
  data() {
    return {
      type: 'group',
      targrtId: '',
      targetOptions: [],
      showTargetManagement: false
    }
  },

  mounted() {
    this.getTargrtList()
  },
  beforeDestroy() {
    this.$root.$store.commit('changeTableExtendActive', -1)
  },
  methods: {
    typeChange() {
      this.clearTargrt()
      this.getTargrtList()
    },
    getTargrtList() {
      const api = this.$root.apiCenter.getTargetByEndpoint + '/' + this.type
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', api, '', responseData => {
        this.targetOptions = responseData
        this.targrtId = this.targetOptions[0].guid
        this.search()
      }, {isNeedloading: false})
    },
    clearTargrt() {
      this.targetOptions = []
      this.targrtId = ''
      this.showTargetManagement = false
      this.$refs.select.query = ''
    },
    search() {
      if (this.targrtId) {
        this.showTargetManagement = true
        this.$refs[this.type].getDetail(this.targrtId)
      }
    },
    openDoc() {
      window.open('https://webankpartners.github.io/wecube-docs/manual-open-monitor-config-metrics/')
    }
  },
  components: {
    endpointManagement,
    groupManagement,
    TagShow
  },
}
</script>

<style lang="less">
.business-monitor {
  .ivu-radio-group-button .ivu-radio-wrapper-checked {
    background: #2d8cf0;
    color: #fff;
  }
  .search-li {
    display: inline-block;
  }
  .search-ul>li:not(:first-child) {
    padding-left: 12px;
  }
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
