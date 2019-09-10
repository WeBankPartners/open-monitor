<template>
  <div class=" " style="text-align: center;padding-top:15%">
      <ul class="search-ul">
        <li class="search-li">
            <Searchinput @sendInputValue="obtainInputValue"></Searchinput> 
        </li>
        <li class="search-li">
            <Button type="primary" @click="sss" icon="ios-search">搜索</Button>
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
      ip: ''
    }
  },
  methods: {
    sss (){
        this.$router.push({ name: 'mainView', params: { ip: this.ip }})
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
