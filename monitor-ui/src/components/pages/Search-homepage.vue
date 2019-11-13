<template>
  <div class=" " style="text-align: center;padding-top:15%">
      <ul class="search-ul">
        <li class="search-li">
            <Searchinput :parentConfig="searchInputConfig"></Searchinput> 
        </li>
        <li class="search-li">
          <button type="button" class="btn btn-sm btn-confirm-f"
            @click="routerChange">
            <i class="fa fa-search" ></i>
            {{$t('button.search')}}
          </button>
        </li>
      </ul>
    <div>  
    </div>
  </div>
</template>

<script>
import Searchinput from '../components/Search-input'
export default {
  name: '',
  data() {
    return {
      searchInputConfig: {
        poptipWidth: 500,
        placeholder: 'placeholder.endpointSearch',
        inputStyle: "width:500px;",
        // api: '/dashboard/search'
        api: this.$root.apiCenter.resourceSearch.api
      }
    }
  },
  methods: {
    routerChange (){
      this.$root.$validate.isEmpty_reset(this.$root.$store.state.ip) ? '' : this.$router.push({ name: 'mainView',params: this.$root.$store.state.ip})
    }, 
  },
  components: {
    Searchinput
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