<template>
  <div class="search-input-content">
        <Poptip placement="bottom" width="300">
          <input v-model.trim="ip.label"
          placeholder="请输入主机名或IP地址，可模糊匹配"
          @input="userInput"
          type="text"
          class="search-input" />
          <div class="api" slot="content" v-if="showSearchTips">
            <ul>
              <template v-for="(resItem, resIndex) in searchResult">
                <li style="line-height: 20px;font-weight: 500;cursor:pointer" @click="choiceRes(resItem)" :key="resIndex">
                  <span>{{resItem.option_text}}</span>
                </li>
              </template>  
                
            </ul> 
          </div>
        </Poptip>
  </div>
</template>

<script>
export default {
  name: '',
  data() {
    return {
      ip: {
        label: '',
        value: '',
      },
      showSearchTips: false,
      searchResult: [],
    }
  },
  methods: {
    userInput () {
      this.showSearchTips = false
      this.request()
    },
    choiceRes (resItem) {
      this.ip.label = resItem.option_text
      this.ip.value = resItem.option_value
      this.$emit('sendInputValue', this.ip)
      this.showSearchTips = false
    },
    request () {
      if (!this.ip.label) {
        return
      }
      let params = {
        search: this.ip.label
      }
      this.$httpRequestEntrance.httpRequestEntrance('GET','/dashboard/search', params, responseData => {
        this.searchResult = responseData
      })
      this.showSearchTips = true
    } 
  },
  components: {},
}
</script>

<style scoped lang="less">
.search-input {
    display: inline-block;
    width: 300px;
    height: 32px;
    line-height: 1.5;
    padding: 4px 7px;
    font-size: 12px;
    border: 1px solid #dcdee2;
    border-radius: 4px;
    color: #515a6e;
    background-color: #fff;
    background-image: none;
    position: relative;
    cursor: text

  }
  .search-input:focus {
    outline: 0;
    border-color: #57a3f3;
  }

  .search-input-content {
    display: inline-block;
    vertical-align: middle; 
  }
</style>
