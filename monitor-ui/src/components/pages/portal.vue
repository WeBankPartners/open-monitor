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
            搜索
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
        placeholder: '请输入主机名或IP地址，可模糊匹配',
        inputStyle: "width:500px;",
        api: '/dashboard/search'
      }
    }
  },
  methods: {
    routerChange (){
      if (Object.keys(this.$store.state.ip).length === 0) {
        this.$Message.warning('请选择有效IP！')
      } else {
        this.$router.push({ name: 'mainView'})
      }
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
