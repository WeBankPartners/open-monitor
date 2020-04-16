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
          @on-open-change="getEndpointList('.')"
          :remote-method="getEndpointList"
          >
          <Option v-for="(option, index) in endpointList" :value="option.option_value" :key="index">
            <Tag :color="endpointTag[option.option_type_name] || choiceColor(option.option_type_name, index)" class="tag-width">{{option.option_type_name}}</Tag>{{option.option_text}}</Option>
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
import {endpointTag, randomColor} from '@/assets/config/common-config'
export default {
  name: '',
  data() {
    return {
      endpoint: '',
      endpointObject: {},
      endpointList: [],
      endpointTag: endpointTag,
      randomColor: randomColor,
      cacheColor: {}
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
    this.getEndpointList()
  },
  methods: {
    choiceColor (type,index) {
      let color = ''
      // eslint-disable-next-line no-prototype-builtins
      if (Object.keys(this.cacheColor).includes(type)) {
        color = this.cacheColor[type]
      } else {
        color = randomColor[index]
        this.cacheColor[type] = randomColor[index]
      }
      return color
    },
    getEndpointList(query='.') {
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