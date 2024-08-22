<template>
  <div class="portal-search">
    <ul class="search-ul">
      <li class="search-li">
        <Select
          style="width:400px;"
          v-model="endpoint"
          filterable
          clearable
          remote
          ref="select"
          :placeholder="$t('m_requestMoreData')"
          @on-open-change="getEndpointList('.')"
          :remote-method="getEndpointList"
        >
          <Option v-for="(option, index) in endpointList" :value="option.option_value" :label="option.option_text" :key="index">
            <TagShow :list="endpointList" name="option_type_name" :tagName="option.option_type_name" :index="index"></TagShow>
            {{option.option_text}}
          </Option>
          <Option value="moreTips" disabled>{{$t('m_tips_requestMoreData')}}</Option>
        </Select>
      </li>
      <li class="search-li">
        <Button type="primary" @click="routerChange">{{$t('m_button_search')}}</Button>
      </li>
    </ul>
  </div>
</template>

<script>
import TagShow from '@/components/Tag-show.vue'
export default {
  name: '',
  data() {
    return {
      endpoint: '',
      endpointObject: {},
      endpointList: []
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
  },
  mounted() {
    this.getEndpointList()
  },
  methods: {
    getEndpointList(query='.') {
      const params = {
        search: query,
        page: 1,
        size: 1000
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.resourceSearch.api, params, responseData => {
        this.endpointList = responseData
      })
    },
    routerChange(){
      if (!this.endpointObject) {
        return
      }
      this.$router.push({
        name: 'endpointView',
        params: this.endpointObject
      })
    },
  },
  components: {
    TagShow
  }
}
</script>

<style scoped lang="less">
.portal-search {
  padding-top:15%;
  width: 512px;
  margin: 0 auto;
}
.search-li {
  display: inline-block;
}
.search-ul>li:not(:first-child) {
  padding-left: 12px;
}
</style>
