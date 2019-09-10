<template>
  <div class=" " style="text-align: center;padding-top:15%">
      <ul class="search-ul">
        <li class="search-li">
            <Searchinput @sendInputValue="obtainInputValue" :parentConfig="searchInputConfig"></Searchinput> 
        </li>
        <li class="search-li">
            <Button type="primary" @click="routerChange" icon="ios-search">搜索</Button>
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
      ip: '',
      searchInputConfig: {
        poptipWidth: 500,
        placeholder: '请输入主机名或IP地址，可模糊匹配',
        inputStyle: "width:500px;"
      }
    }
  },
  mounted: {
  },
  methods: {
    routerChange (){
      if (this.ip) {
        this.$router.push({ name: 'mainView', params: { ip: this.ip }})
      } else {
        this.$Message.warning('请选择有效IP！')
      }
    },  
    obtainInputValue (ip) {
      this.ip = ip
    } 
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
