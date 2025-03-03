<!--业务配置-->
<template>
  <div class="business-monitor">
    <div class="business-monitor-header">
      <RadioGroup
        v-model="type"
        type="button"
        button-style="solid"
        @on-change="typeChange(true)"
        style="margin-right: 5px"
      >
        <Radio label="group">{{ $t('m_field_resourceLevel') }}</Radio>
        <Radio label="endpoint">{{ $t('m_tableKey_endpoint') }}</Radio>
      </RadioGroup>
      <Select
        :key='selectKey'
        style="width:250px;"
        v-model="targrtId"
        filterable
        clearable
        remote
        ref="select"
        @on-change="() => {
          metricKey = ''
          search()
        }"
        @on-clear="typeChange(false)"
        @on-open-change="onSelectOpenChange"
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
      <Input v-model.trim="metricKey"
             :placeholder="$t('m_enter_indicator_key_tips')"
             clearable
             style="width:250px; margin-left: 5px"
             @on-change='search'
      />
      <span style="font-size: 14px; cursor: pointer;" @click="openDoc">
        <i
          class="fa fa-book"
          aria-hidden="true"
          style="font-size:20px;color:#58a0e6;vertical-align: middle;margin-left:20px"
        >
        </i>
        {{$t('m_operationDoc')}}
      </span>
    </div>
    <div v-if="!targrtId" style="margin: 10px 0">
      <Alert type="error">
        <span>{{ $t('m_empty_tip_1') }}</span>
        <span v-if="type === 'group'">{{ $t('m_field_resourceLevel') }}</span>
        <span v-if="type === 'endpoint'">{{ $t('m_field_endpoint') }}</span>
      </Alert>
    </div>
    <div v-if="targrtId && isDataEmpty" style="margin: 10px 0">
      <Alert type="error">
        <span>{{ $t('m_noData') }}</span>
      </Alert>
    </div>

    <section v-show="showTargetManagement" class='business-monitor-content' style="margin-top: 16px;">
      <template v-if="type === 'group'">
        <groupManagement ref="group" @feedbackInfo="onFeedbackInfo"></groupManagement>
      </template>
      <template v-if="type === 'endpoint'">
        <endpointManagement @feedbackInfo="onFeedbackInfo" ref="endpoint"></endpointManagement>
      </template>
    </section>
  </div>
</template>

<script>
import {debounce} from 'lodash'
import endpointManagement from './business-monitor-endpoint.vue'
import groupManagement from './business-monitor-group.vue'
import TagShow from '@/components/Tag-show.vue'

export const custom_api_enum = [
  {
    getTargetByEndpointByType: 'get'
  }
]

export default {
  name: '',
  data() {
    return {
      type: 'group',
      targrtId: '',
      targetOptions: [],
      showTargetManagement: false,
      metricKey: '',
      isDataEmpty: false,
      selectKey: '',
      request: this.$root.$httpRequestEntrance.httpRequestEntrance,
      apiCenter: this.$root.apiCenter,
    }
  },

  mounted() {
    this.getTargrtList()
  },
  beforeDestroy() {
    this.$root.$store.commit('changeTableExtendActive', -1)
  },
  methods: {
    typeChange(needDefaultTarget) {
      this.metricKey = ''
      // this.clearTargrt()
      this.getTargrtList(needDefaultTarget)
      this.selectKey = +new Date() + ''
    },
    getTargrtList(needDefaultTarget = true) {
      const api = this.apiCenter.getTargetByEndpoint + '/' + this.type
      this.request('GET', api, '', responseData => {
        this.targetOptions = responseData || []
        if (this.targetOptions.length > 0 && needDefaultTarget) {
          this.targrtId = this.targetOptions[0].guid
        }
        this.search()
      }, {isNeedloading: false})
    },
    clearTargrt() {
      this.targetOptions = []
      this.targrtId = ''
      this.showTargetManagement = false
      this.$refs.select.query = ''
    },
    search: debounce(function () {
      if (this.targrtId) {
        this.showTargetManagement = true
        this.$refs[this.type].getDetail(this.targrtId, this.metricKey)
      }
    }, 300),
    openDoc() {
      window.open('https://webankpartners.github.io/wecube-docs/manual-open-monitor-config-metrics/')
    },
    onFeedbackInfo(allData) {
      this.isDataEmpty = allData.length === 0
    },
    onSelectOpenChange(open) {
      if (open) {
        this.getTargrtList(false)
      }
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
    background: #5384FF;
    color: #fff;
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
  .business-monitor-header {
    display: flex;
    justify-content: flex-start;
    align-items: center;
    margin-bottom: 8px
  }
  .business-monitor-content {
    max-height: ~'calc(100vh - 170px)';
    overflow-y: auto;
  }
</style>
