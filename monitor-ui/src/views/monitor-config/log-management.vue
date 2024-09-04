<template>
  <div class="log-management">
    <div class='log-management-top'>
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
      <Select
        style="width:300px;"
        v-model="targetId"
        filterable
        clearable
        remote
        ref="select"
        @on-change="onFilterChange"
      >
        <Option v-for="(option, index) in targetOptions" :value="option.guid" :label="option.display_name" :key="index">
          <TagShow :list="targetOptions" name="type" :tagName="option.type" :index="index"></TagShow>
          {{option.display_name}}
        </Option>
      </Select>
      <Input
        v-model="alarmName"
        @on-change="onFilterChange"
        style="width:250px;margin-left:12px;"
        clearable
        :placeholder="$t('m_placeholder_input') + $t('m_alarmName')"
      />
    </div>
    <section v-show="showTargetManagement" style="margin-top: 16px;">
      <keywordContent ref='keywordContent' :keywordType="typeMap[type]"></keywordContent>
    </section>
  </div>
</template>

<script>
import {debounce} from 'lodash'
import keywordContent from './keyword-content.vue'
import TagShow from '@/components/Tag-show.vue'
export default {
  name: '',
  data() {
    return {
      type: 'group',
      targetId: '',
      targetOptions: [],
      showTargetManagement: false,
      typeMap: {
        group: 'service',
        endpoint: 'endpoint'
      },
      alarmName: ''
    }
  },
  async mounted() {
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
        this.targetId = this.targetOptions[0].guid
        this.search()
      }, {isNeedloading: false})
    },
    clearTargrt() {
      this.targetOptions = []
      this.targetId = ''
      this.showTargetManagement = false
      this.$refs.select.query = ''
    },
    search() {
      if (this.targetId) {
        this.showTargetManagement = true
        this.$refs.keywordContent.getDetail(this.targetId, this.alarmName)
      }
    },
    onFilterChange: debounce(function () {
      this.search()
    }, 300)
  },
  components: {
    TagShow,
    keywordContent
  },
}
</script>

<style lang="less">
.log-management {
  .ivu-radio-group-button .ivu-radio-wrapper-checked {
    background: #2d8cf0;
    color: #fff;
  }
}
</style>
<style scoped lang="less">
.log-management-top {
  display: flex;
  align-items: center;
}
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
