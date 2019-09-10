<template>
  <div class="search-input-content">
    <Poptip placement="bottom" :width="parentConfig.poptipWidth">
      <input v-model.trim="ip.label"
      :placeholder="parentConfig.placeholder"
      @input="userInput"
      type="text"
      :style= "parentConfig.inputStyle"
      class="search-input" />
      <div class="poptip-content" slot="content" v-if="showSearchTips">
        <ul>
          <template v-for="(resItem, resIndex) in searchResult">
            <li class="ul-option" @click="choiceRes(resItem)" :key="resIndex">
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
  props: {
    parentConfig: Object
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

<style>
  .ivu-poptip-body {
    padding: 0;
  }
  .poptip-content {
    padding: 4px 0;
  }
  .poptip-content {
    max-height: 300px;
  }
</style>
<style scoped lang="less">
.search-input {
    display: inline-block;
    height: 32px;
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

  .ul-option {
    font-weight: 500;
    text-align: left;
    padding: 7px 16px;
    font-size: 12px;
    white-space: nowrap;
    cursor:pointer;
  }
  .ul-option:hover {
    background: @gray-hover;
  }
</style>
