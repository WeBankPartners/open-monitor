<template>
  <div class="portal-search">
    <ul class="search-ul">
      <li class="search-li">
        <Select
          style="width:400px;"
          v-model="endpoint"
          filterable
          remote
          ref="select"
          clearable
          :placeholder="$t('placeholder.input')"
          :remote-method="getEndpointList"
          >
          <Option v-for="(option, index) in endpointList" :value="option.option_value" :key="index">
            <Tag color="green" class="tag-width" v-if="option.type == 'sys'">system</Tag>
            <Tag color="cyan" class="tag-width" v-if="option.type == 'host'">host</Tag>
            <Tag color="blue" class="tag-width" v-if="option.type == 'mysql'">mysql </Tag>
            <Tag color="geekblue" class="tag-width" v-if="option.type == 'redis'">redis </Tag>
            <Tag color="purple" class="tag-width" v-if="option.type == 'tomcat'">tomcat</Tag>{{option.option_text}}</Option>
        </Select>
      </li>
      <li class="search-li">
        <button type="button" class="btn btn-sm btn-confirm-f"
          @click="routerChange">
          <i class="fa fa-search" ></i>
          {{$t('button.search')}}
        </button>
      </li>
    </ul>
  </div>
</template>

<script>
export default {
  name: '',
  data() {
    return {
      endpoint: '',
      endpointObject: {},
      endpointList: [],
    }
  },
  watch: {
    endpoint: function (val) {
      if (val) {
        this.endpointObject = this.endpointList.find(ep => {
          return ep.option_value === val
        })
      } else {
        this.endpointObject = {}
      }
    },
  },
  mounted () {
    this.getEndpointList('.')
  },
  methods: {
    getEndpointList(query) {
      let params = {
        search: query,
        page: 1,
        size: 1000
      }
      this.$root.$httpRequestEntrance.httpRequestEntrance('GET', this.$root.apiCenter.resourceSearch.api, params, (responseData) => {
        this.endpointList = responseData
      })
    },
    routerChange (){
      if (!this.endpointObject) {
        return
      }
      this.$router.push({ name: 'endpointView',params: this.endpointObject})
    }, 
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
    padding-left: 10px;
  }
</style>